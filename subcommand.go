package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
)

func newSubcommand(config *config, name, full, help string) *subcommand {
	fset := flag.NewFlagSet(name, flag.ContinueOnError)
	fset.SetOutput(io.Discard)
	return &subcommand{
		config:   config,
		fset:     fset,
		name:     name,
		full:     full,
		help:     help,
		commands: map[string]*subcommand{},
	}
}

type subcommand struct {
	config *config
	fset   *flag.FlagSet
	run    func(ctx context.Context) error
	parsed bool

	// state for the template
	name     string
	full     string
	help     string
	hidden   bool
	advanced bool
	commands map[string]*subcommand
	flags    []*Flag
	args     []*Arg
	restArgs *Args // optional, collects the rest of the args
}

var _ Command = (*subcommand)(nil)

func (c *subcommand) printUsage() error {
	return c.config.usage.Execute(c.config.writer, &usage{c})
}

type value interface {
	flag.Value
	verify(displayName string) error
}

// Set flags only once
func (c *subcommand) setFlags() {
	if c.parsed {
		return
	}
	c.parsed = true
	for _, flag := range c.flags {
		c.fset.Var(flag.value, flag.name, flag.help)
		if flag.short != 0 {
			c.fset.Var(flag.value, string(flag.short), flag.help)
		}
	}
}

func (c *subcommand) parse(ctx context.Context, args []string) error {
	// Set flags
	c.setFlags()
	// Parse the arguments
	if err := c.fset.Parse(args); err != nil {
		// Print usage if the developer used -h or --help
		if errors.Is(err, flag.ErrHelp) {
			return c.printUsage()
		}
		return err
	}
	// Check if the first argument is a subcommand
	if sub, ok := c.commands[c.fset.Arg(0)]; ok {
		return sub.parse(ctx, c.fset.Args()[1:])
	}
	// Handle the remaining arguments
	numArgs := len(c.args)
	restArgs := c.fset.Args()
loop:
	for i, arg := range restArgs {
		if i >= numArgs {
			if c.restArgs == nil {
				return fmt.Errorf("%w: %s", ErrInvalidInput, arg)
			}
			// Loop over the remaining unset args, appending them to restArgs
			for _, arg := range restArgs[i:] {
				c.restArgs.value.Set(arg)
			}
			break loop
		}
		if err := c.args[i].value.Set(arg); err != nil {
			return err
		}
	}
	// Verify that all the args have been set or have default values
	if err := verifyArgs(c.args); err != nil {
		return err
	}
	// Also verify rest args if we have any
	if c.restArgs != nil {
		if err := c.restArgs.verify(c.restArgs.name); err != nil {
			return err
		}
	}
	// Print usage if there's no run function defined
	if c.run == nil {
		if len(restArgs) == 0 {
			return c.printUsage()
		}
		return fmt.Errorf("%w: %s", ErrInvalidInput, c.fset.Arg(0))
	}
	// Verify that all the flags have been set or have default values
	if err := verifyFlags(c.flags); err != nil {
		return err
	}
	if err := c.run(ctx); err != nil {
		// Support explicitly printing usage
		if errors.Is(err, flag.ErrHelp) {
			return c.printUsage()
		}
		return err
	}
	return nil
}

func (c *subcommand) Run(runner func(ctx context.Context) error) {
	c.run = runner
}

func (c *subcommand) Command(name, help string) Command {
	if c.commands[name] != nil {
		return c.commands[name]
	}
	cmd := newSubcommand(c.config, name, c.full+" "+name, help)
	c.commands[name] = cmd
	return cmd
}

func (c *subcommand) Hidden() Command {
	c.hidden = true
	return c
}

func (c *subcommand) Advanced() Command {
	c.advanced = true
	return c
}

func (c *subcommand) Arg(name, help string) *Arg {
	arg := &Arg{
		name: name,
		help: help,
	}
	c.args = append(c.args, arg)
	return arg
}

func (c *subcommand) Args(name, help string) *Args {
	if c.restArgs != nil {
		// Panic is okay here because settings commands should be done during
		// initialization. We want to fail fast for invalid usage.
		panic("commander: you can only use cmd.Args(name, usage) once per command")
	}
	args := &Args{
		name: name,
		help: help,
	}
	c.restArgs = args
	return args
}

func (c *subcommand) Flag(name, help string) *Flag {
	flag := &Flag{
		name: name,
		help: help,
	}
	c.flags = append(c.flags, flag)
	return flag
}

func (c *subcommand) Find(cmds ...string) (*subcommand, bool) {
	if len(cmds) == 0 {
		return c, true
	}
	cmd := cmds[0]
	sub, ok := c.commands[cmd]
	if !ok {
		return nil, false
	}
	return sub.Find(cmds[1:]...)
}