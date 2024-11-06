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
	"log"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type tarrable struct {
	Name string
	Body []byte
}

func createBufferedTar(in []tarrable) *bytes.Buffer {
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

func TestResourceFromInsights(t *testing.T) {
	fakeObj := []byte(`{"metadata":{},"kind":"FakeKind","apiVersion":"Fake1.2"}`)
	expectedObj := &unstructured.Unstructured{}
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
			Body: fakeObj,
		},
		tarrable{
			Name: "config/configmaps/openshift-config/openshift-install/invoker",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/configmaps/openshift-config/dummy/key",
			Body: fakeObj,
		},
		tarrable{
			Name: "config/configmaps/namespace/dummy/key",
			Body: fakeObj,
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
	}

	tests := []struct {
		name          string
		resourceGroup string
		namespace     string
		resourceName  string
		expected      []*unstructured.Unstructured
	}{
		{
			name:          "return pods within a namespace",
			resourceGroup: "pod",
			namespace:     "openshift-multus",
			resourceName:  "",
			expected:      []*unstructured.Unstructured{expectedObj, expectedObj},
		},
		{
			name:          "return pods from all namespaces",
			resourceGroup: "pod",
			namespace:     AllNamespaceValue,
			resourceName:  "",
			expected:      []*unstructured.Unstructured{expectedObj, expectedObj, expectedObj},
		},
		{
			name:          "return named pod from a namespace",
			resourceGroup: "pod",
			namespace:     "namespace",
			resourceName:  "pod-name",
			expected:      []*unstructured.Unstructured{expectedObj},
		},
		{
			name:          "return named co",
			resourceGroup: "clusteroperator",
			namespace:     "",
			resourceName:  "network",
			expected:      []*unstructured.Unstructured{expectedObj},
		},
		{
			name:          "return all co",
			resourceGroup: "clusteroperator",
			namespace:     "",
			resourceName:  "",
			expected:      []*unstructured.Unstructured{expectedObj, expectedObj},
		},
		{
			name:          "return named configmap w namespace",
			resourceGroup: "configmap",
			namespace:     "openshift-config",
			resourceName:  "openshift-install",
			expected:      []*unstructured.Unstructured{expectedObj, expectedObj},
		},
		{
			name:          "return all configmap w namespace",
			resourceGroup: "configmap",
			namespace:     "openshift-config",
			resourceName:  "",
			expected:      []*unstructured.Unstructured{expectedObj, expectedObj, expectedObj},
		},
		{
			name:          "return all configmap across all namespaces",
			resourceGroup: "configmap",
			namespace:     AllNamespaceValue,
			resourceName:  "",
			expected:      []*unstructured.Unstructured{expectedObj, expectedObj, expectedObj, expectedObj},
		},
		{
			name:          "return machineconfig",
			resourceGroup: "machineconfig",
			namespace:     "",
			resourceName:  "",
			expected:      []*unstructured.Unstructured{expectedObj},
		},
		{
			name:          "return all storageclass",
			resourceGroup: "storageclass",
			namespace:     "",
			resourceName:  "",
			expected:      []*unstructured.Unstructured{expectedObj, expectedObj},
		},
		{
			name:          "return named storageclass",
			resourceGroup: "storageclass",
			namespace:     "",
			resourceName:  "standard-csi",
			expected:      []*unstructured.Unstructured{expectedObj},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tw := createBufferedTar(files)
			tr := tar.NewReader(tw)
			got := readResources(tr, tc.resourceGroup, tc.resourceName, tc.namespace)

			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("Expected: %v, got: %+v", tc.expected, got)
			}
		})
	}
}

