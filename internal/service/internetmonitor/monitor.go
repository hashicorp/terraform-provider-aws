package internetmonitor

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/internetmonitor"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
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

// @SDKResource("aws_internetmonitor_monitor", name="Monitor")
// @Tags(identifierAttribute="arn")
func ResourceMonitor() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMonitorCreate,
		ReadWithoutTimeout:   resourceMonitorRead,
		UpdateWithoutTimeout: resourceMonitorUpdate,
		DeleteWithoutTimeout: resourceMonitorDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
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
									"bucket_name": {
										Type:     schema.TypeString,
										Required: true,
									},
									"bucket_prefix": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"log_delivery_status": {
										Type:         schema.TypeString,
										Optional:     true,
										Default:      internetmonitor.LogDeliveryStatusEnabled,
										ValidateFunc: validation.StringInSlice(internetmonitor.LogDeliveryStatus_Values(), false),
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
			"resources": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: verify.ValidARN,
				},
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  internetmonitor.MonitorConfigStateActive,
				ValidateFunc: validation.StringInSlice([]string{
					internetmonitor.MonitorConfigStateActive,
					internetmonitor.MonitorConfigStateInactive,
				}, false),
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

func resourceMonitorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InternetMonitorConn()

	monitorName := d.Get("monitor_name").(string)
	input := &internetmonitor.CreateMonitorInput{
		ClientToken: aws.String(id.UniqueId()),
		MonitorName: aws.String(monitorName),
		Tags:        GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("max_city_networks_to_monitor"); ok {
		input.MaxCityNetworksToMonitor = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("traffic_percentage_to_monitor"); ok {
		input.TrafficPercentageToMonitor = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("internet_measurements_log_delivery"); ok {
		input.InternetMeasurementsLogDelivery = expandInternetMeasurementsLogDelivery(v.([]interface{}))
	}

	if v, ok := d.GetOk("resources"); ok && v.(*schema.Set).Len() > 0 {
		input.Resources = flex.ExpandStringSet(v.(*schema.Set))
	}

	log.Printf("[DEBUG] Creating Internet Monitor Monitor: %s", input)
	_, err := conn.CreateMonitorWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Internet Monitor Monitor (%s): %s", monitorName, err)
	}

	d.SetId(monitorName)

	if err := waitMonitor(ctx, conn, monitorName, internetmonitor.MonitorConfigStateActive); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Internet Monitor Monitor (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("status"); ok && v.(string) != internetmonitor.MonitorConfigStateActive {
		input := &internetmonitor.UpdateMonitorInput{
			ClientToken: aws.String(id.UniqueId()),
			MonitorName: aws.String(d.Id()),
			Status:      aws.String(v.(string)),
		}

		_, err := conn.UpdateMonitorWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Internet Monitor Monitor (%s) to inactive at creation: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMonitorRead(ctx, d, meta)...)
}

func resourceMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InternetMonitorConn()

	monitor, err := FindMonitor(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Internet Monitor Monitor (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Internet Monitor Monitor (%s): %s", d.Id(), err)
	}

	err = d.Set("internet_measurements_log_delivery", flattenInternetMeasurementsLogDelivery(monitor.InternetMeasurementsLogDelivery))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting internet_measurements_log_delivery: %s", err)
	}

	d.Set("arn", monitor.MonitorArn)
	d.Set("monitor_name", monitor.MonitorName)
	d.Set("max_city_networks_to_monitor", monitor.MaxCityNetworksToMonitor)
	d.Set("traffic_percentage_to_monitor", monitor.TrafficPercentageToMonitor)
	d.Set("status", monitor.Status)
	d.Set("resources", flex.FlattenStringSet(monitor.Resources))

	SetTagsOut(ctx, monitor.Tags)

	return diags
}

func resourceMonitorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InternetMonitorConn()

	if d.HasChangesExcept("tags", "tags_all") {
		input := &internetmonitor.UpdateMonitorInput{
			ClientToken: aws.String(id.UniqueId()),
			MonitorName: aws.String(d.Id()),
		}

		if d.HasChange("max_city_networks_to_monitor") {
			input.MaxCityNetworksToMonitor = aws.Int64(int64(d.Get("max_city_networks_to_monitor").(int)))
		}

		if d.HasChange("traffic_percentage_to_monitor") {
			input.TrafficPercentageToMonitor = aws.Int64(int64(d.Get("traffic_percentage_to_monitor").(int)))
		}

		if d.HasChange("status") {
			input.Status = aws.String(d.Get("status").(string))
		}

		if d.HasChange("internet_measurements_log_delivery") {
			input.InternetMeasurementsLogDelivery = expandInternetMeasurementsLogDelivery(d.Get("internet_measurements_log_delivery").([]interface{}))
		}

		if d.HasChange("resources") {
			o, n := d.GetChange("resources")
			os, ns := o.(*schema.Set), n.(*schema.Set)
			remove := flex.ExpandStringValueSet(os.Difference(ns))
			add := flex.ExpandStringValueSet(ns.Difference(os))

			if len(add) > 0 {
				input.ResourcesToAdd = aws.StringSlice(add)
			}

			if len(remove) > 0 {
				input.ResourcesToRemove = aws.StringSlice(remove)
			}
		}

		log.Printf("[DEBUG] Updating Internet Monitor Monitor: %s", input)
		_, err := conn.UpdateMonitorWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Internet Monitor Monitor (%s): %s", d.Id(), err)
		}

		if err := waitMonitor(ctx, conn, d.Id(), d.Get("status").(string)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Internet Monitor Monitor (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMonitorRead(ctx, d, meta)...)
}

func resourceMonitorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).InternetMonitorConn()

	input := &internetmonitor.UpdateMonitorInput{
		ClientToken: aws.String(id.UniqueId()),
		MonitorName: aws.String(d.Id()),
		Status:      aws.String(internetmonitor.MonitorConfigStateInactive),
	}

	_, err := conn.UpdateMonitorWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Internet Monitor Monitor (%s) to inactive before deletion: %s", d.Id(), err)
	}

	if err := waitMonitor(ctx, conn, d.Id(), internetmonitor.MonitorConfigStateInactive); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Internet Monitor Monitor (%s) to be inactive before deletion: %s", d.Id(), err)
	}

	log.Printf("[DEBUG] Deleting Internet Monitor Monitor: %s", d.Id())
	_, err = conn.DeleteMonitorWithContext(ctx, &internetmonitor.DeleteMonitorInput{
		MonitorName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, internetmonitor.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Internet Monitor Monitor (%s): %s", d.Id(), err)
	}

	return diags
}

