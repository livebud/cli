package main

import (
	"context"
	"fmt"
	"os"

	"github.com/livebud/cli"
)

func main() {
	flag := new(Flag)
	cli := cli.New("app", "your awesome cli").Writer(os.Stdout)
	cli.Flag("log", "log level").Short('L').String(&flag.Log).Default("info")
	cli.Flag("embed", "embed the code").Bool(&flag.Embed).Default(false)

	{ // new <dir>
		cmd := &New{Flag: flag}
		cli := cli.Command("new", "create a new project")
		cli.Arg("dir").String(&cmd.Dir)
		cli.Run(cmd.Run)
	}

	ctx := context.Background()
	if err := cli.Parse(ctx, os.Args[1:]...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type Flag struct {
	Log   string
	Embed bool
}

type New struct {
	Flag *Flag
	Dir  string
}

// Run new
func (n *New) Run(ctx context.Context) error {
	return nil
}
