// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_health_check", name="Health Check")
// @Tags(identifierAttribute="id", resourceType="healthcheck")
func ResourceHealthCheck() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHealthCheckCreate,
		ReadWithoutTimeout:   resourceHealthCheckRead,
		UpdateWithoutTimeout: resourceHealthCheckUpdate,
		DeleteWithoutTimeout: resourceHealthCheckDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"child_health_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntAtMost(256),
			},
			"child_healthchecks": {
				Type:     schema.TypeSet,
				MaxItems: 256,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringLenBetween(0, 64),
				},
				Optional: true,
			},
			"cloudwatch_alarm_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cloudwatch_alarm_region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"enable_sni": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"failure_threshold": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntBetween(1, 10),
			},
			"fqdn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"insufficient_data_health_status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(route53.InsufficientDataHealthStatus_Values(), true),
			},
			"invert_healthcheck": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"ip_address": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.IsIPAddress,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return net.ParseIP(old).Equal(net.ParseIP(new))
				},
			},
			"measure_latency": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"port": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IsPortNumber,
			},
			"reference_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// The max length of the reference name is 64 characters for the API.
				// Terraform appends a 37-character unique ID to the provided
				// reference_name. This limits the length of the resource argument to 27.
				//
				// Example generated suffix: -terraform-20190122200019880700000001
				ValidateFunc: validation.StringLenBetween(0, (64 - id.UniqueIDSuffixLength - 11)),
			},
			"regions": {
				Type:     schema.TypeSet,
				MinItems: 3,
				MaxItems: 64,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(route53.HealthCheckRegion_Values(), false),
				},
				Optional: true,
			},
			"request_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.IntInSlice([]int{10, 30}),
			},
			"resource_path": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			"routing_control_arn": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"search_string": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 255),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
				ValidateFunc: validation.StringInSlice(route53.HealthCheckType_Values(), true),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHealthCheckCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	healthCheckType := d.Get("type").(string)
	healthCheckConfig := &route53.HealthCheckConfig{
		Type: aws.String(healthCheckType),
	}

	if v, ok := d.GetOk("disabled"); ok {
		healthCheckConfig.Disabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_sni"); ok {
		healthCheckConfig.EnableSNI = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("failure_threshold"); ok {
		healthCheckConfig.FailureThreshold = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("fqdn"); ok {
		healthCheckConfig.FullyQualifiedDomainName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invert_healthcheck"); ok {
		healthCheckConfig.Inverted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("ip_address"); ok {
		healthCheckConfig.IPAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk("port"); ok {
		healthCheckConfig.Port = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("request_interval"); ok {
		healthCheckConfig.RequestInterval = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("resource_path"); ok {
		healthCheckConfig.ResourcePath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("search_string"); ok {
		healthCheckConfig.SearchString = aws.String(v.(string))
	}

	switch healthCheckType {
	case route53.HealthCheckTypeCalculated:
		if v, ok := d.GetOk("child_health_threshold"); ok {
			healthCheckConfig.HealthThreshold = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("child_healthchecks"); ok {
			healthCheckConfig.ChildHealthChecks = flex.ExpandStringSet(v.(*schema.Set))
		}
	case route53.HealthCheckTypeCloudwatchMetric:
		alarmIdentifier := &route53.AlarmIdentifier{}

		if v, ok := d.GetOk("cloudwatch_alarm_name"); ok {
			alarmIdentifier.Name = aws.String(v.(string))
		}

		if v, ok := d.GetOk("cloudwatch_alarm_region"); ok {
			alarmIdentifier.Region = aws.String(v.(string))
		}

		healthCheckConfig.AlarmIdentifier = alarmIdentifier

		if v, ok := d.GetOk("insufficient_data_health_status"); ok {
			healthCheckConfig.InsufficientDataHealthStatus = aws.String(v.(string))
		}
	case route53.HealthCheckTypeRecoveryControl:
		if v, ok := d.GetOk("routing_control_arn"); ok {
			healthCheckConfig.RoutingControlArn = aws.String(v.(string))
		}
		fallthrough
	default:
		if v, ok := d.GetOk("measure_latency"); ok {
			healthCheckConfig.MeasureLatency = aws.Bool(v.(bool))
		}
	}

	if v, ok := d.GetOk("regions"); ok && v.(*schema.Set).Len() > 0 {
		healthCheckConfig.Regions = flex.ExpandStringSet(v.(*schema.Set))
	}

	callerRef := id.UniqueId()
	if v, ok := d.GetOk("reference_name"); ok {
		callerRef = fmt.Sprintf("%s-%s", v.(string), callerRef)
	}

	input := &route53.CreateHealthCheckInput{
		CallerReference:   aws.String(callerRef),
		HealthCheckConfig: healthCheckConfig,
	}

	output, err := conn.CreateHealthCheckWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Health Check: %s", err)
	}

	d.SetId(aws.StringValue(output.HealthCheck.Id))

	if err := createTags(ctx, conn, d.Id(), route53.TagResourceTypeHealthcheck, getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Health Check (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceHealthCheckRead(ctx, d, meta)...)
}

func resourceHealthCheckRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	output, err := FindHealthCheckByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Health Check (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Route53 Health Check (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "route53",
		Resource:  fmt.Sprintf("healthcheck/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	healthCheckConfig := output.HealthCheckConfig
	d.Set("child_health_threshold", healthCheckConfig.HealthThreshold)
	d.Set("child_healthchecks", aws.StringValueSlice(healthCheckConfig.ChildHealthChecks))
	if alarmIdentifier := healthCheckConfig.AlarmIdentifier; alarmIdentifier != nil {
		d.Set("cloudwatch_alarm_name", alarmIdentifier.Name)
		d.Set("cloudwatch_alarm_region", alarmIdentifier.Region)
	}
	d.Set("disabled", healthCheckConfig.Disabled)
	d.Set("enable_sni", healthCheckConfig.EnableSNI)
	d.Set("failure_threshold", healthCheckConfig.FailureThreshold)
	d.Set("fqdn", healthCheckConfig.FullyQualifiedDomainName)
	d.Set("insufficient_data_health_status", healthCheckConfig.InsufficientDataHealthStatus)
	d.Set("invert_healthcheck", healthCheckConfig.Inverted)
	d.Set("ip_address", healthCheckConfig.IPAddress)
	d.Set("measure_latency", healthCheckConfig.MeasureLatency)
	d.Set("port", healthCheckConfig.Port)
	d.Set("regions", aws.StringValueSlice(healthCheckConfig.Regions))
	d.Set("request_interval", healthCheckConfig.RequestInterval)
	d.Set("resource_path", healthCheckConfig.ResourcePath)
	d.Set("routing_control_arn", healthCheckConfig.RoutingControlArn)
	d.Set("search_string", healthCheckConfig.SearchString)
	d.Set("type", healthCheckConfig.Type)

	return diags
}

func resourceHealthCheckUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &route53.UpdateHealthCheckInput{
			HealthCheckId: aws.String(d.Id()),
		}

		if d.HasChange("child_health_threshold") {
			input.HealthThreshold = aws.Int64(int64(d.Get("child_health_threshold").(int)))
		}

		if d.HasChange("child_healthchecks") {
			input.ChildHealthChecks = flex.ExpandStringSet(d.Get("child_healthchecks").(*schema.Set))
		}

		if d.HasChanges("cloudwatch_alarm_name", "cloudwatch_alarm_region") {
			alarmIdentifier := &route53.AlarmIdentifier{
				Name:   aws.String(d.Get("cloudwatch_alarm_name").(string)),
				Region: aws.String(d.Get("cloudwatch_alarm_region").(string)),
			}

			input.AlarmIdentifier = alarmIdentifier
		}

		if d.HasChange("disabled") {
			input.Disabled = aws.Bool(d.Get("disabled").(bool))
		}

		if d.HasChange("enable_sni") {
			input.EnableSNI = aws.Bool(d.Get("enable_sni").(bool))
		}

		if d.HasChange("failure_threshold") {
			input.FailureThreshold = aws.Int64(int64(d.Get("failure_threshold").(int)))
		}

		if d.HasChange("fqdn") {
			input.FullyQualifiedDomainName = aws.String(d.Get("fqdn").(string))
		}

		if d.HasChange("insufficient_data_health_status") {
			input.InsufficientDataHealthStatus = aws.String(d.Get("insufficient_data_health_status").(string))
		}

		if d.HasChange("invert_healthcheck") {
			input.Inverted = aws.Bool(d.Get("invert_healthcheck").(bool))
		}

		if d.HasChange("ip_address") {
			input.IPAddress = aws.String(d.Get("ip_address").(string))
		}

		if d.HasChange("port") {
			input.Port = aws.Int64(int64(d.Get("port").(int)))
		}

		if d.HasChange("regions") {
			input.Regions = flex.ExpandStringSet(d.Get("regions").(*schema.Set))
		}

		if d.HasChange("resource_path") {
			input.ResourcePath = aws.String(d.Get("resource_path").(string))
		}

		if d.HasChange("search_string") {
			input.SearchString = aws.String(d.Get("search_string").(string))
		}

		_, err := conn.UpdateHealthCheckWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Health Check (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceHealthCheckRead(ctx, d, meta)...)
}

func resourceHealthCheckDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Conn(ctx)

	log.Printf("[DEBUG] Deleting Route53 Health Check: %s", d.Id())
	_, err := conn.DeleteHealthCheckWithContext(ctx, &route53.DeleteHealthCheckInput{
		HealthCheckId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHealthCheck) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Health Check (%s): %s", d.Id(), err)
	}

	return diags
}

func FindHealthCheckByID(ctx context.Context, conn *route53.Route53, id string) (*route53.HealthCheck, error) {
	input := &route53.GetHealthCheckInput{
		HealthCheckId: aws.String(id),
	}

	output, err := conn.GetHealthCheckWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53.ErrCodeNoSuchHealthCheck) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.HealthCheck == nil || output.HealthCheck.HealthCheckConfig == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.HealthCheck, nil
}
