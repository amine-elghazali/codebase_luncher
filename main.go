package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type item struct {
	name string
}

func (i item) Title() string       { return i.name }
func (i item) Description() string { return "" }
func (i item) FilterValue() string { return i.name }

type model struct {
	list     list.Model
	choices  []string
	selected map[int]struct{}
}

type itemDelegate struct {
	list.DefaultDelegate
	selected map[int]struct{}
}

func newItemDelegate(selected map[int]struct{}) itemDelegate {
	delegate := itemDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
		selected:        selected,
	}

	// Define base styles with minimal, elegant aesthetics
	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")) // Subtle gray

	// Style for when cursor is on an item - clean and minimal
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		Border(lipgloss.Border{
			Left:  "│",
			Right: "│",
		}).
		BorderForeground(lipgloss.Color("99")). // Soft purple
		Foreground(lipgloss.Color("255")).      // Pure white text
		Background(lipgloss.Color("236")).      // Dark gray background
		Padding(0, 1).
		MarginLeft(1)

	return delegate
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	title := i.name
	var finalStyle lipgloss.Style

	// Check if item is marked for selection
	if _, ok := d.selected[index]; ok {
		title = "● " + title // Filled circle for selected
	} else {
		title = "○ " + title // Empty circle for unselected
	}

	if index == m.Index() {
		// Cursor is on this item
		if _, ok := d.selected[index]; ok {
			// Selected and cursor is on it
			finalStyle = d.Styles.SelectedTitle.Copy().
				Background(lipgloss.Color("238")).      // Slightly lighter background
				BorderForeground(lipgloss.Color("99")). // Soft purple
				Bold(true)
		} else {
			// Just cursor, not selected
			finalStyle = d.Styles.SelectedTitle
		}
	} else {
		baseStyle := lipgloss.NewStyle().
			MarginLeft(2)

		if _, ok := d.selected[index]; ok {
			// Selected but cursor elsewhere
			finalStyle = baseStyle.Copy().
				Foreground(lipgloss.Color("99")). // Soft purple
				Bold(true)
		} else {
			// Neither selected nor cursor
			finalStyle = baseStyle.Copy().
				Foreground(lipgloss.Color("246")) // Subtle gray
		}
	}

	fmt.Fprint(w, finalStyle.Render(title))
}

func main() {
	files, err := ioutil.ReadDir(".")
	if err != nil {
		fmt.Println("Error reading directory:", err)
		os.Exit(1)
	}

	var items []list.Item
	for _, file := range files {
		if file.IsDir() {
			items = append(items, item{name: file.Name()})
		}
	}

	selected := make(map[int]struct{})
	delegate := newItemDelegate(selected)

	l := list.New(items, delegate, 40, 20)

	m := model{
		list:     l,
		selected: selected,
	}

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			for i := range m.selected {
				cmd := exec.Command("code", m.list.Items()[i].(item).name)
				cmd.Start()
			}
			return m, tea.Quit
		case "q":
			return m, tea.Quit
		case " ":
			i := m.list.Index()
			if _, ok := m.selected[i]; ok {
				delete(m.selected, i)
			} else {
				m.selected[i] = struct{}{}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("246")). // Subtle gray
		Render("\n[space] select   [enter] launch   [q] quit")

	return m.list.View() + helpText
}