func expandInternetMeasurementsLogDelivery(vInternetMeasurementsLogDelivery []interface{}) *internetmonitor.InternetMeasurementsLogDelivery {
	if len(vInternetMeasurementsLogDelivery) == 0 || vInternetMeasurementsLogDelivery[0] == nil {
		return nil
	}
	mInternetMeasurementsLogDelivery := vInternetMeasurementsLogDelivery[0].(map[string]interface{})

	logDelivery := &internetmonitor.InternetMeasurementsLogDelivery{}

	if v, ok := mInternetMeasurementsLogDelivery["s3_config"].([]interface{}); ok {
		logDelivery.S3Config = expandS3Config(v)
	}

	return logDelivery
}

func expandS3Config(vS3Config []interface{}) *internetmonitor.S3Config {
	if len(vS3Config) == 0 || vS3Config[0] == nil {
		return nil
	}
	mS3Config := vS3Config[0].(map[string]interface{})

	s3Config := &internetmonitor.S3Config{}

	if v, ok := mS3Config["bucket_name"].(string); ok && v != "" {
		s3Config.BucketName = aws.String(v)
	}

	if v, ok := mS3Config["bucket_prefix"].(string); ok && v != "" {
		s3Config.BucketPrefix = aws.String(v)
	}

	if v, ok := mS3Config["log_delivery_status"].(string); ok && v != "" {
		s3Config.LogDeliveryStatus = aws.String(v)
	}

	return s3Config
}

func flattenInternetMeasurementsLogDelivery(internetMeasurementsLogDelivery *internetmonitor.InternetMeasurementsLogDelivery) []interface{} {
	if internetMeasurementsLogDelivery == nil {
		return []interface{}{}
	}

	mInternetMeasurementsLogDelivery := map[string]interface{}{
		"s3_config": flattenS3Config(internetMeasurementsLogDelivery.S3Config),
	}

	return []interface{}{mInternetMeasurementsLogDelivery}
}

func flattenS3Config(s3Config *internetmonitor.S3Config) []interface{} {
	if s3Config == nil {
		return []interface{}{}
	}

	mS3Config := map[string]interface{}{
		"bucket_name": aws.StringValue(s3Config.BucketName),
	}

	if s3Config.BucketPrefix != nil {
		mS3Config["bucket_prefix"] = aws.StringValue(s3Config.BucketPrefix)
	}

	if s3Config.LogDeliveryStatus != nil {
		mS3Config["log_delivery_status"] = aws.StringValue(s3Config.LogDeliveryStatus)
	}

	return []interface{}{mS3Config}
}
