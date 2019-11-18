package cmd

import (
	"github.com/pkg/errors"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

//NewRoot is for creating commandline
func NewRoot() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "app",
		Short: "app is a birthday application",
		Long: `
               A Fast and Flexible Site 
               that helps you to remind remaining days
               http://www.appholiday.com`,
		Run: func(cmd *cobra.Command, args []string) {},
	}

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.otsample.yaml)")

	return rootCmd
}

func initConfig() {
	// Don't forget to read config either from cfgFile or from home directory!
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			panic(errors.Wrap(err, "could not get HOME dir:"))
		}

		// Search config in home directory with name ".otsample" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".otsample")
	}

	if err := viper.ReadInConfig(); err != nil {
		panic(errors.Wrap(err, "Can't read config:"))
	}
}
