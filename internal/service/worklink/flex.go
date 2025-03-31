// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package worklink

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/worklink"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenIdentityProviderConfigResponse(c *worklink.DescribeIdentityProviderConfigurationOutput) []map[string]any {
	config := make(map[string]any)

	if c.IdentityProviderSamlMetadata == nil {
		return nil
	}

	config[names.AttrType] = c.IdentityProviderType

	if c.IdentityProviderSamlMetadata != nil {
		config["saml_metadata"] = aws.ToString(c.IdentityProviderSamlMetadata)
	}

	return []map[string]any{config}
}

func flattenNetworkConfigResponse(c *worklink.DescribeCompanyNetworkConfigurationOutput) []map[string]any {
	config := make(map[string]any)

	if c == nil {
		return nil
	}

	if len(c.SubnetIds) == 0 && len(c.SecurityGroupIds) == 0 && aws.ToString(c.VpcId) == "" {
		return nil
	}

	config[names.AttrSubnetIDs] = flex.FlattenStringValueSet(c.SubnetIds)
	config[names.AttrSecurityGroupIDs] = flex.FlattenStringValueSet(c.SecurityGroupIds)
	config[names.AttrVPCID] = aws.ToString(c.VpcId)

	return []map[string]any{config}
}
