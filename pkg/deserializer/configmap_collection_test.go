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
package deserializer

import (
	"errors"
	"reflect"
	"testing"
)

func TestUpsert(t *testing.T) {
	tests := []struct {
		name, namespace, cmName, key, value string
		original                            *ConfigMapData
		expected                            *ConfigMapData
	}{
		{
			name:      "insert new configmap in new collection",
			namespace: "kube-system",
			cmName:    "cluster-config-v1",
			key:       "install-config",
			value:     "value",
			original:  NewConfigMapData(),
			expected: &ConfigMapData{
				data: collector{
					"kube-system": {
						"cluster-config-v1": {
							"install-config": "value",
						},
					},
				},
			},
		},
		{
			name:      "update existing configmap in collection",
			namespace: "kube-system",
			cmName:    "cluster-config-v1",
			key:       "new-key",
			value:     "new-value",
			original: &ConfigMapData{
				data: collector{
					"kube-system": {
						"cluster-config-v1": {
							"install-config": "value",
						},
					},
				},
			},
			expected: &ConfigMapData{
				data: collector{
					"kube-system": {
						"cluster-config-v1": {
							"install-config": "value",
							"new-key":        "new-value",
						},
					},
				},
			},
		},
		{
			name:      "add configmap in same namespace",
			namespace: "kube-system",
			cmName:    "new-configmap",
			key:       "new-key",
			value:     "new-value",
			original: &ConfigMapData{
				data: collector{
					"kube-system": {
						"cluster-config-v1": {
							"install-config": "value",
						},
					},
				},
			},
			expected: &ConfigMapData{
				data: collector{
					"kube-system": {
						"cluster-config-v1": {
							"install-config": "value",
						},
						"new-configmap": {
							"new-key": "new-value",
						},
					},
				},
			},
		},
		{
			name:      "add configmap in different namespace",
			namespace: "new-namespace",
			cmName:    "new-configmap",
			key:       "new-key",
			value:     "new-value",
			original: &ConfigMapData{
				data: collector{
					"kube-system": {
						"cluster-config-v1": {
							"install-config": "value",
						},
					},
				},
			},
			expected: &ConfigMapData{
				data: collector{
					"kube-system": {
						"cluster-config-v1": {
							"install-config": "value",
						},
					},
					"new-namespace": {
						"new-configmap": {
							"new-key": "new-value",
						},
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.original.Upsert(tc.namespace, tc.cmName, tc.key, tc.value)
			if !reflect.DeepEqual(tc.original, tc.expected) {
				t.Fatalf("Expected: %#v, got: %#v", tc.expected, tc.original)
			}
		})
	}
}

func TestConfigMapFromFilename(t *testing.T) {
	tests := []struct {
		name, in, expectedName, expectedNamespace, expectedKey string
		expectedErr                                            error
	}{
		{
			name:              "split configmap path correctly",
			in:                "config/configmaps/kube-system/cluster-config-v1/install-config",
			expectedName:      "cluster-config-v1",
			expectedNamespace: "kube-system",
			expectedKey:       "install-config",
			expectedErr:       nil,
		},
		{
			name:              "raise error if insufficient fields",
			in:                "config/configmaps/kube-system/cluster-config-v1/",
			expectedName:      "cluster-config-v1",
			expectedNamespace: "kube-system",
			expectedKey:       "install-config",
			expectedErr:       ErrUnknownResourcePath,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name, namespace, key, err := configMapFromFilename(tc.in)

			if err == nil {
				if name != tc.expectedName || namespace != tc.expectedNamespace || key != tc.expectedKey {
					t.Fatalf("Expected: (%s, %s, %s) got: (%s, %s, %s)", tc.expectedName, tc.expectedNamespace, tc.expectedKey, name, namespace, key)
				}
			} else {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("Expected err='%s', got err='%s'", tc.expectedErr, err)
				}
			}
		})
	}
}
