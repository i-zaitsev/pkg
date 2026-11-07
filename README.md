# pkg

Small, reusable Go packages shared across projects.

| Package | Import path | Purpose |
|---|---|---|
| `assert` | `github.com/i-zaitsev/pkg/assert` | Test assertions, plain or fluent |
| `cmd` | `github.com/i-zaitsev/pkg/cmd` | Subcommand registry, dispatch, and flag helpers |

```sh
go get github.com/i-zaitsev/pkg@latest
```

## assert

```go
func TestSum(t *testing.T) {
	got, err := Sum(1, 2)

	assert.NotErr(t, err)
	assert.Eq(t, got, 3)
	assert.Val(t, got).NotZero().Eq(3)
}

func TestOpen(t *testing.T) {
	_, err := Open("missing.txt")

	var pe *fs.PathError
	assert.ErrVal(t, err).Is(fs.ErrNotExist).As(&pe).Fatal()
	assert.Eq(t, pe.Path, "missing.txt")
}
```

Every assertion reports through `t.Errorf`, so the test continues after a failure.
In an `ErrVal` chain, `Fatal()` stops the test if any earlier check in the chain failed.

## cmd

A command describes itself, parses its own arguments, and runs in the tool's env.
The tool owns the registry, dispatch, global flags, and help text.

```go
type Env struct{ out io.Writer }

type greet struct{ name string }

func (g *greet) Name() string { return "greet" }
func (g *greet) Desc() string { return "Print a greeting" }
func (g *greet) Args() string { return "<name>" }

func (g *greet) Parse(args []string) error {
	fs := cmd.NewFlagSet(g)
	if err := fs.Parse(args); err != nil {
		return cmd.Errorf(g, "%w", err)
	}
	var err error
	g.name, err = cmd.RequiredArg(g, fs, "name")
	return err
}

func (g *greet) Run(e *Env) error {
	_, err := fmt.Fprintf(e.out, "hello, %s\n", g.name)
	return err
}

func main() {
	reg := map[string]cmd.Command[*Env]{}
	if err := cmd.Register(reg, &greet{}); err != nil {
		log.Fatal(err)
	}

	var c cmd.Command[*Env]
	if len(os.Args) > 1 {
		c = reg[os.Args[1]]
	}
	if c == nil {
		fmt.Print("Commands:\n", cmd.Usage(slices.Collect(maps.Values(reg))...))
		return
	}
	if err := c.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}
	if err := c.Run(&Env{out: os.Stdout}); err != nil {
		log.Fatal(err)
	}
}
```

| Helper | Purpose |
|---|---|
| `Register(reg, cmds...)` | Adds commands to the caller's registry; fails on empty or duplicate names and leaves `reg` unchanged |
| `Usage(cmds...)` | Aligned `name  args  desc` lines, sorted by name |
| `NewFlagSet(c)` | Quiet flag set named after the command; `-h` returns `flag.ErrHelp` |
| `RequiredArg(c, fs, name)` | First positional argument or `ArgMissingError` |
| `Errorf(c, format, args...)` | `fmt.Errorf` prefixed with the command name |

## License

[MPL-2.0](LICENSE)
