// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra

import (
	"context"
	"fmt"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKDataSource("aws_kendra_experience")
func DataSourceExperience() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceExperienceRead,
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"content_source_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"data_source_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"direct_put_content": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"faq_ids": {
										Type:     schema.TypeSet,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
								},
							},
						},
						"user_identity_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"identity_attribute_name": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoints": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"endpoint": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"endpoint_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"experience_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 36),
					validation.StringMatch(
						regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z_-]*`),
						"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens.",
					),
				),
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceExperienceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	experienceID := d.Get("experience_id").(string)
	indexID := d.Get("index_id").(string)

	resp, err := FindExperienceByID(ctx, conn, experienceID, indexID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Kendra Experience (%s): %s", experienceID, err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/experience/%s", indexID, experienceID),
	}.String()
	d.Set("arn", arn)
	d.Set("created_at", aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set("description", resp.Description)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("experience_id", resp.Id)
	d.Set("index_id", resp.IndexId)
	d.Set("name", resp.Name)
	d.Set("role_arn", resp.RoleArn)
	d.Set("status", resp.Status)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))

	if err := d.Set("configuration", flattenConfiguration(resp.Configuration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting configuration argument: %s", err)
	}

	if err := d.Set("endpoints", flattenEndpoints(resp.Endpoints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoints argument: %s", err)
	}

	d.SetId(fmt.Sprintf("%s/%s", experienceID, indexID))

	return diags
}
