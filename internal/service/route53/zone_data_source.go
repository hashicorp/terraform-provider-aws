// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_route53_zone")
func DataSourceZone() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceZoneRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"caller_reference": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"comment": {
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
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"name_servers": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
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
				Optional: true,
				Computed: true,
			},
			"tags": tftags.TagsSchemaComputed(),
			"vpc_id": {
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
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	name, nameExists := d.GetOk("name")
	name = name.(string)
	id, idExists := d.GetOk("zone_id")
	vpcId, vpcIdExists := d.GetOk("vpc_id")
	tags := tftags.New(ctx, d.Get("tags").(map[string]interface{})).IgnoreAWS()

	if nameExists && idExists {
		return sdkdiag.AppendErrorf(diags, "zone_id and name arguments can't be used together")
	}

	if !nameExists && !idExists {
		return sdkdiag.AppendErrorf(diags, "Either name or zone_id must be set")
	}

	var nextMarker *string

	var hostedZoneFound *route53.HostedZone
	// We loop through all hostedzone
	for allHostedZoneListed := false; !allHostedZoneListed; {
		req := &route53.ListHostedZonesInput{}
		if nextMarker != nil {
			req.Marker = nextMarker
		}
		log.Printf("[DEBUG] Reading Route53 Zone: %s", req)
		resp, err := conn.ListHostedZonesWithContext(ctx, req)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "finding Route 53 Hosted Zone: %s", err)
		}
		for _, hostedZone := range resp.HostedZones {
			hostedZoneId := CleanZoneID(aws.StringValue(hostedZone.Id))
			if idExists && hostedZoneId == id.(string) {
				hostedZoneFound = hostedZone
				break
				// we check if the name is the same as requested and if private zone field is the same as requested or if there is a vpc_id
			} else if (TrimTrailingPeriod(aws.StringValue(hostedZone.Name)) == TrimTrailingPeriod(name)) && (aws.BoolValue(hostedZone.Config.PrivateZone) == d.Get("private_zone").(bool) || (aws.BoolValue(hostedZone.Config.PrivateZone) && vpcIdExists)) {
				matchingVPC := false
				if vpcIdExists {
					reqHostedZone := &route53.GetHostedZoneInput{}
					reqHostedZone.Id = aws.String(hostedZoneId)

					respHostedZone, errHostedZone := conn.GetHostedZoneWithContext(ctx, reqHostedZone)
					if errHostedZone != nil {
						return sdkdiag.AppendErrorf(diags, "finding Route 53 Hosted Zone: %s", errHostedZone)
					}
					// we go through all VPCs
					for _, vpc := range respHostedZone.VPCs {
						if aws.StringValue(vpc.VPCId) == vpcId.(string) {
							matchingVPC = true
							break
						}
					}
				} else {
					matchingVPC = true
				}
				// we check if tags match
				matchingTags := true
				if len(tags) > 0 {
					listTags, err := listTags(ctx, conn, hostedZoneId, route53.TagResourceTypeHostedzone)

					if err != nil {
						return sdkdiag.AppendErrorf(diags, "finding Route 53 Hosted Zone: %s", err)
					}
					matchingTags = listTags.ContainsAll(tags)
				}

				if matchingTags && matchingVPC {
					if hostedZoneFound != nil {
						return sdkdiag.AppendErrorf(diags, "multiple Route53Zone found please use vpc_id option to filter")
					}

					hostedZoneFound = hostedZone
				}
			}
		}
		if *resp.IsTruncated {
			nextMarker = resp.NextMarker
		} else {
			allHostedZoneListed = true
		}
	}
	if hostedZoneFound == nil {
		return sdkdiag.AppendErrorf(diags, "no matching Route53Zone found")
	}

	idHostedZone := CleanZoneID(aws.StringValue(hostedZoneFound.Id))
	d.SetId(idHostedZone)
	d.Set("zone_id", idHostedZone)
	// To be consistent with other AWS services (e.g. ACM) that do not accept a trailing period,
	// we remove the suffix from the Hosted Zone Name returned from the API
	d.Set("name", TrimTrailingPeriod(aws.StringValue(hostedZoneFound.Name)))
	d.Set("comment", hostedZoneFound.Config.Comment)
	d.Set("private_zone", hostedZoneFound.Config.PrivateZone)
	d.Set("caller_reference", hostedZoneFound.CallerReference)
	d.Set("resource_record_set_count", hostedZoneFound.ResourceRecordSetCount)
	if hostedZoneFound.LinkedService != nil {
		d.Set("linked_service_principal", hostedZoneFound.LinkedService.ServicePrincipal)
		d.Set("linked_service_description", hostedZoneFound.LinkedService.Description)
	}

	nameServers, err := hostedZoneNameServers(ctx, conn, idHostedZone, aws.StringValue(hostedZoneFound.Name))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Route 53 Hosted Zone (%s) name servers: %s", idHostedZone, err)
	}

	if err := d.Set("primary_name_server", nameServers[0]); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting primary_name_server: %s", err)
	}

	if err := d.Set("name_servers", nameServers); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting name_servers: %s", err)
	}

	tags, err = listTags(ctx, conn, idHostedZone, route53.TagResourceTypeHostedzone)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Route 53 Hosted Zone (%s) tags: %s", idHostedZone, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("hostedzone/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return diags
}

// used to retrieve name servers
func hostedZoneNameServers(ctx context.Context, conn *route53.Route53, id string, name string) ([]string, error) {
	input := &route53.GetHostedZoneInput{
		Id: aws.String(id),
	}

	output, err := conn.GetHostedZoneWithContext(ctx, input)

	if err != nil {
		return nil, fmt.Errorf("getting Route 53 Hosted Zone (%s): %w", id, err)
	}

	if output == nil {
		return nil, fmt.Errorf("getting Route 53 Hosted Zone (%s): empty response", id)
	}

	if output.DelegationSet != nil {
		return aws.StringValueSlice(output.DelegationSet.NameServers), nil
	}

	if output.HostedZone != nil && output.HostedZone.Config != nil && aws.BoolValue(output.HostedZone.Config.PrivateZone) {
		nameServers, err := findNameServers(ctx, conn, id, name)

		if err != nil {
			return nil, fmt.Errorf("listing Route 53 Hosted Zone (%s) NS records: %w", id, err)
		}

		return nameServers, nil
	}

	return nil, nil
}
