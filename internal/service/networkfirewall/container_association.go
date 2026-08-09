// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package networkfirewall

import (
	"context"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/networkfirewall"
	awstypes "github.com/aws/aws-sdk-go-v2/service/networkfirewall/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_networkfirewall_container_association", name="Container Association")
// @Tags(identifierAttribute="container_association_arn")
// @ArnIdentity("container_association_arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/networkfirewall;networkfirewall.DescribeContainerAssociationOutput")
// @Testing(preCheck="testAccPreCheck")
// @Testing(importIgnore="update_token", plannableImportAction="NoOp")
// @Testing(hasNoPreExistingResource=true)
func newContainerAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &containerAssociationResource{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type containerAssociationResource struct {
	framework.ResourceWithModel[containerAssociationResourceModel]
	framework.WithTimeouts
	framework.WithImportByIdentity
}

func (r *containerAssociationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"container_association_arn": framework.ARNAttributeComputedOnly(),
			"container_association_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 128),
					stringvalidator.RegexMatches(
						regexache.MustCompile(`^[a-zA-Z0-9-]+$`),
						"must contain only alphanumeric characters and hyphens",
					),
				},
			},
			names.AttrDescription: schema.StringAttribute{
				Optional: true,
			},
			"resolved_cidr_count": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			names.AttrType: schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ContainerMonitoringType](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"update_token": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"container_monitoring_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[containerMonitoringConfigurationModel](ctx),
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeBetween(1, 5),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"cluster_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
					Blocks: map[string]schema.Block{
						"attribute_filter": schema.ListNestedBlock{
							CustomType: fwtypes.NewListNestedObjectTypeOf[containerAttributeModel](ctx),
							NestedObject: schema.NestedBlockObject{
								Attributes: map[string]schema.Attribute{
									names.AttrKey: schema.StringAttribute{
										Required: true,
									},
									names.AttrValue: schema.StringAttribute{
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *containerAssociationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan containerAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	name := fwflex.StringValueFromFramework(ctx, plan.ContainerAssociationName)
	var input networkfirewall.CreateContainerAssociationInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	out, err := conn.CreateContainerAssociation(ctx, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, name)
		return
	}

	arn := aws.ToString(out.ContainerAssociationArn)

	outD, err := waitContainerAssociationCreated(ctx, conn, arn, r.CreateTimeout(ctx, plan.Timeouts))
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, outD, &plan))
	if resp.Diagnostics.HasError() {
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, plan))
}

func (r *containerAssociationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state containerAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := fwflex.StringValueFromFramework(ctx, state.ContainerAssociationARN)
	out, err := findContainerAssociationByARN(ctx, conn, arn)
	if retry.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, out, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	setTagsOut(ctx, out.Tags)

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state))
}

func (r *containerAssociationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var plan, state containerAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	diff, d := fwflex.Diff(ctx, plan, state)
	smerr.AddEnrich(ctx, &resp.Diagnostics, d)
	if resp.Diagnostics.HasError() {
		return
	}

	if diff.HasChanges() {
		arn := fwflex.StringValueFromFramework(ctx, state.ContainerAssociationARN)
		var input networkfirewall.UpdateContainerAssociationInput
		smerr.AddEnrich(ctx, &resp.Diagnostics, fwflex.Expand(ctx, plan, &input))
		if resp.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ContainerAssociationArn = aws.String(arn)
		input.UpdateToken = state.UpdateToken.ValueStringPointer()

		_, err := conn.UpdateContainerAssociation(ctx, &input)
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		outD, err := waitContainerAssociationUpdated(ctx, conn, arn, r.UpdateTimeout(ctx, plan.Timeouts))
		if err != nil {
			smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
			return
		}

		smerr.AddEnrich(ctx, &resp.Diagnostics, r.flatten(ctx, outD, &plan))
		if resp.Diagnostics.HasError() {
			return
		}
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &plan))
}

