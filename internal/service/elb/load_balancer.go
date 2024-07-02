// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elb

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
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
			customdiff.ForceNewIfChange(names.AttrSubnets, func(_ context.Context, o, n, meta interface{}) bool {
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
						names.AttrBucket: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrBucketPrefix: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						names.AttrInterval: {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      60,
							ValidateFunc: ValidAccessLogsInterval,
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZones: {
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
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrHealthCheck: {
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
						names.AttrInterval: {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(5, 300),
						},
						names.AttrTarget: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: ValidHeathCheckTarget,
						},
						names.AttrTimeout: {
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
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  ValidName,
			},
			names.AttrNamePrefix: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validNamePrefix,
			},
			names.AttrSecurityGroups: {
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
			names.AttrSubnets: {
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
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	listeners, err := ExpandListeners(d.Get("listener").(*schema.Set).List())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	elbName := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrName).(string)),
		create.WithConfiguredPrefix(d.Get(names.AttrNamePrefix).(string)),
		create.WithDefaultPrefix("tf-lb-"),
	).Generate()
	input := &elasticloadbalancing.CreateLoadBalancerInput{
		Listeners:        listeners,
		LoadBalancerName: aws.String(elbName),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZones); ok && v.(*schema.Set).Len() > 0 {
		input.AvailabilityZones = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if _, ok := d.GetOk("internal"); ok {
		input.Scheme = aws.String("internal")
	}

	if v, ok := d.GetOk(names.AttrSecurityGroups); ok && v.(*schema.Set).Len() > 0 {
		input.SecurityGroups = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrSubnets); ok && v.(*schema.Set).Len() > 0 {
		input.Subnets = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	_, err = tfresource.RetryWhenIsA[*awstypes.CertificateNotFoundException](ctx, d.Timeout(schema.TimeoutCreate), func() (interface{}, error) {
		return conn.CreateLoadBalancer(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELB Classic Load Balancer (%s): %s", elbName, err)
	}

	d.SetId(elbName)

	return append(diags, resourceLoadBalancerUpdate(ctx, d, meta)...)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

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
	d.Set(names.AttrARN, arn.String())
	d.Set(names.AttrAvailabilityZones, flex.FlattenStringValueList(lb.AvailabilityZones))
	d.Set("connection_draining", lbAttrs.ConnectionDraining.Enabled)
	d.Set("connection_draining_timeout", lbAttrs.ConnectionDraining.Timeout)
	d.Set("cross_zone_load_balancing", lbAttrs.CrossZoneLoadBalancing.Enabled)
	d.Set(names.AttrDNSName, lb.DNSName)
	if lbAttrs.ConnectionSettings != nil {
		d.Set("idle_timeout", lbAttrs.ConnectionSettings.IdleTimeout)
	}
	d.Set("instances", flattenInstances(lb.Instances))
	var scheme bool
	if lb.Scheme != nil {
		scheme = aws.ToString(lb.Scheme) == "internal"
	}
	d.Set("internal", scheme)
	d.Set("listener", flattenListeners(lb.ListenerDescriptions))
	d.Set(names.AttrName, lb.LoadBalancerName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(lb.LoadBalancerName)))
	d.Set(names.AttrSecurityGroups, flex.FlattenStringValueList(lb.SecurityGroups))
	d.Set(names.AttrSubnets, flex.FlattenStringValueList(lb.Subnets))
	d.Set("zone_id", lb.CanonicalHostedZoneNameID)

	if lb.SourceSecurityGroup != nil {
		group := lb.SourceSecurityGroup.GroupName
		if v := aws.ToString(lb.SourceSecurityGroup.OwnerAlias); v != "" {
			group = aws.String(v + "/" + aws.ToString(lb.SourceSecurityGroup.GroupName))
		}
		d.Set("source_security_group", group)

		// Manually look up the ELB Security Group ID, since it's not provided
		if lb.VPCId != nil {
			sg, err := tfec2.FindSecurityGroupByNameAndVPCIDAndOwnerID(ctx, meta.(*conns.AWSClient).EC2Conn(ctx), aws.ToString(lb.SourceSecurityGroup.GroupName), aws.ToString(lb.VPCId), aws.ToString(lb.SourceSecurityGroup.OwnerAlias))
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
		if len(nl) == 0 && !elbal.Enabled {
			elbal = nil
		}
		if err := d.Set("access_logs", flattenAccessLog(elbal)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting access_logs: %s", err)
		}
	}

	for _, attr := range lbAttrs.AdditionalAttributes {
		switch aws.ToString(attr.Key) {
		case "elasticloadbalancing.http.desyncmitigationmode":
			d.Set("desync_mitigation_mode", attr.Value)
		}
	}

	// There's only one health check, so save that to state as we
	// currently can
	if aws.ToString(lb.HealthCheck.Target) != "" {
		if err := d.Set(names.AttrHealthCheck, FlattenHealthCheck(lb.HealthCheck)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting health_check: %s", err)
		}
	}

	return diags
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

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
			ports := make([]int32, 0, len(remove))
			for _, listener := range remove {
				ports = append(ports, listener.LoadBalancerPort)
			}

			input := &elasticloadbalancing.DeleteLoadBalancerListenersInput{
				LoadBalancerName:  aws.String(d.Id()),
				LoadBalancerPorts: ports,
			}

			_, err := conn.DeleteLoadBalancerListeners(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Load Balancer (%s) listeners: %s", d.Id(), err)
			}
		}

		if len(add) > 0 {
			input := &elasticloadbalancing.CreateLoadBalancerListenersInput{
				Listeners:        add,
				LoadBalancerName: aws.String(d.Id()),
			}

			// Occasionally AWS will error with a 'duplicate listener', without any
			// other listeners on the ELB. Retry here to eliminate that.
			_, err := tfresource.RetryWhen(ctx, d.Timeout(schema.TimeoutUpdate),
				func() (interface{}, error) {
					return conn.CreateLoadBalancerListeners(ctx, input)
				},
				func(err error) (bool, error) {
					if errs.IsA[*awstypes.DuplicateListenerException](err) {
						return true, err
					}
					if errs.IsAErrorMessageContains[*awstypes.CertificateNotFoundException](err, "Server Certificate not found for the key: arn") {
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
			input := &elasticloadbalancing.RegisterInstancesWithLoadBalancerInput{
				Instances:        add,
				LoadBalancerName: aws.String(d.Id()),
			}

			_, err := conn.RegisterInstancesWithLoadBalancer(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "registering ELB Classic Load Balancer (%s) instances: %s", d.Id(), err)
			}
		}

		if len(remove) > 0 {
			input := &elasticloadbalancing.DeregisterInstancesFromLoadBalancerInput{
				Instances:        remove,
				LoadBalancerName: aws.String(d.Id()),
			}

			_, err := conn.DeregisterInstancesFromLoadBalancer(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deregistering ELB Classic Load Balancer (%s) instances: %s", d.Id(), err)
			}
		}
	}

	if d.HasChanges("cross_zone_load_balancing", "idle_timeout", "access_logs", "desync_mitigation_mode") {
		input := &elasticloadbalancing.ModifyLoadBalancerAttributesInput{
			LoadBalancerAttributes: &awstypes.LoadBalancerAttributes{
				AdditionalAttributes: []awstypes.AdditionalAttribute{
					{
						Key:   aws.String("elasticloadbalancing.http.desyncmitigationmode"),
						Value: aws.String(d.Get("desync_mitigation_mode").(string)),
					},
				},
				CrossZoneLoadBalancing: &awstypes.CrossZoneLoadBalancing{
					Enabled: d.Get("cross_zone_load_balancing").(bool),
				},
				ConnectionSettings: &awstypes.ConnectionSettings{
					IdleTimeout: aws.Int32(int32(d.Get("idle_timeout").(int))),
				},
			},
			LoadBalancerName: aws.String(d.Id()),
		}

		if logs := d.Get("access_logs").([]interface{}); len(logs) == 1 {
			l := logs[0].(map[string]interface{})
			input.LoadBalancerAttributes.AccessLog = &awstypes.AccessLog{
				Enabled:        l[names.AttrEnabled].(bool),
				EmitInterval:   aws.Int32(int32(l[names.AttrInterval].(int))),
				S3BucketName:   aws.String(l[names.AttrBucket].(string)),
				S3BucketPrefix: aws.String(l[names.AttrBucketPrefix].(string)),
			}
		} else if len(logs) == 0 {
			// disable access logs
			input.LoadBalancerAttributes.AccessLog = &awstypes.AccessLog{
				Enabled: false,
			}
		}

		_, err := conn.ModifyLoadBalancerAttributes(ctx, input)

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
			input := &elasticloadbalancing.ModifyLoadBalancerAttributesInput{
				LoadBalancerAttributes: &awstypes.LoadBalancerAttributes{
					ConnectionDraining: &awstypes.ConnectionDraining{
						Enabled: true,
						Timeout: aws.Int32(int32(d.Get("connection_draining_timeout").(int))),
					},
				},
				LoadBalancerName: aws.String(d.Id()),
			}

			_, err := conn.ModifyLoadBalancerAttributes(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying ELB Classic Load Balancer (%s) attributes: %s", d.Id(), err)
			}
		}

		// Then we always set connection draining even if there is no change.
		// This lets us reset to "false" if requested even with a timeout
		// change.
		input := &elasticloadbalancing.ModifyLoadBalancerAttributesInput{
			LoadBalancerAttributes: &awstypes.LoadBalancerAttributes{
				ConnectionDraining: &awstypes.ConnectionDraining{
					Enabled: d.Get("connection_draining").(bool),
				},
			},
			LoadBalancerName: aws.String(d.Id()),
		}

		_, err := conn.ModifyLoadBalancerAttributes(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELB Classic Load Balancer (%s) attributes: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrHealthCheck) {
		if hc := d.Get(names.AttrHealthCheck).([]interface{}); len(hc) > 0 {
			check := hc[0].(map[string]interface{})
			input := &elasticloadbalancing.ConfigureHealthCheckInput{
				HealthCheck: &awstypes.HealthCheck{
					HealthyThreshold:   aws.Int32(int32(check["healthy_threshold"].(int))),
					Interval:           aws.Int32(int32(check[names.AttrInterval].(int))),
					Target:             aws.String(check[names.AttrTarget].(string)),
					Timeout:            aws.Int32(int32(check[names.AttrTimeout].(int))),
					UnhealthyThreshold: aws.Int32(int32(check["unhealthy_threshold"].(int))),
				},
				LoadBalancerName: aws.String(d.Id()),
			}
			_, err := conn.ConfigureHealthCheck(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "configuring ELB Classic Load Balancer (%s) health check: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange(names.AttrSecurityGroups) {
		input := &elasticloadbalancing.ApplySecurityGroupsToLoadBalancerInput{
			LoadBalancerName: aws.String(d.Id()),
			SecurityGroups:   flex.ExpandStringValueSet(d.Get(names.AttrSecurityGroups).(*schema.Set)),
		}

		_, err := conn.ApplySecurityGroupsToLoadBalancer(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "applying ELB Classic Load Balancer (%s) security groups: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrAvailabilityZones) {
		o, n := d.GetChange(names.AttrAvailabilityZones)
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removed := flex.ExpandStringValueSet(os.Difference(ns))
		added := flex.ExpandStringValueSet(ns.Difference(os))

		if len(added) > 0 {
			input := &elasticloadbalancing.EnableAvailabilityZonesForLoadBalancerInput{
				AvailabilityZones: added,
				LoadBalancerName:  aws.String(d.Id()),
			}

			_, err := conn.EnableAvailabilityZonesForLoadBalancer(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling ELB Classic Load Balancer (%s) Availability Zones: %s", d.Id(), err)
			}
		}

		if len(removed) > 0 {
			input := &elasticloadbalancing.DisableAvailabilityZonesForLoadBalancerInput{
				AvailabilityZones: removed,
				LoadBalancerName:  aws.String(d.Id()),
			}

			_, err := conn.DisableAvailabilityZonesForLoadBalancer(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "enabling ELB Classic Load Balancer (%s) Availability Zones: %s", d.Id(), err)
			}
		}
	}

	if d.HasChange(names.AttrSubnets) {
		o, n := d.GetChange(names.AttrSubnets)
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		removed := flex.ExpandStringValueSet(os.Difference(ns))
		added := flex.ExpandStringValueSet(ns.Difference(os))

		if len(removed) > 0 {
			input := &elasticloadbalancing.DetachLoadBalancerFromSubnetsInput{
				LoadBalancerName: aws.String(d.Id()),
				Subnets:          removed,
			}

			_, err := conn.DetachLoadBalancerFromSubnets(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "detaching ELB Classic Load Balancer (%s) from subnets: %s", d.Id(), err)
			}
		}

		if len(added) > 0 {
			input := &elasticloadbalancing.AttachLoadBalancerToSubnetsInput{
				LoadBalancerName: aws.String(d.Id()),
				Subnets:          added,
			}

			_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidConfigurationRequestException](ctx, d.Timeout(schema.TimeoutUpdate), func() (interface{}, error) {
				return conn.AttachLoadBalancerToSubnets(ctx, input)
			}, "cannot be attached to multiple subnets in the same AZ")

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "attaching ELB Classic Load Balancer (%s) to subnets: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceLoadBalancerRead(ctx, d, meta)...)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBClient(ctx)

	log.Printf("[INFO] Deleting ELB Classic Load Balancer: %s", d.Id())
	_, err := conn.DeleteLoadBalancer(ctx, &elasticloadbalancing.DeleteLoadBalancerInput{
		LoadBalancerName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELB Classic Load Balancer (%s): %s", d.Id(), err)
	}

	err = deleteNetworkInterfaces(ctx, meta.(*conns.AWSClient).EC2Client(ctx), d.Id())

	if err != nil {
		diags = sdkdiag.AppendWarningf(diags, "cleaning up ELB Classic Load Balancer (%s) ENIs: %s", d.Id(), err)
	}

	return diags
}

func FindLoadBalancerByName(ctx context.Context, conn *elasticloadbalancing.Client, name string) (*awstypes.LoadBalancerDescription, error) {
	input := &elasticloadbalancing.DescribeLoadBalancersInput{
		LoadBalancerNames: []string{name},
	}

	output, err := conn.DescribeLoadBalancers(ctx, input)

	if errs.IsA[*awstypes.AccessPointNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.LoadBalancerDescriptions) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.LoadBalancerDescriptions); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	// Eventual consistency check.
	if aws.ToString(output.LoadBalancerDescriptions[0].LoadBalancerName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return tfresource.AssertSingleValueResult(output.LoadBalancerDescriptions)
}

func findLoadBalancerAttributesByName(ctx context.Context, conn *elasticloadbalancing.Client, name string) (*awstypes.LoadBalancerAttributes, error) {
	input := &elasticloadbalancing.DescribeLoadBalancerAttributesInput{
		LoadBalancerName: aws.String(name),
	}

	output, err := conn.DescribeLoadBalancerAttributes(ctx, input)

	if errs.IsA[*awstypes.AccessPointNotFoundException](err) {
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
	matches := regexache.MustCompile(`\A(\w+):(\d+)(.+)?\z`).FindStringSubmatch(value)

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
func deleteNetworkInterfaces(ctx context.Context, conn *ec2.Client, name string) error {
	// https://aws.amazon.com/premiumsupport/knowledge-center/elb-find-load-balancer-IP/.
	networkInterfaces, err := tfec2.FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescription(ctx, conn, "amazon-elb", "ELB "+name)

	if err != nil {
		return err
	}

	var errs []error

	for _, networkInterface := range networkInterfaces {
		if networkInterface.Attachment == nil {
			continue
		}

		attachmentID := aws.ToString(networkInterface.Attachment.AttachmentId)
		networkInterfaceID := aws.ToString(networkInterface.NetworkInterfaceId)

		if err := tfec2.DetachNetworkInterface(ctx, conn, networkInterfaceID, attachmentID, tfec2.NetworkInterfaceDetachedTimeout); err != nil {
			errs = append(errs, err)

			continue
		}

		if err := tfec2.DeleteNetworkInterface(ctx, conn, networkInterfaceID); err != nil {
			errs = append(errs, err)

			continue
		}
	}

	return errors.Join(errs...)
}
