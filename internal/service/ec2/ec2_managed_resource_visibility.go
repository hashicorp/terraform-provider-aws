// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_ec2_managed_resource_visibility", name="Managed Resource Visibility")
// @SingletonIdentity
// @NoImport
// @Testing(hasExistsFunction=false)
// @Testing(generator=false)
// @Testing(hasNoPreExistingResource=true)
// @Testing(checkDestroyNoop=true)
func newManagedResourceVisibilityResource(_ context.Context) (resource.ResourceWithConfigure, error) {
	return &managedResourceVisibilityResource{}, nil
}

type managedResourceVisibilityResource struct {
	framework.ResourceWithModel[managedResourceVisibilityResourceModel]
}

func (r *managedResourceVisibilityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"default_visibility": schema.StringAttribute{
				CustomType: fwtypes.StringEnumType[awstypes.ManagedResourceDefaultVisibility](),
				Required:   true,
			},
			names.AttrID: framework.IDAttribute(),
		},
	}
}

func (r *managedResourceVisibilityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data managedResourceVisibilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.ModifyManagedResourceVisibilityInput{
		DefaultVisibility: data.DefaultVisibility.ValueEnum(),
	}

	_, err := conn.ModifyManagedResourceVisibility(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("creating EC2 Managed Resource Visibility", err.Error())
		return
	}

	data.ID = types.StringValue(r.Meta().Region(ctx))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedResourceVisibilityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data managedResourceVisibilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.GetManagedResourceVisibilityInput{}
	output, err := conn.GetManagedResourceVisibility(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("reading EC2 Managed Resource Visibility (%s)", data.ID.ValueString()), err.Error())
		return
	}

	data.DefaultVisibility = fwtypes.StringEnumValue(output.Visibility.DefaultVisibility)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedResourceVisibilityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data managedResourceVisibilityResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.ModifyManagedResourceVisibilityInput{
		DefaultVisibility: data.DefaultVisibility.ValueEnum(),
	}

	_, err := conn.ModifyManagedResourceVisibility(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("updating EC2 Managed Resource Visibility (%s)", data.ID.ValueString()), err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *managedResourceVisibilityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data managedResourceVisibilityResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().EC2Client(ctx)

	input := ec2.ModifyManagedResourceVisibilityInput{
		DefaultVisibility: awstypes.ManagedResourceDefaultVisibilityVisible,
	}

	_, err := conn.ModifyManagedResourceVisibility(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("deleting EC2 Managed Resource Visibility (%s)", data.ID.ValueString()), err.Error())
		return
	}
}

type managedResourceVisibilityResourceModel struct {
	framework.WithRegionModel
	DefaultVisibility fwtypes.StringEnum[awstypes.ManagedResourceDefaultVisibility] `tfsdk:"default_visibility"`
	ID                types.String                                                  `tfsdk:"id"`
}
