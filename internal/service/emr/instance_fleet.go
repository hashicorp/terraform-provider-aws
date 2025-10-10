// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/emr"
	awstypes "github.com/aws/aws-sdk-go-v2/service/emr/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_emr_instance_fleet", name="Instance Fleet")
func resourceInstanceFleet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceFleetCreate,
		ReadWithoutTimeout:   resourceInstanceFleetRead,
		UpdateWithoutTimeout: resourceInstanceFleetUpdate,
		DeleteWithoutTimeout: resourceInstanceFleetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(_ context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected cluster-id/fleet-id", d.Id())
				}
				clusterID := idParts[0]
				resourceID := idParts[1]
				d.Set("cluster_id", clusterID)
				d.SetId(resourceID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_type_configs": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bid_price": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"bid_price_as_percentage_of_on_demand_price": {
							Type:     schema.TypeFloat,
							Optional: true,
							ForceNew: true,
							Default:  100,
						},
						"configurations": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"classification": {
										Type:     schema.TypeString,
										Optional: true,
										ForceNew: true,
									},
									names.AttrProperties: {
										Type:     schema.TypeMap,
										Optional: true,
										ForceNew: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
								},
							},
						},
						"ebs_config": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrIOPS: {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
									names.AttrSize: {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									names.AttrType: {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validEBSVolumeType(),
									},
									"volumes_per_instance": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										Default:  1,
									},
								},
							},
							Set: resourceClusterEBSHashConfig,
						},
						names.AttrInstanceType: {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"weighted_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							Default:  1,
						},
					},
				},
				Set: resourceInstanceTypeHashConfig,
			},
			"launch_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_demand_specification": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allocation_strategy": {
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.OnDemandProvisioningAllocationStrategy](),
									},
								},
							},
						},
						"spot_specification": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allocation_strategy": {
										Type:             schema.TypeString,
										ForceNew:         true,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.SpotProvisioningAllocationStrategy](),
									},
									"block_duration_minutes": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										Default:  0,
									},
									"timeout_action": {
										Type:             schema.TypeString,
										Required:         true,
										ForceNew:         true,
										ValidateDiagFunc: enum.Validate[awstypes.SpotProvisioningTimeoutAction](),
									},
									"timeout_duration_minutes": {
										Type:     schema.TypeInt,
										ForceNew: true,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"provisioned_on_demand_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"provisioned_spot_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"target_on_demand_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"target_spot_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
		},
	}
}

func resourceInstanceFleetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	tfMap := map[string]any{
		names.AttrName:              d.Get(names.AttrName),
		"target_on_demand_capacity": d.Get("target_on_demand_capacity"),
		"target_spot_capacity":      d.Get("target_spot_capacity"),
		"instance_type_configs":     d.Get("instance_type_configs"),
		"launch_specifications":     d.Get("launch_specifications"),
	}
	input := &emr.AddInstanceFleetInput{
		ClusterId:     aws.String(d.Get("cluster_id").(string)),
		InstanceFleet: expandInstanceFleetConfig(tfMap, awstypes.InstanceFleetTypeTask),
	}

	output, err := conn.AddInstanceFleet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EMR Instance Fleet: %s", err)
	}

	d.SetId(aws.ToString(output.InstanceFleetId))

	return append(diags, resourceInstanceFleetRead(ctx, d, meta)...)
}

func resourceInstanceFleetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	fleet, err := findInstanceFleetByTwoPartKey(ctx, conn, d.Get("cluster_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Instance Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EMR Instance Fleet (%s): %s", d.Id(), err)
	}

	if err := d.Set("instance_type_configs", flattenInstanceTypeSpecifications(fleet.InstanceTypeSpecifications)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_type_configs: %s", err)
	}
	if err := d.Set("launch_specifications", flattenInstanceFleetProvisioningSpecifications(fleet.LaunchSpecifications)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting launch_specifications: %s", err)
	}
	d.Set(names.AttrName, fleet.Name)
	d.Set("provisioned_on_demand_capacity", fleet.ProvisionedOnDemandCapacity)
	d.Set("provisioned_spot_capacity", fleet.ProvisionedSpotCapacity)
	d.Set("target_on_demand_capacity", fleet.TargetOnDemandCapacity)
	d.Set("target_spot_capacity", fleet.TargetSpotCapacity)

	return diags
}

func resourceInstanceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	clusterID := d.Get("cluster_id").(string)
	input := &emr.ModifyInstanceFleetInput{
		ClusterId: aws.String(clusterID),
		InstanceFleet: &awstypes.InstanceFleetModifyConfig{
			InstanceFleetId:        aws.String(d.Id()),
			TargetOnDemandCapacity: aws.Int32(int32(d.Get("target_on_demand_capacity").(int))),
			TargetSpotCapacity:     aws.Int32(int32(d.Get("target_spot_capacity").(int))),
		},
	}

	_, err := conn.ModifyInstanceFleet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EMR Instance Fleet (%s): %s", d.Id(), err)
	}

	const (
		timeout = 75 * time.Minute
	)
	if _, err := waitInstanceFleetRunning(ctx, conn, clusterID, d.Id(), timeout); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EMR Instance Fleet (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceInstanceFleetRead(ctx, d, meta)...)
}

func resourceInstanceFleetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRClient(ctx)

	// AWS EMR Instance Fleet does not support DELETE; resizing cluster to zero before removing from state.
	log.Printf("[DEBUG] Deleting EMR Instance Fleet: %s", d.Id())
	_, err := conn.ModifyInstanceFleet(ctx, &emr.ModifyInstanceFleetInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
		InstanceFleet: &awstypes.InstanceFleetModifyConfig{
			InstanceFleetId:        aws.String(d.Id()),
			TargetOnDemandCapacity: aws.Int32(0),
			TargetSpotCapacity:     aws.Int32(0),
		},
	})

	// Ignore certain errors that indicate the fleet is already (being) deleted
	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "instance fleet may only be modified when the cluster is running or waiting") {
		return diags
	}
	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "A job flow that is shutting down, terminated, or finished may not be modified") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EMR Instance Fleet (%s): %s", d.Id(), err)
	}

	return diags
}

func findInstanceFleetByTwoPartKey(ctx context.Context, conn *emr.Client, clusterID, fleetID string) (*awstypes.InstanceFleet, error) {
	input := &emr.ListInstanceFleetsInput{
		ClusterId: aws.String(clusterID),
	}

	return findInstanceFleet(ctx, conn, input, func(v *awstypes.InstanceFleet) bool {
		return aws.ToString(v.Id) == fleetID
	})
}

func findInstanceFleet(ctx context.Context, conn *emr.Client, input *emr.ListInstanceFleetsInput, filter tfslices.Predicate[*awstypes.InstanceFleet]) (*awstypes.InstanceFleet, error) {
	output, err := findInstanceFleets(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findInstanceFleets(ctx context.Context, conn *emr.Client, input *emr.ListInstanceFleetsInput, filter tfslices.Predicate[*awstypes.InstanceFleet]) ([]awstypes.InstanceFleet, error) {
	var output []awstypes.InstanceFleet

	pages := emr.NewListInstanceFleetsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "is not valid") {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.InstanceFleets {
			if v.Status != nil && filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func statusInstanceFleet(ctx context.Context, conn *emr.Client, clusterID, fleetID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findInstanceFleetByTwoPartKey(ctx, conn, clusterID, fleetID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.State), nil
	}
}

func waitInstanceFleetRunning(ctx context.Context, conn *emr.Client, clusterID, fleetID string, timeout time.Duration) (*awstypes.InstanceFleet, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.InstanceFleetStateProvisioning, awstypes.InstanceFleetStateBootstrapping, awstypes.InstanceFleetStateResizing),
		Target:     enum.Slice(awstypes.InstanceFleetStateRunning),
		Refresh:    statusInstanceFleet(ctx, conn, clusterID, fleetID),
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.InstanceFleet); ok {
		return output, err
	}

	return nil, err
}
