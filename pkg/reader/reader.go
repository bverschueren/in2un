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
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/bverschueren/in2un/pkg/deserializer"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var ErrInvalidInsightsArchive = fmt.Errorf("no valid insights (gzip compressed) archive provided")

const AllNamespaceValue = "_all_"

type InsightsReader struct {
	Path   string
	Reader *tar.Reader
}

func NewInsightsReader(path string) (*InsightsReader, error) {
	r, err := open(path)
	if err != nil {
		return nil, err
	} else {
		return &InsightsReader{Reader: r, Path: path}, nil
	}
}

func (ir *InsightsReader) ReadResource(resourceGroup, resourceName, namespace string) []*unstructured.Unstructured {
	return readResources(ir.Reader, resourceGroup, resourceName, namespace)
}

func (ir *InsightsReader) ReadLog(resourceGroup, resourceName, namespace, containerName string, previous bool) io.Reader {
	return readLogs(ir.Reader, resourceGroup, resourceName, namespace, containerName, previous)
}

// read plain or gzipped tar and return tar.Reader
func open(filename string) (*tar.Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to open insights archive: %w", err)
	}
	var reader io.Reader
	// insights archives are gzipped so try opening as such
	reader, err = gzip.NewReader(file)
	if err != nil {
		return nil, ErrInvalidInsightsArchive
	}
	tr := tar.NewReader(reader)
	return tr, nil
}

// read a resource or log from an archive and return either an unstructed for resource or bytes.Buffer for logs
func readResources(tr *tar.Reader, resourceGroup, resourceName, namespace string) []*unstructured.Unstructured {
	configRegex := &ResourceRegex{base: &ConfigRegex{BaseRegex: BaseRegex{resourceGroup: resourceGroup, namespace: namespace}}, resourceName: resourceName}
	conditionalRegex := &ResourceRegex{base: &ConditionalRegex{BaseRegex: BaseRegex{resourceGroup: resourceGroup, namespace: namespace}}, resourceName: resourceName}
	regs := []string{configRegex.getPart(), conditionalRegex.getPart()}

	log.Debugf("Searching tar file for regex '%s'\n", regs)
	var result []*unstructured.Unstructured
	configMaps := deserializer.NewConfigMapData()
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // end of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		// check if the resourceGroup is found on a deviant but well-known path
		if hdr.Name == wellKnownInsightsJson(resourceGroup) {
			log.Debugf("Found well-known path at '%s'\n", hdr.Name)
			var raw bytes.Buffer
			raw.ReadFrom(tr)
			object, err := deserializer.JsonToUnstructed(raw.Bytes())
			if err != nil {
				log.Fatal(err)
			}
			result = append(result, object)
			return result
		}
		resourceFile := resourceFilename(regs, hdr.Name)
		if resourceFile != "" {
			var raw bytes.Buffer
			raw.ReadFrom(tr)
			object, err := deserializer.JsonToUnstructed(raw.Bytes())
			if err != nil {
				// perhaps it's a configmap
				namespace, name, key, err := configMapFromFilename(hdr.Name)
				if err == nil {
					configMaps.Upsert(namespace, name, key, raw.String())
				} else {
					log.Debug(err)
				}
			} else {
				result = append(result, object)
			}
		}
	}
	result = append(result, configMaps.Flatten()...)
	return result
}

func readLogs(tr *tar.Reader, resourceGroup, resourceName, namespace, containerName string, previous bool) io.ReadCloser {
	regex := &LogRegex{base: &ConfigRegex{BaseRegex: BaseRegex{resourceGroup: resourceGroup, namespace: namespace}}, resourceName: resourceName, containerName: containerName, previous: previous}
	//regex := &LogRegex{base: &ConfigRegex{resourceGroup: resourceGroup, namespace: namespace}, resourceName: resourceName, containerName: containerName, previous: previous}
	regs := []string{regex.getPart()}
	log.Debugf("Searching tar file for regex '%s'\n", regs)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // end of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		resourceFile := resourceFilename(regs, hdr.Name)
		if resourceFile != "" {
			if containerName == "" {
				containerName, _ := containerAndVersionFromFilename(resourceFile)
				log.Printf("Defaulted container \"%s\"\n", containerName)
				// TODO: continue looping tar headers and append additional containers to the previous output
			}
			break
		}
	}
	return io.NopCloser(tr)
}

func wellKnownInsightsJson(resourceGroup string) string {
	return `config/` + resourceGroup + `.json`
}

// check if the current file header name matches any of the regexes
func resourceFilename(regs []string, in string) string {
	log.Tracef("scanning '%s'\n", in)
	for _, r := range regs {
		log.Tracef("with '%s'\n", r)
		re := regexp.MustCompile(r)
		match := re.FindString(in)
		if match != "" {
			log.Tracef("found match for '%s' on %s\n", r, in)
			return match
		}
	}
	return ""
}

func containerAndVersionFromFilename(filename string) (string, string) {
	base := path.Base(strings.TrimSuffix(filename, ".log"))
	parts := strings.Split(base, "_")
	if len(parts) > 1 {
		return parts[0], parts[1]
	} else {
		return parts[0], ""
	}
}

func configMapFromFilename(tarFilePath string) (namespace, name, key string, err error) {
	parts := strings.Split(strings.TrimSuffix(tarFilePath, "/"), "/")
	if len(parts) != 5 {
		return "", "", "", deserializer.ErrUnknownResourcePath
	}
	return parts[2], parts[3], parts[4], nil
}
