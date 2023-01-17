package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceManagedPrefixLists() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceManagedPrefixListsRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceManagedPrefixListsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn()

	input := &ec2.DescribeManagedPrefixListsInput{}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		input.Filters = nil
	}

	prefixLists, err := FindManagedPrefixLists(ctx, conn, input)

	if err != nil {
		return diag.Errorf("reading EC2 Managed Prefix Lists: %s", err)
	}

	var prefixListIDs []string

	for _, v := range prefixLists {
		prefixListIDs = append(prefixListIDs, aws.StringValue(v.PrefixListId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", prefixListIDs)

	return nil
}
