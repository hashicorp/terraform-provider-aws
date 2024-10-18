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
	var (
		diags               diag.Diagnostics
		zoneID, name, vpcID string
		privateZone         bool
	)

	conn := meta.(*conns.AWSClient).Route53Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	zoneIDArg, zoneIDSet := d.GetOk("zone_id")
	if zoneIDSet {
		zoneID = zoneIDArg.(string)
	}

	nameArg, nameSet := d.GetOk(names.AttrName)
	if nameSet {
		name = nameArg.(string)
	}

	if zoneIDSet && nameSet {
		return sdkdiag.AppendErrorf(diags, "only one of `zone_id` or `name` may be set")
	}

	vpcIDArg, vpcIDSet := d.GetOk(names.AttrVPCID)
	if vpcIDSet {
		vpcID = vpcIDArg.(string)
		privateZone = true
	}

	privateZoneArg, privateZoneSet := d.GetOk("private_zone")
	if privateZoneSet {
		privateZone = privateZoneArg.(bool)
	}

	if vpcIDSet && !privateZone {
		return sdkdiag.AppendErrorf(diags, "`vpc_id` can only be set for private zones")
	}

	tags := tftags.New(ctx, d.Get(names.AttrTags).(map[string]interface{})).IgnoreAWS()

	var hostedZone *awstypes.HostedZone

	if zoneIDSet {
		// Perform direct lookup on unique zoneID
		foundZone, err := findHostedZoneByID(ctx, conn, zoneID)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
		hostedZone = foundZone.HostedZone
	} else {
		// As name is not unique, we need to list all zones and filter
		var hostedZones []awstypes.HostedZone
		input := &route53.ListHostedZonesInput{}
		pages := route53.NewListHostedZonesPaginator(conn, input)
		for pages.HasMorePages() {
			page, err := pages.NextPage(ctx)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zones: %s", err)
			}

			for _, zone := range page.HostedZones {
				// skip zone on explicit name mismatch
				if nameSet && (normalizeZoneName(aws.ToString(zone.Name)) != normalizeZoneName(name)) {
					continue
				}

				// skip zone on type mismatch
				if zone.Config.PrivateZone != privateZone {
					continue
				}

				zoneID := cleanZoneID(aws.ToString(zone.Id))

				matchingVPC := false
				if vpcIDSet {
					zoneDetails, err := findHostedZoneByID(ctx, conn, zoneID)
					if err != nil {
						return sdkdiag.AppendErrorf(diags, "reading Route 53 Hosted Zone (%s): %s", zoneID, err)
					}

					for _, v := range zoneDetails.VPCs {
						if aws.ToString(v.VPCId) == vpcID {
							matchingVPC = true
							break
						}
					}
				} else {
					matchingVPC = true
				}

				matchingTags := true
				if len(tags) > 0 {
					zoneTags, err := listTags(ctx, conn, zoneID, string(awstypes.TagResourceTypeHostedzone))
					if err != nil {
						return sdkdiag.AppendErrorf(diags, "listing Route 53 Hosted Zone (%s) tags: %s", zoneID, err)
					}

					matchingTags = zoneTags.ContainsAll(tags)
				}

				if matchingTags && matchingVPC {
					hostedZones = append(hostedZones, zone)
				}
			}
		}
		var err error
		hostedZone, err = tfresource.AssertSingleValueResult(hostedZones)
		if err != nil {
			return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Route 53 Hosted Zone", err))
		}
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
