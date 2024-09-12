package models

import (
	"strings"
	"khronos/constants"
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

func (e *Task) GetUrlAsString() string {
	var result string

	for _, element := range e.Properties {
		if strings.EqualFold(element.Name, constants.URL) {
			result = element.Value
			break
		}
	}

	return result
}