//func TestResourceRegex(t *testing.T) {
//	tests := []struct {
//		name          string
//		resourceGroup string
//		namespace     string
//		resourceName  string
//		expected      string
//	}{
//		{
//			name:          "return regex for crd w/o namespace",
//			resourceGroup: "crd",
//			namespace:     "",
//			resourceName:  "",
//			expected:      `[a-z]+(/storage)?/crds?/[a-z0-9\.-]+(.json|/[a-z0-9\-]+)$`,
//		},
//		{
//			name:          "return regex for pods in a namespace",
//			resourceGroup: "pod",
//			namespace:     "namespace",
//			resourceName:  "",
//			expected:      `[a-z]+(/storage)?/pods?/namespace/[a-z0-9\.-]+(.json|/[a-z0-9\-]+)$`,
//		},
//		{
//			name:          "return regex for specific pod in a namespace",
//			resourceGroup: "pod",
//			namespace:     "namespace",
//			resourceName:  "podname",
//			expected:      `[a-z]+(/storage)?/pods?/namespace/podname(.json|/[a-z0-9\-]+)$`,
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			got := resourceRegex(tc.resourceGroup, tc.namespace, tc.resourceName)
//
//			if got != tc.expected {
//				t.Fatalf("Expected : %v, got: %v", tc.expected, got)
//			}
//			t.Logf("got: %v", got)
//		})
//	}
//}

func TestResourceFilename(t *testing.T) {
	tests := []struct {
		name     string
		regex    []string
		filename string
		expected string
	}{
		{
			name:     "return filename when regex matches",
			regex:    []string{`[a-z]+/pods?/openshift-cluster-samples-operator/[a-z0-9\.-]*`},
			filename: "config/pod/openshift-cluster-samples-operator/cluster-samples-operator-7bdb9db984-2k2l9.json",
			expected: "config/pod/openshift-cluster-samples-operator/cluster-samples-operator-7bdb9db984-2k2l9.json",
		},
		{
			name:     "return empty string when regex does not match",
			regex:    []string{`[a-z]+/pods?/openshift-cluster-samples-operator/[a-z0-9\.-]*`},
			filename: "config/pod/anothernamespace/cluster-samples-operator-7bdb9db984-2k2l9.json",
			expected: "",
		},
		{
			name:     "return filename for non namespaced resource",
			regex:    []string{`[a-z]+/nodes?/[a-z0-9\.-]*`},
			filename: "config/node/bverschu-6dh2k-worker-0-67hrt.json",
			expected: "config/node/bverschu-6dh2k-worker-0-67hrt.json",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := resourceFilename(tc.regex, tc.filename)

			if got != tc.expected {
				t.Fatalf("Expected : %s, got: %s", tc.expected, got)
			}
			t.Logf("got: %v", got)
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

//func TestLogRegex(t *testing.T) {
//	tests := []struct {
//		name                                                  string
//		resourceGroup, namespace, resourceName, containerName string
//		expectedRegex                                         string
//		previous                                              bool
//	}{
//		{
//			name:          "return regex for previous logs from specified ingress-operator pod",
//			resourceGroup: "pod",
//			resourceName:  "ingress-operator-65ccf4f77c-b2hv7",
//			namespace:     "openshift-ingress-operator",
//			containerName: "ingress-operator",
//			previous:      true,
//			expectedRegex: `[a-z]+(/storage)?/pods?/openshift-ingress-operator/logs/ingress-operator-65ccf4f77c-b2hv7/ingress-operator_previous.log`,
//		},
//		{
//			name:          "return regex for current logs from specified ingress-operator pod",
//			resourceGroup: "pod",
//			resourceName:  "ingress-operator-65ccf4f77c-b2hv7",
//			namespace:     "openshift-ingress-operator",
//			containerName: "ingress-operator",
//			previous:      false,
//			expectedRegex: `[a-z]+(/storage)?/pods?/openshift-ingress-operator/logs/ingress-operator-65ccf4f77c-b2hv7/ingress-operator_current.log`,
//		},
//		{
//			name:          "return regex for kube-rbac-proxy if no container name is specified",
//			resourceGroup: "pod",
//			resourceName:  "ingress-operator-65ccf4f77c-b2hv7",
//			namespace:     "openshift-ingress-operator",
//			containerName: "",
//			previous:      false,
//			expectedRegex: `[a-z]+(/storage)?/pods?/openshift-ingress-operator/logs/ingress-operator-65ccf4f77c-b2hv7/[a-z0-9\.\-]+_current.log`,
//		},
//	}
//
//	for _, tc := range tests {
//		t.Run(tc.name, func(t *testing.T) {
//			got := logRegex(tc.resourceGroup, tc.namespace, tc.resourceName, tc.containerName, tc.previous)
//
//			if got != tc.expectedRegex {
//				t.Fatalf("Expected: '%s' got : '%s'", tc.expectedRegex, got)
//			}
//		})
//	}
//}
