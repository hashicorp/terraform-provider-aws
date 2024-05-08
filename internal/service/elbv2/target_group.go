// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
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
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_alb_target_group", name="Target Group")
// @SDKResource("aws_lb_target_group", name="Target Group")
// @Tags(identifierAttribute="id")
func ResourceTargetGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTargetGroupCreate,
		ReadWithoutTimeout:   resourceTargetGroupRead,
		UpdateWithoutTimeout: resourceTargetGroupUpdate,
		DeleteWithoutTimeout: resourceTargetGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			resourceTargetGroupCustomizeDiff,
			customizeDiffTargetGroupTargetTypeLambda,
			customizeDiffTargetGroupTargetTypeNotLambda,
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn_suffix": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_termination": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"deregistration_delay": {
				Type:         nullable.TypeNullableInt,
				Optional:     true,
				Default:      300,
				ValidateFunc: nullable.ValidateTypeStringNullableIntBetween(0, 3600),
			},
			"health_check": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"healthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3,
							ValidateFunc: validation.IntBetween(2, 10),
						},
						"interval": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      30,
							ValidateFunc: validation.IntBetween(5, 300),
						},
						"matcher": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"path": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 1024),
								verify.StringHasPrefix("/"),
							),
						},
						names.AttrPort: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          healthCheckPortTrafficPort,
							ValidateFunc:     validTargetGroupHealthCheckPort,
							DiffSuppressFunc: suppressIfTargetType(elbv2.TargetTypeEnumLambda),
						},
						"protocol": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  elbv2.ProtocolEnumHttp,
							StateFunc: func(v interface{}) string {
								return strings.ToUpper(v.(string))
							},
							ValidateFunc:     validation.StringInSlice(healthCheckProtocolEnumValues(), true),
							DiffSuppressFunc: suppressIfTargetType(elbv2.TargetTypeEnumLambda),
						},
						"timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntBetween(2, 120),
						},
						"unhealthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      3,
							ValidateFunc: validation.IntBetween(2, 10),
						},
					},
				},
			},
			"ip_address_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(elbv2.TargetGroupIpAddressTypeEnum_Values(), false),
			},
			"lambda_multi_value_headers_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"load_balancer_arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"load_balancing_algorithm_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(loadBalancingAlgorithmType_Values(), false),
			},
			"load_balancing_anomaly_mitigation": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(loadBalancingAnomalyMitigationType_Values(), false),
			},
			"load_balancing_cross_zone_enabled": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(loadBalancingCrossZoneEnabled_Values(), false),
			},
			names.AttrName: {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
				ValidateFunc:  validTargetGroupName,
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{names.AttrName},
				ValidateFunc:  validTargetGroupNamePrefix,
			},
			names.AttrPort: {
				Type:             schema.TypeInt,
				Optional:         true,
				ForceNew:         true,
				ValidateFunc:     validation.IntBetween(1, 65535),
				DiffSuppressFunc: suppressIfTargetType(elbv2.TargetTypeEnumLambda),
			},
			"preserve_client_ip": {
				Type:             nullable.TypeNullableBool,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: nullable.DiffSuppressNullableBool,
				ValidateFunc:     nullable.ValidateTypeStringNullableBool,
			},
			"protocol": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateFunc:     validation.StringInSlice(elbv2.ProtocolEnum_Values(), true),
				DiffSuppressFunc: suppressIfTargetType(elbv2.TargetTypeEnumLambda),
			},
			"protocol_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
				ValidateFunc: validation.StringInSlice(protocolVersionEnumValues(), true),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Don't suppress on creation, so that warnings are actually called
					if d.Id() == "" {
						return false
					}
					if d.Get("target_type").(string) == elbv2.TargetTypeEnumLambda {
						return true
					}
					switch d.Get("protocol").(string) {
					case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
						return false
					}
					return true
				},
			},
			"proxy_protocol_v2": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"slow_start": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  0,
				ValidateFunc: validation.Any(
					validation.IntBetween(0, 0),
					validation.IntBetween(30, 900),
				),
			},
			"stickiness": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cookie_duration": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      86400,
							ValidateFunc: validation.IntBetween(0, 604800),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								switch d.Get("protocol").(string) {
								case elbv2.ProtocolEnumTcp, elbv2.ProtocolEnumUdp, elbv2.ProtocolEnumTcpUdp, elbv2.ProtocolEnumTls, elbv2.ProtocolEnumGeneve:
									return true
								}
								return false
							},
						},
						"cookie_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrEnabled: {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						names.AttrType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(stickinessType_Values(), false),
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_failover": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"on_deregistration": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(targetFailover_Values(), false),
						},
						"on_unhealthy": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(targetFailover_Values(), false),
						},
					},
				},
			},
			"target_health_state": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_unhealthy_connection_termination": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},
			"target_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      elbv2.TargetTypeEnumInstance,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(elbv2.TargetTypeEnum_Values(), false),
			},
			names.AttrVPCID: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressIfTargetType(elbv2.TargetTypeEnumLambda),
			},
		},
	}
}

