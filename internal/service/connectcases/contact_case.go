// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connectcases

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/connectcases"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
			"client_token": {
				Type:     schema.TypeString,
				Optional: true,
			},
			//TODO: Implement fields
			// "fields": {
			// 	Type:     schema.TypeSet,
			// 	Required: true,
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"id": {
			// 				Type:     schema.TypeString,
			// 				Required: true,
			// 			},
			// 			"value": {
			// 				Type:     schema.TypeString,
			// 				Required: true,
			// 			},
			// 		},
			// 	},
			// },
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceContactCaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	input := &connectcases.CreateCaseInput{
		TemplateId: aws.String(d.Get("template_id").(string)),
		DomainId:   aws.String(d.Get("domain_id").(string)),
	}

	// if v, ok := d.GetOk("fields"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
	// 	input.Fields = expandFields(v.([]interface{})[0].(map[string]interface{}))
	// }

	if v, ok := d.GetOk("client_token"); ok {
		input.ClientToken = aws.String(v.(string))
	}

	output, err := conn.CreateCase(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Cases Contact Case: %s", err)
	}

	d.SetId(aws.ToString(output.CaseId))

	// The below fields are only returned by the Create API, so we need to set it here.
	d.Set("case_id", output.CaseId)
	d.Set("case_arn", output.CaseArn)

	return append(diags, resourceContactCaseRead(ctx, d, meta)...)
}

func resourceContactCaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	input := &connectcases.GetCaseInput{
		CaseId:   aws.String(d.Id()),
		DomainId: aws.String(d.Get("domain_id").(string)),
	}

	output, err := conn.GetCase(ctx, input)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Connect Cases Contact Case (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Cases Contact Case (%s): %s", d.Id(), err)
	}

	d.Set("template_id", output.TemplateId)

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
