// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package accessanalyzer

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer"
	"github.com/aws/aws-sdk-go-v2/service/accessanalyzer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2/types/nullable"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_accessanalyzer_archive_rule")
func resourceArchiveRule() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceArchiveRuleCreate,
		ReadWithoutTimeout:   resourceArchiveRuleRead,
		UpdateWithoutTimeout: resourceArchiveRuleUpdate,
		DeleteWithoutTimeout: resourceArchiveRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"analyzer_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			names.AttrFilter: {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"criteria": {
							Type:     schema.TypeString,
							Required: true,
						},
						"contains": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"eq": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"exists": {
							Type:         nullable.TypeNullableBool,
							Optional:     true,
							Computed:     true,
							ValidateFunc: nullable.ValidateTypeStringNullableBool,
						},
						"neq": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"rule_name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
		},
	}
}

func resourceArchiveRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	analyzerName := d.Get("analyzer_name").(string)
	ruleName := d.Get("rule_name").(string)
	id := archiveRuleCreateResourceID(analyzerName, ruleName)
	input := &accessanalyzer.CreateArchiveRuleInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(sdkid.UniqueId()),
		RuleName:     aws.String(ruleName),
	}

	if v, ok := d.GetOk(names.AttrFilter); ok {
		input.Filter = expandFilter(v.(*schema.Set))
	}

	_, err := conn.CreateArchiveRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating IAM Access Analyzer Archive Rule (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceArchiveRuleRead(ctx, d, meta)...)
}

func resourceArchiveRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	analyzerName, ruleName, err := archiveRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	archiveRule, err := findArchiveRuleByTwoPartKey(ctx, conn, analyzerName, ruleName)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] IAM Access Analyzer Archive Rule (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading IAM Access Analyzer Archive Rule (%s): %s", d.Id(), err)
	}

	d.Set("analyzer_name", analyzerName)
	d.Set(names.AttrFilter, flattenFilter(archiveRule.Filter))
	d.Set("rule_name", archiveRule.RuleName)

	return diags
}

func resourceArchiveRuleUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	analyzerName, ruleName, err := archiveRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &accessanalyzer.UpdateArchiveRuleInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(sdkid.UniqueId()),
		RuleName:     aws.String(ruleName),
	}

	if d.HasChanges(names.AttrFilter) {
		input.Filter = expandFilter(d.Get(names.AttrFilter).(*schema.Set))
	}

	_, err = conn.UpdateArchiveRule(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating AWS IAM Access Analyzer Archive Rule (%s): %s", d.Id(), err)
	}

	return append(diags, resourceArchiveRuleRead(ctx, d, meta)...)
}

func resourceArchiveRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AccessAnalyzerClient(ctx)

	analyzerName, ruleName, err := archiveRuleParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[INFO] Deleting IAM Access Analyzer Archive Rule: %s", d.Id())
	_, err = conn.DeleteArchiveRule(ctx, &accessanalyzer.DeleteArchiveRuleInput{
		AnalyzerName: aws.String(analyzerName),
		ClientToken:  aws.String(sdkid.UniqueId()),
		RuleName:     aws.String(ruleName),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting IAM Access Analyzer Archive Rule (%s): %s", d.Id(), err)
	}

	return diags
}

func findArchiveRuleByTwoPartKey(ctx context.Context, conn *accessanalyzer.Client, analyzerName, ruleName string) (*types.ArchiveRuleSummary, error) {
	input := &accessanalyzer.GetArchiveRuleInput{
		AnalyzerName: aws.String(analyzerName),
		RuleName:     aws.String(ruleName),
	}

	output, err := conn.GetArchiveRule(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ArchiveRule == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ArchiveRule, nil
}

func flattenFilter(filter map[string]types.Criterion) []interface{} {
	if filter == nil {
		return nil
	}

	l := make([]interface{}, 0)

	for key, value := range filter {
		val := make(map[string]interface{})
		val["criteria"] = key
		val["contains"] = value.Contains
		val["eq"] = value.Eq

		if value.Exists != nil {
			val["exists"] = strconv.FormatBool(aws.ToBool(value.Exists))
		}

		val["neq"] = value.Neq

		l = append(l, val)
	}

	return l
}

func expandFilter(l *schema.Set) map[string]types.Criterion {
	if len(l.List()) == 0 || l.List()[0] == nil {
		return nil
	}

	a := make(map[string]types.Criterion)

	for _, value := range l.List() {
		c := types.Criterion{}
		if v, ok := value.(map[string]interface{})["contains"]; ok {
			if len(v.([]interface{})) > 0 {
				c.Contains = flex.ExpandStringValueList(v.([]interface{}))
			}
		}
		if v, ok := value.(map[string]interface{})["eq"]; ok {
			if len(v.([]interface{})) > 0 {
				c.Eq = flex.ExpandStringValueList(v.([]interface{}))
			}
		}
		if v, ok := value.(map[string]interface{})["neq"]; ok {
			if len(v.([]interface{})) > 0 {
				c.Neq = flex.ExpandStringValueList(v.([]interface{}))
			}
		}
		if v, ok := value.(map[string]interface{})["exists"]; ok {
			if val, null, _ := nullable.Bool(v.(string)).ValueBool(); !null {
				c.Exists = aws.Bool(val)
			}
		}

		a[value.(map[string]interface{})["criteria"].(string)] = c
	}

	return a
}

const archiveRuleResourceIDSeparator = "/"

func archiveRuleCreateResourceID(analyzerName, ruleName string) string {
	parts := []string{analyzerName, ruleName}
	id := strings.Join(parts, archiveRuleResourceIDSeparator)

	return id
}

func archiveRuleParseResourceID(id string) (string, string, error) {
	parts := strings.Split(id, archiveRuleResourceIDSeparator)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected AnalyzerName%[2]sRuleName", id, archiveRuleResourceIDSeparator)
	}

	return parts[0], parts[1], nil
}
