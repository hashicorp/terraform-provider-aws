// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_byte_match_set", name="Byte Match Set")
func resourceByteMatchSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceByteMatchSetCreate,
		ReadWithoutTimeout:   resourceByteMatchSetRead,
		UpdateWithoutTimeout: resourceByteMatchSetUpdate,
		DeleteWithoutTimeout: resourceByteMatchSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"byte_match_tuples": {
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
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"positional_constraint": {
							Type:     schema.TypeString,
							Required: true,
						},
						"target_string": {
							Type:     schema.TypeString,
							Optional: true,
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

func resourceByteMatchSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	log.Printf("[INFO] Creating ByteMatchSet: %s", d.Get(names.AttrName).(string))

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateByteMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}
		return conn.CreateByteMatchSetWithContext(ctx, params)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating ByteMatchSet: %s", err)
	}
	resp := out.(*waf.CreateByteMatchSetOutput)

	d.SetId(aws.StringValue(resp.ByteMatchSet.ByteMatchSetId))

	return append(diags, resourceByteMatchSetUpdate(ctx, d, meta)...)
}

func resourceByteMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	log.Printf("[INFO] Reading ByteMatchSet: %s", d.Get(names.AttrName).(string))

	params := &waf.GetByteMatchSetInput{
		ByteMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetByteMatchSetWithContext(ctx, params)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
		log.Printf("[WARN] WAF Regional Byte Set Match (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting WAF Regional Byte Match Set (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() && (resp == nil || resp.ByteMatchSet == nil) {
		log.Printf("[WARN] WAF Regional Byte Set Match (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err := d.Set("byte_match_tuples", flattenByteMatchTuplesWR(resp.ByteMatchSet.ByteMatchTuples)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting byte_match_tuples: %s", err)
	}

	d.Set(names.AttrName, resp.ByteMatchSet.Name)

	return diags
}

func flattenByteMatchTuplesWR(in []*waf.ByteMatchTuple) []interface{} {
	tuples := make([]interface{}, len(in))

	for i, tuple := range in {
		fieldToMatchMap := map[string]interface{}{
			"data":         aws.StringValue(tuple.FieldToMatch.Data),
			names.AttrType: aws.StringValue(tuple.FieldToMatch.Type),
		}

		m := map[string]interface{}{
			"field_to_match":        []map[string]interface{}{fieldToMatchMap},
			"positional_constraint": aws.StringValue(tuple.PositionalConstraint),
			"target_string":         string(tuple.TargetString),
			"text_transformation":   aws.StringValue(tuple.TextTransformation),
		}
		tuples[i] = m
	}

	return tuples
}

func resourceByteMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("byte_match_tuples") {
		o, n := d.GetChange("byte_match_tuples")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()

		if err := updateByteMatchSetResourceWR(ctx, conn, region, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Byte Match Set (%s): %s", d.Id(), err)
		}
	}
	return append(diags, resourceByteMatchSetRead(ctx, d, meta)...)
}

func resourceByteMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldT := d.Get("byte_match_tuples").(*schema.Set).List(); len(oldT) > 0 {
		var newT []interface{}

		err := updateByteMatchSetResourceWR(ctx, conn, region, d.Id(), oldT, newT)

		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentContainerException, wafregional.ErrCodeWAFNonexistentItemException) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Byte Match Set (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Byte Match Set: %s", d.Id())
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteByteMatchSetInput{
			ByteMatchSetId: aws.String(d.Id()),
			ChangeToken:    token,
		}

		return conn.DeleteByteMatchSetWithContext(ctx, input)
	})

	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Byte Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func updateByteMatchSetResourceWR(ctx context.Context, conn *wafregional.WAFRegional, region, byteMatchSetID string, oldT, newT []interface{}) error {
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateByteMatchSetInput{
			ByteMatchSetId: aws.String(byteMatchSetID),
			ChangeToken:    token,
			Updates:        diffByteMatchSetTuple(oldT, newT),
		}

		return conn.UpdateByteMatchSetWithContext(ctx, input)
	})

	return err
}

func diffByteMatchSetTuple(oldT, newT []interface{}) []*waf.ByteMatchSetUpdate {
	updates := make([]*waf.ByteMatchSetUpdate, 0)

	for _, ot := range oldT {
		tuple := ot.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, tuple); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		updates = append(updates, &waf.ByteMatchSetUpdate{
			Action: aws.String(waf.ChangeActionDelete),
			ByteMatchTuple: &waf.ByteMatchTuple{
				FieldToMatch:         tfwaf.ExpandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				PositionalConstraint: aws.String(tuple["positional_constraint"].(string)),
				TargetString:         []byte(tuple["target_string"].(string)),
				TextTransformation:   aws.String(tuple["text_transformation"].(string)),
			},
		})
	}

	for _, nt := range newT {
		tuple := nt.(map[string]interface{})

		updates = append(updates, &waf.ByteMatchSetUpdate{
			Action: aws.String(waf.ChangeActionInsert),
			ByteMatchTuple: &waf.ByteMatchTuple{
				FieldToMatch:         tfwaf.ExpandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				PositionalConstraint: aws.String(tuple["positional_constraint"].(string)),
				TargetString:         []byte(tuple["target_string"].(string)),
				TextTransformation:   aws.String(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}
