package aws

import (
	"fmt"
	"log"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func dataSourceAwsPrefixList() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsPrefixListRead,

		Schema: map[string]*schema.Schema{
			"prefix_list_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
		},
	}
}

func dataSourceAwsPrefixListRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	filters, filtersOk := d.GetOk("filter")

	req := &ec2.DescribeManagedPrefixListsInput{}
	if filtersOk {
		req.Filters = buildAwsDataSourceFilters(filters.(*schema.Set))
	}
	if prefixListID := d.Get("prefix_list_id"); prefixListID != "" {
		req.PrefixListIds = aws.StringSlice([]string{prefixListID.(string)})
	}
	if prefixListName := d.Get("name"); prefixListName.(string) != "" {
		req.Filters = append(req.Filters, &ec2.Filter{
			Name:   aws.String("prefix-list-name"),
			Values: aws.StringSlice([]string{prefixListName.(string)}),
		})
	}

	log.Printf("[DEBUG] Reading Prefix List: %s", req)
	resp, err := conn.DescribeManagedPrefixLists(req)
	switch {
	case err != nil:
		return err
	case resp == nil || len(resp.PrefixLists) == 0:
		return fmt.Errorf("no matching prefix list found; the prefix list ID or name may be invalid or not exist in the current region")
	case len(resp.PrefixLists) > 1:
		return fmt.Errorf("more than one prefix list matched the given set of criteria")
	}

	pl := resp.PrefixLists[0]

	d.SetId(*pl.PrefixListId)
	d.Set("name", pl.PrefixListName)

	getEntriesInput := ec2.GetManagedPrefixListEntriesInput{
		PrefixListId: pl.PrefixListId,
	}

	cidrs := []string(nil)

	err = conn.GetManagedPrefixListEntriesPages(
		&getEntriesInput, func(output *ec2.GetManagedPrefixListEntriesOutput, last bool) bool {
			for _, entry := range output.Entries {
				cidrs = append(cidrs, aws.StringValue(entry.Cidr))
			}
			return true
		})
	if err != nil {
		return fmt.Errorf("failed to get entries of prefix list %s: %s", *pl.PrefixListId, err)
	}

	sort.Strings(cidrs)

	if err := d.Set("cidr_blocks", cidrs); err != nil {
		return fmt.Errorf("failed to set cidr blocks of prefix list %s: %s", d.Id(), err)
	}

	d.Set("owner_id", pl.OwnerId)
	d.Set("address_family", pl.AddressFamily)
	d.Set("arn", pl.PrefixListArn)

	if actual := aws.Int64Value(pl.MaxEntries); actual > 0 {
		d.Set("max_entries", actual)
	}

	if err := d.Set("tags", keyvaluetags.Ec2KeyValueTags(pl.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("failed to set tags of prefix list %s: %s", d.Id(), err)
	}

	return nil
}
