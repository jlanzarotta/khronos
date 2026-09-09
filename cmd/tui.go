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
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/term"
)

// errNotATerminal is returned by the interactive selectors when stdin/stdout
// are not a real terminal. Bubble Tea cannot drive a selector in that case,
// and letting it try either errors out obscurely or hangs.
var errNotATerminal = errors.New("interactive selection requires a terminal")

// runTUI runs a Bubble Tea program and restores the terminal state that was in
// effect before it started.
//
// This is belt and braces, not a fix for anything. Bubble Tea already restores
// the console input mode twice on its way out: conInputReader.Close puts back
// the mode it captured, and restoreInput puts back the pre-MakeRaw state. The
// restore here only matters if a future Bubble Tea version stops doing that, or
// if a program exits on a path that skips its own teardown.
//
// Do not read this function as the reason the "typing shows nothing" bug went
// away. That bug was the inline renderer erasing the prompt line, and the fix
// for it is tea.WithAltScreen on the selectors. See the comment there.
func runTUI(p *tea.Program) (tea.Model, error) {
	fd := int(os.Stdin.Fd())

	state, err := term.GetState(fd)
	if err != nil {
		// Not a terminal, so there is nothing to save or restore.
		return p.Run()
	}

	defer func() {
		_ = term.Restore(fd, state)
	}()

	return p.Run()
}
