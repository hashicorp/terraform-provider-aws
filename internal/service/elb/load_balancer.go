// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"bytes"
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_elb", name="Classic Load Balancer")
// @Tags(identifierAttribute="id")
func ResourceLoadBalancer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLoadBalancerCreate,
		ReadWithoutTimeout:   resourceLoadBalancerRead,
		UpdateWithoutTimeout: resourceLoadBalancerUpdate,
		DeleteWithoutTimeout: resourceLoadBalancerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.All(
			customdiff.ForceNewIfChange("subnets", func(_ context.Context, o, n, meta interface{}) bool {
				// Force new if removing all current subnets.
				os := o.(*schema.Set)
				ns := n.(*schema.Set)

				removed := os.Difference(ns)

				return removed.Equal(os)
			}),
			verify.SetTagsDiff,
		),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Update: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_logs": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bucket_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"interval": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      60,
							ValidateFunc: ValidAccessLogsInterval,
						},
					},
				},
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zones": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"connection_draining": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"connection_draining_timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  300,
			},
			"cross_zone_load_balancing": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"desync_mitigation_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "defensive",
				ValidateFunc: validation.StringInSlice([]string{
					"monitor",
					"defensive",
					"strictest",
				}, false),
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health_check": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"healthy_threshold": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(2, 10),
						},
						"interval": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(5, 300),
						},
						"target": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: ValidHeathCheckTarget,
						},
						"timeout": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(2, 60),
						},
						"unhealthy_threshold": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(2, 10),
						},
					},
				},
			},
			"idle_timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      60,
				ValidateFunc: validation.IntBetween(1, 4000),
			},
			"instances": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"internal": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"listener": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"instance_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"instance_protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateListenerProtocol(),
						},
						"lb_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 65535),
						},
						"lb_protocol": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateListenerProtocol(),
						},
						"ssl_certificate_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
				Set: ListenerHash,
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  ValidName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validNamePrefix,
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"source_security_group": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"source_security_group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"subnets": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	var elbName string
	if v, ok := d.GetOk("name"); ok {
		elbName = v.(string)
	} else {
		if v, ok := d.GetOk("name_prefix"); ok {
			elbName = id.PrefixedUniqueId(v.(string))
		} else {
			elbName = id.PrefixedUniqueId("tf-lb-")
		}
	}

	listeners, err := ExpandListeners(d.Get("listener").(*schema.Set).List())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &elb.CreateLoadBalancerInput{
		LoadBalancerName: aws.String(elbName),
		Listeners:        listeners,
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("availability_zones"); ok && v.(*schema.Set).Len() > 0 {
		input.AvailabilityZones = flex.ExpandStringSet(v.(*schema.Set))
	}

	if _, ok := d.GetOk("internal"); ok {
		input.Scheme = aws.String("internal")
	}

	if v, ok := d.GetOk("security_groups"); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnets"); ok && v.(*schema.Set).Len() > 0 {
		input.Subnets = flex.ExpandStringSet(v.(*schema.Set))
	}

	_, err = tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateLoadBalancerWithContext(ctx, input)
	}, elb.ErrCodeCertificateNotFoundException)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELB Classic Load Balancer (%s): %s", elbName, err)
	}

	d.SetId(elbName)

	return append(diags, resourceLoadBalancerUpdate(ctx, d, meta)...)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	lb, err := FindLoadBalancerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELB Classic Load Balancer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s): %s", d.Id(), err)
	}

	lbAttrs, err := findLoadBalancerAttributesByName(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s) attributes: %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Service:   "elasticloadbalancing",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("loadbalancer/%s", d.Id()),
	}
	d.Set("arn", arn.String())
	d.Set("availability_zones", flex.FlattenStringList(lb.AvailabilityZones))
	d.Set("connection_draining", lbAttrs.ConnectionDraining.Enabled)
	d.Set("connection_draining_timeout", lbAttrs.ConnectionDraining.Timeout)
	d.Set("cross_zone_load_balancing", lbAttrs.CrossZoneLoadBalancing.Enabled)
	d.Set("dns_name", lb.DNSName)
	if lbAttrs.ConnectionSettings != nil {
		d.Set("idle_timeout", lbAttrs.ConnectionSettings.IdleTimeout)
	}
	d.Set("instances", flattenInstances(lb.Instances))
	var scheme bool
	if lb.Scheme != nil {
		scheme = aws.StringValue(lb.Scheme) == "internal"
	}
	d.Set("internal", scheme)
	d.Set("listener", flattenListeners(lb.ListenerDescriptions))
	d.Set("name", lb.LoadBalancerName)
	d.Set("security_groups", flex.FlattenStringList(lb.SecurityGroups))
	d.Set("subnets", flex.FlattenStringList(lb.Subnets))
	d.Set("zone_id", lb.CanonicalHostedZoneNameID)

	if lb.SourceSecurityGroup != nil {
		group := lb.SourceSecurityGroup.GroupName
		if v := aws.StringValue(lb.SourceSecurityGroup.OwnerAlias); v != "" {
			group = aws.String(v + "/" + aws.StringValue(lb.SourceSecurityGroup.GroupName))
		}
		d.Set("source_security_group", group)

		// Manually look up the ELB Security Group ID, since it's not provided
		if lb.VPCId != nil {
			sg, err := tfec2.FindSecurityGroupByNameAndVPCIDAndOwnerID(ctx, meta.(*conns.AWSClient).EC2Conn(ctx), aws.StringValue(lb.SourceSecurityGroup.GroupName), aws.StringValue(lb.VPCId), aws.StringValue(lb.SourceSecurityGroup.OwnerAlias))
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "reading ELB Classic Load Balancer (%s) security group: %s", d.Id(), err)
			} else {
				d.Set("source_security_group_id", sg.GroupId)
			}
		}
	}

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
			return sdkdiag.AppendErrorf(diags, "setting access_logs: %s", err)
		}
	}

	for _, attr := range lbAttrs.AdditionalAttributes {
		switch aws.StringValue(attr.Key) {
		case "elb.http.desyncmitigationmode":
			d.Set("desync_mitigation_mode", attr.Value)
		}
	}

	// There's only one health check, so save that to state as we
	// currently can
	if aws.StringValue(lb.HealthCheck.Target) != "" {
		if err := d.Set("health_check", FlattenHealthCheck(lb.HealthCheck)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting health_check: %s", err)
		}
	}

	return diags
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	if d.HasChange("listener") {
		o, n := d.GetChange("listener")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		remove, _ := ExpandListeners(os.Difference(ns).List())
		add, err := ExpandListeners(ns.Difference(os).List())

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if len(remove) > 0 {
			ports := make([]*int64, 0, len(remove))
			for _, listener := range remove {
				ports = append(ports, listener.LoadBalancerPort)
			}

			input := &elb.DeleteLoadBalancerListenersInput{
				LoadBalancerName:  aws.String(d.Id()),
				LoadBalancerPorts: ports,
			}

			_, err := conn.DeleteLoadBalancerListenersWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Load Balancer (%s) listeners: %s", d.Id(), err)
			}
		}

		if len(add) > 0 {
			input := &elb.CreateLoadBalancerListenersInput{
				Listeners:        add,
				LoadBalancerName: aws.String(d.Id()),
			}

			// Occasionally AWS will error with a 'duplicate listener', without any
			// other listeners on the ELB. Retry here to eliminate that.
			_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutUpdate),
				func() (interface{}, error) {
					return conn.CreateLoadBalancerListenersWithContext(ctx, input)
				},
				func(err error) (bool, error) {
					if tfawserr.ErrCodeEquals(err, elb.ErrCodeDuplicateListenerException) {
						return true, err
					}
					if tfawserr.ErrMessageContains(err, elb.ErrCodeCertificateNotFoundException, "Server Certificate not found for the key: arn") {
						return true, err
					}

					return false, err
				})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "creating ELB Classic Load Balancer (%s) listeners: %s", d.Id(), err)
			}
		}
	}

	// If we currently have instances, or did have instances,
	// we want to figure out what to add and remove from the load
	// balancer
	if d.HasChange("instances") {
		o, n := d.GetChange("instances")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)
		remove := ExpandInstanceString(os.Difference(ns).List())
		add := ExpandInstanceString(ns.Difference(os).List())

		if len(add) > 0 {
			input := &elb.RegisterInstancesWithLoadBalancerInput{
				Instances:        add,
				LoadBalancerName: aws.String(d.Id()),
			}

			_, err := conn.RegisterInstancesWithLoadBalancerWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "registering ELB Classic Load Balancer (%s) instances: %s", d.Id(), err)
			}
		}

		if len(remove) > 0 {
			input := &elb.DeregisterInstancesFromLoadBalancerInput{
				Instances:        remove,
				LoadBalancerName: aws.String(d.Id()),
			}

			_, err := conn.DeregisterInstancesFromLoadBalancerWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deregistering ELB Classic Load Balancer (%s) instances: %s", d.Id(), err)
			}
		}
	}

	if d.HasChanges("cross_zone_load_balancing", "idle_timeout", "access_logs", "desync_mitigation_mode") {
		input := &elb.ModifyLoadBalancerAttributesInput{
			LoadBalancerAttributes: &elb.LoadBalancerAttributes{
				AdditionalAttributes: []*elb.AdditionalAttribute{
					{
						Key:   aws.String("elb.http.desyncmitigationmode"),
						Value: aws.String(d.Get("desync_mitigation_mode").(string)),
					},
				},
				CrossZoneLoadBalancing: &elb.CrossZoneLoadBalancing{
					Enabled: aws.Bool(d.Get("cross_zone_load_balancing").(bool)),
				},
				ConnectionSettings: &elb.ConnectionSettings{
					IdleTimeout: aws.Int64(int64(d.Get("idle_timeout").(int))),
				},
			},
			LoadBalancerName: aws.String(d.Id()),
		}

		if logs := d.Get("access_logs").([]interface{}); len(logs) == 1 {
			l := logs[0].(map[string]interface{})
			input.LoadBalancerAttributes.AccessLog = &elb.AccessLog{
				Enabled:        aws.Bool(l["enabled"].(bool)),
				EmitInterval:   aws.Int64(int64(l["interval"].(int))),
				S3BucketName:   aws.String(l["bucket"].(string)),
				S3BucketPrefix: aws.String(l["bucket_prefix"].(string)),
			}
		} else if len(logs) == 0 {
			// disable access logs
			input.LoadBalancerAttributes.AccessLog = &elb.AccessLog{
				Enabled: aws.Bool(false),
			}
		}

		_, err := conn.ModifyLoadBalancerAttributesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELB Classic Load Balancer (%s) attributes: %s", d.Id(), err)
		}
	}

	// We have to do these changes separately from everything else since
	// they have some weird undocumented rules. You can't set the timeout
	// without having connection draining to true, so we set that to true,
	// set the timeout, then reset it to false if requested.
	if d.HasChanges("connection_draining", "connection_draining_timeout") {
		// We do timeout changes first since they require us to set draining
		// to true for a hot second.
		if d.HasChange("connection_draining_timeout") {
			input := &elb.ModifyLoadBalancerAttributesInput{
				LoadBalancerAttributes: &elb.LoadBalancerAttributes{
					ConnectionDraining: &elb.ConnectionDraining{
						Enabled: aws.Bool(true),
						Timeout: aws.Int64(int64(d.Get("connection_draining_timeout").(int))),
					},
				},
				LoadBalancerName: aws.String(d.Id()),
			}

			_, err := conn.ModifyLoadBalancerAttributesWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying ELB Classic Load Balancer (%s) attributes: %s", d.Id(), err)
			}
		}

		// Then we always set connection draining even if there is no change.
		// This lets us reset to "false" if requested even with a timeout
		// change.
		input := &elb.ModifyLoadBalancerAttributesInput{
			LoadBalancerAttributes: &elb.LoadBalancerAttributes{
				ConnectionDraining: &elb.ConnectionDraining{
					Enabled: aws.Bool(d.Get("connection_draining").(bool)),
				},
			},
			LoadBalancerName: aws.String(d.Id()),
		}

		_, err := conn.ModifyLoadBalancerAttributesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELB Classic Load Balancer (%s) attributes: %s", d.Id(), err)
		}
	}

	if d.HasChange("health_check") {
		if hc := d.Get("health_check").([]interface{}); len(hc) > 0 {
			check := hc[0].(map[string]interface{})
			input := &elb.ConfigureHealthCheckInput{
				HealthCheck: &elb.HealthCheck{
					HealthyThreshold:   aws.Int64(int64(check["healthy_threshold"].(int))),
					Interval:           aws.Int64(int64(check["interval"].(int))),
					Target:             aws.String(check["target"].(string)),
					Timeout:            aws.Int64(int64(check["timeout"].(int))),
					UnhealthyThreshold: aws.Int64(int64(check["unhealthy_threshold"].(int))),
				},
				LoadBalancerName: aws.String(d.Id()),
			}
			_, err := conn.ConfigureHealthCheckWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "configuring ELB Classic Load Balancer (%s) health check: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("security_groups") {
		input := &elb.ApplySecurityGroupsToLoadBalancerInput{
			LoadBalancerName: aws.String(d.Id()),
			SecurityGroups:   flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
		}

		_, err := conn.ApplySecurityGroupsToLoadBalancerWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "applying ELB Classic Load Balancer (%s) security groups: %s", d.Id(), err)
		}
	}

	if d.HasChange("availability_zones") {
		o, n := d.GetChange("availability_zones")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removed := flex.ExpandStringSet(os.Difference(ns))
		added := flex.ExpandStringSet(ns.Difference(os))

		if len(added) > 0 {
			input := &elb.EnableAvailabilityZonesForLoadBalancerInput{
				AvailabilityZones: added,
				LoadBalancerName:  aws.String(d.Id()),
			}

			_, err := conn.EnableAvailabilityZonesForLoadBalancerWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling ELB Classic Load Balancer (%s) Availability Zones: %s", d.Id(), err)
			}
		}

		if len(removed) > 0 {
			input := &elb.DisableAvailabilityZonesForLoadBalancerInput{
				AvailabilityZones: removed,
				LoadBalancerName:  aws.String(d.Id()),
			}

			_, err := conn.DisableAvailabilityZonesForLoadBalancerWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling ELB Classic Load Balancer (%s) Availability Zones: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange("subnets") {
		o, n := d.GetChange("subnets")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removed := flex.ExpandStringSet(os.Difference(ns))
		added := flex.ExpandStringSet(ns.Difference(os))

		if len(removed) > 0 {
			input := &elb.DetachLoadBalancerFromSubnetsInput{
				LoadBalancerName: aws.String(d.Id()),
				Subnets:          removed,
			}

			_, err := conn.DetachLoadBalancerFromSubnetsWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "detaching ELB Classic Load Balancer (%s) from subnets: %s", d.Id(), err)
			}
		}

		if len(added) > 0 {
			input := &elb.AttachLoadBalancerToSubnetsInput{
				LoadBalancerName: aws.String(d.Id()),
				Subnets:          added,
			}

			_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.AttachLoadBalancerToSubnetsWithContext(ctx, input)
			}, elb.ErrCodeInvalidConfigurationRequestException, "cannot be attached to multiple subnets in the same AZ")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "attaching ELB Classic Load Balancer (%s) to subnets: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceLoadBalancerRead(ctx, d, meta)...)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBConn(ctx)

	log.Printf("[INFO] Deleting ELB Classic Load Balancer: %s", d.Id())
	_, err := conn.DeleteLoadBalancerWithContext(ctx, &elb.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Load Balancer (%s): %s", d.Id(), err)
	}

	err = deleteNetworkInterfaces(ctx, meta.(*conns.AWSClient).EC2Conn(ctx), d.Id())

	if err != nil {
		diags = sdkdiag.AppendWarningf(diags, "cleaning up ELB Classic Load Balancer (%s) ENIs: %s", d.Id(), err)
	}

	return diags
}

