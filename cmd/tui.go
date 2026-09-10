/*
Copyright © 2018-2026 Jeff Lanzarotta
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
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// errNotATerminal is returned by the interactive selectors when stdin/stdout
// are not a real terminal. Bubble Tea cannot drive a selector in that case,
// and letting it try either errors out obscurely or hangs.
var errNotATerminal = errors.New("interactive selection requires a terminal")

// sgrReset clears every character attribute the terminal has active: color,
// bold, faint, reverse and the rest. It is the standard "back to normal" escape
// sequence and is safe to emit on any terminal.
const sgrReset = "\x1b[0m"

// runTUI runs a Bubble Tea program and cleans up the two pieces of terminal
// state that Bubble Tea leaves behind.
//
// Character attributes. Bubble Tea's teardown resets the alt screen, the mouse
// modes, bracketed paste and focus events, but it never emits an SGR reset. Its
// last painted frame is styled (lipgloss for the help line, go-pretty for the
// table), so whatever attribute was active at the end of that frame is still
// active when we return. Worse, standardRenderer.flush truncates each line with
// ansi.Truncate(line, r.width, ""), and on Windows r.width is measured once at
// program start and never updated, because there is no SIGWINCH. Under a
// multiplexer such as psmux, r.width can be wrong for the whole run, so the
// truncation cuts at the wrong column and can drop the trailing reset sequence
// off a styled line. The terminal is then left with a foreground attribute set.
// If that attribute happens to render close to the background, everything
// printed afterward is invisible, including the terminal's echo of what the
// user types at the next prompt. Emitting a reset here costs nothing and closes
// that hole.
//
// Console input mode. Bubble Tea already restores this twice on its way out
// (conInputReader.Close puts back the mode it captured, restoreInput puts back
// the pre-MakeRaw state), so the restore below is belt and braces. It matters
// only if a future version stops doing that, or on an exit path that skips
// Bubble Tea's own teardown.
func runTUI(p *tea.Program) (tea.Model, error) {
	fd := int(os.Stdin.Fd())

	state, err := term.GetState(fd)
	if err != nil {
		// Not a terminal, so there is nothing to save or restore.
		defer fmt.Print(sgrReset)
		return p.Run()
	}

	defer func() {
		_ = term.Restore(fd, state)
		fmt.Print(sgrReset)
	}()

	return p.Run()
}
