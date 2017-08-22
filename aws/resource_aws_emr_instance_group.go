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

var emrInstanceGroupNotFound = errors.New("No matching EMR Instance Group")

func resourceAwsEMRInstanceGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsEMRInstanceGroupCreate,
		Read:   resourceAwsEMRInstanceGroupRead,
		Update: resourceAwsEMRInstanceGroupUpdate,
		Delete: resourceAwsEMRInstanceGroupDelete,
		Schema: map[string]*schema.Schema{
			"cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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
				Elem:     ebsConfigurationSchema(),
			},
		},
	}
}

func resourceAwsEMRInstanceGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn

	clusterId := d.Get("cluster_id").(string)
	instanceType := d.Get("instance_type").(string)
	instanceCount := d.Get("instance_count").(int)
	groupName := d.Get("name").(string)

	ebsConfig := &emr.EbsConfiguration{}
	if v, ok := d.GetOk("ebs_config"); ok && v.(*schema.Set).Len() == 1 {
		ebsConfig = expandEbsConfiguration(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("ebs_optimized"); ok {
		ebsConfig.EbsOptimized = aws.Bool(v.(bool))
	}

	params := &emr.AddInstanceGroupsInput{
		InstanceGroups: []*emr.InstanceGroupConfig{
			{
				InstanceRole:     aws.String(emr.InstanceRoleTypeTask),
				InstanceCount:    aws.Int64(int64(instanceCount)),
				InstanceType:     aws.String(instanceType),
				Name:             aws.String(groupName),
				EbsConfiguration: ebsConfig,
			},
		},
		JobFlowId: aws.String(clusterId),
	}

	log.Printf("[DEBUG] Creating EMR task group params: %s", params)
	resp, err := conn.AddInstanceGroups(params)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Created EMR task group finished: %#v", resp)
	if resp == nil || len(resp.InstanceGroupIds) == 0 {
		return fmt.Errorf("Error creating instance groups: no instance group returned")
	}
	d.SetId(*resp.InstanceGroupIds[0])

	return nil
}

func resourceAwsEMRInstanceGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).emrconn
	group, err := fetchEMRInstanceGroup(conn, d.Get("cluster_id").(string), d.Id())
	if err != nil {
		switch err {
		case emrInstanceGroupNotFound:
			log.Printf("[DEBUG] EMR Instance Group (%s) not found, removing", d.Id())
			d.SetId("")
			return nil
		default:
			return err
		}
	}

	// Guard against the chance of fetchEMRInstanceGroup returning nil group but
	// not a emrInstanceGroupNotFound error
	if group == nil {
		log.Printf("[DEBUG] EMR Instance Group (%s) not found, removing", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("name", group.Name)
	d.Set("instance_count", group.RequestedInstanceCount)
	d.Set("running_instance_count", group.RunningInstanceCount)
	d.Set("instance_type", group.InstanceType)
	if group.Status != nil && group.Status.State != nil {
		d.Set("status", group.Status.State)
	}

	return nil
}

func fetchAllEMRInstanceGroups(conn *emr.EMR, clusterId string) ([]*emr.InstanceGroup, error) {
	req := &emr.ListInstanceGroupsInput{
		ClusterId: aws.String(clusterId),
	}

	var groups []*emr.InstanceGroup
	marker := aws.String("initial")
	for marker != nil {
		log.Printf("[DEBUG] EMR Cluster Instance Marker: %s", *marker)
		respGrps, errGrps := conn.ListInstanceGroups(req)
		if errGrps != nil {
			return nil, fmt.Errorf("[ERR] Error reading EMR cluster (%s): %s", clusterId, errGrps)
		}
		if respGrps == nil {
			return nil, fmt.Errorf("[ERR] Error reading EMR Instance Group for cluster (%s)", clusterId)
		}

		if respGrps.InstanceGroups != nil {
			for _, g := range respGrps.InstanceGroups {
				groups = append(groups, g)
			}
		} else {
			log.Printf("[DEBUG] EMR Instance Group list was empty")
		}
		marker = respGrps.Marker
	}

	if len(groups) == 0 {
		return nil, fmt.Errorf("[WARN] No instance groups found for EMR Cluster (%s)", clusterId)
	}

	return groups, nil
}

func fetchEMRInstanceGroup(conn *emr.EMR, clusterId, groupId string) (*emr.InstanceGroup, error) {
	groups, err := fetchAllEMRInstanceGroups(conn, clusterId)
	if err != nil {
		return nil, err
	}

	var group *emr.InstanceGroup
	for _, ig := range groups {
		if groupId == *ig.Id {
			group = ig
			break
		}
	}

	if group != nil {
		return group, nil
	}

	return nil, emrInstanceGroupNotFound
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
		Pending:    []string{emr.InstanceGroupStateProvisioning, emr.InstanceGroupStateBootstrapping, emr.InstanceGroupStateResizing},
		Target:     []string{emr.InstanceGroupStateRunning},
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
	if err != nil {
		return err
	}
	return nil
}
