// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mediaconvert

import (
	"context"
	"log"
	_ "unsafe" // Required for go:linkname

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert"
	"github.com/aws/aws-sdk-go-v2/service/mediaconvert/types"
	smithyjson "github.com/aws/smithy-go/encoding/json"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Dirty hack to avoid any backwards compatibility issues with the AWS SDK for Go v2 migration.
// Reach down into the SDK and use the same serialization function that the SDK uses.
//
//go:linkname serializeJobTemplateSettings github.com/aws/aws-sdk-go-v2/service/mediaconvert.awsRestjson1_serializeDocumentJobTemplateSettings
func serializeJobTemplateSettings(v *types.JobTemplateSettings, value smithyjson.Value) error

// @SDKResource("aws_media_convert_job_template", name="Job Template")
// @Tags(identifierAttribute="arn")
func resourceJobTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceJobTemplateCreate,
		ReadWithoutTimeout:   resourceJobTemplateRead,
		UpdateWithoutTimeout: resourceJobTemplateUpdate,
		DeleteWithoutTimeout: resourceJobTemplateDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"acceleration_settings": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"mode": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.AccelerationMode](),
						},
					},
				},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"category": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hop_destinations": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"priority": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"queue": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"wait_minutes": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"priority": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(-50, 50),
			},
			"queue": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"settings_json": {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentJSONDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"status_update_interval": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[types.StatusUpdateInterval](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceJobTemplateCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	name := d.Get(names.AttrName).(string)
	settings, err := expandJobTemplateSettings(d.Get("settings_json").(string))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "expanding settings_json for Media Convert Job Template (%s): %s", name, err)
	}

	input := &mediaconvert.CreateJobTemplateInput{
		Name:     aws.String(name),
		Settings: settings,
		Tags:     getTagsIn(ctx),
	}

	if v, ok := d.GetOk("acceleration_settings"); ok && len(v.([]any)) > 0 {
		input.AccelerationSettings = expandJobTemplateAccelerationSettings(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("category"); ok {
		input.Category = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("hop_destinations"); ok {
		input.HopDestinations = expandHopDestinations(v.([]any))
	}

	input.Priority = aws.Int32(int32(d.Get("priority").(int)))

	if v, ok := d.GetOk("queue"); ok {
		input.Queue = aws.String(v.(string))
	}

	if v, ok := d.GetOk("status_update_interval"); ok {
		input.StatusUpdateInterval = types.StatusUpdateInterval(v.(string))
	}

	output, err := conn.CreateJobTemplate(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Media Convert Job Template (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.JobTemplate.Name))

	return append(diags, resourceJobTemplateRead(ctx, d, meta)...)
}

func resourceJobTemplateRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	jobTemplate, err := findJobTemplateByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Media Convert Job Template (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Media Convert Job Template (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, jobTemplate.Arn)
	d.Set("category", jobTemplate.Category)
	d.Set(names.AttrDescription, jobTemplate.Description)
	d.Set(names.AttrName, jobTemplate.Name)
	d.Set("priority", jobTemplate.Priority)
	d.Set("queue", jobTemplate.Queue)
	d.Set("status_update_interval", jobTemplate.StatusUpdateInterval)

	if jobTemplate.AccelerationSettings != nil {
		if err := d.Set("acceleration_settings", []any{flattenJobTemplateAccelerationSettings(jobTemplate.AccelerationSettings)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting acceleration_settings: %s", err)
		}
	} else {
		d.Set("acceleration_settings", nil)
	}

	if err := d.Set("hop_destinations", flattenHopDestinations(jobTemplate.HopDestinations)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting hop_destinations: %s", err)
	}

	if jobTemplate.Settings != nil {
		settingsJSON, err := flattenJobTemplateSettings(jobTemplate.Settings)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "flattening settings_json for Media Convert Job Template (%s): %s", d.Id(), err)
		}
		if err := d.Set("settings_json", settingsJSON); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting settings_json: %s", err)
		}
	}

	return diags
}

func resourceJobTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		settings, err := expandJobTemplateSettings(d.Get("settings_json").(string))
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "expanding settings_json for Media Convert Job Template (%s): %s", d.Id(), err)
		}

		input := &mediaconvert.UpdateJobTemplateInput{
			Name:     aws.String(d.Id()),
			Settings: settings,
		}

		if v, ok := d.GetOk("acceleration_settings"); ok && len(v.([]any)) > 0 {
			input.AccelerationSettings = expandJobTemplateAccelerationSettings(v.([]any)[0].(map[string]any))
		}

		if v, ok := d.GetOk("category"); ok {
			input.Category = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("hop_destinations"); ok {
			input.HopDestinations = expandHopDestinations(v.([]any))
		}

		input.Priority = aws.Int32(int32(d.Get("priority").(int)))

		if v, ok := d.GetOk("queue"); ok {
			input.Queue = aws.String(v.(string))
		}

		if v, ok := d.GetOk("status_update_interval"); ok {
			input.StatusUpdateInterval = types.StatusUpdateInterval(v.(string))
		}

		_, err = conn.UpdateJobTemplate(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Media Convert Job Template (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceJobTemplateRead(ctx, d, meta)...)
}

func resourceJobTemplateDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MediaConvertClient(ctx)

	log.Printf("[DEBUG] Deleting Media Convert Job Template: %s", d.Id())
	_, err := conn.DeleteJobTemplate(ctx, &mediaconvert.DeleteJobTemplateInput{
		Name: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Media Convert Job Template (%s): %s", d.Id(), err)
	}

	return diags
}

func findJobTemplateByName(ctx context.Context, conn *mediaconvert.Client, name string) (*types.JobTemplate, error) {
	input := &mediaconvert.GetJobTemplateInput{
		Name: aws.String(name),
	}

	output, err := conn.GetJobTemplate(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.JobTemplate == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.JobTemplate, nil
}

func expandJobTemplateSettings(s string) (*types.JobTemplateSettings, error) {
	var settings types.JobTemplateSettings
	if err := tfjson.DecodeFromString(s, &settings); err != nil {
		return nil, err
	}
	return &settings, nil
}

func flattenJobTemplateSettings(settings *types.JobTemplateSettings) (string, error) {
	jsonEncoder := smithyjson.NewEncoder()
	if err := serializeJobTemplateSettings(settings, jsonEncoder.Value); err != nil {
		return "", err
	}
	return jsonEncoder.String(), nil
}

func expandJobTemplateAccelerationSettings(tfMap map[string]any) *types.AccelerationSettings {
	if tfMap == nil {
		return nil
	}
	return &types.AccelerationSettings{
		Mode: types.AccelerationMode(tfMap["mode"].(string)),
	}
}

func flattenJobTemplateAccelerationSettings(apiObject *types.AccelerationSettings) map[string]any {
	if apiObject == nil {
		return nil
	}
	return map[string]any{
		"mode": apiObject.Mode,
	}
}

func expandHopDestinations(tfList []any) []types.HopDestination {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []types.HopDestination
	for _, item := range tfList {
		m := item.(map[string]any)
		apiObject := types.HopDestination{
			Priority:    aws.Int32(int32(m["priority"].(int))),
			WaitMinutes: aws.Int32(int32(m["wait_minutes"].(int))),
		}
		if v, ok := m["queue"].(string); ok && v != "" {
			apiObject.Queue = aws.String(v)
		}
		apiObjects = append(apiObjects, apiObject)
	}
	return apiObjects
}

func flattenHopDestinations(apiObjects []types.HopDestination) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []any
	for _, apiObject := range apiObjects {
		tfList = append(tfList, map[string]any{
			"priority":     aws.ToInt32(apiObject.Priority),
			"queue":        aws.ToString(apiObject.Queue),
			"wait_minutes": aws.ToInt32(apiObject.WaitMinutes),
		})
	}
	return tfList
}
