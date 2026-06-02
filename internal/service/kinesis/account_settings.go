// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_kinesis_account_settings", name="Account Settings")
// @SingletonIdentity(identityDuplicateAttributes="id")
// @Testing(hasNoPreExistingResource=true)
// @Testing(generator=false)
// @Testing(serialize=true)
// @Testing(checkDestroyNoop=true)
func newAccountSettingsResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &accountSettingsResource{}
	return r, nil
}

type accountSettingsResource struct {
	framework.ResourceWithModel[accountSettingsResourceModel]
	framework.WithNoOpDelete
	framework.WithImportByIdentity
}

func (r *accountSettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttributeDeprecatedWithAlternate(path.Root(names.AttrRegion)),
		},
		Blocks: map[string]schema.Block{
			"minimum_throughput_billing_commitment": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[minimumThroughputBillingCommitmentModel](ctx),
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"earliest_allowed_end_at": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"ended_at": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						"started_at": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
						},
						names.AttrStatus: schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MinimumThroughputBillingCommitmentInputStatus](),
							Required:   true,
						},
						"status_actual": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.MinimumThroughputBillingCommitmentOutputStatus](),
							Computed:   true,
						},
					},
				},
			},
		},
	}
}

func (r *accountSettingsResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accountSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().KinesisClient(ctx)

	var input kinesis.UpdateAccountSettingsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateAccountSettings(ctx, &input)

	switch status := input.MinimumThroughputBillingCommitment.Status; {
	case status == awstypes.MinimumThroughputBillingCommitmentInputStatusDisabled && errs.IsAErrorMessageContains[*awstypes.ValidationException](err, "Account isn't opted in"):
	case err != nil:
		response.Diagnostics.AddError("creating Kinesis Account Settings", err.Error())
		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, r.Meta().Region(ctx))
	output, err := findAccountSettings(ctx, conn)

	if err != nil {
		response.Diagnostics.AddError("reading Kinesis Account Settings", err.Error())
		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *accountSettingsResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accountSettingsResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().KinesisClient(ctx)

	output, err := findAccountSettings(ctx, conn)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading Kinesis Account Settings", err.Error())
		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accountSettingsResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data accountSettingsResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().KinesisClient(ctx)

	var input kinesis.UpdateAccountSettingsInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	output, err := conn.UpdateAccountSettings(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError("updating Kinesis Account Settings", err.Error())
		return
	}

	// Set values for unknowns.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func findAccountSettings(ctx context.Context, conn *kinesis.Client) (*kinesis.DescribeAccountSettingsOutput, error) {
	var input kinesis.DescribeAccountSettingsInput
	output, err := conn.DescribeAccountSettings(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type accountSettingsResourceModel struct {
	framework.WithRegionModel
	MinimumThroughputBillingCommitment fwtypes.ListNestedObjectValueOf[minimumThroughputBillingCommitmentModel] `tfsdk:"minimum_throughput_billing_commitment"`
	ID                                 types.String                                                             `tfsdk:"id"`
}

type minimumThroughputBillingCommitmentModel struct {
	EarliestAllowedEndAt timetypes.RFC3339                                                           `tfsdk:"earliest_allowed_end_at"`
	EndedAt              timetypes.RFC3339                                                           `tfsdk:"ended_at"`
	StartedAt            timetypes.RFC3339                                                           `tfsdk:"started_at"`
	Status               fwtypes.StringEnum[awstypes.MinimumThroughputBillingCommitmentInputStatus]  `tfsdk:"status"`
	StatusActual         fwtypes.StringEnum[awstypes.MinimumThroughputBillingCommitmentOutputStatus] `tfsdk:"status_actual"`
}

var (
	_ fwflex.Flattener = &minimumThroughputBillingCommitmentModel{}
)

func (m *minimumThroughputBillingCommitmentModel) Flatten(ctx context.Context, v any) diag.Diagnostics {
	var diags diag.Diagnostics
	switch t := v.(type) {
	case awstypes.MinimumThroughputBillingCommitmentOutput:
		m.EarliestAllowedEndAt = timetypes.NewRFC3339TimePointerValue(t.EarliestAllowedEndAt)
		m.EndedAt = timetypes.NewRFC3339TimePointerValue(t.EndedAt)
		m.StartedAt = timetypes.NewRFC3339TimePointerValue(t.StartedAt)
		switch t.Status {
		case awstypes.MinimumThroughputBillingCommitmentOutputStatusEnabled:
			m.Status = fwtypes.StringEnumValue(awstypes.MinimumThroughputBillingCommitmentInputStatusEnabled)
		default:
			m.Status = fwtypes.StringEnumValue(awstypes.MinimumThroughputBillingCommitmentInputStatusDisabled)
		}
		m.StatusActual = fwtypes.StringEnumValue(t.Status)

	default:
		diags.AddError(
			"Unsupported Type",
			fmt.Sprintf("minimum throughput billing commitment flatten: %T", v),
		)
	}
	return diags
}
