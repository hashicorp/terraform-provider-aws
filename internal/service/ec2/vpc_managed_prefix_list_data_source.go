package ec2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourceManagedPrefixList() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceManagedPrefixListRead,

		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"address_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"entries": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"filter": CustomFiltersSchema(),
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"max_entries": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceManagedPrefixListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeManagedPrefixListsInput{
		Filters: BuildAttributeFilterList(map[string]string{
			"prefix-list-name": d.Get("name").(string),
		}),
	}

	if v, ok := d.GetOk("id"); ok {
		input.PrefixListIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	pl, err := FindManagedPrefixList(ctx, conn, input)

	if err != nil {
		return diag.FromErr(tfresource.SingularDataSourceFindError("EC2 Managed Prefix List", err))
	}

	d.SetId(aws.StringValue(pl.PrefixListId))

	prefixListEntries, err := FindManagedPrefixListEntriesByID(ctx, conn, d.Id())

	if err != nil {
		return diag.Errorf("reading EC2 Managed Prefix List (%s) Entries: %s", d.Id(), err)
	}

	d.Set("address_family", pl.AddressFamily)
	d.Set("arn", pl.PrefixListArn)
	if err := d.Set("entries", flattenPrefixListEntries(prefixListEntries)); err != nil {
		return diag.Errorf("setting entries: %s", err)
	}
	d.Set("max_entries", pl.MaxEntries)
	d.Set("name", pl.PrefixListName)
	d.Set("owner_id", pl.OwnerId)
	d.Set("version", pl.Version)

	if err := d.Set("tags", KeyValueTags(pl.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	return nil
}
