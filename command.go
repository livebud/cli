package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

func newCommand(config *config, flags []*Flag, name, full, help string) *command {
	fset := flag.NewFlagSet(name, flag.ContinueOnError)
	fset.SetOutput(io.Discard)
	return &command{
		config:   config,
		fset:     fset,
		name:     name,
		full:     full,
		flags:    flags,
		help:     help,
		commands: map[string]*command{},
	}
}

type command struct {
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
	commands map[string]*command
	flags    []*Flag
	args     []*Arg
	restArgs *Args // optional, collects the rest of the args
}

var _ Command = (*command)(nil)

func (c *command) printUsage() error {
	return c.config.usage.Execute(c.config.writer, &usage{c})
}

type value interface {
	flag.Value
	optional() bool
	verify() error
	Default() (string, bool)
}

// Set flags only once
func (c *command) setFlags() error {
	if c.parsed {
		return nil
	}
	c.parsed = true
	seen := map[string]bool{}
	for _, flag := range c.flags {
		if seen[flag.name] {
			return fmt.Errorf("%w %q command contains a duplicate flag \"--%s\"", ErrInvalidInput, c.full, flag.name)
		}
		seen[flag.name] = true
		c.fset.Var(flag.value, flag.name, flag.help)
		if flag.short != "" {
			if seen[flag.short] {
				return fmt.Errorf("%w %q command contains a duplicate flag \"-%s\"", ErrInvalidInput, c.full, flag.short)
			}
			seen[flag.short] = true
			c.fset.Var(flag.value, flag.short, flag.help)
		}
	}
	return nil
}

func (c *command) parse(ctx context.Context, args []string) error {
	// Set flags
	if err := c.setFlags(); err != nil {
		return err
	}
	// Parse the arguments
	if err := c.fset.Parse(args); err != nil {
		// Print usage if the developer used -h or --help
		if errors.Is(err, flag.ErrHelp) {
			return c.printUsage()
		}
		return maybeTrimError(err)
	}
	// Check if the first argument is a subcommand
	if sub, ok := c.commands[c.fset.Arg(0)]; ok {
		return sub.parse(ctx, c.fset.Args()[1:])
	}
	// Handle the remaining arguments
	numArgs := len(c.args)
	restArgs := c.fset.Args()

	// restArgs will start with an arg, so before parsing flags, check that the
	// command can handle additional args
	if len(restArgs) > 0 && len(c.args) == 0 && c.restArgs == nil {
		return fmt.Errorf("%w with unxpected arg %q", ErrInvalidInput, restArgs[0])
	}

	// Also parse the flags after an arg
	restArgs, err := parseFlags(c.fset, restArgs)
	if err != nil {
		return err
	}

loop:
	for i, arg := range restArgs {
		if i >= numArgs {
			if c.restArgs == nil {
				return fmt.Errorf("%w: %s", ErrInvalidInput, arg)
			}
			// Loop over the remaining unset args, appending them to restArgs
			if c.restArgs != nil {
				for _, arg := range restArgs[i:] {
					if err := c.restArgs.value.Set(arg); err != nil {
						return err
					}
				}
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
		if err := c.restArgs.verify(); err != nil {
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

func (c *command) Run(runner func(ctx context.Context) error) {
	c.run = runner
}

func (c *command) Command(name, help string) Command {
	if c.commands[name] != nil {
		return c.commands[name]
	}
	// Copy the flags from the parent command
	flags := append([]*Flag{}, c.flags...)
	// Create the subcommand
	cmd := newCommand(c.config, flags, name, c.full+" "+name, help)
	c.commands[name] = cmd
	return cmd
}

func (c *command) Hidden() Command {
	c.hidden = true
	return c
}

func (c *command) Advanced() Command {
	c.advanced = true
	return c
}

func (c *command) Arg(name, help string) *Arg {
	arg := &Arg{
		name: name,
		help: help,
	}
	c.args = append(c.args, arg)
	return arg
}

func (c *command) Args(name, help string) *Args {
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

func (c *command) Flag(name, help string) *Flag {
	flag := &Flag{
		name: name,
		help: help,
	}
	c.flags = append(c.flags, flag)
	return flag
}

func (c *command) Find(cmds ...string) (*command, bool) {
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

func parseFlags(fset *flag.FlagSet, args []string) (rest []string, err error) {
	for i, arg := range args {
		if !strings.HasPrefix(arg, "-") {
			rest = append(rest, arg)
			continue
		}
		if err := fset.Parse(args[i:]); err != nil {
			return nil, err
		}
		remaining, err := parseFlags(fset, fset.Args())
		if err != nil {
			return nil, err
		}
		rest = append(rest, remaining...)
		return rest, nil
	}
	return rest, nil
}

// This is a hack to trim the error messages returned by the flag package.
func maybeTrimError(err error) error {
	msg := err.Error()
	idx := -1
	if strings.HasPrefix(msg, "invalid value ") {
		idx = maybeTrimInvalidValue(msg)
	} else if strings.HasPrefix(msg, "invalid boolean value ") {
		idx = maybeTrimInvalidBooleanValue(msg)
	}
	if idx < 0 {
		return err
	}
	return errors.New(msg[idx:])
}

func maybeTrimInvalidValue(msg string) int {
	i1 := strings.Index(msg, `" for flag -`)
	if i1 < 0 {
		return i1
	}
	i1 += 12
	i2 := strings.Index(msg[i1:], ": ")
	if i2 < 0 {
		return i2
	}
	i2 += 2
	return i1 + i2
}

func maybeTrimInvalidBooleanValue(msg string) int {
	i1 := strings.Index(msg, `" for -`)
	if i1 < 0 {
		return i1
	}
	i1 += 7
	i2 := strings.Index(msg[i1:], `: `)
	if i2 < 0 {
		return i2
	}
	i2 += 2
	return i1 + i2
}
