package aws

import (
	"errors"
	"log"
	"time"

	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

var emrInstanceFleetNotFound = errors.New("no matching EMR Instance Fleet")

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
				ValidateFunc: validateAwsEmrInstanceFleetType,
			},
			"instance_type_configs": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     instanceTypeConfigSchema(),
			},
			"launch_specifications": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MinItems: 1,
				MaxItems: 1,
				Elem:     launchSpecificationsSchema(),
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"target_on_demand_capacity": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
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
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func instanceTypeConfigSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bid_price": {
				Type:     schema.TypeString,
				Optional: true,
				Required: false,
				ForceNew: true,
			},
			"bid_price_as_percentage_of_on_demand_price": {
				Type:     schema.TypeFloat,
				Optional: true,
				ForceNew: true,
			},
			"configurations": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     configurationSchema(),
			},
			"ebs_optimized": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"ebs_config": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     ebsConfigurationSchema(),
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
	}
}

func launchSpecificationsSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"spot_specification": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				MinItems: 1,
				Elem:     spotSpecificationSchema(),
			},
		},
	}
}

func spotSpecificationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"block_duration_minutes": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  0,
			},
			"timeout_action": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateAwsEmrSpotProvisioningTimeOutAction,
			},
			"timeout_duration_minutes": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func configurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"classification": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"configurations": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     additionalConfigurationSchema(),
			},
			"properties": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func additionalConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"classification": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"properties": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
			},
		},
	}
}

func expandInstanceFleetConfig(data *schema.ResourceData) *emr.InstanceFleetConfig {
	configInstanceFleetType := data.Get("instance_fleet_type").(string)
	configName := data.Get("name").(string)
	configTargetOnDemandCapacity := data.Get("target_on_demand_capacity").(int)
	configTargetSpotCapacity := data.Get("target_spot_capacity").(int)

	config := &emr.InstanceFleetConfig{
		InstanceFleetType: aws.String(configInstanceFleetType),
		Name:              aws.String(configName),
		TargetOnDemandCapacity: aws.Int64(int64(configTargetOnDemandCapacity)),
		TargetSpotCapacity:     aws.Int64(int64(configTargetSpotCapacity)),
	}

	if v, ok := data.Get("instance_type_configs").(*schema.Set); ok && v.Len() > 0 {
		config.InstanceTypeConfigs = expandInstanceTypeConfigs(v.List())
	}

	if v, ok := data.Get("launch_specifications").(*schema.Set); ok && v.Len() == 1 {
		config.LaunchSpecifications = expandLaunchSpecification(v.List()[0])
	}

	return config
}

func expandInstanceFleetConfigs(instanceFleetConfigs []interface{}) []*emr.InstanceFleetConfig {
	configsOut := []*emr.InstanceFleetConfig{}

	for _, raw := range instanceFleetConfigs {
		configAttributes := raw.(map[string]interface{})

		configInstanceFleetType := configAttributes["instance_fleet_type"].(string)
		configName := configAttributes["name"].(string)
		configTargetOnDemandCapacity := configAttributes["target_on_demand_capacity"].(int)
		configTargetSpotCapacity := configAttributes["target_spot_capacity"].(int)

		config := &emr.InstanceFleetConfig{
			InstanceFleetType: aws.String(configInstanceFleetType),
			Name:              aws.String(configName),
			TargetOnDemandCapacity: aws.Int64(int64(configTargetOnDemandCapacity)),
			TargetSpotCapacity:     aws.Int64(int64(configTargetSpotCapacity)),
		}

		if v, ok := configAttributes["instance_type_configs"].(*schema.Set); ok && v.Len() > 0 {
			config.InstanceTypeConfigs = expandInstanceTypeConfigs(v.List())
		}

		if v, ok := configAttributes["launch_specifications"].(*schema.Set); ok && v.Len() == 1 {
			config.LaunchSpecifications = expandLaunchSpecification(v.List()[0])
		}

		configsOut = append(configsOut, config)
	}

	return configsOut
}

