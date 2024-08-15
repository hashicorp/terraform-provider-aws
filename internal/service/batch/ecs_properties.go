package batch

import (
	"sort"
	_ "unsafe" // Required for go:linkname

	"github.com/aws/aws-sdk-go-v2/aws"
	_ "github.com/aws/aws-sdk-go-v2/service/batch" // Required for go:linkname
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

type ecsProperties struct {
	TaskProperties []*ecsTaskProperties
}

func (ep *ecsProperties) reduce() {
	for _, taskProp := range ep.TaskProperties {
		taskProp.reduce()
	}
}

type ecsTaskProperties awstypes.EcsTaskProperties

func (tp *ecsTaskProperties) reduce() {
	tp.orderContainers()

	for i, c := range tp.Containers {
		cp := taskContainerProperties(c)
		pcp := &cp
		pcp.reduce()
		tp.Containers[i] = awstypes.TaskContainerProperties(cp)
	}

	// Set all empty slices to nil.
	if len(tp.Volumes) == 0 {
		tp.Volumes = nil
	}
}

func (tp *ecsTaskProperties) orderContainers() {
	sort.Slice(tp.Containers, func(i, j int) bool {
		return aws.ToString(tp.Containers[i].Name) < aws.ToString(tp.Containers[j].Name)
	})
}

type taskContainerProperties awstypes.TaskContainerProperties

func (cp *taskContainerProperties) reduce() {
	cp.orderEnvironmentVariables()
	cp.orderSecrets()

	// Remove environment variables with empty values.
	cp.Environment = tfslices.Filter(cp.Environment, func(kvp awstypes.KeyValuePair) bool {
		return aws.ToString(kvp.Value) != ""
	})

	// Deal with special fields which have defaults.
	if cp.Essential == nil {
		cp.Essential = aws.Bool(true)
	}

	// Set all empty slices to nil.
	if len(cp.Command) == 0 {
		cp.Command = nil
	}
	if len(cp.DependsOn) == 0 {
		cp.DependsOn = nil
	}
	if len(cp.Environment) == 0 {
		cp.Environment = nil
	}
	if cp.LogConfiguration != nil && len(cp.LogConfiguration.SecretOptions) == 0 {
		cp.LogConfiguration.SecretOptions = nil
	}
	if len(cp.MountPoints) == 0 {
		cp.MountPoints = nil
	}
	if len(cp.Secrets) == 0 {
		cp.Secrets = nil
	}
	if len(cp.Ulimits) == 0 {
		cp.Ulimits = nil
	}
}

func (cp *taskContainerProperties) orderEnvironmentVariables() {
	sort.Slice(cp.Environment, func(i, j int) bool {
		return aws.ToString(cp.Environment[i].Name) < aws.ToString(cp.Environment[j].Name)
	})
}

func (cp *taskContainerProperties) orderSecrets() {
	sort.Slice(cp.Secrets, func(i, j int) bool {
		return aws.ToString(cp.Secrets[i].Name) < aws.ToString(cp.Secrets[j].Name)
	})
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
