package ec2

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceLocalGateway() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalGatewayRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

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

			"filter": CustomFiltersSchema(),

			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tftags.TagsSchemaComputed(),

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLocalGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeLocalGatewaysInput{}

	if v, ok := d.GetOk("id"); ok {
		req.LocalGatewayIds = []*string{aws.String(v.(string))}
	}

	req.Filters = BuildAttributeFilterList(
		map[string]string{
			"state": d.Get("state").(string),
		},
	)

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, BuildTagFilterList(
			Tags(tftags.New(tags.(map[string]interface{}))),
		)...)
	}

	req.Filters = append(req.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading AWS LOCAL GATEWAY: %s", req)
	resp, err := conn.DescribeLocalGatewaysWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing EC2 Local Gateways: %s", err)
	}
	if resp == nil || len(resp.LocalGateways) == 0 {
		return sdkdiag.AppendErrorf(diags, "no matching Local Gateway found")
	}
	if len(resp.LocalGateways) > 1 {
		return sdkdiag.AppendErrorf(diags, "multiple Local Gateways matched; use additional constraints to reduce matches to a single Local Gateway")
	}

	localGateway := resp.LocalGateways[0]

	d.SetId(aws.StringValue(localGateway.LocalGatewayId))
	d.Set("outpost_arn", localGateway.OutpostArn)
	d.Set("owner_id", localGateway.OwnerId)
	d.Set("state", localGateway.State)

	if err := d.Set("tags", KeyValueTags(localGateway.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
