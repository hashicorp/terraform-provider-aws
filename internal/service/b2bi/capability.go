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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_b2bi_capability", name="Capability")
// @Tags(identifierAttribute="capability_arn")
func resourceCapability() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCapabilityCreate,
		ReadWithoutTimeout:   resourceCapabilityRead,
		UpdateWithoutTimeout: resourceCapabilityUpdate,
		DeleteWithoutTimeout: resourceCapabilityDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"capability_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capability_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"edi": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"input_location": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrKey: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"output_location": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:     schema.TypeString,
													Required: true,
												},
												names.AttrKey: {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"transformer_id": {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrType: {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"x12_details": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"transaction_set": {
																Type:     schema.TypeString,
																Optional: true,
															},
															"version": {
																Type:     schema.TypeString,
																Optional: true,
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
					},
				},
			},
			"instructions_documents": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrKey: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 254),
			},
			names.AttrType: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceCapabilityCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &b2bi.CreateCapabilityInput{
		Name:          aws.String(name),
		Type:          awstypes.CapabilityType(d.Get(names.AttrType).(string)),
		Configuration: expandCapabilityConfiguration(d.Get("configuration").([]any)),
		Tags:          getTagsIn(ctx),
	}

	if v, ok := d.GetOk("instructions_documents"); ok {
		input.InstructionsDocuments = expandS3Locations(v.([]any))
	}

	output, err := conn.CreateCapability(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating B2BI Capability (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.CapabilityId))

	return append(diags, resourceCapabilityRead(ctx, d, meta)...)
}

func resourceCapabilityRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	output, err := findCapabilityByID(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] B2BI Capability (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading B2BI Capability (%s): %s", d.Id(), err)
	}

	d.Set("capability_arn", output.CapabilityArn)
	d.Set("capability_id", output.CapabilityId)

	// The API does not return the edi.type field on read, so we merge the API response
	// with the existing state to preserve the type configuration.
	flatConfig := flattenCapabilityConfiguration(output.Configuration)
	if flatConfig != nil && len(flatConfig) > 0 {
		if existingConfig, ok := d.GetOk("configuration"); ok {
			existing := existingConfig.([]any)
			if len(existing) > 0 && existing[0] != nil {
				existingMap := existing[0].(map[string]any)
				if existingEdi, ok := existingMap["edi"].([]any); ok && len(existingEdi) > 0 && existingEdi[0] != nil {
					existingEdiMap := existingEdi[0].(map[string]any)
					if existingType, ok := existingEdiMap[names.AttrType]; ok {
						// Preserve the type from state if the API didn't return it
						if newEdi, ok := flatConfig[0]["edi"].([]map[string]any); ok && len(newEdi) > 0 {
							if _, hasType := newEdi[0][names.AttrType]; !hasType {
								newEdi[0][names.AttrType] = existingType
							}
						}
					}
				}
			}
		}
	}
	if err := d.Set("configuration", flatConfig); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration: %s", err)
	}
	if err := d.Set("instructions_documents", flattenS3Locations(output.InstructionsDocuments)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instructions_documents: %s", err)
	}
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrType, output.Type)

	return diags
}

func resourceCapabilityUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &b2bi.UpdateCapabilityInput{
			CapabilityId: aws.String(d.Id()),
		}

		if d.HasChange("configuration") {
			input.Configuration = expandCapabilityConfiguration(d.Get("configuration").([]any))
		}

		if d.HasChange("instructions_documents") {
			input.InstructionsDocuments = expandS3Locations(d.Get("instructions_documents").([]any))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		_, err := conn.UpdateCapability(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating B2BI Capability (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceCapabilityRead(ctx, d, meta)...)
}

func resourceCapabilityDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).B2BIClient(ctx)

	log.Printf("[DEBUG] Deleting B2BI Capability: %s", d.Id())
	_, err := conn.DeleteCapability(ctx, &b2bi.DeleteCapabilityInput{
		CapabilityId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting B2BI Capability (%s): %s", d.Id(), err)
	}

	return diags
}

