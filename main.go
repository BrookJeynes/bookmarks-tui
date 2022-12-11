package main

import (
	"bufio"
	"log"
	"os"
	"strings"

	"github.com/atotto/clipboard"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/nsf/termbox-go"
)

const PATH string = ".bookmarks"

type ViewState struct {
	ItemSelect  int
	LastDeleted string
	HelpMenu    bool
}

func main() {
	// Initialise termui
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialise termui: %v", err)
	}

	defer ui.Close()

	// Get terminal window size
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	width, height := termbox.Size()

	defer termbox.Close()

	// Open file
	file, _ := os.OpenFile(PATH, os.O_RDWR, 0644)

	defer file.Close()

	help_menu := initialise_help_menu(width/4, height/4, 3*width/4, 3*height/4)
	bookmark_list := initialise_list(0, 0, width, height)

	vs := ViewState{
		ItemSelect:  0,
		LastDeleted: "",
		HelpMenu:    false,
	}

	// Loop through file contents and add them to list
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		bookmark_list.Rows = append(bookmark_list.Rows, scanner.Text())
	}

	ui.Render(bookmark_list)

	// User input
	previousKey := ""
	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "j", "<Down>":
			if vs.ItemSelect < len(bookmark_list.Rows)-1 {
				vs.ItemSelect++
			}

			bookmark_list.ScrollDown()
		case "k", "<Up>":
			if vs.ItemSelect > 0 {
				vs.ItemSelect--
			}

			bookmark_list.ScrollUp()
		case "g":
			if previousKey == "g" {
				vs.ItemSelect = 0
				bookmark_list.ScrollTop()
			}
		case "G", "<End>":
			vs.ItemSelect = len(bookmark_list.Rows) - 1
			bookmark_list.ScrollBottom()
		case "d":
			index := vs.ItemSelect

			if previousKey == "d" {
				vs.LastDeleted = bookmark_list.Rows[index]

				new_items, err := delete_item(index, bookmark_list.Rows)

				if err != nil {
					panic(err)
				}

				bookmark_list.Rows = new_items
			}
		case "y":
			if previousKey == "y" {
				clipboard.WriteAll(bookmark_list.Rows[vs.ItemSelect])
			}
		case "u":
			if vs.LastDeleted != "" {
				new_items, err := add_item(vs.LastDeleted, bookmark_list.Rows)

				if err != nil {
					panic(err)
				}

				bookmark_list.Rows = new_items
				vs.LastDeleted = ""
			}
		case "a":
			contents, _ := clipboard.ReadAll()

			new_items, err := add_item(contents, bookmark_list.Rows)

			if err != nil {
				panic(err)
			}

			bookmark_list.Rows = new_items
		case "?":
			ui.Clear()

			vs.HelpMenu = !vs.HelpMenu
		}

		if previousKey == "g" || previousKey == "d" || previousKey == "y" {
			previousKey = ""
		} else {
			previousKey = e.ID
		}

		if vs.HelpMenu {
			ui.Render(bookmark_list, help_menu)
		} else {
			ui.Render(bookmark_list)
		}
	}
}

func delete_item(item_index int, items []string) ([]string, error) {
	items = append(items[:item_index], items[item_index+1:]...)
	err := os.WriteFile(PATH, []byte(strings.Join(items, "\n")), 0666)

	return items, err
}

func add_item(contents string, items []string) ([]string, error) {
	items = append(items, contents)
	err := os.WriteFile(PATH, []byte(strings.Join(items, "\n")), 0666)

	return items, err
}

func initialise_list(x1 int, y1 int, x2 int, y2 int) *widgets.List {
	l := widgets.NewList()
	l.Title = "Bookmarks - ? for help"
	l.Rows = []string{}
	l.TextStyle = ui.NewStyle(ui.ColorWhite)
	l.SelectedRowStyle = ui.NewStyle(ui.ColorCyan)
	l.WrapText = false
	l.SetRect(x1, y1, x2, y2)

	return l
}

func initialise_help_menu(x1 int, y1 int, x2 int, y2 int) *widgets.Paragraph {
	p := widgets.NewParagraph()
	p.Title = "Help"
	p.Text = "q - quit\n? - help menu\n\na - add from clipboard\ndd - delete\nyy - copy\nu - undo\n\nj - down\nk - up\ngg - scroll to top\nG - scroll to bottom\n"
	p.SetRect(x1, y1, x2, y2)
	p.TextStyle.Fg = ui.ColorWhite
	p.BorderStyle.Fg = ui.ColorCyan

	return p
}
