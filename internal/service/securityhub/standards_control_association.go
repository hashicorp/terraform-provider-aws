// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	awstypes "github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	autoflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_securityhub_standards_control_association", name="Standards Control Association")
func newResourceStandardsControlAssociation(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceStandardsControlAssociation{}

	return r, nil
}

const (
	ResNameStandardsControlAssociation = "Standards Control Association"
)

type resourceStandardsControlAssociation struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (r *resourceStandardsControlAssociation) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_securityhub_standards_control_association"
}

func (r *resourceStandardsControlAssociation) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"association_status": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.AssociationStatus](),
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
			"security_control_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"standards_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"updated_reason": schema.StringAttribute{
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
		},
	}
}

func (r *resourceStandardsControlAssociation) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data resourceStandardsControlAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	input := &securityhub.BatchUpdateStandardsControlAssociationsInput{
		StandardsControlAssociationUpdates: []awstypes.StandardsControlAssociationUpdate{
			{
				AssociationStatus: awstypes.AssociationStatus(data.AssociationStatus.ValueString()),
				SecurityControlId: data.SecurityControlID.ValueStringPointer(),
				StandardsArn:      data.StandardsARN.ValueStringPointer(),
				UpdatedReason:     data.UpdatedReason.ValueStringPointer(),
			},
		},
	}

	output, err := conn.BatchUpdateStandardsControlAssociations(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("creating Standards Control Association", err.Error())

		return
	}

	if len(output.UnprocessedAssociationUpdates) > 0 {
		resp.Diagnostics.AddError("creating Standards Control Association", errors.New("unprocessed association updates").Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	resp.Diagnostics.Append(resp.State.Set(ctx, data)...)
}

func (r *resourceStandardsControlAssociation) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data resourceStandardsControlAssociationModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(ctx); err != nil {
		resp.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	securityControlID, standardsARN := data.SecurityControlID.ValueString(), data.StandardsARN.ValueString()
	output, err := findStandardsControlAssociationByTwoPartKey(ctx, conn, securityControlID, standardsARN)

	if tfresource.NotFound(err) {
		resp.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		resp.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading SecurityHub Standards Control Association (%s/%s)", securityControlID, standardsARN), err.Error())

		return
	}

	resp.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceStandardsControlAssociation) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data resourceStandardsControlAssociationModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	input := &securityhub.BatchUpdateStandardsControlAssociationsInput{
		StandardsControlAssociationUpdates: []awstypes.StandardsControlAssociationUpdate{
			{
				AssociationStatus: awstypes.AssociationStatus(data.AssociationStatus.ValueString()),
				SecurityControlId: data.SecurityControlID.ValueStringPointer(),
				StandardsArn:      data.StandardsARN.ValueStringPointer(),
				UpdatedReason:     data.UpdatedReason.ValueStringPointer(),
			},
		},
	}

	output, err := conn.BatchUpdateStandardsControlAssociations(ctx, input)

	if err != nil {
		resp.Diagnostics.AddError("creating Standards Control Association", err.Error())

		return
	}

	if len(output.UnprocessedAssociationUpdates) > 0 {
		resp.Diagnostics.AddError("creating Standards Control Association", errors.New("unprocessed association updates").Error())

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *resourceStandardsControlAssociation) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data resourceStandardsControlAssociationModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.AssociationStatus == fwtypes.StringEnumValue(awstypes.AssociationStatusEnabled) {
		return
	}

	if !data.UpdatedReason.IsNull() {
		return
	}

	resp.Diagnostics.Append(
		fwdiag.NewAttributeRequiredWhenError(
			path.Root("updated_reason"),
			path.Root("association_status"),
			data.AssociationStatus.ValueString(),
		),
	)
}

type resourceStandardsControlAssociationModel struct {
	AssociationStatus fwtypes.StringEnum[awstypes.AssociationStatus] `tfsdk:"association_status"`
	ID                types.String                                   `tfsdk:"id"`
	SecurityControlID types.String                                   `tfsdk:"security_control_id"`
	StandardsARN      fwtypes.ARN                                    `tfsdk:"standards_arn"`
	UpdatedReason     types.String                                   `tfsdk:"updated_reason"`
}

const (
	standardsControlAssociationResourceIDPartCount = 2
)

func (m *resourceStandardsControlAssociationModel) InitFromID(ctx context.Context) error {
	parts, err := autoflex.ExpandResourceId(m.ID.ValueString(), standardsControlAssociationResourceIDPartCount, false)
	if err != nil {
		return err
	}

	m.SecurityControlID = types.StringValue(parts[0])
	m.StandardsARN = fwtypes.ARNValue(parts[1])

	return nil
}

func (m *resourceStandardsControlAssociationModel) setID() {
	m.ID = types.StringValue(errs.Must(autoflex.FlattenResourceId([]string{m.SecurityControlID.ValueString(), m.StandardsARN.ValueString()}, standardsControlAssociationResourceIDPartCount, false)))
}

func findStandardsControlAssociationByTwoPartKey(ctx context.Context, conn *securityhub.Client, securityControlID string, standardsARN string) (*awstypes.StandardsControlAssociationSummary, error) {
	input := &securityhub.ListStandardsControlAssociationsInput{
		SecurityControlId: aws.String(securityControlID),
	}

	return findStandardsControlAssociation(ctx, conn, input, func(v *awstypes.StandardsControlAssociationSummary) bool {
		return aws.ToString(v.StandardsArn) == standardsARN
	})
}

func findStandardsControlAssociation(ctx context.Context, conn *securityhub.Client, input *securityhub.ListStandardsControlAssociationsInput, filter tfslices.Predicate[*awstypes.StandardsControlAssociationSummary]) (*awstypes.StandardsControlAssociationSummary, error) {
	output, err := findStandardsControlAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findStandardsControlAssociations(ctx context.Context, conn *securityhub.Client, input *securityhub.ListStandardsControlAssociationsInput, filter tfslices.Predicate[*awstypes.StandardsControlAssociationSummary]) ([]awstypes.StandardsControlAssociationSummary, error) {
	var output []awstypes.StandardsControlAssociationSummary

	pages := securityhub.NewListStandardsControlAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.StandardsControlAssociationSummaries {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}
