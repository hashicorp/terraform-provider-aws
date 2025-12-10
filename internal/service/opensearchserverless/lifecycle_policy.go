// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package opensearchserverless

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/opensearchserverless"
	awstypes "github.com/aws/aws-sdk-go-v2/service/opensearchserverless/types"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_opensearchserverless_lifecycle_policy", name="Lifecycle Policy")
func newLifecyclePolicyResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &lifecyclePolicyResource{}, nil
}

type lifecyclePolicyResource struct {
	framework.ResourceWithModel[lifecyclePolicyResourceModel]
}

func (r *lifecyclePolicyResource) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrDescription: schema.StringAttribute{
				Description: "Description of the policy.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 1000),
				},
			},
			names.AttrID: framework.IDAttribute(),
			names.AttrName: schema.StringAttribute{
				Description: "Name of the policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(3, 32),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			names.AttrPolicy: schema.StringAttribute{
				Description: "JSON policy document to use as the content for the new policy.",
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 20480),
				},
			},
			"policy_version": schema.StringAttribute{
				Description: "Version of the policy.",
				Computed:    true,
			},
			names.AttrType: schema.StringAttribute{
				Description: "Type of lifecycle policy. Must be `retention`.",
				CustomType:  fwtypes.StringEnumType[awstypes.LifecyclePolicyType](),
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *lifecyclePolicyResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data lifecyclePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.Name)
	var input opensearchserverless.CreateLifecyclePolicyInput
	response.Diagnostics.Append(fwflex.Expand(ctx, data, &input)...)
	if response.Diagnostics.HasError() {
		return
	}

	// Additional fields.
	input.ClientToken = aws.String(sdkid.UniqueId())

	output, err := conn.CreateLifecyclePolicy(ctx, &input)

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("creating OpenSearch Serverless Lifecycle Policy (%s)", name), err.Error())

		return
	}

	// Set values for unknowns.
	data.ID = fwflex.StringValueToFramework(ctx, name)
	data.PolicyVersion = fwflex.StringToFramework(ctx, output.LifecyclePolicyDetail.PolicyVersion)

	response.Diagnostics.Append(response.State.Set(ctx, data)...)
}

func (r *lifecyclePolicyResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data lifecyclePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.ID)
	output, err := findLifecyclePolicyByNameAndType(ctx, conn, name, data.Type.ValueString())

	if retry.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)

		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading OpenSearch Serverless Lifecycle Policy (%s)", name), err.Error())

		return
	}

	// Set attributes for import.
	response.Diagnostics.Append(fwflex.Flatten(ctx, output, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *lifecyclePolicyResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	var new, old lifecyclePolicyResourceModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &new)...)
	if response.Diagnostics.HasError() {
		return
	}
	response.Diagnostics.Append(request.State.Get(ctx, &old)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	if !new.Description.Equal(old.Description) || !new.Policy.Equal(old.Policy) {
		name := fwflex.StringValueFromFramework(ctx, new.ID)
		var input opensearchserverless.UpdateLifecyclePolicyInput
		response.Diagnostics.Append(fwflex.Expand(ctx, new, &input)...)
		if response.Diagnostics.HasError() {
			return
		}

		// Additional fields.
		input.ClientToken = aws.String(sdkid.UniqueId())
		input.PolicyVersion = old.PolicyVersion.ValueStringPointer() // use policy version from state since it can be recalculated on update

		output, err := conn.UpdateLifecyclePolicy(ctx, &input)

		if err != nil {
			response.Diagnostics.AddError(fmt.Sprintf("updating OpenSearch Serverless Lifecycle Policy (%s)", name), err.Error())

			return
		}

		// Set values for unknowns.
		new.PolicyVersion = fwflex.StringToFramework(ctx, output.LifecyclePolicyDetail.PolicyVersion)
	}

	response.Diagnostics.Append(response.State.Set(ctx, &new)...)
}

func (r *lifecyclePolicyResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	var data lifecyclePolicyResourceModel
	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().OpenSearchServerlessClient(ctx)

	name := fwflex.StringValueFromFramework(ctx, data.ID)
	input := opensearchserverless.DeleteLifecyclePolicyInput{
		ClientToken: aws.String(sdkid.UniqueId()),
		Name:        aws.String(name),
		Type:        data.Type.ValueEnum(),
	}
	_, err := conn.DeleteLifecyclePolicy(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("deleting OpenSearch Serverless Lifecycle Policy (%s)", name), err.Error())

		return
	}
}

func (r *lifecyclePolicyResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	parts := strings.Split(request.ID, resourceIDSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		err := fmt.Errorf("unexpected format for ID (%[1]s), expected lifecycle-policy-name%[2]slifecycle-policy-type", request.ID, resourceIDSeparator)
		response.Diagnostics.Append(fwdiag.NewParsingResourceIDErrorDiagnostic(err))

		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrID), parts[0])...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root(names.AttrType), parts[1])...)
}

type lifecyclePolicyResourceModel struct {
	framework.WithRegionModel
	Description   types.String                                     `tfsdk:"description"`
	ID            types.String                                     `tfsdk:"id"`
	Name          types.String                                     `tfsdk:"name"`
	Policy        jsontypes.Normalized                             `tfsdk:"policy"`
	PolicyVersion types.String                                     `tfsdk:"policy_version"`
	Type          fwtypes.StringEnum[awstypes.LifecyclePolicyType] `tfsdk:"type"`
}
