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

package jira

import (
	"encoding/json"
	"fmt"
	"khronos/constants"
	"khronos/internal/models"
	"khronos/internal/util"
	"log"
	"os"
	"time"

	"github.com/agrison/go-commons-lang/stringUtils"
	"github.com/fatih/color"
)

var JiraPushUrl string
var JiraLogWorkToTicketUrl string
var JiraBrowseTicketUrl string

//  {
//    "started": "2025-11-04T09:30:00.000-0500",
//    "timeSpentSeconds": 3600,
//    "comment": {
//      "type": "doc",
//      "version": 1,
//      "content": [
//        { "type": "paragraph", "content": [ { "type": "text", "text": "Logged via Bruno." } ] }
//      ]
//    }
//  }

// Top‑level payload
type Payload struct {
	Started          string  `json:"started"`          // we’ll fill this with a formatted time string
	TimeSpentSeconds int64   `json:"timeSpentSeconds"` // seconds as an integer
	Comment          Comment `json:"comment"`
}

// Nested “comment” object
type Comment struct {
	Type    string  `json:"type"`    // always "doc"
	Version int     `json:"version"` // usually 1
	Content []Block `json:"content"` // slice of blocks (paragraphs, tables, …)
}

// One block inside the comment – here we only need a paragraph
type Block struct {
	Type    string     `json:"type"`    // e.g. "paragraph"
	Content []TextNode `json:"content"` // the actual text runs inside the paragraph
}

// Text node inside a paragraph
type TextNode struct {
	Type string `json:"type"` // always "text"
	Text string `json:"text"` // the visible string
}

func FormatJiraUrl(url string, ticket string) string {
	if stringUtils.IsBlank(ticket) {
		return constants.EMPTY
	} else {
		return fmt.Sprintf(url, ticket)
	}
}

type JiraRequest struct {
	EntryUid int64
	Ticket   string
	Payload  []byte
}

// JIRATimeLayout is the layout Jira expects.
// Note the three‑digit millisecond part (".000") and the offset without a colon.
const JIRATimeLayout = "2006-01-02T15:04:05.000-0700"

// UTCToJira takes a timestamp string that is guaranteed to be in UTC
// (e.g. "2025-11-04T16:00:28+00:00") and returns it formatted for Jira.
// If the input cannot be parsed, an error is returned.
func UTCToJira(utcStr string) (string, error) {
	// Parse the incoming string as RFC3339 (covers the "+00:00" suffix).
	t, err := time.Parse(time.RFC3339, utcStr)
	if err != nil {
		return "", fmt.Errorf("cannot parse %q as RFC3339 UTC timestamp: %w", utcStr, err)
	}

	// Ensure the time is in UTC – Parse already does this for "+00:00",
	// but calling UTC() makes the intent explicit and also covers inputs
	// that might omit the offset.
	utc := t.UTC()

	// Format using Jira’s layout.
	return utc.Format(JIRATimeLayout), nil
}

func JiraNewRequests(roundToMinutes int64, entries []models.Entry) (result []JiraRequest) {
	var requests []JiraRequest

	for _, entry := range entries {
		var ticket string = entry.GetTicketAsString()
		var pushed string = entry.GetPushedAsString()

		// Only look at records that have a ticket associated with them and the
		// pushed string is empty.
		if !stringUtils.IsBlank(ticket) && stringUtils.IsBlank(pushed) {
			var request JiraRequest
			request.EntryUid = entry.Uid
			request.Ticket = entry.GetTicketAsString()
			jiraTime, err := UTCToJira(entry.EntryDatetime)
			if err != nil {
				log.Fatalf("%s: Failed to convert %s to Jira time.\n", color.RedString(constants.FATAL_NORMAL_CASE),
					entry.EntryDatetime)
				os.Exit(1)
			}

			// Fill a fresh Payload struct.
			p := Payload{
				Started:          jiraTime,
				TimeSpentSeconds: util.Round(roundToMinutes, entry.Duration),
				Comment: Comment{
					Type:    "doc",
					Version: 1,
					Content: []Block{
						{
							Type: "paragraph",
							Content: []TextNode{
								{
									Type: "text",
									Text: entry.Note,
								},
							},
						},
					},
				},
			}

			bytes, err := json.Marshal(p)
			if err != nil {
				log.Fatalf("Failed to marshal payload %v", err)
				os.Exit(1)
			}

			request.Payload = bytes
			requests = append(requests, request)
		}
	}

	return requests
}
