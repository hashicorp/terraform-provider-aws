// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_sql_injection_match_set", name="SQL Injection Match Set")
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
			"sql_injection_match_tuple": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      resourceSQLInjectionMatchSetTupleHash,
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
											value := v.(string)
											return strings.ToLower(value)
										},
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
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	log.Printf("[INFO] Creating Regional WAF SQL Injection Match Set: %s", d.Get(names.AttrName).(string))

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &wafregional.CreateSqlInjectionMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}

		return conn.CreateSqlInjectionMatchSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Regional WAF SQL Injection Match Set: %s", err)
	}
	resp := out.(*wafregional.CreateSqlInjectionMatchSetOutput)
	d.SetId(aws.ToString(resp.SqlInjectionMatchSet.SqlInjectionMatchSetId))

	return append(diags, resourceSQLInjectionMatchSetUpdate(ctx, d, meta)...)
}

func resourceSQLInjectionMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	log.Printf("[INFO] Reading Regional WAF SQL Injection Match Set: %s", d.Get(names.AttrName).(string))
	params := &wafregional.GetSqlInjectionMatchSetInput{
		SqlInjectionMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetSqlInjectionMatchSet(ctx, params)
	if !d.IsNewResource() && errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		log.Printf("[WARN] Regional WAF SQL Injection Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Regional WAF SQL Injection Match Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, resp.SqlInjectionMatchSet.Name)
	d.Set("sql_injection_match_tuple", flattenSQLInjectionMatchTuples(resp.SqlInjectionMatchSet.SqlInjectionMatchTuples))

	return diags
}

func resourceSQLInjectionMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("sql_injection_match_tuple") {
		o, n := d.GetChange("sql_injection_match_tuple")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateSQLInjectionMatchSetResourceWR(ctx, d.Id(), oldT, newT, conn, region)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Regional WAF SQL Injection Match Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceSQLInjectionMatchSetRead(ctx, d, meta)...)
}

func resourceSQLInjectionMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	oldTuples := d.Get("sql_injection_match_tuple").(*schema.Set).List()

	if len(oldTuples) > 0 {
		noTuples := []interface{}{}
		err := updateSQLInjectionMatchSetResourceWR(ctx, d.Id(), oldTuples, noTuples, conn, region)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating Regional WAF SQL Injection Match Set (%s): %s", d.Id(), err)
			}
		}
	}

	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &wafregional.DeleteSqlInjectionMatchSetInput{
			ChangeToken:            token,
			SqlInjectionMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteSqlInjectionMatchSet(ctx, req)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional SQL Injection Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func updateSQLInjectionMatchSetResourceWR(ctx context.Context, id string, oldT, newT []interface{}, conn *wafregional.Client, region string) error {
	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &wafregional.UpdateSqlInjectionMatchSetInput{
			ChangeToken:            token,
			SqlInjectionMatchSetId: aws.String(id),
			Updates:                diffSQLInjectionMatchTuplesWR(oldT, newT),
		}

		log.Printf("[INFO] Updating Regional WAF SQL Injection Match Set: %s", id)
		return conn.UpdateSqlInjectionMatchSet(ctx, req)
	})

	return err
}

func diffSQLInjectionMatchTuplesWR(oldT, newT []interface{}) []awstypes.SqlInjectionMatchSetUpdate {
	updates := make([]awstypes.SqlInjectionMatchSetUpdate, 0)

	for _, od := range oldT {
		tuple := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, tuple); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		ftm := tuple["field_to_match"].([]interface{})

		updates = append(updates, awstypes.SqlInjectionMatchSetUpdate{
			Action: awstypes.ChangeActionDelete,
			SqlInjectionMatchTuple: &awstypes.SqlInjectionMatchTuple{
				FieldToMatch:       ExpandFieldToMatch(ftm[0].(map[string]interface{})),
				TextTransformation: awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}

	for _, nd := range newT {
		tuple := nd.(map[string]interface{})
		ftm := tuple["field_to_match"].([]interface{})

		updates = append(updates, awstypes.SqlInjectionMatchSetUpdate{
			Action: awstypes.ChangeActionInsert,
			SqlInjectionMatchTuple: &awstypes.SqlInjectionMatchTuple{
				FieldToMatch:       ExpandFieldToMatch(ftm[0].(map[string]interface{})),
				TextTransformation: awstypes.TextTransformation(tuple["text_transformation"].(string)),
			},
		})
	}
	return updates
}

func resourceSQLInjectionMatchSetTupleHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["field_to_match"]; ok {
		ftms := v.([]interface{})
		ftm := ftms[0].(map[string]interface{})

		if v, ok := ftm["data"]; ok {
			buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(v.(string))))
		}
		buf.WriteString(fmt.Sprintf("%s-", ftm[names.AttrType].(string)))
	}
	buf.WriteString(fmt.Sprintf("%s-", m["text_transformation"].(string)))

	return create.StringHashcode(buf.String())
}

func flattenSQLInjectionMatchTuples(ts []awstypes.SqlInjectionMatchTuple) []interface{} {
	out := make([]interface{}, len(ts))
	for i, t := range ts {
		m := make(map[string]interface{})
		m["text_transformation"] = t.TextTransformation
		m["field_to_match"] = FlattenFieldToMatch(t.FieldToMatch)
		out[i] = m
	}

	return out
}
