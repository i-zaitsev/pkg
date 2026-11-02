// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import "fmt"

// Errorf formats an error like fmt.Errorf, prefixed with the command name.
func Errorf(c Info, format string, args ...any) error {
	return fmt.Errorf("%s: "+format, append([]any{c.Name()}, args...)...)
}

// ArgMissingError is returned when a command did not get a required argument.
type ArgMissingError struct {
	Name string
}

func (e *ArgMissingError) Error() string {
	return e.Name + " is required"
}

// ArgInvalidError is returned when a command argument has an invalid value.
type ArgInvalidError struct {
	Name  string
	Value string
}

func (e *ArgInvalidError) Error() string {
	return "invalid " + e.Name + ": " + e.Value
}
