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
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types;types.TargetGroup")
// @Testing(importIgnore="lambda_multi_value_headers_enabled;proxy_protocol_v2")
func resourceTargetGroup() *schema.Resource {
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
			names.AttrHealthCheck: {
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
						names.AttrInterval: {
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
						names.AttrPath: {
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
							DiffSuppressFunc: suppressIfTargetType(awstypes.TargetTypeEnumLambda),
						},
						names.AttrProtocol: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  awstypes.ProtocolEnumHttp,
							StateFunc: func(v interface{}) string {
								return strings.ToUpper(v.(string))
							},
							ValidateFunc:     validation.StringInSlice(healthCheckProtocolEnumValues(), true),
							DiffSuppressFunc: suppressIfTargetType(awstypes.TargetTypeEnumLambda),
						},
						names.AttrTimeout: {
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
			names.AttrIPAddressType: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TargetGroupIpAddressTypeEnum](),
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
				ConflictsWith: []string{names.AttrNamePrefix},
				ValidateFunc:  validTargetGroupName,
			},
			names.AttrNamePrefix: {
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
				DiffSuppressFunc: suppressIfTargetType(awstypes.TargetTypeEnumLambda),
			},
			"preserve_client_ip": {
				Type:             nullable.TypeNullableBool,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: nullable.DiffSuppressNullableBool,
				ValidateFunc:     nullable.ValidateTypeStringNullableBool,
			},
			names.AttrProtocol: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ProtocolEnum](),
				DiffSuppressFunc: suppressIfTargetType(awstypes.TargetTypeEnumLambda),
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
					if awstypes.TargetTypeEnum(d.Get("target_type").(string)) == awstypes.TargetTypeEnumLambda {
						return true
					}
					switch awstypes.ProtocolEnum(d.Get(names.AttrProtocol).(string)) {
					case awstypes.ProtocolEnumHttp, awstypes.ProtocolEnumHttps:
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
								switch awstypes.ProtocolEnum(d.Get(names.AttrProtocol).(string)) {
								case awstypes.ProtocolEnumTcp, awstypes.ProtocolEnumUdp, awstypes.ProtocolEnumTcpUdp, awstypes.ProtocolEnumTls, awstypes.ProtocolEnumGeneve:
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
			"target_group_health": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dns_failover": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"minimum_healthy_targets_count": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "off",
										ValidateFunc: validTargetGroupHealthInput,
									},
									"minimum_healthy_targets_percentage": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "off",
										ValidateFunc: validTargetGroupHealthPercentageInput,
									},
								},
							},
						},
						"unhealthy_state_routing": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"minimum_healthy_targets_count": {
										Type:     schema.TypeInt,
										Optional: true,
										Default:  1,
									},
									"minimum_healthy_targets_percentage": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      "off",
										ValidateFunc: validTargetGroupHealthPercentageInput,
									},
								},
							},
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
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.TargetTypeEnumInstance,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TargetTypeEnum](),
			},
			names.AttrVPCID: {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressIfTargetType(awstypes.TargetTypeEnumLambda),
			},
		},
	}
}

func suppressIfTargetType(t awstypes.TargetTypeEnum) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		// Don't suppress on creation, so that warnings are actually called
		if d.Id() == "" {
			return false
		}
		return awstypes.TargetTypeEnum(d.Get("target_type").(string)) == t
	}
}

func resourceTargetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	name := create.NewNameGenerator(
		create.WithConfiguredName(d.Get(names.AttrName).(string)),
		create.WithConfiguredPrefix(d.Get(names.AttrNamePrefix).(string)),
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

	protocol := awstypes.ProtocolEnum(d.Get(names.AttrProtocol).(string))
	targetType := awstypes.TargetTypeEnum(d.Get("target_type").(string))
	input := &elasticloadbalancingv2.CreateTargetGroupInput{
		Name:       aws.String(name),
		Tags:       getTagsIn(ctx),
		TargetType: targetType,
	}

	if targetType != awstypes.TargetTypeEnumLambda {
		input.Port = aws.Int32(int32(d.Get(names.AttrPort).(int)))
		input.Protocol = protocol
		switch protocol {
		case awstypes.ProtocolEnumHttp, awstypes.ProtocolEnumHttps:
			input.ProtocolVersion = aws.String(d.Get("protocol_version").(string))
		}
		input.VpcId = aws.String(d.Get(names.AttrVPCID).(string))

		switch targetType {
		case awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp:
			if v, ok := d.GetOk(names.AttrIPAddressType); ok {
				input.IpAddressType = awstypes.TargetGroupIpAddressTypeEnum(v.(string))
			}
		}
	}

	if v, ok := d.GetOk(names.AttrHealthCheck); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})

		input.HealthCheckEnabled = aws.Bool(tfMap[names.AttrEnabled].(bool))
		input.HealthCheckIntervalSeconds = aws.Int32(int32(tfMap[names.AttrInterval].(int)))
		input.HealthyThresholdCount = aws.Int32(int32(tfMap["healthy_threshold"].(int)))
		input.UnhealthyThresholdCount = aws.Int32(int32(tfMap["unhealthy_threshold"].(int)))

		if v, ok := tfMap[names.AttrTimeout].(int); ok && v != 0 {
			input.HealthCheckTimeoutSeconds = aws.Int32(int32(v))
		}

		healthCheckProtocol := awstypes.ProtocolEnum(tfMap[names.AttrProtocol].(string))
		if healthCheckProtocol != awstypes.ProtocolEnumTcp {
			if v, ok := tfMap[names.AttrPath].(string); ok && v != "" {
				input.HealthCheckPath = aws.String(v)
			}

			if v, ok := tfMap["matcher"].(string); ok && v != "" {
				if protocolVersion := d.Get("protocol_version").(string); protocolVersion == protocolVersionGRPC {
					input.Matcher = &awstypes.Matcher{
						GrpcCode: aws.String(v),
					}
				} else {
					input.Matcher = &awstypes.Matcher{
						HttpCode: aws.String(v),
					}
				}
			}
		}

		if targetType != awstypes.TargetTypeEnumLambda {
			input.HealthCheckPort = aws.String(tfMap[names.AttrPort].(string))
			input.HealthCheckProtocol = healthCheckProtocol
		}
	}

	output, err := conn.CreateTargetGroup(ctx, input)

	// Some partitions (e.g. ISO) may not support tag-on-create.
	partition := meta.(*conns.AWSClient).Partition
	if input.Tags != nil && errs.IsUnsupportedOperationInPartitionError(partition, err) {
		input.Tags = nil

		output, err = conn.CreateTargetGroup(ctx, input)
	}

	// Tags are not supported on creation with some protocol types(i.e. GENEVE)
	// Retry creation without tags
	if input.Tags != nil && tfawserr.ErrMessageContains(err, errCodeValidationError, tagsOnCreationErrMessage) {
		input.Tags = nil

		output, err = conn.CreateTargetGroup(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ELBv2 Target Group (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.TargetGroups[0].TargetGroupArn))

	_, err = tfresource.RetryWhenNotFound(ctx, elbv2PropagationTimeout, func() (interface{}, error) {
		return findTargetGroupByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Target Group (%s) create: %s", d.Id(), err)
	}

	var attributes []awstypes.TargetGroupAttribute

	switch targetType {
	case awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp:
		if v, ok := d.GetOk("stickiness"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandTargetGroupStickinessAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
		}

		if v, ok := d.GetOk("target_failover"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandTargetGroupTargetFailoverAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
		}

		if v, ok := d.GetOk("target_group_health"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandTargetGroupHealthAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
		}

		if v, ok := d.GetOk("target_health_state"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			attributes = append(attributes, expandTargetGroupTargetHealthStateAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
		}
	}

	attributes = append(attributes, targetGroupAttributes.expand(d, targetType, false)...)

	if len(attributes) > 0 {
		input := &elasticloadbalancingv2.ModifyTargetGroupAttributesInput{
			Attributes:     attributes,
			TargetGroupArn: aws.String(d.Id()),
		}

		_, err := conn.ModifyTargetGroupAttributes(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Target Group (%s) attributes: %s", d.Id(), err)
		}
	}

	// For partitions not supporting tag-on-create, attempt tag after create.
	if tags := getTagsIn(ctx); input.Tags == nil && len(tags) > 0 {
		err := createTags(ctx, conn, d.Id(), tags)

		// If default tags only, continue. Otherwise, error.
		if v, ok := d.GetOk(names.AttrTags); (!ok || len(v.(map[string]interface{})) == 0) && errs.IsUnsupportedOperationInPartitionError(partition, err) {
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
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	targetGroup, err := findTargetGroupByARN(ctx, conn, d.Id())

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
	if err := d.Set(names.AttrHealthCheck, flattenTargetGroupHealthCheck(targetGroup)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting health_check: %s", err)
	}
	d.Set(names.AttrIPAddressType, targetGroup.IpAddressType)
	d.Set("load_balancer_arns", flex.FlattenStringValueSet(targetGroup.LoadBalancerArns))
	d.Set(names.AttrName, targetGroup.TargetGroupName)
	d.Set(names.AttrNamePrefix, create.NamePrefixFromName(aws.ToString(targetGroup.TargetGroupName)))
	targetType := targetGroup.TargetType
	d.Set("target_type", targetType)

	if _, ok := d.GetOk(names.AttrPort); targetGroup.Port != nil || ok {
		d.Set(names.AttrPort, targetGroup.Port)
	}
	var protocol awstypes.ProtocolEnum
	if _, ok := d.GetOk(names.AttrProtocol); targetGroup.Protocol != "" || ok {
		protocol = targetGroup.Protocol
		d.Set(names.AttrProtocol, protocol)
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

	if err := d.Set("target_group_health", []interface{}{flattenTargetGroupHealthAttributes(attributes, protocol)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_group_health: %s", err)
	}

	if err := d.Set("target_health_state", []interface{}{flattenTargetGroupTargetHealthStateAttributes(attributes, protocol)}); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting target_health_state: %s", err)
	}

	targetGroupAttributes.flatten(d, targetType, attributes)

	return diags
}

func resourceTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	protocol := awstypes.ProtocolEnum(d.Get(names.AttrProtocol).(string))
	targetType := awstypes.TargetTypeEnum(d.Get("target_type").(string))

	if d.HasChange(names.AttrHealthCheck) {
		if v, ok := d.GetOk(names.AttrHealthCheck); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			input := &elasticloadbalancingv2.ModifyTargetGroupInput{
				HealthCheckEnabled:         aws.Bool(tfMap[names.AttrEnabled].(bool)),
				HealthCheckIntervalSeconds: aws.Int32(int32(tfMap[names.AttrInterval].(int))),
				HealthyThresholdCount:      aws.Int32(int32(tfMap["healthy_threshold"].(int))),
				TargetGroupArn:             aws.String(d.Id()),
				UnhealthyThresholdCount:    aws.Int32(int32(tfMap["unhealthy_threshold"].(int))),
			}

			if v, ok := tfMap[names.AttrTimeout].(int); ok && v != 0 {
				input.HealthCheckTimeoutSeconds = aws.Int32(int32(v))
			}

			healthCheckProtocol := awstypes.ProtocolEnum(tfMap[names.AttrProtocol].(string))
			if healthCheckProtocol != awstypes.ProtocolEnumTcp {
				if v, ok := tfMap["matcher"].(string); ok {
					if protocolVersion := d.Get("protocol_version").(string); protocolVersion == protocolVersionGRPC {
						input.Matcher = &awstypes.Matcher{
							GrpcCode: aws.String(v),
						}
					} else {
						input.Matcher = &awstypes.Matcher{
							HttpCode: aws.String(v),
						}
					}
				}
				input.HealthCheckPath = aws.String(tfMap[names.AttrPath].(string))
			}

			if targetType != awstypes.TargetTypeEnumLambda {
				input.HealthCheckPort = aws.String(tfMap[names.AttrPort].(string))
				input.HealthCheckProtocol = healthCheckProtocol
			}

			_, err := conn.ModifyTargetGroup(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Target Group (%s): %s", d.Id(), err)
			}
		}
	}

	var attributes []awstypes.TargetGroupAttribute

	switch targetType {
	case awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp:
		if d.HasChange("stickiness") {
			if v, ok := d.GetOk("stickiness"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				attributes = append(attributes, expandTargetGroupStickinessAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
			} else {
				attributes = append(attributes, awstypes.TargetGroupAttribute{
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

		if d.HasChange("target_group_health") {
			if v, ok := d.GetOk("target_group_health"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
				attributes = append(attributes, expandTargetGroupHealthAttributes(v.([]interface{})[0].(map[string]interface{}), protocol)...)
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
		input := &elasticloadbalancingv2.ModifyTargetGroupAttributesInput{
			Attributes:     attributes,
			TargetGroupArn: aws.String(d.Id()),
		}

		_, err := conn.ModifyTargetGroupAttributes(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying ELBv2 Target Group (%s) attributes: %s", d.Id(), err)
		}
	}

	return append(diags, resourceTargetGroupRead(ctx, d, meta)...)
}

func resourceTargetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Client(ctx)

	log.Printf("[DEBUG] Deleting ELBv2 Target Group: %s", d.Id())
	const (
		timeout = 2 * time.Minute
	)
	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.ResourceInUseException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteTargetGroup(ctx, &elasticloadbalancingv2.DeleteTargetGroupInput{
			TargetGroupArn: aws.String(d.Id()),
		})
	}, "is currently in use by a listener or a rule")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting ELBv2 Target Group (%s): %s", d.Id(), err)
	}

	return diags
}

type targetGroupAttributeInfo struct {
	apiAttributeKey      string
	tfType               schema.ValueType
	tfNullableType       schema.ValueType
	targetTypesSupported []awstypes.TargetTypeEnum
}

type targetGroupAttributeMap map[string]targetGroupAttributeInfo

var targetGroupAttributes = targetGroupAttributeMap(map[string]targetGroupAttributeInfo{
	"connection_termination": {
		apiAttributeKey:      targetGroupAttributeDeregistrationDelayConnectionTerminationEnabled,
		tfType:               schema.TypeBool,
		targetTypesSupported: []awstypes.TargetTypeEnum{awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp},
	},
	"deregistration_delay": {
		apiAttributeKey:      targetGroupAttributeDeregistrationDelayTimeoutSeconds,
		tfType:               schema.TypeString,
		tfNullableType:       schema.TypeInt,
		targetTypesSupported: []awstypes.TargetTypeEnum{awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp},
	},
	"lambda_multi_value_headers_enabled": {
		apiAttributeKey:      targetGroupAttributeLambdaMultiValueHeadersEnabled,
		tfType:               schema.TypeBool,
		targetTypesSupported: []awstypes.TargetTypeEnum{awstypes.TargetTypeEnumLambda},
	},
	"load_balancing_algorithm_type": {
		apiAttributeKey:      targetGroupAttributeLoadBalancingAlgorithmType,
		tfType:               schema.TypeString,
		targetTypesSupported: []awstypes.TargetTypeEnum{awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp},
	},
	"load_balancing_anomaly_mitigation": {
		apiAttributeKey:      targetGroupAttributeLoadBalancingAlgorithmAnomalyMitigation,
		tfType:               schema.TypeString,
		targetTypesSupported: []awstypes.TargetTypeEnum{awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp},
	},
	"load_balancing_cross_zone_enabled": {
		apiAttributeKey:      targetGroupAttributeLoadBalancingCrossZoneEnabled,
		tfType:               schema.TypeString,
		targetTypesSupported: []awstypes.TargetTypeEnum{awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp},
	},
	"preserve_client_ip": {
		apiAttributeKey:      targetGroupAttributePreserveClientIPEnabled,
		tfType:               schema.TypeString,
		tfNullableType:       schema.TypeBool,
		targetTypesSupported: []awstypes.TargetTypeEnum{awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp},
	},
	"proxy_protocol_v2": {
		apiAttributeKey:      targetGroupAttributeProxyProtocolV2Enabled,
		tfType:               schema.TypeBool,
		targetTypesSupported: []awstypes.TargetTypeEnum{awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp},
	},
	"slow_start": {
		apiAttributeKey:      targetGroupAttributeSlowStartDurationSeconds,
		tfType:               schema.TypeInt,
		targetTypesSupported: []awstypes.TargetTypeEnum{awstypes.TargetTypeEnumInstance, awstypes.TargetTypeEnumIp},
	},
})

func (m targetGroupAttributeMap) expand(d *schema.ResourceData, targetType awstypes.TargetTypeEnum, update bool) []awstypes.TargetGroupAttribute {
	var apiObjects []awstypes.TargetGroupAttribute

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
				apiObjects = append(apiObjects, awstypes.TargetGroupAttribute{
					Key:   k,
					Value: flex.BoolValueToString(v),
				})
			}
		case schema.TypeInt:
			v := v.(string)
			if v, null, _ := nullable.Int(v).ValueInt64(); !null {
				apiObjects = append(apiObjects, awstypes.TargetGroupAttribute{
					Key:   k,
					Value: flex.Int64ValueToString(v),
				})
			}
		default:
			switch attributeInfo.tfType {
			case schema.TypeBool:
				if v := v.(bool); v || update {
					apiObjects = append(apiObjects, awstypes.TargetGroupAttribute{
						Key:   k,
						Value: flex.BoolValueToString(v),
					})
				}
			case schema.TypeInt:
				if v := v.(int); v > 0 || update {
					apiObjects = append(apiObjects, awstypes.TargetGroupAttribute{
						Key:   k,
						Value: flex.IntValueToString(v),
					})
				}
			case schema.TypeString:
				if v := v.(string); v != "" || update {
					apiObjects = append(apiObjects, awstypes.TargetGroupAttribute{
						Key:   k,
						Value: aws.String(v),
					})
				}
			}
		}
	}

	return apiObjects
}

func (m targetGroupAttributeMap) flatten(d *schema.ResourceData, targetType awstypes.TargetTypeEnum, apiObjects []awstypes.TargetGroupAttribute) {
	for tfAttributeName, attributeInfo := range m {
		if !slices.Contains(attributeInfo.targetTypesSupported, targetType) {
			continue
		}

		k := attributeInfo.apiAttributeKey
		i := slices.IndexFunc(apiObjects, func(v awstypes.TargetGroupAttribute) bool {
			return aws.ToString(v.Key) == k
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

func findTargetGroupByARN(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string) (*awstypes.TargetGroup, error) {
	input := &elasticloadbalancingv2.DescribeTargetGroupsInput{
		TargetGroupArns: []string{arn},
	}

	output, err := findTargetGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TargetGroupArn) != arn {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTargetGroupByName(ctx context.Context, conn *elasticloadbalancingv2.Client, name string) (*awstypes.TargetGroup, error) {
	input := &elasticloadbalancingv2.DescribeTargetGroupsInput{
		Names: []string{name},
	}

	output, err := findTargetGroup(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	// Eventual consistency check.
	if aws.ToString(output.TargetGroupName) != name {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func findTargetGroup(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTargetGroupsInput) (*awstypes.TargetGroup, error) {
	output, err := findTargetGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findTargetGroups(ctx context.Context, conn *elasticloadbalancingv2.Client, input *elasticloadbalancingv2.DescribeTargetGroupsInput) ([]awstypes.TargetGroup, error) {
	var output []awstypes.TargetGroup

	pages := elasticloadbalancingv2.NewDescribeTargetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.TargetGroupNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.TargetGroups...)
	}

	return output, nil
}

func findTargetGroupAttributesByARN(ctx context.Context, conn *elasticloadbalancingv2.Client, arn string) ([]awstypes.TargetGroupAttribute, error) {
	input := &elasticloadbalancingv2.DescribeTargetGroupAttributesInput{
		TargetGroupArn: aws.String(arn),
	}

	output, err := conn.DescribeTargetGroupAttributes(ctx, input)

	if errs.IsA[*awstypes.TargetGroupNotFoundException](err) {
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
	if healthChecks := diff.Get(names.AttrHealthCheck).([]interface{}); len(healthChecks) == 1 {
		healthCheck = healthChecks[0].(map[string]interface{})
	}

	healthCheckPath := cty.GetAttrPath(names.AttrHealthCheck).IndexInt(0)

	if p, ok := healthCheck[names.AttrProtocol].(string); ok && strings.ToUpper(p) == string(awstypes.ProtocolEnumTcp) {
		if m := healthCheck["matcher"].(string); m != "" {
			return sdkdiag.DiagnosticError(errs.NewAttributeConflictsWhenError(
				healthCheckPath.GetAttr("matcher"),
				healthCheckPath.GetAttr(names.AttrProtocol),
				p,
			))
		}

		if m := healthCheck[names.AttrPath].(string); m != "" {
			return sdkdiag.DiagnosticError(errs.NewAttributeConflictsWhenError(
				healthCheckPath.GetAttr(names.AttrPath),
				healthCheckPath.GetAttr(names.AttrProtocol),
				p,
			))
		}
	}

	protocol := awstypes.ProtocolEnum(diff.Get(names.AttrProtocol).(string))

	switch protocol {
	case awstypes.ProtocolEnumHttp, awstypes.ProtocolEnumHttps:
		if p, ok := healthCheck[names.AttrProtocol].(string); ok && strings.ToUpper(p) == string(awstypes.ProtocolEnumTcp) {
			return fmt.Errorf("Attribute %q cannot have value %q when %q is %q.",
				errs.PathString(healthCheckPath.GetAttr(names.AttrProtocol)),
				awstypes.ProtocolEnumTcp,
				errs.PathString(cty.GetAttrPath(names.AttrProtocol)),
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
	targetType := awstypes.TargetTypeEnum(diff.Get("target_type").(string))
	if targetType != awstypes.TargetTypeEnumLambda {
		return nil
	}

	if healthChecks := diff.Get(names.AttrHealthCheck).([]interface{}); len(healthChecks) == 1 {
		healthCheck := healthChecks[0].(map[string]interface{})
		healtCheckPath := cty.GetAttrPath(names.AttrHealthCheck).IndexInt(0)
		healthCheckProtocol := awstypes.ProtocolEnum(healthCheck[names.AttrProtocol].(string))

		if healthCheckProtocol == awstypes.ProtocolEnumTcp {
			return fmt.Errorf("Attribute %q cannot have value %q when %q is %q.",
				errs.PathString(healtCheckPath.GetAttr(names.AttrProtocol)),
				awstypes.ProtocolEnumTcp,
				errs.PathString(cty.GetAttrPath("target_type")),
				targetType,
			)
		}
	}

	return nil
}

func customizeDiffTargetGroupTargetTypeNotLambda(_ context.Context, diff *schema.ResourceDiff, meta any) error {
	targetType := diff.Get("target_type").(string)
	if awstypes.TargetTypeEnum(targetType) == awstypes.TargetTypeEnumLambda {
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

	if v := config.GetAttr(names.AttrProtocol); v.IsKnown() && v.IsNull() {
		return sdkdiag.DiagnosticError(errs.NewAttributeRequiredWhenError(
			cty.GetAttrPath(names.AttrProtocol),
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

func flattenTargetGroupHealthCheck(apiObject *awstypes.TargetGroup) []interface{} {
	tfMap := map[string]interface{}{
		names.AttrEnabled:     aws.ToBool(apiObject.HealthCheckEnabled),
		"healthy_threshold":   aws.ToInt32(apiObject.HealthyThresholdCount),
		names.AttrInterval:    aws.ToInt32(apiObject.HealthCheckIntervalSeconds),
		names.AttrPort:        aws.ToString(apiObject.HealthCheckPort),
		names.AttrProtocol:    apiObject.HealthCheckProtocol,
		names.AttrTimeout:     aws.ToInt32(apiObject.HealthCheckTimeoutSeconds),
		"unhealthy_threshold": aws.ToInt32(apiObject.UnhealthyThresholdCount),
	}

	if v := apiObject.HealthCheckPath; v != nil {
		tfMap[names.AttrPath] = aws.ToString(v)
	}

	if apiObject := apiObject.Matcher; apiObject != nil {
		if v := apiObject.HttpCode; v != nil {
			tfMap["matcher"] = aws.ToString(v)
		}
		if v := apiObject.GrpcCode; v != nil {
			tfMap["matcher"] = aws.ToString(v)
		}
	}

	return []interface{}{tfMap}
}

func expandTargetGroupStickinessAttributes(tfMap map[string]interface{}, protocol awstypes.ProtocolEnum) []awstypes.TargetGroupAttribute {
	if tfMap == nil {
		return nil
	}

	apiObjects := []awstypes.TargetGroupAttribute{
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
	case awstypes.ProtocolEnumHttp, awstypes.ProtocolEnumHttps:
		switch stickinessType := tfMap[names.AttrType].(string); stickinessType {
		case stickinessTypeLBCookie:
			apiObjects = append(apiObjects,
				awstypes.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeStickinessLBCookieDurationSeconds),
					Value: flex.IntValueToString(tfMap["cookie_duration"].(int)),
				})
		case stickinessTypeAppCookie:
			apiObjects = append(apiObjects,
				awstypes.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeStickinessAppCookieCookieName),
					Value: aws.String(tfMap["cookie_name"].(string)),
				},
				awstypes.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeStickinessAppCookieDurationSeconds),
					Value: flex.IntValueToString(tfMap["cookie_duration"].(int)),
				})
		}
	}

	return apiObjects
}

func flattenTargetGroupStickinessAttributes(apiObjects []awstypes.TargetGroupAttribute, protocol awstypes.ProtocolEnum) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{}

	var stickinessType string
	for _, apiObject := range apiObjects {
		switch k, v := aws.ToString(apiObject.Key), apiObject.Value; k {
		case targetGroupAttributeStickinessEnabled:
			tfMap[names.AttrEnabled] = flex.StringToBoolValue(v)
		case targetGroupAttributeStickinessType:
			stickinessType = aws.ToString(v)
			tfMap[names.AttrType] = stickinessType
		}
	}

	switch protocol {
	case awstypes.ProtocolEnumHttp, awstypes.ProtocolEnumHttps:
		for _, apiObject := range apiObjects {
			k, v := aws.ToString(apiObject.Key), apiObject.Value
			switch {
			case k == targetGroupAttributeStickinessLBCookieDurationSeconds && stickinessType == stickinessTypeLBCookie:
				tfMap["cookie_duration"] = flex.StringToIntValue(v)
			case k == targetGroupAttributeStickinessAppCookieCookieName && stickinessType == stickinessTypeAppCookie:
				tfMap["cookie_name"] = aws.ToString(v)
			case k == targetGroupAttributeStickinessAppCookieDurationSeconds && stickinessType == stickinessTypeAppCookie:
				tfMap["cookie_duration"] = flex.StringToIntValue(v)
			}
		}
	}

	return tfMap
}

func expandTargetGroupTargetFailoverAttributes(tfMap map[string]interface{}, protocol awstypes.ProtocolEnum) []awstypes.TargetGroupAttribute {
	if tfMap == nil {
		return nil
	}

	var apiObjects []awstypes.TargetGroupAttribute

	switch protocol {
	case awstypes.ProtocolEnumGeneve:
		apiObjects = append(apiObjects,
			awstypes.TargetGroupAttribute{
				Key:   aws.String(targetGroupAttributeTargetFailoverOnDeregistration),
				Value: aws.String(tfMap["on_deregistration"].(string)),
			},
			awstypes.TargetGroupAttribute{
				Key:   aws.String(targetGroupAttributeTargetFailoverOnUnhealthy),
				Value: aws.String(tfMap["on_unhealthy"].(string)),
			})
	}

	return apiObjects
}

func flattenTargetGroupTargetFailoverAttributes(apiObjects []awstypes.TargetGroupAttribute, protocol awstypes.ProtocolEnum) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{}

	switch protocol {
	case awstypes.ProtocolEnumGeneve:
		for _, apiObject := range apiObjects {
			switch k, v := aws.ToString(apiObject.Key), apiObject.Value; k {
			case targetGroupAttributeTargetFailoverOnDeregistration:
				tfMap["on_deregistration"] = aws.ToString(v)
			case targetGroupAttributeTargetFailoverOnUnhealthy:
				tfMap["on_unhealthy"] = aws.ToString(v)
			}
		}
	}

	return tfMap
}

func expandTargetGroupTargetHealthStateAttributes(tfMap map[string]interface{}, protocol awstypes.ProtocolEnum) []awstypes.TargetGroupAttribute {
	if tfMap == nil {
		return nil
	}

	var apiObjects []awstypes.TargetGroupAttribute

	switch protocol {
	case awstypes.ProtocolEnumTcp, awstypes.ProtocolEnumTls:
		apiObjects = append(apiObjects,
			awstypes.TargetGroupAttribute{
				Key:   aws.String(targetGroupAttributeTargetHealthStateUnhealthyConnectionTerminationEnabled),
				Value: flex.BoolValueToString(tfMap["enable_unhealthy_connection_termination"].(bool)),
			})
	}

	return apiObjects
}

func flattenTargetGroupTargetHealthStateAttributes(apiObjects []awstypes.TargetGroupAttribute, protocol awstypes.ProtocolEnum) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{}

	switch protocol {
	case awstypes.ProtocolEnumTcp, awstypes.ProtocolEnumTls:
		for _, apiObject := range apiObjects {
			switch k, v := aws.ToString(apiObject.Key), apiObject.Value; k {
			case targetGroupAttributeTargetHealthStateUnhealthyConnectionTerminationEnabled:
				tfMap["enable_unhealthy_connection_termination"] = flex.StringToBoolValue(v)
			}
		}
	}

	return tfMap
}

func expandTargetGroupHealthAttributes(tfMap map[string]interface{}, protocol awstypes.ProtocolEnum) []awstypes.TargetGroupAttribute {
	if tfMap == nil {
		return nil
	}

	var apiObjects []awstypes.TargetGroupAttribute

	// Supported on Application Load Balancers and Network Load Balancers.
	switch protocol {
	case awstypes.ProtocolEnumGeneve:
	default:
		if v, ok := tfMap["dns_failover"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})
			apiObjects = append(apiObjects,
				awstypes.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeTargetGroupHealthDNSFailoverMinimumHealthyTargetsCount),
					Value: aws.String(tfMap["minimum_healthy_targets_count"].(string)),
				},
				awstypes.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeTargetGroupHealthDNSFailoverMinimumHealthyTargetsPercentage),
					Value: aws.String(tfMap["minimum_healthy_targets_percentage"].(string)),
				},
			)
		}

		if v, ok := tfMap["unhealthy_state_routing"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
			tfMap := v[0].(map[string]interface{})
			apiObjects = append(apiObjects,
				awstypes.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeTargetGroupHealthUnhealthyStateRoutingMinimumHealthyTargetsCount),
					Value: flex.IntValueToString(tfMap["minimum_healthy_targets_count"].(int)),
				},
				awstypes.TargetGroupAttribute{
					Key:   aws.String(targetGroupAttributeTargetGroupHealthUnhealthyStateRoutingMinimumHealthyTargetsPercentage),
					Value: aws.String(tfMap["minimum_healthy_targets_percentage"].(string)),
				},
			)
		}
	}

	return apiObjects
}

func flattenTargetGroupHealthAttributes(apiObjects []awstypes.TargetGroupAttribute, protocol awstypes.ProtocolEnum) map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	tfMap := map[string]interface{}{}
	dnsFailoverMap := make(map[string]interface{})
	unhealthyStateRoutingMap := make(map[string]interface{})

	// Supported on Application Load Balancers and Network Load Balancers.
	switch protocol {
	case awstypes.ProtocolEnumGeneve:
	default:
		for _, apiObject := range apiObjects {
			switch k, v := aws.ToString(apiObject.Key), apiObject.Value; k {
			case targetGroupAttributeTargetGroupHealthDNSFailoverMinimumHealthyTargetsCount:
				dnsFailoverMap["minimum_healthy_targets_count"] = aws.ToString(v)
			case targetGroupAttributeTargetGroupHealthDNSFailoverMinimumHealthyTargetsPercentage:
				dnsFailoverMap["minimum_healthy_targets_percentage"] = aws.ToString(v)
			case targetGroupAttributeTargetGroupHealthUnhealthyStateRoutingMinimumHealthyTargetsCount:
				unhealthyStateRoutingMap["minimum_healthy_targets_count"] = flex.StringToIntValue(v)
			case targetGroupAttributeTargetGroupHealthUnhealthyStateRoutingMinimumHealthyTargetsPercentage:
				unhealthyStateRoutingMap["minimum_healthy_targets_percentage"] = aws.ToString(v)
			}
		}
	}

	tfMap["dns_failover"] = []interface{}{dnsFailoverMap}
	tfMap["unhealthy_state_routing"] = []interface{}{unhealthyStateRoutingMap}

	return tfMap
}

func targetGroupRuntimeValidation(d *schema.ResourceData, diags *diag.Diagnostics) {
	if targetType := awstypes.TargetTypeEnum(d.Get("target_type").(string)); targetType == awstypes.TargetTypeEnumLambda {
		targetType := string(targetType)
		if _, ok := d.GetOk(names.AttrProtocol); ok {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				cty.GetAttrPath(names.AttrProtocol),
				cty.GetAttrPath("target_type"),
				targetType,
			))
		}

		if _, ok := d.GetOk("protocol_version"); ok {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				cty.GetAttrPath("protocol_version"),
				cty.GetAttrPath("target_type"),
				targetType,
			))
		}

		if _, ok := d.GetOk(names.AttrPort); ok {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				cty.GetAttrPath(names.AttrPort),
				cty.GetAttrPath("target_type"),
				targetType,
			))
		}

		if _, ok := d.GetOk(names.AttrVPCID); ok {
			*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
				cty.GetAttrPath(names.AttrPort),
				cty.GetAttrPath("target_type"),
				targetType,
			))
		}

		if healthChecks := d.Get(names.AttrHealthCheck).([]interface{}); len(healthChecks) == 1 {
			healthCheck := healthChecks[0].(map[string]interface{})
			path := cty.GetAttrPath(names.AttrHealthCheck)

			if healthCheckProtocol := healthCheck[names.AttrProtocol].(string); healthCheckProtocol != "" {
				*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
					path.GetAttr(names.AttrProtocol),
					cty.GetAttrPath("target_type"),
					targetType,
				))
			}
		}
	} else {
		if _, ok := d.GetOk("protocol_version"); ok {
			protocol := awstypes.ProtocolEnum(d.Get(names.AttrProtocol).(string))
			switch protocol {
			case awstypes.ProtocolEnumHttp, awstypes.ProtocolEnumHttps:
				// Noop
			default:
				protocol := string(protocol)
				*diags = append(*diags, errs.NewAttributeConflictsWhenWillBeError(
					cty.GetAttrPath("protocol_version"),
					cty.GetAttrPath(names.AttrProtocol),
					protocol,
				))
			}
		}
	}
}
