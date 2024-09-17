// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ResNameReservedCacheNode = "Reserved Cache Node"
)

// @SDKResource("aws_elasticache_reserved_cache_node")
// @Tags(identifierAttribute="arn")
// @Testing(tagsTests=false)
func ResourceReservedCacheNode() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReservedCacheNodeCreate,
		ReadWithoutTimeout:   resourceReservedCacheNodeRead,
		UpdateWithoutTimeout: resourceReservedCacheNodeUpdate,
		DeleteWithoutTimeout: resourceReservedCacheNodeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cache_node_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDuration: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"fixed_price": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"cache_node_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"offering_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"offering_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"product_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"recurring_charges": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"recurring_charge_amount": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"recurring_charge_frequency": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"reservation_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrStartTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"usage_price": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReservedCacheNodeCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	input := elasticache.PurchaseReservedCacheNodesOfferingInput{
		ReservedCacheNodesOfferingId: aws.String(d.Get("offering_id").(string)),
		Tags:                         getTagsIn(ctx),
	}

	if v, ok := d.Get("cache_node_count").(int); ok && v > 0 {
		input.CacheNodeCount = aws.Int32(int32(d.Get("cache_node_count").(int)))
	}

	if v, ok := d.Get("reservation_id").(string); ok && v != "" {
		input.ReservedCacheNodeId = aws.String(v)
	}

	resp, err := conn.PurchaseReservedCacheNodesOffering(ctx, &input)
	if err != nil {
		return create.AppendDiagError(diags, names.ElastiCache, create.ErrActionCreating, ResNameReservedCacheNode, fmt.Sprintf("offering_id: %s, reservation_id: %s", d.Get("offering_id").(string), d.Get("reservation_id").(string)), err)
	}

	d.SetId(aws.ToString(resp.ReservedCacheNode.ReservedCacheNodeId))

	if err := waitReservedCacheNodeCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return create.AppendDiagError(diags, names.ElastiCache, create.ErrActionWaitingForCreation, ResNameReservedCacheNode, d.Id(), err)
	}

	return append(diags, resourceReservedCacheNodeRead(ctx, d, meta)...)
}

func resourceReservedCacheNodeRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).ElastiCacheClient(ctx)

	reservation, err := findReservedCacheNodeByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		create.LogNotFoundRemoveState(names.ElastiCache, create.ErrActionReading, ResNameReservedCacheNode, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.AppendDiagError(diags, names.ElastiCache, create.ErrActionReading, ResNameReservedCacheNode, d.Id(), err)
	}

	d.Set(names.AttrARN, reservation.ReservationARN)
	d.Set("cache_node_type", reservation.CacheNodeType)
	d.Set(names.AttrDuration, reservation.Duration)
	d.Set("fixed_price", reservation.FixedPrice)
	d.Set("cache_node_count", reservation.CacheNodeCount)
	d.Set("offering_id", reservation.ReservedCacheNodesOfferingId)
	d.Set("offering_type", reservation.OfferingType)
	d.Set("product_description", reservation.ProductDescription)
	d.Set("recurring_charges", flattenRecurringCharges(reservation.RecurringCharges))
	d.Set("reservation_id", reservation.ReservedCacheNodeId)
	d.Set(names.AttrStartTime, (reservation.StartTime).Format(time.RFC3339))
	d.Set(names.AttrState, reservation.State)
	d.Set("usage_price", reservation.UsagePrice)

	return diags
}

func resourceReservedCacheNodeUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceReservedCacheNodeRead(ctx, d, meta)...)
}

func resourceReservedCacheNodeDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	log.Printf("[DEBUG] %s %s cannot be deleted. Removing from state.: %s", names.ElastiCache, ResNameReservedCacheNode, d.Id())

	return diags
}

func flattenRecurringCharges(recurringCharges []awstypes.RecurringCharge) []any {
	if len(recurringCharges) == 0 {
		return []any{}
	}

	var rawRecurringCharges []any
	for _, recurringCharge := range recurringCharges {
		rawRecurringCharge := map[string]any{
			"recurring_charge_amount":    recurringCharge.RecurringChargeAmount,
			"recurring_charge_frequency": aws.ToString(recurringCharge.RecurringChargeFrequency),
		}

		rawRecurringCharges = append(rawRecurringCharges, rawRecurringCharge)
	}

	return rawRecurringCharges
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
