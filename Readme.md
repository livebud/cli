# CLI

Build beautiful CLIs in Go. A simpler alternative to [kingpin](https://github.com/alecthomas/kingpin).

<img width="601" alt="CleanShot 2023-06-20 at 22 23 14@2x" src="https://github.com/livebud/cli/assets/170299/f29b7f43-ff1a-445e-9642-af300742ae4f">

## Features

- Type-safe, fluent API
- Flag, command and argument support
- Required and optional parameters
- Built entirely on the [flag](https://pkg.go.dev/flag) package the standard library
- Colon-based subcommands (e.g. `new:controller`)
- `SIGINT` context cancellation out-of-the-box
- Custom help messages
- Respects `NO_COLOR`

## Install

```sh
go get -u github.com/livebud/cli
```

## Example

```go
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
```

## Contributing

First, clone the repo:

```sh
git clone https://github.com/livebud/cli
cd cli
```

Next, install dependencies:

```sh
go mod tidy
```

Finally, try running the tests:

```sh
go test ./...
```

## License

MIT
