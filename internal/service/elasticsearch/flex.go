// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticsearch

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticsearchservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func expandCognitoOptions(c []any) *awstypes.CognitoOptions {
	options := &awstypes.CognitoOptions{
		Enabled: aws.Bool(false),
	}
	if len(c) < 1 {
		return options
	}

	m := c[0].(map[string]any)

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

func expandDomainEndpointOptions(l []any) *awstypes.DomainEndpointOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)
	domainEndpointOptions := &awstypes.DomainEndpointOptions{}

	if v, ok := m["enforce_https"].(bool); ok {
		domainEndpointOptions.EnforceHTTPS = aws.Bool(v)
	}

	if v, ok := m["tls_security_policy"].(string); ok {
		domainEndpointOptions.TLSSecurityPolicy = awstypes.TLSSecurityPolicy(v)
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

func expandEBSOptions(m map[string]any) *awstypes.EBSOptions {
	options := awstypes.EBSOptions{}

	if ebsEnabled, ok := m["ebs_enabled"]; ok {
		options.EBSEnabled = aws.Bool(ebsEnabled.(bool))

		if ebsEnabled.(bool) {
			if v, ok := m[names.AttrVolumeSize]; ok && v.(int) > 0 {
				options.VolumeSize = aws.Int32(int32(v.(int)))
			}
			var volumeType string
			if v, ok := m[names.AttrVolumeType]; ok && v.(string) != "" {
				volumeType = v.(string)
				options.VolumeType = awstypes.VolumeType(volumeType)
			}
			if v, ok := m[names.AttrIOPS]; ok && v.(int) > 0 && ebsVolumeTypePermitsIopsInput(volumeType) {
				options.Iops = aws.Int32(int32(v.(int)))
			}
			if v, ok := m[names.AttrThroughput]; ok && v.(int) > 0 && ebsVolumeTypePermitsThroughputInput(volumeType) {
				options.Throughput = aws.Int32(int32(v.(int)))
			}
		}
	}

	return &options
}

func expandEncryptAtRestOptions(m map[string]any) *awstypes.EncryptionAtRestOptions {
	options := awstypes.EncryptionAtRestOptions{}

	if v, ok := m[names.AttrEnabled]; ok {
		options.Enabled = aws.Bool(v.(bool))
	}
	if v, ok := m[names.AttrKMSKeyID]; ok && v.(string) != "" {
		options.KmsKeyId = aws.String(v.(string))
	}

	return &options
}

func expandVPCOptions(m map[string]any) *awstypes.VPCOptions {
	if m == nil {
		return nil
	}

	options := awstypes.VPCOptions{}

	if v, ok := m[names.AttrSecurityGroupIDs].(*schema.Set); ok && v.Len() > 0 {
		options.SecurityGroupIds = flex.ExpandStringValueSet(v)
	}
	if v, ok := m[names.AttrSubnetIDs].(*schema.Set); ok && v.Len() > 0 {
		options.SubnetIds = flex.ExpandStringValueSet(v)
	}

	return &options
}

func flattenCognitoOptions(c *awstypes.CognitoOptions) []map[string]any {
	m := map[string]any{}

	m[names.AttrEnabled] = aws.ToBool(c.Enabled)

	if aws.ToBool(c.Enabled) {
		m["identity_pool_id"] = aws.ToString(c.IdentityPoolId)
		m[names.AttrUserPoolID] = aws.ToString(c.UserPoolId)
		m[names.AttrRoleARN] = aws.ToString(c.RoleArn)
	}

	return []map[string]any{m}
}

func flattenDomainEndpointOptions(domainEndpointOptions *awstypes.DomainEndpointOptions) []any {
	if domainEndpointOptions == nil {
		return nil
	}

	m := map[string]any{
		"enforce_https":           aws.ToBool(domainEndpointOptions.EnforceHTTPS),
		"tls_security_policy":     string(domainEndpointOptions.TLSSecurityPolicy),
		"custom_endpoint_enabled": aws.ToBool(domainEndpointOptions.CustomEndpointEnabled),
	}
	if aws.ToBool(domainEndpointOptions.CustomEndpointEnabled) {
		if domainEndpointOptions.CustomEndpoint != nil {
			m["custom_endpoint"] = aws.ToString(domainEndpointOptions.CustomEndpoint)
		}
		if domainEndpointOptions.CustomEndpointCertificateArn != nil {
			m["custom_endpoint_certificate_arn"] = aws.ToString(domainEndpointOptions.CustomEndpointCertificateArn)
		}
	}

	return []any{m}
}

func flattenEBSOptions(o *awstypes.EBSOptions) []map[string]any {
	m := map[string]any{}

	if o.EBSEnabled != nil {
		m["ebs_enabled"] = aws.ToBool(o.EBSEnabled)
	}

	if aws.ToBool(o.EBSEnabled) {
		if o.Iops != nil {
			m[names.AttrIOPS] = aws.ToInt32(o.Iops)
		}
		if o.Throughput != nil {
			m[names.AttrThroughput] = aws.ToInt32(o.Throughput)
		}
		if o.VolumeSize != nil {
			m[names.AttrVolumeSize] = aws.ToInt32(o.VolumeSize)
		}
		m[names.AttrVolumeType] = string(o.VolumeType)
	}

	return []map[string]any{m}
}

func flattenEncryptAtRestOptions(o *awstypes.EncryptionAtRestOptions) []map[string]any {
	if o == nil {
		return []map[string]any{}
	}

	m := map[string]any{}

	if o.Enabled != nil {
		m[names.AttrEnabled] = aws.ToBool(o.Enabled)
	}
	if o.KmsKeyId != nil {
		m[names.AttrKMSKeyID] = aws.ToString(o.KmsKeyId)
	}

	return []map[string]any{m}
}

func flattenSnapshotOptions(snapshotOptions *awstypes.SnapshotOptions) []map[string]any {
	if snapshotOptions == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"automated_snapshot_start_hour": int(aws.ToInt32(snapshotOptions.AutomatedSnapshotStartHour)),
	}

	return []map[string]any{m}
}

func flattenVPCDerivedInfo(o *awstypes.VPCDerivedInfo) map[string]any {
	if o == nil {
		return nil
	}

	m := map[string]any{}

	if o.AvailabilityZones != nil {
		m[names.AttrAvailabilityZones] = flex.FlattenStringValueSet(o.AvailabilityZones)
	}
	if o.SecurityGroupIds != nil {
		m[names.AttrSecurityGroupIDs] = flex.FlattenStringValueSet(o.SecurityGroupIds)
	}
	if o.SubnetIds != nil {
		m[names.AttrSubnetIDs] = flex.FlattenStringValueSet(o.SubnetIds)
	}
	if o.VPCId != nil {
		m[names.AttrVPCID] = aws.ToString(o.VPCId)
	}

	return m
}
