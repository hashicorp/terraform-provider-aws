// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_route53_zone", name="Hosted Zone")
func dataSourceZone() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceZoneRead,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrComment: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"linked_service_description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"linked_service_principal": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name_servers": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"primary_name_server": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"private_zone": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"resource_record_set_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func dataSourceZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name := d.Get(names.AttrName).(string)
	zoneID, zoneIDExists := d.GetOk("zone_id")
	vpcID, vpcIDExists := d.GetOk(names.AttrVPCID)
	tags := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})).IgnoreAWS()

	input := &route53.ListHostedZonesInput{}
	var hostedZones []awstypes.HostedZone
	pages := route53.NewListHostedZonesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zones: %s", err)
		}

		for _, hostedZone := range page.HostedZones {
			hostedZoneID := cleanZoneID(aws.ToString(hostedZone.Id))
			if zoneIDExists && hostedZoneID == zoneID.(string) {
				hostedZones = append(hostedZones, hostedZone)
				// we check if the name is the same as requested and if private zone field is the same as requested or if there is a vpc_id
			} else if (normalizeZoneName(aws.ToString(hostedZone.Name)) == normalizeZoneName(name)) && (hostedZone.Config.PrivateZone == d.Get("private_zone").(bool) || (hostedZone.Config.PrivateZone && vpcIDExists)) {
				matchingVPC := false
				if vpcIDExists {
					hostedZone, err := findHostedZoneByID(ctx, conn, hostedZoneID)

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zone (%s): %s", hostedZoneID, err)
					}

					for _, v := range hostedZone.VPCs {
						if aws.ToString(v.VPCId) == vpcID.(string) {
							matchingVPC = true
							break
						}
					}
				} else {
					matchingVPC = true
				}

				matchingTags := true
				if len(tags) > 0 {
					output, err := listTags(ctx, conn, hostedZoneID, string(awstypes.TagResourceTypeHostedzone))

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "listing Route 53 Hosted Zone (%s) tags: %s", hostedZoneID, err)
					}

					matchingTags = output.ContainsAll(tags)
				}

				if matchingTags && matchingVPC {
					hostedZones = append(hostedZones, hostedZone)
				}
			}
		}
	}

	hostedZone, err := tfresource.AssertSingleValueResult(hostedZones)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Route 53 Hosted Zone", err))
	}

	hostedZoneID := cleanZoneID(aws.ToString(hostedZone.Id))
	d.SetId(hostedZoneID)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  "hostedzone/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("caller_reference", hostedZone.CallerReference)
	d.Set(names.AttrComment, hostedZone.Config.Comment)
	if hostedZone.LinkedService != nil {
		d.Set("linked_service_description", hostedZone.LinkedService.Description)
		d.Set("linked_service_principal", hostedZone.LinkedService.ServicePrincipal)
	}
	// To be consistent with other AWS services (e.g. ACM) that do not accept a trailing period,
	// we remove the suffix from the Hosted Zone Name returned from the API
	d.Set(names.AttrName, normalizeZoneName(aws.ToString(hostedZone.Name)))
	d.Set("private_zone", hostedZone.Config.PrivateZone)
	d.Set("resource_record_set_count", hostedZone.ResourceRecordSetCount)
	d.Set("zone_id", hostedZoneID)

	nameServers, err := hostedZoneNameServers(ctx, conn, hostedZoneID, aws.ToString(hostedZone.Name))

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	d.Set("name_servers", nameServers)
	d.Set("primary_name_server", nameServers[0])

	tags, err = listTags(ctx, conn, hostedZoneID, string(awstypes.TagResourceTypeHostedzone))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Route 53 Hosted Zone (%s) tags: %s", hostedZoneID, err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}

func hostedZoneNameServers(ctx context.Context, conn *route53.Client, zoneID, zoneName string) ([]string, error) {
	output, err := findHostedZoneByID(ctx, conn, zoneID)

	if err != nil {
		return nil, fmt.Errorf("reading Route53 Hosted Zone (%s): %w", zoneID, err)
	}

	var nameServers []string

	if output.DelegationSet != nil {
		nameServers = output.DelegationSet.NameServers
	}

	if output.HostedZone.Config != nil && output.HostedZone.Config.PrivateZone {
		nameServers, err = findNameServersByZone(ctx, conn, zoneID, zoneName)

		if err != nil {
			return nil, fmt.Errorf("reading Route53 Hosted Zone (%s) name servers: %w", zoneID, err)
		}
	}

	return nameServers, nil
}
