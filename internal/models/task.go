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
	"khronos/constants"
	"strings"
)

type Task struct {
	Task       string
	Duration   int64
	Properties []Property
}

func NewTask(task string) Task {
	var t Task = Task{task, 0, make([]Property, 0)}
	return t
}

func (t *Task) AddTaskProperty(name string, value string) {
	if len(value) > 0 {
		var found bool = false
		for _, element := range t.Properties {
			if strings.EqualFold(element.Name, name) && strings.EqualFold(element.Value, value) {
				found = true
				break
			}
		}

		if !found {
			var property Property = NewProperty(constants.UNKNOWN_UID, name, value)
			t.Properties = append(t.Properties, property)
		}
	}
}

func (e *Task) GetProjectsAsString() string {
	var result string

	// Count the number of Projects.
	var projectCount = 0
	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.PROJECT) {
			projectCount += 1
		}
	}

	// Append any Projects to the string.
	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.PROJECT) {
			result += element.Value
		}

		// Count backwards to add our separator.
		if projectCount > 1 {
			result += ", "
			projectCount -= 1
		}
	}

	return result
}

func (e *Task) GetTicketAsString() string {
	var result string

	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.TICKET) {
			result = element.Value
			break
		}
	}

	return result
}

func (e *Task) GetPushedAsString() string {
	var result string

	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.PUSHED) {
			result = element.Value
			break
		}
	}

	return result
}
