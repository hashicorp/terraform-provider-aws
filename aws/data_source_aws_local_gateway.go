package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsLocalGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsLocalGatewayRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"filter": ec2CustomFiltersSchema(),

			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tagsSchemaComputed(),

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsLocalGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeLocalGatewaysInput{}

	var id string
	if cid, ok := d.GetOk("id"); ok {
		id = cid.(string)
	}

	if id != "" {
		req.LocalGatewayIds = []*string{aws.String(id)}
	}

	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"state": d.Get("state").(string),
		},
	)

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, buildEC2TagFilterList(
			keyvaluetags.New(tags.(map[string]interface{})).Ec2Tags(),
		)...)
	}

	req.Filters = append(req.Filters, buildEC2CustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading AWS LOCAL GATEWAY: %s", req)
	resp, err := conn.DescribeLocalGateways(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.LocalGateways) == 0 {
		return fmt.Errorf("no matching VPC found")
	}
	if len(resp.LocalGateways) > 1 {
		return fmt.Errorf("multiple Local Gateways matched; use additional constraints to reduce matches to a single VPC")
	}

	localGateway := resp.LocalGateways[0]

	d.SetId(aws.StringValue(localGateway.LocalGatewayId))
	d.Set("outpost_arn", localGateway.OutpostArn)
	d.Set("state", localGateway.State)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(localGateway.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	d.Set("owner_id", localGateway.OwnerId)

	arn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "outposts",
		Region:    meta.(*AWSClient).region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("outpost/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}
