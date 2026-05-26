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
	"strconv"
	"strings"

	"khronos/constants"
	"khronos/internal/models"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/dromara/carbon/v2"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

// entrySelectorModel renders a list of entries as a fully-bordered go-pretty
// table and lets the user move a cursor, jump to a row by number, and select.
// It mirrors favoriteSelectorModel; the difference is the columns rendered and
// that there is no config path in the help line.
type entrySelectorModel struct {
	entries []models.Entry

	cursor   int
	numBuf   string // accumulates typed digits for jump-to-row
	chosen   int    // selected index, or -1 if none
	quitting bool

	caption string // action text, e.g. "Select an entry to amend"
}

func newEntrySelectorModel(caption string, entries []models.Entry) entrySelectorModel {
	return entrySelectorModel{
		entries: entries,
		cursor:  0,
		chosen:  -1,
		caption: caption,
	}
}

// renderTable builds the go-pretty table string with the cursor row
// highlighted. The "#" column is displayed 1-based to match the original amend
// listing, while the cursor still tracks the 0-based slice index internally.
func (m entrySelectorModel) renderTable() string {
	t := table.NewWriter()

	t.AppendHeader(table.Row{"#", constants.PROJECT_NORMAL_CASE, constants.TASK_NORMAL_CASE, constants.DATE_TIME_NORMAL_CASE})

	style := table.StyleDefault
	style.Format.Header = text.FormatUpper
	t.SetStyle(style)

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, WidthMin: 3, WidthMax: 5},
	})

	for i, entry := range m.entries {
		// Display number is 1-based; store it as an int so the row painter can
		// match against the cursor (cursor+1).
		t.AppendRow(table.Row{
			i + 1,
			entry.Project,
			entry.GetTasksAsString(),
			carbon.Parse(entry.EntryDatetime).SetTimezone(carbon.Local).ToIso8601String(),
		})
	}

	cursor := m.cursor
	t.SetRowPainter(func(row table.Row) text.Colors {
		// row[0] is the 1-based "#"; the cursor is 0-based, so compare to cursor+1.
		if len(row) > 0 {
			if idx, ok := row[0].(int); ok && idx == cursor+1 {
				return text.Colors{text.BgBlue, text.FgHiWhite}
			}
		}
		return text.Colors{}
	})

	return t.Render()
}

func (m entrySelectorModel) Init() tea.Cmd { return nil }

func (m entrySelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.entries)-1 {
				m.cursor++
			}
			return m, nil

		case "enter":
			if m.numBuf != "" {
				// The user types the 1-based number they see; convert to the
				// 0-based cursor index.
				if n, err := strconv.Atoi(m.numBuf); err == nil {
					if n >= 1 && n <= len(m.entries) {
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

func (m entrySelectorModel) View() string {
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

func (m entrySelectorModel) helpLine() string {
	dim := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))

	keys := "up/down: navigate - enter: select - type a number + enter: jump - q/esc: cancel"
	if m.numBuf != "" {
		keys = "jump to row: " + m.numBuf + "  (enter to go, backspace to edit)"
	}

	var b strings.Builder
	if m.caption != "" {
		b.WriteString(dim.Render(m.caption))
		b.WriteString("\n")
	}
	b.WriteString(dim.Render(keys))
	return b.String()
}

// selectEntry launches the interactive entry selector and returns the chosen
// index (0-based, into the entries slice) along with ok=true. On cancel it
// returns (-1, false).
func selectEntry(caption string, entries []models.Entry) (int, bool, error) {
	m := newEntrySelectorModel(caption, entries)

	// Inline (no alt-screen): renders in normal terminal flow.
	p := tea.NewProgram(m)
	final, err := p.Run()
	if err != nil {
		return -1, false, err
	}

	em, ok := final.(entrySelectorModel)
	if !ok || em.chosen < 0 {
		return -1, false, nil
	}
	return em.chosen, true, nil
}
