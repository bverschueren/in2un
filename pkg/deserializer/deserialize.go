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
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const MissingTypeMetaFieldValue = "DUMMY"

var ToBeFixedTypeMetaFields = []string{"kind", "apiVersion"}

func JsonToUnstructed(raw []byte) (*unstructured.Unstructured, error) {
	item := &unstructured.Unstructured{}
	// First, try to unmarshal the raw json into an unstructured
	if err := item.UnmarshalJSON(raw); err != nil {
		// insights removes several typeMeta fields (Kind, apiVersion), eg:
		// https://github.com/openshift/insights-operator/blob/master/docs/insights-archive-sample/config/pod/openshift-insights/insights-operator-65bcbd8bbf-n5xcr.json
		// this causes unmarshal to fail, so try to insert dummy values for these fields and retry to unmarshal
		if fixed, fixErr := insertTypeMeta(raw); fixErr != nil {
			return nil, fixErr
		} else {
			log.Trace("Retrying to unmarshal after fixing missing TypeMeta fields")
			if retryErr := item.UnmarshalJSON(fixed); retryErr != nil {
				return nil, fmt.Errorf("error when retrying to unmarshal into unstructured: %v", retryErr)
			}
			return item, nil
		}
		return nil, fmt.Errorf("Error when trying to unmarshal into unstructured: %v", err)
	}
	return item, nil
}

// Certain resources are stripped off their Kind and apiVersion fields.
// In order to marshal those into Unstructured, add dummy Kind and apiVersions to these definitions
func insertTypeMeta(raw []byte) ([]byte, error) {
	var marshalled map[string]interface{}
	if err := json.Unmarshal(raw, &marshalled); err != nil {
		return nil, fmt.Errorf("error when trying to unmarshal json: %v", err)
	}
	for _, field := range ToBeFixedTypeMetaFields {
		// TODO: use json.patch instead of (un)marshalling
		if _, ok := marshalled[field]; !ok {
			marshalled[field] = MissingTypeMetaFieldValue
		}
	}
	newData, jsonMarshalErr := json.Marshal(&marshalled)
	if jsonMarshalErr != nil {
		return nil, fmt.Errorf("error when trying to marshal json: %v", jsonMarshalErr)
	}
	return newData, nil
}
