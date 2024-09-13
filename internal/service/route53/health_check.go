// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_route53_health_check", name="Health Check")
// @Tags(identifierAttribute="id", resourceType="healthcheck")
func resourceHealthCheck() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceHealthCheckCreate,
		ReadWithoutTimeout:   resourceHealthCheckRead,
		UpdateWithoutTimeout: resourceHealthCheckUpdate,
		DeleteWithoutTimeout: resourceHealthCheckDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.CloudWatchRegion](),
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
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.InsufficientDataHealthStatus](),
			},
			"invert_healthcheck": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			names.AttrIPAddress: {
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
			names.AttrPort: {
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
					Type:             schema.TypeString,
					ValidateDiagFunc: enum.Validate[awstypes.HealthCheckRegion](),
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
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.HealthCheckType](),
				StateFunc: func(val interface{}) string {
					return strings.ToUpper(val.(string))
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHealthCheckCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	healthCheckType := awstypes.HealthCheckType(d.Get(names.AttrType).(string))
	healthCheckConfig := &awstypes.HealthCheckConfig{
		Type: healthCheckType,
	}

	if v, ok := d.GetOk("disabled"); ok {
		healthCheckConfig.Disabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("enable_sni"); ok {
		healthCheckConfig.EnableSNI = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("failure_threshold"); ok {
		healthCheckConfig.FailureThreshold = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("fqdn"); ok {
		healthCheckConfig.FullyQualifiedDomainName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("invert_healthcheck"); ok {
		healthCheckConfig.Inverted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrIPAddress); ok {
		healthCheckConfig.IPAddress = aws.String(v.(string))
	}
	if v, ok := d.GetOk(names.AttrPort); ok {
		healthCheckConfig.Port = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("request_interval"); ok {
		healthCheckConfig.RequestInterval = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("resource_path"); ok {
		healthCheckConfig.ResourcePath = aws.String(v.(string))
	}

	if v, ok := d.GetOk("search_string"); ok {
		healthCheckConfig.SearchString = aws.String(v.(string))
	}

	switch healthCheckType {
	case awstypes.HealthCheckTypeCalculated:
		if v, ok := d.GetOk("child_health_threshold"); ok {
			healthCheckConfig.HealthThreshold = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("child_healthchecks"); ok {
			healthCheckConfig.ChildHealthChecks = flex.ExpandStringValueSet(v.(*schema.Set))
		}
	case awstypes.HealthCheckTypeCloudwatchMetric:
		alarmIdentifier := &awstypes.AlarmIdentifier{}

		if v, ok := d.GetOk("cloudwatch_alarm_name"); ok {
			alarmIdentifier.Name = aws.String(v.(string))
		}

		if v, ok := d.GetOk("cloudwatch_alarm_region"); ok {
			alarmIdentifier.Region = awstypes.CloudWatchRegion(v.(string))
		}

		healthCheckConfig.AlarmIdentifier = alarmIdentifier

		if v, ok := d.GetOk("insufficient_data_health_status"); ok {
			healthCheckConfig.InsufficientDataHealthStatus = awstypes.InsufficientDataHealthStatus(v.(string))
		}
	case awstypes.HealthCheckTypeRecoveryControl:
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
		healthCheckConfig.Regions = flex.ExpandStringyValueSet[awstypes.HealthCheckRegion](v.(*schema.Set))
	}

	callerRef := id.UniqueId()
	if v, ok := d.GetOk("reference_name"); ok {
		callerRef = fmt.Sprintf("%s-%s", v.(string), callerRef)
	}

	input := &route53.CreateHealthCheckInput{
		CallerReference:   aws.String(callerRef),
		HealthCheckConfig: healthCheckConfig,
	}

	output, err := conn.CreateHealthCheck(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Route53 Health Check: %s", err)
	}

	d.SetId(aws.ToString(output.HealthCheck.Id))

	if err := createTags(ctx, conn, d.Id(), string(awstypes.TagResourceTypeHealthcheck), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting Route53 Health Check (%s) tags: %s", d.Id(), err)
	}

	return append(diags, resourceHealthCheckRead(ctx, d, meta)...)
}

func resourceHealthCheckRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	output, err := findHealthCheckByID(ctx, conn, d.Id())

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
		Resource:  "healthcheck/" + d.Id(),
	}.String()
	d.Set(names.AttrARN, arn)
	healthCheckConfig := output.HealthCheckConfig
	d.Set("child_health_threshold", healthCheckConfig.HealthThreshold)
	d.Set("child_healthchecks", healthCheckConfig.ChildHealthChecks)
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
	d.Set(names.AttrIPAddress, healthCheckConfig.IPAddress)
	d.Set("measure_latency", healthCheckConfig.MeasureLatency)
	d.Set(names.AttrPort, healthCheckConfig.Port)
	d.Set("regions", healthCheckConfig.Regions)
	d.Set("request_interval", healthCheckConfig.RequestInterval)
	d.Set("resource_path", healthCheckConfig.ResourcePath)
	d.Set("routing_control_arn", healthCheckConfig.RoutingControlArn)
	d.Set("search_string", healthCheckConfig.SearchString)
	d.Set(names.AttrType, healthCheckConfig.Type)

	return diags
}

func resourceHealthCheckUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &route53.UpdateHealthCheckInput{
			HealthCheckId: aws.String(d.Id()),
		}

		if d.HasChange("child_health_threshold") {
			input.HealthThreshold = aws.Int32(int32(d.Get("child_health_threshold").(int)))
		}

		if d.HasChange("child_healthchecks") {
			input.ChildHealthChecks = flex.ExpandStringValueSet(d.Get("child_healthchecks").(*schema.Set))
		}

		if d.HasChanges("cloudwatch_alarm_name", "cloudwatch_alarm_region") {
			alarmIdentifier := &awstypes.AlarmIdentifier{
				Name:   aws.String(d.Get("cloudwatch_alarm_name").(string)),
				Region: awstypes.CloudWatchRegion(d.Get("cloudwatch_alarm_region").(string)),
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
			input.FailureThreshold = aws.Int32(int32(d.Get("failure_threshold").(int)))
		}

		if d.HasChange("fqdn") {
			input.FullyQualifiedDomainName = aws.String(d.Get("fqdn").(string))
		}

		if d.HasChange("insufficient_data_health_status") {
			input.InsufficientDataHealthStatus = awstypes.InsufficientDataHealthStatus(d.Get("insufficient_data_health_status").(string))
		}

		if d.HasChange("invert_healthcheck") {
			input.Inverted = aws.Bool(d.Get("invert_healthcheck").(bool))
		}

		if d.HasChange(names.AttrIPAddress) {
			input.IPAddress = aws.String(d.Get(names.AttrIPAddress).(string))
		}

		if d.HasChange(names.AttrPort) {
			input.Port = aws.Int32(int32(d.Get(names.AttrPort).(int)))
		}

		if d.HasChange("regions") {
			input.Regions = flex.ExpandStringyValueSet[awstypes.HealthCheckRegion](d.Get("regions").(*schema.Set))
		}

		if d.HasChange("resource_path") {
			input.ResourcePath = aws.String(d.Get("resource_path").(string))
		}

		if d.HasChange("search_string") {
			input.SearchString = aws.String(d.Get("search_string").(string))
		}

		_, err := conn.UpdateHealthCheck(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Route53 Health Check (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceHealthCheckRead(ctx, d, meta)...)
}

func resourceHealthCheckDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).Route53Client(ctx)

	log.Printf("[DEBUG] Deleting Route53 Health Check: %s", d.Id())
	_, err := conn.DeleteHealthCheck(ctx, &route53.DeleteHealthCheckInput{
		HealthCheckId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NoSuchHealthCheck](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Route53 Health Check (%s): %s", d.Id(), err)
	}

	return diags
}

func findHealthCheckByID(ctx context.Context, conn *route53.Client, id string) (*awstypes.HealthCheck, error) {
	input := &route53.GetHealthCheckInput{
		HealthCheckId: aws.String(id),
	}

	output, err := conn.GetHealthCheck(ctx, input)

	if errs.IsA[*awstypes.NoSuchHealthCheck](err) {
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
