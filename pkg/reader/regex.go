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
	log "github.com/sirupsen/logrus"
)

type IRegex interface {
	getPart() string
}

func NewRegex(resourceGroup, resourceName, namespace string) IRegex {
	return &Regex{
		resourceGroup: resourceGroup,
		resourceName:  resourceName,
		namespace:     namespace,
	}
}

type Regex struct {
	resourceGroup, resourceName, namespace string
}

func (b *Regex) getPart() string { return "" }

type ConfigRegex struct {
	Regex
	base IRegex
}

// ConfigRegex is the "intermediate" regex: base->intermediate (e.g. ConditionalRegex) ->final
func NewConfigRegex(resourceGroup, resourceName, namespace string) IRegex {
	return &ConfigRegex{
		Regex: Regex{
			resourceGroup: resourceGroup,
			resourceName:  resourceName,
			namespace:     namespace,
		},
		base: NewRegex(
			resourceGroup,
			resourceName,
			namespace,
		),
	}
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
	log.Tracef("got part '%s'", reg)
	return reg
}

type ConditionalRegex struct {
	Regex
	base IRegex
}

// ConditionalRegex is the "intermediate" regex: base->intermediate (e.g. ConditionalRegex) ->final
func NewConditionalRegex(resourceGroup, resourceName, namespace string) IRegex {
	return &ConditionalRegex{
		Regex: Regex{
			resourceGroup: resourceGroup,
			resourceName:  resourceName,
			namespace:     namespace,
		},
		base: NewRegex(
			resourceGroup,
			resourceName,
			namespace,
		),
	}
}

func (c *ConditionalRegex) getPart() string {
	if c.namespace == "_all_" {
		c.namespace = `[a-z0-9\-]+`
	}
	reg := `^conditional/namespaces/` + c.namespace + `/` + helpers.Plural(c.resourceGroup) + `/`
	log.Tracef("got part '%s'", reg)
	return reg
}

type ResourceRegex struct {
	Regex
	base IRegex
}

// ResourceRegex is the "final" regex: base->intermediate (e.g. ConfigRegex) ->final
func NewResourceRegex(resourceGroup, resourceName, namespace string, base IRegex) IRegex {
	return &ResourceRegex{
		Regex: Regex{
			resourceGroup: resourceGroup,
			resourceName:  resourceName,
			namespace:     namespace,
		},
		base: base,
	}
}

func (r *ResourceRegex) getPart() string {
	reg := r.base.getPart()
	if r.resourceName != "" {
		// configmaps are expanded (namespace/cm-name/key) in insights, other resources are json (namespace/obj-name.json)
		reg += r.resourceName + `(.json|/[a-z0-9\-\.]+)$`
	} else {
		reg += `[a-z0-9\.\-]+(.json|/[a-z0-9\-\.]+)(.json)?$`
	}
	log.Tracef("got part '%s'", reg)
	return reg
}

type LogRegex struct {
	containerName string
	previous      bool
	Regex
	base Regex
}

func NewLogRegex(resourceGroup, resourceName, namespace, containerName string, previous bool) IRegex {
	return &LogRegex{
		Regex: Regex{
			resourceGroup: resourceGroup,
			resourceName:  resourceName,
			namespace:     namespace,
		},
		containerName: containerName,
		previous:      previous,
	}
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
	log.Tracef("got part '%s'", reg)
	return reg + `.log`
}
