package rum

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/cloudwatchrum"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceAppMonitor() *schema.Resource {
	return &schema.Resource{
		Create: resourceAppMonitorCreate,
		Read:   resourceAppMonitorRead,
		Update: resourceAppMonitorUpdate,
		Delete: resourceAppMonitorDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"app_monitor_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_cookies": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"enable_xray": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"excluded_pages": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"favorite_pages": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"guest_role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"identity_pool_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"included_pages": {
							Type:     schema.TypeSet,
							Optional: true,
							MaxItems: 50,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"session_sample_rate": {
							Type:         schema.TypeFloat,
							Optional:     true,
							Default:      0.1,
							ValidateFunc: validation.FloatBetween(0, 1),
						},
						"telemetries": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(cloudwatchrum.Telemetry_Values(), false),
							},
						},
					},
				},
			},
			"cw_log_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"cw_log_group": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 253),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAppMonitorCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RUMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &cloudwatchrum.CreateAppMonitorInput{
		Name:         aws.String(name),
		CwLogEnabled: aws.Bool(d.Get("cw_log_enabled").(bool)),
		Domain:       aws.String(d.Get("domain").(string)),
	}

	if v, ok := d.GetOk("app_monitor_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AppMonitorConfiguration = expandAppMonitorConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	_, err := conn.CreateAppMonitor(input)

	if err != nil {
		return fmt.Errorf("error creating CloudWatch RUM App Monitor %s: %w", name, err)
	}

	d.SetId(name)

	return resourceAppMonitorRead(d, meta)
}

func resourceAppMonitorRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RUMConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	appMon, err := FindAppMonitorByName(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Unable to find CloudWatch RUM App Monitor (%s); removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading CloudWatch RUM App Monitor (%s): %w", d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("appmonitor/%s", aws.StringValue(appMon.Name)),
		Service:   "rum",
	}.String()

	d.Set("name", appMon.Name)
	d.Set("arn", arn)
	d.Set("domain", appMon.Domain)

	d.Set("cw_log_enabled", appMon.DataStorage.CwLog.CwLogEnabled)
	d.Set("cw_log_group", appMon.DataStorage.CwLog.CwLogGroup)

	if err := d.Set("app_monitor_configuration", []interface{}{flattenAppMonitorConfiguration(appMon.AppMonitorConfiguration)}); err != nil {
		return fmt.Errorf("error setting app_monitor_configuration for CloudWatch RUM App Monitor (%s): %w", d.Id(), err)
	}

	tags := KeyValueTags(appMon.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceAppMonitorUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RUMConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &cloudwatchrum.UpdateAppMonitorInput{
			Name: aws.String(d.Id()),
		}

		if d.HasChange("cw_log_enabled") {
			input.CwLogEnabled = aws.Bool(d.Get("cw_log_enabled").(bool))
		}

		if d.HasChange("app_monitor_configuration") {
			input.AppMonitorConfiguration = expandAppMonitorConfiguration(d.Get("app_monitor_configuration").([]interface{})[0].(map[string]interface{}))
		}

		if d.HasChange("domain") {
			input.Domain = aws.String(d.Get("domain").(string))
		}

		log.Printf("[DEBUG] cloudwatchrum AppMonitor update config: %s", input.String())
		_, err := conn.UpdateAppMonitor(input)
		if err != nil {
			return fmt.Errorf("error updating CloudWatch RUM App Monitor: %w", err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating CloudWatch RUM App Monitor (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceAppMonitorRead(d, meta)
}

func resourceAppMonitorDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RUMConn

	input := &cloudwatchrum.DeleteAppMonitorInput{
		Name: aws.String(d.Id()),
	}

	if _, err := conn.DeleteAppMonitor(input); err != nil {
		if tfawserr.ErrCodeEquals(err, cloudwatchrum.ErrCodeResourceNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting CloudWatch RUM App Monitor (%s): %w", d.Id(), err)
	}

	return nil
}

func expandAppMonitorConfiguration(tfMap map[string]interface{}) *cloudwatchrum.AppMonitorConfiguration {
	if tfMap == nil {
		return nil
	}

	config := &cloudwatchrum.AppMonitorConfiguration{}

	if v, ok := tfMap["guest_role_arn"].(string); ok && v != "" {
		config.GuestRoleArn = aws.String(v)
	}

	if v, ok := tfMap["identity_pool_id"].(string); ok && v != "" {
		config.IdentityPoolId = aws.String(v)
	}

	if v, ok := tfMap["session_sample_rate"].(float64); ok {
		config.SessionSampleRate = aws.Float64(v)
	}

	if v, ok := tfMap["allow_cookies"].(bool); ok {
		config.AllowCookies = aws.Bool(v)
	}

	if v, ok := tfMap["enable_xray"].(bool); ok {
		config.EnableXRay = aws.Bool(v)
	}

	if v, ok := tfMap["excluded_pages"].(*schema.Set); ok && v.Len() > 0 {
		config.ExcludedPages = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["favorite_pages"].(*schema.Set); ok && v.Len() > 0 {
		config.FavoritePages = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["included_pages"].(*schema.Set); ok && v.Len() > 0 {
		config.IncludedPages = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["telemetries"].(*schema.Set); ok && v.Len() > 0 {
		config.Telemetries = flex.ExpandStringSet(v)
	}

	return config
}

func flattenAppMonitorConfiguration(apiObject *cloudwatchrum.AppMonitorConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.GuestRoleArn; v != nil {
		tfMap["guest_role_arn"] = aws.StringValue(v)
	}

	if v := apiObject.IdentityPoolId; v != nil {
		tfMap["identity_pool_id"] = aws.StringValue(v)
	}

	if v := apiObject.SessionSampleRate; v != nil {
		tfMap["session_sample_rate"] = aws.Float64Value(v)
	}

	if v := apiObject.AllowCookies; v != nil {
		tfMap["allow_cookies"] = aws.BoolValue(v)
	}

	if v := apiObject.EnableXRay; v != nil {
		tfMap["enable_xray"] = aws.BoolValue(v)
	}

	if v := apiObject.Telemetries; v != nil {
		tfMap["telemetries"] = flex.FlattenStringSet(v)
	}

	if v := apiObject.IncludedPages; v != nil {
		tfMap["included_pages"] = flex.FlattenStringSet(v)
	}

	if v := apiObject.FavoritePages; v != nil {
		tfMap["favorite_pages"] = flex.FlattenStringSet(v)
	}

	if v := apiObject.ExcludedPages; v != nil {
		tfMap["excluded_pages"] = flex.FlattenStringSet(v)
	}

	return tfMap
}
