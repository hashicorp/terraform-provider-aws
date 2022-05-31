package datapipeline

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datapipeline"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourcePipelineDefinition() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePipelineDefinitionPut,
		ReadContext:   resourcePipelineDefinitionRead,
		UpdateContext: resourcePipelineDefinitionPut,
		DeleteContext: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"parameter_object": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attribute": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"string_value": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 10240),
									},
								},
							},
						},
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
					},
				},
			},
			"parameter_value": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 256),
						},
						"string_value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 10240),
						},
					},
				},
			},
			"pipeline_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"pipeline_object": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"ref_value": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
									},
									"string_value": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 10240),
									},
								},
							},
						},
						"id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						"name": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
					},
				},
			},
		},
	}
}

func resourcePipelineDefinitionPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DataPipelineConn

	pipelineID := d.Get("pipeline_id").(string)
	input := &datapipeline.PutPipelineDefinitionInput{
		PipelineId:      aws.String(pipelineID),
		PipelineObjects: expandPipelineDefinitionObjects(d.Get("pipeline_object").(*schema.Set).List()),
	}

	if v, ok := d.GetOk("parameter_object"); ok {
		input.ParameterObjects = expandPipelineDefinitionParameterObjects(v.(*schema.Set).List())
	}

	if v, ok := d.GetOk("parameter_value"); ok {
		input.ParameterValues = expandPipelineDefinitionParameterValues(v.(*schema.Set).List())
	}

	var err error
	var output *datapipeline.PutPipelineDefinitionOutput
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		output, err = conn.PutPipelineDefinitionWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, datapipeline.ErrCodeInternalServiceError) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		if aws.BoolValue(output.Errored) {
			errors := getValidationError(output.ValidationErrors)
			if strings.Contains(errors.Error(), "role") {
				return resource.RetryableError(fmt.Errorf("error validating after creation DataPipeline Pipeline Definition (%s): %w", pipelineID, errors))
			}
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.PutPipelineDefinitionWithContext(ctx, input)
	}

	if err != nil {
		return diag.Errorf("error creating DataPipeline Pipeline Definition (%s): %s", pipelineID, err)
	}

	if aws.BoolValue(output.Errored) {
		return diag.Errorf("error validating after creation DataPipeline Pipeline Definition (%s): %s", pipelineID, getValidationError(output.ValidationErrors))
	}

	// Activate pipeline if enabled
	input2 := &datapipeline.ActivatePipelineInput{
		PipelineId: aws.String(pipelineID),
	}

	_, err = conn.ActivatePipelineWithContext(ctx, input2)
	if err != nil {
		return diag.Errorf("error activating DataPipeline Pipeline Definition (%s): %s", pipelineID, err)
	}

	d.SetId(pipelineID)

	return resourcePipelineDefinitionRead(ctx, d, meta)
}

func resourcePipelineDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DataPipelineConn
	input := &datapipeline.GetPipelineDefinitionInput{
		PipelineId: aws.String(d.Id()),
	}

	resp, err := conn.GetPipelineDefinitionWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineNotFoundException) ||
		tfawserr.ErrCodeEquals(err, datapipeline.ErrCodePipelineDeletedException) {
		log.Printf("[WARN] DataPipeline Pipeline Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error reading DataPipeline Pipeline Definition (%s): %s", d.Id(), err)
	}

	if err = d.Set("parameter_object", flattenPipelineDefinitionParameterObjects(resp.ParameterObjects)); err != nil {
		return diag.Errorf("error setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", d.Id(), err)
	}
	if err = d.Set("parameter_value", flattenPipelineDefinitionParameterValues(resp.ParameterValues)); err != nil {
		return diag.Errorf("error setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", d.Id(), err)
	}
	if err = d.Set("pipeline_object", flattenPipelineDefinitionObjects(resp.PipelineObjects)); err != nil {
		return diag.Errorf("error setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", d.Id(), err)
	}
	d.Set("pipeline_id", d.Id())

	return nil
}

func expandPipelineDefinitionParameterObject(tfMap map[string]interface{}) *datapipeline.ParameterObject {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.ParameterObject{
		Attributes: expandPipelineDefinitionParameterAttributes(tfMap["attribute"].(*schema.Set).List()),
		Id:         aws.String(tfMap["id"].(string)),
	}

	return apiObject
}

func expandPipelineDefinitionParameterAttribute(tfMap map[string]interface{}) *datapipeline.ParameterAttribute {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.ParameterAttribute{
		Key:         aws.String(tfMap["key"].(string)),
		StringValue: aws.String(tfMap["string_value"].(string)),
	}

	return apiObject
}