func suppressIfTargetType(t string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		// Don't suppress on creation, so that warnings are actually called
		if d.Id() == "" {
			return false
		}
		return d.Get("target_type").(string) == t
	}
}

func resourceTargetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrName).(string)),
		create.WithConfiguredPrefix(d.Get("name_prefix").(string)),
		create.WithDefaultPrefix("tf-"),
	).Generate()
	exist, err := findTargetGroupByName(ctx, conn, name)

	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Group (%s): %s", name, err)
	}

	if exist != nil {
		return sdkdiag.AppendErrorf(diags, "ELBv2 Target Group (%s) already exists", name)
	}

	targetGroupRuntimeValidation(d, &diags)

	protocol := d.Get("protocol").(string)
	targetType := d.Get("target_type").(string)
	input := &elbv2.CreateTargetGroupInput{
		Name:       aws.String(name),
		Tags:       getTagsIn(ctx),
		TargetType: aws.String(targetType),
	}

	if targetType != elbv2.TargetTypeEnumLambda {
		input.Port = aws.Int64(int64(d.Get(names.AttrPort).(int)))
		input.Protocol = aws.String(protocol)
		switch protocol {
		case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
			input.ProtocolVersion = aws.String(d.Get("protocol_version").(string))
		}
		input.VpcId = aws.String(d.Get(names.AttrVPCID).(string))

		if targetType == elbv2.TargetTypeEnumIp {
			if v, ok := d.GetOk("ip_address_type"); ok {
				input.IpAddressType = aws.String(v.(string))
			}
		}
	}

	if v, ok := d.GetOk("health_check"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		input.HealthCheckEnabled = aws.Bool(tfMap[names.AttrEnabled].(bool))
		input.HealthCheckIntervalSeconds = aws.Int64(int64(tfMap["interval"].(int)))
		input.HealthyThresholdCount = aws.Int64(int64(tfMap["healthy_threshold"].(int)))
		input.UnhealthyThresholdCount = aws.Int64(int64(tfMap["unhealthy_threshold"].(int)))

		if v, ok := tfMap["timeout"].(int); ok && v != 0 {
			input.HealthCheckTimeoutSeconds = aws.Int64(int64(v))
		}

		healthCheckProtocol := tfMap["protocol"].(string)
		if healthCheckProtocol != elbv2.ProtocolEnumTcp {
			if v, ok := tfMap["path"].(string); ok && v != "" {
				input.HealthCheckPath = aws.String(v)
			}

			if v, ok := tfMap["matcher"].(string); ok && v != "" {
				if protocolVersion := d.Get("protocol_version").(string); protocolVersion == protocolVersionGRPC {
					input.Matcher = &elbv2.Matcher{
						GrpcCode: aws.String(v),
					}
				} else {
					input.Matcher = &elbv2.Matcher{
						HttpCode: aws.String(v),
					}
				}
			}
		}

		if targetType != elbv2.TargetTypeEnumLambda {
			input.HealthCheckPort = aws.String(tfMap[names.AttrPort].(string))
			input.HealthCheckProtocol = aws.String(healthCheckProtocol)
		}
	}

	output, err := conn.CreateTargetGroupWithContext(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
		input.Tags = nil

		output, err = conn.CreateTargetGroupWithContext(ctx, input)
	}

	// Tags are not supported on creation with some protocol types(i.e. GENEVE)
	// Retry creation without tags
	if input.Tags != nil && tfawserr.ErrMessageContains(err, errCodeValidationError, tagsOnCreationErrMessage) {
		input.Tags = nil

		output, err = conn.CreateTargetGroupWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELBv2 Target Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.TargetGroups[0].TargetGroupArn))

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTargetGroupByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Target Group (%s) create: %s", d.Id(), err)
	}

	var attributes []*elbv2.TargetGroupAttribute

	switch targetType {
	case elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp:
		if v, ok := d.GetOk("stickiness"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandTargetGroupStickinessAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
		}

		if v, ok := d.GetOk("target_failover"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandTargetGroupTargetFailoverAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
		}

		if v, ok := d.GetOk("target_health_state"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandTargetGroupTargetHealthStateAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
		}
	}

	attributes = append(attributes, targetGroupAttributes.expand(d, targetType, false)...)

	if len(attributes) > 0 {
		input := &elbv2.ModifyTargetGroupAttributesInput{
			Attributes:     attributes,
			TargetGroupArn: aws.String(d.Id()),
		}

		_, err := conn.ModifyTargetGroupAttributesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Target Group (%s) attributes: %s", d.Id(), err)
		}
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(conn.PartitionID, err) {
			return append(diags, resourceTargetGroupRead(ctx, d, meta)...)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting ELBv2 Target Group (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTargetGroupRead(ctx, d, meta)...)
}

func resourceTargetGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	targetGroup, err := FindTargetGroupByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Target Group %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Group (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() {
		targetGroupRuntimeValidation(d, &diags)
	}

	d.Set(names.AttrARN, targetGroup.TargetGroupArn)
	d.Set("arn_suffix", TargetGroupSuffixFromARN(targetGroup.TargetGroupArn))
	if err := d.Set("health_check", flattenTargetGroupHealthCheck(targetGroup)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting health_check: %s", err)
	}
	d.Set("ip_address_type", targetGroup.IpAddressType)
	d.Set("load_balancer_arns", flex.FlattenStringSet(targetGroup.LoadBalancerArns))
	d.Set(names.AttrName, targetGroup.TargetGroupName)
	d.Set("name_prefix", create.NamePrefixFromName(aws.StringValue(targetGroup.TargetGroupName)))
	targetType := aws.StringValue(targetGroup.TargetType)
	d.Set("target_type", targetType)

	if _, ok := d.GetOk(names.AttrPort); targetGroup.Port != nil || ok {
		d.Set(names.AttrPort, targetGroup.Port)
	}
	var protocol string
	if _, ok := d.GetOk("protocol"); targetGroup.Protocol != nil || ok {
		protocol = aws.StringValue(targetGroup.Protocol)
		d.Set("protocol", protocol)
	}
	if _, ok := d.GetOk("protocol_version"); targetGroup.ProtocolVersion != nil || ok {
		d.Set("protocol_version", targetGroup.ProtocolVersion)
	}
	if _, ok := d.GetOk(names.AttrVPCID); targetGroup.VpcId != nil || ok {
		d.Set(names.AttrVPCID, targetGroup.VpcId)
	}

	attributes, err := findTargetGroupAttributesByARN(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Group (%s) attributes: %s", d.Id(), err)
	}

	if err := d.Set("stickiness", []interface{}{flattenTargetGroupStickinessAttributes(attributes, protocol)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting stickiness: %s", err)
	}

	if err := d.Set("target_failover", []interface{}{flattenTargetGroupTargetFailoverAttributes(attributes, protocol)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_failover: %s", err)
	}

	if err := d.Set("target_health_state", []interface{}{flattenTargetGroupTargetHealthStateAttributes(attributes, protocol)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_health_state: %s", err)
	}

	targetGroupAttributes.flatten(d, targetType, attributes)

	return diags
}

func resourceTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	protocol := d.Get("protocol").(string)
	targetType := d.Get("target_type").(string)

	if d.HasChange("health_check") {
		if v, ok := d.GetOk("health_check"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			input := &elbv2.ModifyTargetGroupInput{
				HealthCheckEnabled:         aws.Bool(tfMap[names.AttrEnabled].(bool)),
				HealthCheckIntervalSeconds: aws.Int64(int64(tfMap["interval"].(int))),
				HealthyThresholdCount:      aws.Int64(int64(tfMap["healthy_threshold"].(int))),
				TargetGroupArn:             aws.String(d.Id()),
				UnhealthyThresholdCount:    aws.Int64(int64(tfMap["unhealthy_threshold"].(int))),
			}

			if v, ok := tfMap["timeout"].(int); ok && v != 0 {
				input.HealthCheckTimeoutSeconds = aws.Int64(int64(v))
			}

			healthCheckProtocol := tfMap["protocol"].(string)
			if healthCheckProtocol != elbv2.ProtocolEnumTcp {
				if v, ok := tfMap["matcher"].(string); ok {
					if protocolVersion := d.Get("protocol_version").(string); protocolVersion == protocolVersionGRPC {
						input.Matcher = &elbv2.Matcher{
							GrpcCode: aws.String(v),
						}
					} else {
						input.Matcher = &elbv2.Matcher{
							HttpCode: aws.String(v),
						}
					}
				}
				input.HealthCheckPath = aws.String(tfMap["path"].(string))
			}

			if targetType != elbv2.TargetTypeEnumLambda {
				input.HealthCheckPort = aws.String(tfMap[names.AttrPort].(string))
				input.HealthCheckProtocol = aws.String(healthCheckProtocol)
			}

			_, err := conn.ModifyTargetGroupWithContext(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Target Group (%s): %s", d.Id(), err)
			}
		}
	}

	var attributes []*elbv2.TargetGroupAttribute

	switch targetType {
	case elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp:
		if d.HasChange("stickiness") {
			if v, ok := d.GetOk("stickiness"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				attributes = append(attributes, expandTargetGroupStickinessAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
			} else {
				attributes = append(attributes, &elbv2.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeStickinessEnabled),
					Value: flex.BoolValueToString(false),
				})
			}
		}

		if d.HasChange("target_failover") {
			if v, ok := d.GetOk("target_failover"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				attributes = append(attributes, expandTargetGroupTargetFailoverAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
			}
		}

		if d.HasChange("target_health_state") {
			if v, ok := d.GetOk("target_health_state"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				attributes = append(attributes, expandTargetGroupTargetHealthStateAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
			}
		}
	}

	attributes = append(attributes, targetGroupAttributes.expand(d, targetType, true)...)

	if len(attributes) > 0 {
		input := &elbv2.ModifyTargetGroupAttributesInput{
			Attributes:     attributes,
			TargetGroupArn: aws.String(d.Id()),
		}

		_, err := conn.ModifyTargetGroupAttributesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Target Group (%s) attributes: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTargetGroupRead(ctx, d, meta)...)
}

func resourceTargetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	log.Printf("[DEBUG] Deleting ELBv2 Target Group: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrMessageContains(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.DeleteTargetGroupWithContext(ctx, &elbv2.DeleteTargetGroupInput{
			TargetGroupArn: aws.String(d.Id()),
		})
	}, elbv2.ErrCodeResourceInUseException, "is currently in use by a listener or a rule")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Target Group (%s): %s", d.Id(), err)
	}

	return diags
}

