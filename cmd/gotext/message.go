// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

// TODO: these definitions should be moved to a package so that the can be used
// by other tools.

// The file contains the structures used to define translations of a certain
// messages.
//
// A translation may have multiple translations strings, or messages, depending
// on the feature values of the various arguments. For instance, consider
// a hypothetical translation from English to English, where the source defines
// the format string "%d file(s) remaining".
// See the examples directory for examples of extracted messages.

// A Message describes a message to be translated.
type Message struct {
	// Key contains a list of identifiers for the message. If this list is empty
	// the message itself is used as the key.
	Key         []string `json:"key,omitempty"`
	Meaning     string   `json:"meaning,omitempty"`
	Message     Text     `json:"message"`
	Translation *Text    `json:"translation,omitempty"`

	Comment           string `json:"comment,omitempty"`
	TranslatorComment string `json:"translatorComment,omitempty"`

	// TODO: have a separate placeholder list, mapping placeholders
	// to arguments or constant strings.
	// TODO: default placeholder syntax is {foo}. Allow alternatives
	// like `foo`.

	Args []Argument `json:"args,omitempty"`

	// Extraction information.
	Position string `json:"position,omitempty"` // filePosition:line
}

// An Argument contains information about the arguments passed to a message.
type Argument struct {
	ID string `json:"id"` // An int for printf-style calls, but could be a string.
	// Argument position for printf-style format strings. ArgNum corresponds to
	// the number that should be used for explicit argument indexes (e.g.
	// "%[1]d").
	ArgNum int      `json:"argNum,omitempty"`
	Format []string `json:"format,omitempty"`

	Type           string `json:"type"`
	UnderlyingType string `json:"underlyingType"`
	Expr           string `json:"expr"`
	Value          string `json:"value,omitempty"`
	Comment        string `json:"comment,omitempty"`
	Position       string `json:"position,omitempty"`

	// Features contains the features that are available for the implementation
	// of this argument.
	Features []Feature `json:"features,omitempty"`
}

// Feature holds information about a feature that can be implemented by
// an Argument.
type Feature struct {
	Type string `json:"type"` // Right now this is only gender and plural.

	// TODO: possible values and examples for the language under consideration.

}

// Text defines a message to be displayed.
type Text struct {
	// Msg and Select contains the message to be displayed. Within a Text value
	// either Msg or Select is defined.
	Msg    string  `json:"msg,omitempty"`
	Select *Select `json:"select,omitempty"`
	// Var defines a map of variables that may be substituted in the selected
	// message.
	Var map[string]Text `json:"var,omitempty"`
	// Example contains an example message formatted with default values.
	Example string `json:"example,omitempty"`
}

// Select selects a Text based on the feature value associated with a feature of
// a certain argument.
type Select struct {
	Feature string          `json:"feature"` // Name of variable or Feature type
	Arg     interface{}     `json:"arg"`     // The argument ID.
	Cases   map[string]Text `json:"cases"`
}
