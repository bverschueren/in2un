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
	"github.com/bverschueren/in2un/pkg/helpers"
)

type Regex interface {
	getPart() string
}

type BaseRegex struct {
	resourceGroup, namespace string
}

func (b *BaseRegex) getPart() string { return "" }

type ConfigRegex struct {
	BaseRegex
}

func (c *ConfigRegex) getPart() string {
	resourceGroupPart := helpers.Plural(c.resourceGroup)
	if c.resourceGroup == "all" {
		resourceGroupPart = `[a-z0-9\-]+`
	}
	reg := `^config(/storage)?/` + resourceGroupPart + `/`
	if helpers.Namespaced(c.resourceGroup) {
		if c.namespace != "" {
			if c.namespace == "_all_" {
				c.namespace = `[a-z0-9\-]+`
			}
			reg += c.namespace + `/`
		}
	}
	return reg
}

type ConditionalRegex struct {
	BaseRegex
}

func (c *ConditionalRegex) getPart() string {
	if c.namespace == "_all_" {
		c.namespace = `[a-z0-9\-]+`
	}
	return `^conditional/namespaces/` + c.namespace + `/` + helpers.Plural(c.resourceGroup) + `/`
}

type ResourceRegex struct {
	resourceName string
	base         Regex
}

func (r *ResourceRegex) getPart() string {
	reg := r.base.getPart()
	if r.resourceName != "" {
		// configmaps are expanded (namespace/cm-name/key) in insights, other resources are json (namespace/obj-name.json)
		reg += r.resourceName + `(.json|/[a-z0-9\-\.]+)$`
	} else {
		reg += `[a-z0-9\.\-]+(.json|/[a-z0-9\-\.]+)(.json)?$`
	}
	return reg
}

type LogRegex struct {
	resourceName, containerName string
	previous                    bool
	base                        Regex
}

func (r *LogRegex) getPart() string {
	reg := r.base.getPart()
	// logs require a specific resourceName
	if r.resourceName == "" {
		return ""
	} else {
		if r.containerName == "" {
			reg += `logs/` + r.resourceName + `/[a-z0-9\.\-]+`
		} else {
			reg += `logs/` + r.resourceName + `/` + r.containerName
		}
	}
	if r.previous {
		reg += `_previous`
	} else {
		reg += `_current`
	}
	return reg + `.log`
}
