/*
Copyright 2015 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package ci defines the internal representation of a continuous integration reports.
package ci

import (
	"encoding/json"
	"github.com/promet/git-appraise/repository"
	"sort"
	"strconv"
)

const (
	// Ref defines the git-notes ref that we expect to contain CI reports.
	Ref = "refs/notes/devtools/ci"

	// StatusSuccess is the status string representing that a build and/or test passed.
	StatusSuccess = "success"
	// StatusFailure is the status string representing that a build and/or test failed.
	StatusFailure = "failure"

	// FormatVersion defines the latest version of the request format supported by the tool.
	FormatVersion = 0
)

// Report represents a build/test status report generated by a continuous integration tool.
//
// Every field is optional.
type Report struct {
	Timestamp string `json:"timestamp,omitempty"`
	URL       string `json:"url,omitempty"`
	Status    string `json:"status,omitempty"`
	Agent     string `json:"agent,omitempty"`
	// Version represents the version of the metadata format.
	Version int `json:"v,omitempty"`
}

// Parse parses a CI report from a git note.
func Parse(note repository.Note) (Report, error) {
	bytes := []byte(note)
	var report Report
	err := json.Unmarshal(bytes, &report)
	return report, err
}

// GetLatestCIReport takes the collection of reports and returns the one with the most recent timestamp.
func GetLatestCIReport(reports []Report) (*Report, error) {
	timestampReportMap := make(map[int]*Report)
	var timestamps []int

	for _, report := range reports {
		timestamp, err := strconv.Atoi(report.Timestamp)
		if err != nil {
			return nil, err
		}
		timestamps = append(timestamps, timestamp)
		timestampReportMap[timestamp] = &report
	}
	if len(timestamps) == 0 {
		return nil, nil
	}
	sort.Sort(sort.Reverse(sort.IntSlice(timestamps)))
	return timestampReportMap[timestamps[0]], nil
}

// ParseAllValid takes collection of git notes and tries to parse a CI report
// from each one. Any notes that are not valid CI reports get ignored, as we
// expect the git notes to be a heterogenous list, with only some of them
// being valid CI status reports.
func ParseAllValid(notes []repository.Note) []Report {
	var reports []Report
	for _, note := range notes {
		report, err := Parse(note)
		if err == nil && report.Version == FormatVersion {
			if report.Status == "" || report.Status == StatusSuccess || report.Status == StatusFailure {
				reports = append(reports, report)
			}
		}
	}
	return reports
}
