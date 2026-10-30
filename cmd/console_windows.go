//go:build windows

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
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

// cookedInput is the console input mode a line prompt needs: characters are
// echoed, input is assembled into lines, and Ctrl+C raises a signal.
const cookedInput = windows.ENABLE_ECHO_INPUT | windows.ENABLE_LINE_INPUT | windows.ENABLE_PROCESSED_INPUT

// ensureCookedInput turns the cooked input bits back on if something cleared
// them. The report is deliberately not behind a flag: a repair is rare, and it
// is the only evidence that a prompt would otherwise have been dead.
func ensureCookedInput() {
	handle := windows.Handle(os.Stdin.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return
	}

	if mode&cookedInput == cookedInput {
		return
	}

	repaired := mode | cookedInput
	if err := windows.SetConsoleMode(handle, repaired); err != nil {
		fmt.Fprintf(os.Stderr, "khronos: console input mode 0x%04X is wrong and could not be repaired: %v\n", mode, err)
		return
	}

	fmt.Fprintf(os.Stderr, "khronos: repaired console input mode 0x%04X -> 0x%04X\n", mode, repaired)
}
