// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import ( // nosemgrep:ci.semgrep.aws.multiple-service-imports

	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_alb", name="Load Balancer")
// @SDKResource("aws_lb", name="Load Balancer")
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

		CustomizeDiff: customdiff.Sequence(
			customizeDiffLoadBalancerALB,
			customizeDiffLoadBalancerNLB,
			customizeDiffLoadBalancerGWLB,
			verify.SetTagsDiff,
		),

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"access_logs": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return !d.Get("access_logs.0.enabled").(bool)
							},
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return !d.Get("access_logs.0.enabled").(bool)
							},
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_keep_alive": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          3600,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},
			"connection_logs": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrBucket: {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return !d.Get("connection_logs.0.enabled").(bool)
							},
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"prefix": {
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return !d.Get("connection_logs.0.enabled").(bool)
							},
						},
					},
				},
			},
			"customer_owned_ipv4_pool": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"desync_mitigation_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          httpDesyncMitigationModeDefensive,
				ValidateFunc:     validation.StringInSlice(httpDesyncMitigationMode_Values(), false),
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},
			"dns_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dns_record_client_routing_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          dnsRecordClientRoutingPolicyAnyAvailabilityZone,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumNetwork),
				ValidateFunc:     validation.StringInSlice(dnsRecordClientRoutingPolicy_Values(), false),
			},
			"drop_invalid_header_fields": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},
			"enable_cross_zone_load_balancing": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBType(elbv2.LoadBalancerTypeEnumApplication),
			},
			"enable_deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_http2": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          true,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},
			"enable_tls_version_and_cipher_suite_headers": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},
			"enable_waf_fail_open": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},
			"enable_xff_client_port": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},
			"enforce_security_group_inbound_rules_on_private_link_traffic": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateFunc:     validation.StringInSlice(elbv2.EnforceSecurityGroupInboundRulesOnPrivateLinkTrafficEnum_Values(), false),
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumNetwork),
			},
			"idle_timeout": {
				Type:             schema.TypeInt,
				Optional:         true,
				Default:          60,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},
			"internal": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"ip_address_type": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(elbv2.IpAddressType_Values(), false),
			},
			"load_balancer_type": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				Default:      elbv2.LoadBalancerTypeEnumApplication,
				ValidateFunc: validation.StringInSlice(elbv2.LoadBalancerTypeEnum_Values(), false),
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validNamePrefix,
			},
			"preserve_host_header": {
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
			},
			"security_groups": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"subnet_mapping": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ipv6_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPv6Address,
						},
						"outpost_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_ipv4_address": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsIPv4Address,
						},
						"subnet_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
				ExactlyOneOf: []string{"subnet_mapping", "subnets"},
			},
			"subnets": {
				Type:         schema.TypeSet,
				Optional:     true,
				Computed:     true,
				Elem:         &schema.Schema{Type: schema.TypeString},
				ExactlyOneOf: []string{"subnet_mapping", "subnets"},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"xff_header_processing_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          httpXFFHeaderProcessingModeAppend,
				DiffSuppressFunc: suppressIfLBTypeNot(elbv2.LoadBalancerTypeEnumApplication),
				ValidateFunc:     validation.StringInSlice(httpXFFHeaderProcessingMode_Values(), false),
			},
			"zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func suppressIfLBType(types ...string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return slices.Contains(types, d.Get("load_balancer_type").(string))
	}
}

func suppressIfLBTypeNot(types ...string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return !slices.Contains(types, d.Get("load_balancer_type").(string))
	}
}

func resourceLoadBalancerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrName).(string)),
		create.WithConfiguredPrefix(d.Get("name_prefix").(string)),
		create.WithDefaultPrefix("tf-lb-"),
	).Generate()
	exist, err := findLoadBalancer(ctx, conn, &elbv2.DescribeLoadBalancersInput{
		Names: aws.StringSlice([]string{name}),
	})

	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Load Balancer (%s): %s", name, err)
	}

	if exist != nil {
		return sdkdiag.AppendErrorf(diags, "ELBv2 Load Balancer (%s) already exists", name)
	}

	d.Set(names.AttrName, name)

	lbType := d.Get("load_balancer_type").(string)
	input := &elbv2.CreateLoadBalancerInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
		Type: aws.String(lbType),
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		input.CustomerOwnedIpv4Pool = aws.String(v.(string))
	}

	if _, ok := d.GetOk("internal"); ok {
		input.Scheme = aws.String(elbv2.LoadBalancerSchemeEnumInternal)
	}

	if v, ok := d.GetOk("ip_address_type"); ok {
		input.IpAddressType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_groups"); ok {
		input.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("subnet_mapping"); ok && v.(*schema.Set).Len() > 0 {
		input.SubnetMappings = expandSubnetMappings(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("subnets"); ok {
		input.Subnets = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.CreateLoadBalancerWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateLoadBalancerWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELBv2 %s Load Balancer (%s): %s", lbType, name, err)
	}

	d.SetId(aws.StringValue(output.LoadBalancers[0].LoadBalancerArn))

	if _, err := waitLoadBalancerActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Load Balancer (%s) create: %s", d.Id(), err)
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceLoadBalancerUpdate(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELBv2 Load Balancer (%s) tags: %s", d.Id(), err)
		}
	}

	var attributes []*elbv2.LoadBalancerAttribute

	if lbType == elbv2.LoadBalancerTypeEnumApplication || lbType == elbv2.LoadBalancerTypeEnumNetwork {
		if v, ok := d.GetOk("access_logs"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandLoadBalancerAccessLogsAttributes(v.([]interface{})[0].(map[string]interface{}), false)...)
		} else {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String(loadBalancerAttributeAccessLogsS3Enabled),
				Value: flex.BoolValueToString(false),
			})
		}
	}

	if lbType == elbv2.LoadBalancerTypeEnumApplication {
		if v, ok := d.GetOk("connection_logs"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandLoadBalancerConnectionLogsAttributes(v.([]interface{})[0].(map[string]interface{}), false)...)
		} else {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String(loadBalancerAttributeConnectionLogsS3Enabled),
				Value: flex.BoolValueToString(false),
			})
		}
	}

	attributes = append(attributes, loadBalancerAttributes.expand(d, false)...)

	wait := false
	if len(attributes) > 0 {
		if err := modifyLoadBalancerAttributes(ctx, conn, d.Id(), attributes); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		wait = true
	}

	if v, ok := d.GetOk("enforce_security_group_inbound_rules_on_private_link_traffic"); ok && lbType == elbv2.LoadBalancerTypeEnumNetwork {
		input := &elbv2.SetSecurityGroupsInput{
			EnforceSecurityGroupInboundRulesOnPrivateLinkTraffic: aws.String(v.(string)),
			LoadBalancerArn: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("security_groups"); ok {
			input.SecurityGroups = flex.ExpandStringSet(v.(*schema.Set))
		}

		_, err := conn.SetSecurityGroupsWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELBv2 Load Balancer (%s) security groups: %s", d.Id(), err)
		}

		wait = true
	}

	if wait {
		if _, err := waitLoadBalancerActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Load Balancer (%s) create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceLoadBalancerRead(ctx, d, meta)...)
}

func resourceLoadBalancerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	lb, err := FindLoadBalancerByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Load Balancer %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Load Balancer (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, lb.LoadBalancerArn)
	d.Set("arn_suffix", SuffixFromARN(lb.LoadBalancerArn))
	d.Set("customer_owned_ipv4_pool", lb.CustomerOwnedIpv4Pool)
	d.Set("dns_name", lb.DNSName)
	d.Set("enforce_security_group_inbound_rules_on_private_link_traffic", lb.EnforceSecurityGroupInboundRulesOnPrivateLinkTraffic)
	d.Set("internal", aws.StringValue(lb.Scheme) == elbv2.LoadBalancerSchemeEnumInternal)
	d.Set("ip_address_type", lb.IpAddressType)
	lbType := aws.StringValue(lb.Type)
	d.Set("load_balancer_type", lbType)
	d.Set(names.AttrName, lb.LoadBalancerName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(lb.LoadBalancerName)))
	d.Set("security_groups", aws.StringValueSlice(lb.SecurityGroups))
	if err := d.Set("subnet_mapping", flattenSubnetMappingsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnet_mapping: %s", err)
	}
	if err := d.Set("subnets", flattenSubnetsFromAvailabilityZones(lb.AvailabilityZones)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting subnets: %s", err)
	}
	d.Set(names.AttrVPCID, lb.VpcId)
	d.Set("zone_id", lb.CanonicalHostedZoneId)

	attributes, err := FindLoadBalancerAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Load Balancer (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set("access_logs", []interface{}{flattenLoadBalancerAccessLogsAttributes(attributes)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting access_logs: %s", err)
	}

	if lbType == elbv2.LoadBalancerTypeEnumApplication {
		if err := d.Set("connection_logs", []interface{}{flattenLoadBalancerConnectionLogsAttributes(attributes)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting connection_logs: %s", err)
		}
	}

	loadBalancerAttributes.flatten(d, attributes)

	return diags
}

func resourceLoadBalancerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	var attributes []*elbv2.LoadBalancerAttribute

	if d.HasChange("access_logs") {
		if v, ok := d.GetOk("access_logs"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandLoadBalancerAccessLogsAttributes(v.([]interface{})[0].(map[string]interface{}), true)...)
		} else {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String(loadBalancerAttributeAccessLogsS3Enabled),
				Value: flex.BoolValueToString(false),
			})
		}
	}

	if d.HasChange("connection_logs") {
		if v, ok := d.GetOk("connection_logs"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandLoadBalancerConnectionLogsAttributes(v.([]interface{})[0].(map[string]interface{}), true)...)
		} else {
			attributes = append(attributes, &elbv2.LoadBalancerAttribute{
				Key:   aws.String(loadBalancerAttributeConnectionLogsS3Enabled),
				Value: flex.BoolValueToString(false),
			})
		}
	}

	attributes = append(attributes, loadBalancerAttributes.expand(d, true)...)

	if len(attributes) > 0 {
		if err := modifyLoadBalancerAttributes(ctx, conn, d.Id(), attributes); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	if d.HasChanges("enforce_security_group_inbound_rules_on_private_link_traffic", "security_groups") {
		input := &elbv2.SetSecurityGroupsInput{
			LoadBalancerArn: aws.String(d.Id()),
			SecurityGroups:  flex.ExpandStringSet(d.Get("security_groups").(*schema.Set)),
		}

		if v := d.Get("load_balancer_type"); v == elbv2.LoadBalancerTypeEnumNetwork {
			if v, ok := d.GetOk("enforce_security_group_inbound_rules_on_private_link_traffic"); ok {
				input.EnforceSecurityGroupInboundRulesOnPrivateLinkTraffic = aws.String(v.(string))
			}
		}

		_, err := conn.SetSecurityGroupsWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELBv2 Load Balancer (%s) security groups: %s", d.Id(), err)
		}
	}

	if d.HasChanges("subnet_mapping", "subnets") {
		input := &elbv2.SetSubnetsInput{
			LoadBalancerArn: aws.String(d.Id()),
		}

		if d.HasChange("subnet_mapping") {
			if v, ok := d.GetOk("subnet_mapping"); ok && v.(*schema.Set).Len() > 0 {
				input.SubnetMappings = expandSubnetMappings(v.(*schema.Set).List())
			}
		}

		if d.HasChange("subnets") {
			if v, ok := d.GetOk("subnets"); ok {
				input.Subnets = flex.ExpandStringSet(v.(*schema.Set))
			}
		}

		_, err := conn.SetSubnetsWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELBv2 Load Balancer (%s) subnets: %s", d.Id(), err)
		}
	}

	if d.HasChange("ip_address_type") {
		input := &elbv2.SetIpAddressTypeInput{
			IpAddressType:   aws.String(d.Get("ip_address_type").(string)),
			LoadBalancerArn: aws.String(d.Id()),
		}

		_, err := conn.SetIpAddressTypeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELBv2 Load Balancer (%s) address type: %s", d.Id(), err)
		}
	}

	if _, err := waitLoadBalancerActive(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Load Balancer (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceLoadBalancerRead(ctx, d, meta)...)
}

func resourceLoadBalancerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	log.Printf("[INFO] Deleting ELBv2 Load Balancer: %s", d.Id())
	_, err := conn.DeleteLoadBalancerWithContext(ctx, &elbv2.DeleteLoadBalancerInput{
		LoadBalancerArn: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Load Balancer (%s): %s", d.Id(), err)
	}

	ec2conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if err := cleanupALBNetworkInterfaces(ctx, ec2conn, d.Id()); err != nil {
		log.Printf("[WARN] Failed to cleanup ENIs for ALB (%s): %s", d.Id(), err)
	}

	if err := waitForNLBNetworkInterfacesToDetach(ctx, ec2conn, d.Id()); err != nil {
		log.Printf("[WARN] Failed to wait for ENIs to disappear for NLB (%s): %s", d.Id(), err)
	}

	return diags
}

