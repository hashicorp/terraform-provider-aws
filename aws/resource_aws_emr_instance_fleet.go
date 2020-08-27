package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsEMRInstanceFleet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEMRInstanceFleetCreate,
		Read:   resourceAwsEMRInstanceFleetRead,
		Update: resourceAwsEMRInstanceFleetUpdate,
		Delete: resourceAwsEMRInstanceFleetDelete,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_fleet_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(emr.InstanceFleetType_Values(), false),
			},
			"instance_fleet": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem:     InstanceFleetConfigSchema(),
			},
		},
	}
}

func resourceAwsEMRInstanceFleetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	addInstanceFleetInput := &emr.AddInstanceFleetInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
	}

	if l := d.Get("instance_fleet").([]interface{}); len(l) > 0 && l[0] != nil {
		addInstanceFleetInput.InstanceFleet = readInstanceFleetConfig(
			l[0].(map[string]interface{}),
			*aws.String(d.Get("instance_fleet_type").(string)))
	}

	log.Printf("[DEBUG] Creating EMR instance fleet params: %s", addInstanceFleetInput)
	resp, err := conn.AddInstanceFleet(addInstanceFleetInput)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Created EMR instance fleet finished: %#v", resp)
	if resp == nil {
		return fmt.Errorf("error creating instance fleet: no instance fleet returned")
	}
	d.SetId(*resp.InstanceFleetId)

	return nil
}

func resourceAwsEMRInstanceFleetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn
	instanceFleets, err := fetchAllEMRInstanceFleets(conn, d.Get("cluster_id").(string))
	if err != nil {
		log.Printf("[DEBUG] EMR doesn't have any Instance Fleet ")
		d.SetId("")
		return nil
	}

	fleet := findInstanceFleetById(instanceFleets, d.Id())
	if fleet == nil {
		log.Printf("[DEBUG] EMR Instance Fleet (%s) not found, removing", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("instance_fleet_type", aws.StringValue(fleet.InstanceFleetType))

	if err := d.Set("instance_fleet", flattenInstanceFleet(fleet)); err != nil {
		return fmt.Errorf("error setting instance_fleet: %s", err)
	}

	return nil
}

func findInstanceFleetById(instanceFleets []*emr.InstanceFleet, fleetId string) *emr.InstanceFleet {
	for _, fleet := range instanceFleets {
		if *fleet.Id == fleetId {
			return fleet
		}
	}
	return nil
}

func resourceAwsEMRInstanceFleetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	log.Printf("[DEBUG] Modify EMR task fleet")

	modifyConfig := &emr.InstanceFleetModifyConfig{
		InstanceFleetId: aws.String(d.Id()),
	}

	data := d.Get("instance_fleet").([]interface{})[0].(map[string]interface{})

	if v, ok := data["target_on_demand_capacity"]; ok {
		modifyConfig.TargetOnDemandCapacity = aws.Int64(int64(v.(int)))
	}
	if v, ok := data["target_spot_capacity"]; ok {
		modifyConfig.TargetSpotCapacity = aws.Int64(int64(v.(int)))
	}

	modifyInstanceFleetInput := &emr.ModifyInstanceFleetInput{
		ClusterId:     aws.String(d.Get("cluster_id").(string)),
		InstanceFleet: modifyConfig,
	}

	_, err := conn.ModifyInstanceFleet(modifyInstanceFleetInput)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{emr.InstanceFleetStateProvisioning, emr.InstanceFleetStateBootstrapping, emr.InstanceFleetStateResizing},
		Target:     []string{emr.InstanceFleetStateRunning},
		Refresh:    instanceFleetStateRefresh(conn, d.Get("cluster_id").(string), d.Id()),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for instance (%s) to terminate: %s", d.Id(), err)
	}

	return resourceAwsEMRInstanceFleetRead(d, meta)
}

func instanceFleetStateRefresh(conn *emr.EMR, clusterID, ifID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		instanceFleets, err := fetchAllEMRInstanceFleets(conn, clusterID)
		if err != nil {
			return nil, "Not Found", err
		}

		fleet := findInstanceFleetById(instanceFleets, ifID)
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

func resourceAwsEMRInstanceFleetDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] AWS EMR Instance Fleet does not support DELETE; resizing cluster to zero before removing from state")
	conn := meta.(*AWSClient).emrconn

	clusterId := d.Get("cluster_id").(string)

	modifyInstanceFleetInput := &emr.ModifyInstanceFleetInput{
		ClusterId: aws.String(clusterId),
		InstanceFleet: &emr.InstanceFleetModifyConfig{
			InstanceFleetId:        aws.String(d.Id()),
			TargetOnDemandCapacity: aws.Int64(0),
			TargetSpotCapacity:     aws.Int64(0),
		},
	}

	_, err := conn.ModifyInstanceFleet(modifyInstanceFleetInput)
	if err != nil {
		return err
	}

	return nil
}
