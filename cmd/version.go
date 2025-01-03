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
	"khronos/constants"
	"log"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var BuildVersion string
var BuildDateTime string

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version information",
	Long:  "Show the version information.",
	Run: func(cmd *cobra.Command, args []string) {
		runVersion(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) {
	log.Printf("%s %s\nBorn on: %s\n" +
"------------------------------------------------------------------------------\n" +
"BSD 3-Clause License\n\n" +
"Copyright (c) 2018-%s, Jeff Lanzarotta All rights reserved.\n" +
"\n" +
"Redistribution and use in source and binary forms, with or without\n" +
"modification, are permitted provided that the following conditions are met:\n" +
"\n" +
"1. Redistributions of source code must retain the above copyright notice, this\n" +
"   list of conditions and the following disclaimer.\n" +
"\n" +
"2. Redistributions in binary form must reproduce the above copyright notice,\n" +
"   this list of conditions and the following disclaimer in the documentation\n" +
"   and/or other materials provided with the distribution.\n" +
"\n" +
"3. Neither the name of the copyright holder nor the names of its\n" +
"   contributors may be used to endorse or promote products derived from\n" +
"   this software without specific prior written permission.\n" +
"\n" +
"THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS \"AS IS\"\n" +
"AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE\n" +
"IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE\n" +
"DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE\n" +
"FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL\n" +
"DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR\n" +
"SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER\n" +
"CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,\n" +
"OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE\n" +
"OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.\n" +
"------------------------------------------------------------------------------\n", 
	color.YellowString(constants.APPLICATION_NAME), BuildVersion, BuildDateTime, time.Now().Format("2006"))
}
