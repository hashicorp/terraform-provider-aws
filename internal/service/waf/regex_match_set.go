// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_regex_match_set", name="Regex Match Set")
func resourceRegexMatchSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegexMatchSetCreate,
		ReadWithoutTimeout:   resourceRegexMatchSetRead,
		UpdateWithoutTimeout: resourceRegexMatchSetUpdate,
		DeleteWithoutTimeout: resourceRegexMatchSetDelete,

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
			"regex_match_tuple": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      regexMatchSetTupleHash,
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
										StateFunc: func(v interface{}) string {
											return strings.ToLower(v.(string))
										},
									},
									names.AttrType: {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"regex_pattern_set_id": {
							Type:     schema.TypeString,
							Required: true,
						},
						"text_transformation": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceRegexMatchSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateRegexMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateRegexMatchSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regex Match Set (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*waf.CreateRegexMatchSetOutput).RegexMatchSet.RegexMatchSetId))

	return append(diags, resourceRegexMatchSetUpdate(ctx, d, meta)...)
}

func resourceRegexMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	regexMatchSet, err := findRegexMatchSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regex Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Regex Match Set (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "regexmatchset/" + d.Id(),
	}
	d.Set(names.AttrARN, arn.String())
	d.Set(names.AttrName, regexMatchSet.Name)
	if err := d.Set("regex_match_tuple", flattenRegexMatchTuples(regexMatchSet.RegexMatchTuples)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting regex_match_tuple: %s", err)
	}

	return diags
}

func resourceRegexMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("regex_match_tuple") {
		o, n := d.GetChange("regex_match_tuple")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateRegexMatchSet(ctx, conn, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRegexMatchSetRead(ctx, d, meta)...)
}

func resourceRegexMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if oldTuples := d.Get("regex_match_tuple").(*schema.Set).List(); len(oldTuples) > 0 {
		noTuples := []interface{}{}
		if err := updateRegexMatchSet(ctx, conn, d.Id(), oldTuples, noTuples); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regex Match Set: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteRegexMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regex Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findRegexMatchSetByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.RegexMatchSet, error) {
	input := &waf.GetRegexMatchSetInput{
		RegexMatchSetId: aws.String(id),
	}

	output, err := conn.GetRegexMatchSet(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RegexMatchSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RegexMatchSet, nil
}

func updateRegexMatchSet(ctx context.Context, conn *waf.Client, id string, oldT, newT []interface{}) error {
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(id),
			Updates:         diffRegexMatchSetTuples(oldT, newT),
		}

		return conn.UpdateRegexMatchSet(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF Regex Match Set (%s): %w", id, err)
	}

	return nil
}

func flattenRegexMatchTuples(tuples []awstypes.RegexMatchTuple) []interface{} {
	out := make([]interface{}, len(tuples))
	for i, t := range tuples {
		m := make(map[string]interface{})

		if t.FieldToMatch != nil {
			m["field_to_match"] = flattenFieldToMatch(t.FieldToMatch)
		}
		m["regex_pattern_set_id"] = aws.ToString(t.RegexPatternSetId)
		m["text_transformation"] = string(t.TextTransformation)

		out[i] = m
	}
	return out
}

func expandRegexMatchTuple(tuple map[string]interface{}) *awstypes.RegexMatchTuple {
	ftm := tuple["field_to_match"].([]interface{})
	return &awstypes.RegexMatchTuple{
		FieldToMatch:       expandFieldToMatch(ftm[0].(map[string]interface{})),
		RegexPatternSetId:  aws.String(tuple["regex_pattern_set_id"].(string)),
		TextTransformation: awstypes.TextTransformation(tuple["text_transformation"].(string)),
	}
}

func diffRegexMatchSetTuples(oldT, newT []interface{}) []awstypes.RegexMatchSetUpdate {
	updates := make([]awstypes.RegexMatchSetUpdate, 0)

	for _, ot := range oldT {
		tuple := ot.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, tuple); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		updates = append(updates, awstypes.RegexMatchSetUpdate{
			Action:          awstypes.ChangeActionDelete,
			RegexMatchTuple: expandRegexMatchTuple(tuple),
		})
	}

	for _, nt := range newT {
		tuple := nt.(map[string]interface{})

		updates = append(updates, awstypes.RegexMatchSetUpdate{
			Action:          awstypes.ChangeActionInsert,
			RegexMatchTuple: expandRegexMatchTuple(tuple),
		})
	}
	return updates
}

func regexMatchSetTupleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["field_to_match"]; ok {
		ftms := v.([]interface{})
		ftm := ftms[0].(map[string]interface{})

		if v, ok := ftm["data"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(v.(string))))
		}
		buf.WriteString(fmt.Sprintf("%s-", ftm[names.AttrType]))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["regex_pattern_set_id"].(string)))
	buf.WriteString(fmt.Sprintf("%s-", m["text_transformation"].(string)))

	return create.StringHashcode(buf.String())
}
