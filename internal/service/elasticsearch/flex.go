// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"github.com/aws/aws-sdk-go/aws"
	elasticsearch "github.com/aws/aws-sdk-go/service/elasticsearchservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandCognitoOptions(c []interface{}) *elasticsearch.CognitoOptions {
	options := &elasticsearch.CognitoOptions{
		Enabled: aws.Bool(false),
	}
	if len(c) < 1 {
		return options
	}

	m := c[0].(map[string]interface{})

	if cognitoEnabled, ok := m[names.AttrEnabled]; ok {
		options.Enabled = aws.Bool(cognitoEnabled.(bool))

		if cognitoEnabled.(bool) {
			if v, ok := m[names.AttrUserPoolID]; ok && v.(string) != "" {
				options.UserPoolId = aws.String(v.(string))
			}
			if v, ok := m["identity_pool_id"]; ok && v.(string) != "" {
				options.IdentityPoolId = aws.String(v.(string))
			}
			if v, ok := m[names.AttrRoleARN]; ok && v.(string) != "" {
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
			if v, ok := m[names.AttrVolumeSize]; ok && v.(int) > 0 {
				options.VolumeSize = aws.Int64(int64(v.(int)))
			}
			var volumeType string
			if v, ok := m[names.AttrVolumeType]; ok && v.(string) != "" {
				volumeType = v.(string)
				options.VolumeType = aws.String(volumeType)
			}
			if v, ok := m[names.AttrIOPS]; ok && v.(int) > 0 && EBSVolumeTypePermitsIopsInput(volumeType) {
				options.Iops = aws.Int64(int64(v.(int)))
			}
			if v, ok := m[names.AttrThroughput]; ok && v.(int) > 0 && EBSVolumeTypePermitsThroughputInput(volumeType) {
				options.Throughput = aws.Int64(int64(v.(int)))
			}
		}
	}

	return &options
}

func expandEncryptAtRestOptions(m map[string]interface{}) *elasticsearch.EncryptionAtRestOptions {
	options := elasticsearch.EncryptionAtRestOptions{}

	if v, ok := m[names.AttrEnabled]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}
	if v, ok := m[names.AttrKMSKeyID]; ok && v.(string) != "" {
		options.KmsKeyId = aws.String(v.(string))
	}

	return &options
}

func expandVPCOptions(m map[string]interface{}) *elasticsearch.VPCOptions {
	if m == nil {
		return nil
	}

	options := elasticsearch.VPCOptions{}

	if v, ok := m[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		options.SecurityGroupIds = flex.ExpandStringSet(v)
	}
	if v, ok := m[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		options.SubnetIds = flex.ExpandStringSet(v)
	}

	return &options
}

func flattenCognitoOptions(c *elasticsearch.CognitoOptions) []map[string]interface{} {
	m := map[string]interface{}{}

	m[names.AttrEnabled] = aws.BoolValue(c.Enabled)

	if aws.BoolValue(c.Enabled) {
		m["identity_pool_id"] = aws.StringValue(c.IdentityPoolId)
		m[names.AttrUserPoolID] = aws.StringValue(c.UserPoolId)
		m[names.AttrRoleARN] = aws.StringValue(c.RoleArn)
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
		m["ebs_enabled"] = aws.BoolValue(o.EBSEnabled)
	}

	if aws.BoolValue(o.EBSEnabled) {
		if o.Iops != nil {
			m[names.AttrIOPS] = aws.Int64Value(o.Iops)
		}
		if o.Throughput != nil {
			m[names.AttrThroughput] = aws.Int64Value(o.Throughput)
		}
		if o.VolumeSize != nil {
			m[names.AttrVolumeSize] = aws.Int64Value(o.VolumeSize)
		}
		if o.VolumeType != nil {
			m[names.AttrVolumeType] = aws.StringValue(o.VolumeType)
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
		m[names.AttrEnabled] = aws.BoolValue(o.Enabled)
	}
	if o.KmsKeyId != nil {
		m[names.AttrKMSKeyID] = aws.StringValue(o.KmsKeyId)
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

func flattenVPCDerivedInfo(o *elasticsearch.VPCDerivedInfo) map[string]interface{} {
	if o == nil {
		return nil
	}

	m := map[string]interface{}{}

	if o.AvailabilityZones != nil {
		m[names.AttrAvailabilityZones] = flex.FlattenStringSet(o.AvailabilityZones)
	}
	if o.SecurityGroupIds != nil {
		m[names.AttrSecurityGroupIDs] = flex.FlattenStringSet(o.SecurityGroupIds)
	}
	if o.SubnetIds != nil {
		m[names.AttrSubnetIDs] = flex.FlattenStringSet(o.SubnetIds)
	}
	if o.VPCId != nil {
		m[names.AttrVPCID] = aws.StringValue(o.VPCId)
	}

	return m
}