func expandInstanceTypeConfigs(instanceTypeConfigs []interface{}) []*emr.InstanceTypeConfig {
	configsOut := []*emr.InstanceTypeConfig{}

	for _, raw := range instanceTypeConfigs {
		configAttributes := raw.(map[string]interface{})

		configInstanceType := configAttributes["instance_type"].(string)

		config := &emr.InstanceTypeConfig{
			InstanceType: aws.String(configInstanceType),
		}

		if bidPrice, ok := configAttributes["bid_price"]; ok {
			if bidPrice != "" {
				config.BidPrice = aws.String(bidPrice.(string))
			}
		}

		if v, ok := configAttributes["bid_price_as_percentage_of_on_demand_price"].(float64); ok && v != 0 {
			config.BidPriceAsPercentageOfOnDemandPrice = aws.Float64(v)
		}

		if v, ok := configAttributes["weighted_capacity"].(int); ok {
			config.WeightedCapacity = aws.Int64(int64(v))
		}

		if v, ok := configAttributes["configurations"].(*schema.Set); ok && v.Len() > 0 {
			config.Configurations = expandConfigurations(v.List())
		}

		if v, ok := configAttributes["ebs_config"].(*schema.Set); ok && v.Len() == 1 {
			config.EbsConfiguration = expandEbsConfiguration(v.List())

			if v, ok := configAttributes["ebs_optimized"].(bool); ok {
				config.EbsConfiguration.EbsOptimized = aws.Bool(v)
			}
		}

		configsOut = append(configsOut, config)
	}

	return configsOut
}

func expandConfigurations(configurations []interface{}) []*emr.Configuration {
	configsOut := []*emr.Configuration{}

	for _, raw := range configurations {
		configAttributes := raw.(map[string]interface{})

		config := &emr.Configuration{}

		if v, ok := configAttributes["classification"].(string); ok {
			config.Classification = aws.String(v)
		}

		if rawConfig, ok := configAttributes["configurations"]; ok {
			config.Configurations = expandConfigurations(rawConfig.([]interface{}))
		}

		if v, ok := configAttributes["properties"]; ok {
			properties := make(map[string]string)
			for k, v := range v.(map[string]interface{}) {
				properties[k] = v.(string)
			}
			config.Properties = aws.StringMap(properties)
		}

		configsOut = append(configsOut, config)
	}

	return configsOut
}

func expandLaunchSpecification(launchSpecification interface{}) *emr.InstanceFleetProvisioningSpecifications {
	configAttributes := launchSpecification.(map[string]interface{})

	return &emr.InstanceFleetProvisioningSpecifications{
		SpotSpecification: expandSpotSpecification(configAttributes["spot_specification"].(*schema.Set).List()[0]),
	}
}

func expandSpotSpecification(spotSpecification interface{}) *emr.SpotProvisioningSpecification {
	configAttributes := spotSpecification.(map[string]interface{})

	spotProvisioning := &emr.SpotProvisioningSpecification{
		TimeoutAction:          aws.String(configAttributes["timeout_action"].(string)),
		TimeoutDurationMinutes: aws.Int64(int64(configAttributes["timeout_duration_minutes"].(int))),
	}

	if v, ok := configAttributes["block_duration_minutes"]; ok && v != 0 {
		spotProvisioning.BlockDurationMinutes = aws.Int64(int64(v.(int)))
	}

	return spotProvisioning
}

func resourceAwsEMRInstanceFleetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	clusterId := d.Get("cluster_id").(string)
	instanceFleetConfig := expandInstanceFleetConfig(d)

	addInstanceFleetInput := &emr.AddInstanceFleetInput{
		ClusterId:     aws.String(clusterId),
		InstanceFleet: instanceFleetConfig,
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
	fleet, err := fetchEMRInstanceFleet(conn, d.Get("cluster_id").(string), d.Id())
	if err != nil {
		switch err {
		case emrInstanceFleetNotFound:
			log.Printf("[DEBUG] EMR Instance Fleet (%s) not found, removing", d.Id())
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	// Guard against the chance of fetchEMRInstanceFleet returning nil fleet but
	// not a emrInstanceFleetNotFound error
	if fleet == nil {
		log.Printf("[DEBUG] EMR Instance Fleet (%s) not found, removing", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("name", fleet.Name)
	d.Set("provisioned_on_demand_capacity", fleet.ProvisionedOnDemandCapacity)
	d.Set("provisioned_spot_capacity", fleet.ProvisionedSpotCapacity)
	if fleet.Status != nil && fleet.Status.State != nil {
		d.Set("status", fleet.Status.State)
	}

	return nil
}

func fetchAllEMRInstanceFleets(conn *emr.EMR, clusterId string) ([]*emr.InstanceFleet, error) {
	listInstanceFleetsInput := &emr.ListInstanceFleetsInput{
		ClusterId: aws.String(clusterId),
	}

	var fleets []*emr.InstanceFleet
	marker := aws.String("initial")
	for marker != nil {
		log.Printf("[DEBUG] EMR Cluster Instance Marker: %s", *marker)
		respFleets, errFleets := conn.ListInstanceFleets(listInstanceFleetsInput)
		if errFleets != nil {
			return nil, fmt.Errorf("[ERR] Error reading EMR cluster (%s): %s", clusterId, errFleets)
		}
		if respFleets == nil {
			return nil, fmt.Errorf("[ERR] Error reading EMR Instance Fleet for cluster (%s)", clusterId)
		}

		if respFleets.InstanceFleets != nil {
			for _, f := range respFleets.InstanceFleets {
				fleets = append(fleets, f)
			}
		} else {
			log.Printf("[DEBUG] EMR Instance Fleet list was empty")
		}
		marker = respFleets.Marker
	}

	if len(fleets) == 0 {
		return nil, fmt.Errorf("[WARN] No instance fleets found for EMR Cluster (%s)", clusterId)
	}

	return fleets, nil
}

func fetchEMRInstanceFleet(conn *emr.EMR, clusterId, fleetId string) (*emr.InstanceFleet, error) {
	fleets, err := fetchAllEMRInstanceFleets(conn, clusterId)
	if err != nil {
		return nil, err
	}

	var instanceFleet *emr.InstanceFleet
	for _, fleet := range fleets {
		if fleetId == *fleet.Id {
			instanceFleet = fleet
			break
		}
	}

	if instanceFleet != nil {
		return instanceFleet, nil
	}

	return nil, emrInstanceFleetNotFound
}

func resourceAwsEMRInstanceFleetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	log.Printf("[DEBUG] Modify EMR task fleet")
	clusterId := d.Get("cluster_id").(string)
	targetOnDemandCapacity := d.Get("target_on_demand_capacity").(int)
	targetSpotCapacity := d.Get("target_spot_capacity").(int)

	modifyInstanceFleetInput := &emr.ModifyInstanceFleetInput{
		ClusterId: aws.String(clusterId),
		InstanceFleet: &emr.InstanceFleetModifyConfig{
			InstanceFleetId:        aws.String(d.Id()),
			TargetOnDemandCapacity: aws.Int64(int64(targetOnDemandCapacity)),
			TargetSpotCapacity:     aws.Int64(int64(targetSpotCapacity)),
		},
	}

	_, err := conn.ModifyInstanceFleet(modifyInstanceFleetInput)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{emr.InstanceFleetStateProvisioning, emr.InstanceFleetStateBootstrapping, emr.InstanceFleetStateResizing},
		Target:     []string{emr.InstanceFleetStateRunning},
		Refresh:    instanceFleetStateRefresh(conn, d.Get("cluster_id").(string), d.Id()),
		Timeout:    10 * time.Minute,
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
		fleet, err := fetchEMRInstanceFleet(conn, clusterID, ifID)
		if err != nil {
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
