package examples

import (
	auto_ampenable "github.com/nkvoll/innosonix-maxx/cmd/examples/auto-ampenable"
	"github.com/nkvoll/innosonix-maxx/cmd/examples/mute"
	"github.com/spf13/cobra"
)

var ExamplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Demo apps using imctl as a library",
}

func init() {
	ExamplesCmd.AddCommand(auto_ampenable.AutoAmpenableCmd)
	ExamplesCmd.AddCommand(mute.MuteCmd)
}
