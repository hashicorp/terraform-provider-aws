package emr

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceInstanceFleet() *schema.Resource {
	return &schema.Resource{
		Create: resourceInstanceFleetCreate,
		Read:   resourceInstanceFleetRead,
		Update: resourceInstanceFleetUpdate,
		Delete: resourceInstanceFleetDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
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

func resourceInstanceFleetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRConn()

	taskFleet := map[string]interface{}{
		"name":                      d.Get("name"),
		"target_on_demand_capacity": d.Get("target_on_demand_capacity"),
		"target_spot_capacity":      d.Get("target_spot_capacity"),
		"instance_type_configs":     d.Get("instance_type_configs"),
		"launch_specifications":     d.Get("launch_specifications"),
	}
	input := &emr.AddInstanceFleetInput{
		ClusterId:     aws.String(d.Get("cluster_id").(string)),
		InstanceFleet: readInstanceFleetConfig(taskFleet, emr.InstanceFleetTypeTask),
	}

	output, err := conn.AddInstanceFleet(input)

	if err != nil {
		return fmt.Errorf("creating EMR Instance Fleet: %w", err)
	}

	d.SetId(aws.StringValue(output.InstanceFleetId))

	return resourceInstanceFleetRead(d, meta)
}

func resourceInstanceFleetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRConn()

	fleet, err := FindInstanceFleetByTwoPartKey(conn, d.Get("cluster_id").(string), d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Instance Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EMR Instance Fleet (%s): %w", d.Id(), err)
	}

	if err := d.Set("instance_type_configs", flatteninstanceTypeConfigs(fleet.InstanceTypeSpecifications)); err != nil {
		return fmt.Errorf("setting instance_type_configs: %w", err)
	}
	if err := d.Set("launch_specifications", flattenLaunchSpecifications(fleet.LaunchSpecifications)); err != nil {
		return fmt.Errorf("setting launch_specifications: %w", err)
	}
	d.Set("name", fleet.Name)
	d.Set("provisioned_on_demand_capacity", fleet.ProvisionedOnDemandCapacity)
	d.Set("provisioned_spot_capacity", fleet.ProvisionedSpotCapacity)
	d.Set("target_on_demand_capacity", fleet.TargetOnDemandCapacity)
	d.Set("target_spot_capacity", fleet.TargetSpotCapacity)

	return nil
}

func resourceInstanceFleetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRConn()

	log.Printf("[DEBUG] Modify EMR task fleet")

	modifyConfig := &emr.InstanceFleetModifyConfig{
		InstanceFleetId:        aws.String(d.Id()),
		TargetOnDemandCapacity: aws.Int64(int64(d.Get("target_on_demand_capacity").(int))),
		TargetSpotCapacity:     aws.Int64(int64(d.Get("target_spot_capacity").(int))),
	}

	input := &emr.ModifyInstanceFleetInput{
		ClusterId:     aws.String(d.Get("cluster_id").(string)),
		InstanceFleet: modifyConfig,
	}

	_, err := conn.ModifyInstanceFleet(input)

	if err != nil {
		return fmt.Errorf("updating EMR Instance Fleet (%s): %w", d.Id(), err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{emr.InstanceFleetStateProvisioning, emr.InstanceFleetStateBootstrapping, emr.InstanceFleetStateResizing},
		Target:     []string{emr.InstanceFleetStateRunning},
		Refresh:    statusInstanceFleet(conn, d.Get("cluster_id").(string), d.Id()),
		Timeout:    75 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 30 * time.Second,
	}

	_, err = stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf("waiting for EMR Instance Fleet (%s) update: %s", d.Id(), err)
	}

	return resourceInstanceFleetRead(d, meta)
}

func resourceInstanceFleetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRConn()

	// AWS EMR Instance Fleet does not support DELETE; resizing cluster to zero before removing from state.
	log.Printf("[DEBUG] Deleting EMR Instance Fleet: %s", d.Id())
	_, err := conn.ModifyInstanceFleet(&emr.ModifyInstanceFleetInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
		InstanceFleet: &emr.InstanceFleetModifyConfig{
			InstanceFleetId:        aws.String(d.Id()),
			TargetOnDemandCapacity: aws.Int64(0),
			TargetSpotCapacity:     aws.Int64(0),
		},
	})

	if tfawserr.ErrMessageContains(err, emr.ErrCodeInvalidRequestException, "instance fleet may only be modified when the cluster is running or waiting") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EMR Instance Fleet (%s): %w", d.Id(), err)
	}

	return nil
}

func FindInstanceFleetByTwoPartKey(conn *emr.EMR, clusterID, fleetID string) (*emr.InstanceFleet, error) {
	input := &emr.ListInstanceFleetsInput{
		ClusterId: aws.String(clusterID),
	}
	var fleets []*emr.InstanceFleet

	err := conn.ListInstanceFleetsPages(input, func(page *emr.ListInstanceFleetsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.InstanceFleets {
			if v != nil && v.Status != nil {
				fleets = append(fleets, v)
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	for _, fleet := range fleets {
		if aws.StringValue(fleet.Id) == fleetID {
			return fleet, nil
		}
	}

	return nil, &resource.NotFoundError{}
}

func statusInstanceFleet(conn *emr.EMR, clusterID, fleetID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindInstanceFleetByTwoPartKey(conn, clusterID, fleetID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status.State), nil
	}
}
