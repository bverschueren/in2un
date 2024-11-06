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
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/bverschueren/in2un/pkg/reader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var useCmd = &cobra.Command{
	Use:   "use",
	Args:  cobra.MinimumNArgs(1),
	Short: "Specify the insights file to read from",
	Run: func(cmd *cobra.Command, args []string) {
		active, err := reader.NewInsightsReader(args[0])
		if err != nil {
			log.Fatal(err)
		}
		configDir := filepath.Dir(ConfigFile)
		err = os.MkdirAll(configDir, 0750)
		if err != nil {
			log.Fatal(err)
		}
		viper.Set("active", active.Path)
		viper.WriteConfigAs(ConfigFile)
	},
}

func init() {
	rootCmd.AddCommand(useCmd)
}
