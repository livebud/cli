package cli

import (
	"bytes"
	_ "embed"
	"flag"
	"sort"
	"strconv"
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
	cmd *command
}

func (u *usage) Name() string {
	return u.cmd.name
}

func (u *usage) Full() string {
	return u.cmd.full
}

func valueIsOptionalOrHasDefault(v value) bool {
	if v == nil {
		return false
	}
	if v.optional() {
		return true
	}
	_, hasDefault := v.Default()
	return hasDefault
}

func formatFixedArgUsage(name string, v value) string {
	if valueIsOptionalOrHasDefault(v) {
		return "[" + name + "]"
	}
	return "<" + name + ">"
}

func formatRestArgsUsage(name string) string {
	return "[" + name + "...]"
}

func (u *usage) Usage() string {
	out := new(strings.Builder)
	if len(u.cmd.flags) > 0 {
		out.WriteString(" ")
		out.WriteString(dim())
		out.WriteString("[flags]")
		out.WriteString(reset())
	}
	if u.cmd.run != nil && (len(u.cmd.args) > 0 || u.cmd.restArgs != nil) {
		for _, arg := range u.cmd.args {
			out.WriteString(" ")
			out.WriteString(dim())
			out.WriteString(formatFixedArgUsage(arg.name, arg.value))
			out.WriteString(reset())
		}
		if u.cmd.restArgs != nil {
			out.WriteString(" ")
			out.WriteString(dim())
			out.WriteString(formatRestArgsUsage(u.cmd.restArgs.name))
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
		args = append(args, &usageArg{
			name:  arg.name,
			help:  arg.help,
			value: arg.value,
		})
	}
	if u.cmd.restArgs != nil {
		args = append(args, &usageArg{
			name:     u.cmd.restArgs.name,
			help:     u.cmd.restArgs.help,
			value:    u.cmd.restArgs.value,
			variadic: true,
		})
	}
	return args
}

type usageArg struct {
	name     string
	help     string
	value    value
	variadic bool
}

func (a *usageArg) Key() string {
	if a.variadic {
		return formatRestArgsUsage(a.name)
	}
	return formatFixedArgUsage(a.name, a.value)
}

func (a *usageArg) Suffix() string {
	if a.value == nil {
		return ""
	}
	if def, ok := a.value.Default(); ok {
		return " (default:" + strconv.Quote(def) + ")"
	} else if a.value.optional() {
		return " (optional)"
	}
	return ""
}

type usageArgs []*usageArg

func (args usageArgs) Usage() (string, error) {
	buf := new(bytes.Buffer)
	tw := tabwriter.NewWriter(buf, 0, 0, 2, ' ', 0)
	for _, arg := range args {
		tw.Write([]byte("\t\t"))
		tw.Write([]byte(arg.Key()))
		if arg.help != "" {
			tw.Write([]byte("\t" + dim()))
			tw.Write([]byte(arg.help))
			tw.Write([]byte(arg.Suffix()))
			tw.Write([]byte(reset()))
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
	c *command
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

func (u *usageFlag) Suffix() string {
	attrs := []string{}
	if u.f.env != nil {
		attrs = append(attrs, "or $"+*u.f.env)
	}
	if def, ok := u.f.value.Default(); ok {
		attrs = append(attrs, "default:"+strconv.Quote(def))
	} else if u.f.value.optional() {
		attrs = append(attrs, "optional")
	}
	if len(attrs) == 0 {
		return ""
	}
	out := new(strings.Builder)
	out.WriteString(" (")
	for i, v := range attrs {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(v)
	}
	out.WriteString(")")
	return out.String()
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
			tw.Write([]byte("\t" + dim()))
			tw.Write([]byte(flag.f.help))
			tw.Write([]byte(flag.Suffix()))
			tw.Write([]byte(reset()))
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
