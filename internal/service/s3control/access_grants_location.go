// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

// @FrameworkResource(name="Access Grants Location")
// @Tags
func newAccessGrantsLocationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accessGrantsLocationResource{}

	return r, nil
}

type accessGrantsLocationResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *accessGrantsLocationResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3control_access_grants_location"
}

func (r *accessGrantsLocationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_grants_location_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_grants_location_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			names.AttrIAMRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
			"location_scope": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrID:      framework.IDAttribute(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *accessGrantsLocationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accessGrantsLocationResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	if data.AccountID.ValueString() == "" {
		data.AccountID = types.StringValue(r.Meta().AccountID)
	}
	input := &s3control.CreateAccessGrantsLocationInput{}
	response.Diagnostics.Append(fwflex.Expand(ctx, data, input)...)
	if response.Diagnostics.HasError() {
		return
	}

	input.Tags = getTagsIn(ctx)

	outputRaw, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateAccessGrantsLocation(ctx, input)
	}, errCodeInvalidIAMRole)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Access Grants Location (%s)", data.LocationScope.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	output := outputRaw.(*s3control.CreateAccessGrantsLocationOutput)
	data.AccessGrantsLocationARN = fwflex.StringToFramework(ctx, output.AccessGrantsLocationArn)
	data.AccessGrantsLocationID = fwflex.StringToFramework(ctx, output.AccessGrantsLocationId)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantsLocationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accessGrantsLocationResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	output, err := findAccessGrantsLocationByTwoPartKey(ctx, conn, data.AccountID.ValueString(), data.AccessGrantsLocationID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Access Grants Location (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	tags, err := listTags(ctx, conn, data.AccessGrantsLocationARN.ValueString(), data.AccountID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("listing tags for S3 Access Grants Location (%s)", data.ID.ValueString()), err.Error())

		return
	}

	setTagsOut(ctx, Tags(tags))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantsLocationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new accessGrantsLocationResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	if !new.IAMRoleARN.Equal(old.IAMRoleARN) {
		input := &s3control.UpdateAccessGrantsLocationInput{}
		response.Diagnostics.Append(fwflex.Expand(ctx, new, input)...)
		if response.Diagnostics.HasError() {
			return
		}

		_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
			return conn.UpdateAccessGrantsLocation(ctx, input)
		}, errCodeInvalidIAMRole)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating S3 Access Grants Location (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	if oldTagsAll, newTagsAll := old.TagsAll, new.TagsAll; !newTagsAll.Equal(oldTagsAll) {
		if err := updateTags(ctx, conn, new.AccessGrantsLocationARN.ValueString(), new.AccountID.ValueString(), oldTagsAll, newTagsAll); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating tags for S3 Access Grants Location (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *accessGrantsLocationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accessGrantsLocationResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	input := &s3control.DeleteAccessGrantsLocationInput{
		AccessGrantsLocationId: fwflex.StringFromFramework(ctx, data.AccessGrantsLocationID),
		AccountId:              fwflex.StringFromFramework(ctx, data.AccountID),
	}

	// "AccessGrantsLocationNotEmptyError: Please delete access grants before deleting access grants location".
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, propagationTimeout, func() (interface{}, error) {
		return conn.DeleteAccessGrantsLocation(ctx, input)
	}, errCodeAccessGrantsLocationNotEmptyError)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Access Grants Location (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *accessGrantsLocationResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findAccessGrantsLocationByTwoPartKey(ctx context.Context, conn *s3control.Client, accountID, locationID string) (*s3control.GetAccessGrantsLocationOutput, error) {
	input := &s3control.GetAccessGrantsLocationInput{
		AccessGrantsLocationId: aws.String(locationID),
		AccountId:              aws.String(accountID),
	}

	output, err := conn.GetAccessGrantsLocation(ctx, input)

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

type accessGrantsLocationResourceModel struct {
	AccessGrantsLocationARN types.String `tfsdk:"access_grants_location_arn"`
	AccessGrantsLocationID  types.String `tfsdk:"access_grants_location_id"`
	AccountID               types.String `tfsdk:"account_id"`
	IAMRoleARN              fwtypes.ARN  `tfsdk:"iam_role_arn"`
	ID                      types.String `tfsdk:"id"`
	LocationScope           types.String `tfsdk:"location_scope"`
	Tags                    types.Map    `tfsdk:"tags"`
	TagsAll                 types.Map    `tfsdk:"tags_all"`
}

const (
	accessGrantsLocationResourceIDPartCount = 2
)

func (data *accessGrantsLocationResourceModel) InitFromID() error {
	id := data.ID.ValueString()
	parts, err := flex.ExpandResourceId(id, accessGrantsLocationResourceIDPartCount, false)

	if err != nil {
		return err
	}

	data.AccountID = types.StringValue(parts[0])
	data.AccessGrantsLocationID = types.StringValue(parts[1])

	return nil
}

func (data *accessGrantsLocationResourceModel) setID() {
	data.ID = types.StringValue(errs.Must(flex.FlattenResourceId([]string{data.AccountID.ValueString(), data.AccessGrantsLocationID.ValueString()}, accessGrantsLocationResourceIDPartCount, false)))
}
