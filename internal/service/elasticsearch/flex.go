package elasticsearch

import (
	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func expandCognitoOptions(c []interface{}) *elasticsearch.CognitoOptions {
	options := &elasticsearch.CognitoOptions{
		Enabled: aws.Bool(false),
	}
	if len(c) < 1 {
		return options
	}

	m := c[0].(map[string]interface{})

	if cognitoEnabled, ok := m["enabled"]; ok {
		options.Enabled = aws.Bool(cognitoEnabled.(bool))

		if cognitoEnabled.(bool) {

			if v, ok := m["user_pool_id"]; ok && v.(string) != "" {
				options.UserPoolId = aws.String(v.(string))
			}
			if v, ok := m["identity_pool_id"]; ok && v.(string) != "" {
				options.IdentityPoolId = aws.String(v.(string))
			}
			if v, ok := m["role_arn"]; ok && v.(string) != "" {
				options.RoleArn = aws.String(v.(string))
			}
		}
	}

	return options
}

func expandDomainEndpointOptions(l []interface{}) *elasticsearch.DomainEndpointOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})
	domainEndpointOptions := &elasticsearch.DomainEndpointOptions{}

	if v, ok := m["enforce_https"].(bool); ok {
		domainEndpointOptions.EnforceHTTPS = aws.Bool(v)
	}

	if v, ok := m["tls_security_policy"].(string); ok {
		domainEndpointOptions.TLSSecurityPolicy = aws.String(v)
	}

	if customEndpointEnabled, ok := m["custom_endpoint_enabled"]; ok {
		domainEndpointOptions.CustomEndpointEnabled = aws.Bool(customEndpointEnabled.(bool))

		if customEndpointEnabled.(bool) {
			if v, ok := m["custom_endpoint"].(string); ok && v != "" {
				domainEndpointOptions.CustomEndpoint = aws.String(v)
			}

			if v, ok := m["custom_endpoint_certificate_arn"].(string); ok && v != "" {
				domainEndpointOptions.CustomEndpointCertificateArn = aws.String(v)
			}
		}
	}

	return domainEndpointOptions
}

func expandEBSOptions(m map[string]interface{}) *elasticsearch.EBSOptions {
	options := elasticsearch.EBSOptions{}

	if ebsEnabled, ok := m["ebs_enabled"]; ok {
		options.EBSEnabled = aws.Bool(ebsEnabled.(bool))

		if ebsEnabled.(bool) {
			if v, ok := m["iops"]; ok && v.(int) > 0 {
				options.Iops = aws.Int64(int64(v.(int)))
			}
			if v, ok := m["volume_size"]; ok && v.(int) > 0 {
				options.VolumeSize = aws.Int64(int64(v.(int)))
			}
			if v, ok := m["volume_type"]; ok && v.(string) != "" {
				options.VolumeType = aws.String(v.(string))
			}
		}
	}

	return &options
}

func expandEncryptAtRestOptions(m map[string]interface{}) *elasticsearch.EncryptionAtRestOptions {
	options := elasticsearch.EncryptionAtRestOptions{}

	if v, ok := m["enabled"]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}
	if v, ok := m["kms_key_id"]; ok && v.(string) != "" {
		options.KmsKeyId = aws.String(v.(string))
	}

	return &options
}

func expandVPCOptions(m map[string]interface{}) *elasticsearch.VPCOptions {
	options := elasticsearch.VPCOptions{}

	if v, ok := m["security_group_ids"]; ok {
		options.SecurityGroupIds = flex.ExpandStringSet(v.(*schema.Set))
	}
	if v, ok := m["subnet_ids"]; ok {
		options.SubnetIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	return &options
}

func flattenCognitoOptions(c *elasticsearch.CognitoOptions) []map[string]interface{} {
	m := map[string]interface{}{}

	m["enabled"] = aws.BoolValue(c.Enabled)

	if aws.BoolValue(c.Enabled) {
		m["identity_pool_id"] = aws.StringValue(c.IdentityPoolId)
		m["user_pool_id"] = aws.StringValue(c.UserPoolId)
		m["role_arn"] = aws.StringValue(c.RoleArn)
	}

	return []map[string]interface{}{m}
}

func flattenDomainEndpointOptions(domainEndpointOptions *elasticsearch.DomainEndpointOptions) []interface{} {
	if domainEndpointOptions == nil {
		return nil
	}

	m := map[string]interface{}{
		"enforce_https":           aws.BoolValue(domainEndpointOptions.EnforceHTTPS),
		"tls_security_policy":     aws.StringValue(domainEndpointOptions.TLSSecurityPolicy),
		"custom_endpoint_enabled": aws.BoolValue(domainEndpointOptions.CustomEndpointEnabled),
	}
	if aws.BoolValue(domainEndpointOptions.CustomEndpointEnabled) {
		if domainEndpointOptions.CustomEndpoint != nil {
			m["custom_endpoint"] = aws.StringValue(domainEndpointOptions.CustomEndpoint)
		}
		if domainEndpointOptions.CustomEndpointCertificateArn != nil {
			m["custom_endpoint_certificate_arn"] = aws.StringValue(domainEndpointOptions.CustomEndpointCertificateArn)
		}
	}

	return []interface{}{m}
}

func flattenEBSOptions(o *elasticsearch.EBSOptions) []map[string]interface{} {
	m := map[string]interface{}{}

	if o.EBSEnabled != nil {
		m["ebs_enabled"] = *o.EBSEnabled
	}

	if aws.BoolValue(o.EBSEnabled) {
		if o.Iops != nil {
			m["iops"] = *o.Iops
		}
		if o.VolumeSize != nil {
			m["volume_size"] = *o.VolumeSize
		}
		if o.VolumeType != nil {
			m["volume_type"] = *o.VolumeType
		}
	}

	return []map[string]interface{}{m}
}

func flattenEncryptAtRestOptions(o *elasticsearch.EncryptionAtRestOptions) []map[string]interface{} {
	if o == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{}

	if o.Enabled != nil {
		m["enabled"] = *o.Enabled
	}
	if o.KmsKeyId != nil {
		m["kms_key_id"] = *o.KmsKeyId
	}

	return []map[string]interface{}{m}
}

func flattenSnapshotOptions(snapshotOptions *elasticsearch.SnapshotOptions) []map[string]interface{} {
	if snapshotOptions == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"automated_snapshot_start_hour": int(aws.Int64Value(snapshotOptions.AutomatedSnapshotStartHour)),
	}

	return []map[string]interface{}{m}
}

func flattenVPCDerivedInfo(o *elasticsearch.VPCDerivedInfo) []map[string]interface{} {
	m := map[string]interface{}{}

	if o.AvailabilityZones != nil {
		m["availability_zones"] = flex.FlattenStringSet(o.AvailabilityZones)
	}
	if o.SecurityGroupIds != nil {
		m["security_group_ids"] = flex.FlattenStringSet(o.SecurityGroupIds)
	}
	if o.SubnetIds != nil {
		m["subnet_ids"] = flex.FlattenStringSet(o.SubnetIds)
	}
	if o.VPCId != nil {
		m["vpc_id"] = *o.VPCId
	}

	return []map[string]interface{}{m}
}
