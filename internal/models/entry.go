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
package models

import (
	"strings"
	"khronos/constants"

	"github.com/fatih/color"
)

type Entry struct {
	Uid           int64
	Project       string
	Note          string
	EntryDatetime string
	Duration      int64
	Properties    []Property
}

func NewEntry(uid int64, project string, note string, entryDatetime string) Entry {
	var e Entry = Entry{uid, project, note, entryDatetime, 0, make([]Property, 0)}
	return e
}

func (e *Entry) AddEntryProperty(name string, value string) {
	var found bool = false
	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, name) && strings.EqualFold(element.Value, value) {
			found = true
			break
		}
	}

	if !found {
		var property Property = NewProperty(e.Uid, name, value)
		e.Properties = append(e.Properties, property)
	}
}

func (e *Entry) UpdateEntryProperty(name string, value string) {
	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, name) {
			element.Value = value
			break
		}
	}
}

func (e *Entry) GetPropertiesAsString() string {
	var result string
	var firstTime bool = true

	for _, element := range e.Properties {
		if firstTime {
			firstTime = false
		} else {
			result += ", "
		}

		result += element.Name
		result += ":"
		result += element.Value
	}

	return result
}

func (e *Entry) GetTasksAsString() string {
	var result string

	// Count the number of Tasks.
	var taskCount = 0
	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.TASK) {
			taskCount += 1
		}
	}

	// Append any Tasks to the string.
	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.TASK) {
			result += element.Value
		}

		// Count backwards to add our separator.
		if taskCount > 1 {
			result += ", "
			taskCount -= 1
		}
	}

	return result
}

func (e *Entry) GetUrlAsString() string {
	var result string

	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.URL) {
			result = element.Value
			break
		}
	}

	return result
}

func (e *Entry) Dump(vertical bool, indent_amount int) string {
	var result string

	if strings.EqualFold(e.Project, constants.BREAK) {
		result = "Break Time"
	} else {
		// Add the project.
		if vertical {
			result = "\n"
		}
		result = result + strings.Repeat(constants.SPACE_CHARACTER, indent_amount) + color.YellowString("Project") + "[" + e.Project + "]"

		// Add the task(s).
		if vertical {
			result = result + "\n  "
		}
		result = result + strings.Repeat(constants.SPACE_CHARACTER, indent_amount) + color.YellowString(" Task") + "[" + e.GetTasksAsString() + "]"
	}

	// Add the note if there is one.
	if len(e.Note) > 0 {
		if vertical {
			result += "\n  "
		}
		result += strings.Repeat(constants.SPACE_CHARACTER, indent_amount) + color.YellowString(" Note") + "[" + e.Note + "]"
	}

	// Add the URL if there is one.
	var url = e.GetUrlAsString()
	if len(url) > 0 {
		if vertical {
			result += "\n   "
		}

		result += strings.Repeat(constants.SPACE_CHARACTER, indent_amount) + color.YellowString(" URL") + "[" + url + "]"
	}

	// Add the Date.
	if vertical {
		result += "\n  "
	}
	result += strings.Repeat(constants.SPACE_CHARACTER, indent_amount) + color.YellowString(" Date") + "[" + e.EntryDatetime + "]"

	return result
}
