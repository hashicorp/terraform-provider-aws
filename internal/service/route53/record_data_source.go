// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"regexp"
	"strconv"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_route53_record", name="Record")
func DataSourceRecords() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRecordRead,

		Schema: map[string]*schema.Schema{
			"zone_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter_record": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"record_sets": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"alias": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"failover": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"geolocation": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"geoproximity": {
							Type:     schema.TypeMap,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"latency_region": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"multivalue_answer": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"set_identifier": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ttl": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"values": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"weight": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"routing_policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var (
		diags  diag.Diagnostics
		filter *regexp.Regexp
		output []types.ResourceRecordSet
	)
	conn := meta.(*conns.AWSClient).Route53Client(ctx)
	zoneid := d.Get("zone_id").(string)
	input := &route53.ListResourceRecordSetsInput{
		HostedZoneId: aws.String(zoneid),
	}
	if v, ok := d.GetOk("filter_record"); ok {
		filter = regexache.MustCompile(v.(string))
	}
	err := listrecordset(conn, input, &output)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting the record sets in (%s): %s", zoneid, err)
	}
	d.SetId(*aws.String(zoneid))
	if err1 := d.Set("record_sets", flattenrecordset(output, filter)); err1 != nil {
		return sdkdiag.AppendErrorf(diags, "setting recordsets: %s", err1)
	}
	return diags
}

func listrecordset(conn *route53.Client, input *route53.ListResourceRecordSetsInput, output *[]types.ResourceRecordSet) error {
	re, err := conn.ListResourceRecordSets(context.TODO(), input)
	if err != nil {
		return err
	}
	*output = append(*output, re.ResourceRecordSets...)
	if re.IsTruncated {
		input := &route53.ListResourceRecordSetsInput{
			HostedZoneId:    input.HostedZoneId,
			StartRecordName: aws.String(*re.NextRecordName),
			StartRecordType: re.NextRecordType,
		}
		listrecordset(conn, input, output)
	}
	return nil
}

func flattenrecordset(config []types.ResourceRecordSet, filter *regexp.Regexp) interface{} {
	if config == nil {
		return []interface{}{}
	}
	y := []interface{}{}
	if len(config) > 0 {
		for _, f := range config {
			if filter == nil {
				y = append(y, extractRecordSet(f))
			} else {
				if filter.MatchString(*f.Name) {
					y = append(y, extractRecordSet(f))
				}
			}
		}
	}
	return interface{}(y)
}

func extractRecordSet(data types.ResourceRecordSet) map[string]interface{} {
	var val []string
	c := map[string]interface{}{
		"name": data.Name,
		"ttl":  data.TTL,
		"type": data.Type,
	}
	if data.Failover != types.ResourceRecordSetFailover("") {
		c["failover"] = data.Failover
		c["routing_policy"] = "failover_routing_policy"
	}
	if flattenGeolocation(data.GeoLocation) != nil {
		c["geolocation"] = flattenGeolocation(data.GeoLocation)
		c["routing_policy"] = "geolocation_routing_policy"
	}
	if flattenGeoProximity(data.GeoProximityLocation) != nil {
		c["geoproximity"] = flattenGeoProximity(data.GeoProximityLocation)
		c["routing_policy"] = "geoproximity_routing_policy"
	}
	if data.Region != types.ResourceRecordSetRegion("") {
		c["latency_region"] = data.Region
		c["routing_policy"] = "latency_routing_policy"
	}
	if data.MultiValueAnswer != nil {
		c["multivalue_answer"] = data.MultiValueAnswer
		c["routing_policy"] = "multivalue_routing_policy"
	}
	if data.Weight != nil {
		c["weight"] = data.Weight
		c["routing_policy"] = "weighted_routing_policy"
	}
	if data.SetIdentifier != nil {
		c["set_identifier"] = data.SetIdentifier
	}
	if data.AliasTarget != nil {
		c["alias"] = flattenAlias(data.AliasTarget)
	}
	if data.Failover == types.ResourceRecordSetFailover("") && flattenGeolocation(data.GeoLocation) == nil && data.Region == types.ResourceRecordSetRegion("") && data.MultiValueAnswer == nil && data.Weight == nil {
		c["routing_policy"] = "simple_routing_policy"
	}
	for _, k := range data.ResourceRecords {
		val = append(val, *k.Value)
	}
	c["values"] = aws.StringSlice(val)
	return c
}

func flattenGeolocation(geo *types.GeoLocation) map[string]interface{} {
	if geo == nil {
		return nil
	}
	c := map[string]interface{}{
		"continent_code":   geo.ContinentCode,
		"country_code":     geo.CountryCode,
		"subdivision_code": geo.SubdivisionCode,
	}
	return c
}

func flattenGeoProximity(geoproxi *types.GeoProximityLocation) map[string]interface{} {
	if geoproxi == nil {
		return nil
	}
	c := map[string]interface{}{
		"aws_region":       geoproxi.AWSRegion,
		"bias":             aws.String(strconv.Itoa(int(aws.ToInt32(geoproxi.Bias)))),
		"local_zone_group": geoproxi.LocalZoneGroup,
	}
	coordinates := flattenCoordinate(geoproxi.Coordinates)
	c["latitude"] = coordinates[0].(map[string]interface{})["latitude"].(string)
	c["longitude"] = coordinates[0].(map[string]interface{})["longitude"].(string)
	return c
}

func flattenAlias(alias *types.AliasTarget) map[string]interface{} {
	if alias == nil {
		return nil
	}
	c := map[string]interface{}{
		"dns_name":               alias.DNSName,
		"hostedzone_id":          alias.HostedZoneId,
		"evaluate_target_health": alias.EvaluateTargetHealth,
	}
	return c
}
