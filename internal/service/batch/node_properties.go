// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch

import (
	_ "unsafe" // Required for go:linkname

	_ "github.com/aws/aws-sdk-go-v2/service/batch" // Required for go:linkname
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

type nodeProperties struct {
	MainNode            *int64
	NodeRangeProperties []*nodeRangeProperty

	NumNodes *int64
}

type nodeRangeProperty struct {
	Container     *containerProperties
	EcsProperties *ecsProperties
	EKSProperties *eksProperties
	TargetNodes   *string
}

func (np *nodeProperties) reduce() {
	// Deal with Environment objects which may be re-ordered in the API.
	for _, node := range np.NodeRangeProperties {
		if node.Container != nil {
			node.Container.reduce()
		}
		if node.EcsProperties != nil {
			node.EcsProperties.reduce()
		}
		if node.EKSProperties != nil {
			node.EKSProperties.reduce()
		}
	}
}

func equivalentNodePropertiesJSON(str1, str2 string) (bool, error) {
	if str1 == "" {
		str1 = "{}"
	}

	if str2 == "" {
		str2 = "{}"
	}

	var np1 nodeProperties
	err := tfjson.DecodeFromString(str1, &np1)
	if err != nil {
		return false, err
	}
	np1.reduce()
	b1, err := tfjson.EncodeToBytes(np1)
	if err != nil {
		return false, err
	}

	var np2 nodeProperties
	err = tfjson.DecodeFromString(str2, &np2)
	if err != nil {
		return false, err
	}
	np2.reduce()
	b2, err := tfjson.EncodeToBytes(np2)
	if err != nil {
		return false, err
	}

	return tfjson.EqualBytes(b1, b2), nil
}

func expandJobNodeProperties(tfString string) (*awstypes.NodeProperties, error) {
	apiObject := &awstypes.NodeProperties{}

	if err := tfjson.DecodeFromString(tfString, apiObject); err != nil {
		return nil, err
	}

	return apiObject, nil
}

// Dirty hack to avoid any backwards compatibility issues with the AWS SDK for Go v2 migration.
// Reach down into the SDK and use the same serialization function that the SDK uses.
//
//go:linkname serializeNodeProperties github.com/aws/aws-sdk-go-v2/service/batch.awsRestjson1_serializeDocumentNodeProperties
func serializeNodeProperties(v *awstypes.NodeProperties, value smithyjson.Value) error

func flattenNodeProperties(apiObject *awstypes.NodeProperties) (string, error) {
	if apiObject == nil {
		return "", nil
	}

	jsonEncoder := smithyjson.NewEncoder()
	err := serializeNodeProperties(apiObject, jsonEncoder.Value)

	if err != nil {
		return "", err
	}

	return jsonEncoder.String(), nil
}
