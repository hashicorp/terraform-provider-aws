// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codepipeline

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/codepipeline"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codepipeline_custom_action_type", name="Custom Action Type")
// @Tags(identifierAttribute="arn")
func ResourceCustomActionType() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomActionTypeCreate,
		ReadWithoutTimeout:   resourceCustomActionTypeRead,
		UpdateWithoutTimeout: resourceCustomActionTypeUpdate,
		DeleteWithoutTimeout: resourceCustomActionTypeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringInSlice(codepipeline.ActionCategory_Values(), false),
			},
			"configuration_property": {
				Type:     schema.TypeList,
				MaxItems: 10,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"key": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"queryable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"required": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"secret": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(codepipeline.ActionConfigurationPropertyType_Values(), false),
						},
					},
				},
			},
			"input_artifact_details": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
						"minimum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
					},
				},
			},
			"output_artifact_details": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"maximum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
						"minimum_count": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(0, 5),
						},
					},
				},
			},
			"owner": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"provider_name": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 35),
			},
			"settings": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"entity_url_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"execution_url_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"revision_url_template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"third_party_configuration_url": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"version": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 9),
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCustomActionTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	category := d.Get("category").(string)
	provider := d.Get("provider_name").(string)
	version := d.Get("version").(string)
	id := CustomActionTypeCreateResourceID(category, provider, version)
	input := &codepipeline.CreateCustomActionTypeInput{
		Category: aws.String(category),
		Provider: aws.String(provider),
		Tags:     getTagsIn(ctx),
		Version:  aws.String(version),
	}

	if v, ok := d.GetOk("configuration_property"); ok && len(v.([]interface{})) > 0 {
		input.ConfigurationProperties = expandActionConfigurationProperties(v.([]interface{}))
	}

	if v, ok := d.GetOk("input_artifact_details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.InputArtifactDetails = expandArtifactDetails(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("output_artifact_details"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.OutputArtifactDetails = expandArtifactDetails(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("settings"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.Settings = expandActionTypeSettings(v.([]interface{})[0].(map[string]interface{}))
	}

	_, err := conn.CreateCustomActionTypeWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating CodePipeline Custom Action Type (%s): %s", id, err)
	}

	d.SetId(id)

	return resourceCustomActionTypeRead(ctx, d, meta)
}

func resourceCustomActionTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	category, provider, version, err := CustomActionTypeParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	actionType, err := FindCustomActionTypeByThreePartKey(ctx, conn, category, provider, version)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] CodePipeline Custom Action Type %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading CodePipeline Custom Action Type (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   codepipeline.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("actiontype:%s/%s/%s/%s", codepipeline.ActionOwnerCustom, category, provider, version),
	}.String()
	d.Set("arn", arn)
	d.Set("category", actionType.Id.Category)
	if err := d.Set("configuration_property", flattenActionConfigurationProperties(d, actionType.ActionConfigurationProperties)); err != nil {
		return diag.Errorf("setting configuration_property: %s", err)
	}
	if actionType.InputArtifactDetails != nil {
		if err := d.Set("input_artifact_details", []interface{}{flattenArtifactDetails(actionType.InputArtifactDetails)}); err != nil {
			return diag.Errorf("setting input_artifact_details: %s", err)
		}
	} else {
		d.Set("input_artifact_details", nil)
	}
	if actionType.OutputArtifactDetails != nil {
		if err := d.Set("output_artifact_details", []interface{}{flattenArtifactDetails(actionType.OutputArtifactDetails)}); err != nil {
			return diag.Errorf("setting output_artifact_details: %s", err)
		}
	} else {
		d.Set("output_artifact_details", nil)
	}
	d.Set("owner", actionType.Id.Owner)
	d.Set("provider_name", actionType.Id.Provider)
	if actionType.Settings != nil &&
		// Service can return empty ({}) Settings.
		(actionType.Settings.EntityUrlTemplate != nil || actionType.Settings.ExecutionUrlTemplate != nil || actionType.Settings.RevisionUrlTemplate != nil || actionType.Settings.ThirdPartyConfigurationUrl != nil) {
		if err := d.Set("settings", []interface{}{flattenActionTypeSettings(actionType.Settings)}); err != nil {
			return diag.Errorf("setting settings: %s", err)
		}
	} else {
		d.Set("settings", nil)
	}
	d.Set("version", actionType.Id.Version)

	return nil
}

func resourceCustomActionTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceCustomActionTypeRead(ctx, d, meta)
}

func resourceCustomActionTypeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).CodePipelineConn(ctx)

	category, provider, version, err := CustomActionTypeParseResourceID(d.Id())

	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Deleting CodePipeline Custom Action Type: %s", d.Id())
	_, err = conn.DeleteCustomActionTypeWithContext(ctx, &codepipeline.DeleteCustomActionTypeInput{
		Category: aws.String(category),
		Provider: aws.String(provider),
		Version:  aws.String(version),
	})

	if tfawserr.ErrCodeEquals(err, codepipeline.ErrCodeActionTypeNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting CodePipeline Custom Action Type (%s): %s", d.Id(), err)
	}

	return nil
}

const customActionTypeResourceIDSeparator = "/"

func CustomActionTypeCreateResourceID(category, provider, version string) string {
	parts := []string{category, provider, version}
	id := strings.Join(parts, customActionTypeResourceIDSeparator)

	return id
}

func CustomActionTypeParseResourceID(id string) (string, string, string, error) {
	parts := strings.Split(id, customActionTypeResourceIDSeparator)

	if len(parts) == 3 && parts[0] != "" && parts[1] != "" && parts[2] != "" {
		return parts[0], parts[1], parts[2], nil
	}

	return "", "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected category%[2]sprovider%[2]sversion", id, customActionTypeResourceIDSeparator)
}

