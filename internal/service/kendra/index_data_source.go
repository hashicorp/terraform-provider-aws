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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_kendra_index")
func DataSourceIndex() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceIndexRead,
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"capacity_units": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"query_capacity_units": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"storage_capacity_units": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"document_metadata_configuration_updates": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Computed: true,
						},
						"relevance": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDuration: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"freshness": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"importance": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"rank_order": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"values_importance_map": {
										Type:     schema.TypeMap,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeInt},
									},
								},
							},
						},
						"search": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"displayable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"facetable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"searchable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
									"sortable": {
										Type:     schema.TypeBool,
										Computed: true,
									},
								},
							},
						},
						names.AttrType: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"edition": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrID: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			"index_statistics": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"faq_statistics": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"indexed_question_answers_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
						"text_document_statistics": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"indexed_text_bytes": {
										Type:     schema.TypeInt,
										Computed: true,
									},
									"indexed_text_documents_count": {
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrRoleARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"server_side_encryption_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKeyID: {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_context_policy": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"user_group_resolution_configuration": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"user_group_resolution_mode": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"user_token_configurations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"json_token_type_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group_attribute_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_name_attribute_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"jwt_token_type_configuration": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"claim_regex": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"group_attribute_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrIssuer: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"key_location": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"secrets_manager_arn": {
										Type:     schema.TypeString,
										Computed: true,
									},
									names.AttrURL: {
										Type:     schema.TypeString,
										Computed: true,
									},
									"user_name_attribute_field": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTags: tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceIndexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	id := d.Get(names.AttrID).(string)

	resp, err := findIndexByID(ctx, conn, id)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Kendra Index (%s): %s", id, err)
	}

	if resp == nil {
		return sdkdiag.AppendErrorf(diags, "getting Kendra Index (%s): empty response", id)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s", id),
	}.String()

	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set(names.AttrDescription, resp.Description)
	d.Set("edition", resp.Edition)
	d.Set("error_message", resp.ErrorMessage)
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrRoleARN, resp.RoleArn)
	d.Set(names.AttrStatus, resp.Status)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))
	d.Set("user_context_policy", resp.UserContextPolicy)

	if err := d.Set("capacity_units", flattenCapacityUnits(resp.CapacityUnits)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("document_metadata_configuration_updates", flattenDocumentMetadataConfigurations(resp.DocumentMetadataConfigurations)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("index_statistics", flattenIndexStatistics(resp.IndexStatistics)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("server_side_encryption_configuration", flattenServerSideEncryptionConfiguration(resp.ServerSideEncryptionConfiguration)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("user_group_resolution_configuration", flattenUserGroupResolutionConfiguration(resp.UserGroupResolutionConfiguration)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("user_token_configurations", flattenUserTokenConfigurations(resp.UserTokenConfigurations)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	tags, err := listTags(ctx, conn, arn)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for resource (%s): %s", arn, err)
	}
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if err := d.Set(names.AttrTags, tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	d.SetId(id)

	return diags
}
