// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceByteMatchSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.CreateByteMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateByteMatchSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Byte Match Set (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*wafregional.CreateByteMatchSetOutput).ByteMatchSet.ByteMatchSetId))

	return append(diags, resourceByteMatchSetUpdate(ctx, d, meta)...)
}

func resourceByteMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	byteMatchSet, err := findByteMatchSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regional Byte Set Match (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Regional Byte Set Match (%s): %s", d.Id(), err)
	}

	if err := d.Set("byte_match_tuples", flattenByteMatchTuples(byteMatchSet.ByteMatchTuples)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting byte_match_tuples: %s", err)
	}
	d.Set(names.AttrName, byteMatchSet.Name)

	return diags
}

func resourceByteMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("byte_match_tuples") {
		o, n := d.GetChange("byte_match_tuples")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateByteMatchSet(ctx, conn, region, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceByteMatchSetRead(ctx, d, meta)...)
}

func resourceByteMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldT := d.Get("byte_match_tuples").(*schema.Set).List(); len(oldT) > 0 {
		var newT []interface{}
		if err := updateByteMatchSet(ctx, conn, region, d.Id(), oldT, newT); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Byte Match Set: %s", d.Id())
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.DeleteByteMatchSetInput{
			ByteMatchSetId: aws.String(d.Id()),
			ChangeToken:    token,
		}

		return conn.DeleteByteMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Byte Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findByteMatchSetByID(ctx context.Context, conn *wafregional.Client, id string) (*awstypes.ByteMatchSet, error) {
	input := &wafregional.GetByteMatchSetInput{
		ByteMatchSetId: aws.String(id),
	}

	output, err := conn.GetByteMatchSet(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ByteMatchSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ByteMatchSet, nil
}

func updateByteMatchSet(ctx context.Context, conn *wafregional.Client, region, byteMatchSetID string, oldT, newT []interface{}) error {
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.UpdateByteMatchSetInput{
			ByteMatchSetId: aws.String(byteMatchSetID),
			ChangeToken:    token,
			Updates:        diffByteMatchSetTuple(oldT, newT),
		}

		return conn.UpdateByteMatchSet(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF Regional Byte Match Set (%s): %w", byteMatchSetID, err)
	}

	return nil
}

func flattenByteMatchTuples(in []awstypes.ByteMatchTuple) []interface{} {
	tuples := make([]interface{}, len(in))

	for i, tuple := range in {
		fieldToMatchMap := map[string]interface{}{
			"data":         aws.ToString(tuple.FieldToMatch.Data),
			names.AttrType: tuple.FieldToMatch.Type,
		}

		m := map[string]interface{}{
			"field_to_match":        []map[string]interface{}{fieldToMatchMap},
			"positional_constraint": tuple.PositionalConstraint,
			"target_string":         string(tuple.TargetString),
			"text_transformation":   tuple.TextTransformation,
		}
		tuples[i] = m
	}

	return tuples
}

func diffByteMatchSetTuple(oldT, newT []interface{}) []awstypes.ByteMatchSetUpdate {
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
				FieldToMatch:         expandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
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
				FieldToMatch:         expandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				PositionalConstraint: awstypes.PositionalConstraint(tuple["positional_constraint"].(string)),
				TargetString:         []byte(tuple["target_string"].(string)),
				TextTransformation:   awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}