func findCapabilityByID(ctx context.Context, conn *b2bi.Client, id string) (*b2bi.GetCapabilityOutput, error) {
	input := &b2bi.GetCapabilityInput{
		CapabilityId: aws.String(id),
	}

	output, err := conn.GetCapability(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{LastError: err}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

// Expand/Flatten helpers for Capability

func expandCapabilityConfiguration(l []any) awstypes.CapabilityConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	data := l[0].(map[string]any)

	if v, ok := data["edi"].([]any); ok && len(v) > 0 && v[0] != nil {
		ediData := v[0].(map[string]any)
		ediConfig := awstypes.EdiConfiguration{}

		if v, ok := ediData["input_location"].([]any); ok && len(v) > 0 && v[0] != nil {
			loc := v[0].(map[string]any)
			ediConfig.InputLocation = &awstypes.S3Location{
				BucketName: aws.String(loc["bucket_name"].(string)),
				Key:        aws.String(loc[names.AttrKey].(string)),
			}
		}

		if v, ok := ediData["output_location"].([]any); ok && len(v) > 0 && v[0] != nil {
			loc := v[0].(map[string]any)
			ediConfig.OutputLocation = &awstypes.S3Location{
				BucketName: aws.String(loc["bucket_name"].(string)),
				Key:        aws.String(loc[names.AttrKey].(string)),
			}
		}

		if v, ok := ediData["transformer_id"].(string); ok && v != "" {
			ediConfig.TransformerId = aws.String(v)
		}

		if v, ok := ediData[names.AttrType].([]any); ok && len(v) > 0 && v[0] != nil {
			typeData := v[0].(map[string]any)
			if x12, ok := typeData["x12_details"].([]any); ok && len(x12) > 0 && x12[0] != nil {
				x12Data := x12[0].(map[string]any)
				x12Details := awstypes.X12Details{}
				if ts, ok := x12Data["transaction_set"].(string); ok && ts != "" {
					x12Details.TransactionSet = awstypes.X12TransactionSet(ts)
				}
				if ver, ok := x12Data["version"].(string); ok && ver != "" {
					x12Details.Version = awstypes.X12Version(ver)
				}
				ediConfig.Type = &awstypes.EdiTypeMemberX12Details{Value: x12Details}
			}
		}

		return &awstypes.CapabilityConfigurationMemberEdi{Value: ediConfig}
	}

	return nil
}

func flattenCapabilityConfiguration(config awstypes.CapabilityConfiguration) []map[string]any {
	if config == nil {
		return nil
	}

	switch v := config.(type) {
	case *awstypes.CapabilityConfigurationMemberEdi:
		edi := v.Value
		ediMap := map[string]any{
			"transformer_id": aws.ToString(edi.TransformerId),
		}

		if edi.InputLocation != nil {
			ediMap["input_location"] = []map[string]any{
				{
					"bucket_name": aws.ToString(edi.InputLocation.BucketName),
					names.AttrKey: aws.ToString(edi.InputLocation.Key),
				},
			}
		}

		if edi.OutputLocation != nil {
			ediMap["output_location"] = []map[string]any{
				{
					"bucket_name": aws.ToString(edi.OutputLocation.BucketName),
					names.AttrKey: aws.ToString(edi.OutputLocation.Key),
				},
			}
		}

		if edi.Type != nil {
			switch t := edi.Type.(type) {
			case *awstypes.EdiTypeMemberX12Details:
				ediMap[names.AttrType] = []map[string]any{
					{
						"x12_details": []map[string]any{
							{
								"transaction_set": string(t.Value.TransactionSet),
								"version":         string(t.Value.Version),
							},
						},
					},
				}
			}
		}

		return []map[string]any{
			{
				"edi": []map[string]any{ediMap},
			},
		}
	}

	return nil
}

func expandS3Locations(l []any) []awstypes.S3Location {
	if len(l) == 0 {
		return nil
	}

	result := make([]awstypes.S3Location, 0, len(l))
	for _, item := range l {
		if item == nil {
			continue
		}
		data := item.(map[string]any)
		result = append(result, awstypes.S3Location{
			BucketName: aws.String(data["bucket_name"].(string)),
			Key:        aws.String(data[names.AttrKey].(string)),
		})
	}

	return result
}

func flattenS3Locations(locations []awstypes.S3Location) []map[string]any {
	if len(locations) == 0 {
		return nil
	}

	result := make([]map[string]any, 0, len(locations))
	for _, loc := range locations {
		result = append(result, map[string]any{
			"bucket_name": aws.ToString(loc.BucketName),
			names.AttrKey: aws.ToString(loc.Key),
		})
	}

	return result
}