func FindLoadBalancerByName(ctx context.Context, conn *elb.ELB, name string) (*elb.LoadBalancerDescription, error) {
	input := &elb.DescribeLoadBalancersInput{
		LoadBalancerNames: aws.StringSlice([]string{name}),
	}

	output, err := conn.DescribeLoadBalancersWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.LoadBalancerDescriptions) == 0 || output.LoadBalancerDescriptions[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.LoadBalancerDescriptions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	// Eventual consistency check.
	if aws.StringValue(output.LoadBalancerDescriptions[0].LoadBalancerName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output.LoadBalancerDescriptions[0], nil
}

func findLoadBalancerAttributesByName(ctx context.Context, conn *elb.ELB, name string) (*elb.LoadBalancerAttributes, error) {
	input := &elb.DescribeLoadBalancerAttributesInput{
		LoadBalancerName: aws.String(name),
	}

	output, err := conn.DescribeLoadBalancerAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, elb.ErrCodeAccessPointNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LoadBalancerAttributes == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LoadBalancerAttributes, nil
}

func ListenerHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["instance_port"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["instance_protocol"].(string))))
	buf.WriteString(fmt.Sprintf("%d-", m["lb_port"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["lb_protocol"].(string))))

	if v, ok := m["ssl_certificate_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	return create.StringHashcode(buf.String())
}

