package apigatewayv2

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	defaultStageName = "$default"
)

func ResourceStage() *schema.Resource {
	return &schema.Resource{
		Create: resourceStageCreate,
		Read:   resourceStageRead,
		Update: resourceStageUpdate,
		Delete: resourceStageDelete,
		Importer: &schema.ResourceImporter{
			State: resourceStageImport,
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
							ValidateFunc: verify.ValidARN,
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
			"auto_deploy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
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
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
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
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								apigatewayv2.LoggingLevelError,
								apigatewayv2.LoggingLevelInfo,
								apigatewayv2.LoggingLevelOff,
							}, false),
						},
						"throttling_burst_limit": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"throttling_rate_limit": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
					},
				},
			},
			"deployment_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
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
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								apigatewayv2.LoggingLevelError,
								apigatewayv2.LoggingLevelInfo,
								apigatewayv2.LoggingLevelOff,
							}, false),
						},
						"route_key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"throttling_burst_limit": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"throttling_rate_limit": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
					},
				},
			},
			"stage_variables": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStageCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	apiId := d.Get("api_id").(string)

	apiOutput, err := conn.GetApi(&apigatewayv2.GetApiInput{
		ApiId: aws.String(apiId),
	})
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 API (%s): %s", apiId, err)
	}

	protocolType := aws.StringValue(apiOutput.ProtocolType)

	req := &apigatewayv2.CreateStageInput{
		ApiId:      aws.String(apiId),
		AutoDeploy: aws.Bool(d.Get("auto_deploy").(bool)),
		StageName:  aws.String(d.Get("name").(string)),
		Tags:       Tags(tags.IgnoreAWS()),
	}
	if v, ok := d.GetOk("access_log_settings"); ok {
		req.AccessLogSettings = expandAccessLogSettings(v.([]interface{}))
	}
	if v, ok := d.GetOk("client_certificate_id"); ok {
		req.ClientCertificateId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("default_route_settings"); ok {
		req.DefaultRouteSettings = expandDefaultRouteSettings(v.([]interface{}), protocolType)
	}
	if v, ok := d.GetOk("deployment_id"); ok {
		req.DeploymentId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		req.Description = aws.String(v.(string))
	}
	if v, ok := d.GetOk("route_settings"); ok {
		req.RouteSettings = expandRouteSettings(v.(*schema.Set).List(), protocolType)
	}
	if v, ok := d.GetOk("stage_variables"); ok {
		req.StageVariables = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	log.Printf("[DEBUG] Creating API Gateway v2 stage: %s", req)
	resp, err := conn.CreateStage(req)
	if err != nil {
		return fmt.Errorf("error creating API Gateway v2 stage: %s", err)
	}

	d.SetId(aws.StringValue(resp.StageName))

	return resourceStageRead(d, meta)
}

func resourceStageRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	apiId := d.Get("api_id").(string)
	resp, err := conn.GetStage(&apigatewayv2.GetStageInput{
		ApiId:     aws.String(apiId),
		StageName: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 stage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 stage (%s): %s", d.Id(), err)
	}

	stageName := aws.StringValue(resp.StageName)
	err = d.Set("access_log_settings", flattenAccessLogSettings(resp.AccessLogSettings))
	if err != nil {
		return fmt.Errorf("error setting access_log_settings: %s", err)
	}
	region := meta.(*conns.AWSClient).Region
	resourceArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    region,
		Resource:  fmt.Sprintf("/apis/%s/stages/%s", apiId, stageName),
	}.String()
	d.Set("arn", resourceArn)
	d.Set("auto_deploy", resp.AutoDeploy)
	d.Set("client_certificate_id", resp.ClientCertificateId)
	err = d.Set("default_route_settings", flattenDefaultRouteSettings(resp.DefaultRouteSettings))
	if err != nil {
		return fmt.Errorf("error setting default_route_settings: %s", err)
	}
	d.Set("deployment_id", resp.DeploymentId)
	d.Set("description", resp.Description)
	executionArn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "execute-api",
		Region:    region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("%s/%s", apiId, stageName),
	}.String()
	d.Set("execution_arn", executionArn)
	d.Set("name", stageName)
	err = d.Set("route_settings", flattenRouteSettings(resp.RouteSettings))
	if err != nil {
		return fmt.Errorf("error setting route_settings: %s", err)
	}
	err = d.Set("stage_variables", flex.PointersMapToStringList(resp.StageVariables))
	if err != nil {
		return fmt.Errorf("error setting stage_variables: %s", err)
	}

	tags := KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	apiOutput, err := conn.GetApi(&apigatewayv2.GetApiInput{
		ApiId: aws.String(apiId),
	})
	if err != nil {
		return fmt.Errorf("error reading API Gateway v2 API (%s): %s", apiId, err)
	}

	switch aws.StringValue(apiOutput.ProtocolType) {
	case apigatewayv2.ProtocolTypeWebsocket:
		d.Set("invoke_url", fmt.Sprintf("wss://%s.execute-api.%s.amazonaws.com/%s", apiId, region, stageName))
	case apigatewayv2.ProtocolTypeHttp:
		if stageName == defaultStageName {
			d.Set("invoke_url", fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/", apiId, region))
		} else {
			d.Set("invoke_url", fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s", apiId, region, stageName))
		}
	}

	return nil
}

func resourceStageUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	if d.HasChanges("access_log_settings", "auto_deploy", "client_certificate_id",
		"default_route_settings", "deployment_id", "description",
		"route_settings", "stage_variables") {
		apiId := d.Get("api_id").(string)

		apiOutput, err := conn.GetApi(&apigatewayv2.GetApiInput{
			ApiId: aws.String(apiId),
		})
		if err != nil {
			return fmt.Errorf("error reading API Gateway v2 API (%s): %s", apiId, err)
		}

		protocolType := aws.StringValue(apiOutput.ProtocolType)

		req := &apigatewayv2.UpdateStageInput{
			ApiId:     aws.String(apiId),
			StageName: aws.String(d.Id()),
		}
		if d.HasChange("access_log_settings") {
			req.AccessLogSettings = expandAccessLogSettings(d.Get("access_log_settings").([]interface{}))
		}
		if d.HasChange("auto_deploy") {
			req.AutoDeploy = aws.Bool(d.Get("auto_deploy").(bool))
		}
		if d.HasChange("client_certificate_id") {
			req.ClientCertificateId = aws.String(d.Get("client_certificate_id").(string))
		}
		if d.HasChange("default_route_settings") {
			req.DefaultRouteSettings = expandDefaultRouteSettings(d.Get("default_route_settings").([]interface{}), protocolType)
		}
		if d.HasChange("deployment_id") {
			req.DeploymentId = aws.String(d.Get("deployment_id").(string))
		}
		if d.HasChange("description") {
			req.Description = aws.String(d.Get("description").(string))
		}
		if d.HasChange("route_settings") {
			o, n := d.GetChange("route_settings")
			os := o.(*schema.Set)
			ns := n.(*schema.Set)

			for _, vRouteSetting := range os.Difference(ns).List() {
				routeKey := vRouteSetting.(map[string]interface{})["route_key"].(string)

				log.Printf("[DEBUG] Deleting API Gateway v2 stage (%s) route settings (%s)", d.Id(), routeKey)
				_, err := conn.DeleteRouteSettings(&apigatewayv2.DeleteRouteSettingsInput{
					ApiId:     aws.String(d.Get("api_id").(string)),
					RouteKey:  aws.String(routeKey),
					StageName: aws.String(d.Id()),
				})
				if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
					continue
				}
				if err != nil {
					return fmt.Errorf("error deleting API Gateway v2 stage (%s) route settings (%s): %w", d.Id(), routeKey, err)
				}
			}

			req.RouteSettings = expandRouteSettings(ns.List(), protocolType)
		}
		if d.HasChange("stage_variables") {
			o, n := d.GetChange("stage_variables")
			add, del, _ := verify.DiffStringMaps(o.(map[string]interface{}), n.(map[string]interface{}))
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

		log.Printf("[DEBUG] Updating API Gateway v2 stage: %s", req)
		_, err = conn.UpdateStage(req)
		if err != nil {
			return fmt.Errorf("error updating API Gateway v2 stage (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating API Gateway v2 stage (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceStageRead(d, meta)
}

func resourceStageDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	log.Printf("[DEBUG] Deleting API Gateway v2 stage (%s)", d.Id())
	_, err := conn.DeleteStage(&apigatewayv2.DeleteStageInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		StageName: aws.String(d.Id()),
	})
	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting API Gateway v2 stage (%s): %s", d.Id(), err)
	}

	return nil
}

func resourceStageImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'api-id/stage-name'", d.Id())
	}

	apiId := parts[0]
	stageName := parts[1]

	conn := meta.(*conns.AWSClient).APIGatewayV2Conn

	resp, err := conn.GetStage(&apigatewayv2.GetStageInput{
		ApiId:     aws.String(apiId),
		StageName: aws.String(stageName),
	})
	if err != nil {
		return nil, err
	}

	if aws.BoolValue(resp.ApiGatewayManaged) {
		return nil, fmt.Errorf("API Gateway v2 stage (%s) was created via quick create", stageName)
	}

	d.SetId(stageName)
	d.Set("api_id", apiId)

	return []*schema.ResourceData{d}, nil
}

func expandAccessLogSettings(vSettings []interface{}) *apigatewayv2.AccessLogSettings {
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

func flattenAccessLogSettings(settings *apigatewayv2.AccessLogSettings) []interface{} {
	if settings == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"destination_arn": aws.StringValue(settings.DestinationArn),
		"format":          aws.StringValue(settings.Format),
	}}
}

func expandDefaultRouteSettings(vSettings []interface{}, protocolType string) *apigatewayv2.RouteSettings {
	routeSettings := &apigatewayv2.RouteSettings{}

	if len(vSettings) == 0 || vSettings[0] == nil {
		return routeSettings
	}
	mSettings := vSettings[0].(map[string]interface{})

	if vDataTraceEnabled, ok := mSettings["data_trace_enabled"].(bool); ok && protocolType == apigatewayv2.ProtocolTypeWebsocket {
		routeSettings.DataTraceEnabled = aws.Bool(vDataTraceEnabled)
	}
	if vDetailedMetricsEnabled, ok := mSettings["detailed_metrics_enabled"].(bool); ok {
		routeSettings.DetailedMetricsEnabled = aws.Bool(vDetailedMetricsEnabled)
	}
	if vLoggingLevel, ok := mSettings["logging_level"].(string); ok && vLoggingLevel != "" && protocolType == apigatewayv2.ProtocolTypeWebsocket {
		routeSettings.LoggingLevel = aws.String(vLoggingLevel)
	}
	if vThrottlingBurstLimit, ok := mSettings["throttling_burst_limit"].(int); ok {
		routeSettings.ThrottlingBurstLimit = aws.Int64(int64(vThrottlingBurstLimit))
	}
	if vThrottlingRateLimit, ok := mSettings["throttling_rate_limit"].(float64); ok {
		routeSettings.ThrottlingRateLimit = aws.Float64(vThrottlingRateLimit)
	}

	return routeSettings
}

func flattenDefaultRouteSettings(routeSettings *apigatewayv2.RouteSettings) []interface{} {
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

func expandRouteSettings(vSettings []interface{}, protocolType string) map[string]*apigatewayv2.RouteSettings {
	settings := map[string]*apigatewayv2.RouteSettings{}

	for _, v := range vSettings {
		routeSettings := &apigatewayv2.RouteSettings{}

		mSettings := v.(map[string]interface{})

		if v, ok := mSettings["data_trace_enabled"].(bool); ok && protocolType == apigatewayv2.ProtocolTypeWebsocket {
			routeSettings.DataTraceEnabled = aws.Bool(v)
		}
		if v, ok := mSettings["detailed_metrics_enabled"].(bool); ok {
			routeSettings.DetailedMetricsEnabled = aws.Bool(v)
		}
		if v, ok := mSettings["logging_level"].(string); ok && v != "" && protocolType == apigatewayv2.ProtocolTypeWebsocket {
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

func flattenRouteSettings(settings map[string]*apigatewayv2.RouteSettings) []interface{} {
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

	return vSettings
}
