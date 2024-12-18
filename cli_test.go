package cli_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/livebud/cli"
	"github.com/matryer/is"
	"github.com/matthewmueller/diff"
)

func isEqual(t testing.TB, actual, expected string) {
	t.Helper()
	equal(t, expected, replaceEscapeCodes(actual))
}

func replaceEscapeCodes(str string) string {
	r := strings.NewReplacer(
		"\033[0m", `{reset}`,
		"\033[1m", `{bold}`,
		"\033[37m", `{dim}`,
		"\033[4m", `{underline}`,
		"\033[36m", `{teal}`,
		"\033[34m", `{blue}`,
		"\033[33m", `{yellow}`,
		"\033[31m", `{red}`,
		"\033[32m", `{green}`,
	)
	return r.Replace(str)
}

// is checks if expect and actual are equal
func equal(t testing.TB, expect, actual string) {
	t.Helper()
	if expect == actual {
		return
	}
	var b bytes.Buffer
	b.WriteString("\n\x1b[4mExpect\x1b[0m:\n")
	b.WriteString(expect)
	b.WriteString("\n\n")
	b.WriteString("\x1b[4mActual\x1b[0m: \n")
	b.WriteString(actual)
	b.WriteString("\n\n")
	b.WriteString("\x1b[4mDifference\x1b[0m: \n")
	b.WriteString(diff.String(expect, actual))
	b.WriteString("\n")
	t.Fatal(b.String())
}

func encode(w io.Writer, cmd, in interface{}) error {
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"cmd": cmd,
		"in":  in,
	})
}

func heroku(w io.Writer) *cli.CLI {
	type global struct {
		App    string
		Remote *string
	}
	var g global
	cli := cli.New("heroku", `CLI to interact with Heroku`).Writer(w)
	cli.Flag("app", "app to run command against").Short('a').String(&g.App)
	cli.Flag("remote", "git remote of app to use").Short('r').Optional().String(&g.Remote)

	{
		var in = struct {
			*global
			All  bool
			Json bool
		}{global: &g}
		cli := cli.Command("addons", `lists your add-ons and attachments`)
		cli.Flag("all", "show add-ons and attachments for all accessible apps").Bool(&in.All).Default(false)
		cli.Flag("json", "return add-ons in json format").Bool(&in.Json).Default(false)
		cli.Run(func(ctx context.Context) error { return encode(w, "addons", in) })

		{
			var in = struct {
				*global
				Name *string
				As   *string
			}{global: &g}
			cli := cli.Command("attach", `attach an existing add-on resource to an app`)
			cli.Flag("name", "name for the add-on resource").Optional().String(&in.Name)
			cli.Flag("as", "name for add-on attachment").Optional().String(&in.As)
			cli.Run(func(ctx context.Context) error { return encode(w, "addons:attach", in) })
		}

		{
			var in = struct {
				*global
				Name *string
				As   *string
				Wait bool
			}{global: &g}
			cli := cli.Command("create", `create a new add-on resource and attach to an app`)
			cli.Flag("name", "name for the add-on resource").Optional().String(&in.Name)
			cli.Flag("as", "name for add-on attachment").Optional().String(&in.As)
			cli.Flag("wait", "watch add-on creation status and exit when complete").Bool(&in.Wait).Default(false)
			cli.Run(func(ctx context.Context) error { return encode(w, "addons:create", in) })
		}

		{
			var in = struct {
				global
				Force bool
			}{global: g}
			cli := cli.Command("destroy", `destroy permanently destroys an add-on resource`)
			cli.Flag("force", "force destroy").Bool(&in.Force).Default(false)
			cli.Run(func(ctx context.Context) error { return encode(w, "addons:destroy", in) })
		}

		{
			var in = struct {
				*global
			}{global: &g}
			cli := cli.Command("info", `show detailed information for an add-on`)
			cli.Run(func(ctx context.Context) error { return encode(w, "addons:info", in) })
		}
	}

	{
		var in = struct {
			*global
			Json bool
		}{global: &g}
		cli := cli.Command("ps", `list dynos for an app`)
		cli.Flag("json", "output in json format").Bool(&in.Json).Default(false)
		cli.Run(func(ctx context.Context) error { return encode(w, "ps", in) })

		{
			var in = struct {
				*global
				Value string
			}{global: &g}
			cli := cli.Command("scale", `scale dyno quantity up or down`)
			cli.Arg("value", "some value").String(&in.Value)
			cli.Run(func(ctx context.Context) error {
				return encode(w, "ps:scale", in)
			})
		}

		{
			cli := cli.Command("autoscale", `enable autoscaling for an app`)

			{
				var in = struct {
					*global
					Min           int
					Max           int
					Notifications bool
					P95           int
				}{global: &g}
				cli := cli.Command("enable", `enable autoscaling for an app`)
				cli.Flag("min", "minimum number of dynos").Int(&in.Min)
				cli.Flag("max", "maximum number of dynos").Int(&in.Max)
				cli.Flag("notifications", "comma-separated list of notifications to enable").Bool(&in.Notifications)
				cli.Flag("p95", "95th percentile response time threshold").Int(&in.P95)
				cli.Run(func(ctx context.Context) error { return encode(w, "ps:autoscale:enable", in) })
			}

			{
				var in = struct {
					*global
				}{global: &g}
				cli := cli.Command("disable", `disable autoscaling for an app`)
				cli.Run(func(ctx context.Context) error { return encode(w, "ps:autoscale:disable", in) })
			}
		}
	}

	return cli
}

