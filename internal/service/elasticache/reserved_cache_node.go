// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @FrameworkResource("aws_elasticache_reserved_cache_node", name="Reserved Cache Node")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTests=false)
func newResourceReservedCacheNode(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourceReservedCacheNode{}
	r.SetDefaultCreateTimeout(30 * time.Minute)
	r.SetDefaultUpdateTimeout(10 * time.Minute)
	r.SetDefaultDeleteTimeout(1 * time.Minute)

	return r, nil
}

type resourceReservedCacheNode struct {
	framework.ResourceWithConfigure
	framework.WithNoOpUpdate[resourceReservedCacheNodeModel]
	framework.WithNoOpDelete
	framework.WithTimeouts
}

func (r *resourceReservedCacheNode) Schema(ctx context.Context, request resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			names.AttrARN: schema.StringAttribute{
				Computed: true,
			},
			"cache_node_count": schema.Int32Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"cache_node_type": schema.StringAttribute{
				Computed: true,
			},
			names.AttrDuration: schema.StringAttribute{
				CustomType: fwtypes.RFC3339DurationType,
				Computed:   true,
			},
			"fixed_price": schema.Float64Attribute{
				Computed: true,
			},
			names.AttrID: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"reserved_cache_nodes_offering_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"offering_type": schema.StringAttribute{
				Computed: true,
			},
			"product_description": schema.StringAttribute{
				Computed: true,
			},
			"recurring_charges": schema.ListAttribute{
				CustomType: fwtypes.NewListNestedObjectTypeOf[recurringChargeModel](ctx),
				Computed:   true,
				ElementType: types.ObjectType{
					AttrTypes: fwtypes.AttributeTypesMust[recurringChargeModel](ctx),
				},
			},
			names.AttrStartTime: schema.StringAttribute{
				CustomType: timetypes.RFC3339Type{},
				Computed:   true,
			},
			names.AttrState: schema.StringAttribute{
				Computed: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
			"usage_price": schema.Float64Attribute{
				Computed: true,
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

// Create is called when the provider must create a new resource.
// Config and planned state values should be read from the CreateRequest and new state values set on the CreateResponse.
func (r *resourceReservedCacheNode) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	var data resourceReservedCacheNodeModel

	response.Diagnostics.Append(request.Plan.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ElastiCacheClient(ctx)

	var input elasticache.PurchaseReservedCacheNodesOfferingInput
	response.Diagnostics.Append(flex.Expand(ctx, data, &input, r.flexOpts()...)...)

	input.Tags = getTagsIn(ctx)

	resp, err := conn.PurchaseReservedCacheNodesOffering(ctx, &input)
	if err != nil {
		response.Diagnostics.AddError(
			"Creating ElastiCache Reserved Cache Node",
			fmt.Sprintf("Could not create ElastiCache Reserved Cache Node with Offering ID %q\nError: %s", data.ReservedCacheNodesOfferingID.ValueString(), err.Error()),
		)
		return
	}

	createTimeout := r.CreateTimeout(ctx, data.Timeouts)
	if err := waitReservedCacheNodeCreated(ctx, conn, aws.ToString(resp.ReservedCacheNode.ReservedCacheNodeId), createTimeout); err != nil {
		response.Diagnostics.AddError(
			"Creating ElastiCache Reserved Cache Node",
			fmt.Sprintf("Creating ElastiCache Reserved Cache Node with Offering ID %q failed while waiting for completion.\nError: %s", data.ReservedCacheNodesOfferingID.ValueString(), err.Error()),
		)
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, resp.ReservedCacheNode, &data, r.flexOpts()...)...)
	if response.Diagnostics.HasError() {
		return
	}

	duration := time.Duration(aws.ToInt32(resp.ReservedCacheNode.Duration)) * time.Second
	data.Duration = fwtypes.RFC3339DurationTimeDurationValue(duration)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceReservedCacheNode) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	var data resourceReservedCacheNodeModel

	response.Diagnostics.Append(request.State.Get(ctx, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ElastiCacheClient(ctx)

	reservation, err := findReservedCacheNodeByID(ctx, conn, data.ID.ValueString())

	if tfresource.NotFound(err) {
		response.Diagnostics.Append(fwdiag.NewResourceNotFoundWarningDiagnostic(err))
		response.State.RemoveResource(ctx)
		return
	}

	if err != nil {
		response.Diagnostics.AddError(fmt.Sprintf("reading ElastiCache Reserved Cache Node (%s)", data.ID.ValueString()), err.Error())
		return
	}

	response.Diagnostics.Append(flex.Flatten(ctx, reservation, &data)...)
	if response.Diagnostics.HasError() {
		return
	}

	duration := time.Duration(aws.ToInt32(reservation.Duration)) * time.Second
	data.Duration = fwtypes.RFC3339DurationTimeDurationValue(duration)

	response.Diagnostics.Append(response.State.Set(ctx, &data)...)
}

func (r *resourceReservedCacheNode) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root(names.AttrID), request, response)
}

