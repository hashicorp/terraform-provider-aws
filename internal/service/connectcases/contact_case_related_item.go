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
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_connectcases_related_item", name="Connect Cases Related Item")
func ResourceRelatedItem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRelatedItemCreate,
		ReadWithoutTimeout:   resourceRelatedItemRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{string(types.RelatedItemTypeComment), string(types.RelatedItemTypeContact)}, false),
			},
			"content": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"comment": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"body": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"content_type": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						"contact": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"contact_arn": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
					},
				},
				ExactlyOneOf: []string{"content.0.comment", "content.0.contact"},
			},
			"performed_by": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_arn": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
				ExactlyOneOf: []string{"content.0.comment", "content.0.contact"},
			},
			"case_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"related_item_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"related_item_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRelatedItemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	input := &connectcases.CreateRelatedItemInput{
		CaseId:   aws.String(d.Get("case_id").(string)),
		DomainId: aws.String(d.Get("domain_id").(string)),
		Type:     types.RelatedItemType(d.Get("type").(string)),
		//TODO: fix this
		//	Content: expandContent(d.Get("content").([]interface{})),
	}

	output, err := conn.CreateRelatedItem(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Connect Cases Related Item: %s", err)
	}

	d.SetId(aws.ToString(output.RelatedItemId))

	// The below fields are only returned by the Create Related Item API, so we need to set it here.
	d.Set("related_item_id", output.RelatedItemId)
	d.Set("related_item_arn", output.RelatedItemArn)

	return append(diags, resourceRelatedItemRead(ctx, d, meta)...)
}

func resourceRelatedItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ConnectCasesClient(ctx)

	output, err := FindRelatedItemByID(ctx, conn, d.Get("case_id").(string), d.Get("domain_id").(string), d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Cases Related Item (%s): %s", d.Id(), err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "reading Connect Cases Related Item (%s): not found", d.Id())
	}

	d.Set("type", output.Type)
	d.Set("content", output.Content)

	return diags
}

// func expandContent(l []interface{}) *types.RelatedItemContent {
// 	if len(l) == 0 || l[0] == nil {
// 		return nil
// 	}

// 	tfMap, ok := l[0].(map[string]interface{})

// 	if !ok {
// 		return nil
// 	}

// 	result := &types.RelatedItemContent{}

// 	if v, ok := tfMap["comment"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
// 		result.Comment = expandComment(v)
// 	}

// 	if v, ok := tfMap["contact"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
// 		result.Contact = expandContact(v)
// 	}

// 	return result
// }
