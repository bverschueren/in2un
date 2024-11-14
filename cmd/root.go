/*
Copyright © 2024 Bram Verschueren <bverschueren@redhat.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	rootCmd = &cobra.Command{
		Use:   "in2un",
		Args:  cobra.MinimumNArgs(1),
		Short: "Parse Insights data as unstructed data or raw log lines.",
	}

	ResourceGroup, ResourceName, Namespace, Active, LogLevel string
	AllNamespaces                                            bool
	ConfigFile                                               = "$HOME/.in2un/in2un.json"
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().StringVarP(&LogLevel, "loglevel", "v", "warning", "Logging level")
	rootCmd.PersistentFlags().StringVarP(&Namespace, "namespace", "n", "", "If present, the namespace scope for this CLI request")
	rootCmd.PersistentFlags().StringVarP(&Active, "insights-file", "", "", "Insights file to read from")
	viper.BindPFlag("active", rootCmd.PersistentFlags().Lookup("active"))
}

func initConfig() {
	level, err := log.ParseLevel(LogLevel)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
	log.SetLevel(level)
	ConfigFile = os.ExpandEnv(ConfigFile)
	viper.SetConfigFile(ConfigFile)

	if err := viper.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			viper.WriteConfigAs(ConfigFile)
		} else {
			fmt.Println("Can't read config:", err)
		}
	}
}