func TestHerokuHelp(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := heroku(actual)
	ctx := context.Background()
	err := cli.Parse(ctx, "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} heroku {dim}[flags]{reset} {dim}[command]{reset}

  {bold}Description:{reset}
    CLI to interact with Heroku

  {bold}Flags:{reset}
    -a, --app     {dim}app to run command against{reset}
    -r, --remote  {dim}git remote of app to use{reset}

  {bold}Commands:{reset}
    addons  {dim}lists your add-ons and attachments{reset}
    ps      {dim}list dynos for an app{reset}

`)
}

func TestHerokuHelpPs(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := heroku(actual)
	ctx := context.Background()
	err := cli.Parse(ctx, "ps", "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} heroku ps {dim}[flags]{reset} {dim}[command]{reset}

  {bold}Description:{reset}
    list dynos for an app

  {bold}Flags:{reset}
    -a, --app     {dim}app to run command against{reset}
    -r, --remote  {dim}git remote of app to use{reset}
    --json        {dim}output in json format{reset}

  {bold}Commands:{reset}
    autoscale  {dim}enable autoscaling for an app{reset}
    scale      {dim}scale dyno quantity up or down{reset}

`)
}

func TestHerokuHelpPsAutoscale(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := heroku(actual)
	ctx := context.Background()
	err := cli.Parse(ctx, "ps", "autoscale", "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} heroku ps autoscale {dim}[flags]{reset} {dim}[command]{reset}

  {bold}Description:{reset}
    enable autoscaling for an app

  {bold}Flags:{reset}
    -a, --app     {dim}app to run command against{reset}
    -r, --remote  {dim}git remote of app to use{reset}
    --json        {dim}output in json format{reset}

  {bold}Commands:{reset}
    disable  {dim}disable autoscaling for an app{reset}
    enable   {dim}enable autoscaling for an app{reset}

`)
}

func TestHerokuHelpPsAutoscaleEnable(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := heroku(actual)
	ctx := context.Background()
	err := cli.Parse(ctx, "ps", "autoscale", "enable", "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} heroku ps autoscale enable {dim}[flags]{reset}

  {bold}Description:{reset}
    enable autoscaling for an app

  {bold}Flags:{reset}
    -a, --app        {dim}app to run command against{reset}
    -r, --remote     {dim}git remote of app to use{reset}
    --json           {dim}output in json format{reset}
    --max            {dim}maximum number of dynos{reset}
    --min            {dim}minimum number of dynos{reset}
    --notifications  {dim}comma-separated list of notifications to enable{reset}
    --p95            {dim}95th percentile response time threshold{reset}

`)
}

// Possible signatures:
// cli [flags|args]
// cli [sub] [flags|args]

func TestHerokuFlagArgOrder(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := heroku(actual)
	ctx := context.Background()
	err := cli.Parse(ctx, "ps", "scale", "--app=foo", "--remote", "bar", "web=1")
	is.NoErr(err)
	is.Equal(actual.String(), `{"cmd":"ps:scale","in":{"App":"foo","Remote":"bar","Value":"web=1"}}`+"\n")
}

func TestHerokuArgFlagOutOfOrder(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cmd := heroku(actual)
	ctx := context.Background()
	err := cmd.Parse(ctx, "ps", "scale", "web=1", "--app=foo", "--remote", "bar")
	is.NoErr(err)
	is.Equal(actual.String(), `{"cmd":"ps:scale","in":{"App":"foo","Remote":"bar","Value":"web=1"}}`+"\n")
}

func TestHerokuInvalidFlagArgFlagOrderTwo(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cmd := heroku(actual)
	ctx := context.Background()
	err := cmd.Parse(ctx, "ps:scale", "--app=foo", "web=1", "--remote", "bar")
	is.True(errors.Is(err, cli.ErrInvalidInput))
}