type targetGroupAttributeInfo struct {
	apiAttributeKey      string
	tfType               schema.ValueType
	tfNullableType       schema.ValueType
	targetTypesSupported []string
}

type targetGroupAttributeMap map[string]targetGroupAttributeInfo

var targetGroupAttributes = targetGroupAttributeMap(map[string]targetGroupAttributeInfo{
	"connection_termination": {
		apiAttributeKey:      targetGroupAttributeDeregistrationDelayConnectionTerminationEnabled,
		tfType:               schema.TypeBool,
		targetTypesSupported: []string{elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp},
	},
	"deregistration_delay": {
		apiAttributeKey:      targetGroupAttributeDeregistrationDelayTimeoutSeconds,
		tfType:               schema.TypeString,
		tfNullableType:       schema.TypeInt,
		targetTypesSupported: []string{elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp},
	},
	"lambda_multi_value_headers_enabled": {
		apiAttributeKey:      targetGroupAttributeLambdaMultiValueHeadersEnabled,
		tfType:               schema.TypeBool,
		targetTypesSupported: []string{elbv2.TargetTypeEnumLambda},
	},
	"load_balancing_algorithm_type": {
		apiAttributeKey:      targetGroupAttributeLoadBalancingAlgorithmType,
		tfType:               schema.TypeString,
		targetTypesSupported: []string{elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp},
	},
	"load_balancing_anomaly_mitigation": {
		apiAttributeKey:      targetGroupAttributeLoadBalancingAlgorithmAnomalyMitigation,
		tfType:               schema.TypeString,
		targetTypesSupported: []string{elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp},
	},
	"load_balancing_cross_zone_enabled": {
		apiAttributeKey:      targetGroupAttributeLoadBalancingCrossZoneEnabled,
		tfType:               schema.TypeString,
		targetTypesSupported: []string{elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp},
	},
	"preserve_client_ip": {
		apiAttributeKey:      targetGroupAttributePreserveClientIPEnabled,
		tfType:               schema.TypeString,
		tfNullableType:       schema.TypeBool,
		targetTypesSupported: []string{elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp},
	},
	"proxy_protocol_v2": {
		apiAttributeKey:      targetGroupAttributeProxyProtocolV2Enabled,
		tfType:               schema.TypeBool,
		targetTypesSupported: []string{elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp},
	},
	"slow_start": {
		apiAttributeKey:      targetGroupAttributeSlowStartDurationSeconds,
		tfType:               schema.TypeInt,
		targetTypesSupported: []string{elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp},
	},
})

