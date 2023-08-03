// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elbv2

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	propagationTimeout = 2 * time.Minute
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

		// NLBs have restrictions on them at this time
		CustomizeDiff: customdiff.Sequence(
			resourceTargetGroupCustomizeDiff,
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"arn": {
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
				Default:  false,
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
						"enabled": {
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
							Type:     schema.TypeInt,
							Optional: true,
							Default:  30,
						},
						"matcher": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
						},
						"path": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validTargetGroupHealthCheckPath,
						},
						"port": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          "traffic-port",
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
							ValidateFunc: validation.StringInSlice([]string{
								elbv2.ProtocolEnumHttp,
								elbv2.ProtocolEnumHttps,
								elbv2.ProtocolEnumTcp,
							}, true),
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
			"load_balancing_algorithm_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"round_robin",
					"least_outstanding_requests",
				}, false),
			},
			"load_balancing_cross_zone_enabled": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"true",
					"false",
					"use_load_balancer_configuration",
				}, false),
			},
			"name": {
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
				ForceNew:      true,
				ConflictsWith: []string{"name"},
				ValidateFunc:  validTargetGroupNamePrefix,
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntBetween(1, 65535),
			},
			"preserve_client_ip": {
				Type:             nullable.TypeNullableBool,
				Optional:         true,
				Computed:         true,
				DiffSuppressFunc: nullable.DiffSuppressNullableBool,
				ValidateFunc:     nullable.ValidateTypeStringNullableBool,
			},
			"protocol": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(elbv2.ProtocolEnum_Values(), true),
			},
			"protocol_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return strings.ToUpper(v.(string))
				},
				ValidateFunc: validation.StringInSlice([]string{
					"GRPC",
					"HTTP1",
					"HTTP2",
				}, true),
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
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
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validateSlowStart,
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
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"lb_cookie",               // Only for ALBs
								"app_cookie",              // Only for ALBs
								"source_ip",               // Only for NLBs
								"source_ip_dest_ip",       // Only for GWLBs
								"source_ip_dest_ip_proto", // Only for GWLBs
							}, false),
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
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"rebalance",
								"no_rebalance",
							}, false),
						},
						"on_unhealthy": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"rebalance",
								"no_rebalance",
							}, false),
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
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func suppressIfTargetType(t string) schema.SchemaDiffSuppressFunc {
	return func(k string, old string, new string, d *schema.ResourceData) bool {
		return d.Get("target_type").(string) == t
	}
}

func resourceTargetGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	var groupName string
	if v, ok := d.GetOk("name"); ok {
		groupName = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		groupName = id.PrefixedUniqueId(v.(string))
	} else {
		groupName = id.PrefixedUniqueId("tf-")
	}

	existingTg, err := FindTargetGroupByName(ctx, conn, groupName)

	if err != nil && !tfresource.NotFound(err) {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Group (%s): %s", groupName, err)
	}

	if existingTg != nil {
		return sdkdiag.AppendErrorf(diags, "ELBv2 Target Group (%s) already exists", groupName)
	}

	input := &elbv2.CreateTargetGroupInput{
		Name:       aws.String(groupName),
		Tags:       getTagsIn(ctx),
		TargetType: aws.String(d.Get("target_type").(string)),
	}

	if d.Get("target_type").(string) != elbv2.TargetTypeEnumLambda {
		if _, ok := d.GetOk("port"); !ok {
			return sdkdiag.AppendErrorf(diags, "port should be set when target type is %s", d.Get("target_type").(string))
		}

		if _, ok := d.GetOk("protocol"); !ok {
			return sdkdiag.AppendErrorf(diags, "protocol should be set when target type is %s", d.Get("target_type").(string))
		}

		if _, ok := d.GetOk("vpc_id"); !ok {
			return sdkdiag.AppendErrorf(diags, "vpc_id should be set when target type is %s", d.Get("target_type").(string))
		}
		input.Port = aws.Int64(int64(d.Get("port").(int)))
		input.Protocol = aws.String(d.Get("protocol").(string))
		switch d.Get("protocol").(string) {
		case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
			input.ProtocolVersion = aws.String(d.Get("protocol_version").(string))
		}
		input.VpcId = aws.String(d.Get("vpc_id").(string))

		if d.Get("target_type").(string) == elbv2.TargetTypeEnumIp {
			if _, ok := d.GetOk("ip_address_type"); ok {
				input.IpAddressType = aws.String(d.Get("ip_address_type").(string))
			}
		}
	}

	if healthChecks := d.Get("health_check").([]interface{}); len(healthChecks) == 1 {
		healthCheck := healthChecks[0].(map[string]interface{})

		input.HealthCheckEnabled = aws.Bool(healthCheck["enabled"].(bool))

		input.HealthCheckIntervalSeconds = aws.Int64(int64(healthCheck["interval"].(int)))

		input.HealthyThresholdCount = aws.Int64(int64(healthCheck["healthy_threshold"].(int)))
		input.UnhealthyThresholdCount = aws.Int64(int64(healthCheck["unhealthy_threshold"].(int)))
		t := healthCheck["timeout"].(int)
		if t != 0 {
			input.HealthCheckTimeoutSeconds = aws.Int64(int64(t))
		}
		healthCheckProtocol := healthCheck["protocol"].(string)

		if healthCheckProtocol != elbv2.ProtocolEnumTcp {
			p := healthCheck["path"].(string)
			if p != "" {
				input.HealthCheckPath = aws.String(p)
			}

			m := healthCheck["matcher"].(string)
			protocolVersion := d.Get("protocol_version").(string)
			if m != "" {
				if protocolVersion == "GRPC" {
					input.Matcher = &elbv2.Matcher{
						GrpcCode: aws.String(m),
					}
				} else {
					input.Matcher = &elbv2.Matcher{
						HttpCode: aws.String(m),
					}
				}
			}
		}
		if d.Get("target_type").(string) != elbv2.TargetTypeEnumLambda {
			input.HealthCheckPort = aws.String(healthCheck["port"].(string))
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
	if input.Tags != nil && tfawserr.ErrMessageContains(err, ErrValidationError, TagsOnCreationErrMessage) {
		input.Tags = nil

		output, err = conn.CreateTargetGroupWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating LB Target Group: %s", err)
	}

	if len(output.TargetGroups) == 0 {
		return sdkdiag.AppendErrorf(diags, "creating LB Target Group: no groups returned in response")
	}

	d.SetId(aws.StringValue(output.TargetGroups[0].TargetGroupArn))

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTargetGroupByARN(ctx, conn, d.Id())
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for ELBv2 Target Group (%s) create: %s", d.Id(), err)
	}

	var attrs []*elbv2.TargetGroupAttribute

	switch d.Get("target_type").(string) {
	case elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp:
		if v, null, _ := nullable.Int(d.Get("deregistration_delay").(string)).Value(); !null {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("deregistration_delay.timeout_seconds"),
				Value: aws.String(fmt.Sprintf("%d", v)),
			})
		}

		if v, ok := d.GetOk("load_balancing_algorithm_type"); ok {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("load_balancing.algorithm.type"),
				Value: aws.String(v.(string)),
			})
		}

		if v, ok := d.GetOk("load_balancing_cross_zone_enabled"); ok {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("load_balancing.cross_zone.enabled"),
				Value: aws.String(v.(string)),
			})
		}

		if v, ok := d.GetOk("preserve_client_ip"); ok {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("preserve_client_ip.enabled"),
				Value: aws.String(v.(string)),
			})
		}

		if v, ok := d.GetOk("proxy_protocol_v2"); ok {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("proxy_protocol_v2.enabled"),
				Value: aws.String(strconv.FormatBool(v.(bool))),
			})
		}

		if v, ok := d.GetOk("connection_termination"); ok {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("deregistration_delay.connection_termination.enabled"),
				Value: aws.String(strconv.FormatBool(v.(bool))),
			})
		}

		if v, ok := d.GetOk("slow_start"); ok {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("slow_start.duration_seconds"),
				Value: aws.String(fmt.Sprintf("%d", v.(int))),
			})
		}

		// Only supported for GWLB
		if v, ok := d.Get("protocol").(string); ok && v == elbv2.ProtocolEnumGeneve {
			if v, ok := d.GetOk("target_failover"); ok {
				failoverBlock := v.([]interface{})
				failover := failoverBlock[0].(map[string]interface{})
				attrs = append(attrs,
					&elbv2.TargetGroupAttribute{
						Key:   aws.String("target_failover.on_deregistration"),
						Value: aws.String(failover["on_deregistration"].(string)),
					},
					&elbv2.TargetGroupAttribute{
						Key:   aws.String("target_failover.on_unhealthy"),
						Value: aws.String(failover["on_unhealthy"].(string)),
					},
				)
			}
		}

		if v, ok := d.GetOk("stickiness"); ok && len(v.([]interface{})) > 0 {
			stickinessBlocks := v.([]interface{})
			stickiness := stickinessBlocks[0].(map[string]interface{})

			attrs = append(attrs,
				&elbv2.TargetGroupAttribute{
					Key:   aws.String("stickiness.enabled"),
					Value: aws.String(strconv.FormatBool(stickiness["enabled"].(bool))),
				},
				&elbv2.TargetGroupAttribute{
					Key:   aws.String("stickiness.type"),
					Value: aws.String(stickiness["type"].(string)),
				})

			switch d.Get("protocol").(string) {
			case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
				switch stickiness["type"].(string) {
				case "lb_cookie":
					attrs = append(attrs,
						&elbv2.TargetGroupAttribute{
							Key:   aws.String("stickiness.lb_cookie.duration_seconds"),
							Value: aws.String(fmt.Sprintf("%d", stickiness["cookie_duration"].(int))),
						})
				case "app_cookie":
					attrs = append(attrs,
						&elbv2.TargetGroupAttribute{
							Key:   aws.String("stickiness.app_cookie.duration_seconds"),
							Value: aws.String(fmt.Sprintf("%d", stickiness["cookie_duration"].(int))),
						},
						&elbv2.TargetGroupAttribute{
							Key:   aws.String("stickiness.app_cookie.cookie_name"),
							Value: aws.String(stickiness["cookie_name"].(string)),
						})
				default:
					log.Printf("[WARN] Unexpected stickiness type. Expected lb_cookie or app_cookie, got %s", stickiness["type"].(string))
				}
			}
		}
	case elbv2.TargetTypeEnumLambda:
		if v, ok := d.GetOk("lambda_multi_value_headers_enabled"); ok {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("lambda.multi_value_headers.enabled"),
				Value: aws.String(strconv.FormatBool(v.(bool))),
			})
		}
	}

	if len(attrs) > 0 {
		params := &elbv2.ModifyTargetGroupAttributesInput{
			TargetGroupArn: aws.String(d.Id()),
			Attributes:     attrs,
		}

		_, err := conn.ModifyTargetGroupAttributesWithContext(ctx, params)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Target Group Attributes: %s", err)
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

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindTargetGroupByARN(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] ELBv2 Target Group %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading ELBv2 Target Group (%s): %s", d.Id(), err)
	}

	if err := flattenTargetGroupResource(ctx, d, meta, outputRaw.(*elbv2.TargetGroup)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	return diags
}

func resourceTargetGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	if d.HasChange("health_check") {
		var params *elbv2.ModifyTargetGroupInput
		healthChecks := d.Get("health_check").([]interface{})
		if len(healthChecks) == 1 {
			healthCheck := healthChecks[0].(map[string]interface{})

			params = &elbv2.ModifyTargetGroupInput{
				TargetGroupArn:             aws.String(d.Id()),
				HealthCheckEnabled:         aws.Bool(healthCheck["enabled"].(bool)),
				HealthCheckIntervalSeconds: aws.Int64(int64(healthCheck["interval"].(int))),
				HealthyThresholdCount:      aws.Int64(int64(healthCheck["healthy_threshold"].(int))),
				UnhealthyThresholdCount:    aws.Int64(int64(healthCheck["unhealthy_threshold"].(int))),
			}

			t := healthCheck["timeout"].(int)
			if t != 0 {
				params.HealthCheckTimeoutSeconds = aws.Int64(int64(t))
			}

			healthCheckProtocol := healthCheck["protocol"].(string)
			protocolVersion := d.Get("protocol_version").(string)
			if healthCheckProtocol != elbv2.ProtocolEnumTcp && !d.IsNewResource() {
				if protocolVersion == "GRPC" {
					params.Matcher = &elbv2.Matcher{
						GrpcCode: aws.String(healthCheck["matcher"].(string)),
					}
				} else {
					params.Matcher = &elbv2.Matcher{
						HttpCode: aws.String(healthCheck["matcher"].(string)),
					}
				}
				params.HealthCheckPath = aws.String(healthCheck["path"].(string))
			}
			if d.Get("target_type").(string) != elbv2.TargetTypeEnumLambda {
				params.HealthCheckPort = aws.String(healthCheck["port"].(string))
				params.HealthCheckProtocol = aws.String(healthCheckProtocol)
			}
		}

		if params != nil {
			_, err := conn.ModifyTargetGroupWithContext(ctx, params)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "modifying Target Group: %s", err)
			}
		}
	}

	var attrs []*elbv2.TargetGroupAttribute

	switch d.Get("target_type").(string) {
	case elbv2.TargetTypeEnumInstance, elbv2.TargetTypeEnumIp:
		if d.HasChange("deregistration_delay") {
			if v, null, _ := nullable.Int(d.Get("deregistration_delay").(string)).Value(); !null {
				attrs = append(attrs, &elbv2.TargetGroupAttribute{
					Key:   aws.String("deregistration_delay.timeout_seconds"),
					Value: aws.String(fmt.Sprintf("%d", v)),
				})
			}
		}

		if d.HasChange("slow_start") {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("slow_start.duration_seconds"),
				Value: aws.String(fmt.Sprintf("%d", d.Get("slow_start").(int))),
			})
		}

		if d.HasChange("proxy_protocol_v2") {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("proxy_protocol_v2.enabled"),
				Value: aws.String(strconv.FormatBool(d.Get("proxy_protocol_v2").(bool))),
			})
		}

		if d.HasChange("connection_termination") {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("deregistration_delay.connection_termination.enabled"),
				Value: aws.String(strconv.FormatBool(d.Get("connection_termination").(bool))),
			})
		}

		if d.HasChange("preserve_client_ip") {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("preserve_client_ip.enabled"),
				Value: aws.String(d.Get("preserve_client_ip").(string)),
			})
		}

		if d.HasChange("stickiness") {
			stickinessBlocks := d.Get("stickiness").([]interface{})
			if len(stickinessBlocks) == 1 {
				stickiness := stickinessBlocks[0].(map[string]interface{})
				attrs = append(attrs,
					&elbv2.TargetGroupAttribute{
						Key:   aws.String("stickiness.enabled"),
						Value: aws.String(strconv.FormatBool(stickiness["enabled"].(bool))),
					},
					&elbv2.TargetGroupAttribute{
						Key:   aws.String("stickiness.type"),
						Value: aws.String(stickiness["type"].(string)),
					})

				switch d.Get("protocol").(string) {
				case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
					switch stickiness["type"].(string) {
					case "lb_cookie":
						attrs = append(attrs,
							&elbv2.TargetGroupAttribute{
								Key:   aws.String("stickiness.lb_cookie.duration_seconds"),
								Value: aws.String(fmt.Sprintf("%d", stickiness["cookie_duration"].(int))),
							})
					case "app_cookie":
						attrs = append(attrs,
							&elbv2.TargetGroupAttribute{
								Key:   aws.String("stickiness.app_cookie.duration_seconds"),
								Value: aws.String(fmt.Sprintf("%d", stickiness["cookie_duration"].(int))),
							},
							&elbv2.TargetGroupAttribute{
								Key:   aws.String("stickiness.app_cookie.cookie_name"),
								Value: aws.String(stickiness["cookie_name"].(string)),
							})
					default:
						log.Printf("[WARN] Unexpected stickiness type. Expected lb_cookie or app_cookie, got %s", stickiness["type"].(string))
					}
				}
			} else if len(stickinessBlocks) == 0 {
				attrs = append(attrs, &elbv2.TargetGroupAttribute{
					Key:   aws.String("stickiness.enabled"),
					Value: aws.String("false"),
				})
			}
		}

		if d.HasChange("load_balancing_algorithm_type") {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("load_balancing.algorithm.type"),
				Value: aws.String(d.Get("load_balancing_algorithm_type").(string)),
			})
		}

		if d.HasChange("load_balancing_cross_zone_enabled") {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("load_balancing.cross_zone.enabled"),
				Value: aws.String(d.Get("load_balancing_cross_zone_enabled").(string)),
			})
		}

		if d.HasChange("target_failover") {
			failoverBlock := d.Get("target_failover").([]interface{})
			if len(failoverBlock) == 1 {
				failover := failoverBlock[0].(map[string]interface{})
				attrs = append(attrs,
					&elbv2.TargetGroupAttribute{
						Key:   aws.String("target_failover.on_deregistration"),
						Value: aws.String(failover["on_deregistration"].(string)),
					},
					&elbv2.TargetGroupAttribute{
						Key:   aws.String("target_failover.on_unhealthy"),
						Value: aws.String(failover["on_unhealthy"].(string)),
					},
				)
			}
		}

	case elbv2.TargetTypeEnumLambda:
		if d.HasChange("lambda_multi_value_headers_enabled") {
			attrs = append(attrs, &elbv2.TargetGroupAttribute{
				Key:   aws.String("lambda.multi_value_headers.enabled"),
				Value: aws.String(strconv.FormatBool(d.Get("lambda_multi_value_headers_enabled").(bool))),
			})
		}
	}

	if len(attrs) > 0 {
		params := &elbv2.ModifyTargetGroupAttributesInput{
			TargetGroupArn: aws.String(d.Id()),
			Attributes:     attrs,
		}

		_, err := conn.ModifyTargetGroupAttributesWithContext(ctx, params)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying Target Group Attributes: %s", err)
		}
	}

	return append(diags, resourceTargetGroupRead(ctx, d, meta)...)
}

func resourceTargetGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	const (
		targetGroupDeleteTimeout = 2 * time.Minute
	)
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	input := &elbv2.DeleteTargetGroupInput{
		TargetGroupArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Target Group (%s): %s", d.Id(), input)
	err := retry.RetryContext(ctx, targetGroupDeleteTimeout, func() *retry.RetryError {
		_, err := conn.DeleteTargetGroupWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, "ResourceInUse", "is currently in use by a listener or a rule") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteTargetGroupWithContext(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Target Group: %s", err)
	}

	return diags
}

func FindTargetGroupByARN(ctx context.Context, conn *elbv2.ELBV2, arn string) (*elbv2.TargetGroup, error) {
	input := &elbv2.DescribeTargetGroupsInput{
		TargetGroupArns: aws.StringSlice([]string{arn}),
	}

	output, err := FindTargetGroup(ctx, conn, input)

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

func FindTargetGroupByName(ctx context.Context, conn *elbv2.ELBV2, name string) (*elbv2.TargetGroup, error) {
	input := &elbv2.DescribeTargetGroupsInput{
		Names: aws.StringSlice([]string{name}),
	}

	output, err := FindTargetGroup(ctx, conn, input)

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

func FindTargetGroups(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTargetGroupsInput) ([]*elbv2.TargetGroup, error) {
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

func FindTargetGroup(ctx context.Context, conn *elbv2.ELBV2, input *elbv2.DescribeTargetGroupsInput) (*elbv2.TargetGroup, error) {
	output, err := FindTargetGroups(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 || output[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return output[0], nil
}

func validTargetGroupHealthCheckPath(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 1024 {
		errors = append(errors, fmt.Errorf(
			"%q cannot be longer than 1024 characters: %q", k, value))
	}
	if len(value) > 0 && !strings.HasPrefix(value, "/") {
		errors = append(errors, fmt.Errorf(
			"%q must begin with a '/' character: %q", k, value))
	}
	return
}

func validateSlowStart(v interface{}, k string) (ws []string, errors []error) {
	value := v.(int)

	// Check if the value is between 30-900 or 0 (seconds).
	if value != 0 && !(value >= 30 && value <= 900) {
		errors = append(errors, fmt.Errorf(
			"%q contains an invalid Slow Start Duration \"%d\". "+
				"Valid intervals are 30-900 or 0 to disable.",
			k, value))
	}
	return
}

func validTargetGroupHealthCheckPort(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)

	if value == "traffic-port" {
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

	if arnComponents := regexp.MustCompile(`arn:.*:targetgroup/(.*)`).FindAllStringSubmatch(*arn, -1); len(arnComponents) == 1 {
		if len(arnComponents[0]) == 2 {
			return fmt.Sprintf("targetgroup/%s", arnComponents[0][1])
		}
	}

	return ""
}

// flattenTargetGroupResource takes a *elbv2.TargetGroup and populates all respective resource fields.
func flattenTargetGroupResource(ctx context.Context, d *schema.ResourceData, meta interface{}, targetGroup *elbv2.TargetGroup) error {
	conn := meta.(*conns.AWSClient).ELBV2Conn(ctx)

	d.Set("arn", targetGroup.TargetGroupArn)
	d.Set("arn_suffix", TargetGroupSuffixFromARN(targetGroup.TargetGroupArn))
	d.Set("name", targetGroup.TargetGroupName)
	d.Set("target_type", targetGroup.TargetType)
	d.Set("ip_address_type", targetGroup.IpAddressType)

	if err := d.Set("health_check", flattenLbTargetGroupHealthCheck(targetGroup)); err != nil {
		return fmt.Errorf("setting health_check: %w", err)
	}

	if v, _ := d.Get("target_type").(string); v != elbv2.TargetTypeEnumLambda {
		d.Set("vpc_id", targetGroup.VpcId)
		d.Set("port", targetGroup.Port)
		d.Set("protocol", targetGroup.Protocol)
	}

	switch d.Get("protocol").(string) {
	case elbv2.ProtocolEnumHttp, elbv2.ProtocolEnumHttps:
		d.Set("protocol_version", targetGroup.ProtocolVersion)
	}

	attrResp, err := conn.DescribeTargetGroupAttributesWithContext(ctx, &elbv2.DescribeTargetGroupAttributesInput{
		TargetGroupArn: aws.String(d.Id()),
	})
	if err != nil {
		return fmt.Errorf("retrieving Target Group Attributes: %w", err)
	}

	for _, attr := range attrResp.Attributes {
		switch aws.StringValue(attr.Key) {
		case "deregistration_delay.timeout_seconds":
			d.Set("deregistration_delay", attr.Value)
		case "lambda.multi_value_headers.enabled":
			enabled, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("converting lambda.multi_value_headers.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("lambda_multi_value_headers_enabled", enabled)
		case "proxy_protocol_v2.enabled":
			enabled, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("converting proxy_protocol_v2.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("proxy_protocol_v2", enabled)
		case "deregistration_delay.connection_termination.enabled":
			enabled, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("converting deregistration_delay.connection_termination.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("connection_termination", enabled)
		case "slow_start.duration_seconds":
			slowStart, err := strconv.Atoi(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("converting slow_start.duration_seconds to int: %s", aws.StringValue(attr.Value))
			}
			d.Set("slow_start", slowStart)
		case "load_balancing.algorithm.type":
			loadBalancingAlgorithm := aws.StringValue(attr.Value)
			d.Set("load_balancing_algorithm_type", loadBalancingAlgorithm)
		case "load_balancing.cross_zone.enabled":
			loadBalancingCrossZoneEnabled := aws.StringValue(attr.Value)
			d.Set("load_balancing_cross_zone_enabled", loadBalancingCrossZoneEnabled)
		case "preserve_client_ip.enabled":
			_, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return fmt.Errorf("converting preserve_client_ip.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			d.Set("preserve_client_ip", attr.Value)
		}
	}

	stickinessAttr, err := flattenTargetGroupStickiness(attrResp.Attributes)
	if err != nil {
		return fmt.Errorf("flattening stickiness: %w", err)
	}

	if err := d.Set("stickiness", stickinessAttr); err != nil {
		return fmt.Errorf("setting stickiness: %w", err)
	}

	// Set target failover attributes for GWLB
	targetFailoverAttr := flattenTargetGroupFailover(attrResp.Attributes)
	if err != nil {
		return fmt.Errorf("flattening target failover: %w", err)
	}

	if err := d.Set("target_failover", targetFailoverAttr); err != nil {
		return fmt.Errorf("setting target failover: %w", err)
	}

	return nil
}

func flattenTargetGroupFailover(attributes []*elbv2.TargetGroupAttribute) []interface{} {
	if len(attributes) == 0 {
		return []interface{}{}
	}

	m := make(map[string]interface{})

	for _, attr := range attributes {
		switch aws.StringValue(attr.Key) {
		case "target_failover.on_deregistration":
			m["on_deregistration"] = aws.StringValue(attr.Value)
		case "target_failover.on_unhealthy":
			m["on_unhealthy"] = aws.StringValue(attr.Value)
		}
	}

	return []interface{}{m}
}

func flattenTargetGroupStickiness(attributes []*elbv2.TargetGroupAttribute) ([]interface{}, error) {
	if len(attributes) == 0 {
		return []interface{}{}, nil
	}

	m := make(map[string]interface{})

	for _, attr := range attributes {
		switch aws.StringValue(attr.Key) {
		case "stickiness.enabled":
			enabled, err := strconv.ParseBool(aws.StringValue(attr.Value))
			if err != nil {
				return nil, fmt.Errorf("converting stickiness.enabled to bool: %s", aws.StringValue(attr.Value))
			}
			m["enabled"] = enabled
		case "stickiness.type":
			m["type"] = aws.StringValue(attr.Value)
		case "stickiness.lb_cookie.duration_seconds":
			if sType, ok := m["type"].(string); !ok || sType == "lb_cookie" {
				duration, err := strconv.Atoi(aws.StringValue(attr.Value))
				if err != nil {
					return nil, fmt.Errorf("converting stickiness.lb_cookie.duration_seconds to int: %s", aws.StringValue(attr.Value))
				}
				m["cookie_duration"] = duration
			}
		case "stickiness.app_cookie.cookie_name":
			m["cookie_name"] = aws.StringValue(attr.Value)
		case "stickiness.app_cookie.duration_seconds":
			if sType, ok := m["type"].(string); !ok || sType == "app_cookie" {
				duration, err := strconv.Atoi(aws.StringValue(attr.Value))
				if err != nil {
					return nil, fmt.Errorf("converting stickiness.app_cookie.duration_seconds to int: %s", aws.StringValue(attr.Value))
				}
				m["cookie_duration"] = duration
			}
		}
	}

	return []interface{}{m}, nil
}

func resourceTargetGroupCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	protocol := diff.Get("protocol").(string)

	// Network Load Balancers have many special quirks to them.
	// See http://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_CreateTargetGroup.html
	if healthChecks := diff.Get("health_check").([]interface{}); len(healthChecks) == 1 {
		healthCheck := healthChecks[0].(map[string]interface{})
		protocol := healthCheck["protocol"].(string)

		if protocol == elbv2.ProtocolEnumTcp {
			// Cannot set custom matcher on TCP health checks
			if m := healthCheck["matcher"].(string); m != "" {
				return fmt.Errorf("%s: health_check.matcher is not supported for target_groups with TCP protocol", diff.Id())
			}
			// Cannot set custom path on TCP health checks
			if m := healthCheck["path"].(string); m != "" {
				return fmt.Errorf("%s: health_check.path is not supported for target_groups with TCP protocol", diff.Id())
			}
		}
	}

	if strings.Contains(protocol, elbv2.ProtocolEnumHttp) {
		if healthChecks := diff.Get("health_check").([]interface{}); len(healthChecks) == 1 {
			healthCheck := healthChecks[0].(map[string]interface{})
			// HTTP(S) Target Groups cannot use TCP health checks
			if p := healthCheck["protocol"].(string); strings.ToLower(p) == "tcp" {
				return fmt.Errorf("HTTP Target Groups cannot use TCP health checks")
			}
		}
	}

	if diff.Id() == "" {
		return nil
	}

	return nil
}

func flattenLbTargetGroupHealthCheck(targetGroup *elbv2.TargetGroup) []interface{} {
	if targetGroup == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"enabled":             aws.BoolValue(targetGroup.HealthCheckEnabled),
		"healthy_threshold":   int(aws.Int64Value(targetGroup.HealthyThresholdCount)),
		"interval":            int(aws.Int64Value(targetGroup.HealthCheckIntervalSeconds)),
		"port":                aws.StringValue(targetGroup.HealthCheckPort),
		"protocol":            aws.StringValue(targetGroup.HealthCheckProtocol),
		"timeout":             int(aws.Int64Value(targetGroup.HealthCheckTimeoutSeconds)),
		"unhealthy_threshold": int(aws.Int64Value(targetGroup.UnhealthyThresholdCount)),
	}

	if targetGroup.HealthCheckPath != nil {
		m["path"] = aws.StringValue(targetGroup.HealthCheckPath)
	}
	if targetGroup.Matcher != nil && targetGroup.Matcher.HttpCode != nil {
		m["matcher"] = aws.StringValue(targetGroup.Matcher.HttpCode)
	}
	if targetGroup.Matcher != nil && targetGroup.Matcher.GrpcCode != nil {
		m["matcher"] = aws.StringValue(targetGroup.Matcher.GrpcCode)
	}

	return []interface{}{m}
}
