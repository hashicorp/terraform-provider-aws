package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsRDSClusterEndpoint() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsRDSClusterEndpointCreate,
		Read:   resourceAwsRDSClusterEndpointRead,
		Update: resourceAwsRDSClusterEndpointUpdate,
		Delete: resourceAwsRDSClusterEndpointDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cluster_endpoint_identifier": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"cluster_identifier": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validateRdsIdentifier,
			},
			"custom_endpoint_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"READER",
					"ANY",
				}, false),
			},
			"excluded_members": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"static_members"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
			},
			"static_members": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"excluded_members"},
				Elem:          &schema.Schema{Type: schema.TypeString},
				Set:           schema.HashString,
			},
			"endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsRDSClusterEndpointCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	var clusterId string
	if v, ok := d.GetOk("cluster_identifier"); ok {
		clusterId = v.(string)
	}

	var endpointId string
	if v, ok := d.GetOk("cluster_endpoint_identifier"); ok {
		endpointId = v.(string)
	}

	var endpointType string
	if v, ok := d.GetOk("custom_endpoint_type"); ok {
		endpointType = v.(string)
	}

	createClusterEndpointInput := &rds.CreateDBClusterEndpointInput{
		DBClusterIdentifier:         aws.String(clusterId),
		DBClusterEndpointIdentifier: aws.String(endpointId),
		EndpointType:                aws.String(endpointType),
	}

	var staticMembers []*string
	if v := d.Get("static_members"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			str := v.(string)
			staticMembers = append(staticMembers, aws.String(str))
		}
		createClusterEndpointInput.StaticMembers = staticMembers
	}
	var excludedMembers []*string
	if v := d.Get("excluded_members"); v != nil {
		for _, v := range v.(*schema.Set).List() {
			str := v.(string)
			excludedMembers = append(excludedMembers, aws.String(str))
		}
		createClusterEndpointInput.ExcludedMembers = excludedMembers
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.CreateDBClusterEndpoint(createClusterEndpointInput)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error creating RDS Cluster Endpoint: %s", err)
	}

	d.SetId(endpointId)
	return resourceAwsRDSClusterEndpointRead(d, meta)
}

func resourceAwsRDSClusterEndpointRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn

	input := &rds.DescribeDBClusterEndpointsInput{
		DBClusterEndpointIdentifier: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Describing RDS Cluster: %s", input)
	resp, err := conn.DescribeDBClusterEndpoints(input)

	if err != nil {
		return fmt.Errorf("error describing RDS Cluster Endpoints (%s): %s", d.Id(), err)
	}

	if resp == nil {
		return fmt.Errorf("Error retrieving RDS Cluster Endpoints: empty response for: %s", input)
	}

	var clusterEp *rds.DBClusterEndpoint
	for _, e := range resp.DBClusterEndpoints {
		if aws.StringValue(e.DBClusterEndpointIdentifier) == d.Id() {
			clusterEp = e
			break
		}
	}

	if clusterEp == nil {
		log.Printf("[WARN] RDS Cluster Endpoint (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("cluster_endpoint_identifier", clusterEp.DBClusterEndpointIdentifier)
	d.Set("cluster_identifier", clusterEp.DBClusterIdentifier)
	d.Set("arn", clusterEp.DBClusterEndpointArn)
	d.Set("endpoint", clusterEp.Endpoint)
	d.Set("custom_endpoint_type", clusterEp.CustomEndpointType)

	excludeMembers := make([]string, 0)
	for _, member := range clusterEp.ExcludedMembers {
		excludeMembers = append(excludeMembers, aws.StringValue(member))
	}
	d.Set("excluded_members", excludeMembers)

	staticMembers := make([]string, 0)
	for _, member := range clusterEp.StaticMembers {
		staticMembers = append(staticMembers, aws.StringValue(member))
	}
	d.Set("static_members", staticMembers)

	return nil
}

func resourceAwsRDSClusterEndpointUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	input := &rds.ModifyDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("custom_endpoint_type"); ok {
		input.EndpointType = aws.String(v.(string))
	}

	if attr := d.Get("excluded_members").(*schema.Set); attr.Len() > 0 {
		input.ExcludedMembers = expandStringList(attr.List())
	} else {
		input.ExcludedMembers = make([]*string, 0)
	}

	if attr := d.Get("static_members").(*schema.Set); attr.Len() > 0 {
		input.StaticMembers = expandStringList(attr.List())
	} else {
		input.StaticMembers = make([]*string, 0)
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.ModifyDBClusterEndpoint(input)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error modifying RDS Cluster Endpoint: %s", err)
	}

	return resourceAwsRDSClusterEndpointRead(d, meta)
}

func resourceAwsRDSClusterEndpointDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).rdsconn
	input := &rds.DeleteDBClusterEndpointInput{
		DBClusterEndpointIdentifier: aws.String(d.Id()),
	}
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDBClusterEndpoint(input)
		if err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("Error deleting RDS Cluster Endpoint: %s", err)
	}
	return nil
}
