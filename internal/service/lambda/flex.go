// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda

import (
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func flattenLayers(layers []types.Layer) []interface{} {
	arns := make([]*string, len(layers))
	for i, layer := range layers {
		arns[i] = layer.Arn
	}
	return flex.FlattenStringList(arns)
}

func flattenVPCConfigResponse(s *types.VpcConfigResponse) []map[string]interface{} {
	settings := make(map[string]interface{})

	if s == nil {
		return nil
	}

	var emptyVpc bool
	if aws.StringValue(s.VpcId) == "" {
		emptyVpc = true
	}
	if len(s.SubnetIds) == 0 && len(s.SecurityGroupIds) == 0 && emptyVpc {
		return nil
	}

	settings["ipv6_allowed_for_dual_stack"] = aws.BoolValue(s.Ipv6AllowedForDualStack)
	settings["subnet_ids"] = flex.FlattenStringValueSet(s.SubnetIds)
	settings["security_group_ids"] = flex.FlattenStringValueSet(s.SecurityGroupIds)
	if s.VpcId != nil {
		settings["vpc_id"] = aws.StringValue(s.VpcId)
	}

	return []map[string]interface{}{settings}
}
