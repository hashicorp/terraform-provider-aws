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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Access Grants Instance Resource Policy")
func newAccessGrantsInstanceResourcePolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accessGrantsInstanceResourcePolicyResource{}

	return r, nil
}

type accessGrantsInstanceResourcePolicyResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *accessGrantsInstanceResourcePolicyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_s3control_access_grants_instance_resource_policy"
}

func (r *accessGrantsInstanceResourcePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
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
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
		},
	}
}

func (r *accessGrantsInstanceResourcePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data accessGrantsInstanceResourcePolicyResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	if data.AccountID.ValueString() == "" {
		data.AccountID = types.StringValue(r.Meta().AccountID)
	}
	input := &s3control.PutAccessGrantsInstanceResourcePolicyInput{
		AccountId: flex.StringFromFramework(ctx, data.AccountID),
		Policy:    flex.StringFromFramework(ctx, data.Policy),
	}

	_, err := conn.PutAccessGrantsInstanceResourcePolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Access Grants Instance Resource Policy (%s)", data.AccountID.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantsInstanceResourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accessGrantsInstanceResourcePolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	output, err := findAccessGrantsInstanceResourcePolicy(ctx, conn, data.AccountID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Access Grants Instance Resource Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantsInstanceResourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new accessGrantsInstanceResourcePolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)

	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	input := &s3control.PutAccessGrantsInstanceResourcePolicyInput{
		AccountId: flex.StringFromFramework(ctx, new.AccountID),
		Policy:    flex.StringFromFramework(ctx, new.Policy),
	}

	_, err := conn.PutAccessGrantsInstanceResourcePolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating S3 Access Grants Instance Resource Policy (%s)", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *accessGrantsInstanceResourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data accessGrantsInstanceResourcePolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().S3ControlClient(ctx)

	_, err := conn.DeleteAccessGrantsInstanceResourcePolicy(ctx, &s3control.DeleteAccessGrantsInstanceResourcePolicyInput{
		AccountId: flex.StringFromFramework(ctx, data.AccountID),
	})

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Access Grants Instance Resource Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findAccessGrantsInstanceResourcePolicy(ctx context.Context, conn *s3control.Client, accountID string) (*s3control.GetAccessGrantsInstanceResourcePolicyOutput, error) {
	input := &s3control.GetAccessGrantsInstanceResourcePolicyInput{
		AccountId: aws.String(accountID),
	}

	output, err := conn.GetAccessGrantsInstanceResourcePolicy(ctx, input)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

type accessGrantsInstanceResourcePolicyResourceModel struct {
	AccountID types.String      `tfsdk:"account_id"`
	ID        types.String      `tfsdk:"id"`
	Policy    fwtypes.IAMPolicy `tfsdk:"policy"`
}

func (data *accessGrantsInstanceResourcePolicyResourceModel) InitFromID() error {
	data.AccountID = data.ID

	return nil
}

func (data *accessGrantsInstanceResourcePolicyResourceModel) setID() {
	data.ID = data.AccountID
}
