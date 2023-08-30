// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	regexPatternSetDeleteTimeout = 5 * time.Minute
)

// @SDKResource("aws_wafv2_regex_pattern_set", name="Regex Pattern Set")
// @Tags(identifierAttribute="arn")
func ResourceRegexPatternSet() *schema.Resource {
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
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"description": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 256),
				},
				"lock_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
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
				"scope": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice(wafv2.Scope_Values(), false),
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceRegexPatternSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	name := d.Get("name").(string)
	input := &wafv2.CreateRegexPatternSetInput{
		Name:                  aws.String(name),
		RegularExpressionList: []*wafv2.Regex{},
		Scope:                 aws.String(d.Get("scope").(string)),
		Tags:                  getTagsIn(ctx),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("regular_expression"); ok && v.(*schema.Set).Len() > 0 {
		input.RegularExpressionList = expandRegexPatternSet(v.(*schema.Set).List())
	}

	output, err := conn.CreateRegexPatternSetWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating WAFv2 RegexPatternSet (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Summary.Id))

	return resourceRegexPatternSetRead(ctx, d, meta)
}

func resourceRegexPatternSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	output, err := FindRegexPatternSetByThreePartKey(ctx, conn, d.Id(), d.Get("name").(string), d.Get("scope").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 RegexPatternSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading WAFv2 RegexPatternSet (%s): %s", d.Id(), err)
	}

	regexPatternSet := output.RegexPatternSet
	arn := aws.StringValue(regexPatternSet.ARN)
	d.Set("arn", arn)
	d.Set("description", regexPatternSet.Description)
	d.Set("lock_token", output.LockToken)
	d.Set("name", regexPatternSet.Name)
	if err := d.Set("regular_expression", flattenRegexPatternSet(regexPatternSet.RegularExpressionList)); err != nil {
		return diag.Errorf("setting regular_expression: %s", err)
	}

	return nil
}

func resourceRegexPatternSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &wafv2.UpdateRegexPatternSetInput{
			Id:                    aws.String(d.Id()),
			LockToken:             aws.String(d.Get("lock_token").(string)),
			Name:                  aws.String(d.Get("name").(string)),
			RegularExpressionList: []*wafv2.Regex{},
			Scope:                 aws.String(d.Get("scope").(string)),
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("regular_expression"); ok && v.(*schema.Set).Len() > 0 {
			input.RegularExpressionList = expandRegexPatternSet(v.(*schema.Set).List())
		}

		log.Printf("[INFO] Updating WAFv2 RegexPatternSet: %s", input)
		_, err := conn.UpdateRegexPatternSetWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating WAFv2 RegexPatternSet (%s): %s", d.Id(), err)
		}
	}

	return resourceRegexPatternSetRead(ctx, d, meta)
}

func resourceRegexPatternSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	input := &wafv2.DeleteRegexPatternSetInput{
		Id:        aws.String(d.Id()),
		LockToken: aws.String(d.Get("lock_token").(string)),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
	}

	log.Printf("[INFO] Deleting WAFv2 RegexPatternSet: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, regexPatternSetDeleteTimeout, func() (interface{}, error) {
		return conn.DeleteRegexPatternSetWithContext(ctx, input)
	}, wafv2.ErrCodeWAFAssociatedItemException)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting WAFv2 RegexPatternSet (%s): %s", d.Id(), err)
	}

	return nil
}

func FindRegexPatternSetByThreePartKey(ctx context.Context, conn *wafv2.WAFV2, id, name, scope string) (*wafv2.GetRegexPatternSetOutput, error) {
	input := &wafv2.GetRegexPatternSetInput{
		Id:    aws.String(id),
		Name:  aws.String(name),
		Scope: aws.String(scope),
	}

	output, err := conn.GetRegexPatternSetWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
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
