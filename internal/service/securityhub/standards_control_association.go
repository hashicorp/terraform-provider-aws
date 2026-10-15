// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	intflex "github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_securityhub_standards_control_association", name="Standards Control Association")
// @IdentityAttribute("security_control_id")
// @IdentityAttribute("standards_arn")
// @ImportIDHandler("standardsControlAssociationImportID", setIDAttribute=true)
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/securityhub/types;awstypes;awstypes.StandardsControlAssociationSummary")
// @Testing(serialize=true)
// @Testing(preIdentityVersion="v6.42.0")
// @Testing(generator=false)
// @Testing(checkDestroyNoop=true)
// @Testing(importStateIdFunc=testAccCheckStandardsControlAssociationImportStateIDFunc)
func newStandardsControlAssociationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &standardsControlAssociationResource{}

	return r, nil
}

type standardsControlAssociationResource struct {
	framework.ResourceWithModel[standardsControlAssociationResourceModel]
	framework.WithNoOpDelete
	framework.WithImportByIdentity
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

	var apiObject awstypes.StandardsControlAssociationUpdate
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &apiObject)...)
	if response.Diagnostics.HasError() {
		return
	}

	securityControlID, standardsARN := fwflex.StringValueFromFramework(ctx, data.SecurityControlID), fwflex.StringValueFromFramework(ctx, data.StandardsARN)
	id := standardsControlAssociationCreateResourceID(securityControlID, standardsARN)
	input := securityhub.BatchUpdateStandardsControlAssociationsInput{
		StandardsControlAssociationUpdates: []awstypes.StandardsControlAssociationUpdate{apiObject},
	}

	output, err := conn.BatchUpdateStandardsControlAssociations(ctx, &input)

	if err == nil {
		err = unprocessedAssociationUpdatesError(output.UnprocessedAssociationUpdates)
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating SecurityHub Standards Control Association (%s)", id), err.Error())
		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, id)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *standardsControlAssociationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data standardsControlAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	securityControlID, standardsARN, err := standardsControlAssociationParseResourceID(id)
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))
		return
	}

	conn := r.Meta().SecurityHubClient(ctx)

	output, err := findStandardsControlAssociationByTwoPartKey(ctx, conn, securityControlID, standardsARN)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading SecurityHub Standards Control Association (%s)", id), err.Error())
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

	id := fwflex.StringValueFromFramework(ctx, data.ID)
	var apiObject awstypes.StandardsControlAssociationUpdate
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &apiObject)...)
	if response.Diagnostics.HasError() {
		return
	}

	input := securityhub.BatchUpdateStandardsControlAssociationsInput{
		StandardsControlAssociationUpdates: []awstypes.StandardsControlAssociationUpdate{apiObject},
	}

	output, err := conn.BatchUpdateStandardsControlAssociations(ctx, &input)

	if err == nil {
		err = unprocessedAssociationUpdatesError(output.UnprocessedAssociationUpdates)
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating SecurityHub Standards Control Association (%s)", id), err.Error())
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
	framework.WithRegionModel
	AssociationStatus fwtypes.StringEnum[awstypes.AssociationStatus] `tfsdk:"association_status"`
	ID                types.String                                   `tfsdk:"id"`
	SecurityControlID types.String                                   `tfsdk:"security_control_id"`
	StandardsARN      fwtypes.ARN                                    `tfsdk:"standards_arn"`
	UpdatedReason     types.String                                   `tfsdk:"updated_reason"`
}

const (
	standardsControlAssociationResourceIDPartCount = 2
)

func standardsControlAssociationCreateResourceID(securityControlID, standardsARN string) string {
	id, _ := intflex.FlattenResourceId([]string{securityControlID, standardsARN}, standardsControlAssociationResourceIDPartCount, false)
	return id
}

func standardsControlAssociationParseResourceID(id string) (string, string, error) {
	parts, err := intflex.ExpandResourceId(id, standardsControlAssociationResourceIDPartCount, false)
	if err != nil {
		return "", "", err
	}

	return parts[0], parts[1], nil
}

func findStandardsControlAssociationByTwoPartKey(ctx context.Context, conn *securityhub.Client, securityControlID string, standardsARN string) (*awstypes.StandardsControlAssociationSummary, error) {
	input := securityhub.ListStandardsControlAssociationsInput{
		SecurityControlId: aws.String(securityControlID),
	}

	return findStandardsControlAssociation(ctx, conn, &input, func(v awstypes.StandardsControlAssociationSummary) bool {
		return aws.ToString(v.StandardsArn) == standardsARN
	})
}

func findStandardsControlAssociation(ctx context.Context, conn *securityhub.Client, input *securityhub.ListStandardsControlAssociationsInput, filter tfslices.Predicate[awstypes.StandardsControlAssociationSummary]) (*awstypes.StandardsControlAssociationSummary, error) {
	output, err := findStandardsControlAssociations(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findStandardsControlAssociations(ctx context.Context, conn *securityhub.Client, input *securityhub.ListStandardsControlAssociationsInput, filter tfslices.Predicate[awstypes.StandardsControlAssociationSummary]) ([]awstypes.StandardsControlAssociationSummary, error) {
	var output []awstypes.StandardsControlAssociationSummary

	pages := securityhub.NewListStandardsControlAssociationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, errCodeInvalidAccessException, "not subscribed to AWS Security Hub") {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.StandardsControlAssociationSummaries {
			if filter(v) {
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
			id := standardsControlAssociationCreateResourceID(aws.ToString(v.SecurityControlId), aws.ToString(v.StandardsArn))
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

var (
	_ inttypes.ImportIDParser           = standardsControlAssociationImportID{}
	_ inttypes.FrameworkImportIDCreator = standardsControlAssociationImportID{}
)

type standardsControlAssociationImportID struct{}

func (standardsControlAssociationImportID) Parse(id string) (string, map[string]any, error) {
	securityControlID, standardsARN, err := standardsControlAssociationParseResourceID(id)
	if err != nil {
		return "", nil, err
	}

	result := map[string]any{
		"security_control_id": securityControlID,
		"standards_arn":       standardsARN,
	}

	return id, result, nil
}

func (standardsControlAssociationImportID) Create(ctx context.Context, state tfsdk.State) string {
	var securityControlID types.String
	state.GetAttribute(ctx, path.Root("security_control_id"), &securityControlID)
	var standardsARN fwtypes.ARN
	state.GetAttribute(ctx, path.Root("standards_arn"), &standardsARN)

	return standardsControlAssociationCreateResourceID(fwflex.StringValueFromFramework(ctx, securityControlID), fwflex.StringValueFromFramework(ctx, standardsARN))
}
