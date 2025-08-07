package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	quiet   bool
	noColor bool
	logJSON bool
)

var rootCmd = &cobra.Command{
	Use:   "portscan",
	Short: "High-performance TUI port scanner",
	Long: `A blazing-fast, cross-platform port scanner with a beautiful terminal UI.
Scans thousands of ports per second with real-time visualization.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.portscan.yaml)")
	rootCmd.PersistentFlags().BoolVar(&quiet, "quiet", false, "suppress non-essential output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&logJSON, "log-json", false, "output logs in JSON format")

	rootCmd.PersistentFlags().Bool("profile", false, "enable pprof profiling")
	rootCmd.PersistentFlags().Bool("trace", false, "enable execution tracing")
	_ = rootCmd.PersistentFlags().MarkHidden("profile")
	_ = rootCmd.PersistentFlags().MarkHidden("trace")

	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("no_color", rootCmd.PersistentFlags().Lookup("no-color"))
	_ = viper.BindPFlag("log_json", rootCmd.PersistentFlags().Lookup("log-json"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName(".portscan")
	}

	viper.SetEnvPrefix("PORTSCAN")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if !quiet {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
