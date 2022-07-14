package ec2

import (
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceSecurityGroupRead,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"filter": CustomFiltersSchema(),

			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchemaComputed(),

			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeSecurityGroupsInput{}

	if id, ok := d.GetOk("id"); ok {
		req.GroupIds = []*string{aws.String(id.(string))}
	}

	req.Filters = BuildAttributeFilterList(
		map[string]string{
			"group-name": d.Get("name").(string),
			"vpc-id":     d.Get("vpc_id").(string),
		},
	)
	req.Filters = append(req.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)
	req.Filters = append(req.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	sg, err := FindSecurityGroup(conn, req)
	if errors.Is(err, tfresource.ErrEmptyResult) {
		return fmt.Errorf("no matching SecurityGroup found")
	}
	if errors.Is(err, tfresource.ErrTooManyResults) {
		return fmt.Errorf("multiple Security Groups matched; use additional constraints to reduce matches to a single Security Group")
	}
	if err != nil {
		return err
	}

	d.SetId(aws.StringValue(sg.GroupId))
	d.Set("name", sg.GroupName)
	d.Set("description", sg.Description)
	d.Set("vpc_id", sg.VpcId)

	if err := d.Set("tags", KeyValueTags(sg.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: *sg.OwnerId,
		Resource:  fmt.Sprintf("security-group/%s", *sg.GroupId),
	}.String()
	d.Set("arn", arn)

	return nil
}
