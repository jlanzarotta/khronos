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
package constants

const ARCHIVE string = "archive"
const ARCHIVE_DESCRIPTION = "Archive all nuked entries."
const ADD_LONG_DESCRIPTION = "Once you have completed a entry (project+task), use this command to add that newly completed task to the database with an optional note."
const ADD_SHORT_DESCRIPTION = "Add a completed entry"
const ADDING string = "Adding"
const AMEND_LONG_DESCRIPTION = "Amend is a convenient way to modify an entry, default is the last entry. It lets you modify the project, task, and/or datetime."
const AMEND_SHORT_DESCRIPTION = "Amend an entry"
const AMENDING string = "Amending"
const APPLICATION_NAME = "Khronos"
const APPLICATION_NAME_LOWERCASE = "khronos"
const AT string = "at"
const BACKEND_LONG_DESCRIPTION = "Open a sqlite shell to the database. REQUIRES sqlite standalone application in user path."
const BACKEND_SHORT_DESCRIPTION = "Open a sqlite shell to the database"
const BACKUP_LONG_DESCRIPTION = "Before making major changes to your database, make a backup."
const BACKUP_SHORT_DESCRIPTION = "Backup your database"
const BREAK string = "***break"
const BREAK_LONG_DESCRIPTION = "If you just spent time on break, use this command to add that time to the database."
const BREAK_SHORT_DESCRIPTION = "Add a break"
const CARBON_DATE_FORMAT string = "Y-m-d"
const CARBON_START_END_TIME_FORMAT string = "h:ia"
const CARBON_START_END_TIME_24H_FORMAT string = "H:i"
const COMPRESS string = "compress"
const COMPRESS_DESCRIPTION string = "Compress archive file in gzip format"
const CONFIGURATION_FILE string = ".khronos.yaml"
const CONFIGURE_LONG_DESCRIPTION = "Write out a YAML config file. Print path to config file."
const CONFIGURE_SHORT_DESCRIPTION = "Write out a YAML config file"
const COMMAND_BACKUP = "backup"
const COMMAND_BACKEND = "backend"
const COMMAND_CONVERT = "convert"
const COMMAND_HELLO = "hello"
const CONVERT_LONG_DESCRIPTION = "Convert all database entries to UTC."
const CONVERT_SHORT_DESCRIPTION = "Convert all database entries to UTC."
const CONVERTED = "converted"
const CSV = "csv"
const DATABASE_FILE string = "database_file"
const DATE_FORMAT string = "2006-01-02" // WTF golang?  Why this date format?
const DATE_NORMAL_CASE = "Date"
const DATE_TIME_NORMAL_CASE = "Date Time"
const DEBUG = "debug"
const DESCRIPTION = "description"
const DISPLAY_BY_DAY_TOTALS string = "display_by_day_totals"
const DISPLAY_TIME_IN_24H_FORMAT = "display_time_in_24h_format"
const DONE = "Done"
const DRY_RUN = "dry-run"
const DRY_RUN_DESCRIPTION = "Do not actually nuke anything, but show what potential would be nuked."
const DURATION_NORMAL_CASE = "Duration"
const EDIT_LONG_DESCRIPTION = "Open the Khronos configuration file in your default editor."
const EDIT_SHORT_DESCRIPTION = "Open the Khronos configuration file in your default editor"
const EMPTY string = ""
const EXPORT = "export"
const EXPORT_TYPE = "type"
const FATAL_NORMAL_CASE string = "Fatal"
const FAVORITE string = "favorite"
const FAVORITES string = "favorites"
const FLAG_CURRENT_WEEK = "current-week"
const FLAG_DATE = "date"
const FLAG_FROM = "from"
const FLAG_LAST_ENTRY = "last-entry"
const FLAG_NO_ROUNDING = "no-rounding"
const FLAG_PREVIOUS_WEEK = "previous-week"
const FLAG_TO = "to"
const FLAG_TODAY = "today"
const FLAG_YESTERDAY = "yesterday"
const HELLO string = "***hello"
const HELLO_LONG_DESCRIPTION = "In order to have khronos start tracking time is to run this command. It informs khronos that you would like it to start tracking your time."
const HELLO_SHORT_DESCRIPTION = "Start time tracking for the day"
const HELP_SHORT_DESCRIPTION = "Show help for command"
const INDENT_AMOUNT int = 4
const INFO_NORMAL_CASE string = "Info"
const MAY_BE_OVERRIDDEN_BY_GLOBAL_CONFIGURATION_SETTING = "* May be overridden by global configuration setting"
const NATURAL_LANGUAGE_DESCRIPTION string = "Natural Language Time, e.g., '18 minutes ago' or '9:45am'"
const NOTE string = "note"
const NOTE_DESCRIPTION string = "A note associated with this entry"
const NOTE_NORMAL_CASE = "Note"
const NUKE_ALL string = "all"
const NUKE_ALL_DESCRIPTION string = "Nuke ALL entries.  Use with extreme caution!!!"
const NUKE_LONG_DESCRIPTION = "As you continuously add completed entries, the database continues to grow unbounded. The nuke command allows you to manage the size of your database."
const NUKE_SHORT_DESCRIPTION = "Nukes entries from the sqlite database"
const PRINT_DATE_WIDTH int = 10
const PRINT_DURATION_WIDTH int = 38
const PRINT_NOTE_WIDTH int = 40
const PRINT_PROJECT_WIDTH int = 20
const PRINT_START_END_WIDTH int = 20
const PRINT_TASK_WIDTH int = 20
const PRIOR_YEARS string = "prior-years"
const PRIOR_YEARS_DESCRIPTION string = "Nuke all entries prior to the current year's entries."
const PROJECT string = "project"
const PROJECT_NORMAL_CASE = "Project"
const PROJECT_TASK = "project+task"
const PROJECTS_NORMAL_CASE = "Project(s)"
const REPORT_BY_DAY = "report.by_day"
const REPORT_BY_DAY_FORMAT string = "%-10s  %-38s  %-20s  %-20s"
const REPORT_BY_ENTRY = "report.by_entry"
const REPORT_BY_ENTRY_FORMAT string = "%-38s  %-10s  %-20s  %-20s  %-20s  %-40s"
const REPORT_BY_PROJECT = "report.by_project"
const REPORT_BY_PROJECT_FORMAT string = "%-38s  %-20s  %-20s"
const REPORT_BY_TASK = "report.by_task"
const REPORT_CARBON_TO_FROM_FORMAT string = "Y-M-d"
const REPORT_LONG_DESCRIPTION = "When you need to generate a report, default today, use this command."
const REPORT_SHORT_DESCRIPTION = "Generate a report"
const REQUIRE_NOTE string = "require_note"
const REQUIRE_NOTE_WITH_ASTERISK string = "require note*"
const ROOT_LONG_DESCRIPTION = "Khronos is a simple command line tool use to track the time you spend on a specific project and the one or more tasks associated with that project. It was inspired by the concepts of utt (Ultimate Time Tracker) and timetrap."
const ROOT_SHORT_DESCRIPTION = "Simple program used to track time spent on projects and tasks"
const ROUND_TO_MINUTES string = "round_to_minutes"
const SECONDS_PER_DAY = 86400
const SHOW_LONG_DESCRIPTION = "Show various information."
const SHOW_SHORT_DESCRIPTION = "Show various information"
const SPACE_CHARACTER string = " "
const SPLIT_WORK_FROM_BREAK_TIME string = "split_work_from_break_time"
const START_END_NORMAL_CASE = "Start-End"
const STATISTICS string = "statistics"
const STRETCH_LONG_DESCRIPTION = "Stretch the latest entry to 'now' or whatever is specified using the 'at' flag command."
const STRETCH_SHORT_DESCRIPTION = "Stretch the latest entry"
const TASK string = "task"
const TASK_DELIMITER string = "+"
const TASK_NORMAL_CASE = "Task"
const TASKS_NORMAL_CASE = "Task(s)"
const TOTAL = "TOTAL"
const UNKNOWN_UID int64 = -1
const URL = "url"
const URL_NORMAL_CASE = "URL"
const VERSION_LONG_DESCRIPTION = "Show the version information."
const VERSION_SHORT_DESCRIPTION = "Show the version information"
const WEB_LONG_DESCRIPTION = "Open the Khronos website in your default browser."
const WEB_SHORT_DESCRIPTION = "Open the Khronos website in your default browser"
const WEB_SITE string = "https://github.com/jlanzarotta/khronos/"
const WEEK_START string = "week_start"
