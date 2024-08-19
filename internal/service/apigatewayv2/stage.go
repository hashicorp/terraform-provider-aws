// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// @SDKResource("aws_apigatewayv2_stage", name="Stage")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigatewayv2;apigatewayv2.GetStageOutput")
// @Testing(importStateIdFunc="testAccStageImportStateIdFunc")
func resourceStage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceStageCreate,
		ReadWithoutTimeout:   resourceStageRead,
		UpdateWithoutTimeout: resourceStageUpdate,
		DeleteWithoutTimeout: resourceStageDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceStageImport,
		},

		Schema: map[string]*schema.Schema{
			"access_log_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDestinationARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrFormat: {
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
			names.AttrARN: {
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
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LoggingLevel](),
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
			names.AttrDescription: {
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
			names.AttrName: {
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
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.LoggingLevel](),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceStageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	apiID := d.Get("api_id").(string)
	outputGA, err := findAPIByID(ctx, conn, apiID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 API (%s): %s", apiID, err)
	}

	name := d.Get(names.AttrName).(string)
	protocolType := outputGA.ProtocolType
	input := &apigatewayv2.CreateStageInput{
		ApiId:      aws.String(apiID),
		AutoDeploy: aws.Bool(d.Get("auto_deploy").(bool)),
		StageName:  aws.String(name),
		Tags:       getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_log_settings"); ok {
		input.AccessLogSettings = expandAccessLogSettings(v.([]interface{}))
	}

	if v, ok := d.GetOk("client_certificate_id"); ok {
		input.ClientCertificateId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("default_route_settings"); ok {
		input.DefaultRouteSettings = expandDefaultRouteSettings(v.([]interface{}), protocolType)
	}

	if v, ok := d.GetOk("deployment_id"); ok {
		input.DeploymentId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("route_settings"); ok {
		input.RouteSettings = expandRouteSettings(v.(*schema.Set).List(), protocolType)
	}

	if v, ok := d.GetOk("stage_variables"); ok {
		input.StageVariables = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	output, err := conn.CreateStage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 Stage (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.StageName))

	return append(diags, resourceStageRead(ctx, d, meta)...)
}

func resourceStageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	apiID := d.Get("api_id").(string)
	outputGS, err := findStageByTwoPartKey(ctx, conn, apiID, d.Id())

	if errs.IsA[*awstypes.NotFoundException](err) && !d.IsNewResource() {
		log.Printf("[WARN] API Gateway v2 Stage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 Stage (%s): %s", d.Id(), err)
	}

	stageName := aws.ToString(outputGS.StageName)
	if err := d.Set("access_log_settings", flattenAccessLogSettings(outputGS.AccessLogSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting access_log_settings: %s", err)
	}
	d.Set(names.AttrARN, stageARN(meta.(*conns.AWSClient), apiID, stageName))
	d.Set("auto_deploy", outputGS.AutoDeploy)
	d.Set("client_certificate_id", outputGS.ClientCertificateId)
	if err := d.Set("default_route_settings", flattenDefaultRouteSettings(outputGS.DefaultRouteSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting default_route_settings: %s", err)
	}
	d.Set("deployment_id", outputGS.DeploymentId)
	d.Set(names.AttrDescription, outputGS.Description)
	d.Set("execution_arn", stageInvokeARN(meta.(*conns.AWSClient), apiID, stageName))
	d.Set(names.AttrName, stageName)
	if err := d.Set("route_settings", flattenRouteSettings(outputGS.RouteSettings)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting route_settings: %s", err)
	}
	d.Set("stage_variables", outputGS.StageVariables)

	setTagsOut(ctx, outputGS.Tags)

	outputGA, err := findAPIByID(ctx, conn, apiID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 API (%s): %s", apiID, err)
	}

	d.Set("invoke_url", meta.(*conns.AWSClient).APIGatewayV2InvokeURL(ctx, outputGA.ProtocolType, apiID, stageName))

	return diags
}

func resourceStageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	if d.HasChanges("access_log_settings", "auto_deploy", "client_certificate_id",
		"default_route_settings", "deployment_id", names.AttrDescription,
		"route_settings", "stage_variables") {
		apiID := d.Get("api_id").(string)
		outputGA, err := findAPIByID(ctx, conn, apiID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 API (%s): %s", apiID, err)
		}

		protocolType := outputGA.ProtocolType
		input := &apigatewayv2.UpdateStageInput{
			ApiId:     aws.String(apiID),
			StageName: aws.String(d.Id()),
		}

		if d.HasChange("access_log_settings") {
			input.AccessLogSettings = expandAccessLogSettings(d.Get("access_log_settings").([]interface{}))
		}

		if d.HasChange("auto_deploy") {
			input.AutoDeploy = aws.Bool(d.Get("auto_deploy").(bool))
		}

		if d.HasChange("client_certificate_id") {
			input.ClientCertificateId = aws.String(d.Get("client_certificate_id").(string))
		}

		if d.HasChange("default_route_settings") {
			input.DefaultRouteSettings = expandDefaultRouteSettings(d.Get("default_route_settings").([]interface{}), protocolType)
		}

		if d.HasChange("deployment_id") {
			input.DeploymentId = aws.String(d.Get("deployment_id").(string))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange("route_settings") {
			o, n := d.GetChange("route_settings")
			os := o.(*schema.Set)
			ns := n.(*schema.Set)

			for _, vRouteSetting := range os.Difference(ns).List() {
				routeKey := vRouteSetting.(map[string]interface{})["route_key"].(string)
				input := &apigatewayv2.DeleteRouteSettingsInput{
					ApiId:     aws.String(d.Get("api_id").(string)),
					RouteKey:  aws.String(routeKey),
					StageName: aws.String(d.Id()),
				}

				_, err := conn.DeleteRouteSettings(ctx, input)

				if errs.IsA[*awstypes.NotFoundException](err) {
					continue
				}

				if err != nil {
					return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Stage (%s) route settings (%s): %s", d.Id(), routeKey, err)
				}
			}

			input.RouteSettings = expandRouteSettings(ns.List(), protocolType)
		}

		if d.HasChange("stage_variables") {
			o, n := d.GetChange("stage_variables")
			add, del, _ := flex.DiffStringValueMaps(o.(map[string]interface{}), n.(map[string]interface{}))
			// Variables are removed by setting the associated value to "".
			for k := range del {
				del[k] = ""
			}
			variables := del
			for k, v := range add {
				variables[k] = v
			}
			input.StageVariables = variables
		}

		_, err = conn.UpdateStage(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 Stage (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceStageRead(ctx, d, meta)...)
}

func resourceStageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 Stage: %s", d.Id())
	_, err := conn.DeleteStage(ctx, &apigatewayv2.DeleteStageInput{
		ApiId:     aws.String(d.Get("api_id").(string)),
		StageName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Stage (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceStageImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'api-id/stage-name'", d.Id())
	}

	apiID := parts[0]
	stageName := parts[1]

	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	output, err := findStageByTwoPartKey(ctx, conn, apiID, stageName)

	if err != nil {
		return nil, err
	}

	if aws.ToBool(output.ApiGatewayManaged) {
		return nil, fmt.Errorf("API Gateway v2 Stage (%s) was created via quick create", stageName)
	}

	d.SetId(stageName)
	d.Set("api_id", apiID)

	return []*schema.ResourceData{d}, nil
}

func findStageByTwoPartKey(ctx context.Context, conn *apigatewayv2.Client, apiID, stageName string) (*apigatewayv2.GetStageOutput, error) {
	input := &apigatewayv2.GetStageInput{
		ApiId:     aws.String(apiID),
		StageName: aws.String(stageName),
	}

	return findStage(ctx, conn, input)
}

func findStage(ctx context.Context, conn *apigatewayv2.Client, input *apigatewayv2.GetStageInput) (*apigatewayv2.GetStageOutput, error) {
	output, err := conn.GetStage(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
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

func expandAccessLogSettings(vSettings []interface{}) *awstypes.AccessLogSettings {
	settings := &awstypes.AccessLogSettings{}

	if len(vSettings) == 0 || vSettings[0] == nil {
		return settings
	}
	mSettings := vSettings[0].(map[string]interface{})

	if vDestinationArn, ok := mSettings[names.AttrDestinationARN].(string); ok && vDestinationArn != "" {
		settings.DestinationArn = aws.String(vDestinationArn)
	}
	if vFormat, ok := mSettings[names.AttrFormat].(string); ok && vFormat != "" {
		settings.Format = aws.String(vFormat)
	}

	return settings
}

func flattenAccessLogSettings(settings *awstypes.AccessLogSettings) []interface{} {
	if settings == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		names.AttrDestinationARN: aws.ToString(settings.DestinationArn),
		names.AttrFormat:         aws.ToString(settings.Format),
	}}
}

func expandDefaultRouteSettings(vSettings []interface{}, protocolType awstypes.ProtocolType) *awstypes.RouteSettings {
	routeSettings := &awstypes.RouteSettings{}

	if len(vSettings) == 0 || vSettings[0] == nil {
		return routeSettings
	}
	mSettings := vSettings[0].(map[string]interface{})

	if vDataTraceEnabled, ok := mSettings["data_trace_enabled"].(bool); ok && protocolType == awstypes.ProtocolTypeWebsocket {
		routeSettings.DataTraceEnabled = aws.Bool(vDataTraceEnabled)
	}
	if vDetailedMetricsEnabled, ok := mSettings["detailed_metrics_enabled"].(bool); ok {
		routeSettings.DetailedMetricsEnabled = aws.Bool(vDetailedMetricsEnabled)
	}
	if vLoggingLevel, ok := mSettings["logging_level"].(string); ok && vLoggingLevel != "" && protocolType == awstypes.ProtocolTypeWebsocket {
		routeSettings.LoggingLevel = awstypes.LoggingLevel(vLoggingLevel)
	}
	if vThrottlingBurstLimit, ok := mSettings["throttling_burst_limit"].(int); ok {
		routeSettings.ThrottlingBurstLimit = aws.Int32(int32(vThrottlingBurstLimit))
	}
	if vThrottlingRateLimit, ok := mSettings["throttling_rate_limit"].(float64); ok {
		routeSettings.ThrottlingRateLimit = aws.Float64(vThrottlingRateLimit)
	}

	return routeSettings
}

func flattenDefaultRouteSettings(routeSettings *awstypes.RouteSettings) []interface{} {
	if routeSettings == nil {
		return []interface{}{}
	}

	return []interface{}{map[string]interface{}{
		"data_trace_enabled":       aws.ToBool(routeSettings.DataTraceEnabled),
		"detailed_metrics_enabled": aws.ToBool(routeSettings.DetailedMetricsEnabled),
		"logging_level":            string(routeSettings.LoggingLevel),
		"throttling_burst_limit":   int(aws.ToInt32(routeSettings.ThrottlingBurstLimit)),
		"throttling_rate_limit":    aws.ToFloat64(routeSettings.ThrottlingRateLimit),
	}}
}

func expandRouteSettings(vSettings []interface{}, protocolType awstypes.ProtocolType) map[string]awstypes.RouteSettings {
	settings := map[string]awstypes.RouteSettings{}

	for _, v := range vSettings {
		routeSettings := awstypes.RouteSettings{}

		mSettings := v.(map[string]interface{})

		if v, ok := mSettings["data_trace_enabled"].(bool); ok && protocolType == awstypes.ProtocolTypeWebsocket {
			routeSettings.DataTraceEnabled = aws.Bool(v)
		}
		if v, ok := mSettings["detailed_metrics_enabled"].(bool); ok {
			routeSettings.DetailedMetricsEnabled = aws.Bool(v)
		}
		if v, ok := mSettings["logging_level"].(string); ok && v != "" && protocolType == awstypes.ProtocolTypeWebsocket {
			routeSettings.LoggingLevel = awstypes.LoggingLevel(v)
		}
		if v, ok := mSettings["throttling_burst_limit"].(int); ok {
			routeSettings.ThrottlingBurstLimit = aws.Int32(int32(v))
		}
		if v, ok := mSettings["throttling_rate_limit"].(float64); ok {
			routeSettings.ThrottlingRateLimit = aws.Float64(v)
		}

		settings[mSettings["route_key"].(string)] = routeSettings
	}

	return settings
}

func flattenRouteSettings(settings map[string]awstypes.RouteSettings) []interface{} {
	vSettings := []interface{}{}

	for k, routeSetting := range settings {
		vSettings = append(vSettings, map[string]interface{}{
			"data_trace_enabled":       aws.ToBool(routeSetting.DataTraceEnabled),
			"detailed_metrics_enabled": aws.ToBool(routeSetting.DetailedMetricsEnabled),
			"logging_level":            routeSetting.LoggingLevel,
			"route_key":                k,
			"throttling_burst_limit":   int(aws.ToInt32(routeSetting.ThrottlingBurstLimit)),
			"throttling_rate_limit":    aws.ToFloat64(routeSetting.ThrottlingRateLimit),
		})
	}

	return vSettings
}

func stageARN(c *conns.AWSClient, apiID, stageName string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   "apigateway",
		Region:    c.Region,
		Resource:  fmt.Sprintf("/apis/%s/stages/%s", apiID, stageName),
	}.String()
}

func stageInvokeARN(c *conns.AWSClient, apiID, stageName string) string {
	return arn.ARN{
		Partition: c.Partition,
		Service:   "execute-api",
		Region:    c.Region,
		AccountID: c.AccountID,
		Resource:  fmt.Sprintf("%s/%s", apiID, stageName),
	}.String()
}