func (r *resourceReservedCacheNode) flexOpts() []flex.AutoFlexOptionsFunc {
	return []flex.AutoFlexOptionsFunc{
		flex.WithFieldNamePrefix("ReservedCacheNode"),
	}
}

type resourceReservedCacheNodeModel struct {
	ReservationARN               types.String                                          `tfsdk:"arn"`
	CacheNodeCount               types.Int32                                           `tfsdk:"cache_node_count"`
	CacheNodeType                types.String                                          `tfsdk:"cache_node_type"`
	Duration                     fwtypes.RFC3339Duration                               `tfsdk:"duration" autoflex:",noflatten"`
	FixedPrice                   types.Float64                                         `tfsdk:"fixed_price"`
	ID                           types.String                                          `tfsdk:"id"`
	ReservedCacheNodesOfferingID types.String                                          `tfsdk:"reserved_cache_nodes_offering_id"`
	OfferingType                 types.String                                          `tfsdk:"offering_type"`
	ProductDescription           types.String                                          `tfsdk:"product_description"`
	RecurringCharges             fwtypes.ListNestedObjectValueOf[recurringChargeModel] `tfsdk:"recurring_charges"`
	StartTime                    timetypes.RFC3339                                     `tfsdk:"start_time"`
	State                        types.String                                          `tfsdk:"state"`
	Tags                         tftags.Map                                            `tfsdk:"tags"`
	TagsAll                      tftags.Map                                            `tfsdk:"tags_all"`
	UsagePrice                   types.Float64                                         `tfsdk:"usage_price"`

	Timeouts timeouts.Value `tfsdk:"timeouts"`
}

type recurringChargeModel struct {
	RecurringChargeAmount    types.Float64 `tfsdk:"recurring_charge_amount"`
	RecurringChargeFrequency types.String  `tfsdk:"recurring_charge_frequency"`
}

func findReservedCacheNodeByID(ctx context.Context, conn *elasticache.Client, id string) (result awstypes.ReservedCacheNode, err error) {
	input := elasticache.DescribeReservedCacheNodesInput{
		ReservedCacheNodeId: aws.String(id),
	}

	output, err := conn.DescribeReservedCacheNodes(ctx, &input)

	if errs.IsA[*awstypes.ReservedCacheNodeNotFoundFault](err) {
		return result, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return result, err
	}

	if output == nil || len(output.ReservedCacheNodes) == 0 {
		return result, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.ReservedCacheNodes); count > 1 {
		return result, tfresource.NewTooManyResultsError(count, input)
	}

	return output.ReservedCacheNodes[0], nil
}

func waitReservedCacheNodeCreated(ctx context.Context, conn *elasticache.Client, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			reservedCacheNodeStatePaymentPending,
		},
		Target:         []string{reservedCacheNodeStateActive},
		Refresh:        statusReservedCacheNode(ctx, conn, id),
		NotFoundChecks: 5,
		Timeout:        timeout,
		MinTimeout:     10 * time.Second,
		Delay:          30 * time.Second,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func statusReservedCacheNode(ctx context.Context, conn *elasticache.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findReservedCacheNodeByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.State), nil
	}
}
