// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"iter"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_instance")
func newInstanceResourceAsListResource() inttypes.ListResourceForSDK {
	return framework.NewSDKListResource(resourceInstance(), framework.SDKListSpec[awstypes.Instance, instanceListResourceModel]{
		ConfigSchema: listInstanceConfigSchema,
		Items: func(ctx context.Context, c *conns.AWSClient, query instanceListResourceModel) (iter.Seq2[awstypes.Instance, error], diag.Diagnostics) {
			var input ec2.DescribeInstancesInput
			if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
				return nil, diags
			}

			// If no instance-state filter is set, default to all states except
			// terminated and shutting-down.
			if !slices.ContainsFunc(input.Filters, func(i awstypes.Filter) bool {
				return aws.ToString(i.Name) == "instance-state-name" || aws.ToString(i.Name) == "instance-state-code"
			}) {
				states := enum.Slice(slices.DeleteFunc(enum.EnumValues[awstypes.InstanceStateName](), func(s awstypes.InstanceStateName) bool {
					return s == awstypes.InstanceStateNameTerminated || s == awstypes.InstanceStateNameShuttingDown
				})...)
				input.Filters = append(input.Filters, awstypes.Filter{
					Name:   aws.String("instance-state-name"),
					Values: states,
				})
			}

			instances := listInstances(ctx, c.EC2Client(ctx), &input)
			if query.IncludeAutoScaled.ValueBool() {
				return instances, nil
			}

			// Auto Scaling membership shows up as a tag, which DescribeInstances
			// cannot filter on.
			return tfiter.FilterValues(instances, func(i awstypes.Instance) bool {
				v, ok := keyValueTags(ctx, i.Tags)["aws:autoscaling:groupName"]

				return !ok || v.ValueString() == ""
			}), nil
		},
		ID: func(i awstypes.Instance, _ instanceListResourceModel) string {
			return aws.ToString(i.InstanceId)
		},
		Prepare: func(ctx context.Context, i awstypes.Instance, _ instanceListResourceModel, _ *schema.ResourceData) {
			setTagsOut(ctx, i.Tags)
		},
		Flatten: resourceInstanceFlatten,
		DisplayName: func(i awstypes.Instance) string {
			return listDisplayName(i.Tags, aws.ToString(i.InstanceId))
		},
	})
}

type instanceListResourceModel struct {
	framework.WithRegionModel
	Filters           customListFilters `tfsdk:"filter"`
	IncludeAutoScaled types.Bool        `tfsdk:"include_auto_scaled"`
}

func listInstanceConfigSchema(ctx context.Context) listschema.Schema {
	return listschema.Schema{
		Attributes: map[string]listschema.Attribute{
			"include_auto_scaled": listschema.BoolAttribute{
				Description: "Whether to include instances that are part of an Auto Scaling group. Auto scaled instances are excluded by default.",
				Optional:    true,
			},
		},
		Blocks: map[string]listschema.Block{
			names.AttrFilter: customListFiltersBlock(ctx),
		},
	}
}
