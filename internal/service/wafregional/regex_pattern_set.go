// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
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
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	outputRaw, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateRegexPatternSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateRegexPatternSetWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Regex Pattern Set (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*waf.CreateRegexPatternSetOutput).RegexPatternSet.RegexPatternSetId))

	return append(diags, resourceRegexPatternSetUpdate(ctx, d, meta)...)
}

func resourceRegexPatternSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	set, err := findRegexPatternSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regional Regex Pattern Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return diag.Errorf("reading WAF Regional Regex Pattern Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, set.Name)
	d.Set("regex_pattern_strings", aws.StringValueSlice(set.RegexPatternStrings))

	return diags
}

func resourceRegexPatternSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("regex_pattern_strings") {
		o, n := d.GetChange("regex_pattern_strings")
		oldPatterns, newPatterns := o.(*schema.Set).List(), n.(*schema.Set).List()

		if err := updateRegexPatternSetPatternStringsWR(ctx, conn, region, d.Id(), oldPatterns, newPatterns); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Regex Pattern Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRegexPatternSetRead(ctx, d, meta)...)
}

func resourceRegexPatternSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldPatterns := d.Get("regex_pattern_strings").(*schema.Set).List(); len(oldPatterns) > 0 {
		var newPatterns []interface{}

		err := updateRegexPatternSetPatternStringsWR(ctx, conn, region, d.Id(), oldPatterns, newPatterns)

		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentContainerException, wafregional.ErrCodeWAFNonexistentItemException) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Regex Pattern Set (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Regex Pattern Set: %s", d.Id())
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(d.Id()),
		}

		return conn.DeleteRegexPatternSetWithContext(ctx, input)
	})

	if tfawserr.ErrCodeEquals(err, waf.ErrCodeNonexistentItemException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Regex Pattern Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findRegexPatternSetByID(ctx context.Context, conn *wafregional.WAFRegional, id string) (*waf.RegexPatternSet, error) {
	input := &waf.GetRegexPatternSetInput{
		RegexPatternSetId: aws.String(id),
	}

	output, err := conn.GetRegexPatternSetWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
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

func updateRegexPatternSetPatternStringsWR(ctx context.Context, conn *wafregional.WAFRegional, region, regexPatternSetID string, oldPatterns, newPatterns []interface{}) error {
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateRegexPatternSetInput{
			ChangeToken:       token,
			RegexPatternSetId: aws.String(regexPatternSetID),
			Updates:           tfwaf.DiffRegexPatternSetPatternStrings(oldPatterns, newPatterns),
		}

		return conn.UpdateRegexPatternSetWithContext(ctx, input)
	})

	return err
}
