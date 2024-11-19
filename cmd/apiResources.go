/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/bverschueren/in2un/pkg/reader"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// apiResourcesCmd represents the apiResources command
var apiResourcesCmd = &cobra.Command{
	Use:   "api-resources",
	Args:  cobra.MaximumNArgs(0),
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		ir, err := reader.NewInsightsReader(viper.GetString("active"))
		if err != nil {
			log.Fatal(err)
		}
		found := ir.ReadResourceTypes()
		fmt.Printf("NAME\n")
		for f := range *found {
			fmt.Printf("%s\n", f)
		}
	},
}

func init() {
	RootCmd.AddCommand(apiResourcesCmd)
}
