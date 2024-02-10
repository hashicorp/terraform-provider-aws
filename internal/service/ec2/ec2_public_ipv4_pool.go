// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource(name="Public IPv4 Pool")
// @Tags(identifierAttribute="id")
func newResourcePublicIPv4Pool(_ context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePublicIPv4Pool{}
	return r, nil
}

const (
	ResNamePublicIPv4Pool = "Public IPv4 Pool"
)

type resourcePublicIPv4Pool struct {
	framework.ResourceWithConfigure
}

func (r *resourcePublicIPv4Pool) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "aws_ec2_public_ipv4_pool"
}

func (r *resourcePublicIPv4Pool) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": framework.IDAttribute(),
			"description": schema.StringAttribute{
				Computed: true,
			},
			"network_border_group": schema.StringAttribute{
				Computed: true,
			},
			// FIXME TODO
			//			"pool_address_ranges": {
			//				Type:     schema.TypeList,
			//				Computed: true,
			//				Elem: &schema.Resource{
			//					Schema: map[string]*schema.Schema{
			//						"address_count": {
			//							Type:     schema.TypeInt,
			//							Computed: true,
			//						},
			//						"available_address_count": {
			//							Type:     schema.TypeInt,
			//							Computed: true,
			//						},
			//						"first_address": {
			//							Type:     schema.TypeString,
			//							Computed: true,
			//						},
			//						"last_address": {
			//							Type:     schema.TypeString,
			//							Computed: true,
			//						},
			//					},
			//				},
			//			},
			//			"tags": tftags.TagsSchemaComputed(),
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"total_address_count": schema.Int64Attribute{
				Computed: true,
			},
			"total_available_address_count": schema.Int64Attribute{
				Computed: true,
			},
		},
	}
}

func (r *resourcePublicIPv4Pool) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	conn := r.Meta().EC2Client(ctx)

	var plan resourcePublicIPv4PoolData
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ec2.CreatePublicIpv4PoolInput{
		TagSpecifications: getTagSpecificationsInV2(ctx, awstypes.ResourceTypeIpv4poolEc2),
	}

	out, err := conn.CreatePublicIpv4Pool(ctx, in)
	if err != nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNamePublicIPv4Pool, "", err),
			err.Error(),
		)
		return
	}
	if out == nil {
		resp.Diagnostics.AddError(
			create.ProblemStandardMessage(names.EC2, create.ErrActionCreating, ResNamePublicIPv4Pool, "", nil),
			errors.New("empty output").Error(),
		)
		return
	}

	plan.ID = flex.StringToFramework(ctx, out.PoolId)

	// those are all empty when creating the pool, but not returned from the create call
	plan.Description = types.StringValue("")
	plan.NetworkBorderGroup = types.StringValue(conn.Options().Region)
	plan.TotalAddressCount = types.Int64Value(0)
	plan.TotalAvailableAddressCount = types.Int64Value(0)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePublicIPv4Pool) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	conn := r.Meta().EC2Client(ctx)

	var state resourcePublicIPv4PoolData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findPublicIPv4PoolByID(ctx, conn, state.ID.ValueString())

	if err != nil {
		// FIXME should only remove from state if it's really not found, not on arbitrary error
		// if tfawserr.ErrCodeEquals(err, errCodeInvalidPublicIpv4PoolIDNotFound) {
		// if tfresource.NotFound(err) {
		//
		//   tflog.Warn(ctx, "Other error related to Public IPv4 Pool", map[string]interface{}{
		//			"id": state.ID.ValueString(),
		// })
		// resp.Diagnostics.AddError(
		//    create.ProblemStandardMessage(names.EC2, create.ErrActionSetting, ResNamePublicIPv4Pool, state.ID.String(), err),
		//    err.Error(),
		// )
		tflog.Warn(ctx, "Public IPv4 Pool not found, removing from state", map[string]interface{}{
			"id": state.ID.ValueString(),
		})
		resp.State.RemoveResource(ctx)
		return
	}

	state.ID = flex.StringToFramework(ctx, out.PoolId)

	state.Description = flex.StringToFramework(ctx, out.Description)
	state.NetworkBorderGroup = flex.StringToFramework(ctx, out.NetworkBorderGroup)
	state.TotalAddressCount = flex.Int32ValueToFramework(ctx, *out.TotalAddressCount)
	state.TotalAvailableAddressCount = flex.Int32ValueToFramework(ctx, *out.TotalAvailableAddressCount)

	setTagsOutV2(ctx, out.Tags)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *resourcePublicIPv4Pool) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Tags only.
}

