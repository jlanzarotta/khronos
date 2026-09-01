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

// runTUI runs a Bubble Tea program and guarantees that the terminal is back in
// cooked mode (echo on, line input on) before control returns to the caller.
//
// Bubble Tea puts the console into raw mode and reads stdin from its own
// goroutine. When p.Run() returns, restoring the console mode and tearing down
// that reader race with whatever we do next. If our own prompts (readLine on
// stdinReader) start before the restore has fully landed, the terminal is still
// in raw mode: keystrokes are accepted but never echoed, and Enter arrives as a
// bare '\r' instead of '\n'. Snapshotting the terminal state here and restoring
// it unconditionally removes the dependency on Bubble Tea's teardown timing.
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
