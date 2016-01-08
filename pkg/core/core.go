// Copyright 2014 Richard Lehane. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package core defines the Identifier, Identification, Recorder, Matcher and Result interfaces.
// The packages within core (bytematcher, containermatcher, etc.) provide a toolkit for building identifiers based on different signature formats (such as PRONOM).
package core

import (
	"errors"

	"github.com/richardlehane/siegfried/config"
	"github.com/richardlehane/siegfried/pkg/core/persist"
	"github.com/richardlehane/siegfried/pkg/core/priority"
	"github.com/richardlehane/siegfried/pkg/core/siegreader"
)

// Identifier describes the implementation of a signature format. E.g. there is a PRONOM identifier that implements the TNA's PRONOM format.
type Identifier interface {
	Recorder() Recorder  // return a recorder for matching
	Describe() [2]string // name and details
	Save(*persist.LoadSaver)
	String() string
	Recognise(MatcherType, int) (bool, string) // do you recognise this result index?
}

// Add additional identifier types here
const (
	Pronom byte = iota // Pronom is the TNA's PRONOM file format registry
)

// IdentifierLoader unmarshals an Identifer from a LoadSaver.
type IdentifierLoader func(*persist.LoadSaver) Identifier

var loaders = [8]IdentifierLoader{nil, nil, nil, nil, nil, nil, nil, nil}

// RegisterIdentifier allows external packages to add new IdentifierLoaders.
func RegisterIdentifier(id byte, l IdentifierLoader) {
	loaders[int(id)] = l
}

// LoadIdentifier applies the appropriate IdentifierLoader to load an identifier.
func LoadIdentifier(ls *persist.LoadSaver) Identifier {
	id := ls.LoadByte()
	l := loaders[int(id)]
	if l == nil {
		if ls.Err == nil {
			ls.Err = errors.New("bad identifier loader")
		}
		return nil
	}
	return l(ls)
}

// Recorder is a mutable object generated by an identifier. It records match results and sends identifications.
type Recorder interface {
	Record(MatcherType, Result) bool // Record results for each matcher; return true if match recorded (siegfried will iterate through the identifiers until an identifier returns true).
	Satisfied(MatcherType) bool      // Called before matcher starts - should we continue onto this matcher?
	Report(chan Identification)      // Now send results.
	Active(MatcherType)              // Instruct Recorder that can expect results of type MatcherType.
}

// Identification is sent by an identifier when a format matches
type Identification interface {
	String() string          // short text that is displayed to indicate the format match
	Known() bool             // does this identifier produce a match
	Warn() string            // identification warning message
	YAML() string            // long text that should be displayed to indicate the format match // TODO: 1.5 get rid of particular encodings.
	JSON() string            // JSON match response // TODO: 1.5 get rid of particular encodings.
	CSV() []string           // CSV match response // TODO: 1.5 get rid of particular encodings.
	Archive() config.Archive // does this format match any of the archive formats (zip, gzip, tar, warc, arc)
}

// Matcher does the matching (against the name/mime string or the byte stream) and sends results
type Matcher interface {
	Identify(string, *siegreader.Buffer) (chan Result, error) // Given a name/MIME string and bytes, identify the file.
	Add(SignatureSet, priority.List) (int, error)             // add a signature set, return total number of signatures in a matcher
	String() string
	Save(*persist.LoadSaver)
}

// MatcherType is used by recorders to tell which type of matcher has sent a result
type MatcherType int

// Add additional Matchers here
const (
	ExtensionMatcher MatcherType = iota
	MIMEMatcher
	ContainerMatcher
	ByteMatcher
	TextMatcher
	XMLMatcher
)

// SignatureSet is added to a matcher. It can take any form, depending on the matcher.
type SignatureSet interface{}

// Result is a raw hit that matchers pass on to Identifiers
type Result interface {
	Index() int
	Basis() string
}
