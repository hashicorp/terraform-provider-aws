package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsEc2ManagedPrefixList() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsEc2ManagedPrefixListRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"entries": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     dataSourceAwsEc2ManagedPrefixListEntrySchema(),
			},
			"filter": dataSourceFiltersSchema(),
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"address_family": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"max_entries": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func dataSourceAwsEc2ManagedPrefixListEntrySchema() *schema.Resource {
	return &schema.Resource{
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
	}
}

func dataSourceAwsEc2ManagedPrefixListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	filters, filtersOk := d.GetOk("filter")

	input := ec2.DescribeManagedPrefixListsInput{}

	if filtersOk {
		input.Filters = buildAwsDataSourceFilters(filters.(*schema.Set))
	}

	if prefixListId, ok := d.GetOk("id"); ok {
		input.PrefixListIds = aws.StringSlice([]string{prefixListId.(string)})
	}

	if prefixListName := d.Get("name"); prefixListName.(string) != "" {
		input.Filters = append(input.Filters, &ec2.Filter{
			Name:   aws.String("prefix-list-name"),
			Values: aws.StringSlice([]string{prefixListName.(string)}),
		})
	}

	out, err := conn.DescribeManagedPrefixListsWithContext(ctx, &input)

	if err != nil {
		return diag.FromErr(err)
	}

	if len(out.PrefixLists) < 1 {
		return diag.Errorf("no managed prefix lists matched the given criteria")
	}

	if len(out.PrefixLists) > 1 {
		return diag.Errorf("more than 1 prefix list matched the given criteria")
	}

	pl := *out.PrefixLists[0]

	d.SetId(aws.StringValue(pl.PrefixListId))
	d.Set("name", pl.PrefixListName)
	d.Set("owner_id", pl.OwnerId)
	d.Set("address_family", pl.AddressFamily)
	d.Set("arn", pl.PrefixListArn)
	d.Set("max_entries", pl.MaxEntries)
	d.Set("version", pl.Version)

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(pl.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(err)
	}

	entries := &schema.Set{
		F: schema.HashResource(dataSourceAwsEc2ManagedPrefixListEntrySchema()),
	}

	err = conn.GetManagedPrefixListEntriesPages(
		&ec2.GetManagedPrefixListEntriesInput{
			PrefixListId: pl.PrefixListId,
		},
		func(output *ec2.GetManagedPrefixListEntriesOutput, last bool) bool {
			for _, entry := range output.Entries {
				entries.Add(map[string]interface{}{
					"cidr":        aws.StringValue(entry.Cidr),
					"description": aws.StringValue(entry.Description),
				})
			}

			return true
		},
	)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("entries", entries); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
