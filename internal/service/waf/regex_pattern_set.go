// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"

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

// @SDKResource("aws_waf_regex_pattern_set")
func ResourceRegexPatternSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegexPatternSetCreate,
		ReadWithoutTimeout:   resourceRegexPatternSetRead,
		UpdateWithoutTimeout: resourceRegexPatternSetUpdate,
		DeleteWithoutTimeout: resourceRegexPatternSetDelete,
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
			"regex_pattern_strings": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceRegexPatternSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	log.Printf("[INFO] Creating WAF Regex Pattern Set: %s", d.Get(names.AttrName).(string))

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateRegexPatternSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}
		return conn.CreateRegexPatternSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regex Pattern Set: %s", err)
	}
	resp := out.(*waf.CreateRegexPatternSetOutput)

	d.SetId(aws.ToString(resp.RegexPatternSet.RegexPatternSetId))

	return append(diags, resourceRegexPatternSetUpdate(ctx, d, meta)...)
}

func resourceRegexPatternSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)
	log.Printf("[INFO] Reading WAF Regex Pattern Set: %s", d.Get(names.AttrName).(string))
	params := &waf.GetRegexPatternSetInput{
		RegexPatternSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetRegexPatternSet(ctx, params)
	if err != nil {
		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] WAF Regex Pattern Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF Regex Pattern Set (%s): %s", d.Get(names.AttrName).(string), err)
	}

	d.Set(names.AttrName, resp.RegexPatternSet.Name)
	d.Set("regex_pattern_strings", aws.StringSlice(resp.RegexPatternSet.RegexPatternStrings))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("regexpatternset/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	return diags
}

func resourceRegexPatternSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	log.Printf("[INFO] Updating WAF Regex Pattern Set: %s", d.Get(names.AttrName).(string))

	if d.HasChange("regex_pattern_strings") {
		o, n := d.GetChange("regex_pattern_strings")
		oldPatterns, newPatterns := o.(*schema.Set).List(), n.(*schema.Set).List()
		err := updateRegexPatternSetPatternStrings(ctx, d.Id(), oldPatterns, newPatterns, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regex Pattern Set: %s", err)
		}
	}

	return append(diags, resourceRegexPatternSetRead(ctx, d, meta)...)
}

func resourceRegexPatternSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	oldPatterns := d.Get("regex_pattern_strings").(*schema.Set).List()
	if len(oldPatterns) > 0 {
		noPatterns := []interface{}{}
		err := updateRegexPatternSetPatternStrings(ctx, d.Id(), oldPatterns, noPatterns, conn)

		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF Regex Pattern Set: %s", err)
			}
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF Regex Pattern Set: %s", d.Id())
		return conn.DeleteRegexPatternSet(ctx, req)
	})
	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regex Pattern Set: %s", err)
	}

	return diags
}

func updateRegexPatternSetPatternStrings(ctx context.Context, id string, oldPatterns, newPatterns []interface{}, conn *waf.Client) error {
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(id),
			Updates:           DiffRegexPatternSetPatternStrings(oldPatterns, newPatterns),
		}

		return conn.UpdateRegexPatternSet(ctx, req)
	})
	return err
}
