// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafv2_regex_pattern_set", name="Regex Pattern Set")
// @Tags(identifierAttribute="arn")
func resourceRegexPatternSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRegexPatternSetCreate,
		ReadWithoutTimeout:   resourceRegexPatternSetRead,
		UpdateWithoutTimeout: resourceRegexPatternSetUpdate,
		DeleteWithoutTimeout: resourceRegexPatternSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set(names.AttrName, name)
				d.Set(names.AttrScope, scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 256),
				},
				"lock_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"regular_expression": {
					Type:     schema.TypeSet,
					Optional: true,
					MaxItems: 10,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"regex_string": {
								Type:     schema.TypeString,
								Required: true,
								ValidateFunc: validation.All(
									validation.StringLenBetween(1, 200),
									validation.StringIsValidRegExp,
								),
							},
						},
					},
				},
				names.AttrScope: {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.Scope](),
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRegexPatternSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &wafv2.CreateRegexPatternSetInput{
		Name:                  aws.String(name),
		RegularExpressionList: []awstypes.Regex{},
		Scope:                 awstypes.Scope(d.Get(names.AttrScope).(string)),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("regular_expression"); ok && v.(*schema.Set).Len() > 0 {
		input.RegularExpressionList = expandRegexPatternSet(v.(*schema.Set).List())
	}

	output, err := conn.CreateRegexPatternSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAFv2 RegexPatternSet (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Summary.Id))

	return append(diags, resourceRegexPatternSetRead(ctx, d, meta)...)
}

func resourceRegexPatternSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	output, err := findRegexPatternSetByThreePartKey(ctx, conn, d.Id(), d.Get(names.AttrName).(string), d.Get(names.AttrScope).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 RegexPatternSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 RegexPatternSet (%s): %s", d.Id(), err)
	}

	regexPatternSet := output.RegexPatternSet
	arn := aws.ToString(regexPatternSet.ARN)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, regexPatternSet.Description)
	d.Set("lock_token", output.LockToken)
	d.Set(names.AttrName, regexPatternSet.Name)
	if err := d.Set("regular_expression", flattenRegexPatternSet(regexPatternSet.RegularExpressionList)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting regular_expression: %s", err)
	}

	return diags
}

func resourceRegexPatternSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &wafv2.UpdateRegexPatternSetInput{
			Id:                    aws.String(d.Id()),
			LockToken:             aws.String(d.Get("lock_token").(string)),
			Name:                  aws.String(d.Get(names.AttrName).(string)),
			RegularExpressionList: []awstypes.Regex{},
			Scope:                 awstypes.Scope(d.Get(names.AttrScope).(string)),
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("regular_expression"); ok && v.(*schema.Set).Len() > 0 {
			input.RegularExpressionList = expandRegexPatternSet(v.(*schema.Set).List())
		}

		log.Printf("[INFO] Updating WAFv2 RegexPatternSet: %s", d.Id())
		_, err := conn.UpdateRegexPatternSet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAFv2 RegexPatternSet (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRegexPatternSetRead(ctx, d, meta)...)
}

func resourceRegexPatternSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	input := &wafv2.DeleteRegexPatternSetInput{
		Id:        aws.String(d.Id()),
		LockToken: aws.String(d.Get("lock_token").(string)),
		Name:      aws.String(d.Get(names.AttrName).(string)),
		Scope:     awstypes.Scope(d.Get(names.AttrScope).(string)),
	}

	log.Printf("[INFO] Deleting WAFv2 RegexPatternSet: %s", d.Id())
	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*awstypes.WAFAssociatedItemException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteRegexPatternSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAFv2 RegexPatternSet (%s): %s", d.Id(), err)
	}

	return diags
}

func findRegexPatternSetByThreePartKey(ctx context.Context, conn *wafv2.Client, id, name, scope string) (*wafv2.GetRegexPatternSetOutput, error) {
	input := &wafv2.GetRegexPatternSetInput{
		Id:    aws.String(id),
		Name:  aws.String(name),
		Scope: awstypes.Scope(scope),
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

	return output, nil
}
