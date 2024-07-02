// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"errors"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Flattens an access log into something that flatmap.Flatten() can handle
func flattenAccessLog(l *awstypes.AccessLog) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	if l == nil {
		return nil
	}

	r := make(map[string]interface{})
	if l.S3BucketName != nil {
		r[names.AttrBucket] = aws.ToString(l.S3BucketName)
	}

	if l.S3BucketPrefix != nil {
		r[names.AttrBucketPrefix] = aws.ToString(l.S3BucketPrefix)
	}

	if l.EmitInterval != nil {
		r[names.AttrInterval] = aws.ToInt32(l.EmitInterval)
	}

	r[names.AttrEnabled] = l.Enabled

	result = append(result, r)

	return result
}

// Flattens an array of Backend Descriptions into a a map of instance_port to policy names.
func flattenBackendPolicies(backends []awstypes.BackendServerDescription) map[int32][]string {
	policies := make(map[int32][]string)
	for _, i := range backends {
		for _, p := range i.PolicyNames {
			policies[*i.InstancePort] = append(policies[*i.InstancePort], p)
		}
		sort.Strings(policies[*i.InstancePort])
	}
	return policies
}

// Flattens a health check into something that flatmap.Flatten()
// can handle
func FlattenHealthCheck(check *awstypes.HealthCheck) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	chk := make(map[string]interface{})
	chk["unhealthy_threshold"] = aws.ToInt32(check.UnhealthyThreshold)
	chk["healthy_threshold"] = aws.ToInt32(check.HealthyThreshold)
	chk[names.AttrTarget] = aws.ToString(check.Target)
	chk[names.AttrTimeout] = aws.ToInt32(check.Timeout)
	chk[names.AttrInterval] = aws.ToInt32(check.Interval)

	result = append(result, chk)

	return result
}

// Flattens an array of Instances into a []string
func flattenInstances(list []awstypes.Instance) []string {
	result := make([]string, 0, len(list))
	for _, i := range list {
		result = append(result, *i.InstanceId)
	}
	return result
}

// Expands an array of String Instance IDs into a []Instances
func ExpandInstanceString(list []interface{}) []awstypes.Instance {
	result := make([]awstypes.Instance, 0, len(list))
	for _, i := range list {
		result = append(result, awstypes.Instance{InstanceId: aws.String(i.(string))})
	}
	return result
}

// Takes the result of flatmap.Expand for an array of listeners and
// returns ELB API compatible objects
func ExpandListeners(configured []interface{}) ([]awstypes.Listener, error) {
	listeners := make([]awstypes.Listener, 0, len(configured))

	// Loop over our configured listeners and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		ip := int32(data["instance_port"].(int))
		lp := int32(data["lb_port"].(int))
		l := awstypes.Listener{
			InstancePort:     &ip,
			InstanceProtocol: aws.String(data["instance_protocol"].(string)),
			LoadBalancerPort: lp,
			Protocol:         aws.String(data["lb_protocol"].(string)),
		}

		if v, ok := data["ssl_certificate_id"]; ok {
			l.SSLCertificateId = aws.String(v.(string))
		}

		var valid bool
		if aws.ToString(l.SSLCertificateId) != "" {
			// validate the protocol is correct
			for _, p := range []string{"https", "ssl"} {
				if (strings.ToLower(*l.InstanceProtocol) == p) || (strings.ToLower(*l.Protocol) == p) {
					valid = true
				}
			}
		} else {
			valid = true
		}

		if valid {
			listeners = append(listeners, l)
		} else {
			return nil, errors.New(`"ssl_certificate_id" may be set only when "protocol" is "https" or "ssl"`)
		}
	}

	return listeners, nil
}

// Flattens an array of Listeners into a []map[string]interface{}
func flattenListeners(list []awstypes.ListenerDescription) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		l := map[string]interface{}{
			"instance_port":     *i.Listener.InstancePort,
			"instance_protocol": strings.ToLower(*i.Listener.InstanceProtocol),
			"lb_port":           i.Listener.LoadBalancerPort,
			"lb_protocol":       strings.ToLower(*i.Listener.Protocol),
		}
		// SSLCertificateID is optional, and may be nil
		if i.Listener.SSLCertificateId != nil {
			l["ssl_certificate_id"] = aws.ToString(i.Listener.SSLCertificateId)
		}
		result = append(result, l)
	}
	return result
}

// Takes the result of flatmap.Expand for an array of policy attributes and
// returns ELB API compatible objects
func ExpandPolicyAttributes(configured []interface{}) []awstypes.PolicyAttribute {
	attributes := make([]awstypes.PolicyAttribute, 0, len(configured))

	// Loop over our configured attributes and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		a := awstypes.PolicyAttribute{
			AttributeName:  aws.String(data[names.AttrName].(string)),
			AttributeValue: aws.String(data[names.AttrValue].(string)),
		}

		attributes = append(attributes, a)
	}

	return attributes
}

// Flattens an array of PolicyAttributes into a []interface{}
func FlattenPolicyAttributes(list []awstypes.PolicyAttributeDescription) []interface{} {
	var attributes []interface{}

	for _, attrdef := range list {
		attribute := map[string]string{
			names.AttrName:  aws.ToString(attrdef.AttributeName),
			names.AttrValue: aws.ToString(attrdef.AttributeValue),
		}

		attributes = append(attributes, attribute)
	}

	return attributes
}
