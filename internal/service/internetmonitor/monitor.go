// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package internetmonitor

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/internetmonitor"
	"github.com/aws/aws-sdk-go-v2/service/internetmonitor/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_internetmonitor_monitor", name="Monitor")
// @Tags(identifierAttribute="arn")
func resourceMonitor() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMonitorCreate,
		ReadWithoutTimeout:   resourceMonitorRead,
		UpdateWithoutTimeout: resourceMonitorUpdate,
		DeleteWithoutTimeout: resourceMonitorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"health_events_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"availability_score_threshold": {
							Type:     schema.TypeFloat,
							Optional: true,
							Default:  95.0,
						},
						"performance_score_threshold": {
							Type:     schema.TypeFloat,
							Optional: true,
							Default:  95.0,
						},
					},
				},
			},
			"internet_measurements_log_delivery": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrBucketName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrBucketPrefix: {
										Type:     schema.TypeString,
										Optional: true,
									},
									"log_delivery_status": {
										Type:             schema.TypeString,
										Optional:         true,
										Default:          types.LogDeliveryStatusEnabled,
										ValidateDiagFunc: enum.Validate[types.LogDeliveryStatus](),
									},
								},
							},
						},
					},
				},
			},
			"max_city_networks_to_monitor": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 500000),
				AtLeastOneOf: []string{"traffic_percentage_to_monitor", "max_city_networks_to_monitor"},
			},
			"monitor_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			names.AttrResources: {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Optional: true,
				Default:  types.MonitorConfigStateActive,
				ValidateFunc: validation.StringInSlice(enum.Slice(
					types.MonitorConfigStateActive,
					types.MonitorConfigStateInactive,
				), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"traffic_percentage_to_monitor": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(1, 100),
				AtLeastOneOf: []string{"traffic_percentage_to_monitor", "max_city_networks_to_monitor"},
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

const (
	errCodeResourceNotFoundException = "ResourceNotFoundException"
)

func resourceMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InternetMonitorClient(ctx)

	name := d.Get("monitor_name").(string)
	input := &internetmonitor.CreateMonitorInput{
		ClientToken: aws.String(id.UniqueId()),
		MonitorName: aws.String(name),
		Tags:        getTagsIn(ctx),
	}

	if v, ok := d.GetOk("health_events_config"); ok {
		input.HealthEventsConfig = expandHealthEventsConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk("internet_measurements_log_delivery"); ok {
		input.InternetMeasurementsLogDelivery = expandInternetMeasurementsLogDelivery(v.([]interface{}))
	}

	if v, ok := d.GetOk("max_city_networks_to_monitor"); ok {
		input.MaxCityNetworksToMonitor = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk(names.AttrResources); ok && v.(*schema.Set).Len() > 0 {
		input.Resources = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("traffic_percentage_to_monitor"); ok {
		input.TrafficPercentageToMonitor = aws.Int32(int32(v.(int)))
	}

	_, err := conn.CreateMonitor(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Internet Monitor Monitor (%s): %s", name, err)
	}

	d.SetId(name)

	if err := waitMonitor(ctx, conn, d.Id(), types.MonitorConfigStateActive); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Internet Monitor Monitor (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk(names.AttrStatus); ok {
		if v := types.MonitorConfigState(v.(string)); v != types.MonitorConfigStateActive {
			input := &internetmonitor.UpdateMonitorInput{
				ClientToken: aws.String(id.UniqueId()),
				MonitorName: aws.String(d.Id()),
				Status:      v,
			}

			_, err := conn.UpdateMonitor(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Internet Monitor Monitor (%s): %s", d.Id(), err)
			}

			if err := waitMonitor(ctx, conn, d.Id(), v); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Internet Monitor Monitor (%s) INACTIVE: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceMonitorRead(ctx, d, meta)...)
}

func resourceMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InternetMonitorClient(ctx)

	monitor, err := findMonitorByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Internet Monitor Monitor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Internet Monitor Monitor (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, monitor.MonitorArn)
	if err := d.Set("health_events_config", flattenHealthEventsConfig(monitor.HealthEventsConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting health_events_config: %s", err)
	}
	if err := d.Set("internet_measurements_log_delivery", flattenInternetMeasurementsLogDelivery(monitor.InternetMeasurementsLogDelivery)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting internet_measurements_log_delivery: %s", err)
	}
	d.Set("monitor_name", monitor.MonitorName)
	d.Set("max_city_networks_to_monitor", monitor.MaxCityNetworksToMonitor)
	d.Set(names.AttrResources, flex.FlattenStringValueSet(monitor.Resources))
	d.Set(names.AttrStatus, monitor.Status)
	d.Set("traffic_percentage_to_monitor", monitor.TrafficPercentageToMonitor)

	setTagsOut(ctx, monitor.Tags)

	return diags
}

func resourceMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InternetMonitorClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &internetmonitor.UpdateMonitorInput{
			ClientToken: aws.String(id.UniqueId()),
			MonitorName: aws.String(d.Id()),
		}

		if d.HasChange("health_events_config") {
			input.HealthEventsConfig = expandHealthEventsConfig(d.Get("health_events_config").([]interface{}))
		}

		if d.HasChange("internet_measurements_log_delivery") {
			input.InternetMeasurementsLogDelivery = expandInternetMeasurementsLogDelivery(d.Get("internet_measurements_log_delivery").([]interface{}))
		}

		if d.HasChange("max_city_networks_to_monitor") {
			input.MaxCityNetworksToMonitor = aws.Int32(int32(d.Get("max_city_networks_to_monitor").(int)))
		}

		if d.HasChange(names.AttrResources) {
			o, n := d.GetChange(names.AttrResources)
			os, ns := o.(*schema.Set), n.(*schema.Set)
			if add := flex.ExpandStringValueSet(ns.Difference(os)); len(add) > 0 {
				input.ResourcesToAdd = add
			}
			if remove := flex.ExpandStringValueSet(os.Difference(ns)); len(remove) > 0 {
				input.ResourcesToRemove = remove
			}
		}

		status := types.MonitorConfigState(d.Get(names.AttrStatus).(string))
		if d.HasChange(names.AttrStatus) {
			input.Status = status
		}

		if d.HasChange("traffic_percentage_to_monitor") {
			input.TrafficPercentageToMonitor = aws.Int32(int32(d.Get("traffic_percentage_to_monitor").(int)))
		}

		_, err := conn.UpdateMonitor(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Internet Monitor Monitor (%s): %s", d.Id(), err)
		}

		if err := waitMonitor(ctx, conn, d.Id(), status); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Internet Monitor Monitor (%s) %s: %s", d.Id(), status, err)
		}
	}

	return append(diags, resourceMonitorRead(ctx, d, meta)...)
}

func resourceMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InternetMonitorClient(ctx)

	input := &internetmonitor.UpdateMonitorInput{
		ClientToken: aws.String(id.UniqueId()),
		MonitorName: aws.String(d.Id()),
		Status:      types.MonitorConfigStateInactive,
	}

	_, err := conn.UpdateMonitor(ctx, input)

	// if errs.IsA[*types.ResourceNotFoundException](err) {
	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Internet Monitor Monitor (%s): %s", d.Id(), err)
	}

	if err := waitMonitor(ctx, conn, d.Id(), types.MonitorConfigStateInactive); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Internet Monitor Monitor (%s) INACTIVE: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Internet Monitor Monitor: %s", d.Id())
	_, err = conn.DeleteMonitor(ctx, &internetmonitor.DeleteMonitorInput{
		MonitorName: aws.String(d.Id()),
	})

	// if errs.IsA[*types.ResourceNotFoundException](err) {
	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Internet Monitor Monitor (%s): %s", d.Id(), err)
	}

	return diags
}

