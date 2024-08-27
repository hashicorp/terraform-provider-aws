// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func flattenLayers(apiObjects []awstypes.Layer) []interface{} {
	return flex.FlattenStringValueList(tfslices.ApplyToAll(apiObjects, func(v awstypes.Layer) string {
		return aws.ToString(v.Arn)
	}))
}

func flattenVPCConfigResponse(apiObject *awstypes.VpcConfigResponse) []interface{} {
	tfMap := make(map[string]interface{})

	if apiObject == nil {
		return nil
	}

	if len(apiObject.SubnetIds) == 0 && len(apiObject.SecurityGroupIds) == 0 && aws.ToString(apiObject.VpcId) == "" {
		return nil
	}

	tfMap["ipv6_allowed_for_dual_stack"] = aws.ToBool(apiObject.Ipv6AllowedForDualStack)
	tfMap[names.AttrSubnetIDs] = apiObject.SubnetIds
	tfMap[names.AttrSecurityGroupIDs] = apiObject.SecurityGroupIds
	if apiObject.VpcId != nil {
		tfMap[names.AttrVPCID] = aws.ToString(apiObject.VpcId)
	}

	return []interface{}{tfMap}
}
