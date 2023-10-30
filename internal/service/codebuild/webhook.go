// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package codebuild

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codebuild"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_codebuild_webhook")
func ResourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWebhookCreate,
		ReadWithoutTimeout:   resourceWebhookRead,
		DeleteWithoutTimeout: resourceWebhookDelete,
		UpdateWithoutTimeout: resourceWebhookUpdate,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"build_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(codebuild.WebhookBuildType_Values(), false),
			},
			"branch_filter": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"filter_group"},
			},
			"filter_group": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(codebuild.WebhookFilterType_Values(), false),
									},
									"exclude_matched_pattern": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
									},
									"pattern": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
				Set:           resourceWebhookFilterHash,
				ConflictsWith: []string{"branch_filter"},
			},
			"payload_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
			"url": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	input := &codebuild.CreateWebhookInput{
		ProjectName:  aws.String(d.Get("project_name").(string)),
		FilterGroups: expandWebhookFilterGroups(d),
	}

	if v, ok := d.GetOk("build_type"); ok {
		input.BuildType = aws.String(v.(string))
	}

	// The CodeBuild API requires this to be non-empty if defined
	if v, ok := d.GetOk("branch_filter"); ok {
		input.BranchFilter = aws.String(v.(string))
	}

	resp, err := conn.CreateWebhookWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CodeBuild Webhook: %s", err)
	}

	// Secret is only returned on create, so capture it at the start
	d.Set("secret", resp.Webhook.Secret)
	d.SetId(d.Get("project_name").(string))

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func expandWebhookFilterGroups(d *schema.ResourceData) [][]*codebuild.WebhookFilter {
	configs := d.Get("filter_group").(*schema.Set).List()

	webhookFilters := make([][]*codebuild.WebhookFilter, 0)

	if len(configs) == 0 {
		return nil
	}

	for _, config := range configs {
		filters := expandWebhookFilterData(config.(map[string]interface{}))
		webhookFilters = append(webhookFilters, filters)
	}

	return webhookFilters
}

func expandWebhookFilterData(data map[string]interface{}) []*codebuild.WebhookFilter {
	filters := make([]*codebuild.WebhookFilter, 0)

	filterConfigs := data["filter"].([]interface{})

	for i, filterConfig := range filterConfigs {
		filter := filterConfig.(map[string]interface{})
		filters = append(filters, &codebuild.WebhookFilter{
			Type:                  aws.String(filter["type"].(string)),
			ExcludeMatchedPattern: aws.Bool(filter["exclude_matched_pattern"].(bool)),
		})
		if v := filter["pattern"]; v != nil {
			filters[i].Pattern = aws.String(v.(string))
		}
	}

	return filters
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	resp, err := conn.BatchGetProjectsWithContext(ctx, &codebuild.BatchGetProjectsInput{
		Names: []*string{
			aws.String(d.Id()),
		},
	})

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, codebuild.ErrCodeResourceNotFoundException) {
		create.LogNotFoundRemoveState(names.CodeBuild, create.ErrActionReading, ResNameWebhook, d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return create.DiagError(names.CodeBuild, create.ErrActionReading, ResNameWebhook, d.Id(), err)
	}

	if d.IsNewResource() && len(resp.Projects) == 0 {
		return create.DiagError(names.CodeBuild, create.ErrActionReading, ResNameWebhook, d.Id(), errors.New("no project found after create"))
	}

	if !d.IsNewResource() && len(resp.Projects) == 0 {
		create.LogNotFoundRemoveState(names.CodeBuild, create.ErrActionReading, ResNameWebhook, d.Id())
		d.SetId("")
		return diags
	}

	project := resp.Projects[0]

	if d.IsNewResource() && project.Webhook == nil {
		return create.DiagError(names.CodeBuild, create.ErrActionReading, ResNameWebhook, d.Id(), errors.New("no webhook after creation"))
	}

	if !d.IsNewResource() && project.Webhook == nil {
		create.LogNotFoundRemoveState(names.CodeBuild, create.ErrActionReading, ResNameWebhook, d.Id())
		d.SetId("")
		return diags
	}

	d.Set("build_type", project.Webhook.BuildType)
	d.Set("branch_filter", project.Webhook.BranchFilter)
	d.Set("filter_group", flattenWebhookFilterGroups(project.Webhook.FilterGroups))
	d.Set("payload_url", project.Webhook.PayloadUrl)
	d.Set("project_name", project.Name)
	d.Set("url", project.Webhook.Url)
	// The secret is never returned after creation, so don't set it here

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	var err error
	filterGroups := expandWebhookFilterGroups(d)

	var buildType *string
	if v, ok := d.GetOk("build_type"); ok {
		buildType = aws.String(v.(string))
	}

	if len(filterGroups) >= 1 {
		_, err = conn.UpdateWebhookWithContext(ctx, &codebuild.UpdateWebhookInput{
			ProjectName:  aws.String(d.Id()),
			BuildType:    buildType,
			FilterGroups: filterGroups,
			RotateSecret: aws.Bool(false),
		})
	} else {
		_, err = conn.UpdateWebhookWithContext(ctx, &codebuild.UpdateWebhookInput{
			ProjectName:  aws.String(d.Id()),
			BuildType:    buildType,
			BranchFilter: aws.String(d.Get("branch_filter").(string)),
			RotateSecret: aws.Bool(false),
		})
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating CodeBuild Webhook (%s): %s", d.Id(), err)
	}

	return append(diags, resourceWebhookRead(ctx, d, meta)...)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CodeBuildConn(ctx)

	_, err := conn.DeleteWebhookWithContext(ctx, &codebuild.DeleteWebhookInput{
		ProjectName: aws.String(d.Id()),
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, codebuild.ErrCodeResourceNotFoundException) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting CodeBuild Webhook (%s): %s", d.Id(), err)
	}

	return diags
}

func flattenWebhookFilterGroups(filterList [][]*codebuild.WebhookFilter) *schema.Set {
	filterSet := schema.Set{
		F: resourceWebhookFilterHash,
	}

	for _, filters := range filterList {
		filterSet.Add(flattenWebhookFilterData(filters))
	}
	return &filterSet
}

func resourceWebhookFilterHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	for _, g := range m {
		for _, f := range g.([]interface{}) {
			r := f.(map[string]interface{})
			buf.WriteString(fmt.Sprintf("%s-", r["type"].(string)))
			buf.WriteString(fmt.Sprintf("%s-", r["pattern"].(string)))
			buf.WriteString(fmt.Sprintf("%q", r["exclude_matched_pattern"]))
		}
	}

	return create.StringHashcode(buf.String())
}

func flattenWebhookFilterData(filters []*codebuild.WebhookFilter) map[string]interface{} {
	values := map[string]interface{}{}
	ff := make([]interface{}, 0)

	for _, f := range filters {
		ff = append(ff, map[string]interface{}{
			"type":                    *f.Type,
			"pattern":                 *f.Pattern,
			"exclude_matched_pattern": *f.ExcludeMatchedPattern,
		})
	}

	values["filter"] = ff

	return values
}
