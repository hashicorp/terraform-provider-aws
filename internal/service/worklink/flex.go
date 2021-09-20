package worklink

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func flattenIdentityProviderConfigResponse(c *worklink.DescribeIdentityProviderConfigurationOutput) []map[string]interface{} {
	config := make(map[string]interface{})

	if c.IdentityProviderType == nil && c.IdentityProviderSamlMetadata == nil {
		return nil
	}

	if c.IdentityProviderType != nil {
		config["type"] = aws.StringValue(c.IdentityProviderType)
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

	config["subnet_ids"] = flex.FlattenStringSet(c.SubnetIds)
	config["security_group_ids"] = flex.FlattenStringSet(c.SecurityGroupIds)
	config["vpc_id"] = aws.StringValue(c.VpcId)

	return []map[string]interface{}{config}
}
