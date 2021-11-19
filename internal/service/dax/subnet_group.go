package dax

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
)

func ResourceSubnetGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceSubnetGroupCreate,
		Read:   resourceSubnetGroupRead,
		Update: resourceSubnetGroupUpdate,
		Delete: resourceSubnetGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSubnetGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DAXConn

	input := &dax.CreateSubnetGroupInput{
		SubnetGroupName: aws.String(d.Get("name").(string)),
		SubnetIds:       flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set)),
	}
	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	_, err := conn.CreateSubnetGroup(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string))
	return resourceSubnetGroupRead(d, meta)
}

func resourceSubnetGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DAXConn

	resp, err := conn.DescribeSubnetGroups(&dax.DescribeSubnetGroupsInput{
		SubnetGroupNames: []*string{aws.String(d.Id())},
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, dax.ErrCodeSubnetGroupNotFoundFault, "") {
			log.Printf("[WARN] DAX SubnetGroup %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	sg := resp.SubnetGroups[0]

	d.Set("name", sg.SubnetGroupName)
	d.Set("description", sg.Description)
	subnetIDs := make([]*string, 0, len(sg.Subnets))
	for _, v := range sg.Subnets {
		subnetIDs = append(subnetIDs, v.SubnetIdentifier)
	}
	d.Set("subnet_ids", flex.FlattenStringList(subnetIDs))
	d.Set("vpc_id", sg.VpcId)
	return nil
}

func resourceSubnetGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DAXConn

	input := &dax.UpdateSubnetGroupInput{
		SubnetGroupName: aws.String(d.Id()),
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("subnet_ids") {
		input.SubnetIds = flex.ExpandStringSet(d.Get("subnet_ids").(*schema.Set))
	}

	_, err := conn.UpdateSubnetGroup(input)
	if err != nil {
		return err
	}

	return resourceSubnetGroupRead(d, meta)
}

func resourceSubnetGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DAXConn

	input := &dax.DeleteSubnetGroupInput{
		SubnetGroupName: aws.String(d.Id()),
	}

	_, err := conn.DeleteSubnetGroup(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, dax.ErrCodeSubnetGroupNotFoundFault, "") {
			return nil
		}
		return err
	}

	return nil
}
