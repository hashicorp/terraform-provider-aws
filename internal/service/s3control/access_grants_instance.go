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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Access Grants Instance")
// @Tags
func newAccessGrantsInstanceResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accessGrantsInstanceResource{}

	return r, nil
}

type accessGrantsInstanceResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *accessGrantsInstanceResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3control_access_grants_instance"
}

func (r *accessGrantsInstanceResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_grants_instance_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_grants_instance_id": schema.StringAttribute{
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
			names.AttrID: framework.IDAttribute(),
			"identity_center_application_arn": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identity_center_arn": schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Optional:   true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
	}
}

func (r *accessGrantsInstanceResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accessGrantsInstanceResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	if data.AccountID.ValueString() == "" {
		data.AccountID = types.StringValue(r.Meta().AccountID)
	}
	input := &s3control.CreateAccessGrantsInstanceInput{
		AccountId:         flex.StringFromFramework(ctx, data.AccountID),
		IdentityCenterArn: flex.StringFromFramework(ctx, data.IdentityCenterARN),
		Tags:              getTagsIn(ctx),
	}

	output, err := conn.CreateAccessGrantsInstance(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Access Grants Instance (%s)", data.AccountID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.AccessGrantsInstanceARN = flex.StringToFramework(ctx, output.AccessGrantsInstanceArn)
	data.AccessGrantsInstanceID = flex.StringToFramework(ctx, output.AccessGrantsInstanceId)
	data.IdentityCenterApplicationARN = flex.StringToFramework(ctx, output.IdentityCenterArn)
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantsInstanceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accessGrantsInstanceResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	output, err := findAccessGrantsInstance(ctx, conn, data.AccountID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Access Grants Instance (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	data.AccessGrantsInstanceARN = flex.StringToFramework(ctx, output.AccessGrantsInstanceArn)
	data.AccessGrantsInstanceID = flex.StringToFramework(ctx, output.AccessGrantsInstanceId)
	data.IdentityCenterApplicationARN = flex.StringToFramework(ctx, output.IdentityCenterArn)

	tags, err := listTags(ctx, conn, data.AccessGrantsInstanceARN.ValueString(), data.AccountID.ValueString())

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("listing tags for S3 Access Grants Instance (%s)", data.ID.ValueString()), err.Error())

		return
	}

	setTagsOut(ctx, Tags(tags))

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantsInstanceResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new accessGrantsInstanceResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	if oldARN, newARN := old.IdentityCenterARN, new.IdentityCenterARN; !newARN.Equal(oldARN) {
		if !oldARN.IsNull() {
			if err := disassociateAccessGrantsInstanceIdentityCenterInstance(ctx, conn, old.ID.ValueString()); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("dissociating S3 Access Grants Instance (%s) IAM Identity Center instance", old.ID.ValueString()), err.Error())

				return
			}
		}

		if !newARN.IsNull() {
			if err := associateAccessGrantsInstanceIdentityCenterInstance(ctx, conn, new.ID.ValueString(), newARN.ValueString()); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("associating S3 Access Grants Instance (%s) IAM Identity Center instance (%s)", new.ID.ValueString(), newARN.ValueString()), err.Error())

				return
			}
		}
	}

	if oldTagsAll, newTagsAll := old.TagsAll, new.TagsAll; !newTagsAll.Equal(oldTagsAll) {
		if err := updateTags(ctx, conn, new.AccessGrantsInstanceARN.ValueString(), new.AccountID.ValueString(), oldTagsAll, newTagsAll); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating tags for S3 Access Grants Instance (%s)", new.ID.ValueString()), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *accessGrantsInstanceResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accessGrantsInstanceResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	if !data.IdentityCenterARN.IsNull() {
		if err := disassociateAccessGrantsInstanceIdentityCenterInstance(ctx, conn, data.ID.ValueString()); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("dissociating S3 Access Grants Instance (%s) IAM Identity Center instance", data.ID.ValueString()), err.Error())

			return
		}
	}

	_, err := conn.DeleteAccessGrantsInstance(ctx, &s3control.DeleteAccessGrantsInstanceInput{
		AccountId: flex.StringFromFramework(ctx, data.AccountID),
	})

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Access Grants Instance (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func (r *accessGrantsInstanceResource) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func associateAccessGrantsInstanceIdentityCenterInstance(ctx context.Context, conn *s3control.Client, accountID, identityCenterARN string) error {
	input := &s3control.AssociateAccessGrantsIdentityCenterInput{
		AccountId:         aws.String(accountID),
		IdentityCenterArn: aws.String(identityCenterARN),
	}

	_, err := conn.AssociateAccessGrantsIdentityCenter(ctx, input)

	return err
}

func disassociateAccessGrantsInstanceIdentityCenterInstance(ctx context.Context, conn *s3control.Client, accountID string) error {
	input := &s3control.DissociateAccessGrantsIdentityCenterInput{
		AccountId: aws.String(accountID),
	}

	_, err := conn.DissociateAccessGrantsIdentityCenter(ctx, input)

	return err
}

func findAccessGrantsInstance(ctx context.Context, conn *s3control.Client, accountID string) (*s3control.GetAccessGrantsInstanceOutput, error) {
	input := &s3control.GetAccessGrantsInstanceInput{
		AccountId: aws.String(accountID),
	}

	output, err := conn.GetAccessGrantsInstance(ctx, input)

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

type accessGrantsInstanceResourceModel struct {
	AccessGrantsInstanceARN      types.String `tfsdk:"access_grants_instance_arn"`
	AccessGrantsInstanceID       types.String `tfsdk:"access_grants_instance_id"`
	AccountID                    types.String `tfsdk:"account_id"`
	ID                           types.String `tfsdk:"id"`
	IdentityCenterApplicationARN types.String `tfsdk:"identity_center_application_arn"`
	IdentityCenterARN            fwtypes.ARN  `tfsdk:"identity_center_arn"`
	Tags                         types.Map    `tfsdk:"tags"`
	TagsAll                      types.Map    `tfsdk:"tags_all"`
}

func (data *accessGrantsInstanceResourceModel) InitFromID() error {
	data.AccountID = data.ID

	return nil
}

func (data *accessGrantsInstanceResourceModel) setID() {
	data.ID = data.AccountID
}
