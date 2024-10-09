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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_xss_match_set", name="XSS Match Set")
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateXssMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateXssMatchSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF XSS Match Set (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*waf.CreateXssMatchSetOutput).XssMatchSet.XssMatchSetId))

	if v, ok := d.GetOk("xss_match_tuples"); ok && v.(*schema.Set).Len() > 0 {
		if err := updateXSSMatchSet(ctx, conn, d.Id(), nil, v.(*schema.Set).List()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceXSSMatchSetRead(ctx, d, meta)...)
}

func resourceXSSMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	xssMatchSet, err := findXSSMatchSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF XSS Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF XSS Match Set (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "xssmatchset/" + d.Id(),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set(names.AttrName, xssMatchSet.Name)
	if err := d.Set("xss_match_tuples", flattenXSSMatchTuples(xssMatchSet.XssMatchTuples)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting xss_match_tuples: %s", err)
	}

	return diags
}

func resourceXSSMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("xss_match_tuples") {
		o, n := d.GetChange("xss_match_tuples")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateXSSMatchSet(ctx, conn, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceXSSMatchSetRead(ctx, d, meta)...)
}

func resourceXSSMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if oldTuples := d.Get("xss_match_tuples").(*schema.Set).List(); len(oldTuples) > 0 {
		if err := updateXSSMatchSet(ctx, conn, d.Id(), oldTuples, nil); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF XSS Match Set: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteXssMatchSetInput{
			ChangeToken:   token,
			XssMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteXssMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF XSS Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findXSSMatchSetByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.XssMatchSet, error) {
	input := &waf.GetXssMatchSetInput{
		XssMatchSetId: aws.String(id),
	}

	output, err := conn.GetXssMatchSet(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.XssMatchSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.XssMatchSet, nil
}

func updateXSSMatchSet(ctx context.Context, conn *waf.Client, id string, oldT, newT []interface{}) error {
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateXssMatchSetInput{
			ChangeToken:   token,
			Updates:       diffXSSMatchSetTuples(oldT, newT),
			XssMatchSetId: aws.String(id),
		}

		return conn.UpdateXssMatchSet(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF XSS Match Set (%s): %w", id, err)
	}

	return nil
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
