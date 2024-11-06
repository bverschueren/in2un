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
package reader

import (
	"testing"
)

func TestiConfigResourceRegex(t *testing.T) {
	tests := []struct {
		name          string
		resourceGroup string
		resourceName  string
		namespace     string
		containerName string
		previous      bool
		expected      string
	}{
		{
			name:          "return regex for namespaced pod w/o recourceName",
			resourceGroup: "pod",
			namespace:     "openshift-ingress-operator",
			resourceName:  "",
			expected:      `'^config(/storage)?/pods?/openshift-ingress-operator/[a-z0-9\.\-]+(.json|/[a-z0-9\-]+)(.json)?$`,
		},
		{
			name:          "return regex for namespaced pod w/ recourceName",
			resourceGroup: "pod",
			namespace:     "openshift-ingress-operator",
			resourceName:  "ingress-operator-65ccf4f77c-b2hv7",
			expected:      `^config(/storage)?/pods?/openshift-ingress-operator/ingress-operator-65ccf4f77c-b2hv7(.json|/[a-z0-9\-]+)$`,
		},
		{
			name:          "return regex for pods in all namespaces",
			resourceGroup: "pod",
			namespace:     "_all_",
			resourceName:  "",
			expected:      `^config(/storage)?/pods?/[a-z0-9\-]+/[a-z0-9\.\-]+(.json|/[a-z0-9\-]+)(.json)?$`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			regex := &ResourceRegex{base: &ConfigRegex{BaseRegex: BaseRegex{resourceGroup: tc.resourceGroup, namespace: tc.namespace}}, resourceName: tc.resourceName}
			got := regex.getPart()

			if got != tc.expected {
				t.Fatalf("Expected : %v, got: %v", tc.expected, got)
			}
			t.Logf("got: %v", got)
		})
	}
}
