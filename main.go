package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/jhunt/go-ansi"
)

var cli = struct {
	Character  CharacterCmd `cmd:"" name:"character" help:"manage characters in the game" aliases:"char"`
	Roll       RollCmd      `cmd:"" name:"roll" help:"roll dice for characters"`
	ConfigPath string       `name:"config" short:"C" help:"path to game config file"`
}{}

func main() {
	cliCtx := kong.Parse(&cli,
		kong.Description("ACV Database Manager"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact:             true,
			NoExpandSubcommands: true,
			FlagsLast:           true,
		}),
	)
	if cli.ConfigPath == "" {
		if os.Getenv("HOME") == "" {
			bail(fmt.Errorf("no config path given and HOME envvar not set"), "")
		}

		cli.ConfigPath = fmt.Sprintf("%s/.p2roll", os.Getenv("HOME"))
	}

	bindConfig(cliCtx, cli.ConfigPath)

	err := cliCtx.Run()
	if err != nil {
		bail(err, cliCtx.Command())
	}
}

func bindConfig(ctx *kong.Context, filepath string) {
	cfg, err := LoadConfig(filepath)
	if err != nil {
		bail(fmt.Errorf("opening config: %s", err), "")
	}

	ctx.Bind(cfg)
}

func bail(err error, command string) {
	ansi.Fprintf(os.Stderr, " @R{!! Error running `%s': %s}\n", command, err)
	os.Exit(1)
}
