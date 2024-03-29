// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_elb")
func DataSourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLoadBalancerRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"access_logs": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"bucket": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"bucket_prefix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},

			"availability_zones": {
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

			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"health_check": {
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

						"target": {
							Type:     schema.TypeString,
							Computed: true,
						},

						"interval": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"timeout": {
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
				Set: ListenerHash,
			},

			"security_groups": {
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

			"subnets": {
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"desync_mitigation_mode": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchemaComputed(),

			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)
	ec2conn := meta.(*conns.AWSClient).EC2Conn(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	lbName := d.Get("name").(string)
	lb, err := FindLoadBalancerByName(ctx, conn, lbName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s): %s", lbName, err)
	}

	d.SetId(aws.StringValue(lb.LoadBalancerName))

	input := &elb.DescribeLoadBalancerAttributesInput{
		LoadBalancerName: aws.String(d.Id()),
	}

	output, err := conn.DescribeLoadBalancerAttributesWithContext(ctx, input)

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
	d.Set("arn", arn.String())
	d.Set("name", lb.LoadBalancerName)
	d.Set("dns_name", lb.DNSName)
	d.Set("zone_id", lb.CanonicalHostedZoneNameID)

	var scheme bool
	if lb.Scheme != nil {
		scheme = aws.StringValue(lb.Scheme) == "internal"
	}
	d.Set("internal", scheme)
	d.Set("availability_zones", flex.FlattenStringList(lb.AvailabilityZones))
	d.Set("instances", flattenInstances(lb.Instances))
	d.Set("listener", flattenListeners(lb.ListenerDescriptions))
	d.Set("security_groups", flex.FlattenStringList(lb.SecurityGroups))
	if lb.SourceSecurityGroup != nil {
		group := lb.SourceSecurityGroup.GroupName
		if lb.SourceSecurityGroup.OwnerAlias != nil && aws.StringValue(lb.SourceSecurityGroup.OwnerAlias) != "" {
			group = aws.String(aws.StringValue(lb.SourceSecurityGroup.OwnerAlias) + "/" + aws.StringValue(lb.SourceSecurityGroup.GroupName))
		}
		d.Set("source_security_group", group)

		// Manually look up the ELB Security Group ID, since it's not provided
		var elbVpc string
		if lb.VPCId != nil {
			elbVpc = aws.StringValue(lb.VPCId)
			sg, err := tfec2.FindSecurityGroupByNameAndVPCIDAndOwnerID(ctx, ec2conn, aws.StringValue(lb.SourceSecurityGroup.GroupName), elbVpc, aws.StringValue(lb.SourceSecurityGroup.OwnerAlias))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "looking up ELB Security Group ID: %s", err)
			} else {
				d.Set("source_security_group_id", sg.GroupId)
			}
		}
	}
	d.Set("subnets", flex.FlattenStringList(lb.Subnets))
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
		if len(nl) == 0 && !aws.BoolValue(elbal.Enabled) {
			elbal = nil
		}
		if err := d.Set("access_logs", flattenAccessLog(elbal)); err != nil {
			return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s): setting access_logs: %s", d.Id(), err)
		}
	}

	for _, attr := range lbAttrs.AdditionalAttributes {
		switch aws.StringValue(attr.Key) {
		case "elb.http.desyncmitigationmode":
			d.Set("desync_mitigation_mode", attr.Value)
		}
	}

	tags, err := listTags(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for ELB (%s): %s", d.Id(), err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	// There's only one health check, so save that to state as we
	// currently can
	if aws.StringValue(lb.HealthCheck.Target) != "" {
		d.Set("health_check", FlattenHealthCheck(lb.HealthCheck))
	}

	return diags
}
