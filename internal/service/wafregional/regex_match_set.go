// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
)

// @SDKResource("aws_wafregional_regex_match_set")
func ResourceRegexMatchSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegexMatchSetCreate,
		ReadWithoutTimeout:   resourceRegexMatchSetRead,
		UpdateWithoutTimeout: resourceRegexMatchSetUpdate,
		DeleteWithoutTimeout: resourceRegexMatchSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"regex_match_tuple": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      tfwaf.RegexMatchSetTupleHash,
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
									"type": {
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
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	log.Printf("[INFO] Creating WAF Regional Regex Match Set: %s", d.Get("name").(string))

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateRegexMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}
		return conn.CreateRegexMatchSetWithContext(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Regex Match Set: %s", err)
	}
	resp := out.(*waf.CreateRegexMatchSetOutput)

	d.SetId(aws.StringValue(resp.RegexMatchSet.RegexMatchSetId))

	return append(diags, resourceRegexMatchSetUpdate(ctx, d, meta)...)
}

func resourceRegexMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	set, err := FindRegexMatchSetByID(ctx, conn, d.Id())
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		log.Printf("[WARN] WAF Regional Regex Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting WAF Regional Regex Match Set (%s): %s", d.Id(), err)
	}

	d.Set("name", set.Name)
	d.Set("regex_match_tuple", tfwaf.FlattenRegexMatchTuples(set.RegexMatchTuples))

	return diags
}

func resourceRegexMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("regex_match_tuple") {
		o, n := d.GetChange("regex_match_tuple")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		err := updateRegexMatchSetResourceWR(ctx, d.Id(), oldT, newT, conn, region)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Regex Match Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRegexMatchSetRead(ctx, d, meta)...)
}

func resourceRegexMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	err := DeleteRegexMatchSetResource(ctx, conn, region, "global", d.Id(), getRegexMatchTuplesFromResourceData(d))

	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Regex Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func getRegexMatchTuplesFromResourceData(d *schema.ResourceData) []*waf.RegexMatchTuple {
	result := []*waf.RegexMatchTuple{}

	for _, t := range d.Get("regex_match_tuple").(*schema.Set).List() {
		result = append(result, tfwaf.ExpandRegexMatchTuple(t.(map[string]interface{})))
	}

	return result
}

func GetRegexMatchTuplesFromAPIResource(r *waf.RegexMatchSet) []*waf.RegexMatchTuple {
	return r.RegexMatchTuples
}

func clearRegexMatchTuples(ctx context.Context, conn *wafregional.WAFRegional, region string, id string, tuples []*waf.RegexMatchTuple) error {
	if len(tuples) > 0 {
		input := &waf.UpdateRegexMatchSetInput{
			RegexMatchSetId: aws.String(id),
		}
		for _, tuple := range tuples {
			input.Updates = append(input.Updates, &waf.RegexMatchSetUpdate{
				Action:          aws.String(waf.ChangeActionDelete),
				RegexMatchTuple: tuple,
			})
		}

		log.Printf("[INFO] Clearing WAF Regional Regex Match Set: %s", id)
		wr := NewRetryer(conn, region)
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			input.ChangeToken = token
			return conn.UpdateRegexMatchSetWithContext(ctx, input)
		})
		if err != nil {
			return fmt.Errorf("clearing WAF Regional Regex Match Set (%s): %w", id, err)
		}
	}
	return nil
}

func deleteRegexMatchSet(ctx context.Context, conn *wafregional.WAFRegional, region, id string) error {
	log.Printf("[INFO] Deleting WAF Regional Regex Match Set: %s", id)
	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(id),
		}
		return conn.DeleteRegexMatchSetWithContext(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("deleting WAF Regional Regex Match Set (%s): %w", id, err)
	}
	return nil
}

func DeleteRegexMatchSetResource(ctx context.Context, conn *wafregional.WAFRegional, region, tokenRegion, id string, tuples []*waf.RegexMatchTuple) error {
	err := clearRegexMatchTuples(ctx, conn, region, id, tuples)
	if err != nil {
		return err
	}

	return deleteRegexMatchSet(ctx, conn, tokenRegion, id)
}

func updateRegexMatchSetResourceWR(ctx context.Context, id string, oldT, newT []interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(id),
			Updates:         tfwaf.DiffRegexMatchSetTuples(oldT, newT),
		}

		return conn.UpdateRegexMatchSetWithContext(ctx, req)
	})

	return err
}
