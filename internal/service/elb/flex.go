package elb

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elb"
)

// Flattens an access log into something that flatmap.Flatten() can handle
func flattenAccessLog(l *elb.AccessLog) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	if l == nil {
		return nil
	}

	r := make(map[string]interface{})
	if l.S3BucketName != nil {
		r["bucket"] = aws.StringValue(l.S3BucketName)
	}

	if l.S3BucketPrefix != nil {
		r["bucket_prefix"] = aws.StringValue(l.S3BucketPrefix)
	}

	if l.EmitInterval != nil {
		r["interval"] = aws.Int64Value(l.EmitInterval)
	}

	if l.Enabled != nil {
		r["enabled"] = aws.BoolValue(l.Enabled)
	}

	result = append(result, r)

	return result
}

// Flattens an array of Backend Descriptions into a a map of instance_port to policy names.
func flattenBackendPolicies(backends []*elb.BackendServerDescription) map[int64][]string {
	policies := make(map[int64][]string)
	for _, i := range backends {
		for _, p := range i.PolicyNames {
			policies[*i.InstancePort] = append(policies[*i.InstancePort], *p)
		}
		sort.Strings(policies[*i.InstancePort])
	}
	return policies
}

// Flattens a health check into something that flatmap.Flatten()
// can handle
func FlattenHealthCheck(check *elb.HealthCheck) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, 1)

	chk := make(map[string]interface{})
	chk["unhealthy_threshold"] = aws.Int64Value(check.UnhealthyThreshold)
	chk["healthy_threshold"] = aws.Int64Value(check.HealthyThreshold)
	chk["target"] = aws.StringValue(check.Target)
	chk["timeout"] = aws.Int64Value(check.Timeout)
	chk["interval"] = aws.Int64Value(check.Interval)

	result = append(result, chk)

	return result
}

// Flattens an array of Instances into a []string
func flattenInstances(list []*elb.Instance) []string {
	result := make([]string, 0, len(list))
	for _, i := range list {
		result = append(result, *i.InstanceId)
	}
	return result
}

// Expands an array of String Instance IDs into a []Instances
func ExpandInstanceString(list []interface{}) []*elb.Instance {
	result := make([]*elb.Instance, 0, len(list))
	for _, i := range list {
		result = append(result, &elb.Instance{InstanceId: aws.String(i.(string))})
	}
	return result
}

// Takes the result of flatmap.Expand for an array of listeners and
// returns ELB API compatible objects
func ExpandListeners(configured []interface{}) ([]*elb.Listener, error) {
	listeners := make([]*elb.Listener, 0, len(configured))

	// Loop over our configured listeners and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		ip := int64(data["instance_port"].(int))
		lp := int64(data["lb_port"].(int))
		l := &elb.Listener{
			InstancePort:     &ip,
			InstanceProtocol: aws.String(data["instance_protocol"].(string)),
			LoadBalancerPort: &lp,
			Protocol:         aws.String(data["lb_protocol"].(string)),
		}

		if v, ok := data["ssl_certificate_id"]; ok {
			l.SSLCertificateId = aws.String(v.(string))
		}

		var valid bool
		if aws.StringValue(l.SSLCertificateId) != "" {
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
			return nil, fmt.Errorf("ELB Listener: ssl_certificate_id may be set only when protocol is 'https' or 'ssl'")
		}
	}

	return listeners, nil
}

// Flattens an array of Listeners into a []map[string]interface{}
func flattenListeners(list []*elb.ListenerDescription) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(list))
	for _, i := range list {
		l := map[string]interface{}{
			"instance_port":     *i.Listener.InstancePort,
			"instance_protocol": strings.ToLower(*i.Listener.InstanceProtocol),
			"lb_port":           *i.Listener.LoadBalancerPort,
			"lb_protocol":       strings.ToLower(*i.Listener.Protocol),
		}
		// SSLCertificateID is optional, and may be nil
		if i.Listener.SSLCertificateId != nil {
			l["ssl_certificate_id"] = aws.StringValue(i.Listener.SSLCertificateId)
		}
		result = append(result, l)
	}
	return result
}

// Takes the result of flatmap.Expand for an array of policy attributes and
// returns ELB API compatible objects
func ExpandPolicyAttributes(configured []interface{}) []*elb.PolicyAttribute {
	attributes := make([]*elb.PolicyAttribute, 0, len(configured))

	// Loop over our configured attributes and create
	// an array of aws-sdk-go compatible objects
	for _, lRaw := range configured {
		data := lRaw.(map[string]interface{})

		a := &elb.PolicyAttribute{
			AttributeName:  aws.String(data["name"].(string)),
			AttributeValue: aws.String(data["value"].(string)),
		}

		attributes = append(attributes, a)

	}

	return attributes
}

// Flattens an array of PolicyAttributes into a []interface{}
func FlattenPolicyAttributes(list []*elb.PolicyAttributeDescription) []interface{} {
	var attributes []interface{}

	for _, attrdef := range list {
		if attrdef == nil {
			continue
		}

		attribute := map[string]string{
			"name":  aws.StringValue(attrdef.AttributeName),
			"value": aws.StringValue(attrdef.AttributeValue),
		}

		attributes = append(attributes, attribute)
	}

	return attributes
}