func modifyLoadBalancerAttributes(ctx context.Context, conn *elbv2.ELBV2, arn string, attributes []*elbv2.LoadBalancerAttribute) error {
	input := &elbv2.ModifyLoadBalancerAttributesInput{
		Attributes:      attributes,
		LoadBalancerArn: aws.String(arn),
	}

	// Not all attributes are supported in all partitions.
	for {
		if len(input.Attributes) == 0 {
			return nil
		}

		_, err := conn.ModifyLoadBalancerAttributesWithContext(ctx, input)

		if err != nil {
			// "Validation error: Load balancer attribute key 'routing.http.desync_mitigation_mode' is not recognized"
			// "InvalidConfigurationRequest: Load balancer attribute key 'dns_record.client_routing_policy' is not supported on load balancers with type 'network'"
			re := regexache.MustCompile(`attribute key ('|")?([^'" ]+)('|")? is not (recognized|supported)`)
			if sm := re.FindStringSubmatch(err.Error()); len(sm) > 1 {
				key := sm[2]
				input.Attributes = slices.DeleteFunc(input.Attributes, func(v *elbv2.LoadBalancerAttribute) bool {
					return aws.StringValue(v.Key) == key
				})

				continue
			}

			return fmt.Errorf("modifying ELBv2 Load Balancer (%s) attributes: %w", arn, err)
		}

		return nil
	}
}

type loadBalancerAttributeInfo struct {
	apiAttributeKey            string
	tfType                     schema.ValueType
	loadBalancerTypesSupported []string
}

type loadBalancerAttributeMap map[string]loadBalancerAttributeInfo

