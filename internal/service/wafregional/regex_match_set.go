// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_regex_match_set", name="Regex Match Set")
func resourceRegexMatchSet() *schema.Resource {
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
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	outputRaw, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.CreateRegexMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateRegexMatchSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Regex Match Set (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*wafregional.CreateRegexMatchSetOutput).RegexMatchSet.RegexMatchSetId))

	return append(diags, resourceRegexMatchSetUpdate(ctx, d, meta)...)
}

func resourceRegexMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	set, err := findRegexMatchSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regional Regex Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return diag.Errorf("reading WAF Regional Regex Match Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, set.Name)
	d.Set("regex_match_tuple", FlattenRegexMatchTuples(set.RegexMatchTuples))

	return diags
}

func resourceRegexMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("regex_match_tuple") {
		o, n := d.GetChange("regex_match_tuple")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()

		if err := updateRegexMatchSetResourceWR(ctx, conn, region, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Regex Match Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRegexMatchSetRead(ctx, d, meta)...)
}

func resourceRegexMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldT := d.Get("regex_match_tuple").(*schema.Set).List(); len(oldT) > 0 {
		var newT []interface{}
		err := updateRegexMatchSetResourceWR(ctx, conn, region, d.Id(), oldT, newT)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF Regional Regex Match Set (%s): %s", d.Id(), err)
			}
		}
	}

	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.DeleteRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteRegexMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Regex Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func findRegexMatchSetByID(ctx context.Context, conn *wafregional.Client, id string) (*awstypes.RegexMatchSet, error) {
	input := &wafregional.GetRegexMatchSetInput{
		RegexMatchSetId: aws.String(id),
	}

	output, err := conn.GetRegexMatchSet(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RegexMatchSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RegexMatchSet, nil
}

func updateRegexMatchSetResourceWR(ctx context.Context, conn *wafregional.Client, region, regexMatchSetID string, oldT, newT []interface{}) error {
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.UpdateRegexMatchSetInput{
			ChangeToken:     token,
			RegexMatchSetId: aws.String(regexMatchSetID),
			Updates:         DiffRegexMatchSetTuples(oldT, newT),
		}

		return conn.UpdateRegexMatchSet(ctx, input)
	})

	return err
}