func FindCustomActionTypeByThreePartKey(ctx context.Context, conn *codepipeline.CodePipeline, category, provider, version string) (*codepipeline.ActionType, error) {
	input := &codepipeline.ListActionTypesInput{
		ActionOwnerFilter: aws.String(codepipeline.ActionOwnerCustom),
	}
	var output *codepipeline.ActionType

	err := conn.ListActionTypesPagesWithContext(ctx, input, func(page *codepipeline.ListActionTypesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ActionTypes {
			if v == nil || v.Id == nil {
				continue
			}

			if aws.StringValue(v.Id.Category) == category && aws.StringValue(v.Id.Provider) == provider && aws.StringValue(v.Id.Version) == version {
				output = v

				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{
			LastRequest: input,
		}
	}

	return output, nil
}

func expandActionConfigurationProperty(tfMap map[string]interface{}) *codepipeline.ActionConfigurationProperty {
	if tfMap == nil {
		return nil
	}

	apiObject := &codepipeline.ActionConfigurationProperty{}

	if v, ok := tfMap["description"].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["key"].(bool); ok {
		apiObject.Key = aws.Bool(v)
	}

	if v, ok := tfMap["name"].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap["queryable"].(bool); ok && v {
		apiObject.Queryable = aws.Bool(v)
	}

	if v, ok := tfMap["required"].(bool); ok {
		apiObject.Required = aws.Bool(v)
	}

	if v, ok := tfMap["secret"].(bool); ok {
		apiObject.Secret = aws.Bool(v)
	}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		apiObject.Type = aws.String(v)
	}

	return apiObject
}

func expandActionConfigurationProperties(tfList []interface{}) []*codepipeline.ActionConfigurationProperty {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*codepipeline.ActionConfigurationProperty

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandActionConfigurationProperty(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func expandArtifactDetails(tfMap map[string]interface{}) *codepipeline.ArtifactDetails {
	if tfMap == nil {
		return nil
	}

	apiObject := &codepipeline.ArtifactDetails{}

	if v, ok := tfMap["maximum_count"].(int); ok {
		apiObject.MaximumCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["minimum_count"].(int); ok {
		apiObject.MinimumCount = aws.Int64(int64(v))
	}

	return apiObject
}

func expandActionTypeSettings(tfMap map[string]interface{}) *codepipeline.ActionTypeSettings {
	if tfMap == nil {
		return nil
	}

	apiObject := &codepipeline.ActionTypeSettings{}

	if v, ok := tfMap["entity_url_template"].(string); ok && v != "" {
		apiObject.EntityUrlTemplate = aws.String(v)
	}

	if v, ok := tfMap["execution_url_template"].(string); ok && v != "" {
		apiObject.ExecutionUrlTemplate = aws.String(v)
	}

	if v, ok := tfMap["revision_url_template"].(string); ok && v != "" {
		apiObject.RevisionUrlTemplate = aws.String(v)
	}

	if v, ok := tfMap["third_party_configuration_url"].(string); ok && v != "" {
		apiObject.ThirdPartyConfigurationUrl = aws.String(v)
	}

	return apiObject
}

func flattenActionConfigurationProperty(d *schema.ResourceData, i int, apiObject *codepipeline.ActionConfigurationProperty) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Description; v != nil {
		tfMap["description"] = aws.StringValue(v)
	}

	if v := apiObject.Key; v != nil {
		tfMap["key"] = aws.BoolValue(v)
	}

	if v := apiObject.Name; v != nil {
		tfMap["name"] = aws.StringValue(v)
	}

	if v := apiObject.Queryable; v != nil {
		tfMap["queryable"] = aws.BoolValue(v)
	}

	if v := apiObject.Required; v != nil {
		tfMap["required"] = aws.BoolValue(v)
	}

	if v := apiObject.Secret; v != nil {
		tfMap["secret"] = aws.BoolValue(v)
	}

	if v := apiObject.Type; v != nil {
		tfMap["type"] = aws.StringValue(v)
	} else {
		// The AWS API does not return Type.
		key := fmt.Sprintf("configuration_property.%d.type", i)
		tfMap["type"] = d.Get(key).(string)
	}

	return tfMap
}

func flattenActionConfigurationProperties(d *schema.ResourceData, apiObjects []*codepipeline.ActionConfigurationProperty) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for i, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		tfList = append(tfList, flattenActionConfigurationProperty(d, i, apiObject))
	}

	return tfList
}

func flattenArtifactDetails(apiObject *codepipeline.ArtifactDetails) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.MaximumCount; v != nil {
		tfMap["maximum_count"] = aws.Int64Value(v)
	}

	if v := apiObject.MinimumCount; v != nil {
		tfMap["minimum_count"] = aws.Int64Value(v)
	}

	return tfMap
}

func flattenActionTypeSettings(apiObject *codepipeline.ActionTypeSettings) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.EntityUrlTemplate; v != nil {
		tfMap["entity_url_template"] = aws.StringValue(v)
	}

	if v := apiObject.ExecutionUrlTemplate; v != nil {
		tfMap["execution_url_template"] = aws.StringValue(v)
	}

	if v := apiObject.RevisionUrlTemplate; v != nil {
		tfMap["revision_url_template"] = aws.StringValue(v)
	}

	if v := apiObject.ThirdPartyConfigurationUrl; v != nil {
		tfMap["third_party_configuration_url"] = aws.StringValue(v)
	}

	return tfMap
}
