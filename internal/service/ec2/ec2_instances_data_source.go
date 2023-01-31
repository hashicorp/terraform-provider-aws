package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceInstancesRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"instance_tags": tftags.TagsSchemaComputed(),
			"instance_state_names": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(ec2.InstanceStateName_Values(), false),
				},
			},
			"private_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"public_ips": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstancesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.DescribeInstancesInput{}

	if v, ok := d.GetOk("instance_state_names"); ok && v.(*schema.Set).Len() > 0 {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: flex.ExpandStringSet(v.(*schema.Set)),
		})
	} else {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String("instance-state-name"),
			Values: aws.StringSlice([]string{ec2.InstanceStateNameRunning}),
		})
	}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("instance_tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindInstances(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Instances: %s", err)
	}

	var instanceIDs, privateIPs, publicIPs []string

	for _, v := range output {
		instanceIDs = append(instanceIDs, aws.StringValue(v.InstanceId))
		if privateIP := aws.StringValue(v.PrivateIpAddress); privateIP != "" {
			privateIPs = append(privateIPs, privateIP)
		}
		if publicIP := aws.StringValue(v.PublicIpAddress); publicIP != "" {
			publicIPs = append(publicIPs, publicIP)
		}
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", instanceIDs)
	d.Set("private_ips", privateIPs)
	d.Set("public_ips", publicIPs)

	return diags
}
