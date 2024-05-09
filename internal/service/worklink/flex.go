// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package worklink

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenIdentityProviderConfigResponse(c *worklink.DescribeIdentityProviderConfigurationOutput) []map[string]interface{} {
	config := make(map[string]interface{})

	if c.IdentityProviderType == nil && c.IdentityProviderSamlMetadata == nil {
		return nil
	}

	if c.IdentityProviderType != nil {
		config[names.AttrType] = aws.StringValue(c.IdentityProviderType)
	}
	if c.IdentityProviderSamlMetadata != nil {
		config["saml_metadata"] = aws.StringValue(c.IdentityProviderSamlMetadata)
	}

	return []map[string]interface{}{config}
}

func flattenNetworkConfigResponse(c *worklink.DescribeCompanyNetworkConfigurationOutput) []map[string]interface{} {
	config := make(map[string]interface{})

	if c == nil {
		return nil
	}

	if len(c.SubnetIds) == 0 && len(c.SecurityGroupIds) == 0 && aws.StringValue(c.VpcId) == "" {
		return nil
	}

	config[names.AttrSubnetIDs] = flex.FlattenStringSet(c.SubnetIds)
	config[names.AttrSecurityGroupIDs] = flex.FlattenStringSet(c.SecurityGroupIds)
	config[names.AttrVPCID] = aws.StringValue(c.VpcId)

	return []map[string]interface{}{config}
}
