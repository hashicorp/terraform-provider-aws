package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/neptune"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsNeptuneSubnetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsNeptuneSubnetGroupCreate,
		Read:   resourceAwsNeptuneSubnetGroupRead,
		Update: resourceAwsNeptuneSubnetGroupUpdate,
		Delete: resourceAwsNeptuneSubnetGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validateNeptuneSubnetGroupName,
			},
			"name_prefix": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validateNeptuneSubnetGroupNamePrefix,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},

			"subnet_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsNeptuneSubnetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn
	tags := tagsFromMapNeptune(d.Get("tags").(map[string]interface{}))

	subnetIdsSet := d.Get("subnet_ids").(*schema.Set)
	subnetIds := make([]*string, subnetIdsSet.Len())
	for i, subnetId := range subnetIdsSet.List() {
		subnetIds[i] = aws.String(subnetId.(string))
	}

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = resource.PrefixedUniqueId(v.(string))
	} else {
		groupName = resource.UniqueId()
	}

	createOpts := neptune.CreateDBSubnetGroupInput{
		DBSubnetGroupName:        aws.String(groupName),
		DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
		SubnetIds:                subnetIds,
		Tags:                     tags,
	}

	log.Printf("[DEBUG] Create Neptune Subnet Group: %#v", createOpts)
	_, err := conn.CreateDBSubnetGroup(&createOpts)
	if err != nil {
		return fmt.Errorf("Error creating Neptune Subnet Group: %s", err)
	}

	d.SetId(*createOpts.DBSubnetGroupName)
	log.Printf("[INFO] Neptune Subnet Group ID: %s", d.Id())
	return resourceAwsNeptuneSubnetGroupRead(d, meta)
}

func resourceAwsNeptuneSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn

	describeOpts := neptune.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(d.Id()),
	}

	describeResp, err := conn.DescribeDBSubnetGroups(&describeOpts)
	if err != nil {
		if neptuneerr, ok := err.(awserr.Error); ok && neptuneerr.Code() == "DBSubnetGroupNotFoundFault" {
			// Update state to indicate the neptune subnet no longer exists.
			d.SetId("")
			return nil
		}
		return err
	}

	if len(describeResp.DBSubnetGroups) == 0 {
		return fmt.Errorf("Unable to find Neptune Subnet Group: %#v", describeResp.DBSubnetGroups)
	}

	var subnetGroup *neptune.DBSubnetGroup
	for _, s := range describeResp.DBSubnetGroups {
		// AWS is down casing the name provided, so we compare lower case versions
		// of the names. We lower case both our name and their name in the check,
		// incase they change that someday.
		if strings.ToLower(d.Id()) == strings.ToLower(*s.DBSubnetGroupName) {
			subnetGroup = describeResp.DBSubnetGroups[0]
		}
	}

	if subnetGroup.DBSubnetGroupName == nil {
		return fmt.Errorf("Unable to find Neptune Subnet Group: %#v", describeResp.DBSubnetGroups)
	}

	d.Set("name", subnetGroup.DBSubnetGroupName)
	d.Set("description", subnetGroup.DBSubnetGroupDescription)

	subnets := make([]string, 0, len(subnetGroup.Subnets))
	for _, s := range subnetGroup.Subnets {
		subnets = append(subnets, *s.SubnetIdentifier)
	}
	d.Set("subnet_ids", subnets)

	// list tags for resource
	// set tags

	//Amazon Neptune shares the format of Amazon RDS ARNs. Neptune ARNs contain rds and not neptune.
	//https://docs.aws.amazon.com/neptune/latest/userguide/tagging.ARN.html
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "rds",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("subgrp:%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	resp, err := conn.ListTagsForResource(&neptune.ListTagsForResourceInput{
		ResourceName: aws.String(arn),
	})

	if err != nil {
		log.Printf("[DEBUG] Error retreiving tags for ARN: %s", arn)
	}

	var dt []*neptune.Tag
	if len(resp.TagList) > 0 {
		dt = resp.TagList
	}
	d.Set("tags", tagsToMapNeptune(dt))

	return nil
}

func resourceAwsNeptuneSubnetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).neptuneconn
	if d.HasChange("subnet_ids") || d.HasChange("description") {
		_, n := d.GetChange("subnet_ids")
		if n == nil {
			n = new(schema.Set)
		}
		ns := n.(*schema.Set)

		var sIds []*string
		for _, s := range ns.List() {
			sIds = append(sIds, aws.String(s.(string)))
		}

		_, err := conn.ModifyDBSubnetGroup(&neptune.ModifyDBSubnetGroupInput{
			DBSubnetGroupName:        aws.String(d.Id()),
			DBSubnetGroupDescription: aws.String(d.Get("description").(string)),
			SubnetIds:                sIds,
		})

		if err != nil {
			return err
		}
	}

	//Amazon Neptune shares the format of Amazon RDS ARNs. Neptune ARNs contain rds and not neptune.
	//https://docs.aws.amazon.com/neptune/latest/userguide/tagging.ARN.html
	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "rds",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("subgrp:%s", d.Id()),
	}.String()
	if err := setTagsNeptune(conn, d, arn); err != nil {
		return err
	} else {
		d.SetPartial("tags")
	}

	return resourceAwsNeptuneSubnetGroupRead(d, meta)
}

func resourceAwsNeptuneSubnetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"destroyed"},
		Refresh:    resourceAwsNeptuneSubnetGroupDeleteRefreshFunc(d, meta),
		Timeout:    3 * time.Minute,
		MinTimeout: 1 * time.Second,
	}
	_, err := stateConf.WaitForState()
	return err
}

func resourceAwsNeptuneSubnetGroupDeleteRefreshFunc(
	d *schema.ResourceData,
	meta interface{}) resource.StateRefreshFunc {
	conn := meta.(*AWSClient).neptuneconn

	return func() (interface{}, string, error) {

		deleteOpts := neptune.DeleteDBSubnetGroupInput{
			DBSubnetGroupName: aws.String(d.Id()),
		}

		if _, err := conn.DeleteDBSubnetGroup(&deleteOpts); err != nil {
			neptuneerr, ok := err.(awserr.Error)
			if !ok {
				return d, "error", err
			}

			if neptuneerr.Code() != "DBSubnetGroupNotFoundFault" {
				return d, "error", err
			}
		}

		return d, "destroyed", nil
	}
}
