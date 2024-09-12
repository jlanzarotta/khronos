package models

type Property struct {
	EntryUid int64
	Name     string
	Value    string
}

func NewProperty(entryUid int64, name string, value string) Property {
	var p Property = Property{entryUid, name, value}
	return p
}
