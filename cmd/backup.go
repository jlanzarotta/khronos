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
	"fmt"
	"io"
	"khronos/constants"
	"log"
	"os"
	"path/filepath"

	"github.com/dromara/carbon/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var BUFFERSIZE int64 = 1024

var backupCmd = &cobra.Command{
	Use:     constants.COMMAND_BACKUP,
	Args:    cobra.ExactArgs(0),
	Short:   constants.BACKUP_SHORT_DESCRIPTION,
	Long:    constants.BACKUP_LONG_DESCRIPTION,
	Run: func(cmd *cobra.Command, args []string) {
		runBackup(args)
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}

func runBackup(_ []string) {
	var databaseFilename string = viper.GetString(constants.DATABASE_FILE)
	dir, file := filepath.Split(databaseFilename)
	var backupFilename = filepath.Join(dir, file + "-backup_" + carbon.Now(carbon.Local).ToShortDateTimeString())
	log.Printf("Backing up %s to %s...", databaseFilename, backupFilename)
	err := backup(databaseFilename, backupFilename, BUFFERSIZE)
	if err != nil {
		log.Fatalf("%s: Error trying to backup database file. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err)
		os.Exit(1)
	}

	log.Printf("%s.\n", color.GreenString(constants.DONE))
	os.Exit(0)
}

func backup(src, dst string, BUFFERSIZE int64) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("backup file %s already exists", dst)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err != nil {
		panic(err)
	}

	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}
	return err
}