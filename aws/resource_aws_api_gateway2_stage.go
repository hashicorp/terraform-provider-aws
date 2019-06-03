package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

const (
	apigatewayv2LoggingLevelOff = "OFF"
)

func resourceAwsApiGateway2Stage() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGateway2StageCreate,
		Read:   resourceAwsApiGateway2StageRead,
		Update: resourceAwsApiGateway2StageUpdate,
		Delete: resourceAwsApiGateway2StageDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsApiGateway2StageImport,
		},

		Schema: map[string]*schema.Schema{
			"access_log_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"destination_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
							StateFunc: func(v interface{}) string {
								return strings.TrimSuffix(v.(string), ":*")
							},
						},
						"format": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_certificate_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"default_route_settings": {
				Type:             schema.TypeList,
				Optional:         true,
				MinItems:         0,
				MaxItems:         1,
				DiffSuppressFunc: suppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_trace_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"detailed_metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"logging_level": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  apigatewayv2LoggingLevelOff,
							ValidateFunc: validation.StringInSlice([]string{
								apigatewayv2.LoggingLevelError,
								apigatewayv2.LoggingLevelInfo,
								apigatewayv2LoggingLevelOff,
							}, false),
						},
						"throttling_burst_limit": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  5000,
						},
						"throttling_rate_limit": {
							Type:     schema.TypeFloat,
							Optional: true,
							Default:  10000.00,
						},
					},
				},
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"execution_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"invoke_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"route_settings": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_trace_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"detailed_metrics_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"logging_level": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  apigatewayv2LoggingLevelOff,
							ValidateFunc: validation.StringInSlice([]string{
								apigatewayv2.LoggingLevelError,
								apigatewayv2.LoggingLevelInfo,
								apigatewayv2LoggingLevelOff,
							}, false),
						},
						"route_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"throttling_burst_limit": {
							Type:     schema.TypeInt,
							Default:  5000,
							Optional: true,
						},
						"throttling_rate_limit": {
							Type:     schema.TypeFloat,
							Optional: true,
							Default:  10000.00,
						},
					},
				},
				Set: apiGateway2RouteSettingsHash,
			},
			"stage_variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags": tagsSchema(),
		},
	}
}

