// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
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

// @SDKResource("aws_waf_byte_match_set")
func ResourceByteMatchSet() *schema.Resource {
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
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[awstypes.MatchFieldType](),
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
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	log.Printf("[INFO] Creating WAF ByteMatchSet: %s", d.Get(names.AttrName).(string))

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateByteMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}
		return conn.CreateByteMatchSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF ByteMatchSet: %s", err)
	}
	resp := out.(*waf.CreateByteMatchSetOutput)

	d.SetId(aws.ToString(resp.ByteMatchSet.ByteMatchSetId))

	return append(diags, resourceByteMatchSetUpdate(ctx, d, meta)...)
}

func resourceByteMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)
	log.Printf("[INFO] Reading WAF ByteMatchSet: %s", d.Get(names.AttrName).(string))
	params := &waf.GetByteMatchSetInput{
		ByteMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetByteMatchSet(ctx, params)
	if err != nil {
		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] WAF ByteMatchSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF ByteMatchSet (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, resp.ByteMatchSet.Name)
	d.Set("byte_match_tuples", flattenByteMatchTuples(resp.ByteMatchSet.ByteMatchTuples))

	return diags
}

func resourceByteMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	log.Printf("[INFO] Updating WAF ByteMatchSet: %s", d.Get(names.AttrName).(string))

	if d.HasChange("byte_match_tuples") {
		o, n := d.GetChange("byte_match_tuples")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		err := updateByteMatchSetResource(ctx, d.Id(), oldT, newT, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF ByteMatchSet: %s", err)
		}
	}

	return append(diags, resourceByteMatchSetRead(ctx, d, meta)...)
}

func resourceByteMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	oldTuples := d.Get("byte_match_tuples").(*schema.Set).List()
	if len(oldTuples) > 0 {
		noTuples := []interface{}{}
		err := updateByteMatchSetResource(ctx, d.Id(), oldTuples, noTuples, conn)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF ByteMatchSet: %s", err)
			}
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteByteMatchSetInput{
			ChangeToken:    token,
			ByteMatchSetId: aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF ByteMatchSet: %s", d.Id())
		return conn.DeleteByteMatchSet(ctx, req)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF ByteMatchSet: %s", err)
	}

	return diags
}

func updateByteMatchSetResource(ctx context.Context, id string, oldT, newT []interface{}, conn *waf.Client) error {
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateByteMatchSetInput{
			ChangeToken:    token,
			ByteMatchSetId: aws.String(id),
			Updates:        diffByteMatchSetTuples(oldT, newT),
		}

		return conn.UpdateByteMatchSet(ctx, req)
	})
	return err
}

func flattenByteMatchTuples(bmt []awstypes.ByteMatchTuple) []interface{} {
	out := make([]interface{}, len(bmt))
	for i, t := range bmt {
		m := make(map[string]interface{})

		if t.FieldToMatch != nil {
			m["field_to_match"] = FlattenFieldToMatch(t.FieldToMatch)
		}
		m["positional_constraint"] = t.PositionalConstraint
		m["target_string"] = string(t.TargetString)
		m["text_transformation"] = t.TextTransformation

		out[i] = m
	}
	return out
}

func diffByteMatchSetTuples(oldT, newT []interface{}) []awstypes.ByteMatchSetUpdate {
	updates := make([]awstypes.ByteMatchSetUpdate, 0)

	for _, ot := range oldT {
		tuple := ot.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, tuple); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		updates = append(updates, awstypes.ByteMatchSetUpdate{
			Action: awstypes.ChangeActionDelete,
			ByteMatchTuple: &awstypes.ByteMatchTuple{
				FieldToMatch:         ExpandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				PositionalConstraint: awstypes.PositionalConstraint(tuple["positional_constraint"].(string)),
				TargetString:         []byte(tuple["target_string"].(string)),
				TextTransformation:   awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}

	for _, nt := range newT {
		tuple := nt.(map[string]interface{})

		updates = append(updates, awstypes.ByteMatchSetUpdate{
			Action: awstypes.ChangeActionInsert,
			ByteMatchTuple: &awstypes.ByteMatchTuple{
				FieldToMatch:         ExpandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				PositionalConstraint: awstypes.PositionalConstraint(tuple["positional_constraint"].(string)),
				TargetString:         []byte(tuple["target_string"].(string)),
				TextTransformation:   awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}