func ValidAccessLogsInterval(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)

	// Check if the value is either 5 or 60 (minutes).
	if value != 5 && value != 60 {
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid Access Logs interval \"%d\". "+
				"Valid intervals are either 5 or 60 (minutes).",
			k, value))
	}
	return
}

func ValidHeathCheckTarget(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	// Parse the Health Check target value.
	matches := regexp.MustCompile(`\A(\w+):(\d+)(.+)?\z`).FindStringSubmatch(value)

	// Check if the value contains a valid target.
	if matches == nil || len(matches) < 1 {
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid Health Check: %s",
			k, value))

		// Invalid target? Return immediately,
		// there is no need to collect other
		// errors.
		return ws, errors
	}

	// Check if the value contains a valid protocol.
	if !isValidProtocol(matches[1]) {
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid Health Check protocol %q. "+
				"Valid protocols are either %q, %q, %q, or %q.",
			k, matches[1], "TCP", "SSL", "HTTP", "HTTPS"))
	}

	// Check if the value contains a valid port range.
	port, _ := strconv.Atoi(matches[2])
	if port < 1 || port > 65535 {
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid Health Check target port \"%d\". "+
				"Valid port is in the range from 1 to 65535 inclusive.",
			k, port))
	}

	switch strings.ToLower(matches[1]) {
	case "tcp", "ssl":
		// Check if value is in the form <PROTOCOL>:<PORT> for TCP and/or SSL.
		if matches[3] != "" {
			errors = append(errors, fmt.Errorf(
				"%q cannot contain a path in the Health Check target: %s",
				k, value))
		}

	case "http", "https":
		// Check if value is in the form <PROTOCOL>:<PORT>/<PATH> for HTTP and/or HTTPS.
		if matches[3] == "" {
			errors = append(errors, fmt.Errorf(
				"%q must contain a path in the Health Check target: %s",
				k, value))
		}

		// Cannot be longer than 1024 multibyte characters.
		if len([]rune(matches[3])) > 1024 {
			errors = append(errors, fmt.Errorf("%q cannot contain a path longer "+
				"than 1024 characters in the Health Check target: %s",
				k, value))
		}
	}

	return ws, errors
}

