package aws

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/structure"
	"github.com/hashicorp/terraform/helper/validation"
)

var errEMRInstanceGroupNotFound = errors.New("No matching EMR Instance Group")

func resourceAwsEMRInstanceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEMRInstanceGroupCreate,
		Read:   resourceAwsEMRInstanceGroupRead,
		Update: resourceAwsEMRInstanceGroupUpdate,
		Delete: resourceAwsEMRInstanceGroupDelete,
		Schema: map[string]*schema.Schema{
			"autoscaling_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressEquivalentJsonDiffs,
				ValidateFunc:     validation.ValidateJsonString,
				StateFunc: func(v interface{}) string {
					s, _ := structure.NormalizeJsonString(v.(string))
					return s
				},
			},
			"bid_price": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_role": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "TASK",
			},
			"instance_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"instance_count": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
			},
			"running_instance_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"iops": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateAwsEmrEbsVolumeType(),
						},
						"volumes_per_instance": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

// Populates an emr.EbsConfiguration struct
func readEmrEBSConfig(d *schema.ResourceData) *emr.EbsConfiguration {
	result := &emr.EbsConfiguration{}
	if v, ok := d.GetOk("ebs_optimized"); ok {
		result.EbsOptimized = aws.Bool(v.(bool))
	}

	ebsConfigs := make([]*emr.EbsBlockDeviceConfig, 0)
	if rawConfig, ok := d.GetOk("ebs_config"); ok {
		configList := rawConfig.(*schema.Set).List()
		for _, config := range configList {
			conf := config.(map[string]interface{})
			ebs := &emr.EbsBlockDeviceConfig{}
			volumeSpec := &emr.VolumeSpecification{
				SizeInGB:   aws.Int64(int64(conf["size"].(int))),
				VolumeType: aws.String(conf["type"].(string)),
			}
			if v, ok := conf["iops"].(int); ok && v != 0 {
				volumeSpec.Iops = aws.Int64(int64(v))
			}
			if v, ok := conf["volumes_per_instance"].(int); ok && v != 0 {
				ebs.VolumesPerInstance = aws.Int64(int64(v))
			}
			ebs.VolumeSpecification = volumeSpec
			ebsConfigs = append(ebsConfigs, ebs)
		}
	}
	result.EbsBlockDeviceConfigs = ebsConfigs
	return result
}

func resourceAwsEMRInstanceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	instanceRole := d.Get("instance_role").(string)
	groupConfig := &emr.InstanceGroupConfig{
		EbsConfiguration: readEmrEBSConfig(d),
		InstanceRole:     aws.String(instanceRole),
		InstanceCount:    aws.Int64(int64(d.Get("instance_count").(int))),
		InstanceType:     aws.String(d.Get("instance_type").(string)),
		Name:             aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("autoscaling_policy"); ok {
		var autoScalingPolicy *emr.AutoScalingPolicy

		if err := json.Unmarshal([]byte(v.(string)), &autoScalingPolicy); err != nil {
			return fmt.Errorf("[DEBUG] error parsing Auto Scaling Policy %s", err)
		}
		groupConfig.AutoScalingPolicy = autoScalingPolicy
	}

	if v, ok := d.GetOk("bid_price"); ok {
		groupConfig.BidPrice = aws.String(v.(string))
	}

	params := &emr.AddInstanceGroupsInput{
		InstanceGroups: []*emr.InstanceGroupConfig{groupConfig},
		JobFlowId:      aws.String(d.Get("cluster_id").(string)),
	}

	log.Printf("[DEBUG] Creating EMR %s group with the following params: %s", instanceRole, params)
	resp, err := conn.AddInstanceGroups(params)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Created EMR %s group finished: %#v", instanceRole, resp)
	if resp == nil || len(resp.InstanceGroupIds) == 0 {
		return fmt.Errorf("Error creating instance groups: no instance group returned")
	}
	d.SetId(*resp.InstanceGroupIds[0])

	return resourceAwsEMRInstanceGroupRead(d, meta)
}

func resourceAwsEMRInstanceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	ig, err := fetchEMRInstanceGroup(conn, d.Get("cluster_id").(string), d.Id())

	if err == errEMRInstanceGroupNotFound {
		log.Printf("[DEBUG] EMR Instance Group (%s) not found, removing", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return err
	}

	// Guard against the chance of fetchEMRInstanceGroup returning nil group but not a errEMRInstanceGroupNotFound error
	if ig == nil {
		log.Printf("[DEBUG] EMR Instance Group (%s) not found, removing", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("instance_count", ig.RequestedInstanceCount)
	d.Set("instance_role", ig.InstanceGroupType)
	d.Set("instance_type", ig.InstanceType)
	d.Set("name", ig.Name)
	d.Set("running_instance_count", ig.RunningInstanceCount)

	if ig.BidPrice != nil {
		d.Set("bid_price", ig.BidPrice)
	}

	if ig.Status != nil && ig.Status.State != nil {
		d.Set("status", ig.Status.State)
	}

	var autoscalingPolicyString string
	if ig.AutoScalingPolicy != nil {
		// AutoScalingPolicy has an additional Status field and null values that are causing a new hashcode to be generated for `instance_group`.
		// We are purposefully omitting that field and the null values here when we flatten the autoscaling policy string for the statefile.
		for i, rule := range ig.AutoScalingPolicy.Rules {
			for j, dimension := range rule.Trigger.CloudWatchAlarmDefinition.Dimensions {
				if *dimension.Key == "JobFlowId" {
					tmpDimensions := append(ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions[:j], ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions[j+1:]...)
					ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions = tmpDimensions
				}
			}

			if len(ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions) == 0 {
				ig.AutoScalingPolicy.Rules[i].Trigger.CloudWatchAlarmDefinition.Dimensions = nil
			}
		}

		tmpAutoScalingPolicy := emr.AutoScalingPolicy{
			Constraints: ig.AutoScalingPolicy.Constraints,
			Rules:       ig.AutoScalingPolicy.Rules,
		}
		autoscalingPolicyConstraintsBytes, err := json.Marshal(tmpAutoScalingPolicy.Constraints)
		if err != nil {
			return fmt.Errorf("error parsing EMR Cluster Instance Groups AutoScalingPolicy Constraints: %s", err)
		}

		autoscalingPolicyRulesBytes, err := normalizeEmptyRules(tmpAutoScalingPolicy.Rules)
		if err != nil {
			return fmt.Errorf("error parsing EMR Cluster Instance Groups AutoScalingPolicy Rules: %s", err)
		}

		autoscalingPolicyString = fmt.Sprintf("{\"Constraints\":%s,\"Rules\":%s}", string(autoscalingPolicyConstraintsBytes), string(autoscalingPolicyRulesBytes))
	}
	d.Set("autoscaling_policy", autoscalingPolicyString)

	return nil
}

func fetchAllEMRInstanceGroups(conn *emr.EMR, clusterID string) ([]*emr.InstanceGroup, error) {
	req := &emr.ListInstanceGroupsInput{
		ClusterId: aws.String(clusterID),
	}

	var groups []*emr.InstanceGroup
	marker := aws.String("intitial")
	for marker != nil {
		log.Printf("[DEBUG] EMR Cluster Instance Marker: %s", *marker)
		respGrps, errGrps := conn.ListInstanceGroups(req)
		if errGrps != nil {
			return nil, fmt.Errorf("Error reading EMR cluster (%s): %s", clusterID, errGrps)
		}
		if respGrps == nil {
			return nil, fmt.Errorf("Error reading EMR Instance Group for cluster (%s)", clusterID)
		}

		if respGrps.InstanceGroups != nil {
			groups = append(groups, respGrps.InstanceGroups...)
		} else {
			log.Printf("[DEBUG] EMR Instance Group list was empty")
		}
		marker = respGrps.Marker
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("No instance groups found for EMR Cluster (%s)", clusterID)
	}

	return groups, nil
}

func fetchEMRInstanceGroup(conn *emr.EMR, clusterID, groupID string) (*emr.InstanceGroup, error) {
	// Is this needed or can we consolidate this a bit?

	groups, err := fetchAllEMRInstanceGroups(conn, clusterID)
	if err != nil {
		return nil, err
	}

	var group *emr.InstanceGroup
	for _, ig := range groups {
		if groupID == *ig.Id {
			group = ig
			break
		}
	}

	if group != nil {
		return group, nil
	}

	return nil, errEMRInstanceGroupNotFound
}

func resourceAwsEMRInstanceGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	log.Printf("[DEBUG] Modify EMR task group")
	instanceCount := d.Get("instance_count").(int)

	params := &emr.ModifyInstanceGroupsInput{
		InstanceGroups: []*emr.InstanceGroupModifyConfig{
			{
				InstanceGroupId: aws.String(d.Id()),
				InstanceCount:   aws.Int64(int64(instanceCount)),
			},
		},
	}

	_, err := conn.ModifyInstanceGroups(params)
	if err != nil {
		return err
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"PROVISIONING", "BOOTSTRAPPING", "RESIZING"},
		Target:     []string{"RUNNING"},
		Refresh:    instanceGroupStateRefresh(conn, d.Get("cluster_id").(string), d.Id()),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf(
			"Error waiting for instance (%s) to terminate: %s", d.Id(), err)
	}

	return resourceAwsEMRInstanceGroupRead(d, meta)
}

func instanceGroupStateRefresh(conn *emr.EMR, clusterID, igID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		group, err := fetchEMRInstanceGroup(conn, clusterID, igID)
		if err != nil {
			return nil, "Not Found", err
		}

		if group.Status == nil || group.Status.State == nil {
			log.Printf("[WARN] ERM Instance Group found, but without state")
			return nil, "Undefined", fmt.Errorf("Undefined EMR Cluster Instance Group state")
		}

		return group, *group.Status.State, nil
	}
}

func resourceAwsEMRInstanceGroupDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[WARN] AWS EMR Instance Group does not support DELETE; resizing cluster to zero before removing from state")
	conn := meta.(*AWSClient).emrconn
	params := &emr.ModifyInstanceGroupsInput{
		InstanceGroups: []*emr.InstanceGroupModifyConfig{
			{
				InstanceGroupId: aws.String(d.Id()),
				InstanceCount:   aws.Int64(0),
			},
		},
	}

	_, err := conn.ModifyInstanceGroups(params)
	return err
}
