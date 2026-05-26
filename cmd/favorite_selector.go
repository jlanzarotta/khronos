/*
Copyright © 2018-2025 Jeff Lanzarotta
All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

 1. Redistributions of source code must retain the above copyright notice,
    this list of conditions and the following disclaimer.

 2. Redistributions in binary form must reproduce the above copyright notice,
    this list of conditions and the following disclaimer in the documentation
    and/or other materials provided with the distribution.

 3. Neither the name of the copyright holder nor the names of its contributors
    may be used to endorse or promote products derived from this software
    without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
POSSIBILITY OF SUCH DAMAGE.
*/
package cmd

import (
	"os"
	"strconv"
	"strings"

	"khronos/constants"
	"khronos/internal/jira"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"github.com/mattn/go-isatty"
)

// favoriteSelectorModel renders the favorites as a fully-bordered go-pretty
// table (identical borders to `show --favorites`) and lets the user move a
// cursor, jump to a row by number, and select. The highlight is done by
// go-pretty's SetRowPainter, so the box-drawing stays intact rather than being
// overlaid with ANSI after the fact.
type favoriteSelectorModel struct {
	favs             []Favorite
	descriptionFound bool
	ticketFound      bool

	cursor   int
	numBuf   string // accumulates typed digits for jump-to-row
	chosen   int    // selected index, or -1 if none
	quitting bool

	caption string // action text, e.g. "Select a favorite to add"
	config  string // config file path, shown in the help line
}

func newFavoriteSelectorModel(caption, config string, favs []Favorite) favoriteSelectorModel {
	var descriptionFound, ticketFound bool
	for _, f := range favs {
		if len(f.Ticket) > 0 {
			ticketFound = true
		}
		if len(f.Description) > 0 {
			descriptionFound = true
		}
	}

	return favoriteSelectorModel{
		favs:             favs,
		descriptionFound: descriptionFound,
		ticketFound:      ticketFound,
		cursor:           0,
		chosen:           -1,
		caption:          caption,
		config:           config,
	}
}

// renderTable builds the go-pretty table string with the cursor row highlighted.
// This mirrors showFavoritesTable's column logic so the look matches exactly.
func (m favoriteSelectorModel) renderTable() string {
	t := table.NewWriter()

	// Match the column layout used by showFavoritesTable.
	var requiredNoteColumn int
	if m.descriptionFound && m.ticketFound {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.DESCRIPTION, constants.URL, constants.REQUIRE_NOTE_WITH_ASTERISK})
		requiredNoteColumn = 5
	} else if m.descriptionFound {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.DESCRIPTION, constants.REQUIRE_NOTE_WITH_ASTERISK})
		requiredNoteColumn = 4
	} else if m.ticketFound {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.URL, constants.REQUIRE_NOTE_WITH_ASTERISK})
		requiredNoteColumn = 4
	} else {
		t.AppendHeader(table.Row{"#", constants.PROJECT_TASK, constants.REQUIRE_NOTE_WITH_ASTERISK})
		requiredNoteColumn = 3
	}

	style := table.StyleDefault
	style.Format.Header = text.FormatUpper
	t.SetStyle(style)

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: 3, WidthMax: 3},
		{Number: requiredNoteColumn, WidthMin: 13, WidthMax: 13},
	})

	for i, f := range m.favs {
		// Display number is 1-based; stored as an int so the row painter can
		// match against the cursor (cursor+1). The selector still returns the
		// 0-based slice index to callers.
		num := i + 1
		if m.descriptionFound && m.ticketFound {
			t.AppendRow(table.Row{num, f.Favorite, f.Description, jira.FormatJiraUrl(jira.JiraBrowseTicketUrl, f.Ticket), f.RequireNote})
		} else if m.descriptionFound {
			t.AppendRow(table.Row{num, f.Favorite, f.Description, f.RequireNote})
		} else if m.ticketFound {
			t.AppendRow(table.Row{num, f.Favorite, jira.FormatJiraUrl(jira.JiraBrowseTicketUrl, f.Ticket), f.RequireNote})
		} else {
			t.AppendRow(table.Row{num, f.Favorite, f.RequireNote})
		}
	}

	// Highlight the cursor row. RowPainter receives the row and is invoked per
	// row before render; we color the row whose number matches cursor+1 (the
	// "#" column is 1-based while the cursor is 0-based).
	cursor := m.cursor
	t.SetRowPainter(func(row table.Row) text.Colors {
		if len(row) > 0 {
			if num, ok := row[0].(int); ok && num == cursor+1 {
				return text.Colors{text.BgBlue, text.FgHiWhite}
			}
		}
		return text.Colors{}
	})

	return t.Render()
}

func (m favoriteSelectorModel) Init() tea.Cmd { return nil }

func (m favoriteSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil

		case "down", "j":
			if m.cursor < len(m.favs)-1 {
				m.cursor++
			}
			return m, nil

		case "enter":
			if m.numBuf != "" {
				// The user types the 1-based number they see; convert to the
				// 0-based cursor index.
				if n, err := strconv.Atoi(m.numBuf); err == nil {
					if n >= 1 && n <= len(m.favs) {
						m.cursor = n - 1
					}
				}
				m.numBuf = ""
				return m, nil
			}
			m.chosen = m.cursor
			m.quitting = true
			return m, tea.Quit

		case "backspace":
			if m.numBuf != "" {
				m.numBuf = m.numBuf[:len(m.numBuf)-1]
			}
			return m, nil

		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			m.numBuf += msg.String()
			return m, nil
		}
	}

	return m, nil
}

func (m favoriteSelectorModel) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.renderTable())
	b.WriteString("\n")
	b.WriteString(m.helpLine())
	b.WriteString("\n")
	return b.String()
}

func (m favoriteSelectorModel) helpLine() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	keys := "up/down: navigate - enter: select - type a number + enter: jump - q/esc: cancel"
	if m.numBuf != "" {
		keys = "jump to row: " + m.numBuf + "  (enter to go, backspace to edit)"
	}

	// First line: the action caption plus the config file path. Second line:
	// the keybindings (or the jump prompt while typing a number).
	var b strings.Builder
	header := m.caption
	if m.config != "" {
		if header != "" {
			header += "  "
		}
		header += "(config: " + m.config + ")"
	}
	if header != "" {
		b.WriteString(dim.Render(header))
		b.WriteString("\n")
	}
	b.WriteString(dim.Render(keys))
	return b.String()
}

// selectFavorite launches the interactive favorites selector and returns the
// chosen index along with ok=true. On cancel it returns (-1, false). The
// caption and config path are shown in the help line below the table.
func selectFavorite(caption, config string, favs []Favorite) (int, bool, error) {
	m := newFavoriteSelectorModel(caption, config, favs)

	// Inline (no alt-screen): renders in normal terminal flow.
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return -1, false, err
	}

	fm, ok := final.(favoriteSelectorModel)
	if !ok || fm.chosen < 0 {
		return -1, false, nil
	}
	return fm.chosen, true, nil
}

// interactiveTerminal reports whether stdin/stdout are a real terminal.
func interactiveTerminal() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) && isatty.IsTerminal(os.Stdin.Fd())
}
