package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func DataSourcePublicIpv4Pools() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePublicIpv4PoolsRead,
		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"pool_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"pools": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{Type: schema.TypeString},
				},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNamePublicIpv4Pools = "Public IPv4 Pools Data Source"
)

func dataSourcePublicIpv4PoolsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribePublicIpv4PoolsInput{}

	if v, ok := d.GetOk("pool_ids"); ok {
		input.PoolIds = aws.StringSlice([]string{v.(string)})
	}

	input.Filters = append(input.Filters, BuildTagFilterList(
		Tags(tftags.New(d.Get("tags").(map[string]interface{}))),
	)...)

	input.Filters = append(input.Filters, BuildFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)...)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	publicIpv4Pools := []map[string]interface{}{}

	output, err := FindPublicIpv4Pools(ctx, conn, input)
	if err != nil {
		create.DiagError(names.EC2, create.ErrActionSetting, DSNamePublicIpv4Pools, d.Id(), err)
	}

	for _, v := range output {
		pool := flattenPublicIpv4Pool(v)
		publicIpv4Pools = append(publicIpv4Pools, pool)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("pools", publicIpv4Pools)

	return nil
}

func flattenPublicIpv4Pool(pool *ec2.PublicIpv4Pool) map[string]interface{} {
	if pool == nil {
		return map[string]interface{}{}
	}

	m := map[string]interface{}{
		"description":                   aws.StringValue(pool.Description),
		"network_border_group":          aws.StringValue(pool.NetworkBorderGroup),
		"pool_address_ranges":           flattenPublicIpv4PoolRanges(pool.PoolAddressRanges),
		"pool_id":                       aws.StringValue(pool.PoolId),
		"tags":                          flattenTags(pool.Tags),
		"total_address_count":           aws.Int64Value(pool.TotalAddressCount),
		"total_available_address_count": aws.Int64Value(pool.TotalAvailableAddressCount),
	}

	return m
}

func flattenPublicIpv4PoolRanges(pool_ranges []*ec2.PublicIpv4PoolRange) []interface{} {
	result := []interface{}{}

	if pool_ranges == nil {
		return result
	}

	for _, v := range pool_ranges {
		range_map := map[string]interface{}{
			"address_count":           aws.Int64Value(v.AddressCount),
			"available_address_count": aws.Int64Value(v.AvailableAddressCount),
			"first_address":           aws.StringValue(v.FirstAddress),
			"last_address":            aws.StringValue(v.LastAddress),
		}
		result = append(result, range_map)
	}

	return result
}

func flattenTags(tags []*ec2.Tag) map[string]string {
	result := make(map[string]string)
	for _, t := range tags {
		result[aws.StringValue(t.Key)] = aws.StringValue(t.Value)
	}

	return result
}
