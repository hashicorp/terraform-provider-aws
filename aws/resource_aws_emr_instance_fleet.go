package aws

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/hashcode"
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
		addInstanceFleetInput.InstanceFleet = readInstanceFleetConfig(l[0].(map[string]interface{}))
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

	if err := d.Set("instance_fleet", flattenInstanceFleet(fleet)); err != nil {
		return fmt.Errorf("error setting instance_fleet: %s", err)
	}

	return nil
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

// duplicate from resource_aws_emr_cluster.go
func expandEbsConfiguration(ebsConfigurations []interface{}) *emr.EbsConfiguration {
	ebsConfig := &emr.EbsConfiguration{}
	ebsConfigs := make([]*emr.EbsBlockDeviceConfig, 0)
	for _, ebsConfiguration := range ebsConfigurations {
		cfg := ebsConfiguration.(map[string]interface{})
		ebsBlockDeviceConfig := &emr.EbsBlockDeviceConfig{
			VolumesPerInstance: aws.Int64(int64(cfg["volumes_per_instance"].(int))),
			VolumeSpecification: &emr.VolumeSpecification{
				SizeInGB:   aws.Int64(int64(cfg["size"].(int))),
				VolumeType: aws.String(cfg["type"].(string)),
			},
		}
		if v, ok := cfg["iops"].(int); ok && v != 0 {
			ebsBlockDeviceConfig.VolumeSpecification.Iops = aws.Int64(int64(v))
		}
		ebsConfigs = append(ebsConfigs, ebsBlockDeviceConfig)
	}
	ebsConfig.EbsBlockDeviceConfigs = ebsConfigs
	return ebsConfig
}

// duplicate from resource_aws_emr_cluster.go
func expandInstanceTypeConfigs(instanceTypeConfigs []interface{}) []*emr.InstanceTypeConfig {
	configsOut := []*emr.InstanceTypeConfig{}

	for _, raw := range instanceTypeConfigs {
		configAttributes := raw.(map[string]interface{})

		config := &emr.InstanceTypeConfig{
			InstanceType: aws.String(configAttributes["instance_type"].(string)),
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
		}

		configsOut = append(configsOut, config)
	}

	return configsOut
}

// duplicate from resource_aws_emr_cluster.go
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

// duplicate from resource_aws_emr_cluster.go
func expandLaunchSpecification(launchSpecification map[string]interface{}) *emr.InstanceFleetProvisioningSpecifications {
	onDemandSpecification := launchSpecification["on_demand_specification"].([]interface{})
	spotSpecification := launchSpecification["spot_specification"].([]interface{})

	fleetSpecification := &emr.InstanceFleetProvisioningSpecifications{}

	if len(onDemandSpecification) > 0 {
		fleetSpecification.OnDemandSpecification = &emr.OnDemandProvisioningSpecification{
			AllocationStrategy: aws.String(onDemandSpecification[0].(map[string]interface{})["allocation_strategy"].(string)),
		}
	}

	if len(spotSpecification) > 0 {
		configAttributes := spotSpecification[0].(map[string]interface{})
		spotProvisioning := &emr.SpotProvisioningSpecification{
			TimeoutAction:          aws.String(configAttributes["timeout_action"].(string)),
			TimeoutDurationMinutes: aws.Int64(int64(configAttributes["timeout_duration_minutes"].(int))),
		}
		if v, ok := configAttributes["block_duration_minutes"]; ok && v != 0 {
			spotProvisioning.BlockDurationMinutes = aws.Int64(int64(v.(int)))
		}
		if v, ok := configAttributes["allocation_strategy"]; ok {

			spotProvisioning.AllocationStrategy = aws.String(v.(string))
		}

		fleetSpecification.SpotSpecification = spotProvisioning
	}

	return fleetSpecification
}

// duplicate from resource_aws_emr_cluster.go
func flattenInstanceFleet(instanceFleet *emr.InstanceFleet) []interface{} {
	if instanceFleet == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"id":                             aws.StringValue(instanceFleet.Id),
		"name":                           aws.StringValue(instanceFleet.Name),
		"instance_fleet_type":            aws.StringValue(instanceFleet.InstanceFleetType),
		"target_on_demand_capacity":      aws.Int64Value(instanceFleet.TargetOnDemandCapacity),
		"target_spot_capacity":           aws.Int64Value(instanceFleet.TargetSpotCapacity),
		"provisioned_on_demand_capacity": aws.Int64Value(instanceFleet.ProvisionedOnDemandCapacity),
		"provisioned_spot_capacity":      aws.Int64Value(instanceFleet.ProvisionedSpotCapacity),
		"instance_type_configs":          flatteninstanceTypeConfigs(instanceFleet.InstanceTypeSpecifications),
		"launch_specifications":          flattenLaunchSpecifications(instanceFleet.LaunchSpecifications),
	}

	return []interface{}{m}

}

// duplicate from resource_aws_emr_cluster.go
func flatteninstanceTypeConfigs(instanceTypeSpecifications []*emr.InstanceTypeSpecification) *schema.Set {
	instanceTypeConfigs := make([]interface{}, 0)

	for _, itc := range instanceTypeSpecifications {
		flattenTypeConfig := make(map[string]interface{})

		if itc.BidPrice != nil {
			flattenTypeConfig["bid_price"] = aws.StringValue(itc.BidPrice)
		}

		if itc.BidPriceAsPercentageOfOnDemandPrice != nil {
			flattenTypeConfig["bid_price_as_percentage_of_on_demand_price"] = aws.Float64Value(itc.BidPriceAsPercentageOfOnDemandPrice)
		}

		flattenTypeConfig["instance_type"] = aws.StringValue(itc.InstanceType)
		flattenTypeConfig["weighted_capacity"] = int(aws.Int64Value(itc.WeightedCapacity))

		flattenTypeConfig["ebs_config"] = flattenEBSConfig(itc.EbsBlockDevices)

		instanceTypeConfigs = append(instanceTypeConfigs, flattenTypeConfig)
	}

	return schema.NewSet(resourceAwsEMRInstanceTypeConfigHash, instanceTypeConfigs)
}

