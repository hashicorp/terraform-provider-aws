package aws

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
				ValidateFunc: validation.StringInSlice(emr.InstanceFleetType_Values(), false),
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
									"configurations": {
										Type:     schema.TypeSet,
										Optional: true,
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
			},
			"launch_specifications": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"spot_specification": {
							Type:     schema.TypeList,
							Required: true,
							ForceNew: true,
							MinItems: 1,
							Elem: &schema.Resource{
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

func expandInstanceFleetConfig(data *schema.ResourceData) *emr.InstanceFleetConfig {

	config := &emr.InstanceFleetConfig{
		InstanceFleetType:      aws.String(data.Get("instance_fleet_type").(string)),
		Name:                   aws.String(data.Get("name").(string)),
		TargetOnDemandCapacity: aws.Int64(int64(data.Get("target_on_demand_capacity").(int))),
		TargetSpotCapacity:     aws.Int64(int64(data.Get("target_spot_capacity").(int))),
	}

	if v, ok := data.Get("instance_type_configs").(*schema.Set); ok && v.Len() > 0 {
		config.InstanceTypeConfigs = expandInstanceTypeConfigs(v.List())
	}

	if v, ok := data.Get("launch_specifications").(*schema.Set); ok && v.Len() == 1 {
		config.LaunchSpecifications = expandLaunchSpecification(v.List()[0])
	}

	return config
}

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

	modifyInstanceFleetInput := &emr.ModifyInstanceFleetInput{
		ClusterId: aws.String(d.Get("cluster_id").(string)),
		InstanceFleet: &emr.InstanceFleetModifyConfig{
			InstanceFleetId:        aws.String(d.Id()),
			TargetOnDemandCapacity: aws.Int64(int64(d.Get("target_on_demand_capacity").(int))),
			TargetSpotCapacity:     aws.Int64(int64(d.Get("target_spot_capacity").(int))),
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
