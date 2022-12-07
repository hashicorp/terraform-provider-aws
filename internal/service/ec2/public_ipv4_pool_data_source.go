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

func DataSourcePublicIpv4Pool() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourcePublicIpv4PoolRead,
		Schema: map[string]*schema.Schema{
			"filter": DataSourceFiltersSchema(),
			"pool_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"pool": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNamePublicIpv4Pool = "Public IPv4 Pool Data Source"
)

func dataSourcePublicIpv4PoolRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).EC2Conn
	input := &ec2.DescribePublicIpv4PoolsInput{}

	if v, ok := d.GetOk("pool_id"); ok {
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

	output, err := FindPublicIpv4Pool(ctx, conn, input)
	if err != nil {
		create.DiagError(names.EC2, create.ErrActionSetting, DSNamePublicIpv4Pool, d.Id(), err)
	}

	pool := flattenPublicIpv4Pool(output[0])

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("pool", pool)

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
