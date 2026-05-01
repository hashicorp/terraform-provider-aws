// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package b2bi

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/b2bi"
	awstypes "github.com/aws/aws-sdk-go-v2/service/b2bi/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_b2bi_transformer", name="Transformer")
// @Tags(identifierAttribute="transformer_arn")
func resourceTransformer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTransformerCreate,
		ReadWithoutTimeout:   resourceTransformerRead,
		UpdateWithoutTimeout: resourceTransformerUpdate,
		DeleteWithoutTimeout: resourceTransformerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"input_conversion": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_format": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.FromFormat](),
						},
						"format_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"x12": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"transaction_set": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.X12TransactionSet](),
												},
												"version": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.X12Version](),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"mapping": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"template": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 350000),
						},
						"template_language": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.MappingTemplateLanguage](),
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 254),
			},
			"output_conversion": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"to_format": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ToFormat](),
						},
						"format_options": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"x12": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"transaction_set": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.X12TransactionSet](),
												},
												"version": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[awstypes.X12Version](),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"sample_documents": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"keys": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"output": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrStatus: {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.TransformerStatus](),
			},
			"transformer_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"transformer_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceTransformerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &b2bi.CreateTransformerInput{
		Name: aws.String(name),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("input_conversion"); ok && len(v.([]any)) > 0 {
		input.InputConversion = expandInputConversion(v.([]any))
	}

	if v, ok := d.GetOk("mapping"); ok && len(v.([]any)) > 0 {
		input.Mapping = expandMapping(v.([]any))
	}

	if v, ok := d.GetOk("output_conversion"); ok && len(v.([]any)) > 0 {
		input.OutputConversion = expandOutputConversion(v.([]any))
	}

	if v, ok := d.GetOk("sample_documents"); ok && len(v.([]any)) > 0 {
		input.SampleDocuments = expandSampleDocuments(v.([]any))
	}

	output, err := conn.CreateTransformer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating B2BI Transformer (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.TransformerId))

	// Transformers are created as inactive. If the user wants it active, update immediately.
	if v, ok := d.GetOk(names.AttrStatus); ok && awstypes.TransformerStatus(v.(string)) == awstypes.TransformerStatusActive {
		_, err := conn.UpdateTransformer(ctx, &b2bi.UpdateTransformerInput{
			TransformerId: output.TransformerId,
			Status:        awstypes.TransformerStatusActive,
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "activating B2BI Transformer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTransformerRead(ctx, d, meta)...)
}

func resourceTransformerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	output, err := findTransformerByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] B2BI Transformer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading B2BI Transformer (%s): %s", d.Id(), err)
	}

	if err := d.Set("input_conversion", flattenInputConversion(output.InputConversion)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting input_conversion: %s", err)
	}
	if err := d.Set("mapping", flattenMapping(output.Mapping)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mapping: %s", err)
	}
	d.Set(names.AttrName, output.Name)
	if err := d.Set("output_conversion", flattenOutputConversion(output.OutputConversion)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting output_conversion: %s", err)
	}
	if err := d.Set("sample_documents", flattenSampleDocuments(output.SampleDocuments)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sample_documents: %s", err)
	}
	d.Set(names.AttrStatus, output.Status)
	d.Set("transformer_arn", output.TransformerArn)
	d.Set("transformer_id", output.TransformerId)

	return diags
}

func resourceTransformerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &b2bi.UpdateTransformerInput{
			TransformerId: aws.String(d.Id()),
		}

		if d.HasChange("input_conversion") {
			input.InputConversion = expandInputConversion(d.Get("input_conversion").([]any))
		}

		if d.HasChange("mapping") {
			input.Mapping = expandMapping(d.Get("mapping").([]any))
		}

		if d.HasChange("output_conversion") {
			input.OutputConversion = expandOutputConversion(d.Get("output_conversion").([]any))
		}

		if d.HasChange("sample_documents") {
			input.SampleDocuments = expandSampleDocuments(d.Get("sample_documents").([]any))
		}

		if d.HasChange(names.AttrStatus) {
			input.Status = awstypes.TransformerStatus(d.Get(names.AttrStatus).(string))
		}

		_, err := conn.UpdateTransformer(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating B2BI Transformer (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTransformerRead(ctx, d, meta)...)
}

func resourceTransformerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	log.Printf("[DEBUG] Deleting B2BI Transformer: %s", d.Id())
	_, err := conn.DeleteTransformer(ctx, &b2bi.DeleteTransformerInput{
		TransformerId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting B2BI Transformer (%s): %s", d.Id(), err)
	}

	return diags
}

