// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"flag"
	"io"
)

// NewFlagSet returns a flag set named after the command that prints nothing
// and returns parse errors, so the tool decides what to print. On -h, Parse
// returns flag.ErrHelp.
func NewFlagSet(c Info) *flag.FlagSet {
	fs := flag.NewFlagSet(c.Name(), flag.ContinueOnError)
	fs.Usage = func() {}
	fs.SetOutput(io.Discard)
	return fs
}

// RequiredArg returns the first positional argument or an ArgMissingError.
func RequiredArg(c Info, fs *flag.FlagSet, name string) (string, error) {
	if fs.NArg() == 0 {
		return "", Errorf(c, "%w", &ArgMissingError{Name: name})
	}
	return fs.Arg(0), nil
}
