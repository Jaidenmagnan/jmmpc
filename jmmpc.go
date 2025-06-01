package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/fhs/gompd/v2/mpd"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

type tickMsg time.Time

const (
	padding  = 2
	maxWidth = 60
)

type model struct {
	table        table.Model
	conn         *mpd.Client
	progress     progress.Model
	percent      float64
	current_song string
	majorno      int
}

func getThemeColor(colorName string) string {
	colors := map[string]string{
		"red":    "1",
		"green":  "2",
		"yellow": "3",
		"blue":   "4",
		"purple": "5",
		"cyan":   "6",
		"white":  "7",
	}

	if color, exists := colors[colorName]; exists {
		return color
	}

	return "5"
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Microsecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func updateProgress(status mpd.Attrs) float64 {
	if timeStr, ok := status["time"]; ok {
		parts := strings.Split(timeStr, ":")
		if len(parts) == 2 {
			current, _ := strconv.ParseFloat(parts[0], 64)
			total, _ := strconv.ParseFloat(parts[1], 64)

			if total > 0 {
				return current / total
			}
		}
	}
	return 0
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	status, err := m.conn.Status()
	if err != nil {
		fmt.Println(err)
	}

	cs, err := m.conn.CurrentSong()
	if err != nil {
		log.Fatalln(err)
	}
	m.percent = updateProgress(status)
	m.current_song = cs["file"]

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.progress.Width = min(msg.Width-padding*2-4, maxWidth)
		return m, nil
	case tickMsg:
		return m, tickCmd()
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "p":

			if status["state"] == "play" {
				m.conn.Pause(true)
			} else {
				m.conn.Pause(false)
			}
		case "n":
			m.conn.Next()
			m.conn.Pause(true)
			m.conn.Pause(false)
		case "b":
			m.conn.Previous()
			m.conn.Pause(true)
			m.conn.Pause(false)
		case "enter":
			minorno, err := strconv.Atoi(m.table.SelectedRow()[0])
			if err != nil {
				log.Fatalln(err)
			}
			m.conn.PlayID(minorno + m.majorno)
			m.conn.Pause(true)
			m.conn.Pause(false)

		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func cleanSongName(filename string) string {
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	return name
}

func (m model) View() string {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width, height = 80, 24
	}

	var song_string = ""
	if m.current_song == "" {
		song_string = "♫ no song playing ♫"
	} else {
		song_string = "♫ " + cleanSongName(m.current_song) + " ♫"
	}

	content := lipgloss.JoinVertical(lipgloss.Center,
		baseStyle.Render(m.table.View()),
		m.progress.ViewAs(m.percent),
		song_string,
		helpStyle("Press q to quit"),
	)

	return lipgloss.Place(width, height,
		lipgloss.Center, lipgloss.Center,
		content)
}

func createTable(conn *mpd.Client, color string) (table.Model, int) {
	columns := []table.Column{
		{Title: "Track", Width: 5},
		{Title: "Song", Width: 30},
		{Title: "Artist", Width: 20},
		{Title: "Album", Width: 10},
	}

	rows := []table.Row{}

	songs, err := conn.ListAllInfo("/")
	if err != nil {
		log.Fatalln(err)
	}

	var majorno = 9999999

	for _, song := range songs {
		minorno, err := conn.AddID(song["file"], -1)
		majorno = min(minorno, majorno)
		if err != nil {
			log.Fatalln(err)
		}

		rows = append(rows, table.Row{
			fmt.Sprintf("%d", minorno-majorno),
			song["file"],
			song["Artist"],
			song["Album"],
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("8")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color(color)).
		Bold(true)
	t.SetStyles(s)

	return t, majorno
}

func main() {
	var err error

	colorFlag := flag.String("c", "purple", "Theme color (red, purple, blue, green, etc.)")
	flag.Parse()
	color := getThemeColor(*colorFlag)

	conn, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		log.Fatalln(err)
	}

	conn.Update("/")
	time.Sleep(300 * time.Millisecond)

	conn.Repeat(true)

	prog := progress.New(progress.WithSolidFill(color))

	t, majorno := createTable(conn, color)
	m := model{t, conn, prog, 0, "", majorno}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