func TestHelpArgs(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cmd := cli.New("cp", "copy files").Writer(actual)
	cmd.Arg("src", "source").String(nil)
	cmd.Arg("dst", "destination").String(nil).Default(".")
	cmd.Run(func(ctx context.Context) error { return nil })
	ctx := context.Background()
	err := cmd.Parse(ctx, "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} cp {dim}<src>{reset} {dim}<dst>{reset}

  {bold}Description:{reset}
    copy files

  {bold}Args:{reset}
    <src>  {dim}source{reset}
    <dst>  {dim}destination{reset}

`)
}

func TestInvalid(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cmd := cli.New("cli", "desc").Writer(actual)
	ctx := context.Background()
	err := cmd.Parse(ctx, "blargle")
	is.True(err != nil)
	is.True(errors.Is(err, cli.ErrInvalidInput))
	isEqual(t, actual.String(), ``)
}

func TestSimple(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("cli", "desc").Writer(actual)
	called := 0
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	isEqual(t, actual.String(), ``)
}

func TestFlagString(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag string
	cli.Flag("flag", "cli flag").String(&flag)
	ctx := context.Background()
	err := cli.Parse(ctx, "--flag", "cool")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(flag, "cool")
	isEqual(t, actual.String(), ``)
}

func TestFlagStringDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag string
	cli.Flag("flag", "cli flag").String(&flag).Default("default")
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(flag, "default")
	isEqual(t, actual.String(), ``)
}

func TestFlagStringRequired(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag string
	cli.Flag("flag", "cli flag").String(&flag)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), "missing --flag")
}

func TestFlagInt(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag int
	cli.Flag("flag", "cli flag").Int(&flag)
	ctx := context.Background()
	err := cli.Parse(ctx, "--flag", "10")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(flag, 10)
	isEqual(t, actual.String(), ``)
}

func TestFlagIntDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag int
	cli.Flag("flag", "cli flag").Int(&flag).Default(10)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(flag, 10)
	isEqual(t, actual.String(), ``)
}

func TestFlagIntRequired(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag int
	cli.Flag("flag", "cli flag").Int(&flag)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), "missing --flag")
}

func TestFlagBool(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag bool
	cli.Flag("flag", "cli flag").Bool(&flag)
	ctx := context.Background()
	err := cli.Parse(ctx, "--flag")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(flag, true)
	isEqual(t, actual.String(), ``)
}

func TestFlagBoolDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag bool
	cli.Flag("flag", "cli flag").Bool(&flag).Default(true)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(flag, true)
	isEqual(t, actual.String(), ``)
}

func TestFlagBoolDefaultFalse(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag bool
	cli.Flag("flag", "cli flag").Bool(&flag).Default(true)
	ctx := context.Background()
	err := cli.Parse(ctx, "--flag=false")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(flag, false)
	isEqual(t, actual.String(), ``)
}

func TestFlagBoolRequired(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag bool
	cli.Flag("flag", "cli flag").Bool(&flag)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), "missing --flag")
}

func TestFlagStrings(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flags []string
	cli.Flag("flag", "cli flag").Strings(&flags)
	ctx := context.Background()
	err := cli.Parse(ctx, "--flag", "1", "--flag", "2")
	is.NoErr(err)
	is.Equal(len(flags), 2)
	is.Equal(flags[0], "1")
	is.Equal(flags[1], "2")
}

func TestFlagStringsRequired(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flags []string
	cli.Flag("flag", "cli flag").Strings(&flags)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), "missing --flag")
}

func TestFlagStringsDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flags []string
	cli.Flag("flag", "cli flag").Strings(&flags).Default("a", "b")
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(len(flags), 2)
	is.Equal(flags[0], "a")
	is.Equal(flags[1], "b")
}

func TestFlagStringMap(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flags map[string]string
	cli.Flag("flag", "cli flag").StringMap(&flags)
	ctx := context.Background()
	err := cli.Parse(ctx, "--flag", "a:1 + 1", "--flag", "b:2")
	is.NoErr(err)
	is.Equal(len(flags), 2)
	is.Equal(flags["a"], "1 + 1")
	is.Equal(flags["b"], "2")
}

func TestFlagStringMapRequired(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flags map[string]string
	cli.Flag("flag", "cli flag").StringMap(&flags)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), "missing --flag")
}

func TestFlagStringMapDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flags map[string]string
	cli.Flag("flag", "cli flag").StringMap(&flags).Default(map[string]string{
		"a": "1",
		"b": "2",
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(len(flags), 2)
	is.Equal(flags["a"], "1")
	is.Equal(flags["b"], "2")
}

func TestArgStringMap(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var args map[string]string
	cli.Arg("arg", "arg map").StringMap(&args)
	// Can have only one arg
	ctx := context.Background()
	err := cli.Parse(ctx, "a:1 + 1")
	is.NoErr(err)
	is.Equal(len(args), 1)
	is.Equal(args["a"], "1 + 1")
}

func TestArgStringMapRequired(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var args map[string]string
	cli.Arg("arg", "arg map").StringMap(&args)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), "missing <arg>")
}

func TestArgStringMapDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var args map[string]string
	cli.Arg("arg", "arg map").StringMap(&args).Default(map[string]string{
		"a": "1",
		"b": "2",
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(len(args), 2)
	is.Equal(args["a"], "1")
	is.Equal(args["b"], "2")
}

func TestSub(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("bud", "bud CLI").Writer(actual)
	var trace []string
	cli.Run(func(ctx context.Context) error {
		trace = append(trace, "bud")
		return nil
	})
	{
		sub := cli.Command("run", "run your application")
		sub.Run(func(ctx context.Context) error {
			trace = append(trace, "run")
			return nil
		})
	}
	{
		sub := cli.Command("build", "build your application")
		sub.Run(func(ctx context.Context) error {
			trace = append(trace, "build")
			return nil
		})
	}
	ctx := context.Background()
	err := cli.Parse(ctx, "build")
	is.NoErr(err)
	is.Equal(len(trace), 1)
	is.Equal(trace[0], "build")
	isEqual(t, actual.String(), ``)
}

func TestSubHelp(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("bud", "bud CLI").Writer(actual)
	cli.Flag("log", "specify the logger").Bool(nil)
	cli.Command("run", "run your application")
	cli.Command("build", "build your application")
	ctx := context.Background()
	err := cli.Parse(ctx, "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} bud {dim}[flags]{reset} {dim}[command]{reset}

  {bold}Description:{reset}
    bud CLI

  {bold}Flags:{reset}
    --log  {dim}specify the logger{reset}

  {bold}Commands:{reset}
    build  {dim}build your application{reset}
    run    {dim}run your application{reset}

`)
}

