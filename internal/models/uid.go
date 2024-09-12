package models

type UID struct {
	Uid           int64
	EntryDatetime string
	Duration      int64
}

func NewUID(uid int64, entryDatetime string, duration int64) UID {
	var u UID = UID{uid, entryDatetime, duration}
	return u
}
