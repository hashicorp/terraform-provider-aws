package outposts

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

func DataSourceOutposts() *schema.Resource { // nosemgrep:ci.outposts-in-func-name
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOutpostsRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"site_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceOutpostsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.outposts-in-func-name
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).OutpostsConn()

	input := &outposts.ListOutpostsInput{}

	var arns, ids []string

	err := conn.ListOutpostsPagesWithContext(ctx, input, func(page *outposts.ListOutpostsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, outpost := range page.Outposts {
			if outpost == nil {
				continue
			}

			if v, ok := d.GetOk("availability_zone"); ok && v.(string) != aws.StringValue(outpost.AvailabilityZone) {
				continue
			}

			if v, ok := d.GetOk("availability_zone_id"); ok && v.(string) != aws.StringValue(outpost.AvailabilityZoneId) {
				continue
			}

			if v, ok := d.GetOk("site_id"); ok && v.(string) != aws.StringValue(outpost.SiteId) {
				continue
			}

			if v, ok := d.GetOk("owner_id"); ok && v.(string) != aws.StringValue(outpost.OwnerId) {
				continue
			}

			arns = append(arns, aws.StringValue(outpost.OutpostArn))
			ids = append(ids, aws.StringValue(outpost.OutpostId))
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Outposts Outposts: %s", err)
	}

	if err := d.Set("arns", arns); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arns: %s", err)
	}

	if err := d.Set("ids", ids); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ids: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)

	return diags
}
