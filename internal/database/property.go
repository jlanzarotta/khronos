package database

import "database/sql"

type Property struct {
	Name  sql.NullString
	Value sql.NullString
}
