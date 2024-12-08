package cli

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"
	"text/template"
)

var ErrInvalidInput = errors.New("cli: invalid input")
var ErrCommandNotFound = errors.New("cli: command not found")

type Command interface {
	Command(name, help string) Command
	Hidden() Command
	Advanced() Command
	Flag(name, help string) *Flag
	Arg(name, help string) *Arg
	Args(name, help string) *Args
	Run(runner func(ctx context.Context) error)
}

func New(name, help string) *CLI {
	config := &config{"", os.Stdout, defaultUsage, []os.Signal{os.Interrupt}}
	return &CLI{newSubcommand(config, []*Flag{}, name, name, help), config}
}

type CLI struct {
	root   *subcommand
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

func (c *CLI) Template(template *template.Template) {
	c.config.usage = template
}

func (c *CLI) Trap(signals ...os.Signal) {
	c.config.signals = signals
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

func (c *CLI) Find(subcommand ...string) (Command, error) {
	return c.find(subcommand...)
}

func (c *CLI) find(subcommand ...string) (*subcommand, error) {
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
