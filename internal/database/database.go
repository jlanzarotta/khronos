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
package database

import (
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"khronos/constants"
	"khronos/internal/models"

	"github.com/dromara/carbon/v2"
	"github.com/fatih/color"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

type Database struct {
	Filename string
	Conn     *sql.DB
	Context  context.Context
}

func New(filename string) *Database {
	// NOTE: Make sure '_foreign_keys=on' is set or 'DELETE ON CASCADE' will not work.
	conn, err := sql.Open("sqlite3", filename+"?_loc=UTC&_foreign_keys=on")
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	db := Database{}
	db.Filename = filename
	db.Conn = conn
	db.Context = context.Background()

	// Ping the database to ensure we are connected.
	err = db.Conn.Ping()
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	return &db
}

func (db *Database) Close() {
	db.Conn.Close()
}

func (db *Database) Create() {
	// Create the entry table.
	query := "CREATE TABLE entry (uid INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, project TEXT(128) NOT NULL, note TEXT(128), entry_datetime TEXT NOT NULL);"
	_, err := db.Conn.Exec(query)
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	// Create the property table.
	query = "CREATE TABLE property (entry_uid INTEGER NOT NULL, name TEXT(128) NOT NULL, value TEXT(128) NOT NULL, CONSTRAINT property_FK FOREIGN KEY (entry_uid) REFERENCES entry(uid) ON DELETE CASCADE);"
	_, err = db.Conn.Exec(query)
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}
}

