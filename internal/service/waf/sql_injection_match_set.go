// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_sql_injection_match_set", name="SqlInjectionMatchSet")
func resourceSQLInjectionMatchSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSQLInjectionMatchSetCreate,
		ReadWithoutTimeout:   resourceSQLInjectionMatchSetRead,
		UpdateWithoutTimeout: resourceSQLInjectionMatchSetUpdate,
		DeleteWithoutTimeout: resourceSQLInjectionMatchSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"sql_injection_match_tuples": {
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

func resourceSQLInjectionMatchSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateSqlInjectionMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateSqlInjectionMatchSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF SqlInjectionMatchSet (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*waf.CreateSqlInjectionMatchSetOutput).SqlInjectionMatchSet.SqlInjectionMatchSetId))

	return append(diags, resourceSQLInjectionMatchSetUpdate(ctx, d, meta)...)
}

func resourceSQLInjectionMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	sqlInjectionMatchSet, err := findSQLInjectionMatchSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF SqlInjectionMatchSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF SqlInjectionMatchSet (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, sqlInjectionMatchSet.Name)
	if err := d.Set("sql_injection_match_tuples", flattenSQLInjectionMatchTuples(sqlInjectionMatchSet.SqlInjectionMatchTuples)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sql_injection_match_tuples: %s", err)
	}

	return diags
}

func resourceSQLInjectionMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("sql_injection_match_tuples") {
		o, n := d.GetChange("sql_injection_match_tuples")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateSQLInjectionMatchSet(ctx, conn, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceSQLInjectionMatchSetRead(ctx, d, meta)...)
}

func resourceSQLInjectionMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if oldTuples := d.Get("sql_injection_match_tuples").(*schema.Set).List(); len(oldTuples) > 0 {
		noTuples := []interface{}{}
		if err := updateSQLInjectionMatchSet(ctx, conn, d.Id(), oldTuples, noTuples); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF SqlInjectionMatchSet: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteSqlInjectionMatchSetInput{
			ChangeToken:            token,
			SqlInjectionMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteSqlInjectionMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF SqlInjectionMatchSet (%s): %s", d.Id(), err)
	}

	return diags
}

func findSQLInjectionMatchSetByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.SqlInjectionMatchSet, error) {
	input := &waf.GetSqlInjectionMatchSetInput{
		SqlInjectionMatchSetId: aws.String(id),
	}

	output, err := conn.GetSqlInjectionMatchSet(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SqlInjectionMatchSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.SqlInjectionMatchSet, nil
}

func updateSQLInjectionMatchSet(ctx context.Context, conn *waf.Client, id string, oldT, newT []interface{}) error {
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateSqlInjectionMatchSetInput{
			ChangeToken:            token,
			SqlInjectionMatchSetId: aws.String(id),
			Updates:                diffSQLInjectionMatchTuples(oldT, newT),
		}

		return conn.UpdateSqlInjectionMatchSet(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF SqlInjectionMatchSet (%s): %w", id, err)
	}

	return nil
}

func flattenSQLInjectionMatchTuples(ts []awstypes.SqlInjectionMatchTuple) []interface{} {
	out := make([]interface{}, len(ts))
	for i, t := range ts {
		m := make(map[string]interface{})
		m["text_transformation"] = t.TextTransformation
		m["field_to_match"] = flattenFieldToMatch(t.FieldToMatch)
		out[i] = m
	}

	return out
}

func diffSQLInjectionMatchTuples(oldT, newT []interface{}) []awstypes.SqlInjectionMatchSetUpdate {
	updates := make([]awstypes.SqlInjectionMatchSetUpdate, 0)

	for _, od := range oldT {
		tuple := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, tuple); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		updates = append(updates, awstypes.SqlInjectionMatchSetUpdate{
			Action: awstypes.ChangeActionDelete,
			SqlInjectionMatchTuple: &awstypes.SqlInjectionMatchTuple{
				FieldToMatch:       expandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}

	for _, nd := range newT {
		tuple := nd.(map[string]interface{})

		updates = append(updates, awstypes.SqlInjectionMatchSetUpdate{
			Action: awstypes.ChangeActionInsert,
			SqlInjectionMatchTuple: &awstypes.SqlInjectionMatchTuple{
				FieldToMatch:       expandFieldToMatch(tuple["field_to_match"].([]interface{})[0].(map[string]interface{})),
				TextTransformation: awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}
