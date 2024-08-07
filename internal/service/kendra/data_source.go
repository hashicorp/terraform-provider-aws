// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kendra

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/kendra"
	"github.com/aws/aws-sdk-go-v2/service/kendra/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ISO8601UTC = "2006-01-02T15:04:05-07:00"

	// validationExceptionMessageDataSourceSecrets describes the error returned when the IAM permission has not yet propagated
	validationExceptionMessageDataSourceSecrets = "Secrets Manager throws the exception"
)

// @SDKResource("aws_kendra_data_source", name="Data Source")
// @Tags(identifierAttribute="arn")
func ResourceDataSource() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDataSourceCreate,
		ReadWithoutTimeout:   resourceDataSourceRead,
		UpdateWithoutTimeout: resourceDataSourceUpdate,
		DeleteWithoutTimeout: resourceDataSourceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
		CustomizeDiff: customdiff.All(
			func(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
				if configuration, dataSourcetype := diff.Get(names.AttrConfiguration).([]interface{}), diff.Get(names.AttrType).(string); len(configuration) > 0 && dataSourcetype == string(types.DataSourceTypeCustom) {
					return fmt.Errorf("configuration must not be set when type is %s", string(types.DataSourceTypeCustom))
				}

				if roleArn, dataSourcetype := diff.Get(names.AttrRoleARN).(string), diff.Get(names.AttrType).(string); roleArn != "" && dataSourcetype == string(types.DataSourceTypeCustom) {
					return fmt.Errorf("role_arn must not be set when type is %s", string(types.DataSourceTypeCustom))
				}

				if schedule, dataSourcetype := diff.Get(names.AttrSchedule).(string), diff.Get(names.AttrType).(string); schedule != "" && dataSourcetype == string(types.DataSourceTypeCustom) {
					return fmt.Errorf("schedule must not be set when type is %s", string(types.DataSourceTypeCustom))
				}

				return nil
			},
			verify.SetTagsDiff,
		),
		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_control_list_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"key_path": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},
									names.AttrBucketName: {
										Type:     schema.TypeString,
										Required: true,
										ValidateFunc: validation.All(
											validation.StringLenBetween(3, 63),
											validation.StringMatch(
												regexache.MustCompile(`[0-9a-z][0-9a-z.-]{1,61}[0-9a-z]`),
												"Must be a valid bucket name",
											),
										),
									},
									"documents_metadata_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"s3_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},
											},
										},
									},
									"exclusion_patterns": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 0,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 150),
										},
									},
									"inclusion_patterns": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 0,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 150),
										},
									},
									"inclusion_prefixes": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 0,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 150),
										},
									},
								},
							},
						},
						"web_crawler_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"authentication_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"basic_authentication": {
													Type:     schema.TypeSet,
													Optional: true,
													MinItems: 0,
													MaxItems: 10,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"credentials": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: verify.ValidARN,
															},
															"host": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 253),
															},
															names.AttrPort: {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(1, 65535),
															},
														},
													},
												},
											},
										},
									},
									"crawl_depth": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      2,
										ValidateFunc: validation.IntBetween(0, 10),
									},
									"max_content_size_per_page_in_mega_bytes": {
										Type:     schema.TypeFloat,
										Optional: true,
										// Default:      50,
										ValidateFunc: validation.FloatBetween(0.000001, 50),
									},
									"max_links_per_page": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      100,
										ValidateFunc: validation.IntBetween(1, 1000),
									},
									"max_urls_per_minute_crawl_rate": {
										Type:         schema.TypeInt,
										Optional:     true,
										Default:      300,
										ValidateFunc: validation.IntBetween(1, 300),
									},
									"proxy_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"credentials": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: verify.ValidARN,
												},
												"host": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 253),
												},
												names.AttrPort: {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IntBetween(1, 65535),
												},
											},
										},
									},
									"url_exclusion_patterns": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 0,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 150),
										},
									},
									"url_inclusion_patterns": {
										Type:     schema.TypeSet,
										Optional: true,
										MinItems: 0,
										MaxItems: 100,
										Elem: &schema.Schema{
											Type:         schema.TypeString,
											ValidateFunc: validation.StringLenBetween(1, 150),
										},
									},
									"urls": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"seed_url_configuration": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"seed_urls": {
																Type:     schema.TypeSet,
																Required: true,
																MinItems: 0,
																MaxItems: 100,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																	ValidateFunc: validation.All(
																		validation.StringLenBetween(1, 2048),
																		validation.StringMatch(regexache.MustCompile(`^(https?):\/\/([^\s]*)`), "must provide a valid url"),
																	),
																},
															},
															"web_crawler_mode": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[types.WebCrawlerMode](),
															},
														},
													},
												},
												"site_maps_configuration": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"site_maps": {
																Type:     schema.TypeSet,
																Required: true,
																MinItems: 0,
																MaxItems: 3,
																Elem: &schema.Schema{
																	Type: schema.TypeString,
																	ValidateFunc: validation.All(
																		validation.StringLenBetween(1, 2048),
																		validation.StringMatch(regexache.MustCompile(`^(https?):\/\/([^\s]*)`), "must provide a valid url"),
																	),
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"custom_document_enrichment_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"inline_configurations": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 0,
							MaxItems: 100,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrCondition: func() *schema.Schema {
										schema := documentAttributeConditionSchema()
										return schema
									}(),
									"document_content_deletion": {
										Type:     schema.TypeBool,
										Optional: true,
									},
									names.AttrTarget: {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"target_document_attribute_key": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.All(
														validation.StringLenBetween(1, 200),
														validation.StringMatch(
															regexache.MustCompile(`[0-9A-Za-z_][0-9A-Za-z_-]*`),
															"Starts with an alphanumeric character or underscore. Subsequently, can contain alphanumeric characters, underscores and hyphens.",
														),
													),
												},
												"target_document_attribute_value": func() *schema.Schema {
													schema := documentAttributeValueSchema()
													return schema
												}(),
												"target_document_attribute_value_deletion": {
													Type:     schema.TypeBool,
													Optional: true,
												},
											},
										},
									},
								},
							},
						},
						"post_extraction_hook_configuration": func() *schema.Schema {
							schema := hookConfigurationSchema()
							return schema
						}(),
						"pre_extraction_hook_configuration": func() *schema.Schema {
							schema := hookConfigurationSchema()
							return schema
						}(),
						names.AttrRoleARN: {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"data_source_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1000),
			},
			"error_message": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"index_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringMatch(
					regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z-]{35}`),
					"Starts with an alphanumeric character. Subsequently, can contain alphanumeric characters and hyphens. Fixed length of 36.",
				),
			},
			names.AttrLanguageCode: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 10),
					validation.StringMatch(
						regexache.MustCompile(`[A-Za-z-]*`),
						"Must have alphanumeric characters or hyphens.",
					),
				),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 1000),
					validation.StringMatch(
						regexache.MustCompile(`[0-9A-Za-z][0-9A-Za-z_-]*`),
						"Starts with an alphanumeric character. Subsequently, the name must consist of alphanumerics, hyphens or underscores.",
					),
				),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrSchedule: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[types.DataSourceType](),
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

// post_extraction_hook_configuration and pre_extraction_hook_configuration share the same schema
func hookConfigurationSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"invocation_condition": func() *schema.Schema {
					schema := documentAttributeConditionSchema()
					return schema
				}(),
				"lambda_arn": {
					Type:         schema.TypeString,
					Required:     true,
					ValidateFunc: verify.ValidARN,
				},
				names.AttrS3Bucket: {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(3, 63),
						validation.StringMatch(
							regexache.MustCompile(`[0-9a-z][0-9a-z.-]{1,61}[0-9a-z]`),
							"Must be a valid bucket name",
						),
					),
				},
			},
		},
	}
}

func documentAttributeConditionSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"condition_document_attribute_key": {
					Type:     schema.TypeString,
					Required: true,
					ValidateFunc: validation.All(
						validation.StringLenBetween(1, 200),
						validation.StringMatch(
							regexache.MustCompile(`[0-9A-Za-z_][0-9A-Za-z_-]*`),
							"Starts with an alphanumeric character or underscore. Subsequently, can contain alphanumeric characters, underscores and hyphens.",
						),
					),
				},
				"condition_on_value": func() *schema.Schema {
					schema := documentAttributeValueSchema()
					return schema
				}(),
				"operator": {
					Type:             schema.TypeString,
					Required:         true,
					ValidateDiagFunc: enum.Validate[types.ConditionOperator](),
				},
			},
		},
	}
}

func documentAttributeValueSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"date_value": {
					Type:     schema.TypeString,
					Optional: true,
					// DiffSuppressFunc does not work on attributes that are part of another attribute of TypeSet
					// DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// 	oldTime, err := time.Parse(ISO8601UTC, old)
					// 	if err != nil {
					// 		return false
					// 	}
					// 	newTime, err := time.Parse(ISO8601UTC, new)
					// 	if err != nil {
					// 		return false
					// 	}
					// 	return oldTime.Equal(newTime)
					// },
					// DiffSuppressOnRefresh: true,
				},
				"long_value": {
					Type:     schema.TypeInt,
					Optional: true,
				},
				"string_list_value": {
					Type:     schema.TypeSet,
					Optional: true,
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validation.StringLenBetween(1, 2048),
					},
				},
				"string_value": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 2048),
				},
			},
		},
	}
}

func resourceDataSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &kendra.CreateDataSourceInput{
		ClientToken: aws.String(id.UniqueId()),
		IndexId:     aws.String(d.Get("index_id").(string)),
		Name:        aws.String(name),
		Tags:        getTagsIn(ctx),
		Type:        types.DataSourceType(d.Get(names.AttrType).(string)),
	}

	if v, ok := d.GetOk(names.AttrConfiguration); ok {
		input.Configuration = expandDataSourceConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("custom_document_enrichment_configuration"); ok {
		input.CustomDocumentEnrichmentConfiguration = expandCustomDocumentEnrichmentConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrLanguageCode); ok {
		input.LanguageCode = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrRoleARN); ok {
		input.RoleArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrSchedule); ok {
		input.Schedule = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhen(ctx, propagationTimeout,
		func() (interface{}, error) {
			return conn.CreateDataSource(ctx, input)
		},
		func(err error) (bool, error) {
			var validationException *types.ValidationException

			if errors.As(err, &validationException) && (strings.Contains(validationException.ErrorMessage(), validationExceptionMessage) || strings.Contains(validationException.ErrorMessage(), validationExceptionMessageDataSourceSecrets)) {
				return true, err
			}

			return false, err
		},
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Kendra Data Source (%s): %s", name, err)
	}

	if outputRaw == nil {
		return sdkdiag.AppendErrorf(diags, "creating Kendra Data Source (%s): empty output", name)
	}

	output := outputRaw.(*kendra.CreateDataSourceOutput)

	id := aws.ToString(output.Id)
	indexId := d.Get("index_id").(string)

	d.SetId(fmt.Sprintf("%s/%s", id, indexId))

	if _, err := waitDataSourceCreated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Data Source (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	id, indexId, err := DataSourceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	resp, err := FindDataSourceByID(ctx, conn, id, indexId)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kendra Data Source (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting Kendra Data Source (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "kendra",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("index/%s/data-source/%s", indexId, id),
	}.String()

	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedAt, aws.ToTime(resp.CreatedAt).Format(time.RFC3339))
	d.Set("data_source_id", resp.Id)
	d.Set(names.AttrDescription, resp.Description)
	d.Set("error_message", resp.ErrorMessage)
	d.Set("index_id", resp.IndexId)
	d.Set(names.AttrLanguageCode, resp.LanguageCode)
	d.Set(names.AttrName, resp.Name)
	d.Set(names.AttrRoleARN, resp.RoleArn)
	d.Set(names.AttrSchedule, resp.Schedule)
	d.Set(names.AttrStatus, resp.Status)
	d.Set(names.AttrType, resp.Type)
	d.Set("updated_at", aws.ToTime(resp.UpdatedAt).Format(time.RFC3339))

	if err := d.Set(names.AttrConfiguration, flattenDataSourceConfiguration(resp.Configuration)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := d.Set("custom_document_enrichment_configuration", flattenCustomDocumentEnrichmentConfiguration(resp.CustomDocumentEnrichmentConfiguration)); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func resourceDataSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	if d.HasChanges(names.AttrConfiguration, "custom_document_enrichment_configuration", names.AttrDescription, names.AttrLanguageCode, names.AttrName, names.AttrRoleARN, names.AttrSchedule) {
		id, indexId, err := DataSourceParseResourceID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &kendra.UpdateDataSourceInput{
			Id:      aws.String(id),
			IndexId: aws.String(indexId),
		}

		if d.HasChange(names.AttrConfiguration) {
			input.Configuration = expandDataSourceConfiguration(d.Get(names.AttrConfiguration).([]interface{}))
		}

		if d.HasChange("custom_document_enrichment_configuration") {
			input.CustomDocumentEnrichmentConfiguration = expandCustomDocumentEnrichmentConfiguration(d.Get("custom_document_enrichment_configuration").([]interface{}))
		}

		if d.HasChange(names.AttrDescription) {
			input.Description = aws.String(d.Get(names.AttrDescription).(string))
		}

		if d.HasChange(names.AttrLanguageCode) {
			input.LanguageCode = aws.String(d.Get(names.AttrLanguageCode).(string))
		}

		if d.HasChange(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChange(names.AttrRoleARN) {
			input.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
		}

		if d.HasChange(names.AttrSchedule) {
			input.Schedule = aws.String(d.Get(names.AttrSchedule).(string))
		}

		log.Printf("[DEBUG] Updating Kendra Data Source (%s): %#v", d.Id(), input)

		_, err = tfresource.RetryWhen(ctx, propagationTimeout,
			func() (interface{}, error) {
				return conn.UpdateDataSource(ctx, input)
			},
			func(err error) (bool, error) {
				var validationException *types.ValidationException

				if errors.As(err, &validationException) && strings.Contains(validationException.ErrorMessage(), validationExceptionMessage) {
					return true, err
				}

				return false, err
			},
		)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Kendra Data Source(%s): %s", d.Id(), err)
		}

		if _, err := waitDataSourceUpdated(ctx, conn, id, indexId, d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Kendra Data Source (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDataSourceRead(ctx, d, meta)...)
}

func resourceDataSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KendraClient(ctx)

	log.Printf("[INFO] Deleting Kendra Data Source %s", d.Id())

	id, indexId, err := DataSourceParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = conn.DeleteDataSource(ctx, &kendra.DeleteDataSourceInput{
		Id:      aws.String(id),
		IndexId: aws.String(indexId),
	})

	var resourceNotFoundException *types.ResourceNotFoundException
	if errors.As(err, &resourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Kendra Data Source (%s): %s", d.Id(), err)
	}

	if _, err := waitDataSourceDeleted(ctx, conn, id, indexId, d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Kendra Data Source (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func waitDataSourceCreated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeDataSourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.DataSourceStatusCreating),
		Target:                    enum.Slice(types.DataSourceStatusActive),
		Timeout:                   timeout,
		Refresh:                   statusDataSource(ctx, conn, id, indexId),
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeDataSourceOutput); ok {
		if output.Status == types.DataSourceStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}
		return output, err
	}

	return nil, err
}

func waitDataSourceUpdated(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeDataSourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.DataSourceStatusUpdating),
		Target:                    enum.Slice(types.DataSourceStatusActive),
		Timeout:                   timeout,
		Refresh:                   statusDataSource(ctx, conn, id, indexId),
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeDataSourceOutput); ok {
		if output.Status == types.DataSourceStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))
		}
		return output, err
	}

	return nil, err
}

func waitDataSourceDeleted(ctx context.Context, conn *kendra.Client, id, indexId string, timeout time.Duration) (*kendra.DescribeDataSourceOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.DataSourceStatusDeleting),
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusDataSource(ctx, conn, id, indexId),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kendra.DescribeDataSourceOutput); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.ErrorMessage)))

		return output, err
	}

	return nil, err
}

func statusDataSource(ctx context.Context, conn *kendra.Client, id, indexId string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDataSourceByID(ctx, conn, id, indexId)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func expandDataSourceConfiguration(tfList []interface{}) *types.DataSourceConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.DataSourceConfiguration{}

	if v, ok := tfMap["s3_configuration"].([]interface{}); ok && len(v) > 0 {
		result.S3Configuration = expandS3Configuration(v)
	}

	if v, ok := tfMap["web_crawler_configuration"].([]interface{}); ok && len(v) > 0 {
		result.WebCrawlerConfiguration = expandWebCrawlerConfiguration(v)
	}

	return result
}

// S3 Configuration
func expandS3Configuration(tfList []interface{}) *types.S3DataSourceConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.S3DataSourceConfiguration{
		BucketName: aws.String(tfMap[names.AttrBucketName].(string)),
	}

	if v, ok := tfMap["access_control_list_configuration"].([]interface{}); ok && len(v) > 0 {
		result.AccessControlListConfiguration = expandAccessControlListConfiguration(v)
	}

	if v, ok := tfMap["documents_metadata_configuration"].([]interface{}); ok && len(v) > 0 {
		result.DocumentsMetadataConfiguration = expandDocumentsMetadataConfiguration(v)
	}

	if v, ok := tfMap["exclusion_patterns"]; ok && v.(*schema.Set).Len() >= 0 {
		result.ExclusionPatterns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["inclusion_patterns"]; ok && v.(*schema.Set).Len() >= 0 {
		result.InclusionPatterns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["inclusion_prefixes"]; ok && v.(*schema.Set).Len() >= 0 {
		result.InclusionPrefixes = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	return result
}

func expandAccessControlListConfiguration(tfList []interface{}) *types.AccessControlListConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.AccessControlListConfiguration{}

	if v, ok := tfMap["key_path"].(string); ok && v != "" {
		result.KeyPath = aws.String(v)
	}

	return result
}

func expandDocumentsMetadataConfiguration(tfList []interface{}) *types.DocumentsMetadataConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.DocumentsMetadataConfiguration{}

	if v, ok := tfMap["s3_prefix"].(string); ok && v != "" {
		result.S3Prefix = aws.String(v)
	}

	return result
}

// Web Crawler Configuration
func expandWebCrawlerConfiguration(tfList []interface{}) *types.WebCrawlerConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.WebCrawlerConfiguration{
		Urls: expandURLs(tfMap["urls"].([]interface{})),
	}

	if v, ok := tfMap["authentication_configuration"].([]interface{}); ok && len(v) > 0 {
		result.AuthenticationConfiguration = expandAuthenticationConfiguration(v)
	}

	if v, ok := tfMap["crawl_depth"].(int); ok {
		result.CrawlDepth = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_content_size_per_page_in_mega_bytes"].(float32); ok {
		result.MaxContentSizePerPageInMegaBytes = aws.Float32(v)
	}

	if v, ok := tfMap["max_links_per_page"].(int); ok {
		result.MaxLinksPerPage = aws.Int32(int32(v))
	}

	if v, ok := tfMap["max_urls_per_minute_crawl_rate"].(int); ok {
		result.MaxUrlsPerMinuteCrawlRate = aws.Int32(int32(v))
	}

	if v, ok := tfMap["proxy_configuration"].([]interface{}); ok && len(v) > 0 {
		result.ProxyConfiguration = expandProxyConfiguration(v)
	}

	if v, ok := tfMap["url_exclusion_patterns"]; ok && v.(*schema.Set).Len() >= 0 {
		result.UrlExclusionPatterns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := tfMap["url_inclusion_patterns"]; ok && v.(*schema.Set).Len() >= 0 {
		result.UrlInclusionPatterns = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	return result
}

func expandAuthenticationConfiguration(tfList []interface{}) *types.AuthenticationConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.AuthenticationConfiguration{}

	if v, ok := tfMap["basic_authentication"]; ok && v.(*schema.Set).Len() > 0 {
		result.BasicAuthentication = expandBasicAuthentication(v.(*schema.Set).List())
	}

	return result
}

func expandBasicAuthentication(tfList []interface{}) []types.BasicAuthenticationConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	result := []types.BasicAuthenticationConfiguration{}

	for _, basicAuthenticationConfig := range tfList {
		data := basicAuthenticationConfig.(map[string]interface{})
		basicAuthenticationConfigExpanded := types.BasicAuthenticationConfiguration{
			Credentials: aws.String(data["credentials"].(string)),
			Host:        aws.String(data["host"].(string)),
			Port:        aws.Int32(int32(data[names.AttrPort].(int))),
		}

		result = append(result, basicAuthenticationConfigExpanded)
	}

	return result
}

func expandProxyConfiguration(tfList []interface{}) *types.ProxyConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.ProxyConfiguration{
		Host: aws.String(tfMap["host"].(string)),
		Port: aws.Int32(int32(tfMap[names.AttrPort].(int))),
	}

	if v, ok := tfMap["credentials"].(string); ok && v != "" {
		result.Credentials = aws.String(tfMap["credentials"].(string))
	}

	return result
}

func expandURLs(tfList []interface{}) *types.Urls {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.Urls{}

	if v, ok := tfMap["seed_url_configuration"].([]interface{}); ok && len(v) > 0 {
		result.SeedUrlConfiguration = expandSeedURLConfiguration(v)
	}

	if v, ok := tfMap["site_maps_configuration"].([]interface{}); ok && len(v) > 0 {
		result.SiteMapsConfiguration = expandSiteMapsConfiguration(v)
	}

	return result
}

func expandSeedURLConfiguration(tfList []interface{}) *types.SeedUrlConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.SeedUrlConfiguration{
		SeedUrls: flex.ExpandStringValueSet(tfMap["seed_urls"].(*schema.Set)),
	}

	if v, ok := tfMap["web_crawler_mode"].(string); ok && v != "" {
		result.WebCrawlerMode = types.WebCrawlerMode(tfMap["web_crawler_mode"].(string))
	}

	return result
}

func expandSiteMapsConfiguration(tfList []interface{}) *types.SiteMapsConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.SiteMapsConfiguration{
		SiteMaps: flex.ExpandStringValueSet(tfMap["site_maps"].(*schema.Set)),
	}

	return result
}

// Custom document enrichment configuration
func expandCustomDocumentEnrichmentConfiguration(tfList []interface{}) *types.CustomDocumentEnrichmentConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.CustomDocumentEnrichmentConfiguration{}

	if v, ok := tfMap["inline_configurations"]; ok && v.(*schema.Set).Len() > 0 {
		result.InlineConfigurations = expandInlineCustomDocumentEnrichmentConfiguration(v.(*schema.Set).List())
	}

	if v, ok := tfMap["post_extraction_hook_configuration"].([]interface{}); ok && len(v) > 0 {
		result.PostExtractionHookConfiguration = expandHookConfiguration(v)
	}

	if v, ok := tfMap["pre_extraction_hook_configuration"].([]interface{}); ok && len(v) > 0 {
		result.PreExtractionHookConfiguration = expandHookConfiguration(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		result.RoleArn = aws.String(v)
	}

	return result
}

func expandInlineCustomDocumentEnrichmentConfiguration(tfList []interface{}) []types.InlineCustomDocumentEnrichmentConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	result := []types.InlineCustomDocumentEnrichmentConfiguration{}

	for _, inlineConfig := range tfList {
		data := inlineConfig.(map[string]interface{})
		inlineConfigExpanded := types.InlineCustomDocumentEnrichmentConfiguration{}

		if v, ok := data[names.AttrCondition].([]interface{}); ok && len(v) > 0 {
			inlineConfigExpanded.Condition = expandDocumentAttributeCondition(v)
		}

		if v, ok := data["document_content_deletion"].(bool); ok {
			inlineConfigExpanded.DocumentContentDeletion = v
		}

		if v, ok := data[names.AttrTarget].([]interface{}); ok && len(v) > 0 {
			inlineConfigExpanded.Target = expandDocumentAttributeTarget(v)
		}

		result = append(result, inlineConfigExpanded)
	}

	return result
}

func expandDocumentAttributeTarget(tfList []interface{}) *types.DocumentAttributeTarget {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.DocumentAttributeTarget{}

	if v, ok := tfMap["target_document_attribute_key"].(string); ok && v != "" {
		result.TargetDocumentAttributeKey = aws.String(v)
	}

	if v, ok := tfMap["target_document_attribute_value"].([]interface{}); ok && len(v) > 0 {
		result.TargetDocumentAttributeValue = expandDocumentAttributeValue(v)
	}

	if v, ok := tfMap["target_document_attribute_value_deletion"].(bool); ok {
		result.TargetDocumentAttributeValueDeletion = v
	}

	return result
}

func expandHookConfiguration(tfList []interface{}) *types.HookConfiguration {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.HookConfiguration{
		LambdaArn: aws.String(tfMap["lambda_arn"].(string)),
		S3Bucket:  aws.String(tfMap[names.AttrS3Bucket].(string)),
	}

	if v, ok := tfMap["invocation_condition"].([]interface{}); ok && len(v) > 0 {
		result.InvocationCondition = expandDocumentAttributeCondition(v)
	}

	return result
}

func expandDocumentAttributeCondition(tfList []interface{}) *types.DocumentAttributeCondition {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.DocumentAttributeCondition{
		ConditionDocumentAttributeKey: aws.String(tfMap["condition_document_attribute_key"].(string)),
		Operator:                      types.ConditionOperator(tfMap["operator"].(string)),
	}

	if v, ok := tfMap["condition_on_value"].([]interface{}); ok && len(v) > 0 {
		result.ConditionOnValue = expandDocumentAttributeValue(v)
	}

	return result
}

func expandDocumentAttributeValue(tfList []interface{}) *types.DocumentAttributeValue {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap, ok := tfList[0].(map[string]interface{})
	if !ok {
		return nil
	}

	result := &types.DocumentAttributeValue{}

	// Only one of these values can be set at a time
	if v, ok := tfMap["date_value"].(string); ok && v != "" {
		// A date expressed as an ISO 8601 string.
		timeValue, _ := time.Parse(ISO8601UTC, v)
		result.DateValue = aws.Time(timeValue)
	} else if v, ok := tfMap["string_value"].(string); ok && v != "" {
		result.StringValue = aws.String(v)
	} else if v, ok := tfMap["string_list_value"]; ok && v.(*schema.Set).Len() > 0 {
		result.StringListValue = flex.ExpandStringValueSet(v.(*schema.Set))
	} else if v, ok := tfMap["long_value"]; ok {
		// When no value was passed it was interpreted as a 0 leading to errors if other values like DateValue, StringValue, StringListValue were defined
		// ValidationException: DocumentAttributeValue can only have 1 non-null field, but given value for key <> has too many non-null fields.
		// hence check this as the last else if
		result.LongValue = aws.Int64(int64(v.(int)))
	}

	return result
}

func flattenDataSourceConfiguration(apiObject *types.DataSourceConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.S3Configuration; v != nil {
		m["s3_configuration"] = flattenS3Configuration(v)
	}

	if v := apiObject.WebCrawlerConfiguration; v != nil {
		m["web_crawler_configuration"] = flattenWebCrawlerConfiguration(v)
	}

	return []interface{}{m}
}

// S3 Configuration
func flattenS3Configuration(apiObject *types.S3DataSourceConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		names.AttrBucketName: aws.ToString(apiObject.BucketName),
	}

	if v := apiObject.AccessControlListConfiguration; v != nil {
		m["access_control_list_configuration"] = flattenAccessControlListConfiguration(v)
	}

	if v := apiObject.DocumentsMetadataConfiguration; v != nil {
		m["documents_metadata_configuration"] = flattenDocumentsMetadataConfiguration(v)
	}

	if v := apiObject.ExclusionPatterns; v != nil {
		m["exclusion_patterns"] = flex.FlattenStringValueSet(v)
	}

	if v := apiObject.InclusionPatterns; v != nil {
		m["inclusion_patterns"] = flex.FlattenStringValueSet(v)
	}

	if v := apiObject.InclusionPrefixes; v != nil {
		m["inclusion_prefixes"] = flex.FlattenStringValueSet(v)
	}

	return []interface{}{m}
}

func flattenAccessControlListConfiguration(apiObject *types.AccessControlListConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.KeyPath; v != nil {
		m["key_path"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenDocumentsMetadataConfiguration(apiObject *types.DocumentsMetadataConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.S3Prefix; v != nil {
		m["s3_prefix"] = aws.ToString(v)
	}

	return []interface{}{m}
}

// Web Crawler Configuration
func flattenWebCrawlerConfiguration(apiObject *types.WebCrawlerConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"crawl_depth": aws.ToInt32(apiObject.CrawlDepth),
		"urls":        flattenURLs(apiObject.Urls),
	}

	if v := apiObject.AuthenticationConfiguration; v != nil {
		m["authentication_configuration"] = flattenAuthenticationConfiguration(v)
	}

	if v := apiObject.MaxContentSizePerPageInMegaBytes; v != nil {
		m["max_content_size_per_page_in_mega_bytes"] = aws.ToFloat32(v)
	}

	if v := apiObject.MaxLinksPerPage; v != nil {
		m["max_links_per_page"] = aws.ToInt32(v)
	}

	if v := apiObject.MaxUrlsPerMinuteCrawlRate; v != nil {
		m["max_urls_per_minute_crawl_rate"] = aws.ToInt32(v)
	}

	if v := apiObject.ProxyConfiguration; v != nil {
		m["proxy_configuration"] = flattenProxyConfiguration(v)
	}

	if v := apiObject.UrlExclusionPatterns; v != nil {
		m["url_exclusion_patterns"] = flex.FlattenStringValueSet(v)
	}

	if v := apiObject.UrlInclusionPatterns; v != nil {
		m["url_inclusion_patterns"] = flex.FlattenStringValueSet(v)
	}

	return []interface{}{m}
}

func flattenAuthenticationConfiguration(apiObject *types.AuthenticationConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.BasicAuthentication; v != nil {
		m["basic_authentication"] = flattenBasicAuthentication(v)
	}

	return []interface{}{m}
}

func flattenBasicAuthentication(basicAuthentications []types.BasicAuthenticationConfiguration) []interface{} {
	BasicAuthenticationList := []interface{}{}

	for _, basicAuthentication := range basicAuthentications {
		m := map[string]interface{}{
			"credentials":  aws.ToString(basicAuthentication.Credentials),
			"host":         aws.ToString(basicAuthentication.Host),
			names.AttrPort: aws.ToInt32(basicAuthentication.Port),
		}

		BasicAuthenticationList = append(BasicAuthenticationList, m)
	}

	return BasicAuthenticationList
}

func flattenProxyConfiguration(apiObject *types.ProxyConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"host":         aws.ToString(apiObject.Host),
		names.AttrPort: aws.ToInt32(apiObject.Port),
	}

	if v := apiObject.Credentials; v != nil {
		m["credentials"] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenURLs(apiObject *types.Urls) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SeedUrlConfiguration; v != nil {
		m["seed_url_configuration"] = flattenSeedURLConfiguration(v)
	}

	if v := apiObject.SiteMapsConfiguration; v != nil {
		m["site_maps_configuration"] = flattenSiteMapsConfiguration(v)
	}

	return []interface{}{m}
}

func flattenSeedURLConfiguration(apiObject *types.SeedUrlConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"seed_urls":        flex.FlattenStringValueSet(apiObject.SeedUrls),
		"web_crawler_mode": string(apiObject.WebCrawlerMode),
	}

	return []interface{}{m}
}

func flattenSiteMapsConfiguration(apiObject *types.SiteMapsConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"site_maps": flex.FlattenStringValueSet(apiObject.SiteMaps),
	}

	return []interface{}{m}
}

// Custom Document Enrichment Configuration
func flattenCustomDocumentEnrichmentConfiguration(apiObject *types.CustomDocumentEnrichmentConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.InlineConfigurations; v != nil {
		m["inline_configurations"] = flattenInlineConfigurations(v)
	}

	if v := apiObject.PostExtractionHookConfiguration; v != nil {
		m["post_extraction_hook_configuration"] = flattenHookConfiguration(v)
	}

	if v := apiObject.PreExtractionHookConfiguration; v != nil {
		m["pre_extraction_hook_configuration"] = flattenHookConfiguration(v)
	}

	if v := apiObject.RoleArn; v != nil {
		m[names.AttrRoleARN] = aws.ToString(v)
	}

	return []interface{}{m}
}

func flattenInlineConfigurations(inlineConfigurations []types.InlineCustomDocumentEnrichmentConfiguration) []interface{} {
	inlineConfigurationList := []interface{}{}

	for _, inlineConfiguration := range inlineConfigurations {
		m := map[string]interface{}{
			"document_content_deletion": inlineConfiguration.DocumentContentDeletion,
		}

		if v := inlineConfiguration.Condition; v != nil {
			m[names.AttrCondition] = flattenDocumentAttributeCondition(v)
		}

		if v := inlineConfiguration.Target; v != nil {
			m[names.AttrTarget] = flattenDocumentAttributeTarget(v)
		}

		inlineConfigurationList = append(inlineConfigurationList, m)
	}

	return inlineConfigurationList
}

func flattenDocumentAttributeTarget(apiObject *types.DocumentAttributeTarget) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"target_document_attribute_value_deletion": apiObject.TargetDocumentAttributeValueDeletion,
	}

	if v := apiObject.TargetDocumentAttributeKey; v != nil {
		m["target_document_attribute_key"] = aws.ToString(v)
	}

	if v := apiObject.TargetDocumentAttributeValue; v != nil {
		m["target_document_attribute_value"] = flattenDocumentAttributeValue(v)
	}

	return []interface{}{m}
}

func flattenHookConfiguration(apiObject *types.HookConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"lambda_arn":       aws.ToString(apiObject.LambdaArn),
		names.AttrS3Bucket: aws.ToString(apiObject.S3Bucket),
	}

	if v := apiObject.InvocationCondition; v != nil {
		m["invocation_condition"] = flattenDocumentAttributeCondition(v)
	}

	return []interface{}{m}
}

func flattenDocumentAttributeCondition(apiObject *types.DocumentAttributeCondition) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{
		"condition_document_attribute_key": aws.ToString(apiObject.ConditionDocumentAttributeKey),
		"operator":                         string(apiObject.Operator),
	}

	if v := apiObject.ConditionOnValue; v != nil {
		m["condition_on_value"] = flattenDocumentAttributeValue(v)
	}

	return []interface{}{m}
}

func flattenDocumentAttributeValue(apiObject *types.DocumentAttributeValue) []interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	// only one of these values should be set at a time
	if v := apiObject.DateValue; v != nil {
		// A date expressed as an ISO 8601 string.
		m["date_value"] = aws.ToTime(v).Format(ISO8601UTC)
	} else if v := apiObject.StringValue; v != nil {
		m["string_value"] = aws.ToString(v)
	} else if v := apiObject.StringListValue; v != nil {
		m["string_list_value"] = flex.FlattenStringValueSet(v)
	} else if v := apiObject.LongValue; v != nil {
		m["long_value"] = aws.ToInt64(v)
	}

	return []interface{}{m}
}
