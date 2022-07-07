package ec2

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func DataSourcePrefixList() *schema.Resource {
	return &schema.Resource{
		Read: dataSourcePrefixListRead,

		Schema: map[string]*schema.Schema{
			"cidr_blocks": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"filter": CustomFiltersSchema(),
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"prefix_list_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourcePrefixListRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DescribePrefixListsInput{}

	if v, ok := d.GetOk("name"); ok {
		input.Filters = append(input.Filters, BuildAttributeFilterList(map[string]string{
			"prefix-list-name": v.(string),
		})...)
	}

	if v, ok := d.GetOk("prefix_list_id"); ok {
		input.PrefixListIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)

	pl, err := FindPrefixList(conn, input)

	if err != nil {
		return tfresource.SingularDataSourceFindError("EC2 Prefix List", err)
	}

	d.SetId(aws.StringValue(pl.PrefixListId))
	d.Set("cidr_blocks", aws.StringValueSlice(pl.Cidrs))
	d.Set("name", pl.PrefixListName)

	return nil
}
