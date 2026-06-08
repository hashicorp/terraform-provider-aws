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
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_s3control_access_grants_instance", name="Access Grants Instance")
// @Tags(identifierAttribute="access_grants_instance_arn")
func newAccessGrantsInstanceResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &accessGrantsInstanceResource{}

	return r, nil
}

type accessGrantsInstanceResource struct {
	framework.ResourceWithModel[accessGrantsInstanceResourceModel]
	framework.WithImportByID
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
	if data.AccountID.IsUnknown() {
		data.AccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().S3ControlClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AccountID)
	var input s3control.CreateAccessGrantsInstanceInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreateAccessGrantsInstance(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating S3 Access Grants Instance (%s)", accountID), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, accountID)
	// Backwards compatibility, don't use AutoFlEx.
	data.AccessGrantsInstanceARN = fwflex.StringToFramework(ctx, output.AccessGrantsInstanceArn)
	data.AccessGrantsInstanceID = fwflex.StringToFramework(ctx, output.AccessGrantsInstanceId)
	data.IdentityCenterApplicationARN = fwflex.StringToFramework(ctx, output.IdentityCenterArn)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *accessGrantsInstanceResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data accessGrantsInstanceResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	data.AccountID = data.ID // From import.

	conn := r.Meta().S3ControlClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AccountID)
	output, err := findAccessGrantsInstanceByID(ctx, conn, accountID)

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading S3 Access Grants Instance (%s)", accountID), err.Error())

		return
	}

	// Set attributes for import.
	// Backwards compatibility, don't use AutoFlEx.
	data.AccessGrantsInstanceARN = fwflex.StringToFramework(ctx, output.AccessGrantsInstanceArn)
	data.AccessGrantsInstanceID = fwflex.StringToFramework(ctx, output.AccessGrantsInstanceId)
	data.IdentityCenterApplicationARN = fwflex.StringToFramework(ctx, output.IdentityCenterArn)

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

	if accountID, oldARN, newARN := fwflex.StringValueFromFramework(ctx, new.AccountID), old.IdentityCenterARN, new.IdentityCenterARN; !newARN.Equal(oldARN) {
		if !oldARN.IsNull() {
			if err := disassociateAccessGrantsInstanceIdentityCenterInstance(ctx, conn, accountID); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("dissociating S3 Access Grants Instance (%s) IAM Identity Center instance", accountID), err.Error())

				return
			}
		}

		if !newARN.IsNull() {
			if err := associateAccessGrantsInstanceIdentityCenterInstance(ctx, conn, accountID, newARN.ValueString()); err != nil {
				response.Diagnostics.AddError(fmt.Sprintf("associating S3 Access Grants Instance (%s) IAM Identity Center instance (%s)", accountID, newARN.ValueString()), err.Error())

				return
			}
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

	accountID := fwflex.StringValueFromFramework(ctx, data.AccountID)
	if !data.IdentityCenterARN.IsNull() {
		if err := disassociateAccessGrantsInstanceIdentityCenterInstance(ctx, conn, accountID); err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("dissociating S3 Access Grants Instance (%s) IAM Identity Center instance", accountID), err.Error())

			return
		}
	}

	input := s3control.DeleteAccessGrantsInstanceInput{
		AccountId: aws.String(accountID),
	}
	_, err := conn.DeleteAccessGrantsInstance(ctx, &input)

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusNotFound) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting S3 Access Grants Instance (%s)", accountID), err.Error())

		return
	}
}

func associateAccessGrantsInstanceIdentityCenterInstance(ctx context.Context, conn *s3control.Client, accountID, identityCenterARN string) error {
	input := s3control.AssociateAccessGrantsIdentityCenterInput{
		AccountId:         aws.String(accountID),
		IdentityCenterArn: aws.String(identityCenterARN),
	}

	_, err := conn.AssociateAccessGrantsIdentityCenter(ctx, &input)

	return err
}

func disassociateAccessGrantsInstanceIdentityCenterInstance(ctx context.Context, conn *s3control.Client, accountID string) error {
	input := s3control.DissociateAccessGrantsIdentityCenterInput{
		AccountId: aws.String(accountID),
	}

	_, err := conn.DissociateAccessGrantsIdentityCenter(ctx, &input)

	return err
}

func findAccessGrantsInstanceByID(ctx context.Context, conn *s3control.Client, accountID string) (*s3control.GetAccessGrantsInstanceOutput, error) {
	input := s3control.GetAccessGrantsInstanceInput{
		AccountId: aws.String(accountID),
	}

	return findAccessGrantsInstance(ctx, conn, &input)
}

func findAccessGrantsInstance(ctx context.Context, conn *s3control.Client, input *s3control.GetAccessGrantsInstanceInput) (*s3control.GetAccessGrantsInstanceOutput, error) {
	output, err := conn.GetAccessGrantsInstance(ctx, input)

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

type accessGrantsInstanceResourceModel struct {
	framework.WithRegionModel
	AccessGrantsInstanceARN      types.String `tfsdk:"access_grants_instance_arn"`
	AccessGrantsInstanceID       types.String `tfsdk:"access_grants_instance_id"`
	AccountID                    types.String `tfsdk:"account_id"`
	ID                           types.String `tfsdk:"id"`
	IdentityCenterApplicationARN types.String `tfsdk:"identity_center_application_arn"`
	IdentityCenterARN            fwtypes.ARN  `tfsdk:"identity_center_arn"`
	Tags                         tftags.Map   `tfsdk:"tags"`
	TagsAll                      tftags.Map   `tfsdk:"tags_all"`
}
