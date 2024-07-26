// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datapipeline

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datapipeline"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datapipeline/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datapipeline_pipeline_definition")
func ResourcePipelineDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePipelineDefinitionPut,
		ReadWithoutTimeout:   resourcePipelineDefinitionRead,
		UpdateWithoutTimeout: resourcePipelineDefinitionPut,
		DeleteWithoutTimeout: schema.NoopContext,
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
									names.AttrKey: {
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
						names.AttrID: {
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
						names.AttrID: {
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
						names.AttrField: {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
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
						names.AttrID: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(1, 1024),
						},
						names.AttrName: {
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
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DataPipelineClient(ctx)

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
	err = retry.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *retry.RetryError {
		output, err = conn.PutPipelineDefinition(ctx, input)
		if err != nil {
			if errs.IsA[*awstypes.InternalServiceError](err) {
				return retry.RetryableError(err)
			}

			return retry.NonRetryableError(err)
		}
		if output.Errored {
			errors := getValidationError(output.ValidationErrors)
			if strings.Contains(errors.Error(), names.AttrRole) {
				return retry.RetryableError(fmt.Errorf("validating after creation DataPipeline Pipeline Definition (%s): %w", pipelineID, errors))
			}
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.PutPipelineDefinition(ctx, input)
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataPipeline Pipeline Definition (%s): %s", pipelineID, err)
	}

	if output.Errored {
		return sdkdiag.AppendErrorf(diags, "validating after creation DataPipeline Pipeline Definition (%s): %s", pipelineID, getValidationError(output.ValidationErrors))
	}

	// Activate pipeline if enabled
	input2 := &datapipeline.ActivatePipelineInput{
		PipelineId: aws.String(pipelineID),
	}

	_, err = conn.ActivatePipeline(ctx, input2)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "activating DataPipeline Pipeline Definition (%s): %s", pipelineID, err)
	}

	d.SetId(pipelineID)

	return append(diags, resourcePipelineDefinitionRead(ctx, d, meta)...)
}

func resourcePipelineDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).DataPipelineClient(ctx)
	input := &datapipeline.GetPipelineDefinitionInput{
		PipelineId: aws.String(d.Id()),
	}

	resp, err := conn.GetPipelineDefinition(ctx, input)

	if !d.IsNewResource() && errs.IsA[*awstypes.PipelineNotFoundException](err) ||
		errs.IsA[*awstypes.PipelineDeletedException](err) {
		log.Printf("[WARN] DataPipeline Pipeline Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataPipeline Pipeline Definition (%s): %s", d.Id(), err)
	}

	if err = d.Set("parameter_object", flattenPipelineDefinitionParameterObjects(resp.ParameterObjects)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_object", d.Id(), err)
	}
	if err = d.Set("parameter_value", flattenPipelineDefinitionParameterValues(resp.ParameterValues)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for DataPipeline Pipeline Definition (%s): %s", "parameter_value", d.Id(), err)
	}
	if err = d.Set("pipeline_object", flattenPipelineDefinitionObjects(resp.PipelineObjects)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting `%s` for DataPipeline Pipeline Definition (%s): %s", "pipeline_object", d.Id(), err)
	}
	d.Set("pipeline_id", d.Id())

	return diags
}

func expandPipelineDefinitionParameterObject(tfMap map[string]interface{}) awstypes.ParameterObject {
	if tfMap == nil {
		return awstypes.ParameterObject{}
	}

	apiObject := awstypes.ParameterObject{
		Attributes: expandPipelineDefinitionParameterAttributes(tfMap["attribute"].(*schema.Set).List()),
		Id:         aws.String(tfMap[names.AttrID].(string)),
	}

	return apiObject
}

func expandPipelineDefinitionParameterAttribute(tfMap map[string]interface{}) awstypes.ParameterAttribute {
	if tfMap == nil {
		return awstypes.ParameterAttribute{}
	}

	apiObject := awstypes.ParameterAttribute{
		Key:         aws.String(tfMap[names.AttrKey].(string)),
		StringValue: aws.String(tfMap["string_value"].(string)),
	}

	return apiObject
}

