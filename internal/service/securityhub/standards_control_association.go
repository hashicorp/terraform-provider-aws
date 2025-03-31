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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_securityhub_standards_control_association", name="Standards Control Association")
func newStandardsControlAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &standardsControlAssociationResource{}

	return r, nil
}

type standardsControlAssociationResource struct {
	framework.ResourceWithConfigure
	framework.WithNoOpDelete
}

func (r *standardsControlAssociationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
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

func (r *standardsControlAssociationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data standardsControlAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
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

	if err == nil {
		err = unprocessedAssociationUpdatesError(output.UnprocessedAssociationUpdates)
	}

	if err != nil {
		response.Diagnostics.AddError("creating Standards Control Association", err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *standardsControlAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data standardsControlAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(ctx); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	securityControlID, standardsARN := data.SecurityControlID.ValueString(), data.StandardsARN.ValueString()
	output, err := findStandardsControlAssociationByTwoPartKey(ctx, conn, securityControlID, standardsARN)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SecurityHub Standards Control Association (%s/%s)", securityControlID, standardsARN), err.Error())

		return
	}

	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *standardsControlAssociationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var data standardsControlAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
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

	if err == nil {
		err = unprocessedAssociationUpdatesError(output.UnprocessedAssociationUpdates)
	}

	if err != nil {
		response.Diagnostics.AddError("updating Standards Control Association", err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *standardsControlAssociationResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var data standardsControlAssociationResourceModel
	response.Diagnostics.Append(request.Config.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if data.AssociationStatus == fwtypes.StringEnumValue(awstypes.AssociationStatusEnabled) {
		return
	}

	if !data.UpdatedReason.IsNull() {
		return
	}

	response.Diagnostics.Append(
		fwdiag.NewAttributeRequiredWhenError(
			path.Root("updated_reason"),
			path.Root("association_status"),
			data.AssociationStatus.ValueString(),
		),
	)
}

type standardsControlAssociationResourceModel struct {
	AssociationStatus fwtypes.StringEnum[awstypes.AssociationStatus] `tfsdk:"association_status"`
	ID                types.String                                   `tfsdk:"id"`
	SecurityControlID types.String                                   `tfsdk:"security_control_id"`
	StandardsARN      fwtypes.ARN                                    `tfsdk:"standards_arn"`
	UpdatedReason     types.String                                   `tfsdk:"updated_reason"`
}

const (
	standardsControlAssociationResourceIDPartCount = 2
)

func (m *standardsControlAssociationResourceModel) InitFromID(ctx context.Context) error {
	parts, err := flex.ExpandResourceId(m.ID.ValueString(), standardsControlAssociationResourceIDPartCount, false)
	if err != nil {
		return err
	}

	m.SecurityControlID = types.StringValue(parts[0])
	m.StandardsARN = fwtypes.ARNValue(parts[1])

	return nil
}

func (m *standardsControlAssociationResourceModel) setID() {
	id, _ := standardsControlAssociationCreateResourceID(m.SecurityControlID.ValueString(), m.StandardsARN.ValueString())
	m.ID = types.StringValue(id)
}

func standardsControlAssociationCreateResourceID(securityControlID, standardsARN string) (string, error) {
	return flex.FlattenResourceId([]string{securityControlID, standardsARN}, standardsControlAssociationResourceIDPartCount, false)
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

func unprocessedAssociationUpdatesError(apiObjects []awstypes.UnprocessedStandardsControlAssociationUpdate) error {
	var errs []error

	for _, apiObject := range apiObjects {
		err := unprocessedAssociationUpdateError(&apiObject)
		if v := apiObject.StandardsControlAssociationUpdate; v != nil {
			id, _ := standardsControlAssociationCreateResourceID(aws.ToString(v.SecurityControlId), aws.ToString(v.StandardsArn))
			err = fmt.Errorf("%s: %w", id, err)
		}
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func unprocessedAssociationUpdateError(apiObject *awstypes.UnprocessedStandardsControlAssociationUpdate) error {
	if apiObject == nil {
		return nil
	}

	return fmt.Errorf("%s: %s", apiObject.ErrorCode, aws.ToString(apiObject.ErrorReason))
}
