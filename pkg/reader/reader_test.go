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
	"archive/tar"
	"bytes"
	"errors"
	"io/fs"
	"log"
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type tarrable struct {
	Name string
	Body []byte
}

func generateBufferedTar(in []tarrable) *bytes.Buffer {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	for _, file := range in {
		hdr := &tar.Header{
			Name: file.Name,
			Mode: 0600,
			Size: int64(len(file.Body)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			log.Fatal(err)
		}
		if _, err := tw.Write([]byte(file.Body)); err != nil {
			log.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		log.Fatal(err)
	}
	return &buf
}

func generateUnstructuredList(u ...unstructured.Unstructured) *unstructured.UnstructuredList {
	return &unstructured.UnstructuredList{
		Object: map[string]interface{}{"kind": "List", "apiVersion": "v1"},
		Items:  u}
}

func generateUnstructuredConfigMap(name, namespace string, data map[string]string) unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.GetObjectKind().SetGroupVersionKind(u.GetObjectKind().GroupVersionKind())
	newObject := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Data: data,
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
	}
	result, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&newObject)
	if err != nil {
		log.Fatal(err)
	}
	u.SetUnstructuredContent(result)
	return *u
}

func TestResourceFromInsights(t *testing.T) {
	fakeObj := []byte(`{"metadata":{},"kind":"FakeKind","apiVersion":"Fake1.2"}`)
	expectedObj := unstructured.Unstructured{}
	_ = expectedObj.UnmarshalJSON(fakeObj)
	var files = []tarrable{
		tarrable{
			Name: "config/clusteroperator/network.json",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/clusteroperator/ingress.json",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/pod/openshift-multus/multus-sns4n.json",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/pod/openshift-multus/multus-a3e4d.json",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/pod/namespace/pod-name.json",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/configmaps/openshift-config/openshift-install/version",
			Body: []byte("v1.2.3"),
		},
		tarrable{
			Name: "config/configmaps/openshift-config/openshift-install/invoker",
			Body: []byte("user"),
		},
		tarrable{
			Name: "config/configmaps/openshift-config/dummy/key",
			Body: []byte("value"),
		},
		tarrable{
			Name: "config/configmaps/namespace/dummy/key",
			Body: []byte("value"),
		},
		tarrable{
			Name: "config/machineconfigs/00-master.json",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/storage/storageclasses/standard-csi.json",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/storage/storageclasses/csi-manila-ceph.json",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/ingress.json",
			Body: fakeObj,
		},
	}

	tests := []struct {
		name                             string
		resourceGroup                    string
		namespace                        string
		resourceName                     string
		overrideApiVersion, overrideKind string
		expected                         *unstructured.UnstructuredList
	}{
		{
			name:               "return pods within a namespace",
			resourceGroup:      "pod",
			namespace:          "openshift-multus",
			resourceName:       "",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(expectedObj, expectedObj),
		},
		{
			name:               "return pods from all namespaces",
			resourceGroup:      "pod",
			namespace:          AllNamespaceValue,
			resourceName:       "",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(expectedObj, expectedObj, expectedObj),
		},
		{
			name:               "return named pod from a namespace",
			resourceGroup:      "pod",
			namespace:          "namespace",
			resourceName:       "pod-name",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(expectedObj),
		},
		{
			name:               "return named co",
			resourceGroup:      "clusteroperator",
			namespace:          "",
			resourceName:       "network",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(expectedObj),
		},
		{
			name:               "return all co",
			resourceGroup:      "clusteroperator",
			namespace:          "",
			resourceName:       "",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(expectedObj, expectedObj),
		},
		{
			name:               "return named configmap w namespace",
			resourceGroup:      "configmap",
			namespace:          "openshift-config",
			resourceName:       "openshift-install",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(generateUnstructuredConfigMap("openshift-install", "openshift-config", map[string]string{"version": "v1.2.3", "invoker": "user"})),
		},
		{
			name:               "return all configmap w namespace",
			resourceGroup:      "configmap",
			namespace:          "openshift-config",
			resourceName:       "",
			overrideApiVersion: "",
			overrideKind:       "",
			expected: generateUnstructuredList(
				generateUnstructuredConfigMap("openshift-install", "openshift-config", map[string]string{"version": "v1.2.3", "invoker": "user"}),
				generateUnstructuredConfigMap("dummy", "openshift-config", map[string]string{"key": "value"}),
			),
		},
		{
			name:               "return all configmap across all namespaces",
			resourceGroup:      "configmap",
			namespace:          AllNamespaceValue,
			resourceName:       "",
			overrideApiVersion: "",
			overrideKind:       "",
			expected: generateUnstructuredList(
				generateUnstructuredConfigMap("openshift-install", "openshift-config", map[string]string{"version": "v1.2.3", "invoker": "user"}),
				generateUnstructuredConfigMap("dummy", "openshift-config", map[string]string{"key": "value"}),
				generateUnstructuredConfigMap("dummy", "namespace", map[string]string{"key": "value"}),
			),
		},
		{
			name:               "return machineconfig",
			resourceGroup:      "machineconfig",
			namespace:          "",
			resourceName:       "",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(expectedObj),
		},
		{
			name:               "return all storageclass",
			resourceGroup:      "storageclass",
			namespace:          "",
			resourceName:       "",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(expectedObj, expectedObj),
		},
		{
			name:               "return named storageclass",
			resourceGroup:      "storageclass",
			namespace:          "",
			resourceName:       "standard-csi",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(expectedObj),
		},
		{
			name:               "return ingress",
			resourceGroup:      "ingress",
			namespace:          "",
			resourceName:       "",
			overrideApiVersion: "",
			overrideKind:       "",
			expected:           generateUnstructuredList(expectedObj),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tw := generateBufferedTar(files)
			tr := tar.NewReader(tw)
			configRegex := NewResourceRegex(tc.resourceGroup, tc.resourceName, tc.namespace,
				NewConfigRegex(
					tc.resourceGroup,
					tc.resourceName,
					tc.namespace,
				))
			conditionalRegex := NewResourceRegex(tc.resourceGroup, tc.resourceName, tc.namespace,
				NewConditionalRegex(
					tc.resourceGroup,
					tc.resourceName,
					tc.namespace,
				),
			)
			operatorConfigRegex := NewOperatorConfigRegex(
				tc.resourceGroup,
				tc.resourceName,
			)
			got := readResources(tr, []IRegex{configRegex, conditionalRegex, operatorConfigRegex}, "", "")

			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("\nExpected: %+v,\n\t got: %+v", tc.expected, got)
			}
		})
	}
}