var loadBalancerAttributes = loadBalancerAttributeMap(map[string]loadBalancerAttributeInfo{
	"client_keep_alive": {
		apiAttributeKey:            loadBalancerAttributeClientKeepAliveSeconds,
		tfType:                     schema.TypeInt,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
	"desync_mitigation_mode": {
		apiAttributeKey:            loadBalancerAttributeRoutingHTTPDesyncMitigationMode,
		tfType:                     schema.TypeString,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
	"dns_record_client_routing_policy": {
		apiAttributeKey:            loadBalancerAttributeDNSRecordClientRoutingPolicy,
		tfType:                     schema.TypeString,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumNetwork},
	},
	"drop_invalid_header_fields": {
		apiAttributeKey:            loadBalancerAttributeRoutingHTTPDropInvalidHeaderFieldsEnabled,
		tfType:                     schema.TypeBool,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
	"enable_cross_zone_load_balancing": {
		apiAttributeKey: loadBalancerAttributeLoadBalancingCrossZoneEnabled,
		tfType:          schema.TypeBool,
		// Although this attribute is supported for ALBs, it must always be true.
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumNetwork, elbv2.LoadBalancerTypeEnumGateway},
	},
	"enable_deletion_protection": {
		apiAttributeKey:            loadBalancerAttributeDeletionProtectionEnabled,
		tfType:                     schema.TypeBool,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication, elbv2.LoadBalancerTypeEnumNetwork, elbv2.LoadBalancerTypeEnumGateway},
	},
	"enable_http2": {
		apiAttributeKey:            loadBalancerAttributeRoutingHTTP2Enabled,
		tfType:                     schema.TypeBool,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
	"enable_tls_version_and_cipher_suite_headers": {
		apiAttributeKey:            loadBalancerAttributeRoutingHTTPXAmznTLSVersionAndCipherSuiteEnabled,
		tfType:                     schema.TypeBool,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
	"enable_waf_fail_open": {
		apiAttributeKey:            loadBalancerAttributeWAFFailOpenEnabled,
		tfType:                     schema.TypeBool,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
	"enable_xff_client_port": {
		apiAttributeKey:            loadBalancerAttributeRoutingHTTPXFFClientPortEnabled,
		tfType:                     schema.TypeBool,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
	"idle_timeout": {
		apiAttributeKey:            loadBalancerAttributeIdleTimeoutTimeoutSeconds,
		tfType:                     schema.TypeInt,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
	"preserve_host_header": {
		apiAttributeKey:            loadBalancerAttributeRoutingHTTPPreserveHostHeaderEnabled,
		tfType:                     schema.TypeBool,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
	"xff_header_processing_mode": {
		apiAttributeKey:            loadBalancerAttributeRoutingHTTPXFFHeaderProcessingMode,
		tfType:                     schema.TypeString,
		loadBalancerTypesSupported: []string{elbv2.LoadBalancerTypeEnumApplication},
	},
})

func (m loadBalancerAttributeMap) expand(d *schema.ResourceData, update bool) []*elbv2.LoadBalancerAttribute {
	var apiObjects []*elbv2.LoadBalancerAttribute

	loadBalancerType := d.Get("load_balancer_type").(string)
	for tfAttributeName, attributeInfo := range m {
		if update && !d.HasChange(tfAttributeName) {
			continue
		}

		if !slices.Contains(attributeInfo.loadBalancerTypesSupported, loadBalancerType) {
			continue
		}

		switch v, t, k := d.Get(tfAttributeName), attributeInfo.tfType, aws.String(attributeInfo.apiAttributeKey); t {
		case schema.TypeBool:
			v := v.(bool)
			apiObjects = append(apiObjects, &elbv2.LoadBalancerAttribute{
				Key:   k,
				Value: flex.BoolValueToString(v),
			})
		case schema.TypeInt:
			v := v.(int)
			apiObjects = append(apiObjects, &elbv2.LoadBalancerAttribute{
				Key:   k,
				Value: flex.IntValueToString(v),
			})
		case schema.TypeString:
			if v := v.(string); v != "" {
				apiObjects = append(apiObjects, &elbv2.LoadBalancerAttribute{
					Key:   k,
					Value: aws.String(v),
				})
			}
		}
	}

	return apiObjects
}

func (m loadBalancerAttributeMap) flatten(d *schema.ResourceData, apiObjects []*elbv2.LoadBalancerAttribute) {
	for tfAttributeName, attributeInfo := range m {
		k := attributeInfo.apiAttributeKey
		i := slices.IndexFunc(apiObjects, func(v *elbv2.LoadBalancerAttribute) bool {
			return aws.StringValue(v.Key) == k
		})

		if i == -1 {
			continue
		}

		switch v, t := apiObjects[i].Value, attributeInfo.tfType; t {
		case schema.TypeBool:
			d.Set(tfAttributeName, flex.StringToBoolValue(v))
		case schema.TypeInt:
			d.Set(tfAttributeName, flex.StringToIntValue(v))
		case schema.TypeString:
			d.Set(tfAttributeName, v)
		}
	}
}

func FindLoadBalancerByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) (*elbv2.LoadBalancer, error) {
	input := &elbv2.DescribeLoadBalancersInput{
		LoadBalancerArns: aws.StringSlice([]string{arn}),
	}

	output, err := findLoadBalancer(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.LoadBalancerArn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findLoadBalancer(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeLoadBalancersInput) (*elbv2.LoadBalancer, error) {
	output, err := findLoadBalancers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findLoadBalancers(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeLoadBalancersInput) ([]*elbv2.LoadBalancer, error) {
	var output []*elbv2.LoadBalancer

	err := conn.DescribeLoadBalancersPagesWithContext(ctx, input, func(page *elbv2.DescribeLoadBalancersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.LoadBalancers {
			if v != nil && v.State != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeLoadBalancerNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindLoadBalancerAttributesByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) ([]*elbv2.LoadBalancerAttribute, error) {
	input := &elbv2.DescribeLoadBalancerAttributesInput{
		LoadBalancerArn: aws.String(arn),
	}

	output, err := conn.DescribeLoadBalancerAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeLoadBalancerNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Attributes, nil
}

func statusLoadBalancer(ctx context.Context, conn *elbv2.ELBV2, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindLoadBalancerByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State.Code), nil
	}
}

func waitLoadBalancerActive(ctx context.Context, conn *elbv2.ELBV2, arn string, timeout time.Duration) (*elbv2.LoadBalancer, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending:    []string{elbv2.LoadBalancerStateEnumProvisioning, elbv2.LoadBalancerStateEnumFailed},
		Target:     []string{elbv2.LoadBalancerStateEnumActive},
		Refresh:    statusLoadBalancer(ctx, conn, arn),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*elbv2.LoadBalancer); ok {
		tfresource.SetLastError(err, errors.New(aws.StringValue(output.State.Reason)))

		return output, err
	}

	return nil, err
}

// ALB automatically creates ENI(s) on creation
// but the cleanup is asynchronous and may take time
// which then blocks IGW, SG or VPC on deletion
// So we make the cleanup "synchronous" here
func cleanupALBNetworkInterfaces(ctx context.Context, conn *ec2.Client, arn string) error {
	name, err := loadBalancerNameFromARN(arn)
	if err != nil {
		return err
	}

	networkInterfaces, err := tfec2.FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescriptionV2(ctx, conn, "amazon-elb", "ELB "+name)
	if err != nil {
		return err
	}

	var errs []error

	for _, v := range networkInterfaces {
		if v.Attachment == nil {
			continue
		}

		attachmentID := aws.StringValue(v.Attachment.AttachmentId)
		networkInterfaceID := aws.StringValue(v.NetworkInterfaceId)

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

func waitForNLBNetworkInterfacesToDetach(ctx context.Context, conn *ec2.Client, lbArn string) error {
	name, err := loadBalancerNameFromARN(lbArn)
	if err != nil {
		return err
	}

	const (
		timeout = 5 * time.Minute
	)
	_, err = tfresource.RetryUntilEqual(ctx, timeout, 0, func() (int, error) {
		networkInterfaces, err := tfec2.FindNetworkInterfacesByAttachmentInstanceOwnerIDAndDescriptionV2(ctx, conn, "amazon-aws", "ELB "+name)
		if err != nil {
			return 0, err
		}

		return len(networkInterfaces), nil
	})

	return err
}

func loadBalancerNameFromARN(s string) (string, error) {
	v, err := arn.Parse(s)
	if err != nil {
		return "", err
	}

	matches := regexache.MustCompile("([^/]+/[^/]+/[^/]+)$").FindStringSubmatch(v.Resource)
	if len(matches) != 2 {
		return "", fmt.Errorf("unexpected ELBv2 Load Balancer ARN format: %q", s)
	}

	// e.g. app/example-alb/b26e625cdde161e6
	return matches[1], nil
}

func flattenSubnetsFromAvailabilityZones(apiObjects []*elbv2.AvailabilityZone) []string {
	return tfslices.ApplyToAll(apiObjects, func(apiObject *elbv2.AvailabilityZone) string {
		return aws.StringValue(apiObject.SubnetId)
	})
}

func flattenSubnetMappingsFromAvailabilityZones(apiObjects []*elbv2.AvailabilityZone) []map[string]interface{} {
	return tfslices.ApplyToAll(apiObjects, func(apiObject *elbv2.AvailabilityZone) map[string]interface{} {
		tfMap := map[string]interface{}{
			"outpost_id": aws.StringValue(apiObject.OutpostId),
			"subnet_id":  aws.StringValue(apiObject.SubnetId),
		}
		if apiObjects := apiObject.LoadBalancerAddresses; len(apiObjects) > 0 {
			apiObject := apiObjects[0]
			tfMap["allocation_id"] = aws.StringValue(apiObject.AllocationId)
			tfMap["ipv6_address"] = aws.StringValue(apiObject.IPv6Address)
			tfMap["private_ipv4_address"] = aws.StringValue(apiObject.PrivateIPv4Address)
		}

		return tfMap
	})
}

func SuffixFromARN(arn *string) string {
	if arn == nil {
		return ""
	}

	if arnComponents := regexache.MustCompile(`arn:.*:loadbalancer/(.*)`).FindAllStringSubmatch(*arn, -1); len(arnComponents) == 1 {
		if len(arnComponents[0]) == 2 {
			return arnComponents[0][1]
		}
	}

	return ""
}

// Load balancers of type 'network' cannot have their subnets updated,
// cannot have security groups added if none are present, and cannot have
// all security groups removed. If the type is 'network' and any of these
// conditions are met, mark the diff as a ForceNew operation.
func customizeDiffLoadBalancerNLB(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// The current criteria for determining if the operation should be ForceNew:
	// - lb of type "network"
	// - existing resource (id is not "")
	// - there are subnet removals
	//   OR security groups are being added where none currently exist
	//   OR all security groups are being removed
	//
	// Any other combination should be treated as normal. At this time, subnet
	// handling is the only known difference between Network Load Balancers and
	// Application Load Balancers, so the logic below is simple individual checks.
	// If other differences arise we'll want to refactor to check other
	// conditions in combinations, but for now all we handle is subnets
	if lbType := diff.Get("load_balancer_type").(string); lbType != elbv2.LoadBalancerTypeEnumNetwork {
		return nil
	}

	if diff.Id() == "" {
		return nil
	}

	config := diff.GetRawConfig()

	// Subnet diffs.
	// Check for changes here -- SetNewComputed will modify HasChange.
	hasSubnetMappingChanges, hasSubnetsChanges := diff.HasChange("subnet_mapping"), diff.HasChange("subnets")
	if hasSubnetMappingChanges {
		if v := config.GetAttr("subnet_mapping"); v.IsWhollyKnown() {
			o, n := diff.GetChange("subnet_mapping")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			deltaN := ns.Len() - os.Len()
			switch {
			case deltaN == 0:
				// No change in number of subnet mappings, but one of the mappings did change.
				fallthrough
			case deltaN < 0:
				// Subnet mappings removed.
				if err := diff.ForceNew("subnet_mapping"); err != nil {
					return err
				}
			case deltaN > 0:
				// Subnet mappings added. Ensure that the previous mappings didn't change.
				if ns.Intersection(os).Len() != os.Len() {
					if err := diff.ForceNew("subnet_mapping"); err != nil {
						return err
					}
				}
			}
		}

		if err := diff.SetNewComputed("subnets"); err != nil {
			return err
		}
	}
	if hasSubnetsChanges {
		if v := config.GetAttr("subnets"); v.IsWhollyKnown() {
			o, n := diff.GetChange("subnets")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			// In-place increase in number of subnets only.
			if deltaN := ns.Len() - os.Len(); deltaN <= 0 {
				if err := diff.ForceNew("subnets"); err != nil {
					return err
				}
			}
		}

		if err := diff.SetNewComputed("subnet_mapping"); err != nil {
			return err
		}
	}

	// Get diff for security groups.
	if diff.HasChange("security_groups") {
		if v := config.GetAttr("security_groups"); v.IsWhollyKnown() {
			o, n := diff.GetChange("security_groups")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			if (os.Len() == 0 && ns.Len() > 0) || (ns.Len() == 0 && os.Len() > 0) {
				if err := diff.ForceNew("security_groups"); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func customizeDiffLoadBalancerALB(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if lbType := diff.Get("load_balancer_type").(string); lbType != elbv2.LoadBalancerTypeEnumApplication {
		return nil
	}

	if diff.Id() == "" {
		return nil
	}

	config := diff.GetRawConfig()

	// Subnet diffs.
	// Check for changes here -- SetNewComputed will modify HasChange.
	hasSubnetMappingChanges, hasSubnetsChanges := diff.HasChange("subnet_mapping"), diff.HasChange("subnets")
	if hasSubnetMappingChanges {
		if v := config.GetAttr("subnet_mapping"); v.IsWhollyKnown() {
			o, n := diff.GetChange("subnet_mapping")
			os, ns := o.(*schema.Set), n.(*schema.Set)

			deltaN := ns.Len() - os.Len()
			switch {
			case deltaN == 0:
				// No change in number of subnet mappings, but one of the mappings did change.
				if err := diff.ForceNew("subnet_mapping"); err != nil {
					return err
				}
			case deltaN < 0:
				// Subnet mappings removed. Ensure that the remaining mappings didn't change.
				if os.Intersection(ns).Len() != ns.Len() {
					if err := diff.ForceNew("subnet_mapping"); err != nil {
						return err
					}
				}
			case deltaN > 0:
				// Subnet mappings added. Ensure that the previous mappings didn't change.
				if ns.Intersection(os).Len() != os.Len() {
					if err := diff.ForceNew("subnet_mapping"); err != nil {
						return err
					}
				}
			}
		}

		if err := diff.SetNewComputed("subnets"); err != nil {
			return err
		}
	}
	if hasSubnetsChanges {
		if err := diff.SetNewComputed("subnet_mapping"); err != nil {
			return err
		}
	}

	return nil
}

func customizeDiffLoadBalancerGWLB(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if lbType := diff.Get("load_balancer_type").(string); lbType != elbv2.LoadBalancerTypeEnumGateway {
		return nil
	}

	if diff.Id() == "" {
		return nil
	}

	return nil
}

func expandLoadBalancerAccessLogsAttributes(tfMap map[string]interface{}, update bool) []*elbv2.LoadBalancerAttribute {
	if tfMap == nil {
		return nil
	}

	var apiObjects []*elbv2.LoadBalancerAttribute

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObjects = append(apiObjects, &elbv2.LoadBalancerAttribute{
			Key:   aws.String(loadBalancerAttributeAccessLogsS3Enabled),
			Value: flex.BoolValueToString(v),
		})

		if v {
			if v, ok := tfMap[names.AttrBucket].(string); ok && (update || v != "") {
				apiObjects = append(apiObjects, &elbv2.LoadBalancerAttribute{
					Key:   aws.String(loadBalancerAttributeAccessLogsS3Bucket),
					Value: aws.String(v),
				})
			}

			if v, ok := tfMap["prefix"].(string); ok && (update || v != "") {
				apiObjects = append(apiObjects, &elbv2.LoadBalancerAttribute{
					Key:   aws.String(loadBalancerAttributeAccessLogsS3Prefix),
					Value: aws.String(v),
				})
			}
		}
	}

	return apiObjects
}

func expandLoadBalancerConnectionLogsAttributes(tfMap map[string]interface{}, update bool) []*elbv2.LoadBalancerAttribute {
	if tfMap == nil {
		return nil
	}

	var apiObjects []*elbv2.LoadBalancerAttribute

	if v, ok := tfMap[names.AttrEnabled].(bool); ok {
		apiObjects = append(apiObjects, &elbv2.LoadBalancerAttribute{
			Key:   aws.String(loadBalancerAttributeConnectionLogsS3Enabled),
			Value: flex.BoolValueToString(v),
		})

		if v {
			if v, ok := tfMap[names.AttrBucket].(string); ok && (update || v != "") {
				apiObjects = append(apiObjects, &elbv2.LoadBalancerAttribute{
					Key:   aws.String(loadBalancerAttributeConnectionLogsS3Bucket),
					Value: aws.String(v),
				})
			}

			if v, ok := tfMap["prefix"].(string); ok && (update || v != "") {
				apiObjects = append(apiObjects, &elbv2.LoadBalancerAttribute{
					Key:   aws.String(loadBalancerAttributeConnectionLogsS3Prefix),
					Value: aws.String(v),
				})
			}
		}
	}

	return apiObjects
}

func flattenLoadBalancerAccessLogsAttributes(apiObjects []*elbv2.LoadBalancerAttribute) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{}

	for _, apiObject := range apiObjects {
		switch k, v := aws.StringValue(apiObject.Key), apiObject.Value; k {
		case loadBalancerAttributeAccessLogsS3Enabled:
			tfMap[names.AttrEnabled] = flex.StringToBoolValue(v)
		case loadBalancerAttributeAccessLogsS3Bucket:
			tfMap[names.AttrBucket] = aws.StringValue(v)
		case loadBalancerAttributeAccessLogsS3Prefix:
			tfMap["prefix"] = aws.StringValue(v)
		}
	}

	return tfMap
}

func flattenLoadBalancerConnectionLogsAttributes(apiObjects []*elbv2.LoadBalancerAttribute) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{}

	for _, apiObject := range apiObjects {
		switch k, v := aws.StringValue(apiObject.Key), apiObject.Value; k {
		case loadBalancerAttributeConnectionLogsS3Enabled:
			tfMap[names.AttrEnabled] = flex.StringToBoolValue(v)
		case loadBalancerAttributeConnectionLogsS3Bucket:
			tfMap[names.AttrBucket] = aws.StringValue(v)
		case loadBalancerAttributeConnectionLogsS3Prefix:
			tfMap["prefix"] = aws.StringValue(v)
		}
	}

	return tfMap
}

func expandSubnetMapping(tfMap map[string]interface{}) *elbv2.SubnetMapping {
	if tfMap == nil {
		return nil
	}

	apiObject := &elbv2.SubnetMapping{}

	if v, ok := tfMap["allocation_id"].(string); ok && v != "" {
		apiObject.AllocationId = aws.String(v)
	}

	if v, ok := tfMap["ipv6_address"].(string); ok && v != "" {
		apiObject.IPv6Address = aws.String(v)
	}

	if v, ok := tfMap["private_ipv4_address"].(string); ok && v != "" {
		apiObject.PrivateIPv4Address = aws.String(v)
	}

	if v, ok := tfMap["subnet_id"].(string); ok && v != "" {
		apiObject.SubnetId = aws.String(v)
	}

	return apiObject
}

func expandSubnetMappings(tfList []interface{}) []*elbv2.SubnetMapping {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*elbv2.SubnetMapping

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandSubnetMapping(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}
