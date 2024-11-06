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
	"reflect"
	"testing"
)

func TestInsertTypeMeta(t *testing.T) {
	tests := []struct {
		name     string
		in       []byte
		expected []byte
	}{
		{
			name:     "re-insert dummy kind",
			in:       []byte(`{"apiVersion":"v1","metadata":{"name":"test-pod","namespace":"test-namespace"}}`),
			expected: []byte(`{"apiVersion":"v1","kind":"` + MissingTypeMetaFieldValue + `","metadata":{"name":"test-pod","namespace":"test-namespace"}}`),
		},
		{
			name:     "re-insert dummy apiVersion",
			in:       []byte(`{"kind":"Pod","metadata":{"name":"test-pod","namespace":"test-namespace"}}`),
			expected: []byte(`{"apiVersion":"` + MissingTypeMetaFieldValue + `","kind":"Pod","metadata":{"name":"test-pod","namespace":"test-namespace"}}`),
		},
		{
			name:     "re-insert dummy kind and apiVersion",
			in:       []byte(`{"metadata":{"name":"test-pod","namespace":"test-namespace"}}`),
			expected: []byte(`{"apiVersion":"` + MissingTypeMetaFieldValue + `","kind":"` + MissingTypeMetaFieldValue + `","metadata":{"name":"test-pod","namespace":"test-namespace"}}`),
		},
		{
			name:     "do not overwrite existing kind and apiVersion",
			in:       []byte(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod","namespace":"test-namespace"}}`),
			expected: []byte(`{"apiVersion":"v1","kind":"Pod","metadata":{"name":"test-pod","namespace":"test-namespace"}}`),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := insertTypeMeta(tc.in)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tc.expected) {
				t.Fatalf("Expected : %s, got: %s", tc.expected, got)
			}
			t.Logf("got: %v", got)
		})
	}
}
