// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package serviceindex

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/awsmapping"
	"github.com/hashicorp/terraform-provider-aws/tools/drift-detect/internal/tfschema"
)

type ResourceServiceInfo struct {
	Resource          string                   `json:"resource"`
	TFFile            string                   `json:"tf_file"`
	TFResourceSegment string                   `json:"tf_resource_segment"`
	AWSClient         string                   `json:"aws_client"`
	AWSService        string                   `json:"aws_service"`
	AWSFile           string                   `json:"aws_file"`
	AWSNamespace      string                   `json:"aws_namespace"`
	Skipped           *awsmapping.SkipResource `json:"skipped,omitempty"`
	Error             string                   `json:"error,omitempty"`
}

const (
	cacheFileName = "resource-service-index.json"
	cacheTTL      = 24 * time.Hour
)

func LoadResourceServiceIndex(repoRoot string, refresh bool, tfSchema *tfschema.ProviderSchema, m *awsmapping.File) (map[string]*ResourceServiceInfo, error) {
	cachePath := filepath.Join(repoRoot, ".cache", cacheFileName)

	if refresh {
		_ = os.Remove(cachePath)
	}

	if isResourceServiceCacheFresh(cachePath) {
		return readResourceServiceCache(cachePath)
	}

	index, err := BuildResourceServiceIndex(repoRoot, tfSchema, m.SkipResources)
	if err != nil {
		return nil, err
	}

	err = writeResourceServiceCache(cachePath, index)
	if err != nil {
		fmt.Printf("warning: unable to write resource service cache: %v\n", err)
	}

	return index, nil
}

func isResourceServiceCacheFresh(cachePath string) bool {
	info, err := os.Stat(cachePath)
	if err != nil {
		return false
	}

	return time.Since(info.ModTime()) < cacheTTL
}

func readResourceServiceCache(cachePath string) (map[string]*ResourceServiceInfo, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	var result map[string]*ResourceServiceInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func ReadResourceServiceCache(cachePath string) (map[string]*ResourceServiceInfo, error) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}
	var result map[string]*ResourceServiceInfo
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func writeResourceServiceCache(cachePath string, index map[string]*ResourceServiceInfo) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0o644)
}

func BuildResourceServiceIndex(repoRoot string, tfSchema *tfschema.ProviderSchema, skipResources map[string]*awsmapping.SkipResource) (map[string]*ResourceServiceInfo, error) {
	serviceIndexes, err := BuildServiceIndexes(repoRoot)
	if err != nil {
		return nil, err
	}
	resourceNameToFileLoc := serviceIndexes.ResourceIndex
	fileToClientList := serviceIndexes.FileConnIndex

	awsClientToService, err := BuildAWSClientIndex(repoRoot)
	if err != nil {
		return nil, err
	}

	awsServiceIndex, err := BuildAWSServiceIndex()
	if err != nil {
		return nil, err
	}

	resourceServiceIndex := make(map[string]*ResourceServiceInfo)

	for _, tfResourceName := range tfSchema.ResourceNames() {
		info := &ResourceServiceInfo{
			Resource: tfResourceName,
		}
		resourceServiceIndex[tfResourceName] = info

		if skip, ok := skipResources[tfResourceName]; ok {
			info.Skipped = skip
			continue
		}

		// Check resourceNameToFileLoc
		fileLocations, ok := resourceNameToFileLoc[tfResourceName]
		if !ok || len(fileLocations) == 0 {
			info.Error = fmt.Sprintf("Could not find TF file defining resource '%s'", tfResourceName)
			continue
		}
		if len(fileLocations) > 1 {
			info.Error = fmt.Sprintf("Multiple TF files defining resource '%s'", tfResourceName)
			continue
		}
		resourceFileLoc := fileLocations[0]
		info.TFFile = resourceFileLoc.File
		info.TFResourceSegment = resourceFileLoc.TFResourceSegment
		resourceService := resourceFileLoc.Service

		// Get and check clients from fileToListOfClients
		clientList, ok := fileToClientList[resourceFileLoc.File]
		if !ok || len(clientList) == 0 {
			info.Error = fmt.Sprintf("Could not find AWS client used for resource '%s'", tfResourceName)
			continue
		}
		var awsClient string
		if len(clientList) == 1 {
			awsClient = clientList[0]
		} else {
			// If multiple clients, try to find the one that matches the service name
			expectedClient := strings.ReplaceAll(resourceService, "_", "") + "Client"
			for _, client := range clientList {
				if strings.EqualFold(client, expectedClient) {
					awsClient = client
					break
				}
			}
			if awsClient == "" {
				info.Error = fmt.Sprintf("Resource %q uses several AWS clients (%v) and none matched expected client %q", tfResourceName, clientList, expectedClient)
				continue
			}
		}
		info.AWSClient = awsClient

		// Find AWS service name using awsClientToService map
		awsServiceName, ok := awsClientToService[awsClient]
		if !ok || awsServiceName == "" {
			info.Error = fmt.Sprintf("AWS Client '%s' does not have a mapped AWS Service Name", awsClient)
			continue
		}

		// Check if extracted awsServiceName is listed in the AWS Service Index
		if awsServiceIndex.HasService(awsServiceName) {
			info.AWSService = awsServiceName
		} else if alias, ok := awsServiceIndex.aliases[awsServiceName]; ok && awsServiceIndex.HasService(alias) {
			info.AWSService = alias
		} else {
			info.Error = fmt.Sprintf("Could not find file for AWS service '%s'", awsServiceName)
			continue
		}

		// Get the file name and namespace for the resource
		awsFile, awsNamespace, err := awsServiceIndex.ResolveModelPath(info.AWSService)
		if err != nil {
			info.Error = err.Error()
			continue
		}
		info.AWSFile = awsFile
		info.AWSNamespace = awsNamespace
	}

	return resourceServiceIndex, nil
}
