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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_regex_pattern_set", name="Regex Pattern Set")
func resourceRegexPatternSet() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	outputRaw, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.CreateRegexPatternSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateRegexPatternSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Regex Pattern Set (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*wafregional.CreateRegexPatternSetOutput).RegexPatternSet.RegexPatternSetId))

	return append(diags, resourceRegexPatternSetUpdate(ctx, d, meta)...)
}

func resourceRegexPatternSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	regexPatternSet, err := findRegexPatternSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regional Regex Pattern Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Regional Regex Pattern Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, regexPatternSet.Name)
	d.Set("regex_pattern_strings", regexPatternSet.RegexPatternStrings)

	return diags
}

func resourceRegexPatternSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("regex_pattern_strings") {
		o, n := d.GetChange("regex_pattern_strings")
		oldPatterns, newPatterns := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateRegexPatternSet(ctx, conn, region, d.Id(), oldPatterns, newPatterns); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceRegexPatternSetRead(ctx, d, meta)...)
}

func resourceRegexPatternSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldPatterns := d.Get("regex_pattern_strings").(*schema.Set).List(); len(oldPatterns) > 0 {
		var newPatterns []interface{}
		if err := updateRegexPatternSet(ctx, conn, region, d.Id(), oldPatterns, newPatterns); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Regex Pattern Set: %s", d.Id())
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.DeleteRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(d.Id()),
		}

		return conn.DeleteRegexPatternSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Regex Pattern Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findRegexPatternSetByID(ctx context.Context, conn *wafregional.Client, id string) (*awstypes.RegexPatternSet, error) {
	input := &wafregional.GetRegexPatternSetInput{
		RegexPatternSetId: aws.String(id),
	}

	output, err := conn.GetRegexPatternSet(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RegexPatternSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RegexPatternSet, nil
}

func updateRegexPatternSet(ctx context.Context, conn *wafregional.Client, region, regexPatternSetID string, oldPatterns, newPatterns []interface{}) error {
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.UpdateRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(regexPatternSetID),
			Updates:           diffRegexPatternSetPatternStrings(oldPatterns, newPatterns),
		}

		return conn.UpdateRegexPatternSet(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF Regional Regex Pattern Set (%s): %w", regexPatternSetID, err)
	}

	return nil
}

func diffRegexPatternSetPatternStrings(oldPatterns, newPatterns []interface{}) []awstypes.RegexPatternSetUpdate {
	updates := make([]awstypes.RegexPatternSetUpdate, 0)

	for _, op := range oldPatterns {
		if idx := tfslices.IndexOf(newPatterns, op.(string)); idx > -1 {
			newPatterns = append(newPatterns[:idx], newPatterns[idx+1:]...)
			continue
		}

		updates = append(updates, awstypes.RegexPatternSetUpdate{
			Action:             awstypes.ChangeActionDelete,
			RegexPatternString: aws.String(op.(string)),
		})
	}

	for _, np := range newPatterns {
		updates = append(updates, awstypes.RegexPatternSetUpdate{
			Action:             awstypes.ChangeActionInsert,
			RegexPatternString: aws.String(np.(string)),
		})
	}
	return updates
}