func (m targetGroupAttributeMap) expand(d *schema.ResourceData, targetType string, update bool) []*elbv2.TargetGroupAttribute {
	var apiObjects []*elbv2.TargetGroupAttribute

	for tfAttributeName, attributeInfo := range m {
		if update && !d.HasChange(tfAttributeName) {
			continue
		}

		if !slices.Contains(attributeInfo.targetTypesSupported, targetType) {
			continue
		}

		switch v, nt, k := d.Get(tfAttributeName), attributeInfo.tfNullableType, aws.String(attributeInfo.apiAttributeKey); nt {
		case schema.TypeBool:
			v := v.(string)
			if v, null, _ := nullable.Bool(v).ValueBool(); !null {
				apiObjects = append(apiObjects, &elbv2.TargetGroupAttribute{
					Key:   k,
					Value: flex.BoolValueToString(v),
				})
			}
		case schema.TypeInt:
			v := v.(string)
			if v, null, _ := nullable.Int(v).ValueInt64(); !null {
				apiObjects = append(apiObjects, &elbv2.TargetGroupAttribute{
					Key:   k,
					Value: flex.Int64ValueToString(v),
				})
			}
		default:
			switch attributeInfo.tfType {
			case schema.TypeBool:
				if v := v.(bool); v || update {
					apiObjects = append(apiObjects, &elbv2.TargetGroupAttribute{
						Key:   k,
						Value: flex.BoolValueToString(v),
					})
				}
			case schema.TypeInt:
				if v := v.(int); v > 0 || update {
					apiObjects = append(apiObjects, &elbv2.TargetGroupAttribute{
						Key:   k,
						Value: flex.IntValueToString(v),
					})
				}
			case schema.TypeString:
				if v := v.(string); v != "" || update {
					apiObjects = append(apiObjects, &elbv2.TargetGroupAttribute{
						Key:   k,
						Value: aws.String(v),
					})
				}
			}
		}
	}

	return apiObjects
}

