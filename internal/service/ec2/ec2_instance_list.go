// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/go-cty/cty"
	frameworkdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/list"
	listschema "github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKListResource("aws_instance")
func newInstanceResourceAsListResource() inttypes.ListResourceForSDK {
	l := instanceListResource{}
	l.SetResourceSchema(resourceInstance())

	return &l
}

var _ list.ListResourceWithRawV5Schemas = &instanceListResource{}

type instanceListResource struct {
	framework.ListResourceWithSDKv2Resource
}

type instanceListResourceModel struct {
	framework.WithRegionModel
	Filters           customListFilters `tfsdk:"filter"`
	IncludeAutoScaled types.Bool        `tfsdk:"include_auto_scaled"`
}

func (l *instanceListResource) ListResourceConfigSchema(ctx context.Context, request list.ListResourceSchemaRequest, response *list.ListResourceSchemaResponse) {
	response.Schema = listschema.Schema{
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

func (l *instanceListResource) List(ctx context.Context, request list.ListRequest, stream *list.ListResultsStream) {
	awsClient := l.Meta()
	conn := awsClient.EC2Client(ctx)

	var query instanceListResourceModel
	if request.Config.Raw.IsKnown() && !request.Config.Raw.IsNull() {
		if diags := request.Config.Get(ctx, &query); diags.HasError() {
			stream.Results = list.ListResultsStreamDiagnostics(diags)
			return
		}
	}

	var input ec2.DescribeInstancesInput
	if diags := fwflex.Expand(ctx, query, &input); diags.HasError() {
		stream.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// If no instance-state filter is set, default to all states except terminated and shutting-down
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

	includeAutoScaled := query.IncludeAutoScaled.ValueBool()

	stream.Results = func(yield func(list.ListResult) bool) {
		result := request.NewListResult(ctx)

		for instance, err := range listInstances(ctx, conn, &input) {
			if err != nil {
				result = fwdiag.NewListResultErrorDiagnostic(err)
				yield(result)
				return
			}

			tags := keyValueTags(ctx, instance.Tags)

			if !includeAutoScaled {
				// Exclude Auto Scaled Instances
				if v, ok := tags["aws:autoscaling:groupName"]; ok && v.ValueString() != "" {
					continue
				}
			}

			rd := l.ResourceData()
			rd.SetId(aws.ToString(instance.InstanceId))
			result.Diagnostics.Append(translateDiags(resourceInstanceFlatten(ctx, awsClient, &instance, rd))...)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if v, ok := tags["Name"]; ok {
				result.DisplayName = fmt.Sprintf("%s (%s)", v.ValueString(), aws.ToString(instance.InstanceId))
			} else {
				result.DisplayName = aws.ToString(instance.InstanceId)
			}

			l.SetResult(ctx, awsClient, request.IncludeResource, &result, rd)
			if result.Diagnostics.HasError() {
				yield(result)
				return
			}

			if !yield(result) {
				return
			}
		}
	}
}

func translateDiags(in diag.Diagnostics) frameworkdiag.Diagnostics {
	out := make(frameworkdiag.Diagnostics, len(in))
	for i, diagIn := range in {
		var diagOut frameworkdiag.Diagnostic
		if diagIn.Severity == diag.Error {
			if len(diagIn.AttributePath) == 0 {
				diagOut = frameworkdiag.NewErrorDiagnostic(diagIn.Summary, diagIn.Detail)
			} else {
				diagOut = frameworkdiag.NewAttributeErrorDiagnostic(translatePath(diagIn.AttributePath), diagIn.Summary, diagIn.Detail)
			}
		} else {
			if len(diagIn.AttributePath) == 0 {
				diagOut = frameworkdiag.NewWarningDiagnostic(diagIn.Summary, diagIn.Detail)
			} else {
				diagOut = frameworkdiag.NewAttributeWarningDiagnostic(translatePath(diagIn.AttributePath), diagIn.Summary, diagIn.Detail)
			}
		}
		out[i] = diagOut
	}
	return out
}

func translatePath(in cty.Path) path.Path {
	var out path.Path

	if len(in) == 0 {
		return out
	}

	step := in[0]
	switch v := step.(type) {
	case cty.GetAttrStep:
		out = path.Root(v.Name)
	}

	for i := 1; i < len(in); i++ {
		step := in[i]
		switch v := step.(type) {
		case cty.GetAttrStep:
			out = out.AtName(v.Name)

		case cty.IndexStep:
			switch v.Key.Type() {
			case cty.Number:
				v, _ := v.Key.AsBigFloat().Int64()
				out = out.AtListIndex(int(v))
			case cty.String:
				out = out.AtMapKey(v.Key.AsString())
			}
		}
	}

	return out
}
