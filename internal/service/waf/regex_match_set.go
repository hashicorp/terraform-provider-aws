// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_regex_match_set")
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regex_match_tuple": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      RegexMatchSetTupleHash,
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

	log.Printf("[INFO] Creating WAF Regex Match Set: %s", d.Get(names.AttrName).(string))

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateRegexMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}
		return conn.CreateRegexMatchSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regex Match Set: %s", err)
	}
	resp := out.(*waf.CreateRegexMatchSetOutput)

	d.SetId(aws.ToString(resp.RegexMatchSet.RegexMatchSetId))

	return append(diags, resourceRegexMatchSetUpdate(ctx, d, meta)...)
}

func resourceRegexMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)
	log.Printf("[INFO] Reading WAF Regex Match Set: %s", d.Get(names.AttrName).(string))
	params := &waf.GetRegexMatchSetInput{
		RegexMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetRegexMatchSet(ctx, params)
	if err != nil {
		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] WAF Regex Match Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF Regex Match Set (%s): %s", d.Get(names.AttrName).(string), err)
	}

	d.Set(names.AttrName, resp.RegexMatchSet.Name)
	d.Set("regex_match_tuple", FlattenRegexMatchTuples(resp.RegexMatchSet.RegexMatchTuples))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("regexmatchset/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	return diags
}

func resourceRegexMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	log.Printf("[INFO] Updating WAF Regex Match Set: %s", d.Get(names.AttrName).(string))

	if d.HasChange("regex_match_tuple") {
		o, n := d.GetChange("regex_match_tuple")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		err := updateRegexMatchSetResource(ctx, d.Id(), oldT, newT, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regex Match Set: %s", err)
		}
	}

	return append(diags, resourceRegexMatchSetRead(ctx, d, meta)...)
}

func resourceRegexMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	oldTuples := d.Get("regex_match_tuple").(*schema.Set).List()
	if len(oldTuples) > 0 {
		noTuples := []interface{}{}
		err := updateRegexMatchSetResource(ctx, d.Id(), oldTuples, noTuples, conn)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF Regex Match Set: %s", err)
			}
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF Regex Match Set: %s", d.Id())
		return conn.DeleteRegexMatchSet(ctx, req)
	})
	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regex Match Set: %s", err)
	}

	return diags
}

func updateRegexMatchSetResource(ctx context.Context, id string, oldT, newT []interface{}, conn *waf.Client) error {
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(id),
			Updates:         DiffRegexMatchSetTuples(oldT, newT),
		}

		return conn.UpdateRegexMatchSet(ctx, req)
	})
	return err
}
