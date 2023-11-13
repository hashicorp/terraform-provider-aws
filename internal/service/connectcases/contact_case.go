// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connectcases

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connectcases"
	"github.com/aws/aws-sdk-go-v2/service/connectcases/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_connectcases_contact_case", name="Connect Cases Contact Case")
func ResourceContactCase() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceContactCaseCreate,
		ReadWithoutTimeout:   resourceContactCaseRead,
		UpdateWithoutTimeout: resourceContactCaseUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"fields": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Required: true,
						},
						//@bschaatsbergen, can we enforce it that only one of these 3 can be set?
						"bool_value": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"decimal_value": {
							Type:     schema.TypeFloat,
							Optional: true,
						},
						"string_value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"template_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"case_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"case_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceContactCaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	input := &connectcases.CreateCaseInput{
		TemplateId: aws.String(d.Get("template_id").(string)),
		DomainId:   aws.String(d.Get("domain_id").(string)),
	}

	if v, ok := d.GetOk("fields"); ok {
		input.Fields = expandFields(v.(*schema.Set).List())
	}

	output, err := conn.CreateCase(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Cases Contact Case: %s", err)
	}

	d.SetId(aws.ToString(output.CaseId))

	// The below fields are only returned by the Create Case API, so we need to set it here.
	d.Set("case_id", output.CaseId)
	d.Set("case_arn", output.CaseArn)

	return append(diags, resourceContactCaseRead(ctx, d, meta)...)
}

func resourceContactCaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	domainId := d.Get("domain_id").(string)
	output, err := FindContactCaseByDomainAndId(ctx, conn, d.Id(), domainId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Case Contact Case (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Case Contact Case (%s): %s", d.Id(), err)
	}

	d.Set("template_id", output.TemplateId)
	d.Set("fields", flattenFields(output.Fields))

	return diags
}

func resourceContactCaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	if d.HasChange("fields") {
		input := &connectcases.UpdateCaseInput{
			Fields: nil,
		}

		_, err := conn.UpdateCase(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Connect Cases Contact Case (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceContactCaseRead(ctx, d, meta)...)
}

func expandFields(fields []interface{}) []types.FieldValue {
	if len(fields) == 0 || fields[0] == nil {
		return nil
	}

	apiObject := make([]types.FieldValue, 0, len(fields))

	for _, object := range fields {
		field, ok := object.(map[string]interface{})
		if !ok {
			return nil
		}

		fieldValue := &types.FieldValue{}

		if v, ok := field["id"].(string); ok && v != "" {
			fieldValue.Id = aws.String(v)
		}

		if v, ok := field["bool_value"].(bool); ok {
			fieldValue.Value = &types.FieldValueUnionMemberBooleanValue{Value: v}
		}

		if v, ok := field["decimal_value"].(float64); ok {
			fieldValue.Value = &types.FieldValueUnionMemberDoubleValue{Value: v}
		}

		if v, ok := field["string_value"].(string); ok {
			fieldValue.Value = &types.FieldValueUnionMemberStringValue{Value: v}
		}

		apiObject = append(apiObject, *fieldValue)
	}

	return apiObject
}

func flattenFields(apiObject []types.FieldValue) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	var apiResult []interface{}
	for _, field := range apiObject {
		result := map[string]interface{}{
			"id":    aws.ToString(field.Id),
			"value": flattenFieldValue(field.Value),
		}

		apiResult = append(apiResult, result)
	}

	return []interface{}{apiResult}
}

func flattenFieldValue(apiObject types.FieldValueUnion) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	apiResult := map[string]interface{}{}

	switch v := apiObject.(type) {
	case *types.FieldValueUnionMemberBooleanValue:
		apiResult["bool_value"] = v.Value

	case *types.FieldValueUnionMemberDoubleValue:
		apiResult["double_value"] = v.Value

	case *types.FieldValueUnionMemberStringValue:
		apiResult["string_value"] = v.Value

	default:
		log.Println("union is nil or unknown type")
	}

	return []interface{}{apiResult}
}
