// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Access Grant")
// @Tags
func newAccessGrantResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accessGrantResource{}

	return r, nil
}

type accessGrantResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *accessGrantResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3control_access_grant"
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
			"permission": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.Permission](),
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID: framework.IDAttribute(),
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

	conn := r.Meta().S3ControlClient(ctx)

	if data.AccountID.ValueString() == "" {
		data.AccountID = types.StringValue(r.Meta().AccountID)
	}
	input := &s3control.CreateAccessGrantInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	// "InvalidRequest: Invalid Grantee in the request".
	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateAccessGrant(ctx, input)
	}, errCodeInvalidRequest, "Invalid Grantee in the request")

	if err != nil {
		response.Diagnostics.AddError("creating S3 Access Grant", err.Error())

		return
	}

	// Set values for unknowns.
	output := outputRaw.(*s3control.CreateAccessGrantOutput)
	data.AccessGrantARN = fwflex.StringToFramework(ctx, output.AccessGrantArn)
	data.AccessGrantID = fwflex.StringToFramework(ctx, output.AccessGrantId)
	data.GrantScope = fwflex.StringToFramework(ctx, output.GrantScope)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accessGrantResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	output, err := findAccessGrantByTwoPartKey(ctx, conn, data.AccountID.ValueString(), data.AccessGrantID.ValueString())

	if tfresource.NotFound(err) {
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

	tags, err := listTags(ctx, conn, data.AccessGrantARN.ValueString(), data.AccountID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("listing tags for S3 Access Grant (%s)", data.ID.ValueString()), err.Error())

		return
	}

	setTagsOut(ctx, Tags(tags))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new accessGrantResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	if oldTagsAll, newTagsAll := old.TagsAll, new.TagsAll; !newTagsAll.Equal(oldTagsAll) {
		if err := updateTags(ctx, conn, new.AccessGrantARN.ValueString(), new.AccountID.ValueString(), oldTagsAll, newTagsAll); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating tags for S3 Access Grant (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *accessGrantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accessGrantResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	_, err := conn.DeleteAccessGrant(ctx, &s3control.DeleteAccessGrantInput{
		AccessGrantId: fwflex.StringFromFramework(ctx, data.AccessGrantID),
		AccountId:     fwflex.StringFromFramework(ctx, data.AccountID),
	})

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Access Grant (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *accessGrantResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findAccessGrantByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, grantID string) (*s3control.GetAccessGrantOutput, error) {
	input := &s3control.GetAccessGrantInput{
		AccessGrantId: aws.String(grantID),
		AccountId:     aws.String(accountID),
	}

	output, err := conn.GetAccessGrant(ctx, input)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type accessGrantResourceModel struct {
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
	Tags                              types.Map                                                               `tfsdk:"tags"`
	TagsAll                           types.Map                                                               `tfsdk:"tags_all"`
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

func (data *accessGrantResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.AccountID.ValueString(), data.AccessGrantID.ValueString()}, accessGrantResourceIDPartCount, false)))
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