// duplicate from resource_aws_emr_cluster.go
func flattenLaunchSpecifications(launchSpecifications *emr.InstanceFleetProvisioningSpecifications) []interface{} {
	if launchSpecifications == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"on_demand_specification": flattenOnDemandSpecification(launchSpecifications.OnDemandSpecification),
		"spot_specification":      flattenSpotSpecification(launchSpecifications.SpotSpecification),
	}
	return []interface{}{m}
}

// duplicate from resource_aws_emr_cluster.go
func flattenOnDemandSpecification(onDemandSpecification *emr.OnDemandProvisioningSpecification) []interface{} {
	if onDemandSpecification == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"allocation_strategy": aws.StringValue(onDemandSpecification.AllocationStrategy),
	}
	return []interface{}{m}
}

// duplicate from resource_aws_emr_cluster.go
func flattenSpotSpecification(spotSpecification *emr.SpotProvisioningSpecification) []interface{} {
	if spotSpecification == nil {
		return []interface{}{}
	}
	m := map[string]interface{}{
		"timeout_action":           aws.StringValue(spotSpecification.TimeoutAction),
		"timeout_duration_minutes": aws.Int64Value(spotSpecification.TimeoutDurationMinutes),
	}
	if spotSpecification.BlockDurationMinutes != nil {
		m["block_duration_minutes"] = aws.Int64Value(spotSpecification.BlockDurationMinutes)
	}
	if spotSpecification.AllocationStrategy != nil {
		m["allocation_strategy"] = aws.StringValue(spotSpecification.AllocationStrategy)
	}

	return []interface{}{m}
}

// duplicate from resource_aws_emr_cluster.go
func fetchAllEMRInstanceFleets(conn *emr.EMR, clusterID string) ([]*emr.InstanceFleet, error) {
	input := &emr.ListInstanceFleetsInput{
		ClusterId: aws.String(clusterID),
	}
	var fleets []*emr.InstanceFleet

	err := conn.ListInstanceFleetsPages(input, func(page *emr.ListInstanceFleetsOutput, lastPage bool) bool {
		fleets = append(fleets, page.InstanceFleets...)

		return !lastPage
	})

	return fleets, err
}

// duplicate from resource_aws_emr_cluster.go
func readInstanceFleetConfig(data map[string]interface{}) *emr.InstanceFleetConfig {

	config := &emr.InstanceFleetConfig{
		InstanceFleetType:      aws.String(data["instance_fleet_type"].(string)),
		Name:                   aws.String(data["name"].(string)),
		TargetOnDemandCapacity: aws.Int64(int64(data["target_on_demand_capacity"].(int))),
		TargetSpotCapacity:     aws.Int64(int64(data["target_spot_capacity"].(int))),
	}

	if v, ok := data["instance_type_configs"].(*schema.Set); ok && v.Len() > 0 {
		config.InstanceTypeConfigs = expandInstanceTypeConfigs(v.List())
	}

	if v, ok := data["launch_specifications"].([]interface{}); ok && len(v) == 1 {
		config.LaunchSpecifications = expandLaunchSpecification(v[0].(map[string]interface{}))
	}

	return config
}

// duplicate from resource_aws_emr_cluster.go
func resourceAwsEMRInstanceTypeConfigHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["instance_type"].(string)))
	if v, ok := m["bid_price"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["weighted_capacity"]; ok && v.(int) > 0 {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["bid_price_as_percentage_of_on_demand_price"]; ok && v.(float64) != 0 {
		buf.WriteString(fmt.Sprintf("%f-", v.(float64)))
	}
	return hashcode.String(buf.String())
}

// duplicate from resource_aws_emr_cluster.go
func InstanceFleetConfigSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_fleet_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(emr.InstanceFleetType_Values(), false),
			},
			"instance_type_configs": {
				Type:     schema.TypeSet,
				Optional: true,
				// ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bid_price": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"bid_price_as_percentage_of_on_demand_price": {
							Type:     schema.TypeFloat,
							Optional: true,
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
									},
									"properties": {
										Type:     schema.TypeMap,
										Optional: true,
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
										ValidateFunc: validateAwsEmrEbsVolumeType(),
									},
									"volumes_per_instance": {
										Type:     schema.TypeInt,
										Optional: true,
										ForceNew: true,
										Default:  1,
									},
								},
							},
							Set: resourceAwsEMRClusterEBSConfigHash,
						},
						"instance_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"weighted_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  1,
						},
					},
				},
				Set: resourceAwsEMRInstanceTypeConfigHash,
			},
			"launch_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				// ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_demand_specification": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allocation_strategy": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(emr.OnDemandProvisioningAllocationStrategy_Values(), false),
									},
								},
							},
						},
						"spot_specification": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"allocation_strategy": {
										Type:         schema.TypeString,
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
										ValidateFunc: validation.StringInSlice(emr.SpotProvisioningTimeoutAction_Values(), false),
									},
									"timeout_duration_minutes": {
										Type:     schema.TypeInt,
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
