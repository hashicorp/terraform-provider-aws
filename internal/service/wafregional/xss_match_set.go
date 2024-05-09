// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_xss_match_set", name="XSS Match Set")
func resourceXSSMatchSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceXSSMatchSetCreate,
		ReadWithoutTimeout:   resourceXSSMatchSetRead,
		UpdateWithoutTimeout: resourceXSSMatchSetUpdate,
		DeleteWithoutTimeout: resourceXSSMatchSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"xss_match_tuple": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field_to_match": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data": {
										Type:     schema.TypeString,
										Optional: true,
									},
									names.AttrType: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.MatchFieldType](),
									},
								},
							},
						},
						"text_transformation": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TextTransformation](),
						},
					},
				},
			},
		},
	}
}

func resourceXSSMatchSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	log.Printf("[INFO] Creating regional WAF XSS Match Set: %s", d.Get(names.AttrName).(string))

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &wafregional.CreateXssMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}

		return conn.CreateXssMatchSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional XSS Match Set: %s", err)
	}
	resp := out.(*wafregional.CreateXssMatchSetOutput)

	d.SetId(aws.ToString(resp.XssMatchSet.XssMatchSetId))

	if v, ok := d.Get("xss_match_tuple").(*schema.Set); ok && v.Len() > 0 {
		if err := updateXSSMatchSetResourceWR(ctx, conn, region, d.Id(), nil, v.List()); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional XSS Match Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceXSSMatchSetRead(ctx, d, meta)...)
}

func resourceXSSMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	params := &wafregional.GetXssMatchSetInput{
		XssMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetXssMatchSet(ctx, params)
	if err != nil {
		if !d.IsNewResource() && errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] Regional WAF XSS Match Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF Regional XSS Match Set: %s", err)
	}

	set := resp.XssMatchSet

	if err := d.Set("xss_match_tuple", flattenXSSMatchTuples(set.XssMatchTuples)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting xss_match_tuple: %s", err)
	}
	d.Set(names.AttrName, set.Name)

	return diags
}

func resourceXSSMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("xss_match_tuple") {
		o, n := d.GetChange("xss_match_tuple")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()

		if err := updateXSSMatchSetResourceWR(ctx, conn, region, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional XSS Match Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceXSSMatchSetRead(ctx, d, meta)...)
}

func resourceXSSMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldT := d.Get("xss_match_tuple").(*schema.Set).List(); len(oldT) > 0 {
		var newT []interface{}

		err := updateXSSMatchSetResourceWR(ctx, conn, region, d.Id(), oldT, newT)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF Regional XSS Match Set (%s): %s", d.Id(), err)
			}
		}
	}

	log.Printf("[INFO] Deleting WAF Regional XSS Match Set: %s", d.Id())
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.DeleteXssMatchSetInput{
			ChangeToken:   token,
			XssMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteXssMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional XSS Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func updateXSSMatchSetResourceWR(ctx context.Context, conn *wafregional.Client, region, xssMatchSetID string, oldT, newT []interface{}) error {
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.UpdateXssMatchSetInput{
			ChangeToken:   token,
			Updates:       diffXSSMatchSetTuples(oldT, newT),
			XssMatchSetId: aws.String(xssMatchSetID),
		}

		return conn.UpdateXssMatchSet(ctx, input)
	})

	return err
}

func flattenXSSMatchTuples(ts []awstypes.XssMatchTuple) []interface{} {
	out := make([]interface{}, len(ts))
	for i, t := range ts {
		m := make(map[string]interface{})
		m["field_to_match"] = FlattenFieldToMatch(t.FieldToMatch)
		m["text_transformation"] = t.TextTransformation
		out[i] = m
	}
	return out
}

func diffXSSMatchSetTuples(oldT, newT []interface{}) []awstypes.XssMatchSetUpdate {
	updates := make([]awstypes.XssMatchSetUpdate, 0)

	for _, od := range oldT {
		tuple := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, tuple); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		updates = append(updates, awstypes.XssMatchSetUpdate{
			Action: awstypes.ChangeActionDelete,
			XssMatchTuple: &awstypes.XssMatchTuple{
				FieldToMatch:       ExpandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}

	for _, nd := range newT {
		tuple := nd.(map[string]interface{})

		updates = append(updates, awstypes.XssMatchSetUpdate{
			Action: awstypes.ChangeActionInsert,
			XssMatchTuple: &awstypes.XssMatchTuple{
				FieldToMatch:       ExpandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}
