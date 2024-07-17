// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"errors"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenAccessLog(apiObject *awstypes.AccessLog) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := make([]interface{}, 0, 1)
	tfMap := make(map[string]interface{})

	if apiObject.S3BucketName != nil {
		tfMap[names.AttrBucket] = aws.ToString(apiObject.S3BucketName)
	}

	if apiObject.S3BucketPrefix != nil {
		tfMap[names.AttrBucketPrefix] = aws.ToString(apiObject.S3BucketPrefix)
	}

	if apiObject.EmitInterval != nil {
		tfMap[names.AttrInterval] = aws.ToInt32(apiObject.EmitInterval)
	}

	tfMap[names.AttrEnabled] = apiObject.Enabled

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenBackendServerDescriptionPolicies(apiObjects []awstypes.BackendServerDescription) map[int32][]string {
	tfMap := make(map[int32][]string)

	for _, apiObject := range apiObjects {
		k := aws.ToInt32(apiObject.InstancePort)
		tfMap[k] = append(tfMap[k], apiObject.PolicyNames...)
		sort.Strings(tfMap[k])
	}

	return tfMap
}

func flattenHealthCheck(apiObject *awstypes.HealthCheck) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfList := make([]interface{}, 0, 1)
	tfMap := make(map[string]interface{})

	tfMap["unhealthy_threshold"] = aws.ToInt32(apiObject.UnhealthyThreshold)
	tfMap["healthy_threshold"] = aws.ToInt32(apiObject.HealthyThreshold)
	tfMap[names.AttrTarget] = aws.ToString(apiObject.Target)
	tfMap[names.AttrTimeout] = aws.ToInt32(apiObject.Timeout)
	tfMap[names.AttrInterval] = aws.ToInt32(apiObject.Interval)

	tfList = append(tfList, tfMap)

	return tfList
}

func flattenInstances(apiObjects []awstypes.Instance) []string {
	return tfslices.ApplyToAll(apiObjects, func(v awstypes.Instance) string {
		return aws.ToString(v.InstanceId)
	})
}

func expandInstances(tfList []interface{}) []awstypes.Instance {
	return tfslices.ApplyToAll(tfList, func(v interface{}) awstypes.Instance {
		return awstypes.Instance{
			InstanceId: aws.String(v.(string)),
		}
	})
}

func expandListeners(tfList []interface{}) ([]awstypes.Listener, error) {
	apiObjects := make([]awstypes.Listener, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.Listener{
			InstancePort:     aws.Int32(int32(tfMap["instance_port"].(int))),
			InstanceProtocol: aws.String(tfMap["instance_protocol"].(string)),
			LoadBalancerPort: int32(tfMap["lb_port"].(int)),
			Protocol:         aws.String(tfMap["lb_protocol"].(string)),
		}

		if v, ok := tfMap["ssl_certificate_id"]; ok {
			apiObject.SSLCertificateId = aws.String(v.(string))
		}

		var valid bool

		if aws.ToString(apiObject.SSLCertificateId) != "" {
			// validate the protocol is correct
			for _, p := range []string{"https", "ssl"} {
				if (strings.ToLower(aws.ToString(apiObject.InstanceProtocol)) == p) || (strings.ToLower(aws.ToString(apiObject.Protocol)) == p) {
					valid = true
				}
			}
		} else {
			valid = true
		}

		if !valid {
			return nil, errors.New(`"ssl_certificate_id" may be set only when "protocol" is "https" or "ssl"`)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects, nil
}

func flattenListenerDescriptions(apiObjects []awstypes.ListenerDescription) []interface{} {
	tfList := make([]interface{}, 0, len(apiObjects))

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"instance_port":     aws.ToInt32(apiObject.Listener.InstancePort),
			"instance_protocol": strings.ToLower(aws.ToString(apiObject.Listener.InstanceProtocol)),
			"lb_port":           apiObject.Listener.LoadBalancerPort,
			"lb_protocol":       strings.ToLower(*apiObject.Listener.Protocol),
		}

		if apiObject.Listener.SSLCertificateId != nil {
			tfMap["ssl_certificate_id"] = aws.ToString(apiObject.Listener.SSLCertificateId)
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandPolicyAttributes(tfList []interface{}) []awstypes.PolicyAttribute {
	apiObjects := make([]awstypes.PolicyAttribute, 0)

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.PolicyAttribute{
			AttributeName:  aws.String(tfMap[names.AttrName].(string)),
			AttributeValue: aws.String(tfMap[names.AttrValue].(string)),
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenPolicyAttributeDescriptions(apiObjects []awstypes.PolicyAttributeDescription) []interface{} {
	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]string{
			names.AttrName:  aws.ToString(apiObject.AttributeName),
			names.AttrValue: aws.ToString(apiObject.AttributeValue),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}
