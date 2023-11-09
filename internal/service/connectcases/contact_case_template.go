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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_connectcases_template", name="Connect Cases Template")
func ResourceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTemplateCreate,
		ReadWithoutTimeout:   resourceTemplateRead,
		UpdateWithoutTimeout: resourceTemplateUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"template_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"layout_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_layout": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"required_fields": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(flattenTemplateStatusValues(types.TemplateStatus("").Values()), false),
				Default:      types.TemplateStatusInactive,
			},
		},
	}
}

func resourceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)
	log.Print("[DEBUG] Creating Connect Case Template")

	name := d.Get("name").(string)
	params := &connectcases.CreateTemplateInput{
		Name:     aws.String(name),
		DomainId: aws.String(d.Get("domain_id").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		params.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("layout_configuration"); ok {
		params.LayoutConfiguration = expandLayoutConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("required_fields"); ok {
		params.RequiredFields = expandRequiredFields(v.([]interface{}))
	}

	if v, ok := d.GetOk("status"); ok {
		params.Status = types.TemplateStatus(v.(string))
	}

	output, err := conn.CreateTemplate(ctx, params)
	if err != nil {
		return diag.Errorf("creating Connect Case Template (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.TemplateId))
	d.Set("template_arn", aws.ToString(output.TemplateArn))

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func resourceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	domainId := d.Get("domain_id").(string)
	output, err := FindTemplateByDomainAndId(ctx, conn, d.Id(), domainId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Case Template %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Case Template (%s): %s", d.Id(), err)
	}

	d.Set("name", output.Name)
	d.Set("description", output.Description)
	d.Set("layout_configuration", flattenLayoutConfiguration(output.LayoutConfiguration))
	d.Set("required_fields", flattenRequiredFields(output.RequiredFields))
	d.Set("status", output.Status)

	return diags
}

func resourceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	input := &connectcases.UpdateTemplateInput{
		TemplateId: aws.String(d.Id()),
	}

	if d.HasChange("name") {
		input.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("domain_id") {
		input.DomainId = aws.String(d.Get("domain_id").(string))
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("layout_configuration") {
		input.LayoutConfiguration = expandLayoutConfiguration(d.Get("layout_configuration").([]interface{}))
	}

	if d.HasChange("required_fields") {
		input.RequiredFields = expandRequiredFields(d.Get("required_fields").([]interface{}))
	}

	if d.HasChange("status") {
		input.Status = types.TemplateStatus(d.Get("status").(string))
	}

	_, err := conn.UpdateTemplate(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Connect Cases Template (%s): %s", d.Id(), err)
	}

	return append(diags, resourceTemplateRead(ctx, d, meta)...)
}

func flattenLayoutConfiguration(apiObject *types.LayoutConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}
	if v := apiObject.DefaultLayout; v != nil {
		tfMap["default_layout"] = aws.ToString(v)
	}

	return []interface{}{tfMap}
}

func flattenRequiredFields(apiObject []types.RequiredField) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}
	for _, requiredField := range apiObject {

		tfMap := map[string]interface{}{
			"field_id": aws.ToString(requiredField.FieldId),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandLayoutConfiguration(tfMap []interface{}) *types.LayoutConfiguration {
	if tfMap == nil || tfMap[0] == nil {
		return nil
	}

	tfList, ok := tfMap[0].(map[string]interface{})
	if !ok {
		return nil
	}

	apiObject := &types.LayoutConfiguration{}
	apiObject.DefaultLayout = aws.String(tfList["default_layout"].(string))

	return apiObject
}

func expandRequiredFields(tfMap []interface{}) []types.RequiredField {
	if tfMap == nil || tfMap[0] == nil {
		return nil
	}

	tfList := make([]types.RequiredField, 0, len(tfMap))
	for _, object := range tfMap {
		if object == nil {
			continue
		}

		field := object.(map[string]interface{})
		requiredField := types.RequiredField{
			FieldId: aws.String(field["field_id"].(string)),
		}

		tfList = append(tfList, requiredField)
	}

	return tfList
}

func flattenTemplateStatusValues(t []types.TemplateStatus) []string {
	var out []string

	for _, v := range t {
		out = append(out, string(v))
	}

	return out
}