func findTransformerByID(ctx context.Context, conn *b2bi.Client, id string) (*b2bi.GetTransformerOutput, error) {
	input := &b2bi.GetTransformerInput{
		TransformerId: aws.String(id),
	}

	output, err := conn.GetTransformer(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

// Expand/Flatten helpers

func expandInputConversion(l []any) *awstypes.InputConversion {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]any)
	result := &awstypes.InputConversion{
		FromFormat: awstypes.FromFormat(data["from_format"].(string)),
	}

	if v, ok := data["format_options"].([]any); ok && len(v) > 0 {
		result.FormatOptions = expandFormatOptions(v)
	}

	return result
}

func expandOutputConversion(l []any) *awstypes.OutputConversion {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]any)
	result := &awstypes.OutputConversion{
		ToFormat: awstypes.ToFormat(data["to_format"].(string)),
	}

	if v, ok := data["format_options"].([]any); ok && len(v) > 0 {
		result.FormatOptions = expandFormatOptions(v)
	}

	return result
}

func expandFormatOptions(l []any) awstypes.FormatOptions {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]any)

	if v, ok := data["x12"].([]any); ok && len(v) > 0 && v[0] != nil {
		x12Data := v[0].(map[string]any)
		x12Details := &awstypes.X12Details{}

		if ts, ok := x12Data["transaction_set"].(string); ok && ts != "" {
			x12Details.TransactionSet = awstypes.X12TransactionSet(ts)
		}

		if ver, ok := x12Data["version"].(string); ok && ver != "" {
			x12Details.Version = awstypes.X12Version(ver)
		}

		return &awstypes.FormatOptionsMemberX12{
			Value: *x12Details,
		}
	}

	return nil
}

func expandMapping(l []any) *awstypes.Mapping {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]any)
	result := &awstypes.Mapping{
		TemplateLanguage: awstypes.MappingTemplateLanguage(data["template_language"].(string)),
	}

	if v, ok := data["template"].(string); ok && v != "" {
		result.Template = aws.String(v)
	}

	return result
}

func expandSampleDocuments(l []any) *awstypes.SampleDocuments {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]any)
	result := &awstypes.SampleDocuments{
		BucketName: aws.String(data["bucket_name"].(string)),
	}

	if v, ok := data["keys"].([]any); ok && len(v) > 0 {
		keys := make([]awstypes.SampleDocumentKeys, 0, len(v))
		for _, item := range v {
			if item == nil {
				continue
			}
			keyData := item.(map[string]any)
			key := awstypes.SampleDocumentKeys{}
			if input, ok := keyData["input"].(string); ok && input != "" {
				key.Input = aws.String(input)
			}
			if output, ok := keyData["output"].(string); ok && output != "" {
				key.Output = aws.String(output)
			}
			keys = append(keys, key)
		}
		result.Keys = keys
	}

	return result
}

func flattenInputConversion(ic *awstypes.InputConversion) []map[string]any {
	if ic == nil {
		return nil
	}

	m := map[string]any{
		"from_format":    string(ic.FromFormat),
		"format_options": flattenFormatOptions(ic.FormatOptions),
	}

	return []map[string]any{m}
}

func flattenOutputConversion(oc *awstypes.OutputConversion) []map[string]any {
	if oc == nil {
		return nil
	}

	m := map[string]any{
		"to_format":      string(oc.ToFormat),
		"format_options": flattenFormatOptions(oc.FormatOptions),
	}

	return []map[string]any{m}
}

func flattenFormatOptions(fo awstypes.FormatOptions) []map[string]any {
	if fo == nil {
		return nil
	}

	switch v := fo.(type) {
	case *awstypes.FormatOptionsMemberX12:
		return []map[string]any{
			{
				"x12": []map[string]any{
					{
						"transaction_set": string(v.Value.TransactionSet),
						"version":         string(v.Value.Version),
					},
				},
			},
		}
	}

	return nil
}

func flattenMapping(m *awstypes.Mapping) []map[string]any {
	if m == nil {
		return nil
	}

	result := map[string]any{
		"template_language": string(m.TemplateLanguage),
	}

	if m.Template != nil {
		result["template"] = aws.ToString(m.Template)
	}

	return []map[string]any{result}
}

func flattenSampleDocuments(sd *awstypes.SampleDocuments) []map[string]any {
	if sd == nil {
		return nil
	}

	m := map[string]any{
		"bucket_name": aws.ToString(sd.BucketName),
	}

	if len(sd.Keys) > 0 {
		keys := make([]map[string]any, 0, len(sd.Keys))
		for _, k := range sd.Keys {
			key := map[string]any{}
			if k.Input != nil {
				key["input"] = aws.ToString(k.Input)
			}
			if k.Output != nil {
				key["output"] = aws.ToString(k.Output)
			}
			keys = append(keys, key)
		}
		m["keys"] = keys
	}

	return []map[string]any{m}
}
