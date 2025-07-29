// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	awstypes "github.com/aws/aws-sdk-go-v2/service/quicksight/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_quicksight_key_registration", name="Key Registration")
func newKeyRegistrationResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &keyRegistrationResource{}

	return r, nil
}

type keyRegistrationResource struct {
	framework.ResourceWithModel[keyRegistrationResourceModel]
}

func (r *keyRegistrationResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrAWSAccountID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					fwvalidators.AWSAccountID(),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"key_registration": schema.SetNestedBlock{
				CustomType: fwtypes.NewSetNestedObjectTypeOf[registeredCustomerManagedKeyModel](ctx),
				Validators: []validator.Set{
					setvalidator.IsRequired(),
					setvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"default_key": schema.BoolAttribute{
							Optional: true,
							Computed: true,
							Default:  booldefault.StaticBool(false),
						},
						"key_arn": schema.StringAttribute{
							CustomType: fwtypes.ARNType,
							Required:   true,
						},
					},
				},
			},
		},
	}
}

func (r *keyRegistrationResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data keyRegistrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}
	if data.AWSAccountID.IsUnknown() {
		data.AWSAccountID = fwflex.StringValueToFramework(ctx, r.Meta().AccountID(ctx))
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AWSAccountID)
	var input quicksight.UpdateKeyRegistrationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateKeyRegistration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight Key Registration (%s)", accountID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *keyRegistrationResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data keyRegistrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AWSAccountID)
	output, err := findKeyRegistrationByID(ctx, conn, accountID)

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Quicksight Key Registration (%s)", accountID), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data.KeyRegistration)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *keyRegistrationResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old keyRegistrationResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, new.AWSAccountID)
	var input quicksight.UpdateKeyRegistrationInput
	response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	_, err := conn.UpdateKeyRegistration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Quicksight Key Registration (%s)", accountID), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *keyRegistrationResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data keyRegistrationResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().QuickSightClient(ctx)

	accountID := fwflex.StringValueFromFramework(ctx, data.AWSAccountID)
	input := quicksight.UpdateKeyRegistrationInput{
		AwsAccountId:    aws.String(accountID),
		KeyRegistration: []awstypes.RegisteredCustomerManagedKey{},
	}
	_, err := conn.UpdateKeyRegistration(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Quicksight Key Registration (%s)", accountID), err.Error())

		return
	}
}

func (r *keyRegistrationResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrAWSAccountID), request, response)
}

func findKeyRegistrationByID(ctx context.Context, conn *quicksight.Client, id string) ([]awstypes.RegisteredCustomerManagedKey, error) {
	input := quicksight.DescribeKeyRegistrationInput{
		AwsAccountId: aws.String(id),
	}
	output, err := conn.DescribeKeyRegistration(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.KeyRegistration) == 0 {
		return nil, tfresource.NewEmptyResultError(&input)
	}

	return output.KeyRegistration, nil
}

type keyRegistrationResourceModel struct {
	framework.WithRegionModel
	AWSAccountID    types.String                                                      `tfsdk:"aws_account_id"`
	KeyRegistration fwtypes.SetNestedObjectValueOf[registeredCustomerManagedKeyModel] `tfsdk:"key_registration"`
}

type registeredCustomerManagedKeyModel struct {
	DefaultKey types.Bool  `tfsdk:"default_key"`
	KeyARN     fwtypes.ARN `tfsdk:"key_arn"`
}