func (m targetGroupAttributeMap) flatten(d *schema.ResourceData, targetType string, apiObjects []*elbv2.TargetGroupAttribute) {
	for tfAttributeName, attributeInfo := range m {
		if !slices.Contains(attributeInfo.targetTypesSupported, targetType) {
			continue
		}

		k := attributeInfo.apiAttributeKey
		i := slices.IndexFunc(apiObjects, func(v *elbv2.TargetGroupAttribute) bool {
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

func FindTargetGroupByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) (*elbv2.TargetGroup, error) {
	input := &elbv2.DescribeTargetGroupsInput{
		TargetGroupArns: aws.StringSlice([]string{arn}),
	}

	output, err := findTargetGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TargetGroupArn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTargetGroupByName(ctx context.Context, conn *elbv2.ELBV2, name string) (*elbv2.TargetGroup, error) {
	input := &elbv2.DescribeTargetGroupsInput{
		Names: aws.StringSlice([]string{name}),
	}

	output, err := findTargetGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.StringValue(output.TargetGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTargetGroup(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTargetGroupsInput) (*elbv2.TargetGroup, error) {
	output, err := findTargetGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findTargetGroups(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTargetGroupsInput) ([]*elbv2.TargetGroup, error) {
	var output []*elbv2.TargetGroup

	err := conn.DescribeTargetGroupsPagesWithContext(ctx, input, func(page *elbv2.DescribeTargetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.TargetGroups {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
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

func findTargetGroupAttributesByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) ([]*elbv2.TargetGroupAttribute, error) {
	input := &elbv2.DescribeTargetGroupAttributesInput{
		TargetGroupArn: aws.String(arn),
	}

	output, err := conn.DescribeTargetGroupAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, elbv2.ErrCodeTargetGroupNotFoundException) {
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

func validTargetGroupHealthCheckPort(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == healthCheckPortTrafficPort {
		return
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("%q must be a valid port number (1-65536) or %q", k, "traffic-port"))
	}

	if port < 1 || port > 65536 {
		errors = append(errors, fmt.Errorf("%q must be a valid port number (1-65536) or %q", k, "traffic-port"))
	}

	return
}

func TargetGroupSuffixFromARN(arn *string) string {
	if arn == nil {
		return ""
	}

	if arnComponents := regexache.MustCompile(`arn:.*:targetgroup/(.*)`).FindAllStringSubmatch(*arn, -1); len(arnComponents) == 1 {
		if len(arnComponents[0]) == 2 {
			return fmt.Sprintf("targetgroup/%s", arnComponents[0][1])
		}
	}

	return ""
}

func resourceTargetGroupCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	healthCheck := make(map[string]any)
	if healthChecks := diff.Get("health_check").([]interface{}); len(healthChecks) == 1 {
		healthCheck = healthChecks[0].(map[string]interface{})
	}

	healtCheckPath := cty.GetAttrPath("health_check").IndexInt(0)

	if p, ok := healthCheck["protocol"].(string); ok && strings.ToUpper(p) == elbv2.ProtocolEnumTcp {
		if m := healthCheck["matcher"].(string); m != "" {
			return sdkdiag.DiagnosticError(errs.NewAttributeConflictsWhenError(
				healtCheckPath.GetAttr("matcher"),
				healtCheckPath.GetAttr("protocol"),
				elbv2.ProtocolEnumTcp,
			))
		}

		if m := healthCheck["path"].(string); m != "" {
			return sdkdiag.DiagnosticError(errs.NewAttributeConflictsWhenError(
				healtCheckPath.GetAttr("path"),
				healtCheckPath.GetAttr("protocol"),
				elbv2.ProtocolEnumTcp,
			))
		}
	}

	protocol := diff.Get("protocol").(string)

	switch protocol {
	case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
		if p, ok := healthCheck["protocol"].(string); ok && strings.ToUpper(p) == elbv2.ProtocolEnumTcp {
			return fmt.Errorf("Attribute %q cannot have value %q when %q is %q.",
				errs.PathString(healtCheckPath.GetAttr("protocol")),
				elbv2.ProtocolEnumTcp,
				errs.PathString(cty.GetAttrPath("protocol")),
				protocol,
			)
		}
	}

	if diff.Id() == "" {
		return nil
	}

	return nil
}

func customizeDiffTargetGroupTargetTypeLambda(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	if diff.Get("target_type").(string) != elbv2.TargetTypeEnumLambda {
		return nil
	}

	if healthChecks := diff.Get("health_check").([]interface{}); len(healthChecks) == 1 {
		healthCheck := healthChecks[0].(map[string]interface{})
		healtCheckPath := cty.GetAttrPath("health_check").IndexInt(0)
		healthCheckProtocol := healthCheck["protocol"].(string)

		if healthCheckProtocol == elbv2.ProtocolEnumTcp {
			return fmt.Errorf("Attribute %q cannot have value %q when %q is %q.",
				errs.PathString(healtCheckPath.GetAttr("protocol")),
				elbv2.ProtocolEnumTcp,
				errs.PathString(cty.GetAttrPath("target_type")),
				elbv2.TargetTypeEnumLambda,
			)
		}
	}

	return nil
}

func customizeDiffTargetGroupTargetTypeNotLambda(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	targetType := diff.Get("target_type").(string)
	if targetType == elbv2.TargetTypeEnumLambda {
		return nil
	}

	config := diff.GetRawConfig()

	if v := config.GetAttr(names.AttrPort); v.IsKnown() && v.IsNull() {
		return sdkdiag.DiagnosticError(errs.NewAttributeRequiredWhenError(
			cty.GetAttrPath(names.AttrPort),
			cty.GetAttrPath("target_type"),
			targetType,
		))
	}

	if v := config.GetAttr("protocol"); v.IsKnown() && v.IsNull() {
		return sdkdiag.DiagnosticError(errs.NewAttributeRequiredWhenError(
			cty.GetAttrPath("protocol"),
			cty.GetAttrPath("target_type"),
			targetType,
		))
	}

	if v := config.GetAttr(names.AttrVPCID); v.IsKnown() && v.IsNull() {
		return sdkdiag.DiagnosticError(errs.NewAttributeRequiredWhenError(
			cty.GetAttrPath(names.AttrVPCID),
			cty.GetAttrPath("target_type"),
			targetType,
		))
	}

	return nil
}

func flattenTargetGroupHealthCheck(apiObject *elbv2.TargetGroup) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrEnabled:     aws.BoolValue(apiObject.HealthCheckEnabled),
		"healthy_threshold":   int(aws.Int64Value(apiObject.HealthyThresholdCount)),
		"interval":            int(aws.Int64Value(apiObject.HealthCheckIntervalSeconds)),
		names.AttrPort:        aws.StringValue(apiObject.HealthCheckPort),
		"protocol":            aws.StringValue(apiObject.HealthCheckProtocol),
		"timeout":             int(aws.Int64Value(apiObject.HealthCheckTimeoutSeconds)),
		"unhealthy_threshold": int(aws.Int64Value(apiObject.UnhealthyThresholdCount)),
	}

	if v := apiObject.HealthCheckPath; v != nil {
		tfMap["path"] = aws.StringValue(v)
	}

	if apiObject := apiObject.Matcher; apiObject != nil {
		if v := apiObject.HttpCode; v != nil {
			tfMap["matcher"] = aws.StringValue(v)
		}
		if v := apiObject.GrpcCode; v != nil {
			tfMap["matcher"] = aws.StringValue(v)
		}
	}

	return []interface{}{tfMap}
}

func expandTargetGroupStickinessAttributes(tfMap map[string]interface{}, protocol string) []*elbv2.TargetGroupAttribute {
	if tfMap == nil {
		return nil
	}

	apiObjects := []*elbv2.TargetGroupAttribute{
		{
			Key:   aws.String(targetGroupAttributeStickinessEnabled),
			Value: flex.BoolValueToString(tfMap[names.AttrEnabled].(bool)),
		},
		{
			Key:   aws.String(targetGroupAttributeStickinessType),
			Value: aws.String(tfMap[names.AttrType].(string)),
		},
	}

	switch protocol {
	case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
		switch stickinessType := tfMap[names.AttrType].(string); stickinessType {
		case stickinessTypeLBCookie:
			apiObjects = append(apiObjects,
				&elbv2.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeStickinessLBCookieDurationSeconds),
					Value: flex.IntValueToString(tfMap["cookie_duration"].(int)),
				})
		case stickinessTypeAppCookie:
			apiObjects = append(apiObjects,
				&elbv2.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeStickinessAppCookieCookieName),
					Value: aws.String(tfMap["cookie_name"].(string)),
				},
				&elbv2.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeStickinessAppCookieDurationSeconds),
					Value: flex.IntValueToString(tfMap["cookie_duration"].(int)),
				})
		}
	}

	return apiObjects
}

func flattenTargetGroupStickinessAttributes(apiObjects []*elbv2.TargetGroupAttribute, protocol string) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{}

	var stickinessType string
	for _, apiObject := range apiObjects {
		switch k, v := aws.StringValue(apiObject.Key), apiObject.Value; k {
		case targetGroupAttributeStickinessEnabled:
			tfMap[names.AttrEnabled] = flex.StringToBoolValue(v)
		case targetGroupAttributeStickinessType:
			stickinessType = aws.StringValue(v)
			tfMap[names.AttrType] = stickinessType
		}
	}

	switch protocol {
	case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
		for _, apiObject := range apiObjects {
			k, v := aws.StringValue(apiObject.Key), apiObject.Value
			switch {
			case k == targetGroupAttributeStickinessLBCookieDurationSeconds && stickinessType == stickinessTypeLBCookie:
				tfMap["cookie_duration"] = flex.StringToIntValue(v)
			case k == targetGroupAttributeStickinessAppCookieCookieName && stickinessType == stickinessTypeAppCookie:
				tfMap["cookie_name"] = aws.StringValue(v)
			case k == targetGroupAttributeStickinessAppCookieDurationSeconds && stickinessType == stickinessTypeAppCookie:
				tfMap["cookie_duration"] = flex.StringToIntValue(v)
			}
		}
	}

	return tfMap
}