func (db *Database) ConvertAllEntriesToUTC() {
	var s string = fmt.Sprintf("SELECT e.uid, e.project, e.note, e.entry_datetime FROM entry e ORDER BY e.uid;")
	results, err := db.Conn.Query(s)
	if err != nil {
		log.Fatalf("%s: Error trying to retrieve Entry records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	// Create a transaction.
	tx, err := db.Conn.BeginTx(db.Context, nil)
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	var query strings.Builder

	// Loop over all records, checking if it needs to be converted. If it does, convert it.
	for results.Next() {
		var entry models.Entry
		err = results.Scan(&entry.Uid, &entry.Project, &entry.Note, &entry.EntryDatetime)
		if err != nil {
			log.Fatalf("%s: Error trying to Scan Entries results into data structure. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		// Convert entry datetime to UTC.
		var utc carbon.Carbon = *carbon.Parse(entry.EntryDatetime).SetTimezone(carbon.UTC)

		// Update the record.
		query.Reset()
		query.WriteString("UPDATE entry")
		query.WriteString(" SET")
		query.WriteString(fmt.Sprintf(" entry_datetime = '%s'", utc.ToIso8601String()))
		query.WriteString(fmt.Sprintf(" WHERE uid = %d;", entry.Uid))

		log.Printf("Query[%s]\n", query.String())

		_, err = tx.ExecContext(db.Context, query.String())
		if err != nil {
			log.Fatalf("%s: Error trying to update entries records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			tx.Rollback()
			os.Exit(1)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}
}

func (db *Database) InsertNewEntry(entry models.Entry) {
	tx, err := db.Conn.BeginTx(db.Context, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	result, err := tx.ExecContext(db.Context, "INSERT INTO entry (uid, project, note, entry_datetime) VALUES (?, ?, ?, ?);", nil, entry.Project, entry.Note, entry.EntryDatetime)
	if err != nil {
		rollBackError := tx.Rollback()
		if rollBackError != nil {
			log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), rollBackError.Error())
			os.Exit(1)
		}

		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	// Now that the record was inserted, get the last inserted id... in our case it it the UID.
	uid, err := result.LastInsertId()
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	// Now insert each of the properties for this entry.
	for _, v := range entry.Properties {
		_, err := tx.ExecContext(db.Context, "INSERT INTO property (entry_uid, name, value) VALUES (?, ?, ?);", uid, v.Name, v.Value)
		if err != nil {
			rollBackError := tx.Rollback()
			if rollBackError != nil {
				log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), rollBackError.Error())
				os.Exit(1)
			}

			log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}
	}

	err = tx.Commit()
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}
}

func (db *Database) GetDistinctUIDs(start carbon.Carbon, end carbon.Carbon) []DistinctUID {
	results, err := db.Conn.Query(`
		SELECT DISTINCT
			e.uid, e.project, e.entry_datetime
		FROM entry e
		WHERE e.entry_datetime BETWEEN ? AND ?
		ORDER BY entry_datetime;
		`, start.ToIso8601String(), end.ToIso8601String(),
	)

	if err != nil {
		log.Fatalf("%s: Error trying to retrieve distinct uids. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	records := []DistinctUID{}
	for results.Next() {
		var distinctUID DistinctUID
		err = results.Scan(&distinctUID.Uid, &distinctUID.Project, &distinctUID.EntryDatetime)
		if err != nil {
			log.Fatalf("%s: Error trying to scan results into DistinctUID data structure. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		records = append(records, distinctUID)
	}

	return records
}

func (db *Database) GetProperties(entryUid int64) []Property {
	var s string = fmt.Sprintf("SELECT p.name, p.value FROM property p WHERE p.entry_uid = %d;", entryUid)

	results, err := db.Conn.Query(s)
	if err != nil {
		log.Fatalf("%s: Error trying to retrieve Property records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	records := []Property{}
	for results.Next() {
		var property Property
		err = results.Scan(&property.Name, &property.Value)
		if err != nil {
			log.Fatalf("%s: Error trying to Scan Property results into data structure. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		records = append(records, property)
	}

	return records
}

func (db *Database) GetEntries(in string) []models.Entry {
	var s string = fmt.Sprintf("SELECT e.uid, e.project, e.note, e.entry_datetime FROM entry e WHERE e.uid IN (%s) ORDER BY entry_datetime;", in)

	results, err := db.Conn.Query(s)
	if err != nil {
		log.Fatalf("%s: Error trying to retrieve Entry records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	records := []models.Entry{}
	for results.Next() {
		var entry models.Entry
		err = results.Scan(&entry.Uid, &entry.Project, &entry.Note, &entry.EntryDatetime)
		if err != nil {
			log.Fatalf("%s: Error trying to Scan Entries results into data structure. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		var properties []Property = db.GetProperties(entry.Uid)
		for _, p := range properties {
			entry.AddEntryProperty(p.Name.String, p.Value.String)
		}

		records = append(records, entry)
	}

	return records
}

func (db *Database) GetEntriesForToday(start carbon.Carbon, end carbon.Carbon) []models.Entry {
	var s string = fmt.Sprintf("SELECT e.uid, e.project, e.note, e.entry_datetime FROM entry e WHERE e.entry_datetime between '%s' AND '%s' ORDER BY entry_datetime;", start.ToIso8601String(), end.ToIso8601String())

	results, err := db.Conn.Query(s)
	if err != nil {
		log.Fatalf("%s: Error trying to retrieve Entry records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	records := []Entry{}
	for results.Next() {
		var entry Entry
		err = results.Scan(&entry.Uid, &entry.Project, &entry.Note, &entry.EntryDatetime)
		if err != nil {
			log.Fatalf("%s: Error trying to Scan Entries results into data structure. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		records = append(records, entry)
	}

	var entries = []models.Entry{}
	for _, e := range records {
		var entry models.Entry = models.NewEntry(e.Uid, e.Project, e.Note.String, e.EntryDatetime)
		var properties []Property = db.GetProperties(entry.Uid)
		for _, p := range properties {
			entry.AddEntryProperty(p.Name.String, p.Value.String)
		}
		entries = append(entries, entry)
	}

	return entries
}

func (db *Database) getEntry(uid int64) models.Entry {
	var s string = fmt.Sprintf("SELECT e.uid, e.project, e.note, e.entry_datetime FROM entry e WHERE e.uid = %d ORDER BY entry_datetime;", uid)
	results, err := db.Conn.QueryContext(db.Context, s)
	if err != nil {
		log.Fatalf("%s: Error trying to retrieve Uid's Entry records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	records := []Entry{}
	for results.Next() {
		var entry Entry
		err = results.Scan(&entry.Uid, &entry.Project, &entry.Note, &entry.EntryDatetime)
		if err != nil {
			log.Fatalf("%s: Error trying to Scan Uid's Entries results into data structure. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		records = append(records, entry)
	}

	results.Close()

	var entry models.Entry
	for i, e := range records {
		if i == 0 {
			entry = models.NewEntry(e.Uid, e.Project, e.Note.String, e.EntryDatetime)
			if strings.EqualFold(e.Project, constants.HELLO) {
				break
			}
		}

		var properties []Property = db.GetProperties(entry.Uid)
		for _, p := range properties {
			entry.AddEntryProperty(p.Name.String, p.Value.String)
		}
	}

	return entry
}

func (db *Database) GetFirstEntry() models.Entry {
	result, err := db.Conn.QueryContext(db.Context, "SELECT e.uid FROM entry e ORDER BY entry_datetime LIMIT 1;")
	if err != nil {
		log.Fatalf("%s: Error trying to retrieve first Uid. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	var firstUid int64
	result.Next()
	err = result.Scan(&firstUid)
	if err != nil {
		log.Fatalf("%s: Error trying to Scan first Uid into data structure. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	result.Close()

	// Create entry from the data from the database.
	var entry models.Entry = db.getEntry(firstUid)

	return entry
}

func (db *Database) GetLastEntry() models.Entry {
	result, err := db.Conn.QueryContext(db.Context, "SELECT e.uid FROM entry e ORDER BY entry_datetime DESC LIMIT 1;")
	if err != nil {
		log.Fatalf("%s: Error trying to retrieve last Uid. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	var lastUid int64
	result.Next()
	err = result.Scan(&lastUid)
	if err != nil {
		return models.NewEntry(constants.UNKNOWN_UID, constants.EMPTY, constants.EMPTY, constants.EMPTY)
	}

	result.Close()

	// Create entry from the data from the database.
	var entry models.Entry = db.getEntry(lastUid)

	return entry
}

func (db *Database) GetCountEntries() int64 {
	result, err := db.Conn.QueryContext(db.Context, "SELECT COUNT(*) FROM entry;")
	if err != nil {
		log.Fatalf("%s: Error trying to retrieve count of entries. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	var count int64
	result.Next()
	err = result.Scan(&count)
	if err != nil {
		log.Fatalf("%s: Error trying to Scan count. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	result.Close()

	return count
}

func CreateArchiveFile(entryWithProperty []EntryWithProperty, compress bool) {
	// Create our unique archive file.
	var filename = constants.APPLICATION_NAME_LOWERCASE + "_archive_" + carbon.Now().ToShortDateTimeString()
	archiveFile, err := os.Create(filename + ".csv")
	if err != nil {
		log.Fatalf("%s: Error trying to create archive file. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}
	defer archiveFile.Close()

	_, err = archiveFile.WriteString("uid,project,note,entry_date_time,name,value\n")
	if err != nil {
		log.Fatalf("%s: Error writing to archive file. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	// Write each record to the archive file.
	for _, ewp := range entryWithProperty {
		_, err = archiveFile.WriteString(fmt.Sprintf("%d, %s, %s, %s, %s, %s\n", ewp.Uid, ewp.Project, ewp.Note.String, ewp.EntryDatetime, ewp.Name.String, ewp.Value.String))
		if err != nil {
			log.Fatalf("%s: Error writing to archive file. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}
	}

	// Flush the archive file to disk so it can be closed.
	archiveFile.Sync()
	archiveFile.Close()

	if compress {
		// Open the original archive file.
		originalFile, err := os.Open(filename + ".csv")
		if err != nil {
			log.Fatalf("%s: Error opening original archive file. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}
		defer originalFile.Close()

		// Create a new gzipped file.
		gzippedFile, err := os.Create(filename + ".csv.gz")
		if err != nil {
			log.Fatalf("%s: Error creating gzip file. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}
		defer gzippedFile.Close()

		// Create a new gzip writer.
		gzipWriter := gzip.NewWriter(gzippedFile)
		defer gzipWriter.Close()

		// Copy the contents of the original file to the gzip writer.
		_, err = io.Copy(gzipWriter, originalFile)
		if err != nil {
			log.Fatalf("%s: Error writing to gzip file. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		// Flush the gzip writer to ensure all data is written.
		gzipWriter.Flush()

		// Delete the original archive file.
		originalFile.Close()
		err = os.Remove(filename + ".csv")
		if err != nil {
			log.Fatalf("%s: Error deleting original archive file after compressing. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}
	}
}

func (db *Database) NukePriorYearsEntries(dryRun bool, year int, archive bool, compress bool) int64 {
	var count int64 = 0
	var query strings.Builder

	if !dryRun {
		var archiveRecords = []EntryWithProperty{}
		var uuids string = constants.EMPTY
		var s string = fmt.Sprintf("%s != '%d';", "SELECT e.uid, e.project, e.note, e.entry_datetime, p.name, p.value FROM entry e JOIN property p ON p.entry_uid = e.uid WHERE strftime('%Y', e.entry_datetime)", year)
		results, err := db.Conn.Query(s)
		if err != nil {
			log.Fatalf("%s: Error trying to retrieve Entry records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		// Loop over our results.
		var lastUid int64 = -1
		for results.Next() {
			var ewp EntryWithProperty
			err = results.Scan(&ewp.Uid, &ewp.Project, &ewp.Note, &ewp.EntryDatetime, &ewp.Name, &ewp.Value)
			if err != nil {
				log.Fatalf("%s: Error trying to Scan EntriesWithProperty results into data structure. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
				os.Exit(1)
			}

			// Save the uid of the record being deleted.
			if strings.EqualFold(uuids, constants.EMPTY) {
				uuids = fmt.Sprintf("%d", ewp.Uid)
			} else {
				if lastUid != ewp.Uid {
					uuids = fmt.Sprintf("%s, %d", uuids, ewp.Uid)
				}
			}
			lastUid = ewp.Uid

			// If the user wants an archive of the records being deleted, save
			// the information in a collection.
			if archive {
				archiveRecords = append(archiveRecords, ewp)
			}
		}

		results.Close()

		// If the user wants an archive, write the archived collection to a
		// file.
		if archive {
			CreateArchiveFile(archiveRecords, compress)
		}

		// Create a transaction.
		tx, err := db.Conn.BeginTx(db.Context, nil)
		if err != nil {
			log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		// Via the transaction, delete all the entry and associated property records.
		query.WriteString(fmt.Sprintf("DELETE FROM entry WHERE uid IN (%s)", uuids))
		_, err = tx.ExecContext(db.Context, query.String())
		if err != nil {
			log.Fatalf("%s: Error trying to delete all entry records prior to %d. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), year, err.Error())
			tx.Rollback()
			os.Exit(1)
		} else {
			query.Reset()
			query.WriteString(fmt.Sprintf("DELETE FROM property WHERE entry_uid IN (%s)", uuids))
			_, err = tx.ExecContext(db.Context, query.String())
			if err != nil {
				log.Fatalf("%s: Error trying to delete all property records prior to %d. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), year, err.Error())
				tx.Rollback()
				os.Exit(1)
			}

			// Commit our changes to the database.
			err = tx.Commit()
			if err != nil {
				log.Fatalf("%s: Error committing transaction. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
				os.Exit(1)
			}
		}
	} else {
		query.WriteString(fmt.Sprintf("%s != '%d';", "SELECT COUNT(*) FROM entry WHERE strftime('%Y', entry_datetime)", year))
		result, err := db.Conn.QueryContext(db.Context, query.String())
		if err != nil {
			log.Fatalf("%s: Error trying to retrieve count of entries. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		result.Next()
		err = result.Scan(&count)
		if err != nil {
			log.Fatalf("%s: Error trying to Scan count. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		result.Close()
	}

	return count
}

func (db *Database) NukeAllEntries(dryRun bool, archive bool, compress bool) int64 {
	var count int64 = 0

	if !dryRun {
		var archiveRecords = []EntryWithProperty{}
		var s string = "SELECT e.uid, e.project, e.note, e.entry_datetime, p.name, p.value FROM entry e JOIN property p ON p.entry_uid = e.uid ORDER BY e.uid"
		results, err := db.Conn.Query(s)
		if err != nil {
			log.Fatalf("%s: Error trying to retrieve Entry records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		// Loop over our results.
		for results.Next() {
			var ewp EntryWithProperty
			err = results.Scan(&ewp.Uid, &ewp.Project, &ewp.Note, &ewp.EntryDatetime, &ewp.Name, &ewp.Value)
			if err != nil {
				log.Fatalf("%s: Error trying to Scan EntriesWithProperty results into data structure. %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
				os.Exit(1)
			}

			// If the user wants an archive of the records being deleted, save
			// the information in a collection.
			if archive {
				archiveRecords = append(archiveRecords, ewp)
			}
		}

		// If the user wants an archive, write the archived collection to a
		// file.
		if archive {
			CreateArchiveFile(archiveRecords, compress)
		}

		// Create a transaction.
		tx, err := db.Conn.BeginTx(db.Context, nil)
		if err != nil {
			log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}

		// Via the transaction, delete all the entry and associated property records.
		_, err = tx.ExecContext(db.Context, "DELETE FROM entry;")
		if err != nil {
			log.Fatalf("%s: Error trying to delete all entry records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			tx.Rollback()
			os.Exit(1)
		} else {
			_, err = tx.ExecContext(db.Context, "DELETE FROM property;")
			if err != nil {
				log.Fatalf("%s: Error trying to delete all property records. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
				tx.Rollback()
				os.Exit(1)
			} else {
				err = tx.Commit()
				if err != nil {
					log.Fatalf("%s: Error committing transaction. %s.\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
					os.Exit(1)
				}
			}
		}
	} else {
		count = db.GetCountEntries()
		log.Printf("%d entries would have been nuked.", count)
	}

	return count
}

func (db *Database) UpdateEntry(entry models.Entry) {
	var previous bool = false
	var query strings.Builder

	// Update the Entry.
	query.WriteString("UPDATE entry")
	query.WriteString(" SET")

	if entry.Project != constants.EMPTY {
		query.WriteString(fmt.Sprintf(" project = '%s'", entry.Project))
		previous = true
	}

	if len(entry.Note) > 0 {
		if previous {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf(" note = '%s'", entry.Note))
		previous = true
	}

	if entry.EntryDatetime != constants.EMPTY {
		if previous {
			query.WriteString(", ")
		}
		query.WriteString(fmt.Sprintf(" entry_datetime = '%s'", entry.EntryDatetime))
	}

	query.WriteString(fmt.Sprintf(" WHERE uid = %d;", entry.Uid))

	if viper.GetBool(constants.DEBUG) {
		log.Printf("Query[%s]\n", query.String())
	}

	// Execute the update.
	_, err := db.Conn.ExecContext(db.Context, query.String())
	if err != nil {
		log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
		os.Exit(1)
	}

	// Update the TASK property if one exists.
	var task = entry.GetTasksAsString()
	if len(task) > 0 {
		query.Reset()
		query.WriteString("UPDATE property")
		query.WriteString(" SET")
		query.WriteString(fmt.Sprintf(" value = '%s'", task))
		query.WriteString(fmt.Sprintf(" WHERE entry_uid = %d and name = '%s';", entry.Uid, constants.TASK))

		if viper.GetBool(constants.DEBUG) {
			log.Printf("Query[%s]\n", query.String())
		}

		// Execute the update.
		_, err = db.Conn.ExecContext(db.Context, query.String())
		if err != nil {
			log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}
	}

	// Update the URL property if one exists.
	var url = entry.GetUrlAsString()
	if len(url) > 0 {
		query.Reset()
		query.WriteString("UPDATE property")
		query.WriteString(" SET")
		query.WriteString(fmt.Sprintf(" value = '%s'", url))
		query.WriteString(fmt.Sprintf(" WHERE entry_uid = %d and name = '%s';", entry.Uid, constants.URL))

		if viper.GetBool(constants.DEBUG) {
			log.Printf("Query[%s]\n", query.String())
		}

		// Execute the update.
		_, err = db.Conn.ExecContext(db.Context, query.String())
		if err != nil {
			log.Fatalf("%s: %s\n", color.RedString(constants.FATAL_NORMAL_CASE), err.Error())
			os.Exit(1)
		}
	}
}
