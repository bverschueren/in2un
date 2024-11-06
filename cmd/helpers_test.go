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
	"testing"
)

func TestProcessArgs(t *testing.T) {
	tests := []struct {
		name                        string
		in                          []string
		resourceGroup, resourceName string
	}{
		{
			name:          "single argument resourcegroup",
			in:            []string{"pod"},
			resourceGroup: "pod",
			resourceName:  "",
		},
		{
			name:          "single argument separated by a slash",
			in:            []string{"pod/name"},
			resourceGroup: "pod",
			resourceName:  "name",
		},
		{
			name:          "sequential argument resourcegroup/resourcename",
			in:            []string{"pod", "name"},
			resourceGroup: "pod",
			resourceName:  "name",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resourceGroup, resourceName := processArgs(tc.in)

			if (resourceGroup != tc.resourceGroup) || (resourceName != tc.resourceName) {
				t.Fatalf("Expected: resourceName=%s, resourceGroup=%s got: resourceName=%s, resourceGroup=%s", tc.resourceName, tc.resourceGroup, resourceName, resourceGroup)
			}
		})
	}
}