func TestEmptyUsage(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("bud", "bud CLI").Writer(actual)
	cli.Flag("log", "").Bool(nil)
	cli.Command("run", "")
	ctx := context.Background()
	err := cli.Parse(ctx, "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} bud {dim}[flags]{reset} {dim}[command]{reset}

  {bold}Description:{reset}
    bud CLI

  {bold}Flags:{reset}
    --log

  {bold}Commands:{reset}
    run

`)
}

func TestSubHelpShort(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("bud", "bud CLI").Writer(actual)
	cli.Flag("log", "specify the logger").Short('L').Bool(nil).Default(false)
	cli.Flag("debug", "set the debugger").Bool(nil).Default(true)
	var trace []string
	cli.Run(func(ctx context.Context) error {
		trace = append(trace, "bud")
		return nil
	})
	{
		sub := cli.Command("run", "run your application")
		sub.Run(func(ctx context.Context) error {
			trace = append(trace, "run")
			return nil
		})
	}
	{
		sub := cli.Command("build", "build your application")
		sub.Run(func(ctx context.Context) error {
			trace = append(trace, "build")
			return nil
		})
	}
	ctx := context.Background()
	err := cli.Parse(ctx, "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} bud {dim}[flags]{reset} {dim}[command]{reset}

  {bold}Description:{reset}
    bud CLI

  {bold}Flags:{reset}
    -L, --log  {dim}specify the logger{reset}
    --debug    {dim}set the debugger{reset}

  {bold}Commands:{reset}
    build  {dim}build your application{reset}
    run    {dim}run your application{reset}

`)
}

func TestArgString(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var arg string
	cli.Arg("arg", "arg string").String(&arg)
	ctx := context.Background()
	err := cli.Parse(ctx, "cool")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(arg, "cool")
	isEqual(t, actual.String(), ``)
}

func TestArgStringDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var arg string
	cli.Arg("arg", "arg string").String(&arg).Default("default")
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(arg, "default")
	isEqual(t, actual.String(), ``)
}

func TestArgStringRequired(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var arg string
	cli.Arg("arg", "arg string").String(&arg)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), "missing <arg>")
}

func TestSubArgString(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var arg string
	cli.Command("build", "build command")
	cli.Command("run", "run command")
	cli.Arg("arg", "arg string").String(&arg)
	ctx := context.Background()
	err := cli.Parse(ctx, "deploy")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(arg, "deploy")
	isEqual(t, actual.String(), ``)
}

// TestInterrupt tests interrupts canceling context. It spawns a copy of itself
// to run a subcommand. I learned this trick from Mitchell Hashimoto's excellent
// "Advanced Testing with Go" talk. We use stdout to synchronize between the
// process and subprocess.
func TestInterrupt(t *testing.T) {
	is := is.New(t)
	if value := os.Getenv("TEST_INTERRUPT"); value == "" {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		// Ignore -test.count otherwise this will continue recursively
		var args []string
		for _, arg := range os.Args[1:] {
			if strings.HasPrefix(arg, "-test.count=") {
				continue
			}
			args = append(args, arg)
		}
		cmd := exec.CommandContext(ctx, os.Args[0], append(args, "-test.v=true", "-test.run=^TestInterrupt$")...)
		cmd.Env = append(os.Environ(), "TEST_INTERRUPT=1")
		stdout, err := cmd.StdoutPipe()
		is.NoErr(err)
		cmd.Stderr = os.Stderr
		is.NoErr(cmd.Start())
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "ready" {
				break
			}
		}
		cmd.Process.Signal(os.Interrupt)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "cancelled" {
				break
			}
		}
		if err := cmd.Wait(); err != nil {
			is.True(errors.Is(err, context.Canceled))
		}
		return
	}
	cli := cli.New("cli", "cli command")
	cli.Run(func(ctx context.Context) error {
		os.Stdout.Write([]byte("ready\n"))
		<-ctx.Done()
		os.Stdout.Write([]byte("cancelled\n"))
		return nil
	})
	ctx := context.Background()
	if err := cli.Parse(ctx); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		is.NoErr(err)
	}
}

// TODO: example support

func TestArgsStrings(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var args []string
	cli.Command("build", "build command")
	cli.Command("run", "run command")
	cli.Args("custom", "custom strings").Strings(&args)
	ctx := context.Background()
	err := cli.Parse(ctx, "new", "view")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(len(args), 2)
	is.Equal(args[0], "new")
	is.Equal(args[1], "view")
	isEqual(t, actual.String(), ``)
}

func TestUsageError(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cmd := cli.New("cli", "cli command").Writer(actual)
	cmd.Run(func(ctx context.Context) error {
		called++
		return cli.Usage()
	})
	ctx := context.Background()
	err := cmd.Parse(ctx)
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} cli

  {bold}Description:{reset}
    cli command

