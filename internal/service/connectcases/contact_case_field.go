// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connectcases

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connectcases"
	"github.com/aws/aws-sdk-go-v2/service/connectcases/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_connectcases_field", name="Connect Cases Field")
func ResourceField() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFieldCreate,
		ReadWithoutTimeout:   resourceFieldRead,
		UpdateWithoutTimeout: resourceFieldUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"namespace": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(fieldType_Values(), false),
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"field_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"field_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFieldCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	input := &connectcases.CreateFieldInput{
		DomainId: aws.String(d.Get("domain_id").(string)),
		Name:     aws.String(d.Get("name").(string)),
		Type:     d.Get("type").(types.FieldType),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateField(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Cases Field: %s", err)
	}

	d.SetId(aws.ToString(output.FieldId))

	return append(diags, resourceFieldRead(ctx, d, meta)...)
}

func resourceFieldRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	output, err := FindFieldByDomainAndID(ctx, conn, d.Get("domain_id").(string), d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Cases Field (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Cases Field (%s): not found", d.Id())
	}

	d.Set("name", output.Name)
	d.Set("namespace", output.Namespace)
	d.Set("type", output.Type)
	d.Set("field_arn", output.FieldArn)
	d.Set("field_id", output.FieldId)

	return diags
}

func resourceFieldUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	input := &connectcases.UpdateFieldInput{
		FieldId: aws.String(d.Id()),
	}

	if d.HasChange("name") {
		input.Name = aws.String(d.Get("name").(string))
	}

	if d.HasChange("description") {
		input.Description = aws.String(d.Get("description").(string))
	}

	if d.HasChange("domain_id") {
		input.DomainId = aws.String(d.Get("domain_id").(string))
	}

	_, err := conn.UpdateField(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Connect Cases Field (%s): %s", d.Id(), err)
	}

	return append(diags, resourceFieldRead(ctx, d, meta)...)
}