func expandTargetGroupTargetFailoverAttributes(tfMap map[string]interface{}, protocol string) []*elbv2.TargetGroupAttribute {
	if tfMap == nil {
		return nil
	}

	var apiObjects []*elbv2.TargetGroupAttribute

	switch protocol {
	case elbv2.ProtocolEnumGeneve:
		apiObjects = append(apiObjects,
			&elbv2.TargetGroupAttribute{
				Key:   aws.String(targetGroupAttributeTargetFailoverOnDeregistration),
				Value: aws.String(tfMap["on_deregistration"].(string)),
			},
			&elbv2.TargetGroupAttribute{
				Key:   aws.String(targetGroupAttributeTargetFailoverOnUnhealthy),
				Value: aws.String(tfMap["on_unhealthy"].(string)),
			})
	}

	return apiObjects
}

func flattenTargetGroupTargetFailoverAttributes(apiObjects []*elbv2.TargetGroupAttribute, protocol string) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{}

	switch protocol {
	case elbv2.ProtocolEnumGeneve:
		for _, apiObject := range apiObjects {
			switch k, v := aws.StringValue(apiObject.Key), apiObject.Value; k {
			case targetGroupAttributeTargetFailoverOnDeregistration:
				tfMap["on_deregistration"] = aws.StringValue(v)
			case targetGroupAttributeTargetFailoverOnUnhealthy:
				tfMap["on_unhealthy"] = aws.StringValue(v)
			}
		}
	}

	return tfMap
}