`)
}

func TestIdempotent(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("cli", "cli command").Writer(actual)
	var f1 string
	cmd := cli.Command("run", "run command")
	cmd.Flag("f1", "cli flag").Short('f').String(&f1)
	var f2 string
	cmd.Flag("f2", "cli flag").String(&f2)
	var f3 string
	cmd.Flag("f3", "cli flag").String(&f3)
	ctx := context.Background()
	args := []string{"run", "--f1=a", "--f2=b", "--f3", "c"}
	err := cli.Parse(ctx, args...)
	is.NoErr(err)
	is.Equal(f1, "a")
	is.Equal(f2, "b")
	is.Equal(f3, "c")
	f1 = ""
	f2 = ""
	f3 = ""
	err = cli.Parse(ctx, args...)
	is.NoErr(err)
	is.Equal(f1, "a")
	is.Equal(f2, "b")
	is.Equal(f3, "c")
}

func TestManualHelp(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("cli", "cli command").Writer(actual)
	var help bool
	var dir string
	cli.Flag("help", "help menu").Short('h').Bool(&help).Default(false)
	cli.Flag("chdir", "change directory").Short('C').String(&dir)
	called := 0
	cli.Run(func(ctx context.Context) error {
		is.Equal(help, true)
		is.Equal(dir, "somewhere")
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "--help", "--chdir", "somewhere")
	is.NoErr(err)
	is.Equal(actual.String(), "")
	is.Equal(called, 1)
}

func TestManualHelpUsage(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cmd := cli.New("cli", "cli command").Writer(actual)
	var help bool
	var dir string
	cmd.Flag("help", "help menu").Short('h').Bool(&help).Default(false)
	cmd.Flag("chdir", "change directory").Short('C').String(&dir)
	called := 0
	cmd.Run(func(ctx context.Context) error {
		is.Equal(help, true)
		is.Equal(dir, "somewhere")
		called++
		return cli.Usage()
	})
	ctx := context.Background()
	err := cmd.Parse(ctx, "--help", "--chdir", "somewhere")
	is.NoErr(err)
	is.Equal(called, 1)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} cli {dim}[flags]{reset}

  {bold}Description:{reset}
    cli command

  {bold}Flags:{reset}
    -C, --chdir  {dim}change directory{reset}
    -h, --help   {dim}help menu{reset}

`)
}

func TestAfterRun(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("cli", "cli command").Writer(actual)
	called := 0
	var ctx context.Context
	cli.Run(func(c context.Context) error {
		called++
		ctx = c
		return nil
	})
	err := cli.Parse(context.Background())
	is.NoErr(err)
	is.Equal(called, 1)
	select {
	case <-ctx.Done():
		is.Fail() // Context shouldn't have been cancelled
	default:
	}
}

func TestArgsClearSlice(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	args := []string{"a", "b"}
	cli.Args("custom", "custom strings").Strings(&args)
	ctx := context.Background()
	err := cli.Parse(ctx, "c", "d")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(len(args), 2)
	is.Equal(args[0], "c")
	is.Equal(args[1], "d")
	isEqual(t, actual.String(), ``)
}

func TestArgClearMap(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	args := map[string]string{"a": "a"}
	cli.Arg("custom", "custom string map").StringMap(&args)
	ctx := context.Background()
	err := cli.Parse(ctx, "b:b")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(len(args), 1)
	is.Equal(args["b"], "b")
	isEqual(t, actual.String(), ``)
}

func TestFlagClearSlice(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	args := []string{"a", "b"}
	cli.Flag("f", "flag").Strings(&args)
	ctx := context.Background()
	err := cli.Parse(ctx, "-f", "c", "-f", "d")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(len(args), 2)
	is.Equal(args[0], "c")
	is.Equal(args[1], "d")
	isEqual(t, actual.String(), ``)
}

func TestFlagClearMap(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	args := map[string]string{"a": "a"}
	cli.Flag("f", "flag").StringMap(&args)
	ctx := context.Background()
	err := cli.Parse(ctx, "-f", "b:b")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(len(args), 1)
	is.Equal(args["b"], "b")
	isEqual(t, actual.String(), ``)
}

func TestFlagCustom(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	hot := ""
	cli.Flag("hot", "hot server").Custom(func(v string) error {
		hot = v
		return nil
	})
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "--hot=:35729")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(hot, ":35729")
}

func TestFlagCustomError(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Flag("hot", "hot server").Custom(func(v string) error {
		return fmt.Errorf("unable to parse")
	})
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "--hot=:35729")
	is.True(err != nil)
	is.Equal(err.Error(), `invalid value ":35729" for flag -hot: unable to parse`)
}

func TestFlagCustomMissing(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Flag("hot", "hot server").Custom(func(v string) error {
		return nil
	})
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), `missing --hot`)
}

func TestFlagCustomMissingDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	hot := ""
	cli.Flag("hot", "hot server").Custom(func(v string) error {
		hot = v
		return nil
	}).Default(":35729")
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(hot, ":35729")
}

func TestFlagCustomDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	hot := ""
	cli.Flag("hot", "hot server").Custom(func(v string) error {
		hot = v
		return nil
	}).Default(":35729")
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "--hot=false")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(hot, "false")
}

func TestFlagOptionalString(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s *string
	cli.Flag("s", "string").Optional().String(&s)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "--s", "foo")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(*s, "foo")
}

func TestFlagOptionalStringDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s *string
	cli.Flag("s", "string").Optional().String(&s).Default("foo")
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(*s, "foo")
}

func TestFlagOptionalStringNil(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s *string
	cli.Flag("s", "string").Optional().String(&s)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.True(s == nil)
}

func TestFlagOptionalBoolTrue(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var b *bool
	cli.Flag("b", "bool").Optional().Bool(&b)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "--b")
	is.NoErr(err)
	is.Equal(1, called)
	is.True(*b)
}

func TestFlagOptionalBoolFalse(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var b *bool
	cli.Flag("b", "bool").Optional().Bool(&b)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "--b=false")
	is.NoErr(err)
	is.Equal(1, called)
	is.True(!*b)
}

func TestFlagOptionalBoolDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var b *bool
	cli.Flag("b", "bool").Optional().Bool(&b).Default(true)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.True(*b)
}

