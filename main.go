package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
	"go.coldcutz.net/bake/bakefile"
	"go.coldcutz.net/bake/executor"
	"go.coldcutz.net/bake/options"
	"go.coldcutz.net/go-stuff/utils"
)

func main() {
	// TODO: set log level to info unless verbose, in utils like here https://stackoverflow.com/questions/76970895/change-log-level-of-go-lang-slog-in-runtime
	ctx, log, opts, err := utils.StdSetup[options.Opts]()
	if err != nil {
		os.Exit(1)
	}

	if err := run(ctx, log, opts); err != nil {
		log.Warn("exiting with error", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *slog.Logger, opts options.Opts) error {
	contents, err := os.ReadFile(opts.BakefilePath)
	if err != nil {
		return fmt.Errorf("reading Bakefile (%q): %w", opts.BakefilePath, err)
	}
	config := bakefile.Config{}
	meta, err := toml.Decode(string(contents), &config)
	if err != nil {
		return fmt.Errorf("decoding Bakefile (%q): %w", opts.BakefilePath, err)
	}
	_ = meta
	log.Info("successfully decoded Bakefile", "bakefile", config)

	ex, err := executor.New(config, "main", opts, log)
	if err != nil {
		return fmt.Errorf("creating executor")
	}

	return ex.Exec(ctx)
}

// syntax similar to make but less shit. what are the features cases i want?
// - [ ] basic dags, with time based caching
// - [X] globstar paths
// - [ ] what you can do with builtin variables
// 	- [ ] protobuf style generation
// - [ ] parallelism
// -  more than one target?
