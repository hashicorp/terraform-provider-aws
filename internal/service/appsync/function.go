package appsync

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceFunctionCreate,
		Read:   resourceFunctionRead,
		Update: resourceFunctionUpdate,
		Delete: resourceFunctionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"data_source": {
				Type:     schema.TypeString,
				Required: true,
			},
			"max_batch_size": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 2000),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`), "must match [_A-Za-z][_0-9A-Za-z]*"),
			},
			"request_mapping_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"response_mapping_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"function_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2018-05-29",
				ValidateFunc: validation.StringInSlice([]string{
					"2018-05-29",
				}, true),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"sync_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"conflict_detection": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(appsync.ConflictDetectionType_Values(), false),
						},
						"conflict_handler": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(appsync.ConflictHandlerType_Values(), false),
						},
						"lambda_conflict_handler_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"lambda_conflict_handler_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
			"function_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID := d.Get("api_id").(string)

	input := &appsync.CreateFunctionInput{
		ApiId:                  aws.String(apiID),
		DataSourceName:         aws.String(d.Get("data_source").(string)),
		FunctionVersion:        aws.String(d.Get("function_version").(string)),
		Name:                   aws.String(d.Get("name").(string)),
		RequestMappingTemplate: aws.String(d.Get("request_mapping_template").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_mapping_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOkExists("max_batch_size"); ok {
		input.MaxBatchSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	resp, err := conn.CreateFunction(input)
	if err != nil {
		return fmt.Errorf("Error creating AppSync Function: %w", err)
	}

	d.SetId(fmt.Sprintf("%s-%s", apiID, aws.StringValue(resp.FunctionConfiguration.FunctionId)))

	return resourceFunctionRead(d, meta)
}

func resourceFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, functionID, err := DecodeFunctionID(d.Id())
	if err != nil {
		return err
	}

	input := &appsync.GetFunctionInput{
		ApiId:      aws.String(apiID),
		FunctionId: aws.String(functionID),
	}

	resp, err := conn.GetFunction(input)
	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) && !d.IsNewResource() {
		log.Printf("[WARN] AppSync Function (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting AppSync Function %s: %w", d.Id(), err)
	}

	function := resp.FunctionConfiguration
	d.Set("api_id", apiID)
	d.Set("function_id", functionID)
	d.Set("data_source", function.DataSourceName)
	d.Set("description", function.Description)
	d.Set("arn", function.FunctionArn)
	d.Set("function_version", function.FunctionVersion)
	d.Set("name", function.Name)
	d.Set("request_mapping_template", function.RequestMappingTemplate)
	d.Set("response_mapping_template", function.ResponseMappingTemplate)
	d.Set("max_batch_size", function.MaxBatchSize)

	if err := d.Set("sync_config", flattenSyncConfig(function.SyncConfig)); err != nil {
		return fmt.Errorf("error setting sync_config: %w", err)
	}

	return nil
}

func resourceFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, functionID, err := DecodeFunctionID(d.Id())
	if err != nil {
		return err
	}

	input := &appsync.UpdateFunctionInput{
		ApiId:                  aws.String(apiID),
		DataSourceName:         aws.String(d.Get("data_source").(string)),
		FunctionId:             aws.String(functionID),
		FunctionVersion:        aws.String(d.Get("function_version").(string)),
		Name:                   aws.String(d.Get("name").(string)),
		RequestMappingTemplate: aws.String(d.Get("request_mapping_template").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_mapping_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_batch_size"); ok {
		input.MaxBatchSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("sync_config"); ok && len(v.([]interface{})) > 0 {
		input.SyncConfig = expandSyncConfig(v.([]interface{}))
	}

	_, err = conn.UpdateFunction(input)
	if err != nil {
		return fmt.Errorf("Error updating AppSync Function %s: %w", d.Id(), err)
	}

	return resourceFunctionRead(d, meta)
}

func resourceFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, functionID, err := DecodeFunctionID(d.Id())
	if err != nil {
		return err
	}

	input := &appsync.DeleteFunctionInput{
		ApiId:      aws.String(apiID),
		FunctionId: aws.String(functionID),
	}

	_, err = conn.DeleteFunction(input)
	if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting AppSync Function %s: %w", d.Id(), err)
	}

	return nil
}

func DecodeFunctionID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "-", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format ApiID-FunctionID, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}

func expandSyncConfig(l []interface{}) *appsync.SyncConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.SyncConfig{}

	if v, ok := configured["conflict_detection"].(string); ok {
		result.ConflictDetection = aws.String(v)
	}

	if v, ok := configured["conflict_handler"].(string); ok {
		result.ConflictHandler = aws.String(v)
	}

	if v, ok := configured["lambda_conflict_handler_config"].([]interface{}); ok && len(v) > 0 {
		result.LambdaConflictHandlerConfig = expandLambdaConflictHandlerConfig(v)
	}

	return result
}

func flattenSyncConfig(config *appsync.SyncConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"conflict_detection":             aws.StringValue(config.ConflictDetection),
		"conflict_handler":               aws.StringValue(config.ConflictHandler),
		"lambda_conflict_handler_config": flattenLambdaConflictHandlerConfig(config.LambdaConflictHandlerConfig),
	}

	return []map[string]interface{}{result}
}

func expandLambdaConflictHandlerConfig(l []interface{}) *appsync.LambdaConflictHandlerConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	configured := l[0].(map[string]interface{})

	result := &appsync.LambdaConflictHandlerConfig{}

	if v, ok := configured["lambda_conflict_handler_arn"].(string); ok {
		result.LambdaConflictHandlerArn = aws.String(v)
	}

	return result
}

func flattenLambdaConflictHandlerConfig(config *appsync.LambdaConflictHandlerConfig) []map[string]interface{} {
	if config == nil {
		return nil
	}

	result := map[string]interface{}{
		"lambda_conflict_handler_arn": aws.StringValue(config.LambdaConflictHandlerArn),
	}

	return []map[string]interface{}{result}
}