func TestFlagOptionalBoolNil(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var b *bool
	cli.Flag("b", "bool").Optional().Bool(&b)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(b, nil)
}

func TestFlagOptionalInt(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var i *int
	cli.Flag("i", "int").Optional().Int(&i)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "--i=1")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(*i, 1)
}

func TestFlagOptionalIntDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var i *int
	cli.Flag("i", "int").Optional().Int(&i).Default(1)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(*i, 1)
}

func TestFlagOptionalIntNil(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var i *int
	cli.Flag("i", "int").Optional().Int(&i)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(i, nil)
}

func TestArgOptionalString(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s *string
	cli.Arg("s", "string arg").Optional().String(&s)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "foo")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(*s, "foo")
}

func TestArgOptionalStringDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s *string
	cli.Arg("s", "string arg").Optional().String(&s).Default("foo")
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(*s, "foo")
}

func TestArgOptionalStringNil(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s *string
	cli.Arg("s", "string arg").Optional().String(&s)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.True(s == nil)
}

func TestArgOptionalBoolTrue(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var b *bool
	cli.Arg("b", "bool arg").Optional().Bool(&b)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "true")
	is.NoErr(err)
	is.Equal(1, called)
	is.True(*b)
}

func TestArgOptionalBoolFalse(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var b *bool
	cli.Arg("b", "bool arg").Optional().Bool(&b)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "false")
	is.NoErr(err)
	is.Equal(1, called)
	is.True(!*b)
}

func TestArgOptionalBoolDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var b *bool
	cli.Arg("b", "bool arg").Optional().Bool(&b).Default(true)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.True(*b)
}

func TestArgOptionalBoolNil(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var b *bool
	cli.Arg("b", "bool arg").Optional().Bool(&b)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(b, nil)
}

func TestArgOptionalInt(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var i *int
	cli.Arg("i", "int arg").Optional().Int(&i)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "1")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(*i, 1)
}

func TestArgOptionalIntDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var i *int
	cli.Arg("i", "int arg").Optional().Int(&i).Default(1)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(*i, 1)
}

func TestArgOptionalIntNil(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var i *int
	cli.Arg("i", "int arg").Optional().Int(&i)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(i, nil)
}

func TestFlagOptionalStrings(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s []string
	cli.Flag("s", "s").Optional().Strings(&s)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "--s=foo", "--s=bar")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(s, []string{"foo", "bar"})
}

func TestFlagOptionalStringsDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s []string
	cli.Flag("s", "s").Optional().Strings(&s).Default("foo", "bar")
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(s, []string{"foo", "bar"})
}

func TestFlagOptionalStringsEmpty(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s []string
	cli.Flag("s", "s").Optional().Strings(&s)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(s, []string{})
}

func TestArgsOptionalStrings(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s []string
	cli.Args("s", "strings args").Optional().Strings(&s)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "foo", "bar")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(s, []string{"foo", "bar"})
}

func TestArgsOptionalStringsDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s []string
	cli.Args("s", "strings args").Optional().Strings(&s).Default("foo", "bar")
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(s, []string{"foo", "bar"})
}

func TestArgsOptionalStringsEmpty(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s []string
	cli.Args("s", "strings args").Optional().Strings(&s)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(s, []string{})
}

func TestArgsRequiredStringsMissing(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	var s []string
	cli.Args("strings", "strings args").Strings(&s)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.True(err != nil)
	is.Equal(err.Error(), "missing <strings...>")
	is.Equal(0, called)
}

func TestUsageNestCommandArg(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	var path string
	called := 0
	cli := cli.New("bud", "bud cli").Writer(actual)
	{
		cli := cli.Command("fs", "filesystem tools")
		{
			cli := cli.Command("cat", "cat a file")
			cli.Arg("path", "path string").String(&path)
			cli.Run(func(ctx context.Context) error {
				called++
				return nil
			})
		}
	}
	ctx := context.Background()
	err := cli.Parse(ctx, "fs", "cat", "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} bud fs cat {dim}<path>{reset}

  {bold}Description:{reset}
    cat a file

  {bold}Args:{reset}
    <path>  {dim}path string{reset}

`)
}

func TestHiddenCommand(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Command("foo", "foo command").Hidden()
	cli.Command("bar", "bar command")
	ctx := context.Background()
	err := cli.Parse(ctx, "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} cli {dim}[command]{reset}

  {bold}Description:{reset}
    cli command

  {bold}Commands:{reset}
    bar  {dim}bar command{reset}

`)
}

func TestHiddenCommandRunnable(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "cli command").Writer(actual)
	cmd := cli.Command("foo", "foo command").Hidden()
	cmd.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	ctx := context.Background()
	err := cli.Parse(ctx, "foo")
	is.NoErr(err)
	is.Equal(1, called)
}

func TestAdvancedCommand(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("cli", "cli command").Writer(actual)
	cli.Command("foo", "foo command")
	cli.Command("bar", "bar command").Advanced()
	ctx := context.Background()
	err := cli.Parse(ctx, "-h")
	is.NoErr(err)
	isEqual(t, actual.String(), `
  {bold}Usage:{reset}
    {dim}${reset} cli {dim}[command]{reset}

  {bold}Description:{reset}
    cli command

  {bold}Commands:{reset}
    foo  {dim}foo command{reset}

  {bold}Advanced Commands:{reset}
    bar  {dim}bar command{reset}

`)
}

type appCmd struct {
	Chdir string
	Embed bool
}