func isValidProtocol(s string) bool {
	if s == "" {
		return false
	}
	s = strings.ToLower(s)

	validProtocols := map[string]bool{
		"http":  true,
		"https": true,
		"ssl":   true,
		"tcp":   true,
	}

	if _, ok := validProtocols[s]; !ok {
		return false
	}

	return true
}

func validateListenerProtocol() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"HTTP",
		"HTTPS",
		"SSL",
		"TCP",
	}, true)
}

// ELB automatically creates ENI(s) on creation
// but the cleanup is asynchronous and may take time
// which then blocks IGW, SG or VPC on deletion
// So we make the cleanup "synchronous" here
func deleteNetworkInterfaces(ctx context.Context, conn *ec2.EC2, name string) error {
	// https://aws.amazon.com/premiumsupport/knowledge-center/elb-find-load-balancer-IP/.
	networkInterfaces, err := tfec2.FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription(ctx, conn, "amazon-elb", "ELB "+name)

	if err != nil {
		return err
	}

	var errs *multierror.Error

	for _, networkInterface := range networkInterfaces {
		if networkInterface.Attachment == nil {
			continue
		}

		attachmentID := aws.StringValue(networkInterface.Attachment.AttachmentId)
		networkInterfaceID := aws.StringValue(networkInterface.NetworkInterfaceId)

		err = tfec2.DetachNetworkInterface(ctx, conn, networkInterfaceID, attachmentID, tfec2.NetworkInterfaceDetachedTimeout)

		if err != nil {
			errs = multierror.Append(errs, err)

			continue
		}

		err = tfec2.DeleteNetworkInterface(ctx, conn, networkInterfaceID)

		if err != nil {
			errs = multierror.Append(errs, err)

			continue
		}
	}

	return errs.ErrorOrNil()
}
