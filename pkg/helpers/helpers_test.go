package helpers

import "testing"

func TestNamespaced(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected bool
	}{
		{
			name:     "static check for MachineConfig",
			in:       "machineconfig",
			expected: false,
		},
		{
			name:     "static check for MachineConfigPool",
			in:       "machineconfigpool",
			expected: false,
		},
		{
			name:     "static check for ClusterOperator",
			in:       "clusteroperator",
			expected: false,
		},
		{
			name:     "static check for Nod3",
			in:       "node",
			expected: false,
		},
		{
			name:     "static check for StorageClass",
			in:       "storageclass",
			expected: false,
		},
		{
			name:     "static check for HostSubnet",
			in:       "hostsubnet",
			expected: false,
		},
		{
			name:     "static check for Pod",
			in:       "pod",
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Namespaced(tc.in)

			if got != tc.expected {
				t.Fatalf("Expected: %t, got: %t", tc.expected, got)
			}
		})
	}
}

func TestPlural(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "static pluralize for Pod",
			in:       "pod",
			expected: "pods?",
		},
		{
			name:     "static pluralize for StorageClass",
			in:       "storageclass",
			expected: "storageclasses?",
		},
		{
			name:     "static pluralize regexify already plural Pods",
			in:       "pods",
			expected: "pods?",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Plural(tc.in)

			if got != tc.expected {
				t.Fatalf("Expected: %s, got: %s", tc.expected, got)
			}
		})
	}
}

func TestUnalias(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "static unalias for MachineConfig",
			in:       "mc",
			expected: "machineconfig",
		},
		{
			name:     "static unalias for ConfigMap",
			in:       "cm",
			expected: "configmap",
		},
		{
			name:     "static unalias for ClusterOperator",
			in:       "co",
			expected: "clusteroperator",
		},
		{
			name:     "static unalias for Namespace",
			in:       "ns",
			expected: "namespace",
		},
		{
			name:     "static unalias for PersistentVolume",
			in:       "pv",
			expected: "persistentvolume",
		},
		{
			name:     "static unalias for PersistentVolumeClaim",
			in:       "pvc",
			expected: "persistentvolumeclaim",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Unalias(tc.in)

			if got != tc.expected {
				t.Fatalf("Expected: %s, got: %s", tc.expected, got)
			}
		})
	}
}