type newCmd struct {
	Dir    string
	Minify bool
}

// Run new
func (c *newCmd) Run(ctx context.Context) error {
	os.Stdout.WriteString("running new on " + c.Dir)
	return nil
}

func ExampleCLI() {
	cmd := &appCmd{}
	cli := cli.New("app", "your awesome cli").Writer(os.Stderr)
	cli.Flag("chdir", "change the dir").Short('C').String(&cmd.Chdir).Default(".")
	cli.Flag("embed", "embed the code").Bool(&cmd.Embed).Default(false)

	{ // new <dir>
		cmd := &newCmd{}
		cli := cli.Command("new", "create a new project")
		cli.Arg("dir", "directory to scaffold in").String(&cmd.Dir)
		cli.Run(cmd.Run)
	}

	ctx := context.Background()
	cli.Parse(ctx, "new", ".")
	// Output:
	// running new on .
}

func TestFlagsAnywhere(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	var dir string
	var src string
	var path string
	called := 0
	cli := cli.New("bud", "bud cli").Writer(actual)
	cli.Flag("chdir", "change the dir").Short('C').String(&dir).Default(".")
	{
		cli := cli.Command("fs", "filesystem tools")
		cli.Flag("src", "source directory").String(&src)
		{
			cli := cli.Command("cat", "cat a file")
			cli.Flag("path", "path to file").String(&path)
			cli.Run(func(ctx context.Context) error {
				called++
				return nil
			})
		}
	}
	ctx := context.Background()
	err := cli.Parse(ctx, "fs", "cat", "-C", "cool", "--src", "http://url.com", "--path", "mypath")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(dir, "cool")
	is.Equal(src, "http://url.com")
	is.Equal(path, "mypath")
}

func TestUnexpectedArg(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	var dir string
	var src string
	var path string
	called := 0
	cmd := cli.New("bud", "bud cli").Writer(actual)
	cmd.Flag("chdir", "change the dir").Short('C').String(&dir).Default(".")
	{
		cmd := cmd.Command("fs", "filesystem tools")
		cmd.Flag("src", "source directory").String(&src)
		{
			cmd := cmd.Command("cat", "cat a file")
			cmd.Flag("path", "path to file").String(&path)
			cmd.Run(func(ctx context.Context) error {
				called++
				return nil
			})
		}
	}
	ctx := context.Background()
	err := cmd.Parse(ctx, "fs:cat", "-C", "cool", "--src", "http://url.com", "--path", "mypath")
	is.True(err != nil)
	is.True(errors.Is(err, cli.ErrInvalidInput))
	is.Equal(err.Error(), `cli: invalid input with unxpected arg "fs:cat"`)
}

func TestFlagEnum(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag string
	cli.Flag("flag", "cli flag").Enum(&flag, "a", "b", "c")
	ctx := context.Background()
	err := cli.Parse(ctx, "--flag", "a")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(flag, "a")
	isEqual(t, actual.String(), ``)
}

func TestFlagEnumDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag string
	cli.Flag("flag", "cli flag").Enum(&flag, "a", "b", "c").Default("b")
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(flag, "b")
	isEqual(t, actual.String(), ``)
}

func TestFlagEnumRequired(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag string
	cli.Flag("flag", "cli flag").Enum(&flag)
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), "missing --flag")
}

func TestFlagEnumInvalid(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var flag string
	cli.Flag("flag", "cli flag").Enum(&flag, "a", "b", "c")
	ctx := context.Background()
	err := cli.Parse(ctx, "--flag", "d")
	is.True(err != nil)
	is.Equal(err.Error(), `--flag "d" must be either "a", "b" or "c"`)
}

func TestArgEnum(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var arg string
	cli.Arg("arg", "enum arg").Enum(&arg, "a", "b", "c")
	ctx := context.Background()
	err := cli.Parse(ctx, "a")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(arg, "a")
	isEqual(t, actual.String(), ``)
}

func TestArgEnumDefault(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var arg string
	cli.Arg("arg", "enum arg").Enum(&arg, "a", "b", "c").Default("b")
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(arg, "b")
	isEqual(t, actual.String(), ``)
}

func TestArgEnumRequired(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var arg string
	cli.Arg("arg", "enum arg").Enum(&arg, "a", "b", "c")
	ctx := context.Background()
	err := cli.Parse(ctx)
	is.True(err != nil)
	is.Equal(err.Error(), "missing <arg>")
}

func TestArgEnumInvalid(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	called := 0
	cli := cli.New("cli", "desc").Writer(actual)
	cli.Run(func(ctx context.Context) error {
		called++
		return nil
	})
	var arg string
	cli.Arg("arg", "enum arg").Enum(&arg, "a", "b", "c")
	ctx := context.Background()
	err := cli.Parse(ctx, "d")
	is.True(err != nil)
	is.Equal(err.Error(), `<arg> "d" must be either "a", "b" or "c"`)
}

