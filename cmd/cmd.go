package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/nkvoll/innosonix-maxx/cmd/common"
	"github.com/nkvoll/innosonix-maxx/cmd/examples"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		// configure logging
		switch logFormat {
		case "text":
			log.SetFormatter(&log.TextFormatter{})
		case "json":
			log.SetFormatter(&log.JSONFormatter{})
		default:
			log.Fatalf("Unknown log format: %s", logFormat)
		}

		ll, err := log.ParseLevel(logLevel)
		if err != nil {
			return err
		}
		log.SetLevel(ll)

		log.SetOutput(os.Stdout)

		// configure viper for fetching arguments from environment variables
		// enable using dashed notation in flags and underscores in env
		viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
		viper.SetEnvPrefix("INNOSONIX")

		viper.AutomaticEnv()

		if err := viper.BindPFlags(cmd.Flags()); err != nil {
			return fmt.Errorf("failed to bind flags: %w", err)
		}

		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if viper.IsSet(f.Name) && viper.GetString(f.Name) != "" {
				cmd.Flags().Set(f.Name, viper.GetString(f.Name))
			}
		})

		return nil
	},
}

var (
	logLevel  string
	logFormat string
)

func init() {
	RootCmd.AddCommand(examples.ExamplesCmd)

	pf := RootCmd.PersistentFlags()
	pf.StringVar(&logLevel, "log-level", "info", fmt.Sprintf("logging level, supported: %s", log.AllLevels))
	pf.StringVar(&logFormat, "log-format", "text", "logging format, supported: [text, json]")

	pf.StringVar(&common.Addr, "addr", "", "amplifier address")
	RootCmd.MarkPersistentFlagRequired("addr")

	pf.StringVar(&common.Token, "token", "", "token to use for the REST API calls")
	RootCmd.MarkPersistentFlagRequired("token")
}
