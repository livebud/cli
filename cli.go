package cli

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"text/template"
)

var ErrInvalidInput = errors.New("cli: invalid input")
var ErrCommandNotFound = errors.New("cli: command not found")

type Middleware = func(next func(ctx context.Context) error) func(ctx context.Context) error

type Command interface {
	Command(name, help string) Command
	Hidden() Command
	Advanced() Command
	Flag(name, help string) *Flag
	Arg(name, help string) *Arg
	Args(name, help string) *Args
	Use(middlewares ...Middleware) Command
	Run(runner func(ctx context.Context) error)
}

func New(name, help string) *CLI {
	config := &config{"", os.Stdout, defaultUsage, defaultSignals()}
	return &CLI{newCommand(config, []*Flag{}, name, name, help), config}
}

type CLI struct {
	root   *command
	config *config
}

var _ Command = (*CLI)(nil)

type config struct {
	version string
	writer  io.Writer
	usage   *template.Template
	signals []os.Signal
}

func (c *CLI) Writer(writer io.Writer) *CLI {
	c.config.writer = writer
	return c
}

func (c *CLI) Version(version string) *CLI {
	c.config.version = version
	return c
}

func (c *CLI) Template(template *template.Template) *CLI {
	c.config.usage = template
	return c
}

func (c *CLI) Trap(signals ...os.Signal) *CLI {
	c.config.signals = signals
	return c
}

func (c *CLI) Parse(ctx context.Context, args ...string) error {
	// Trap signals if any were provided
	ctx = trap(ctx, c.config.signals...)
	// Support basic tab completion
	if compline := os.Getenv("COMP_LINE"); compline != "" {
		return c.complete(compline)
	}
	// Parse the command line arguments
	if err := c.root.parse(ctx, args); err != nil {
		return err
	}
	// Give the caller a chance to handle context cancellations and therefore
	// interrupts specifically.
	return ctx.Err()
}

func (c *CLI) Command(name, help string) Command {
	return c.root.Command(name, help)
}

func (c *CLI) Hidden() Command {
	return c.root.Hidden()
}

func (c *CLI) Advanced() Command {
	return c.root.Advanced()
}

func (c *CLI) Flag(name, help string) *Flag {
	return c.root.Flag(name, help)
}

func (c *CLI) Arg(name, help string) *Arg {
	return c.root.Arg(name, help)
}

func (c *CLI) Args(name, help string) *Args {
	return c.root.Args(name, help)
}

func (c *CLI) Run(runner func(ctx context.Context) error) {
	c.root.Run(runner)
}

func (c *CLI) Use(middlewares ...Middleware) Command {
	return c.root.Use(middlewares...)
}

func (c *CLI) Find(subcommand ...string) (Command, error) {
	return c.find(subcommand...)
}

func (c *CLI) find(subcommand ...string) (*command, error) {
	sub, ok := c.root.Find(subcommand...)
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrCommandNotFound, strings.Join(subcommand, " "))
	}
	return sub, nil
}

func (c *CLI) complete(compline string) error {
	fields := strings.Fields(compline)
	cmd, err := c.find(fields[1:]...)
	if err != nil {
		// If the command wasn't found, don't print anything
		return nil
	}
	for _, cmd := range cmd.commands {
		if cmd.hidden {
			continue
		}
		c.config.writer.Write([]byte(cmd.name + "\n"))
	}
	return nil
}

func trap(parent context.Context, signals ...os.Signal) context.Context {
	if len(signals) == 0 {
		return parent
	}
	ctx, stop := signal.NotifyContext(parent, signals...)
	// If context was canceled, stop catching signals
	go func() {
		<-ctx.Done()
		stop()
	}()
	return ctx
}

func lookupEnv(key *string) (string, bool) {
	if key == nil {
		return "", false
	}
	return os.LookupEnv(*key)
}

type missingInputError struct {
	Key string
	Env *string
}

func (m *missingInputError) Error() string {
	s := new(strings.Builder)
	s.WriteString("missing ")
	s.WriteString(m.Key)
	if m.Env != nil {
		s.WriteString(" or ")
		s.WriteString("$" + *m.Env)
		s.WriteString(" environment variable")
	}
	return s.String()
}

// If called from `go run` or `go test` don't trap any signals by default. This
// avoids the "double Ctrl-C" problem where the user has to hit Ctrl-C twice to
// exit the program.
func defaultSignals() []os.Signal {
	exe, err := os.Executable()
	if err != nil {
		return []os.Signal{os.Interrupt}
	}
	if strings.Contains(exe, string(filepath.Separator)+"go-build") {
		return []os.Signal{}
	}
	// Otherwise, trap interrupts by default
	return []os.Signal{os.Interrupt}
}
