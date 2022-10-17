package globalaccelerator

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceAccelerator() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceAcceleratorRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"attributes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flow_logs_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"flow_logs_s3_bucket": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"flow_logs_s3_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"hosted_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_address_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_addresses": {
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"ip_family": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceAcceleratorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).GlobalAcceleratorConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var results []*globalaccelerator.Accelerator
	err := conn.ListAcceleratorsPagesWithContext(ctx, &globalaccelerator.ListAcceleratorsInput{}, func(page *globalaccelerator.ListAcceleratorsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, accelerator := range page.Accelerators {
			if accelerator == nil {
				continue
			}

			if v, ok := d.GetOk("arn"); ok && v.(string) != aws.StringValue(accelerator.AcceleratorArn) {
				continue
			}

			if v, ok := d.GetOk("name"); ok && v.(string) != aws.StringValue(accelerator.Name) {
				continue
			}

			results = append(results, accelerator)
		}

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("listing Global Accelerator Accelerators: %s", err)
	}

	if n := len(results); n == 0 {
		return diag.Errorf("no matching Global Accelerator Accelerator found")
	} else if n > 1 {
		return diag.Errorf("multiple Global Accelerator Accelerators matched; use additional constraints to reduce matches to a single Global Accelerator Accelerator")
	}

	accelerator := results[0]
	arn := aws.StringValue(accelerator.AcceleratorArn)
	d.SetId(arn)
	d.Set("arn", arn)
	d.Set("dns_name", accelerator.DnsName)
	d.Set("enabled", accelerator.Enabled)
	d.Set("hosted_zone_id", route53ZoneID)
	d.Set("ip_address_type", accelerator.IpAddressType)
	if err := d.Set("ip_sets", flattenIPSets(accelerator.IpSets)); err != nil {
		return diag.Errorf("setting ip_sets: %s", err)
	}
	d.Set("name", accelerator.Name)

	acceleratorAttributes, err := FindAcceleratorAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("reading Global Accelerator Accelerator (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set("attributes", []interface{}{flattenAcceleratorAttributes(acceleratorAttributes)}); err != nil {
		return diag.Errorf("setting attributes: %s", err)
	}

	tags, err := ListTagsWithContext(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("listing tags for Global Accelerator Accelerator (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
