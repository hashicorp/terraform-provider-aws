// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package batch

import (
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

func serializeNodeProperties(v *awstypes.NodeProperties, value smithyjson.Value) error {
	o := value.Object()
	defer o.Close()

	if v.MainNode != nil {
		o.Key("mainNode").Integer(*v.MainNode)
	}
	if v.NodeRangeProperties != nil {
		a := o.Key("nodeRangeProperties").Array()
		for _, node := range v.NodeRangeProperties {
			if err := serializeNodeRangeProperty(&node, a.Value()); err != nil {
				return err
			}
		}
		a.Close()
	}
	if v.NumNodes != nil {
		o.Key("numNodes").Integer(*v.NumNodes)
	}

	return nil
}

func serializeNodeRangeProperty(v *awstypes.NodeRangeProperty, value smithyjson.Value) error {
	o := value.Object()
	defer o.Close()

	if v.ConsumableResourceProperties != nil {
		serializeConsumableResourceProperties(v.ConsumableResourceProperties, o.Key("consumableResourceProperties"))
	}
	if v.Container != nil {
		if err := serializeContainerProperties(v.Container, o.Key("container")); err != nil {
			return err
		}
	}
	if v.EcsProperties != nil {
		if err := serializeECSPProperties(v.EcsProperties, o.Key("ecsProperties")); err != nil {
			return err
		}
	}
	if v.EksProperties != nil {
		serializeEKSProperties(v.EksProperties, o.Key("eksProperties"))
	}
	if v.InstanceTypes != nil {
		serializeStringList(v.InstanceTypes, o.Key("instanceTypes"))
	}
	if v.TargetNodes != nil {
		o.Key("targetNodes").String(*v.TargetNodes)
	}

	return nil
}

func serializeConsumableResourceProperties(v *awstypes.ConsumableResourceProperties, value smithyjson.Value) {
	o := value.Object()
	defer o.Close()

	if v.ConsumableResourceList != nil {
		a := o.Key("consumableResourceList").Array()
		for _, resource := range v.ConsumableResourceList {
			r := a.Value().Object()
			if resource.ConsumableResource != nil {
				r.Key("consumableResource").String(*resource.ConsumableResource)
			}
			if resource.Quantity != nil {
				r.Key("quantity").Long(*resource.Quantity)
			}
			r.Close()
		}
		a.Close()
	}
}

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
