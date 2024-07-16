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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &wafregional.CreateSqlInjectionMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateSqlInjectionMatchSet(ctx, params)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional SQL Injection Match Set (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*wafregional.CreateSqlInjectionMatchSetOutput).SqlInjectionMatchSet.SqlInjectionMatchSetId))

	return append(diags, resourceSQLInjectionMatchSetUpdate(ctx, d, meta)...)
}

func resourceSQLInjectionMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	sqlInjectionMatchSet, err := findSQLInjectionMatchSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regional SQL Injection Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Regional SQL Injection Match Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, sqlInjectionMatchSet.Name)
	if err := d.Set("sql_injection_match_tuple", flattenSQLInjectionMatchTuples(sqlInjectionMatchSet.SqlInjectionMatchTuples)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting sql_injection_match_tuple: %s", err)
	}

	return diags
}

func resourceSQLInjectionMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("sql_injection_match_tuple") {
		o, n := d.GetChange("sql_injection_match_tuple")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateSQLInjectionMatchSet(ctx, conn, region, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceSQLInjectionMatchSetRead(ctx, d, meta)...)
}

func resourceSQLInjectionMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldTuples := d.Get("sql_injection_match_tuple").(*schema.Set).List(); len(oldTuples) > 0 {
		noTuples := []interface{}{}
		if err := updateSQLInjectionMatchSet(ctx, conn, region, d.Id(), oldTuples, noTuples); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.DeleteSqlInjectionMatchSetInput{
			ChangeToken:            token,
			SqlInjectionMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteSqlInjectionMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional SQL Injection Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findSQLInjectionMatchSetByID(ctx context.Context, conn *wafregional.Client, id string) (*awstypes.SqlInjectionMatchSet, error) {
	input := &wafregional.GetSqlInjectionMatchSetInput{
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

func updateSQLInjectionMatchSet(ctx context.Context, conn *wafregional.Client, region, id string, oldT, newT []interface{}) error {
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.UpdateSqlInjectionMatchSetInput{
			ChangeToken:            token,
			SqlInjectionMatchSetId: aws.String(id),
			Updates:                diffSQLInjectionMatchTuplesWR(oldT, newT),
		}

		return conn.UpdateSqlInjectionMatchSet(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF Regional SQL Injection Match Set (%s): %w", id, err)
	}

	return nil
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
				FieldToMatch:       expandFieldToMatch(ftm[0].(map[string]interface{})),
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
				FieldToMatch:       expandFieldToMatch(ftm[0].(map[string]interface{})),
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
		m["field_to_match"] = flattenFieldToMatch(t.FieldToMatch)
		out[i] = m
	}

	return out
}
