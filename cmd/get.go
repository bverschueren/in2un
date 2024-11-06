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

	log "github.com/sirupsen/logrus"

	"github.com/bverschueren/in2un/pkg/reader"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var getCmd = &cobra.Command{
	Use:  "get",
	Args: cobra.MinimumNArgs(1),
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if AllNamespaces {
			Namespace = "_all_" // TODO: export and use global (?) AllNamespaceValue
		}
	},

	Short: "Parse Insights data as generic unstructured (https://pkg.go.dev/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured) data.",
	Run: func(cmd *cobra.Command, args []string) {
		resourceGroup, resourceName := processArgs(args)
		ir, err := reader.NewInsightsReader(viper.GetString("active"))
		if err != nil {
			log.Fatal(err)
		}
		found := ir.ReadResource(resourceGroup, resourceName, Namespace)
		if len(found) == 0 {
			fmt.Printf("No resources found")
			if len(Namespace) != 0 && !AllNamespaces {
				fmt.Printf(" in %s namespace.\n", Namespace)
			}
		} else {
			fmt.Println("NAME")
			for _, f := range found {
				fmt.Printf("%s\n", f.GetName())
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.PersistentFlags().BoolVarP(&AllNamespaces, "all-namespaces", "A", false, "Set the namespace scope for this CLI request to all namespaces")
}
