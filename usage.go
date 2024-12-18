package cli

import (
	"bytes"
	_ "embed"
	"flag"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"
)

func Usage() error {
	return flag.ErrHelp
}

//go:embed usage.gotext
var usageTemplate string

var defaultUsage = template.Must(template.New("usage").Funcs(colors).Parse(usageTemplate))

type usage struct {
	cmd *subcommand
}

func (u *usage) Name() string {
	return u.cmd.name
}

func (u *usage) Full() string {
	return u.cmd.full
}

func (u *usage) Usage() string {
	out := new(strings.Builder)
	if len(u.cmd.flags) > 0 {
		out.WriteString(" ")
		out.WriteString(dim())
		out.WriteString("[flags]")
		out.WriteString(reset())
	}
	if u.cmd.run != nil && len(u.cmd.args) > 0 {
		for _, arg := range u.cmd.args {
			out.WriteString(" ")
			out.WriteString(dim())
			out.WriteString("<")
			out.WriteString(arg.name)
			out.WriteString(">")
			out.WriteString(reset())
		}
	} else if len(u.cmd.commands) > 0 {
		out.WriteString(" ")
		out.WriteString(dim())
		out.WriteString("[command]")
		out.WriteString(reset())
	}
	return out.String()
}

func (u *usage) Description() string {
	return u.cmd.help
}

func (u *usage) Args() (args usageArgs) {
	for _, arg := range u.cmd.args {
		args = append(args, &usageArg{arg})
	}
	return args
}

type usageArg struct {
	a *Arg
}

type usageArgs []*usageArg

func (args usageArgs) Usage() (string, error) {
	buf := new(bytes.Buffer)
	tw := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
	for _, arg := range args {
		tw.Write([]byte("\t\t<" + arg.a.name + ">"))
		if arg.a.help != "" {
			tw.Write([]byte("\t" + dim() + arg.a.help + reset()))
		}
		tw.Write([]byte("\n"))
	}
	if err := tw.Flush(); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func (u *usage) Commands() (commands usageCommands) {
	for _, cmd := range u.cmd.commands {
		if cmd.advanced || cmd.hidden {
			continue
		}
		commands = append(commands, &usageCommand{cmd})
	}
	// Sort by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].c.name < commands[j].c.name
	})
	return commands
}

func (u *usage) Advanced() (commands usageCommands) {
	for _, cmd := range u.cmd.commands {
		if !cmd.advanced || cmd.hidden {
			continue
		}
		commands = append(commands, &usageCommand{cmd})
	}
	// Sort by name
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].c.name < commands[j].c.name
	})
	return commands
}

func (u *usage) Flags() (flags usageFlags) {
	flags = make(usageFlags, len(u.cmd.flags))
	for i, flag := range u.cmd.flags {
		flags[i] = &usageFlag{flag}
	}
	// Sort by name
	sort.Slice(flags, func(i, j int) bool {
		if hasShort(flags[i]) == hasShort(flags[j]) {
			// Both have shorts or don't have shorts, so sort by name
			return flags[i].f.name < flags[j].f.name
		}
		// Shorts above non-shorts
		return flags[i].f.short > flags[j].f.short
	})
	return flags
}

type usageCommand struct {
	c *subcommand
}

type usageCommands []*usageCommand

func (cmds usageCommands) Usage() (string, error) {
	buf := new(bytes.Buffer)
	tw := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
	for _, cmd := range cmds {
		tw.Write([]byte("\t\t" + cmd.c.name))
		if cmd.c.help != "" {
			tw.Write([]byte("\t" + dim() + cmd.c.help + reset()))
		}
		tw.Write([]byte("\n"))
	}
	if err := tw.Flush(); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

type usageFlag struct {
	f *Flag
}

type usageFlags []*usageFlag

func (flags usageFlags) Usage() (string, error) {
	buf := new(bytes.Buffer)
	tw := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
	for _, flag := range flags {
		tw.Write([]byte("\t\t"))
		if flag.f.short != "" {
			tw.Write([]byte("-" + string(flag.f.short) + ", "))
		}
		tw.Write([]byte("--" + flag.f.name))
		if flag.f.help != "" {
			tw.Write([]byte("\t"))
			tw.Write([]byte(dim() + flag.f.help + reset()))
		}
		tw.Write([]byte("\n"))
	}
	if err := tw.Flush(); err != nil {
		return "", err
	}
	return strings.TrimSpace(buf.String()), nil
}

func hasShort(flag *usageFlag) bool {
	return flag.f.short != ""
}
