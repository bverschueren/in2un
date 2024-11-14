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
	"strings"

	log "github.com/sirupsen/logrus"
)

// follow kubectl logic and expect args to be either:
// a resource as single argument: "<recource-type>"
// a resource and its name as a single argument seperated by a slash: "<recource-type>/<resource-name>"
// a resource and its name as sequential arguments: "<recource-type>/<resource-name>"
func processArgs(args []string) (string, string) {
	var resourceGroup, resourceName string
	if len(args) == 1 && strings.Contains(args[0], "/") {
		parts := strings.Split(args[0], "/")
		resourceGroup = parts[0]
		resourceName = parts[1]
	} else {
		resourceGroup = args[0]
		if len(args) > 1 {
			resourceName = args[1]
		}
	}
	resourceGroup = Unalias(resourceGroup)
	return resourceGroup, resourceName
}

func Unalias(alias string) string {
	log.Debug("Using static alias map as best effort")
	aliases := map[string]string{
		"mc":  "machineconfig",
		"mcp": "machineconfigpool",
		"cm":  "configmap",
		"co":  "clusteroperator",
		"ns":  "namespace",
	}
	if unalias, ok := aliases[alias]; ok {
		return unalias
	}
	return alias
}