func resourceAwsApiGateway2StageCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	req := &apigatewayv2.CreateStageInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		StageName: aws.String(d.Get("name").(string)),
		Tags:      keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().Apigatewayv2Tags(),
	}
	if v, ok := d.GetOk("access_log_settings"); ok {
		req.AccessLogSettings = expandApiGateway2AccessLogSettings(v.([]interface{}))
	}
	if v, ok := d.GetOk("client_certificate_id"); ok {
		req.ClientCertificateId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("default_route_settings"); ok {
		req.DefaultRouteSettings = expandApiGateway2DefaultRouteSettings(v.([]interface{}))
	}
	if v, ok := d.GetOk("deployment_id"); ok {
		req.DeploymentId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("route_settings"); ok {
		req.RouteSettings = expandApiGateway2RouteSettings(v.(*schema.Set))
	}
	if v, ok := d.GetOk("stage_variables"); ok {
		req.StageVariables = stringMapToPointers(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 stage: %s", req)
	resp, err := conn.CreateStage(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 stage: %s", err)
	}

	d.SetId(aws.StringValue(resp.StageName))

	return resourceAwsApiGateway2StageRead(d, meta)
}

func resourceAwsApiGateway2StageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	apiId := d.Get("api_id").(string)
	resp, err := conn.GetStage(&apigatewayv2.GetStageInput{
		ApiId:     aws.String(apiId),
		StageName: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] API Gateway v2 stage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 stage: %s", err)
	}

	stageName := aws.StringValue(resp.StageName)
	err = d.Set("access_log_settings", flattenApiGateway2AccessLogSettings(resp.AccessLogSettings))
	if err != nil {
		return fmt.Errorf("error setting access_log_settings: %s", err)
	}
	region := meta.(*AWSClient).region
	resourceArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "apigateway",
		Region:    region,
		Resource:  fmt.Sprintf("/apis/%s/stages/%s", apiId, stageName),
	}.String()
	d.Set("arn", resourceArn)
	d.Set("client_certificate_id", resp.ClientCertificateId)
	err = d.Set("default_route_settings", flattenApiGateway2DefaultRouteSettings(resp.DefaultRouteSettings))
	if err != nil {
		return fmt.Errorf("error setting default_route_settings: %s", err)
	}
	d.Set("deployment_id", resp.DeploymentId)
	d.Set("description", resp.Description)
	executionArn := arn.ARN{
		Partition: meta.(*AWSClient).partition,
		Service:   "execute-api",
		Region:    region,
		AccountID: meta.(*AWSClient).accountid,
		Resource:  fmt.Sprintf("%s/%s", apiId, stageName),
	}.String()
	d.Set("execution_arn", executionArn)
	d.Set("invoke_url", fmt.Sprintf("wss://%s.execute-api.%s.amazonaws.com/%s", apiId, region, stageName))
	d.Set("name", stageName)
	err = d.Set("route_settings", flattenApiGateway2RouteSettings(resp.RouteSettings))
	if err != nil {
		return fmt.Errorf("error setting route_settings: %s", err)
	}
	err = d.Set("stage_variables", pointersMapToStringList(resp.StageVariables))
	if err != nil {
		return fmt.Errorf("error setting stage_variables: %s", err)
	}
	if err := d.Set("tags", keyvaluetags.Apigatewayv2KeyValueTags(resp.Tags).IgnoreAws().Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsApiGateway2StageUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	updateStage := false
	req := &apigatewayv2.UpdateStageInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		StageName: aws.String(d.Id()),
	}
	if d.HasChange("access_log_settings") {
		updateStage = true
		req.AccessLogSettings = expandApiGateway2AccessLogSettings(d.Get("access_log_settings").([]interface{}))
	}
	if d.HasChange("client_certificate_id") {
		updateStage = true
		req.ClientCertificateId = aws.String(d.Get("client_certificate_id").(string))
	}
	if d.HasChange("default_route_settings") {
		updateStage = true
		req.DefaultRouteSettings = expandApiGateway2DefaultRouteSettings(d.Get("default_route_settings").([]interface{}))
	}
	if d.HasChange("deployment_id") {
		updateStage = true
		req.DeploymentId = aws.String(d.Get("deployment_id").(string))
	}
	if d.HasChange("description") {
		updateStage = true
		req.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("route_settings") {
		req.RouteSettings = expandApiGateway2RouteSettings(d.Get("route_settings").(*schema.Set))
	}
	if d.HasChange("stage_variables") {
		updateStage = true
		oraw, nraw := d.GetChange("stage_variables")
		add, del := diffStringMaps(oraw.(map[string]interface{}), nraw.(map[string]interface{}))
		// Variables are removed by setting the associated value to "".
		for k := range del {
			del[k] = aws.String("")
		}
		variables := del
		for k, v := range add {
			variables[k] = v
		}
		req.StageVariables = variables
	}

	if updateStage {
		log.Printf("[DEBUG] Updating API Gateway v2 stage: %s", req)
		_, err := conn.UpdateStage(req)
		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 stage: %s", err)
		}
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")
		if err := keyvaluetags.Apigatewayv2UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating API Gateway v2 stage (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceAwsApiGateway2StageRead(d, meta)
}

func resourceAwsApiGateway2StageDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigatewayv2conn

	log.Printf("[DEBUG] Deleting API Gateway v2 stage (%s)", d.Id())
	_, err := conn.DeleteStage(&apigatewayv2.DeleteStageInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		StageName: aws.String(d.Id()),
	})
	if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 stage: %s", err)
	}

	return nil
}

func resourceAwsApiGateway2StageImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/stage-name'", d.Id())
	}

	d.SetId(parts[1])
	d.Set("api_id", parts[0])

	return []*schema.ResourceData{d}, nil
}

func expandApiGateway2AccessLogSettings(vSettings []interface{}) *apigatewayv2.AccessLogSettings {
	settings := &apigatewayv2.AccessLogSettings{}

	if len(vSettings) == 0 || vSettings[0] == nil {
		return settings
	}
	mSettings := vSettings[0].(map[string]interface{})

	if vDestinationArn, ok := mSettings["destination_arn"].(string); ok && vDestinationArn != "" {
		settings.DestinationArn = aws.String(vDestinationArn)
	}
	if vFormat, ok := mSettings["format"].(string); ok && vFormat != "" {
		settings.Format = aws.String(vFormat)
	}

	return settings
}

func flattenApiGateway2AccessLogSettings(settings *apigatewayv2.AccessLogSettings) []interface{} {
	if settings == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"destination_arn": aws.StringValue(settings.DestinationArn),
		"format":          aws.StringValue(settings.Format),
	}}
}