func expandTargetGroupTargetHealthStateAttributes(tfMap map[string]interface{}, protocol string) []*elbv2.TargetGroupAttribute {
	if tfMap == nil {
		return nil
	}

	var apiObjects []*elbv2.TargetGroupAttribute

	switch protocol {
	case elbv2.ProtocolEnumTcp, elbv2.ProtocolEnumTls:
		apiObjects = append(apiObjects,
			&elbv2.TargetGroupAttribute{
				Key:   aws.String(targetGroupAttributeTargetHealthStateUnhealthyConnectionTerminationEnabled),
				Value: flex.BoolValueToString(tfMap["enable_unhealthy_connection_termination"].(bool)),
			})
	}

	return apiObjects
}

func flattenTargetGroupTargetHealthStateAttributes(apiObjects []*elbv2.TargetGroupAttribute, protocol string) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{}

	switch protocol {
	case elbv2.ProtocolEnumTcp, elbv2.ProtocolEnumTls:
		for _, apiObject := range apiObjects {
			switch k, v := aws.StringValue(apiObject.Key), apiObject.Value; k {
			case targetGroupAttributeTargetHealthStateUnhealthyConnectionTerminationEnabled:
				tfMap["enable_unhealthy_connection_termination"] = flex.StringToBoolValue(v)
			}
		}
	}

	return tfMap
}

func targetGroupRuntimeValidation(d *schema.ResourceData, diags *diag.Diagnostics) {
	targetType := d.Get("target_type").(string)
	if targetType == elbv2.TargetTypeEnumLambda {
		if _, ok := d.GetOk("protocol"); ok {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				cty.GetAttrPath("protocol"),
				cty.GetAttrPath("target_type"),
				elbv2.TargetTypeEnumLambda,
			))
		}

		if _, ok := d.GetOk("protocol_version"); ok {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				cty.GetAttrPath("protocol_version"),
				cty.GetAttrPath("target_type"),
				elbv2.TargetTypeEnumLambda,
			))
		}

		if _, ok := d.GetOk(names.AttrPort); ok {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				cty.GetAttrPath(names.AttrPort),
				cty.GetAttrPath("target_type"),
				elbv2.TargetTypeEnumLambda,
			))
		}

		if _, ok := d.GetOk(names.AttrVPCID); ok {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				cty.GetAttrPath(names.AttrPort),
				cty.GetAttrPath("target_type"),
				elbv2.TargetTypeEnumLambda,
			))
		}

		if healthChecks := d.Get("health_check").([]interface{}); len(healthChecks) == 1 {
			healthCheck := healthChecks[0].(map[string]interface{})
			path := cty.GetAttrPath("health_check")

			if healthCheckProtocol := healthCheck["protocol"].(string); healthCheckProtocol != "" {
				*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
					path.GetAttr("protocol"),
					cty.GetAttrPath("target_type"),
					elbv2.TargetTypeEnumLambda,
				))
			}
		}
	} else {
		if _, ok := d.GetOk("protocol_version"); ok {
			protocol := d.Get("protocol").(string)
			switch protocol {
			case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
				// Noop
			default:
				*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
					cty.GetAttrPath("protocol_version"),
					cty.GetAttrPath("protocol"),
					protocol,
				))
			}
		}
	}
}