func expandPipelineDefinitionParameterAttributes(tfList []interface{}) []*datapipeline.ParameterAttribute {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.ParameterAttribute

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPipelineDefinitionParameterAttribute(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandPipelineDefinitionParameterObjects(tfList []interface{}) []*datapipeline.ParameterObject {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.ParameterObject

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPipelineDefinitionParameterObject(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenPipelineDefinitionParameterObject(apiObject *datapipeline.ParameterObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["attribute"] = flattenPipelineDefinitionParameterAttributes(apiObject.Attributes)
	tfMap["id"] = aws.StringValue(apiObject.Id)

	return tfMap
}

func flattenPipelineDefinitionParameterAttribute(apiObject *datapipeline.ParameterAttribute) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["string_value"] = aws.StringValue(apiObject.StringValue)

	return tfMap
}

func flattenPipelineDefinitionParameterAttributes(apiObjects []*datapipeline.ParameterAttribute) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenPipelineDefinitionParameterAttribute(apiObject))
	}

	return tfList
}

func flattenPipelineDefinitionParameterObjects(apiObjects []*datapipeline.ParameterObject) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenPipelineDefinitionParameterObject(apiObject))
	}

	return tfList
}

func expandPipelineDefinitionParameterValue(tfMap map[string]interface{}) *datapipeline.ParameterValue {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.ParameterValue{
		Id:          aws.String(tfMap["id"].(string)),
		StringValue: aws.String(tfMap["string_value"].(string)),
	}

	return apiObject
}

func expandPipelineDefinitionParameterValues(tfList []interface{}) []*datapipeline.ParameterValue {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.ParameterValue

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPipelineDefinitionParameterValue(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenPipelineDefinitionParameterValue(apiObject *datapipeline.ParameterValue) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["id"] = aws.StringValue(apiObject.Id)
	tfMap["string_value"] = aws.StringValue(apiObject.StringValue)

	return tfMap
}

func flattenPipelineDefinitionParameterValues(apiObjects []*datapipeline.ParameterValue) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenPipelineDefinitionParameterValue(apiObject))
	}

	return tfList
}

func expandPipelineDefinitionObject(tfMap map[string]interface{}) *datapipeline.PipelineObject {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.PipelineObject{
		Fields: expandPipelineDefinitionPipelineFields(tfMap["field"].(*schema.Set).List()),
		Id:     aws.String(tfMap["id"].(string)),
		Name:   aws.String(tfMap["name"].(string)),
	}

	return apiObject
}

func expandPipelineDefinitionPipelineField(tfMap map[string]interface{}) *datapipeline.Field {
	if tfMap == nil {
		return nil
	}

	apiObject := &datapipeline.Field{
		Key: aws.String(tfMap["key"].(string)),
	}

	if v, ok := tfMap["ref_value"]; ok && v.(string) != "" {
		apiObject.RefValue = aws.String(v.(string))
	}
	if v, ok := tfMap["string_value"]; ok && v.(string) != "" {
		apiObject.StringValue = aws.String(v.(string))
	}

	return apiObject
}

func expandPipelineDefinitionPipelineFields(tfList []interface{}) []*datapipeline.Field {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.Field

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPipelineDefinitionPipelineField(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandPipelineDefinitionObjects(tfList []interface{}) []*datapipeline.PipelineObject {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*datapipeline.PipelineObject

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandPipelineDefinitionObject(tfMap)

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenPipelineDefinitionObject(apiObject *datapipeline.PipelineObject) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["field"] = flattenPipelineDefinitionParameterFields(apiObject.Fields)
	tfMap["id"] = aws.StringValue(apiObject.Id)
	tfMap["name"] = aws.StringValue(apiObject.Name)

	return tfMap
}

func flattenPipelineDefinitionParameterField(apiObject *datapipeline.Field) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	tfMap["key"] = aws.StringValue(apiObject.Key)
	tfMap["ref_value"] = aws.StringValue(apiObject.RefValue)
	tfMap["string_value"] = aws.StringValue(apiObject.StringValue)

	return tfMap
}

func flattenPipelineDefinitionParameterFields(apiObjects []*datapipeline.Field) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenPipelineDefinitionParameterField(apiObject))
	}

	return tfList
}

func flattenPipelineDefinitionObjects(apiObjects []*datapipeline.PipelineObject) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenPipelineDefinitionObject(apiObject))
	}

	return tfList
}

func getValidationError(validationError []*datapipeline.ValidationError) error {
	var validationErrors error
	for _, error := range validationError {
		validationErrors = multierror.Append(validationErrors, fmt.Errorf("id: %s, error: %v", aws.StringValue(error.Id), aws.StringValueSlice(error.Errors)))
	}

	return validationErrors
}