func expandPipelineDefinitionParameterAttributes(tfList []interface{}) []awstypes.ParameterAttribute {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ParameterAttribute

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

func expandPipelineDefinitionParameterObjects(tfList []interface{}) []awstypes.ParameterObject {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ParameterObject

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

func flattenPipelineDefinitionParameterObject(apiObject awstypes.ParameterObject) map[string]interface{} {
	tfMap := map[string]interface{}{}
	tfMap["attribute"] = flattenPipelineDefinitionParameterAttributes(apiObject.Attributes)
	tfMap[names.AttrID] = aws.ToString(apiObject.Id)

	return tfMap
}

func flattenPipelineDefinitionParameterAttribute(apiObject awstypes.ParameterAttribute) map[string]interface{} {
	tfMap := map[string]interface{}{}
	tfMap[names.AttrKey] = aws.ToString(apiObject.Key)
	tfMap["string_value"] = aws.ToString(apiObject.StringValue)

	return tfMap
}

func flattenPipelineDefinitionParameterAttributes(apiObjects []awstypes.ParameterAttribute) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPipelineDefinitionParameterAttribute(apiObject))
	}

	return tfList
}

func flattenPipelineDefinitionParameterObjects(apiObjects []awstypes.ParameterObject) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPipelineDefinitionParameterObject(apiObject))
	}

	return tfList
}

func expandPipelineDefinitionParameterValue(tfMap map[string]interface{}) awstypes.ParameterValue {
	if tfMap == nil {
		return awstypes.ParameterValue{}
	}

	apiObject := awstypes.ParameterValue{
		Id:          aws.String(tfMap[names.AttrID].(string)),
		StringValue: aws.String(tfMap["string_value"].(string)),
	}

	return apiObject
}

func expandPipelineDefinitionParameterValues(tfList []interface{}) []awstypes.ParameterValue {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.ParameterValue

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

func flattenPipelineDefinitionParameterValue(apiObject awstypes.ParameterValue) map[string]interface{} {
	tfMap := map[string]interface{}{}
	tfMap[names.AttrID] = aws.ToString(apiObject.Id)
	tfMap["string_value"] = aws.ToString(apiObject.StringValue)

	return tfMap
}

func flattenPipelineDefinitionParameterValues(apiObjects []awstypes.ParameterValue) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPipelineDefinitionParameterValue(apiObject))
	}

	return tfList
}

func expandPipelineDefinitionObject(tfMap map[string]interface{}) awstypes.PipelineObject {
	if tfMap == nil {
		return awstypes.PipelineObject{}
	}

	apiObject := awstypes.PipelineObject{
		Fields: expandPipelineDefinitionPipelineFields(tfMap[names.AttrField].(*schema.Set).List()),
		Id:     aws.String(tfMap[names.AttrID].(string)),
		Name:   aws.String(tfMap[names.AttrName].(string)),
	}

	return apiObject
}

func expandPipelineDefinitionPipelineField(tfMap map[string]interface{}) awstypes.Field {
	if tfMap == nil {
		return awstypes.Field{}
	}

	apiObject := awstypes.Field{
		Key: aws.String(tfMap[names.AttrKey].(string)),
	}

	if v, ok := tfMap["ref_value"]; ok && v.(string) != "" {
		apiObject.RefValue = aws.String(v.(string))
	}
	if v, ok := tfMap["string_value"]; ok && v.(string) != "" {
		apiObject.StringValue = aws.String(v.(string))
	}

	return apiObject
}

func expandPipelineDefinitionPipelineFields(tfList []interface{}) []awstypes.Field {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.Field

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

func expandPipelineDefinitionObjects(tfList []interface{}) []awstypes.PipelineObject {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.PipelineObject

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

func flattenPipelineDefinitionObject(apiObject awstypes.PipelineObject) map[string]interface{} {
	tfMap := map[string]interface{}{}
	tfMap[names.AttrField] = flattenPipelineDefinitionParameterFields(apiObject.Fields)
	tfMap[names.AttrID] = aws.ToString(apiObject.Id)
	tfMap[names.AttrName] = aws.ToString(apiObject.Name)

	return tfMap
}

func flattenPipelineDefinitionParameterField(apiObject awstypes.Field) map[string]interface{} {
	tfMap := map[string]interface{}{}
	tfMap[names.AttrKey] = aws.ToString(apiObject.Key)
	tfMap["ref_value"] = aws.ToString(apiObject.RefValue)
	tfMap["string_value"] = aws.ToString(apiObject.StringValue)

	return tfMap
}

func flattenPipelineDefinitionParameterFields(apiObjects []awstypes.Field) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPipelineDefinitionParameterField(apiObject))
	}

	return tfList
}

func flattenPipelineDefinitionObjects(apiObjects []awstypes.PipelineObject) []map[string]interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []map[string]interface{}

	for _, apiObject := range apiObjects {
		tfList = append(tfList, flattenPipelineDefinitionObject(apiObject))
	}

	return tfList
}

func getValidationError(validationErrors []awstypes.ValidationError) error {
	var errs []error

	for _, err := range validationErrors {
		errs = append(errs, fmt.Errorf("id: %s, error: %v", aws.ToString(err.Id), err.Errors))
	}

	return errors.Join(errs...)
}
