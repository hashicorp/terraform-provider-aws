package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsPinpointApp() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsPinpointAppCreate,
		Read:   resourceAwsPinpointAppRead,
		Update: resourceAwsPinpointAppUpdate,
		Delete: resourceAwsPinpointAppDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"application_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			//"cloudwatch_metrics_enabled": {
			//	Type:     schema.TypeBool,
			//	Optional: true,
			//	Default:  false,
			//},
			"campaign_hook": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      campaignHookHash,
				MaxItems: 1,
				MinItems: 0,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lambda_function_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								pinpoint.ModeDelivery,
								pinpoint.ModeFilter,
							}, false),
						},
						"web_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"limits": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      campaignLimitsHash,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"daily": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"maximum_duration": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"messages_per_second": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"total": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			"quiet_time": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      quietTimeHash,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"end": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"start": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsPinpointAppCreate(d *schema.ResourceData, meta interface{}) error {
	pinpointconn := meta.(*AWSClient).pinpointconn

	var name string
	if v, ok := d.GetOk("name"); ok {
		name = v.(string)
	} else if v, ok := d.GetOk("name_prefix"); ok {
		name = resource.PrefixedUniqueId(v.(string))
	} else {
		name = resource.UniqueId()
	}

	log.Printf("[DEBUG] Pinpoint create app: %s", name)

	req := &pinpoint.CreateAppInput{
		CreateApplicationRequest: &pinpoint.CreateApplicationRequest{
			Name: aws.String(name),
		},
	}

	output, err := pinpointconn.CreateApp(req)
	if err != nil {
		return fmt.Errorf("[ERROR] creating Pinpoint app: %s", err)
	}

	d.SetId(*output.ApplicationResponse.Id)

	return resourceAwsPinpointAppUpdate(d, meta)
}

func resourceAwsPinpointAppUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).pinpointconn

	appSettings := &pinpoint.WriteApplicationSettingsRequest{}

	//if d.HasChange("cloudwatch_metrics_enabled") {
	//	appSettings.CloudWatchMetricsEnabled = aws.Bool(d.Get("cloudwatch_metrics_enabled").(bool));
	//}

	if d.HasChange("campaign_hook") {
		appSettings.CampaignHook = expandCampaignHook(d)
	}

	if d.HasChange("limits") {
		appSettings.Limits = expandCampaignLimits(d.Get("limits").(*schema.Set))
	}

	if d.HasChange("quiet_time") {
		appSettings.QuietTime = expandQuietTime(d.Get("quiet_time").(*schema.Set))
	}

	req := pinpoint.UpdateApplicationSettingsInput{
		ApplicationId:                   aws.String(d.Id()),
		WriteApplicationSettingsRequest: appSettings,
	}

	_, err := conn.UpdateApplicationSettings(&req)
	if err != nil {
		return err
	}

	return resourceAwsPinpointAppRead(d, meta)
}

func resourceAwsPinpointAppRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).pinpointconn

	log.Printf("[INFO] Reading Pinpoint App Attributes for %s", d.Id())

	app, err := conn.GetApp(&pinpoint.GetAppInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, pinpoint.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Pinpoint App (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	settings, err := conn.GetApplicationSettings(&pinpoint.GetApplicationSettingsInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if isAWSErr(err, pinpoint.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Pinpoint App (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("name", app.ApplicationResponse.Name)
	d.Set("application_id", app.ApplicationResponse.Id)

	d.Set("campaign_hook", flattenCampaignHook(settings.ApplicationSettingsResource.CampaignHook))
	d.Set("limits", flattenCampaignLimits(settings.ApplicationSettingsResource.Limits))
	d.Set("quiet_time", flattenQuietTime(settings.ApplicationSettingsResource.QuietTime))

	return nil
}

func resourceAwsPinpointAppDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).pinpointconn

	log.Printf("[DEBUG] Pinpoint Delete App: %s", d.Id())
	_, err := conn.DeleteApp(&pinpoint.DeleteAppInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		return err
	}
	return nil
}

func expandCampaignHook(d *schema.ResourceData) *pinpoint.CampaignHook {
	configs := d.Get("campaign_hook").(*schema.Set).List()
	if configs == nil || len(configs) == 0 {
		return nil
	}

	m := configs[0].(map[string]interface{})

	ch := &pinpoint.CampaignHook{}

	if m["lambda_function_name"] != nil {
		ch.LambdaFunctionName = aws.String(m["lambda_function_name"].(string))
	}

	if m["mode"] != nil {
		ch.Mode = aws.String(m["mode"].(string))
	}

	if m["web_url"] != nil {
		ch.WebUrl = aws.String(m["web_url"].(string))
	}

	return ch
}

func flattenCampaignHook(ch *pinpoint.CampaignHook) []interface{} {
	l := make([]interface{}, 0)

	m := map[string]interface{}{}

	if ch.LambdaFunctionName != nil {
		m["lambda_function_name"] = *ch.LambdaFunctionName
	}

	if ch.Mode != nil && *ch.Mode != "" {
		m["mode"] = *ch.Mode
	}

	if ch.WebUrl != nil {
		m["web_url"] = *ch.WebUrl
	}

	if len(m) <= 0 {
		return nil
	}

	l = append(l, m)

	return l
}

func expandCampaignLimits(s *schema.Set) *pinpoint.CampaignLimits {
	if s == nil || s.Len() == 0 {
		return nil
	}
	m := s.List()[0].(map[string]interface{})

	cl := pinpoint.CampaignLimits{}

	if m["daily"] != nil {
		cl.Daily = aws.Int64(int64(m["daily"].(int)))
	}

	if m["maximum_duration"] != nil {
		cl.MaximumDuration = aws.Int64(int64(m["maximum_duration"].(int)))
	}

	if m["messages_per_second"] != nil {
		cl.MessagesPerSecond = aws.Int64(int64(m["messages_per_second"].(int)))
	}
	if m["total"] != nil {
		cl.Total = aws.Int64(int64(m["total"].(int)))
	}

	return &cl
}

func flattenCampaignLimits(cl *pinpoint.CampaignLimits) []interface{} {
	l := make([]interface{}, 0)

	m := map[string]interface{}{}

	if cl.Daily != nil {
		m["daily"] = *cl.Daily
	}
	if cl.MaximumDuration != nil {
		m["maximum_duration"] = *cl.MaximumDuration
	}
	if cl.MessagesPerSecond != nil {
		m["messages_per_second"] = *cl.MessagesPerSecond
	}

	if cl.Total != nil {
		m["total"] = *cl.Total
	}

	if len(m) <= 0 {
		return nil
	}

	l = append(l, m)

	return l
}

func expandQuietTime(s *schema.Set) *pinpoint.QuietTime {
	if s == nil || s.Len() == 0 {
		return nil
	}

	m := s.List()[0].(map[string]interface{})

	qt := pinpoint.QuietTime{}

	if m["end"] != nil {
		qt.End = aws.String(m["end"].(string))
	}

	if m["start"] != nil {
		qt.Start = aws.String(m["start"].(string))
	}

	return &qt
}

func flattenQuietTime(qt *pinpoint.QuietTime) []interface{} {
	l := make([]interface{}, 0)

	m := map[string]interface{}{}

	if qt.End != nil {
		m["end"] = qt.End
	}
	if qt.Start != nil {
		m["start"] = qt.Start
	}

	if len(m) <= 0 {
		return nil
	}

	l = append(l, m)

	return l
}

// Assemble the hash for the aws_pinpoint_app campaignHook
// TypeSet attribute.
func campaignHookHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["lambda_function_name"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["mode"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["web_url"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return hashcode.String(buf.String())
}

// Assemble the hash for the aws_pinpoint_app campaignLimits
// TypeSet attribute.
func campaignLimitsHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["daily"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["maximum_duration"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["messages_per_second"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["total"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	return hashcode.String(buf.String())
}

// Assemble the hash for the aws_pinpoint_app quietTime
// TypeSet attribute.
func quietTimeHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["end"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	if v, ok := m["start"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}
	return hashcode.String(buf.String())
}
