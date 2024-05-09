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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_sql_injection_match_set")
func ResourceSQLInjectionMatchSet() *schema.Resource {
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

	log.Printf("[INFO] Creating SqlInjectionMatchSet: %s", d.Get(names.AttrName).(string))

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateSqlInjectionMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}

		return conn.CreateSqlInjectionMatchSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SqlInjectionMatchSet: %s", err)
	}
	resp := out.(*waf.CreateSqlInjectionMatchSetOutput)
	d.SetId(aws.ToString(resp.SqlInjectionMatchSet.SqlInjectionMatchSetId))

	return append(diags, resourceSQLInjectionMatchSetUpdate(ctx, d, meta)...)
}

func resourceSQLInjectionMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)
	log.Printf("[INFO] Reading SqlInjectionMatchSet: %s", d.Get(names.AttrName).(string))
	params := &waf.GetSqlInjectionMatchSetInput{
		SqlInjectionMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetSqlInjectionMatchSet(ctx, params)
	if err != nil {
		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] WAF IPSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF SQL Injection Match Set (%s): %s", d.Get(names.AttrName).(string), err)
	}

	d.Set(names.AttrName, resp.SqlInjectionMatchSet.Name)

	if err := d.Set("sql_injection_match_tuples", flattenSQLInjectionMatchTuples(resp.SqlInjectionMatchSet.SqlInjectionMatchTuples)); err != nil {
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

		err := updateSQLInjectionMatchSetResource(ctx, d.Id(), oldT, newT, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SqlInjectionMatchSet: %s", err)
		}
	}

	return append(diags, resourceSQLInjectionMatchSetRead(ctx, d, meta)...)
}

func resourceSQLInjectionMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	oldTuples := d.Get("sql_injection_match_tuples").(*schema.Set).List()

	if len(oldTuples) > 0 {
		noTuples := []interface{}{}
		err := updateSQLInjectionMatchSetResource(ctx, d.Id(), oldTuples, noTuples, conn)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "deleting SqlInjectionMatchSet: %s", err)
			}
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteSqlInjectionMatchSetInput{
			ChangeToken:            token,
			SqlInjectionMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteSqlInjectionMatchSet(ctx, req)
	})
	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SqlInjectionMatchSet: %s", err)
	}

	return diags
}

func updateSQLInjectionMatchSetResource(ctx context.Context, id string, oldT, newT []interface{}, conn *waf.Client) error {
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateSqlInjectionMatchSetInput{
			ChangeToken:            token,
			SqlInjectionMatchSetId: aws.String(id),
			Updates:                diffSQLInjectionMatchTuples(oldT, newT),
		}

		log.Printf("[INFO] Updating SqlInjectionMatchSet: %s", id)
		return conn.UpdateSqlInjectionMatchSet(ctx, req)
	})
	return err
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
