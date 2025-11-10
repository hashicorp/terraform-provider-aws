// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"
	"slices"

	"github.com/aws/aws-sdk-go-v2/aws"
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

// @SDKResource("aws_waf_byte_match_set", name="ByteMatchSet")
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceByteMatchSetCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (any, error) {
		input := &waf.CreateByteMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateByteMatchSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF ByteMatchSet (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*waf.CreateByteMatchSetOutput).ByteMatchSet.ByteMatchSetId))

	return append(diags, resourceByteMatchSetUpdate(ctx, d, meta)...)
}

func resourceByteMatchSetRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	byteMatchSet, err := findByteMatchSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF ByteMatchSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF ByteMatchSet (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, byteMatchSetARN(ctx, meta.(*conns.AWSClient), d.Id()))
	if err := d.Set("byte_match_tuples", flattenByteMatchTuples(byteMatchSet.ByteMatchTuples)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting byte_match_tuples: %s", err)
	}
	d.Set(names.AttrName, byteMatchSet.Name)

	return diags
}

func resourceByteMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("byte_match_tuples") {
		o, n := d.GetChange("byte_match_tuples")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateByteMatchSet(ctx, conn, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceByteMatchSetRead(ctx, d, meta)...)
}

func resourceByteMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if oldTuples := d.Get("byte_match_tuples").(*schema.Set).List(); len(oldTuples) > 0 {
		noTuples := []any{}
		if err := updateByteMatchSet(ctx, conn, d.Id(), oldTuples, noTuples); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF ByteMatchSet: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (any, error) {
		input := &waf.DeleteByteMatchSetInput{
			ByteMatchSetId: aws.String(d.Id()),
			ChangeToken:    token,
		}

		return conn.DeleteByteMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF ByteMatchSet (%s): %s", d.Id(), err)
	}

	return diags
}

func findByteMatchSetByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.ByteMatchSet, error) {
	input := &waf.GetByteMatchSetInput{
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

func updateByteMatchSet(ctx context.Context, conn *waf.Client, id string, oldT, newT []any) error {
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (any, error) {
		input := &waf.UpdateByteMatchSetInput{
			ByteMatchSetId: aws.String(id),
			ChangeToken:    token,
			Updates:        diffByteMatchSetTuples(oldT, newT),
		}

		return conn.UpdateByteMatchSet(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF ByteMatchSet (%s): %w", id, err)
	}

	return nil
}

func flattenByteMatchTuples(bmt []awstypes.ByteMatchTuple) []any {
	out := make([]any, len(bmt))
	for i, t := range bmt {
		m := make(map[string]any)

		if t.FieldToMatch != nil {
			m["field_to_match"] = flattenFieldToMatch(t.FieldToMatch)
		}
		m["positional_constraint"] = t.PositionalConstraint
		m["target_string"] = string(t.TargetString)
		m["text_transformation"] = t.TextTransformation

		out[i] = m
	}
	return out
}

func diffByteMatchSetTuples(oldT, newT []any) []awstypes.ByteMatchSetUpdate {
	updates := make([]awstypes.ByteMatchSetUpdate, 0)

	for _, ot := range oldT {
		tuple := ot.(map[string]any)

		if idx, contains := sliceContainsMap(newT, tuple); contains {
			newT = slices.Delete(newT, idx, idx+1)
			continue
		}

		updates = append(updates, awstypes.ByteMatchSetUpdate{
			Action: awstypes.ChangeActionDelete,
			ByteMatchTuple: &awstypes.ByteMatchTuple{
				FieldToMatch:         expandFieldToMatch(tuple["field_to_match"].([]any)[0].(map[string]any)),
				PositionalConstraint: awstypes.PositionalConstraint(tuple["positional_constraint"].(string)),
				TargetString:         []byte(tuple["target_string"].(string)),
				TextTransformation:   awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}

	for _, nt := range newT {
		tuple := nt.(map[string]any)

		updates = append(updates, awstypes.ByteMatchSetUpdate{
			Action: awstypes.ChangeActionInsert,
			ByteMatchTuple: &awstypes.ByteMatchTuple{
				FieldToMatch:         expandFieldToMatch(tuple["field_to_match"].([]any)[0].(map[string]any)),
				PositionalConstraint: awstypes.PositionalConstraint(tuple["positional_constraint"].(string)),
				TargetString:         []byte(tuple["target_string"].(string)),
				TextTransformation:   awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}

// See https://docs.aws.amazon.com/service-authorization/latest/reference/list_awswaf.html#awswaf-resources-for-iam-policies.
func byteMatchSetARN(ctx context.Context, c *conns.AWSClient, id string) string {
	return c.GlobalARN(ctx, "waf", "bytematchset/"+id)
}
