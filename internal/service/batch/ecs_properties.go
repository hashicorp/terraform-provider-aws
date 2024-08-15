package batch

import (
	_ "unsafe" // Required for go:linkname

	_ "github.com/aws/aws-sdk-go-v2/service/batch" // Required for go:linkname
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

type ecsProperties awstypes.EcsProperties

func (ep *ecsProperties) reduce() {
}

func equivalentECSPropertiesJSON(str1, str2 string) (bool, error) {
	if str1 == "" {
		str1 = "{}"
	}

	if str2 == "" {
		str2 = "{}"
	}

	var ep1 ecsProperties
	err := tfjson.DecodeFromString(str1, &ep1)
	if err != nil {
		return false, err
	}
	ep1.reduce()
	b1, err := tfjson.EncodeToBytes(ep1)
	if err != nil {
		return false, err
	}

	var ep2 ecsProperties
	err = tfjson.DecodeFromString(str2, &ep2)
	if err != nil {
		return false, err
	}
	ep2.reduce()
	b2, err := tfjson.EncodeToBytes(ep2)
	if err != nil {
		return false, err
	}

	return tfjson.EqualBytes(b1, b2), nil
}

func expandECSProperties(tfString string) (*awstypes.EcsProperties, error) {
	apiObject := &awstypes.EcsProperties{}

	if err := tfjson.DecodeFromString(tfString, apiObject); err != nil {
		return nil, err
	}

	return apiObject, nil
}

// Dirty hack to avoid any backwards compatibility issues with the AWS SDK for Go v2 migration.
// Reach down into the SDK and use the same serialization function that the SDK uses.
//
//go:linkname serializeECSPProperties github.com/aws/aws-sdk-go-v2/service/batch.awsRestjson1_serializeDocumentEcsProperties
func serializeECSPProperties(v *awstypes.EcsProperties, value smithyjson.Value) error

func flattenECSProperties(apiObject *awstypes.EcsProperties) (string, error) {
	if apiObject == nil {
		return "", nil
	}

	jsonEncoder := smithyjson.NewEncoder()
	err := serializeECSPProperties(apiObject, jsonEncoder.Value)

	if err != nil {
		return "", err
	}

	return jsonEncoder.String(), nil
}
