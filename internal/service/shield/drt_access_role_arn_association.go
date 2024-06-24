// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="DRT Role ARN Association")
func newDRTAccessRoleARNAssociationResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceDRTAccessRoleARNAssociation{}

	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(30 * time.Minute)
	r.SetDefaultDeleteTimeout(30 * time.Minute)

	return r, nil
}

type resourceDRTAccessRoleARNAssociation struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
	framework.WithTimeouts
}

func (r *resourceDRTAccessRoleARNAssociation) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_shield_drt_access_role_arn_association"
}

func (r *resourceDRTAccessRoleARNAssociation) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrRoleARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
			},
		},
		Blocks: map[string]schema.Block{
			names.AttrTimeouts: timeouts.Block(ctx, timeouts.Opts{
				Create: true,
				Update: true,
				Delete: true,
			}),
		},
	}
}

func (r *resourceDRTAccessRoleARNAssociation) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data drtAccessRoleARNAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	roleARN := data.RoleARN.ValueString()
	input := &shield.AssociateDRTRoleInput{
		RoleArn: aws.String(roleARN),
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.AssociateDRTRole(ctx, input)
	}, "role does not have a valid DRT managed policy")

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Shield DRT Role ARN Association (%s)", roleARN), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = types.StringValue(r.Meta().AccountID)

	_, err = tfresource.RetryWhenNotFound(ctx, r.CreateTimeout(ctx, data.Timeouts), func() (interface{}, error) {
		return findDRTRoleARNAssociation(ctx, conn, roleARN)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Shield DRT Role ARN Association (%s) create", roleARN), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *resourceDRTAccessRoleARNAssociation) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data drtAccessRoleARNAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	output, err := findDRTAccess(ctx, conn)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError("reading Shield DRT Access", err.Error())

		return
	}

	data.RoleARN = fwflex.StringToFrameworkARN(ctx, output.RoleArn)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceDRTAccessRoleARNAssociation) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new drtAccessRoleARNAssociationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}
	conn := r.Meta().ShieldClient(ctx)

	if !new.RoleARN.Equal(old.RoleARN) {
		roleARN := new.RoleARN.ValueString()
		input := &shield.AssociateDRTRoleInput{
			RoleArn: aws.String(roleARN),
		}

		_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParameterException](ctx, propagationTimeout, func() (interface{}, error) {
			return conn.AssociateDRTRole(ctx, input)
		}, "role does not have a valid DRT managed policy")

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating Shield DRT Role ARN Association (%s)", roleARN), err.Error())

			return
		}

		_, err = tfresource.RetryWhenNotFound(ctx, r.UpdateTimeout(ctx, new.Timeouts), func() (interface{}, error) {
			return findDRTRoleARNAssociation(ctx, conn, roleARN)
		})

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("waiting for Shield DRT Role ARN Association (%s) update", roleARN), err.Error())

			return
		}
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourceDRTAccessRoleARNAssociation) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data drtAccessRoleARNAssociationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ShieldClient(ctx)

	roleARN := data.ID.ValueString()
	input := &shield.DisassociateDRTRoleInput{}

	_, err := conn.DisassociateDRTRole(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Shield DRT Role ARN Association (%s)", roleARN), err.Error())

		return
	}

	_, err = tfresource.RetryUntilNotFound(ctx, r.DeleteTimeout(ctx, data.Timeouts), func() (interface{}, error) {
		return findDRTRoleARNAssociation(ctx, conn, roleARN)
	})

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("waiting for Shield DRT Role ARN Association (%s) delete", roleARN), err.Error())

		return
	}
}

func findDRTRoleARNAssociation(ctx context.Context, conn *shield.Client, arn string) (*string, error) {
	output, err := findDRTAccess(ctx, conn)

	if err != nil {
		return nil, err
	}

	if aws.ToString(output.RoleArn) != arn {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output.RoleArn, nil
}

type drtAccessRoleARNAssociationResourceModel struct {
	ID       types.String   `tfsdk:"id"`
	RoleARN  fwtypes.ARN    `tfsdk:"role_arn"`
	Timeouts timeouts.Value `tfsdk:"timeouts"`
}