func (r *containerAssociationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	conn := r.Meta().NetworkFirewallClient(ctx)

	var state containerAssociationResourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
	if resp.Diagnostics.HasError() {
		return
	}

	arn := state.ContainerAssociationARN.ValueString()
	input := networkfirewall.DeleteContainerAssociationInput{
		ContainerAssociationArn: aws.String(arn),
	}
	_, err := conn.DeleteContainerAssociation(ctx, &input)
	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}

	if _, err := waitContainerAssociationDeleted(ctx, conn, arn, r.DeleteTimeout(ctx, state.Timeouts)); err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, arn)
		return
	}
}

func (r *containerAssociationResource) flatten(ctx context.Context, out *networkfirewall.DescribeContainerAssociationOutput, data *containerAssociationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	diags.Append(fwflex.Flatten(ctx, out, data)...)

	return diags
}

func waitContainerAssociationCreated(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeContainerAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ContainerAssociationStatusCreating),
		Target:                    enum.Slice(awstypes.ContainerAssociationStatusActive),
		Refresh:                   statusContainerAssociation(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeContainerAssociationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitContainerAssociationUpdated(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeContainerAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ContainerAssociationStatusUpdating),
		Target:                    enum.Slice(awstypes.ContainerAssociationStatusActive),
		Refresh:                   statusContainerAssociation(conn, arn),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeContainerAssociationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitContainerAssociationDeleted(ctx context.Context, conn *networkfirewall.Client, arn string, timeout time.Duration) (*networkfirewall.DescribeContainerAssociationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ContainerAssociationStatusActive, awstypes.ContainerAssociationStatusDeleting),
		Target:  []string{},
		Refresh: statusContainerAssociation(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*networkfirewall.DescribeContainerAssociationOutput); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusContainerAssociation(conn *networkfirewall.Client, arn string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		out, err := findContainerAssociationByARN(ctx, conn, arn)
		if retry.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		return out, string(out.Status), nil
	}
}

func findContainerAssociationByARN(ctx context.Context, conn *networkfirewall.Client, arn string) (*networkfirewall.DescribeContainerAssociationOutput, error) {
	input := networkfirewall.DescribeContainerAssociationInput{
		ContainerAssociationArn: aws.String(arn),
	}

	return findContainerAssociation(ctx, conn, &input)
}

func findContainerAssociation(ctx context.Context, conn *networkfirewall.Client, input *networkfirewall.DescribeContainerAssociationInput) (*networkfirewall.DescribeContainerAssociationOutput, error) {
	out, err := conn.DescribeContainerAssociation(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError: err,
		})
	}

	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError())
	}

	return out, nil
}

type containerAssociationResourceModel struct {
	framework.WithRegionModel
	ContainerAssociationARN           types.String                                                           `tfsdk:"container_association_arn"`
	ContainerAssociationName          types.String                                                           `tfsdk:"container_association_name"`
	ContainerMonitoringConfigurations fwtypes.ListNestedObjectValueOf[containerMonitoringConfigurationModel] `tfsdk:"container_monitoring_configuration"`
	Description                       types.String                                                           `tfsdk:"description"`
	ResolvedCIDRCount                 types.Int64                                                            `tfsdk:"resolved_cidr_count"`
	Tags                              tftags.Map                                                             `tfsdk:"tags"`
	TagsAll                           tftags.Map                                                             `tfsdk:"tags_all"`
	Timeouts                          timeouts.Value                                                         `tfsdk:"timeouts"`
	Type                              fwtypes.StringEnum[awstypes.ContainerMonitoringType]                   `tfsdk:"type"`
	UpdateToken                       types.String                                                           `tfsdk:"update_token"`
}

type containerMonitoringConfigurationModel struct {
	AttributeFilters fwtypes.ListNestedObjectValueOf[containerAttributeModel] `tfsdk:"attribute_filter"`
	ClusterARN       fwtypes.ARN                                              `tfsdk:"cluster_arn"`
}

type containerAttributeModel struct {
	Key   types.String `tfsdk:"key"`
	Value types.String `tfsdk:"value"`
}
