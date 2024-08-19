// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_elb", name="Classic Load Balancer")
func dataSourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLoadBalancerRead,
		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},

			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			"access_logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrInterval: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						names.AttrBucket: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrBucketPrefix: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			names.AttrAvailabilityZones: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"connection_draining": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"connection_draining_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"cross_zone_load_balancing": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrHealthCheck: {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"healthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"unhealthy_threshold": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						names.AttrTarget: {
							Type:     schema.TypeString,
							Computed: true,
						},

						names.AttrInterval: {
							Type:     schema.TypeInt,
							Computed: true,
						},

						names.AttrTimeout: {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},

			"idle_timeout": {
				Type:     schema.TypeInt,
				Computed: true,
			},

			"instances": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"internal": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"listener": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"instance_protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"lb_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"lb_protocol": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ssl_certificate_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"source_security_group": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrSubnets: {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"desync_mitigation_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrTags: tftags.TagsSchemaComputed(),

			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)
	ec2conn := meta.(*conns.AWSClient).EC2Client(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	lbName := d.Get(names.AttrName).(string)
	lb, err := findLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s): %s", lbName, err)
	}

	d.SetId(aws.ToString(lb.LoadBalancerName))

	input := &elasticloadbalancing.DescribeLoadBalancerAttributesInput{
		LoadBalancerName: aws.String(d.Id()),
	}

	output, err := conn.DescribeLoadBalancerAttributes(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s) attributes: %s", d.Id(), err)
	}

	lbAttrs := output.LoadBalancerAttributes

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "elasticloadbalancing",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("loadbalancer/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set(names.AttrName, lb.LoadBalancerName)
	d.Set(names.AttrDNSName, lb.DNSName)
	d.Set("zone_id", lb.CanonicalHostedZoneNameID)

	var scheme bool
	if lb.Scheme != nil {
		scheme = aws.ToString(lb.Scheme) == "internal"
	}
	d.Set("internal", scheme)
	d.Set(names.AttrAvailabilityZones, flex.FlattenStringValueList(lb.AvailabilityZones))
	d.Set("instances", flattenInstances(lb.Instances))
	d.Set("listener", flattenListenerDescriptions(lb.ListenerDescriptions))
	d.Set(names.AttrSecurityGroups, flex.FlattenStringValueList(lb.SecurityGroups))
	if lb.SourceSecurityGroup != nil {
		group := lb.SourceSecurityGroup.GroupName
		if lb.SourceSecurityGroup.OwnerAlias != nil && aws.ToString(lb.SourceSecurityGroup.OwnerAlias) != "" {
			group = aws.String(aws.ToString(lb.SourceSecurityGroup.OwnerAlias) + "/" + aws.ToString(lb.SourceSecurityGroup.GroupName))
		}
		d.Set("source_security_group", group)

		// Manually look up the ELB Security Group ID, since it's not provided
		var elbVpc string
		if lb.VPCId != nil {
			elbVpc = aws.ToString(lb.VPCId)
			sg, err := tfec2.FindSecurityGroupByNameAndVPCIDAndOwnerID(ctx, ec2conn, aws.ToString(lb.SourceSecurityGroup.GroupName), elbVpc, aws.ToString(lb.SourceSecurityGroup.OwnerAlias))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "looking up ELB Security Group ID: %s", err)
			} else {
				d.Set("source_security_group_id", sg.GroupId)
			}
		}
	}
	d.Set(names.AttrSubnets, flex.FlattenStringValueList(lb.Subnets))
	if lbAttrs.ConnectionSettings != nil {
		d.Set("idle_timeout", lbAttrs.ConnectionSettings.IdleTimeout)
	}
	d.Set("connection_draining", lbAttrs.ConnectionDraining.Enabled)
	d.Set("connection_draining_timeout", lbAttrs.ConnectionDraining.Timeout)
	d.Set("cross_zone_load_balancing", lbAttrs.CrossZoneLoadBalancing.Enabled)
	if lbAttrs.AccessLog != nil {
		// The AWS API does not allow users to remove access_logs, only disable them.
		// During creation of the ELB, Terraform sets the access_logs to disabled,
		// so there should not be a case where lbAttrs.AccessLog above is nil.

		// Here we do not record the remove value of access_log if:
		// - there is no access_log block in the configuration
		// - the remote access_logs are disabled
		//
		// This indicates there is no access_log in the configuration.
		// - externally added access_logs will be enabled, so we'll detect the drift
		// - locally added access_logs will be in the config, so we'll add to the
		// API/state
		// See https://github.com/hashicorp/terraform/issues/10138
		_, n := d.GetChange("access_logs")
		elbal := lbAttrs.AccessLog
		nl := n.([]interface{})
		if len(nl) == 0 && !elbal.Enabled {
			elbal = nil
		}
		if err := d.Set("access_logs", flattenAccessLog(elbal)); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s): setting access_logs: %s", d.Id(), err)
		}
	}

	for _, attr := range lbAttrs.AdditionalAttributes {
		switch aws.ToString(attr.Key) {
		case loadBalancerAttributeDesyncMitigationMode:
			d.Set("desync_mitigation_mode", attr.Value)
		}
	}

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for ELB (%s): %s", d.Id(), err)
	}

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	// There's only one health check, so save that to state as we
	// currently can
	if aws.ToString(lb.HealthCheck.Target) != "" {
		d.Set(names.AttrHealthCheck, flattenHealthCheck(lb.HealthCheck))
	}

	return diags
}
