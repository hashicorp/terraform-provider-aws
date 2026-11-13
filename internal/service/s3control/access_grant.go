// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package s3control

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	awstypes "github.com/aws/aws-sdk-go-v2/service/s3control/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3control_access_grant", name="Access Grant")
// @Tags(identifierAttribute="access_grant_arn")
func newAccessGrantResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accessGrantResource{}

	return r, nil
}

type accessGrantResource struct {
	framework.ResourceWithModel[accessGrantResourceModel]
	framework.WithImportByID
}

func (r *accessGrantResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_grant_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_grant_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_grants_location_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
			},
			"grant_scope": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			names.AttrID: framework.IDAttribute(),
			"permission": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Permission](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"s3_prefix_type": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.S3PrefixType](),
				Optional:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]schema.Block{
			"access_grants_location_configuration": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[accessGrantsLocationConfigurationModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"s3_sub_prefix": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"grantee": schema.ListNestedBlock{
				CustomType: fwtypes.NewListNestedObjectTypeOf[granteeModel](ctx),
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"grantee_identifier": schema.StringAttribute{
							Required: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
						"grantee_type": schema.StringAttribute{
							CustomType: fwtypes.StringEnumType[awstypes.GranteeType](),
							Required:   true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
					listvalidator.IsRequired(),
				},
			},
		},
	}
}

func (r *accessGrantResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accessGrantResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AccountID.IsUnknown() {
		data.AccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().S3ControlClient(ctx)

	var input s3control.CreateAccessGrantInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	// "InvalidRequest: Invalid Grantee in the request".
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, s3PropagationTimeout, func(ctx context.Context) (any, error) {
		return conn.CreateAccessGrant(ctx, &input)
	}, errCodeInvalidRequest, "Invalid Grantee in the request")

	if err != nil {
		response.Diagnostics.AddError("creating S3 Access Grant", err.Error())

		return
	}

	// Set values for unknowns.
	output := outputRaw.(*s3control.CreateAccessGrantOutput)
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	id, err := data.setID()
	if err != nil {
		response.Diagnostics.Append(fwdiag.NewCreatingResourceIDErrorDiagnostic(err))
		return
	}
	data.ID = fwflex.StringValueToFramework(ctx, id)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accessGrantResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	output, err := findAccessGrantByTwoPartKey(ctx, conn, data.AccountID.ValueString(), data.AccessGrantID.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Access Grant (%s)", data.ID.ValueString()), err.Error())

		return
	}

	if isNullAccessGrantsLocationConfiguration(output.AccessGrantsLocationConfiguration) {
		output.AccessGrantsLocationConfiguration = nil
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accessGrantResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	input := s3control.DeleteAccessGrantInput{
		AccessGrantId: fwflex.StringFromFramework(ctx, data.AccessGrantID),
		AccountId:     fwflex.StringFromFramework(ctx, data.AccountID),
	}
	_, err := conn.DeleteAccessGrant(ctx, &input)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Access Grant (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findAccessGrantByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, grantID string) (*s3control.GetAccessGrantOutput, error) {
	input := s3control.GetAccessGrantInput{
		AccessGrantId: aws.String(grantID),
		AccountId:     aws.String(accountID),
	}

	return findAccessGrant(ctx, conn, &input)
}

func findAccessGrant(ctx context.Context, conn *s3control.Client, input *s3control.GetAccessGrantInput) (*s3control.GetAccessGrantOutput, error) {
	output, err := conn.GetAccessGrant(ctx, input)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

type accessGrantResourceModel struct {
	framework.WithRegionModel
	AccessGrantARN                    types.String                                                            `tfsdk:"access_grant_arn"`
	AccessGrantID                     types.String                                                            `tfsdk:"access_grant_id"`
	AccessGrantsLocationConfiguration fwtypes.ListNestedObjectValueOf[accessGrantsLocationConfigurationModel] `tfsdk:"access_grants_location_configuration"`
	AccessGrantsLocationID            types.String                                                            `tfsdk:"access_grants_location_id"`
	AccountID                         types.String                                                            `tfsdk:"account_id"`
	Grantee                           fwtypes.ListNestedObjectValueOf[granteeModel]                           `tfsdk:"grantee"`
	GrantScope                        types.String                                                            `tfsdk:"grant_scope"`
	ID                                types.String                                                            `tfsdk:"id"`
	Permission                        fwtypes.StringEnum[awstypes.Permission]                                 `tfsdk:"permission"`
	S3PrefixType                      fwtypes.StringEnum[awstypes.S3PrefixType]                               `tfsdk:"s3_prefix_type"`
	Tags                              tftags.Map                                                              `tfsdk:"tags"`
	TagsAll                           tftags.Map                                                              `tfsdk:"tags_all"`
}

type accessGrantsLocationConfigurationModel struct {
	S3SubPrefix types.String `tfsdk:"s3_sub_prefix"`
}

type granteeModel struct {
	GranteeIdentifier types.String                             `tfsdk:"grantee_identifier"`
	GranteeType       fwtypes.StringEnum[awstypes.GranteeType] `tfsdk:"grantee_type"`
}

const (
	accessGrantResourceIDPartCount = 2
)

func (data *accessGrantResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, accessGrantResourceIDPartCount, false)

	if err != nil {
		return err
	}

	data.AccountID = types.StringValue(parts[0])
	data.AccessGrantID = types.StringValue(parts[1])

	return nil
}

func (data *accessGrantResourceModel) setID() (string, error) {
	parts := []string{
		data.AccountID.ValueString(),
		data.AccessGrantID.ValueString(),
	}

	return flex.FlattenResourceId(parts, accessGrantResourceIDPartCount, false)
}

// API returns <AccessGrantsLocationConfiguration><S3SubPrefix></S3SubPrefix></AccessGrantsLocationConfiguration>.
func isNullAccessGrantsLocationConfiguration(v *awstypes.AccessGrantsLocationConfiguration) bool {
	if v == nil {
		return true
	}
	if aws.ToString(v.S3SubPrefix) == "" {
		return true
	}
	return false
}