func (r *resourcePublicIPv4Pool) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Warn(ctx, "DELETE start")
	conn := r.Meta().EC2Client(ctx)

	var state resourcePublicIPv4PoolData
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	in := &ec2.DeletePublicIpv4PoolInput{
		PoolId: aws.String(state.ID.ValueString()),
	}

	// FIXME debug
	tflog.Warn(ctx, "DELETE doing delete")
	_, err := conn.DeletePublicIpv4Pool(ctx, in)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidPublicIpv4PoolIDNotFound) {
		// FIXME debug
		tflog.Warn(ctx, "DELETE error not found, all good")
		return
	}
	// TODO check for pool not found
	if err != nil {
		tflog.Warn(ctx, "DELETE in error return")
		return

		// resp.Diagnostics.AddError(
		// 	create.ProblemStandardMessage(names.EC2, create.ErrActionDeleting, ResNamePublicIPv4Pool, state.ID.String(), err),
		// 	err.Error(),
		// )
		// return
	}
	return
}

func (r *resourcePublicIPv4Pool) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourcePublicIPv4Pool) ModifyPlan(ctx context.Context, request resource.ModifyPlanRequest, response *resource.ModifyPlanResponse) {
	r.SetTagsAll(ctx, request, response)
}

func findPublicIPv4PoolByID(ctx context.Context, conn *ec2.Client, id string) (*awstypes.PublicIpv4Pool, error) {
	in := &ec2.DescribePublicIpv4PoolsInput{
		PoolIds: []string{id},
	}

	out, err := conn.DescribePublicIpv4Pools(ctx, in)
	// FIXME just for testing --------------------
	if err != nil {
		tflog.Warn(ctx, "Public IPv4 Pool not found during findPublicIPv4PoolByID", map[string]interface{}{
			"id": id,
		})
		return nil, err
	}

	//	if err != nil {
	//		var oe *smithy.OperationError
	//		if errors.As(err, &oe) {
	//			var responseError *http.ResponseError
	//			if errors.As(oe.Unwrap(), &responseError) {
	//				log.Printf("failed to call service: %s, operation: %s, error: %v", oe.Service(), oe.Operation(), oe.Unwrap())
	//			}
	//		}

	// it's not an awserr
	// if tfawserr.ErrCodeEquals(err, errCodeInvalidPublicIpv4PoolIDNotFound) {
	//	tflog.Warn(ctx, "Public IPv4 Pool not found", map[string]interface{}{
	//		"id": id,
	//	})
	//	return nil, &retry.NotFoundError{
	//		LastError:   err,
	//		LastRequest: in,
	//	}
	//}

	//		return nil, err
	//	}

	if out == nil || out.PublicIpv4Pools == nil || len(out.PublicIpv4Pools) == 0 {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return &out.PublicIpv4Pools[0], nil
}

type resourcePublicIPv4PoolData struct {
	ID                         types.String `tfsdk:"id"`
	Description                types.String `tfsdk:"description"`
	NetworkBorderGroup         types.String `tfsdk:"network_border_group"`
	Tags                       types.Map    `tfsdk:"tags"`
	TagsAll                    types.Map    `tfsdk:"tags_all"`
	TotalAddressCount          types.Int64  `tfsdk:"total_address_count"`
	TotalAvailableAddressCount types.Int64  `tfsdk:"total_available_address_count"`
}
