package main

import (
	"fmt"
	"log"

	"github.com/fhs/gompd/v2/mpd"
)

// var baseStyle = lipgloss.NewStyle().
// 	BorderStyle(lipgloss.NormalBorder()).
// 	BorderForeground(lipgloss.Color("240"))

// type model struct {
// 	songs    []string
// 	cursor   int
// 	selected map[int]struct{}
// }

//func makeTable(artists []string, songs []string) {
//	columns := []table.Column{
//		{Title: "Song Name", Width: 20},
//		{Title: "Artist", Width: 4},
//	}
//
//	rows := table[row]
//
//}

// func initialModel(songs []string) model {
// 	return model{
// 		songs:    songs,
// 		selected: make(map[int]struct{}),
// 	}
// }

// func (m model) View() string {
// 	s := "Songs Listed?\n\n"
// 	for i, song := range m.songs {
// 		cursor := " "
// 		if m.cursor == i {
// 			cursor = ">"
// 		}
// 		s += fmt.Sprintf("%s %s\n", cursor, song)
// 	}
// 	s += "\nPress q to quit.\n"

// 	return s
// }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "ctrl+c", "q":
// 			return m, tea.Quit
// 		case "up", "k":
// 			if m.cursor > 0 {
// 				m.cursor--
// 			}
// 		case "down", "j":
// 			if m.cursor < len(m.songs)-1 {
// 				m.cursor++
// 			}
// 		case "enter", " ":
// 			_, ok := m.selected[m.cursor]
// 			if ok {
// 				delete(m.selected, m.cursor)
// 			} else {
// 				m.selected[m.cursor] = struct{}{}
// 			}
// 		}
// 	}
// 	return m, nil
// }

// func (m model) Init() tea.Cmd {
// 	return nil
// }

func main() {
	var err error

	conn, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		log.Fatalln(err)
	}

	defer conn.Close()

	_, err = conn.Update("")
	if err != nil {
		log.Fatalln(err)
	}

	songs, err := conn.ListAllInfo("/")
	if err != nil {
		log.Fatalln(err)
	}

	for _, song := range songs {
		fmt.Println(song)
	}

	//p := tea.NewProgram(initialModel(songs))
	//if _, err := p.Run(); err != nil {
	//	fmt.Printf("Alas, there's been an error: %v", err)
	//	os.Exit(1)
	//}

}