func findMonitorByName(ctx context.Context, conn *internetmonitor.Client, name string) (*internetmonitor.GetMonitorOutput, error) {
	input := &internetmonitor.GetMonitorInput{
		MonitorName: aws.String(name),
	}

	output, err := conn.GetMonitor(ctx, input)

	// if errs.IsA[*types.ResourceNotFoundException](err) {
	if tfawserr.ErrCodeEquals(err, errCodeResourceNotFoundException) {
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

	return output, nil
}

func statusMonitor(ctx context.Context, conn *internetmonitor.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		monitor, err := findMonitorByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return monitor, string(monitor.Status), nil
	}
}

func waitMonitor(ctx context.Context, conn *internetmonitor.Client, name string, targetState types.MonitorConfigState) error {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.MonitorConfigStatePending),
		Target:  enum.Slice(targetState),
		Refresh: statusMonitor(ctx, conn, name),
		Timeout: timeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*internetmonitor.GetMonitorOutput); ok {
		if status := output.Status; status == types.MonitorConfigStateError {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ProcessingStatusInfo)))
		}

		return err
	}

	return err
}

func expandHealthEventsConfig(tfList []interface{}) *types.HealthEventsConfig {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &types.HealthEventsConfig{}

	if v, ok := tfMap["availability_score_threshold"].(float64); ok && v != 0.0 {
		apiObject.AvailabilityScoreThreshold = v
	}

	if v, ok := tfMap["performance_score_threshold"].(float64); ok && v != 0.0 {
		apiObject.PerformanceScoreThreshold = v
	}

	return apiObject
}

func expandInternetMeasurementsLogDelivery(tfList []interface{}) *types.InternetMeasurementsLogDelivery {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &types.InternetMeasurementsLogDelivery{}

	if v, ok := tfMap["s3_config"].([]interface{}); ok {
		apiObject.S3Config = expandS3Config(v)
	}

	return apiObject
}

func expandS3Config(tfList []interface{}) *types.S3Config {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &types.S3Config{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		apiObject.BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrBucketPrefix].(string); ok && v != "" {
		apiObject.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["log_delivery_status"].(string); ok && v != "" {
		apiObject.LogDeliveryStatus = types.LogDeliveryStatus(v)
	}

	return apiObject
}

func flattenHealthEventsConfig(apiObject *types.HealthEventsConfig) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"availability_score_threshold": apiObject.AvailabilityScoreThreshold,
		"performance_score_threshold":  apiObject.PerformanceScoreThreshold,
	}

	return []interface{}{tfMap}
}

func flattenInternetMeasurementsLogDelivery(apiObject *types.InternetMeasurementsLogDelivery) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"s3_config": flattenS3Config(apiObject.S3Config),
	}

	return []interface{}{tfMap}
}

func flattenS3Config(apiObject *types.S3Config) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		names.AttrBucketName:  aws.ToString(apiObject.BucketName),
		"log_delivery_status": string(apiObject.LogDeliveryStatus),
	}

	if apiObject.BucketPrefix != nil {
		tfMap[names.AttrBucketPrefix] = aws.ToString(apiObject.BucketPrefix)
	}

	return []interface{}{tfMap}
}