func expandApiGateway2DefaultRouteSettings(vSettings []interface{}) *apigatewayv2.RouteSettings {
	routeSettings := &apigatewayv2.RouteSettings{}

	if len(vSettings) == 0 || vSettings[0] == nil {
		return routeSettings
	}
	mSettings := vSettings[0].(map[string]interface{})

	if vDataTraceEnabled, ok := mSettings["data_trace_enabled"].(bool); ok {
		routeSettings.DataTraceEnabled = aws.Bool(vDataTraceEnabled)
	}
	if vDetailedMetricsEnabled, ok := mSettings["detailed_metrics_enabled"].(bool); ok {
		routeSettings.DetailedMetricsEnabled = aws.Bool(vDetailedMetricsEnabled)
	}
	if vLoggingLevel, ok := mSettings["logging_level"].(string); ok && vLoggingLevel != "" {
		routeSettings.LoggingLevel = aws.String(vLoggingLevel)
	}
	if vThrottlingBurstLimit, ok := mSettings["throttling_burst_limit"].(int); ok && vThrottlingBurstLimit > 0 {
		routeSettings.ThrottlingBurstLimit = aws.Int64(int64(vThrottlingBurstLimit))
	}
	if vThrottlingRateLimit, ok := mSettings["throttling_rate_limit"].(float64); ok && vThrottlingRateLimit > 0.0 {
		routeSettings.ThrottlingRateLimit = aws.Float64(vThrottlingRateLimit)
	}

	return routeSettings
}

func flattenApiGateway2DefaultRouteSettings(routeSettings *apigatewayv2.RouteSettings) []interface{} {
	if routeSettings == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"data_trace_enabled":       aws.BoolValue(routeSettings.DataTraceEnabled),
		"detailed_metrics_enabled": aws.BoolValue(routeSettings.DetailedMetricsEnabled),
		"logging_level":            aws.StringValue(routeSettings.LoggingLevel),
		"throttling_burst_limit":   int(aws.Int64Value(routeSettings.ThrottlingBurstLimit)),
		"throttling_rate_limit":    aws.Float64Value(routeSettings.ThrottlingRateLimit),
	}}
}

func expandApiGateway2RouteSettings(vSettings *schema.Set) map[string]*apigatewayv2.RouteSettings {
	settings := map[string]*apigatewayv2.RouteSettings{}

	for _, v := range vSettings.List() {
		routeSettings := &apigatewayv2.RouteSettings{}

		mSettings := v.(map[string]interface{})

		if v, ok := mSettings["data_trace_enabled"].(bool); ok {
			routeSettings.DataTraceEnabled = aws.Bool(v)
		}
		if v, ok := mSettings["detailed_metrics_enabled"].(bool); ok {
			routeSettings.DetailedMetricsEnabled = aws.Bool(v)
		}
		if v, ok := mSettings["logging_level"].(string); ok {
			routeSettings.LoggingLevel = aws.String(v)
		}
		if v, ok := mSettings["throttling_burst_limit"].(int); ok {
			routeSettings.ThrottlingBurstLimit = aws.Int64(int64(v))
		}
		if v, ok := mSettings["throttling_rate_limit"].(float64); ok {
			routeSettings.ThrottlingRateLimit = aws.Float64(v)
		}

		settings[mSettings["route_key"].(string)] = routeSettings
	}

	return settings
}

func flattenApiGateway2RouteSettings(settings map[string]*apigatewayv2.RouteSettings) *schema.Set {
	vSettings := []interface{}{}

	for k, routeSetting := range settings {
		vSettings = append(vSettings, map[string]interface{}{
			"data_trace_enabled":       aws.BoolValue(routeSetting.DataTraceEnabled),
			"detailed_metrics_enabled": aws.BoolValue(routeSetting.DetailedMetricsEnabled),
			"logging_level":            aws.StringValue(routeSetting.LoggingLevel),
			"route_key":                k,
			"throttling_burst_limit":   int(aws.Int64Value(routeSetting.ThrottlingBurstLimit)),
			"throttling_rate_limit":    aws.Float64Value(routeSetting.ThrottlingRateLimit),
		})
	}

	return schema.NewSet(apiGateway2RouteSettingsHash, vSettings)
}

func apiGateway2RouteSettingsHash(vSettings interface{}) int {
	var buf bytes.Buffer

	mSettings := vSettings.(map[string]interface{})

	if v, ok := mSettings["route_key"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := mSettings["data_trace_enabled"].(bool); ok {
		buf.WriteString(fmt.Sprintf("%t-", v))
	}
	if v, ok := mSettings["detailed_metrics_enabled"].(bool); ok {
		buf.WriteString(fmt.Sprintf("%t-", v))
	}
	if v, ok := mSettings["logging_level"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := mSettings["throttling_burst_limit"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}
	if v, ok := mSettings["throttling_rate_limit"].(float64); ok {
		buf.WriteString(fmt.Sprintf("%g-", v))
	}

	return hashcode.String(buf.String())
}