func TestNewInsightsReader(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expected    *InsightsReader
		expectedErr error
	}{
		{
			name:        "return an InsightsReader",
			path:        "../../testdata/fake-insights-archive",
			expected:    &InsightsReader{Reader: new(tar.Reader), Path: "testdata/fake-insights-archive"},
			expectedErr: nil,
		},
		{
			name:        "return error for non-existing inpurt file",
			path:        "../../testdata/non-existing-insights-archive",
			expected:    nil,
			expectedErr: fs.ErrNotExist,
		},
		{
			name:        "return error for non-gzip inpurt file",
			path:        "../../testdata/non-gzip-file",
			expected:    nil,
			expectedErr: ErrInvalidInsightsArchive,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewInsightsReader(tc.path)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("Expected err='%s', got err='%s'", tc.expectedErr, err)
				}
			} else {
				// path presense should suffice to verify if an instance is returned
				if got != nil && tc.expected.Path == got.Path {
					t.Fatalf("Expected: %v got: %v with", tc.expected, got)
				}

			}
		})
	}
}

func TestOpen(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		expected    *tar.Reader
		expectedErr error
	}{
		{
			name:        "return a tar.Reader",
			path:        "../../testdata/fake-insights-archive",
			expected:    &tar.Reader{},
			expectedErr: nil,
		},
		{
			name:        "return error for non-existing input file",
			path:        "../../testdata/non-existing-insights-archive",
			expected:    nil,
			expectedErr: fs.ErrNotExist,
		},
		{
			name:        "return error for non-gzip inpurt file",
			path:        "../../testdata/non-gzip-file",
			expected:    nil,
			expectedErr: ErrInvalidInsightsArchive,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := open(tc.path)

			if tc.expectedErr != nil {
				if !errors.Is(err, tc.expectedErr) {
					t.Fatalf("Expected err='%s', got err='%s'", tc.expectedErr, err)
				}
			} else {
				if got != nil {
					if reflect.TypeOf(got) != reflect.TypeOf(tc.expected) {
						t.Fatalf("Expected: %s got: %s with", reflect.TypeOf(tc.expected), reflect.TypeOf(got))
					}
				}
			}
		})
	}
}

func TestContainerAndVersionFromFilename(t *testing.T) {
	tests := []struct {
		name             string
		in               string
		containerName    string
		containerVersion string
	}{
		{
			name:             "return container name for current log",
			in:               "ingress-operator_current.log",
			containerName:    "ingress-operator",
			containerVersion: "current",
		},
		{
			name:             "return container name for previous log",
			in:               "ingress-operator_current.log",
			containerName:    "ingress-operator",
			containerVersion: "current",
		},
		{
			name:             "return container name for previous log",
			in:               "ingress-operator.log",
			containerName:    "ingress-operator",
			containerVersion: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			name, version := containerAndVersionFromFilename(tc.in)

			if name != tc.containerName || version != tc.containerVersion {
				t.Fatalf("Expected: name=%s,version=%s got: name=%s,verison=%s", tc.containerName, tc.containerVersion, name, version)
			}
		})
	}
}
