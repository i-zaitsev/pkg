// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cmd

import (
	"errors"
	"flag"
	"io"
	"testing"

	"github.com/i-zaitsev/pkg/assert"
)

type stubCmd struct {
	name, desc, args string
}

func (s *stubCmd) Name() string { return s.name }
func (s *stubCmd) Desc() string { return s.desc }
func (s *stubCmd) Args() string { return s.args }

func (s *stubCmd) Parse([]string) error { return nil }
func (s *stubCmd) Run(*testEnv) error   { return nil }

type testEnv struct{}

type otherCmd struct{}

func (otherCmd) Name() string         { return "other" }
func (otherCmd) Desc() string         { return "" }
func (otherCmd) Args() string         { return "" }
func (otherCmd) Parse([]string) error { return nil }
func (otherCmd) Run(*testEnv) error   { return nil }

func TestRegister_addsToCallerRegistry(t *testing.T) {
	custom := &stubCmd{name: "custom"}
	run := &stubCmd{name: "run"}
	reg := map[string]Command[*testEnv]{"custom": custom}

	err := Register(reg, run, otherCmd{})

	assert.NotErr(t, err)
	assert.Eq(t, reg, map[string]Command[*testEnv]{"custom": custom, "run": run, "other": otherCmd{}})
}

func TestRegister_invalidBatch_leavesRegistryUnchanged(t *testing.T) {
	existing := &stubCmd{name: "run"}
	tests := []struct {
		name    string
		cmds    []Command[*testEnv]
		wantErr string
	}{
		{"duplicate_of_existing", []Command[*testEnv]{&stubCmd{name: "list"}, &stubCmd{name: "run"}}, `duplicate command "run"`},
		{"duplicate_in_batch", []Command[*testEnv]{&stubCmd{name: "list"}, &stubCmd{name: "list"}}, `duplicate command "list"`},
		{"empty_name", []Command[*testEnv]{&stubCmd{name: "list"}, &stubCmd{}}, "empty command name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := map[string]Command[*testEnv]{"run": existing}

			err := Register(reg, tt.cmds...)

			assert.ErrContains(t, err, tt.wantErr)
			assert.Eq(t, reg, map[string]Command[*testEnv]{"run": existing})
		})
	}
}

func TestUsage(t *testing.T) {
	tests := []struct {
		name string
		cmds []Info
		want string
	}{
		{
			name: "none",
			want: "",
		},
		{
			name: "single_with_args",
			cmds: []Info{&stubCmd{name: "run", desc: "Run a task", args: "<file>"}},
			want: "  run  <file>  Run a task\n",
		},
		{
			name: "aligned_and_sorted_by_name",
			cmds: []Info{
				&stubCmd{name: "status", desc: "Show status"},
				&stubCmd{name: "run", desc: "Run a task", args: "<file>"},
			},
			want: "  run     <file>  Run a task\n" +
				"  status          Show status\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Eq(t, Usage(tt.cmds...), tt.want)
		})
	}
}

func TestUsage_doesNotReorderCallerSlice(t *testing.T) {
	cmds := []Info{&stubCmd{name: "b"}, &stubCmd{name: "a"}}

	_ = Usage(cmds...)

	assert.Eq(t, cmds[0].Name(), "b")
}

func TestNewFlagSet_quietAndNamedAfterCommand(t *testing.T) {
	fs := NewFlagSet(&stubCmd{name: "alpha"})

	assert.Eq(t, fs.Name(), "alpha")
	assert.Eq(t, fs.ErrorHandling(), flag.ContinueOnError)
	assert.True(t, fs.Output() == io.Discard)
}

func TestNewFlagSet_help_returnsErrHelp(t *testing.T) {
	fs := NewFlagSet(&stubCmd{name: "alpha"})

	err := fs.Parse([]string{"-h"})

	assert.ErrIs(t, err, flag.ErrHelp)
}

func TestRequiredArg(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		want    string
		wantErr bool
	}{
		{name: "present", args: []string{"task.yaml", "extra"}, want: "task.yaml"},
		{name: "missing_nil", wantErr: true},
		{name: "missing_empty", args: []string{}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &stubCmd{name: "demo"}
			fs := NewFlagSet(c)
			assert.NotErr(t, fs.Parse(tt.args))

			got, err := RequiredArg(c, fs, "task file")

			if !tt.wantErr {
				assert.NotErr(t, err)
				assert.Eq(t, got, tt.want)
				return
			}
			missing := assert.AsErr[*ArgMissingError](t, err)
			assert.Eq(t, missing.Name, "task file")
			assert.ErrContains(t, err, "demo: task file is required")
		})
	}
}

func TestErrorf_prefixesCommandName(t *testing.T) {
	errBase := errors.New("boom")

	err := Errorf(&stubCmd{name: "alpha"}, "step %d: %w", 2, errBase)

	assert.Eq(t, err.Error(), "alpha: step 2: boom")
	assert.ErrIs(t, err, errBase)
}

func TestArgInvalidError(t *testing.T) {
	err := &ArgInvalidError{Name: "mode", Value: "fast"}

	assert.Eq(t, err.Error(), "invalid mode: fast")
}