func TestColonBased(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	var dir string
	var src string
	var path string
	called := 0
	cli := cli.New("bud", "bud cli").Writer(actual)
	cli.Flag("chdir", "change the dir").Short('C').String(&dir).Default(".")
	{
		cli := cli.Command("fs:cat", "cat a file")
		cli.Flag("src", "source directory").String(&src)
		cli.Flag("path", "path to file").String(&path)
		cli.Run(func(ctx context.Context) error {
			called++
			return nil
		})
	}
	{
		cli := cli.Command("fs:list", "list a directory")
		cli.Flag("src", "source directory").String(&src)
		cli.Flag("path", "path to directory").String(&path)
		cli.Run(func(ctx context.Context) error {
			called++
			return nil
		})
	}
	ctx := context.Background()
	err := cli.Parse(ctx, "-C", "cool", "fs:cat", "--src=http://url.com", "--path", "mypath")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(dir, "cool")
	is.Equal(src, "http://url.com")
	is.Equal(path, "mypath")
	err = cli.Parse(ctx, "-C", "cool", "fs:list", "--src=http://url.com", "--path", "mypath")
	is.NoErr(err)
	is.Equal(2, called)
	is.Equal(dir, "cool")
	is.Equal(src, "http://url.com")
	is.Equal(path, "mypath")
}

func TestFindAndChange(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	cli := cli.New("cli", "desc").Writer(actual)
	called := []string{}
	cmd := cli.Command("a", "a command")
	cmd.Run(func(ctx context.Context) error {
		called = append(called, "a")
		return nil
	})
	cmd = cmd.Command("b", "b command")
	cmd.Run(func(ctx context.Context) error {
		called = append(called, "b1")
		return nil
	})
	cmd, err := cli.Find("a")
	is.NoErr(err)
	cmd.Run(func(ctx context.Context) error {
		called = append(called, "a2")
		return nil
	})

	// Change a
	cmd, err = cli.Find("a")
	is.NoErr(err)
	cmd.Run(func(ctx context.Context) error {
		called = append(called, "a2")
		return nil
	})
	ctx := context.Background()
	called = []string{}
	err = cli.Parse(ctx, "a")
	is.NoErr(err)
	is.Equal(called, []string{"a2"})

	// Change a b
	cmd, err = cli.Find("a", "b")
	is.NoErr(err)
	cmd.Run(func(ctx context.Context) error {
		called = append(called, "b2")
		return nil
	})
	called = []string{}
	err = cli.Parse(ctx, "a", "b")
	is.NoErr(err)
	is.Equal(called, []string{"b2"})
}

func TestFindNotFound(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	app := cli.New("cli", "desc").Writer(actual)
	called := []string{}
	cmd := app.Command("a", "a command")
	cmd.Run(func(ctx context.Context) error {
		called = append(called, "a")
		return nil
	})
	cmd = app.Command("b", "b command")
	cmd.Run(func(ctx context.Context) error {
		called = append(called, "b1")
		return nil
	})
	cmd, err := app.Find("a", "c")
	is.True(errors.Is(err, cli.ErrCommandNotFound))
	is.Equal(cmd, nil)
}

func TestOutOfOrderFlags(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	var dir string
	var src string
	var path string
	var sync bool
	called := 0
	cli := cli.New("bud", "bud cli").Writer(actual)
	cli.Flag("chdir", "change the dir").Short('C').String(&dir).Default(".")
	{
		cli := cli.Command("fs:cat", "cat a file")
		cli.Flag("src", "source directory").String(&src)
		cli.Flag("path", "path to file").String(&path)
		cli.Run(func(ctx context.Context) error {
			called++
			return nil
		})
	}
	{
		cli := cli.Command("fs:list", "list a directory")
		cli.Flag("src", "source directory").String(&src)
		cli.Flag("path", "path to directory").String(&path)
		cli.Run(func(ctx context.Context) error {
			called++
			return nil
		})
	}
	{
		cli := cli.Command("fs:cp", "cp a directory")
		cli.Arg("from", "from directory").String(&src)
		cli.Arg("to", "to directory").String(&path)
		cli.Flag("sync", "sync the directory").Bool(&sync)
		cli.Run(func(ctx context.Context) error {
			called++
			return nil
		})
	}
	ctx := context.Background()
	err := cli.Parse(ctx, "fs:cat", "--src=http://url.com", "--path", "mypath", "-C", "cool")
	is.NoErr(err)
	is.Equal(1, called)
	is.Equal(dir, "cool")
	is.Equal(src, "http://url.com")
	is.Equal(path, "mypath")
	err = cli.Parse(ctx, "fs:list", "--src=http://url.com", "-C", "cool", "--path", "mypath")
	is.NoErr(err)
	is.Equal(2, called)
	is.Equal(dir, "cool")
	is.Equal(src, "http://url.com")
	is.Equal(path, "mypath")
	err = cli.Parse(ctx, "fs:cp", "some-from", "some-to", "-C", "cool", "--sync")
	is.NoErr(err)
	is.Equal(3, called)
	is.Equal(src, "some-from")
	is.Equal(path, "some-to")
	is.Equal(dir, "cool")
	is.Equal(sync, true)
}

func TestFlagsConflictPanic(t *testing.T) {
	is := is.New(t)
	actual := new(bytes.Buffer)
	var chdir string
	var copy bool
	called := 0

	cli := cli.New("bud", "bud cli").Writer(actual)
	cli.Flag("chdir", "change the dir").Short('C').String(&chdir).Default(".")

	cmd := cli.Command("sub", "subcommand")
	cmd.Flag("copy", "copy flag").Short('C').Bool(&copy)
	cmd.Run(func(ctx context.Context) error {
		called++
		return nil
	})

	ctx := context.Background()
	err := cli.Parse(ctx, "-C", "dir", "sub", "--copy")
	is.True(err != nil)
	is.Equal(err.Error(), `cli: invalid input "bud sub" command contains a duplicate flag "-C"`)
}
