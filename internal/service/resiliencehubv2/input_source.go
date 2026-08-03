// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package resiliencehubv2

import (
	"context"
	"errors"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resiliencehubv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const inputSourceImportIDPartCount = 2

// @FrameworkResource("aws_resiliencehubv2_input_source", name="Input Source")
// @IdentityAttribute("service_arn")
// @IdentityAttribute("input_source_id")
// @ImportIDHandler("inputSourceImportID", setIDAttribute=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/resiliencehubv2/types;awstypes;awstypes.InputSourceSummary")
// @Testing(hasNoPreExistingResource=true)
func newResourceInputSource(context.Context) (resource.ResourceWithConfigure, error) {
	return &resourceInputSource{}, nil
}

type resourceInputSource struct {
	framework.ResourceWithModel[resourceInputSourceModel]
	framework.WithImportByIdentity
}

func (r *resourceInputSource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrID: fwschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"service_arn": fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"input_source_id": fwschema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cfn_stack_arn": fwschema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(
						path.MatchRoot("cfn_stack_arn"),
						path.MatchRoot("tf_state_file_url"),
						path.MatchRoot("eks_cluster_arn"),
					),
				},
			},
			"tf_state_file_url": fwschema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"eks_cluster_arn": fwschema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"eks_namespaces": fwschema.ListAttribute{
				Optional:   true,
				CustomType: fwtypes.ListOfStringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourceInputSource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourceInputSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.CreateInputSourceInput{
		ServiceArn: plan.ServiceArn.ValueStringPointer(),
	}

	switch {
	case !plan.CfnStackArn.IsNull():
		input.ResourceConfiguration = &awstypes.ResourceConfigurationMemberCfnStackArn{
			Value: plan.CfnStackArn.ValueString(),
		}
	case !plan.TfStateFileUrl.IsNull():
		input.ResourceConfiguration = &awstypes.ResourceConfigurationMemberTfStateFileUrl{
			Value: plan.TfStateFileUrl.ValueString(),
		}
	case !plan.EksClusterArn.IsNull():
		var namespaces []string
		if !plan.EksNamespaces.IsNull() {
			smerr.AddEnrich(ctx, &resp.Diagnostics, plan.EksNamespaces.ElementsAs(ctx, &namespaces, false))
			if resp.Diagnostics.HasError() {
				return
			}
		}
		input.ResourceConfiguration = &awstypes.ResourceConfigurationMemberEks{
			Value: awstypes.EksSource{
				ClusterArn: plan.EksClusterArn.ValueStringPointer(),
				Namespaces: namespaces,
			},
		}
	}

	output, err := conn.CreateInputSource(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	plan.InputSourceId = types.StringPointerValue(output.InputSourceId)
	plan.ID = types.StringValue(plan.ServiceArn.ValueString() + "," + plan.InputSourceId.ValueString())

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *resourceInputSource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourceInputSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	is, err := findInputSourceByID(ctx, conn, state.ServiceArn.ValueString(), state.InputSourceId.ValueString())
	if retry.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
		return
	}

	r.flatten(is, &state)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, state))
}

func (r *resourceInputSource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// No Update API — all attributes are ForceNew
}

func (r *resourceInputSource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourceInputSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ResilienceHubV2Client(ctx)

	input := resiliencehubv2.DeleteInputSourceInput{
		ServiceArn:    state.ServiceArn.ValueStringPointer(),
		InputSourceId: state.InputSourceId.ValueStringPointer(),
	}
	_, err := conn.DeleteInputSource(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return
		}
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.ID.ValueString())
	}
}

type inputSourceImportID struct{}

func (inputSourceImportID) Parse(id string) (string, map[string]any, error) {
	parts, err := flex.ExpandResourceId(id, inputSourceImportIDPartCount, false)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"service_arn":     parts[0],
		"input_source_id": parts[1],
	}

	return id, result, nil
}

func (inputSourceImportID) Create(ctx context.Context, state tfsdk.State) string {
	var serviceArn, inputSourceID types.String
	state.GetAttribute(ctx, path.Root("service_arn"), &serviceArn)
	state.GetAttribute(ctx, path.Root("input_source_id"), &inputSourceID)

	return serviceArn.ValueString() + "," + inputSourceID.ValueString()
}

func (r *resourceInputSource) flatten(is *awstypes.InputSourceSummary, data *resourceInputSourceModel) {
	data.InputSourceId = types.StringPointerValue(is.InputSourceId)
	if is.CfnStackArn != nil {
		data.CfnStackArn = types.StringPointerValue(is.CfnStackArn)
	}
	if is.TfStateFileUrl != nil {
		data.TfStateFileUrl = types.StringPointerValue(is.TfStateFileUrl)
	}
	if is.Eks != nil {
		data.EksClusterArn = types.StringPointerValue(is.Eks.ClusterArn)
	}
}

func findInputSourceByID(ctx context.Context, conn *resiliencehubv2.Client, serviceArn, inputSourceId string) (*awstypes.InputSourceSummary, error) {
	input := resiliencehubv2.ListInputSourcesInput{
		ServiceArn: aws.String(serviceArn),
	}
	output, err := conn.ListInputSources(ctx, &input)
	if err != nil {
		var nfe *awstypes.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, smarterr.NewError(&retry.NotFoundError{LastError: err})
		}
		return nil, smarterr.NewError(err)
	}

	for _, is := range output.InputSourceSummaries {
		if aws.ToString(is.InputSourceId) == inputSourceId {
			return &is, nil
		}
	}

	return nil, smarterr.NewError(tfresource.NewEmptyResultError())
}

type resourceInputSourceModel struct {
	framework.WithRegionModel
	CfnStackArn    types.String         `tfsdk:"cfn_stack_arn"`
	EksClusterArn  types.String         `tfsdk:"eks_cluster_arn"`
	EksNamespaces  fwtypes.ListOfString `tfsdk:"eks_namespaces"`
	ID             types.String         `tfsdk:"id"`
	InputSourceId  types.String         `tfsdk:"input_source_id"`
	ServiceArn     types.String         `tfsdk:"service_arn"`
	TfStateFileUrl types.String         `tfsdk:"tf_state_file_url"`
}
