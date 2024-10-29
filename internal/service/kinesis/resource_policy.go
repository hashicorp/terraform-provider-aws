// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesis

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesis/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Resource Policy")
func newResourcePolicyResource(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePolicyResource{}

	return r, nil
}

type resourcePolicyResource struct {
	framework.ResourceWithConfigure
	framework.WithImportByID
}

func (r *resourcePolicyResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "aws_kinesis_resource_policy"
}

func (r *resourcePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrID: framework.IDAttribute(),
			names.AttrPolicy: schema.StringAttribute{
				CustomType: fwtypes.IAMPolicyType,
				Required:   true,
			},
			names.AttrResourceARN: schema.StringAttribute{
				CustomType: fwtypes.ARNType,
				Required:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *resourcePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourcePolicyResourceModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().KinesisClient(ctx)

	input := &kinesis.PutResourcePolicyInput{
		Policy:      flex.StringFromFramework(ctx, data.Policy),
		ResourceARN: flex.StringFromFramework(ctx, data.ResourceARN),
	}

	_, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating Kinesis Resource Policy (%s)", data.ResourceARN.ValueString()), err.Error())

		return
	}

	// Set values for unknowns.
	data.setID()

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourcePolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)

	if response.Diagnostics.HasError() {
		return
	}

	if err := data.InitFromID(); err != nil {
		response.Diagnostics.AddError("parsing resource ID", err.Error())

		return
	}

	conn := r.Meta().KinesisClient(ctx)

	output, err := findResourcePolicyByARN(ctx, conn, data.ResourceARN.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading Kinesis Resource Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(flex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourcePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var old, new resourcePolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().KinesisClient(ctx)

	input := &kinesis.PutResourcePolicyInput{
		Policy:      flex.StringFromFramework(ctx, new.Policy),
		ResourceARN: flex.StringFromFramework(ctx, new.ResourceARN),
	}

	_, err := conn.PutResourcePolicy(ctx, input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("updating Kinesis Resource Policy (%s)", new.ID.ValueString()), err.Error())

		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *resourcePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data resourcePolicyResourceModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().KinesisClient(ctx)

	_, err := conn.DeleteResourcePolicy(ctx, &kinesis.DeleteResourcePolicyInput{
		ResourceARN: flex.StringFromFramework(ctx, data.ResourceARN),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting Kinesis Resource Policy (%s)", data.ID.ValueString()), err.Error())

		return
	}
}

func findResourcePolicyByARN(ctx context.Context, conn *kinesis.Client, resourceARN string) (*kinesis.GetResourcePolicyOutput, error) {
	input := &kinesis.GetResourcePolicyInput{
		ResourceARN: aws.String(resourceARN),
	}

	output, err := conn.GetResourcePolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

type resourcePolicyResourceModel struct {
	ID          types.String      `tfsdk:"id"`
	Policy      fwtypes.IAMPolicy `tfsdk:"policy"`
	ResourceARN fwtypes.ARN       `tfsdk:"resource_arn"`
}

func (data *resourcePolicyResourceModel) InitFromID() error {
	_, err := arn.Parse(data.ID.ValueString())
	if err != nil {
		return err
	}

	data.ResourceARN = fwtypes.ARNValue(data.ID.ValueString())

	return nil
}

func (data *resourcePolicyResourceModel) setID() {
	data.ID = data.ResourceARN.StringValue
}
