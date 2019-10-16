// Entry-Point to our code.

package main

import (
	"context"
	flag "github.com/spf13/pflag"
	"os"

	"github.com/google/subcommands"
)

//
// Open the named configuration file, and parse it
//
func main() {

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&dumpCmd{}, "")
	subcommands.Register(&enqueueCmd{}, "")
	subcommands.Register(&examplesCmd{}, "")
	subcommands.Register(&versionCmd{}, "")
	subcommands.Register(&workerCmd{}, "")
	subcommands.Register(&k8sEventWatcherCmd{}, "")

	flag.Parse()
	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))

}
