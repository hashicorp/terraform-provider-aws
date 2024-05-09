// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_xss_match_set")
func ResourceXSSMatchSet() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"xss_match_tuples": {
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
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	log.Printf("[INFO] Creating XssMatchSet: %s", d.Get(names.AttrName).(string))

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateXssMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}

		return conn.CreateXssMatchSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF XSS Match Set: %s", err)
	}
	resp := out.(*waf.CreateXssMatchSetOutput)

	d.SetId(aws.ToString(resp.XssMatchSet.XssMatchSetId))

	if v, ok := d.GetOk("xss_match_tuples"); ok && v.(*schema.Set).Len() > 0 {
		err := updateXSSMatchSetResource(ctx, d.Id(), nil, v.(*schema.Set).List(), conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting WAF XSS Match Set tuples: %s", err)
		}
	}
	return append(diags, resourceXSSMatchSetRead(ctx, d, meta)...)
}

func resourceXSSMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)
	log.Printf("[INFO] Reading WAF XSS Match Set: %s", d.Get(names.AttrName).(string))
	params := &waf.GetXssMatchSetInput{
		XssMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetXssMatchSet(ctx, params)
	if err != nil {
		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] WAF XSS Match Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF XSS Match Set (%s): %s", d.Get(names.AttrName).(string), err)
	}

	d.Set(names.AttrName, resp.XssMatchSet.Name)
	if err := d.Set("xss_match_tuples", flattenXSSMatchTuples(resp.XssMatchSet.XssMatchTuples)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting xss_match_tuples: %s", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("xssmatchset/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	return diags
}

func resourceXSSMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("xss_match_tuples") {
		o, n := d.GetChange("xss_match_tuples")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateXSSMatchSetResource(ctx, d.Id(), oldT, newT, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF XSS Match Set: %s", err)
		}
	}

	return append(diags, resourceXSSMatchSetRead(ctx, d, meta)...)
}

func resourceXSSMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	oldTuples := d.Get("xss_match_tuples").(*schema.Set).List()
	if len(oldTuples) > 0 {
		err := updateXSSMatchSetResource(ctx, d.Id(), oldTuples, nil, conn)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "removing WAF XSS Match Set tuples: %s", err)
			}
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteXssMatchSetInput{
			ChangeToken:   token,
			XssMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteXssMatchSet(ctx, req)
	})
	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF XSS Match Set: %s", err)
	}

	return diags
}

func updateXSSMatchSetResource(ctx context.Context, id string, oldT, newT []interface{}, conn *waf.Client) error {
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateXssMatchSetInput{
			ChangeToken:   token,
			XssMatchSetId: aws.String(id),
			Updates:       diffXSSMatchSetTuples(oldT, newT),
		}

		log.Printf("[INFO] Updating WAF XSS Match Set tuples: %s", id)
		return conn.UpdateXssMatchSet(ctx, req)
	})
	return err
}

func flattenXSSMatchTuples(ts []awstypes.XssMatchTuple) []interface{} {
	out := make([]interface{}, len(ts))
	for i, t := range ts {
		m := make(map[string]interface{})
		m["field_to_match"] = flattenFieldToMatch(t.FieldToMatch)
		m["text_transformation"] = string(t.TextTransformation)
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
				FieldToMatch:       expandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}

	for _, nd := range newT {
		tuple := nd.(map[string]interface{})

		updates = append(updates, awstypes.XssMatchSetUpdate{
			Action: awstypes.ChangeActionInsert,
			XssMatchTuple: &awstypes.XssMatchTuple{
				FieldToMatch:       expandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}
