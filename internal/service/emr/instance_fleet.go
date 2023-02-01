package emr

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func ResourceInstanceFleet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInstanceFleetCreate,
		ReadWithoutTimeout:   resourceInstanceFleetRead,
		UpdateWithoutTimeout: resourceInstanceFleetUpdate,
		DeleteWithoutTimeout: resourceInstanceFleetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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
									"properties": {
										Type:     schema.TypeMap,
										Optional: true,
										ForceNew: true,
										Elem:     schema.TypeString,
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
									"iops": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
									},
									"size": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									"type": {
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
						"instance_type": {
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
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(emr.OnDemandProvisioningAllocationStrategy_Values(), false),
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
										Type:         schema.TypeString,
										ForceNew:     true,
										Required:     true,
										ValidateFunc: validation.StringInSlice(emr.SpotProvisioningAllocationStrategy_Values(), false),
									},
									"block_duration_minutes": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										Default:  0,
									},
									"timeout_action": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice(emr.SpotProvisioningTimeoutAction_Values(), false),
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
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
			"provisioned_on_demand_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"provisioned_spot_capacity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceInstanceFleetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()

	addInstanceFleetInput := &emr.AddInstanceFleetInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
	}

	taskFleet := map[string]interface{}{
		"name":                      d.Get("name"),
		"target_on_demand_capacity": d.Get("target_on_demand_capacity"),
		"target_spot_capacity":      d.Get("target_spot_capacity"),
		"instance_type_configs":     d.Get("instance_type_configs"),
		"launch_specifications":     d.Get("launch_specifications"),
	}
	addInstanceFleetInput.InstanceFleet = readInstanceFleetConfig(taskFleet, emr.InstanceFleetTypeTask)

	log.Printf("[DEBUG] Creating EMR instance fleet params: %s", addInstanceFleetInput)
	resp, err := conn.AddInstanceFleetWithContext(ctx, addInstanceFleetInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding EMR Instance Fleet: %s", err)
	}

	log.Printf("[DEBUG] Created EMR instance fleet finished: %#v", resp)
	if resp == nil {
		return sdkdiag.AppendErrorf(diags, "creating instance fleet: no instance fleet returned")
	}
	d.SetId(aws.StringValue(resp.InstanceFleetId))

	return diags
}

func resourceInstanceFleetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()
	instanceFleets, err := FetchAllInstanceFleets(ctx, conn, d.Get("cluster_id").(string))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing EMR Instance Fleets for Cluster (%s): %s", d.Get("cluster_id").(string), err)
	}

	fleet := FindInstanceFleetByID(instanceFleets, d.Id())
	if fleet == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "finding EMR Instance Fleet (%s): not found after creation", d.Id())
		}

		log.Printf("[DEBUG] EMR Instance Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err := d.Set("instance_type_configs", flatteninstanceTypeConfigs(fleet.InstanceTypeSpecifications)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_type_configs: %s", err)
	}

	if err := d.Set("launch_specifications", flattenLaunchSpecifications(fleet.LaunchSpecifications)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting launch_specifications: %s", err)
	}
	d.Set("name", fleet.Name)
	d.Set("provisioned_on_demand_capacity", fleet.ProvisionedOnDemandCapacity)
	d.Set("provisioned_spot_capacity", fleet.ProvisionedSpotCapacity)
	d.Set("target_on_demand_capacity", fleet.TargetOnDemandCapacity)
	d.Set("target_spot_capacity", fleet.TargetSpotCapacity)
	return diags
}

func FindInstanceFleetByID(instanceFleets []*emr.InstanceFleet, fleetId string) *emr.InstanceFleet {
	for _, fleet := range instanceFleets {
		if fleet != nil && aws.StringValue(fleet.Id) == fleetId {
			return fleet
		}
	}
	return nil
}

func resourceInstanceFleetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EMRConn()

	log.Printf("[DEBUG] Modify EMR task fleet")

	modifyConfig := &emr.InstanceFleetModifyConfig{
		InstanceFleetId:        aws.String(d.Id()),
		TargetOnDemandCapacity: aws.Int64(int64(d.Get("target_on_demand_capacity").(int))),
		TargetSpotCapacity:     aws.Int64(int64(d.Get("target_spot_capacity").(int))),
	}

	modifyInstanceFleetInput := &emr.ModifyInstanceFleetInput{
		ClusterId:     aws.String(d.Get("cluster_id").(string)),
		InstanceFleet: modifyConfig,
	}

	_, err := conn.ModifyInstanceFleetWithContext(ctx, modifyInstanceFleetInput)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "modifying EMR Instance Fleet (%s): %s", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{emr.InstanceFleetStateProvisioning, emr.InstanceFleetStateBootstrapping, emr.InstanceFleetStateResizing},
		Target:     []string{emr.InstanceFleetStateRunning},
		Refresh:    instanceFleetStateRefresh(ctx, conn, d.Get("cluster_id").(string), d.Id()),
		Timeout:    75 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for instance (%s) to terminate: %s", d.Id(), err)
	}

	return append(diags, resourceInstanceFleetRead(ctx, d, meta)...)
}

func instanceFleetStateRefresh(ctx context.Context, conn *emr.EMR, clusterID, ifID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		instanceFleets, err := FetchAllInstanceFleets(ctx, conn, clusterID)
		if err != nil {
			return nil, "Not Found", err
		}

		fleet := FindInstanceFleetByID(instanceFleets, ifID)
		if fleet == nil {
			return nil, "Not Found", err
		}

		if fleet.Status == nil || fleet.Status.State == nil {
			log.Printf("[WARN] ERM Instance Fleet found, but without state")
			return nil, "Undefined", fmt.Errorf("undefined EMR Cluster Instance Fleet state")
		}

		return fleet, *fleet.Status.State, nil
	}
}

func resourceInstanceFleetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	log.Printf("[WARN] AWS EMR Instance Fleet does not support DELETE; resizing cluster to zero before removing from state")
	conn := meta.(*conns.AWSClient).EMRConn()

	clusterId := d.Get("cluster_id").(string)

	modifyInstanceFleetInput := &emr.ModifyInstanceFleetInput{
		ClusterId: aws.String(clusterId),
		InstanceFleet: &emr.InstanceFleetModifyConfig{
			InstanceFleetId:        aws.String(d.Id()),
			TargetOnDemandCapacity: aws.Int64(0),
			TargetSpotCapacity:     aws.Int64(0),
		},
	}

	_, err := conn.ModifyInstanceFleetWithContext(ctx, modifyInstanceFleetInput)

	if tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "instance fleet may only be modified when the cluster is running or waiting") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting/modifying EMR Instance Fleet (%s): %s", d.Id(), err)
	}

	return diags
}
