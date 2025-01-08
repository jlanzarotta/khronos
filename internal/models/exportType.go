package models

import "errors"

type ExportType string

const (
	ExportTypeCSV      ExportType = "csv"
	ExportTypeHTML     ExportType = "html"
	ExportTypeMarkDown ExportType = "md"
)

// String is used both by fmt.Print and by Cobra in help text.
func (e *ExportType) String() string {
	return string(*e)
}

// Set must have pointer receiver so it doesn't change the value of a copy.
func (e *ExportType) Set(v string) error {
	switch v {
	case string(ExportTypeCSV), string(ExportTypeHTML), string(ExportTypeMarkDown):
		*e = ExportType(v)
		return nil
	default:
		return errors.New(`must be one of "csv", "html", or "md"`)
	}
}

// Type is only used in help text
func (e *ExportType) Type() string {
	return "string"
}
