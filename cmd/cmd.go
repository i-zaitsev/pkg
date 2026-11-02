// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package cmd defines subcommands of command-line tools and helpers for them.
//
// A subcommand implements Command: it describes itself, parses its own
// arguments, and runs in a tool-specific environment E. The tool owns the
// registry, dispatch, global flags, and help text, and composes them from
// the free functions here.
package cmd

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"
	"text/tabwriter"
)

// Info describes a command: its name, a short description,
// and the positional arguments it expects.
type Info interface {
	Name() string
	Desc() string
	Args() string
}

// Command is a subcommand that parses its arguments and runs in env E.
type Command[E any] interface {
	Info
	Parse(args []string) error
	Run(env E) error
}

// Register adds cmds to the caller's registry, keyed by name.
// It fails on an empty name or a name that is already taken, in reg or
// within cmds, and then leaves reg unchanged.
func Register[E any](reg map[string]Command[E], cmds ...Command[E]) error {
	seen := make(map[string]bool, len(cmds))
	for _, c := range cmds {
		name := c.Name()
		if name == "" {
			return errors.New("cmd: empty command name")
		}
		if _, ok := reg[name]; ok || seen[name] {
			return fmt.Errorf("cmd: duplicate command %q", name)
		}
		seen[name] = true
	}
	for _, c := range cmds {
		reg[c.Name()] = c
	}
	return nil
}

// Usage lists the commands sorted by name, one per line,
// with names, arguments, and descriptions aligned in columns.
func Usage[C Info](cmds ...C) string {
	sorted := slices.SortedFunc(slices.Values(cmds), func(a, b C) int {
		return cmp.Compare(a.Name(), b.Name())
	})

	var sb strings.Builder
	tw := tabwriter.NewWriter(&sb, 0, 4, 2, ' ', 0)
	for _, c := range sorted {
		_, _ = fmt.Fprintf(tw, "\t%s\t%s\t%s\n", c.Name(), c.Args(), c.Desc())
	}
	_ = tw.Flush()
	return sb.String()
}
