// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd_test

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"

	"github.com/i-zaitsev/pkg/cmd"
)

type Env struct{ out io.Writer }

type greet struct {
	name  string
	shout bool
}

func (g *greet) Name() string { return "greet" }
func (g *greet) Desc() string { return "Print a greeting" }
func (g *greet) Args() string { return "<name>" }

func (g *greet) Parse(args []string) error {
	fs := cmd.NewFlagSet(g)
	fs.BoolVar(&g.shout, "shout", false, "greet loudly")
	if err := fs.Parse(args); err != nil {
		return cmd.Errorf(g, "%w", err)
	}
	var err error
	g.name, err = cmd.RequiredArg(g, fs, "name")
	return err
}

func (g *greet) Run(e *Env) error {
	greeting := "hello"
	if g.shout {
		greeting = "HELLO"
	}
	_, err := fmt.Fprintf(e.out, "%s, %s\n", greeting, g.name)
	return err
}

type version struct{}

func (version) Name() string         { return "version" }
func (version) Desc() string         { return "Print the version" }
func (version) Args() string         { return "" }
func (version) Parse([]string) error { return nil }

func (version) Run(e *Env) error {
	_, err := fmt.Fprintln(e.out, "0.1.0")
	return err
}

func Example() {
	reg := map[string]cmd.Command[*Env]{}
	if err := cmd.Register(reg, &greet{}); err != nil {
		fmt.Println(err)
	}
	// Commands can be added from anywhere that has the registry.
	if err := cmd.Register(reg, version{}); err != nil {
		fmt.Println(err)
	}

	env := &Env{out: os.Stdout}
	run := func(args ...string) {
		c, ok := reg[args[0]]
		if !ok {
			fmt.Print("Commands:\n", cmd.Usage(slices.Collect(maps.Values(reg))...))
			return
		}
		err := c.Parse(args[1:])
		if errors.Is(err, flag.ErrHelp) {
			fmt.Printf("Usage: tool %s %s\n", c.Name(), c.Args())
			return
		}
		if err == nil {
			err = c.Run(env)
		}
		if err != nil {
			fmt.Println(err)
		}
	}

	run("greet", "-shout", "gopher")
	run("greet")
	run("greet", "-h")
	run("help")
	// Output:
	// HELLO, gopher
	// greet: name is required
	// Usage: tool greet <name>
	// Commands:
	//   greet    <name>  Print a greeting
	//   version          Print the version
}
