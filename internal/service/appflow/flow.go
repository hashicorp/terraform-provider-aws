// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appflow

import (
	"context"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appflow"
	"github.com/aws/aws-sdk-go-v2/service/appflow/types"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appflow_flow", name="Flow")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/appflow;appflow.DescribeFlowOutput")
func resourceFlow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFlowCreate,
		ReadWithoutTimeout:   resourceFlowRead,
		UpdateWithoutTimeout: resourceFlowUpdate,
		DeleteWithoutTimeout: resourceFlowDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				p, err := arn.Parse(d.Id())
				if err != nil {
					return nil, err
				}
				name := strings.TrimPrefix(p.Resource, "flow/")
				d.Set(names.AttrName, name)

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`[\w!@#\-.?,\s]*`), "must contain only alphanumeric, underscore (_), exclamation point (!), at sign (@), number sign (#), hyphen (-), period (.), question mark (?), comma (,), and whitespace characters"),
			},
			"destination_flow_config": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 256)),
						},
						"connector_profile_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`[\w\/!@#+=.-]+`), "must contain only alphanumeric, underscore (_), forward slash (/), exclamation point (!), at sign (@), number sign (#), plus sign (+), equals sign (=), period (.), and hyphen (-) characters"), validation.StringLenBetween(1, 256)),
						},
						"connector_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ConnectorType](),
						},
						"destination_connector_properties": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"custom_connector": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"custom_properties": {
													Type:     schema.TypeMap,
													Optional: true,
													ValidateDiagFunc: validation.AllDiag(
														validation.MapKeyLenBetween(1, 128),
														validation.MapKeyMatch(regexache.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
													),
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 2048)),
													},
												},
												"entity_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 1024)),
												},
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"id_field_names": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 128)),
													},
												},
												"write_operation_type": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.WriteOperationType](),
												},
											},
										},
									},
									"customer_profiles": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrDomainName: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 64)),
												},
												"object_type_name": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 255)),
												},
											},
										},
									},
									"event_bridge": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"honeycode": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"lookout_metrics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{},
										},
									},
									"marketo": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"redshift": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrBucketPrefix: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"intermediate_bucket_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"s3": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrBucketName: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												names.AttrBucketPrefix: {
													Type:         schema.TypeString,
													Optional:     true,
													Computed:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"s3_output_format_config": {
													Type:     schema.TypeList,
													Optional: true,
													Computed: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"aggregation_config": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"aggregation_type": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			Computed:         true,
																			ValidateDiagFunc: enum.Validate[types.AggregationType](),
																		},
																		"target_file_size": {
																			Type:     schema.TypeInt,
																			Optional: true,
																			Computed: true,
																		},
																	},
																},
															},
															"file_type": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[types.FileType](),
															},
															"prefix_config": {
																Type:     schema.TypeList,
																Optional: true,
																Computed: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"prefix_hierarchy": {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			Elem: &schema.Schema{
																				Type:             schema.TypeString,
																				ValidateDiagFunc: enum.Validate[types.PathPrefix](),
																			},
																		},
																		"prefix_format": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			ValidateDiagFunc: enum.Validate[types.PrefixFormat](),
																		},
																		"prefix_type": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			ValidateDiagFunc: enum.Validate[types.PrefixType](),
																		},
																	},
																},
															},
															"preserve_source_data_typing": {
																Type:     schema.TypeBool,
																Optional: true,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},
									"salesforce": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"id_field_names": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 128)),
													},
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
												"write_operation_type": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.WriteOperationType](),
												},
											},
										},
									},
									"sapo_data": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"id_field_names": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 128)),
													},
												},
												"object_path": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
												"success_response_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
														},
													},
												},
												"write_operation_type": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.WriteOperationType](),
												},
											},
										},
									},
									"snowflake": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrBucketPrefix: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"intermediate_bucket_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"upsolver": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrBucketName: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`^(upsolver-appflow)\S*`), "must start with 'upsolver-appflow' and can not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												names.AttrBucketPrefix: {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"s3_output_format_config": {
													Type:     schema.TypeList,
													Required: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"aggregation_config": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"aggregation_type": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			ValidateDiagFunc: enum.Validate[types.AggregationType](),
																		},
																	},
																},
															},
															"file_type": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[types.FileType](),
															},
															"prefix_config": {
																Type:     schema.TypeList,
																Required: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"prefix_hierarchy": {
																			Type:     schema.TypeList,
																			Optional: true,
																			Computed: true,
																			Elem: &schema.Schema{
																				Type:             schema.TypeString,
																				ValidateDiagFunc: enum.Validate[types.PathPrefix](),
																			},
																		},
																		"prefix_format": {
																			Type:             schema.TypeString,
																			Optional:         true,
																			ValidateDiagFunc: enum.Validate[types.PrefixFormat](),
																		},
																		"prefix_type": {
																			Type:             schema.TypeString,
																			Required:         true,
																			ValidateDiagFunc: enum.Validate[types.PrefixType](),
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
									"zendesk": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															names.AttrBucketName: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															names.AttrBucketPrefix: {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
															"fail_on_first_destination_error": {
																Type:     schema.TypeBool,
																Optional: true,
															},
														},
													},
												},
												"id_field_names": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 128)),
													},
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
												"write_operation_type": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.WriteOperationType](),
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
			"flow_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"kms_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`arn:.*:kms:.*:[0-9]+:.*`), "must be a valid ARN of a Key Management Services (KMS) key"),
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`[0-9A-Za-z][\w!@#.-]+`), "must contain only alphanumeric, exclamation point (!), at sign (@), number sign (#), period (.), and hyphen (-) characters"), validation.StringLenBetween(1, 256)),
			},
			"source_flow_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 256)),
						},
						"connector_profile_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`[\w\/!@#+=.-]+`), "must contain only alphanumeric, underscore (_), forward slash (/), exclamation point (!), at sign (@), number sign (#), plus sign (+), equals sign (=), period (.), and hyphen (-) characters"), validation.StringLenBetween(1, 256)),
						},
						"connector_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.ConnectorType](),
						},
						"incremental_pull_config": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"datetime_type_field_name": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringLenBetween(0, 256),
									},
								},
							},
						},
						"source_connector_properties": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amplitude": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"custom_connector": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"custom_properties": {
													Type:     schema.TypeMap,
													Optional: true,
													ValidateDiagFunc: validation.AllDiag(
														validation.MapKeyLenBetween(1, 128),
														validation.MapKeyMatch(regexache.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
													),
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 2048)),
													},
												},
												"entity_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 1024)),
												},
											},
										},
									},
									"datadog": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"dynatrace": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"google_analytics": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"infor_nexus": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"marketo": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"s3": {
										Type:     schema.TypeList,
										Optional: true,
										Computed: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												names.AttrBucketName: {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												names.AttrBucketPrefix: {
													Type:         schema.TypeString,
													Required:     true,
													ForceNew:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"s3_input_format_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"s3_input_file_type": {
																Type:             schema.TypeString,
																Optional:         true,
																ValidateDiagFunc: enum.Validate[types.S3InputFileType](),
															},
														},
													},
												},
											},
										},
									},
									"salesforce": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"enable_dynamic_field_update": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"include_deleted_records": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"sapo_data": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object_path": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"service_now": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"singular": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"slack": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"trendmicro": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"veeva": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"document_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`[\s\w_-]+`), "must contain only alphanumeric, underscore (_), and hyphen (-) characters"), validation.StringLenBetween(1, 512)),
												},
												"include_all_versions": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"include_renditions": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"include_source_files": {
													Type:     schema.TypeBool,
													Optional: true,
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"zendesk": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexache.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"task": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector_operator": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"amplitude": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.AmplitudeConnectorOperator](),
									},
									"custom_connector": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.Operator](),
									},
									"datadog": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.DatadogConnectorOperator](),
									},
									"dynatrace": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.DynatraceConnectorOperator](),
									},
									"google_analytics": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.GoogleAnalyticsConnectorOperator](),
									},
									"infor_nexus": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.InforNexusConnectorOperator](),
									},
									"marketo": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.MarketoConnectorOperator](),
									},
									"s3": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.S3ConnectorOperator](),
									},
									"salesforce": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.SalesforceConnectorOperator](),
									},
									"sapo_data": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.SAPODataConnectorOperator](),
									},
									"service_now": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.ServiceNowConnectorOperator](),
									},
									"singular": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.SingularConnectorOperator](),
									},
									"slack": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.SlackConnectorOperator](),
									},
									"trendmicro": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.TrendmicroConnectorOperator](),
									},
									"veeva": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.VeevaConnectorOperator](),
									},
									"zendesk": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[types.ZendeskConnectorOperator](),
									},
								},
							},
						},
						"destination_field": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 256),
						},
						"source_fields": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 2048),
							},
							DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
								if v, ok := d.Get("task").(*schema.Set); ok && v.Len() == 1 {
									if tl, ok := v.List()[0].(map[string]interface{}); ok && len(tl) > 0 {
										if sf, ok := tl["source_fields"].([]interface{}); ok && len(sf) == 1 {
											if sf[0] == "" {
												return oldValue == "0" && newValue == "1"
											}
										}
									}
								}
								return false
							},
						},
						"task_properties": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 2048),
							},
						},
						"task_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.TaskType](),
						},
					},
				},
			},
			"trigger_config": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"trigger_properties": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"scheduled": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"data_pull_mode": {
													Type:             schema.TypeString,
													Optional:         true,
													ValidateDiagFunc: enum.Validate[types.DataPullMode](),
												},
												"first_execution_from": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.IsRFC3339Time,
												},
												"schedule_end_time": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.IsRFC3339Time,
												},
												names.AttrScheduleExpression: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 256),
												},
												"schedule_offset": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IntBetween(0, 36000),
												},
												"schedule_start_time": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.IsRFC3339Time,
												},
												"timezone": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 256),
												},
											},
										},
									},
								},
							},
						},
						"trigger_type": {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[types.TriggerType](),
						},
					},
				},
			},
			"metadata_catalog_config": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"glue_data_catalog": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrDatabaseName: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrRoleARN: {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: validation.ToDiagFunc(verify.ValidARN),
									},
									"table_prefix": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppFlowClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appflow.CreateFlowInput{
		FlowName:                  aws.String(name),
		DestinationFlowConfigList: expandDestinationFlowConfigs(d.Get("destination_flow_config").([]interface{})),
		SourceFlowConfig:          expandSourceFlowConfig(d.Get("source_flow_config").([]interface{})[0].(map[string]interface{})),
		Tags:                      getTagsIn(ctx),
		Tasks:                     expandTasks(d.Get("task").(*schema.Set).List()),
		TriggerConfig:             expandTriggerConfig(d.Get("trigger_config").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk("metadata_catalog_config"); ok {
		input.MetadataCatalogConfig = expandMetadataCatalogConfig(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_arn"); ok {
		input.KmsArn = aws.String(v.(string))
	}

	output, err := conn.CreateFlow(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating AppFlow Flow (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.FlowArn))

	return append(diags, resourceFlowRead(ctx, d, meta)...)
}

func resourceFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppFlowClient(ctx)

	flowDefinition, err := findFlowByName(ctx, conn, d.Get(names.AttrName).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppFlow Flow (%s) not found, removing from state", d.Get(names.AttrName))
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppFlow Flow (%s): %s", d.Get(names.AttrName), err)
	}

	output, err := findFlowByName(ctx, conn, aws.ToString(flowDefinition.FlowName))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AppFlow Flow (%s): %s", d.Get(names.AttrName), err)
	}

	d.Set(names.AttrARN, output.FlowArn)
	d.Set(names.AttrDescription, output.Description)
	if err := d.Set("destination_flow_config", flattenDestinationFlowConfigs(output.DestinationFlowConfigList)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting destination_flow_config: %s", err)
	}
	d.Set("flow_status", output.FlowStatus)
	d.Set("kms_arn", output.KmsArn)
	d.Set(names.AttrName, output.FlowName)
	if output.SourceFlowConfig != nil {
		if err := d.Set("source_flow_config", []interface{}{flattenSourceFlowConfig(output.SourceFlowConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting source_flow_config: %s", err)
		}
	} else {
		d.Set("source_flow_config", nil)
	}
	if err := d.Set("task", flattenTasks(output.Tasks)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting task: %s", err)
	}
	if output.TriggerConfig != nil {
		if err := d.Set("trigger_config", []interface{}{flattenTriggerConfig(output.TriggerConfig)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting trigger_config: %s", err)
		}
	} else {
		d.Set("trigger_config", nil)
	}

	if output.MetadataCatalogConfig != nil {
		if err := d.Set("metadata_catalog_config", flattenMetadataCatalogConfig(output.MetadataCatalogConfig)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting metadata_catalog_config: %s", err)
		}
	} else {
		d.Set("metadata_catalog_config", nil)
	}

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppFlowClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &appflow.UpdateFlowInput{
			DestinationFlowConfigList: expandDestinationFlowConfigs(d.Get("destination_flow_config").([]interface{})),
			FlowName:                  aws.String(d.Get(names.AttrName).(string)),
			SourceFlowConfig:          expandSourceFlowConfig(d.Get("source_flow_config").([]interface{})[0].(map[string]interface{})),
			Tasks:                     expandTasks(d.Get("task").(*schema.Set).List()),
			TriggerConfig:             expandTriggerConfig(d.Get("trigger_config").([]interface{})[0].(map[string]interface{})),
		}

		if v, ok := d.GetOk("metadata_catalog_config"); ok {
			input.MetadataCatalogConfig = expandMetadataCatalogConfig(v.([]any))
		}

		// always send description when updating a task
		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		_, err := conn.UpdateFlow(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating AppFlow Flow (%s): %s", d.Get(names.AttrName), err)
		}
	}

	return append(diags, resourceFlowRead(ctx, d, meta)...)
}

func resourceFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppFlowClient(ctx)

	log.Printf("[INFO] Deleting AppFlow Flow: %s", d.Get(names.AttrName))
	_, err := conn.DeleteFlow(ctx, &appflow.DeleteFlowInput{
		FlowName: aws.String(d.Get(names.AttrName).(string)),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AppFlow Flow (%s): %s", d.Get(names.AttrName), err)
	}

	if _, err := waitFlowDeleted(ctx, conn, d.Get(names.AttrName).(string)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for AppFlow Flow (%s) delete: %s", d.Get(names.AttrName), err)
	}

	return diags
}

func findFlowByName(ctx context.Context, conn *appflow.Client, name string) (*appflow.DescribeFlowOutput, error) {
	input := &appflow.DescribeFlowInput{
		FlowName: aws.String(name),
	}

	output, err := conn.DescribeFlow(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.FlowStatus; status == types.FlowStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func statusFlow(ctx context.Context, conn *appflow.Client, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFlowByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.FlowStatus), nil
	}
}

func waitFlowDeleted(ctx context.Context, conn *appflow.Client, name string) (*types.FlowDefinition, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Target:  []string{},
		Refresh: statusFlow(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.FlowDefinition); ok {
		return output, err
	}

	return nil, err
}

func expandErrorHandlingConfig(tfMap map[string]interface{}) *types.ErrorHandlingConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.ErrorHandlingConfig{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrBucketPrefix].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["fail_on_first_destination_error"].(bool); ok {
		a.FailOnFirstDestinationError = v
	}

	return a
}

func expandAggregationConfig(tfMap map[string]interface{}) *types.AggregationConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.AggregationConfig{}

	if v, ok := tfMap["aggregation_type"].(string); ok && v != "" {
		a.AggregationType = types.AggregationType(v)
	}

	if v, ok := tfMap["target_file_size"].(int); ok && v != 0 {
		a.TargetFileSize = aws.Int64(int64(v))
	}

	return a
}

func expandPrefixConfig(tfMap map[string]interface{}) *types.PrefixConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.PrefixConfig{}

	if v, ok := tfMap["prefix_format"].(string); ok && v != "" {
		a.PrefixFormat = types.PrefixFormat(v)
	}

	if v, ok := tfMap["prefix_type"].(string); ok && v != "" {
		a.PrefixType = types.PrefixType(v)
	}

	if v, ok := tfMap["prefix_hierarchy"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.PathPrefixHierarchy = flex.ExpandStringyValueList[types.PathPrefix](v)
	}

	return a
}

func expandDestinationFlowConfigs(tfList []interface{}) []types.DestinationFlowConfig {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.DestinationFlowConfig

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandDestinationFlowConfig(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandDestinationFlowConfig(tfMap map[string]interface{}) *types.DestinationFlowConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.DestinationFlowConfig{}

	if v, ok := tfMap["api_version"].(string); ok && v != "" {
		a.ApiVersion = aws.String(v)
	}

	if v, ok := tfMap["connector_profile_name"].(string); ok && v != "" {
		a.ConnectorProfileName = aws.String(v)
	}

	if v, ok := tfMap["connector_type"].(string); ok && v != "" {
		a.ConnectorType = types.ConnectorType(v)
	} else {
		// https://github.com/hashicorp/terraform-provider-aws/issues/26491.
		return nil
	}

	if v, ok := tfMap["destination_connector_properties"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.DestinationConnectorProperties = expandDestinationConnectorProperties(v[0].(map[string]interface{}))
	}

	return a
}

func expandDestinationConnectorProperties(tfMap map[string]interface{}) *types.DestinationConnectorProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.DestinationConnectorProperties{}

	if v, ok := tfMap["custom_connector"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.CustomConnector = expandCustomConnectorDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["customer_profiles"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.CustomerProfiles = expandCustomerProfilesDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["event_bridge"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.EventBridge = expandEventBridgeDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["honeycode"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Honeycode = expandHoneycodeDestinationProperties(v[0].(map[string]interface{}))
	}

	// API reference does not list valid attributes for LookoutMetricsDestinationProperties
	// https://docs.aws.amazon.com/appflow/1.0/APIReference/API_LookoutMetricsDestinationProperties.html
	if v, ok := tfMap["lookout_metrics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.LookoutMetrics = v[0].(*types.LookoutMetricsDestinationProperties)
	}

	if v, ok := tfMap["marketo"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Marketo = expandMarketoDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["redshift"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Redshift = expandRedshiftDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.S3 = expandS3DestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["salesforce"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Salesforce = expandSalesforceDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sapo_data"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.SAPOData = expandSAPODataDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["snowflake"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Snowflake = expandSnowflakeDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["upsolver"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Upsolver = expandUpsolverDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["zendesk"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Zendesk = expandZendeskDestinationProperties(v[0].(map[string]interface{}))
	}

	return a
}

func expandCustomConnectorDestinationProperties(tfMap map[string]interface{}) *types.CustomConnectorDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.CustomConnectorDestinationProperties{}

	if v, ok := tfMap["custom_properties"].(map[string]interface{}); ok && len(v) > 0 {
		a.CustomProperties = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["entity_name"].(string); ok && v != "" {
		a.EntityName = aws.String(v)
	}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["id_field_names"].([]interface{}); ok && len(v) > 0 {
		a.IdFieldNames = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["write_operation_type"].(string); ok && v != "" {
		a.WriteOperationType = types.WriteOperationType(v)
	}

	return a
}

func expandCustomerProfilesDestinationProperties(tfMap map[string]interface{}) *types.CustomerProfilesDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.CustomerProfilesDestinationProperties{}

	if v, ok := tfMap[names.AttrDomainName].(string); ok && v != "" {
		a.DomainName = aws.String(v)
	}

	if v, ok := tfMap["object_type_name"].(string); ok && v != "" {
		a.ObjectTypeName = aws.String(v)
	}

	return a
}

func expandEventBridgeDestinationProperties(tfMap map[string]interface{}) *types.EventBridgeDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.EventBridgeDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandHoneycodeDestinationProperties(tfMap map[string]interface{}) *types.HoneycodeDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.HoneycodeDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandMarketoDestinationProperties(tfMap map[string]interface{}) *types.MarketoDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.MarketoDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandRedshiftDestinationProperties(tfMap map[string]interface{}) *types.RedshiftDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.RedshiftDestinationProperties{}

	if v, ok := tfMap[names.AttrBucketPrefix].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["intermediate_bucket_name"].(string); ok && v != "" {
		a.IntermediateBucketName = aws.String(v)
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandS3DestinationProperties(tfMap map[string]interface{}) *types.S3DestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.S3DestinationProperties{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrBucketPrefix].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["s3_output_format_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.S3OutputFormatConfig = expandS3OutputFormatConfig(v[0].(map[string]interface{}))
	}

	return a
}

func expandS3OutputFormatConfig(tfMap map[string]interface{}) *types.S3OutputFormatConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.S3OutputFormatConfig{}

	if v, ok := tfMap["aggregation_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.AggregationConfig = expandAggregationConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["file_type"].(string); ok && v != "" {
		a.FileType = types.FileType(v)
	}

	if v, ok := tfMap["prefix_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.PrefixConfig = expandPrefixConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["preserve_source_data_typing"].(bool); ok {
		a.PreserveSourceDataTyping = aws.Bool(v)
	}

	return a
}

func expandSalesforceDestinationProperties(tfMap map[string]interface{}) *types.SalesforceDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.SalesforceDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["id_field_names"].([]interface{}); ok && len(v) > 0 {
		a.IdFieldNames = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	if v, ok := tfMap["write_operation_type"].(string); ok && v != "" {
		a.WriteOperationType = types.WriteOperationType(v)
	}

	return a
}

func expandSAPODataDestinationProperties(tfMap map[string]interface{}) *types.SAPODataDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.SAPODataDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["id_field_names"].([]interface{}); ok && len(v) > 0 {
		a.IdFieldNames = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["object_path"].(string); ok && v != "" {
		a.ObjectPath = aws.String(v)
	}

	if v, ok := tfMap["success_response_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.SuccessResponseHandlingConfig = expandSuccessResponseHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["write_operation_type"].(string); ok && v != "" {
		a.WriteOperationType = types.WriteOperationType(v)
	}

	return a
}

func expandSuccessResponseHandlingConfig(tfMap map[string]interface{}) *types.SuccessResponseHandlingConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.SuccessResponseHandlingConfig{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrBucketPrefix].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	return a
}

func expandSnowflakeDestinationProperties(tfMap map[string]interface{}) *types.SnowflakeDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.SnowflakeDestinationProperties{}

	if v, ok := tfMap[names.AttrBucketPrefix].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["intermediate_bucket_name"].(string); ok && v != "" {
		a.IntermediateBucketName = aws.String(v)
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandUpsolverDestinationProperties(tfMap map[string]interface{}) *types.UpsolverDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.UpsolverDestinationProperties{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrBucketPrefix].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["s3_output_format_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.S3OutputFormatConfig = expandUpsolverS3OutputFormatConfig(v[0].(map[string]interface{}))
	}

	return a
}

func expandUpsolverS3OutputFormatConfig(tfMap map[string]interface{}) *types.UpsolverS3OutputFormatConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.UpsolverS3OutputFormatConfig{}

	if v, ok := tfMap["aggregation_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.AggregationConfig = expandAggregationConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["file_type"].(string); ok && v != "" {
		a.FileType = types.FileType(v)
	}

	if v, ok := tfMap["prefix_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.PrefixConfig = expandPrefixConfig(v[0].(map[string]interface{}))
	}

	return a
}

func expandZendeskDestinationProperties(tfMap map[string]interface{}) *types.ZendeskDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.ZendeskDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["id_field_names"].([]interface{}); ok && len(v) > 0 {
		a.IdFieldNames = flex.ExpandStringValueList(v)
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	if v, ok := tfMap["write_operation_type"].(string); ok && v != "" {
		a.WriteOperationType = types.WriteOperationType(v)
	}

	return a
}

func expandSourceFlowConfig(tfMap map[string]interface{}) *types.SourceFlowConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.SourceFlowConfig{}

	if v, ok := tfMap["api_version"].(string); ok && v != "" {
		a.ApiVersion = aws.String(v)
	}

	if v, ok := tfMap["connector_profile_name"].(string); ok && v != "" {
		a.ConnectorProfileName = aws.String(v)
	}

	if v, ok := tfMap["connector_type"].(string); ok && v != "" {
		a.ConnectorType = types.ConnectorType(v)
	}

	if v, ok := tfMap["incremental_pull_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.IncrementalPullConfig = expandIncrementalPullConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["source_connector_properties"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.SourceConnectorProperties = expandSourceConnectorProperties(v[0].(map[string]interface{}))
	}

	return a
}

func expandIncrementalPullConfig(tfMap map[string]interface{}) *types.IncrementalPullConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.IncrementalPullConfig{}

	if v, ok := tfMap["datetime_type_field_name"].(string); ok && v != "" {
		a.DatetimeTypeFieldName = aws.String(v)
	}

	return a
}

func expandSourceConnectorProperties(tfMap map[string]interface{}) *types.SourceConnectorProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.SourceConnectorProperties{}

	if v, ok := tfMap["amplitude"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Amplitude = expandAmplitudeSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["custom_connector"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.CustomConnector = expandCustomConnectorSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["datadog"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Datadog = expandDatadogSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["dynatrace"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Dynatrace = expandDynatraceSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["google_analytics"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.GoogleAnalytics = expandGoogleAnalyticsSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["infor_nexus"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.InforNexus = expandInforNexusSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["marketo"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Marketo = expandMarketoSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.S3 = expandS3SourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sapo_data"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.SAPOData = expandSAPODataSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["salesforce"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Salesforce = expandSalesforceSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["service_now"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ServiceNow = expandServiceNowSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["singular"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Singular = expandSingularSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["slack"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Slack = expandSlackSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["trendmicro"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Trendmicro = expandTrendmicroSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["veeva"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Veeva = expandVeevaSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["zendesk"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Zendesk = expandZendeskSourceProperties(v[0].(map[string]interface{}))
	}

	return a
}

func expandAmplitudeSourceProperties(tfMap map[string]interface{}) *types.AmplitudeSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.AmplitudeSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandCustomConnectorSourceProperties(tfMap map[string]interface{}) *types.CustomConnectorSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.CustomConnectorSourceProperties{}

	if v, ok := tfMap["custom_properties"].(map[string]interface{}); ok && len(v) > 0 {
		a.CustomProperties = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["entity_name"].(string); ok && v != "" {
		a.EntityName = aws.String(v)
	}

	return a
}

func expandDatadogSourceProperties(tfMap map[string]interface{}) *types.DatadogSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.DatadogSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandDynatraceSourceProperties(tfMap map[string]interface{}) *types.DynatraceSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.DynatraceSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandGoogleAnalyticsSourceProperties(tfMap map[string]interface{}) *types.GoogleAnalyticsSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.GoogleAnalyticsSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandInforNexusSourceProperties(tfMap map[string]interface{}) *types.InforNexusSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.InforNexusSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandMarketoSourceProperties(tfMap map[string]interface{}) *types.MarketoSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.MarketoSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandS3SourceProperties(tfMap map[string]interface{}) *types.S3SourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.S3SourceProperties{}

	if v, ok := tfMap[names.AttrBucketName].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrBucketPrefix].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["s3_input_format_config"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.S3InputFormatConfig = expandS3InputFormatConfig(v[0].(map[string]interface{}))
	}

	return a
}

func expandS3InputFormatConfig(tfMap map[string]interface{}) *types.S3InputFormatConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.S3InputFormatConfig{}

	if v, ok := tfMap["s3_input_file_type"].(string); ok && v != "" {
		a.S3InputFileType = types.S3InputFileType(v)
	}

	return a
}

func expandSalesforceSourceProperties(tfMap map[string]interface{}) *types.SalesforceSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.SalesforceSourceProperties{}

	if v, ok := tfMap["enable_dynamic_field_update"].(bool); ok {
		a.EnableDynamicFieldUpdate = v
	}

	if v, ok := tfMap["include_deleted_records"].(bool); ok {
		a.IncludeDeletedRecords = v
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandSAPODataSourceProperties(tfMap map[string]interface{}) *types.SAPODataSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.SAPODataSourceProperties{}

	if v, ok := tfMap["object_path"].(string); ok && v != "" {
		a.ObjectPath = aws.String(v)
	}

	return a
}

func expandServiceNowSourceProperties(tfMap map[string]interface{}) *types.ServiceNowSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.ServiceNowSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandSingularSourceProperties(tfMap map[string]interface{}) *types.SingularSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.SingularSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandSlackSourceProperties(tfMap map[string]interface{}) *types.SlackSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.SlackSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandTrendmicroSourceProperties(tfMap map[string]interface{}) *types.TrendmicroSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.TrendmicroSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandVeevaSourceProperties(tfMap map[string]interface{}) *types.VeevaSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.VeevaSourceProperties{}

	if v, ok := tfMap["document_type"].(string); ok && v != "" {
		a.DocumentType = aws.String(v)
	}

	if v, ok := tfMap["include_all_versions"].(bool); ok {
		a.IncludeAllVersions = v
	}

	if v, ok := tfMap["include_renditions"].(bool); ok {
		a.IncludeRenditions = v
	}

	if v, ok := tfMap["include_source_files"].(bool); ok {
		a.IncludeSourceFiles = v
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandZendeskSourceProperties(tfMap map[string]interface{}) *types.ZendeskSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.ZendeskSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandTasks(tfList []interface{}) []types.Task {
	if len(tfList) == 0 {
		return nil
	}

	var s []types.Task

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandTask(m)

		if a == nil {
			continue
		}

		s = append(s, *a)
	}

	return s
}

func expandTask(tfMap map[string]interface{}) *types.Task {
	if tfMap == nil {
		return nil
	}

	a := &types.Task{}

	if v, ok := tfMap["connector_operator"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.ConnectorOperator = expandConnectorOperator(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["destination_field"].(string); ok && v != "" {
		a.DestinationField = aws.String(v)
	}

	if v, ok := tfMap["source_fields"].([]interface{}); ok && len(v) > 0 {
		a.SourceFields = flex.ExpandStringValueList(v)
	} else {
		a.SourceFields = []string{} // send an empty object if source_fields is empty (required by API)
	}

	if v, ok := tfMap["task_properties"].(map[string]interface{}); ok && len(v) > 0 {
		a.TaskProperties = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["task_type"].(string); ok && v != "" {
		a.TaskType = types.TaskType(v)
	} else {
		// https://github.com/hashicorp/terraform-provider-aws/issues/28237.
		return nil
	}

	return a
}

func expandConnectorOperator(tfMap map[string]interface{}) *types.ConnectorOperator {
	if tfMap == nil {
		return nil
	}

	a := &types.ConnectorOperator{}

	if v, ok := tfMap["amplitude"].(string); ok && v != "" {
		a.Amplitude = types.AmplitudeConnectorOperator(v)
	}

	if v, ok := tfMap["custom_connector"].(string); ok && v != "" {
		a.CustomConnector = types.Operator(v)
	}

	if v, ok := tfMap["datadog"].(string); ok && v != "" {
		a.Datadog = types.DatadogConnectorOperator(v)
	}

	if v, ok := tfMap["dynatrace"].(string); ok && v != "" {
		a.Dynatrace = types.DynatraceConnectorOperator(v)
	}

	if v, ok := tfMap["google_analytics"].(string); ok && v != "" {
		a.GoogleAnalytics = types.GoogleAnalyticsConnectorOperator(v)
	}

	if v, ok := tfMap["infor_nexus"].(string); ok && v != "" {
		a.InforNexus = types.InforNexusConnectorOperator(v)
	}

	if v, ok := tfMap["marketo"].(string); ok && v != "" {
		a.Marketo = types.MarketoConnectorOperator(v)
	}

	if v, ok := tfMap["s3"].(string); ok && v != "" {
		a.S3 = types.S3ConnectorOperator(v)
	}

	if v, ok := tfMap["sapo_data"].(string); ok && v != "" {
		a.SAPOData = types.SAPODataConnectorOperator(v)
	}

	if v, ok := tfMap["salesforce"].(string); ok && v != "" {
		a.Salesforce = types.SalesforceConnectorOperator(v)
	}

	if v, ok := tfMap["service_now"].(string); ok && v != "" {
		a.ServiceNow = types.ServiceNowConnectorOperator(v)
	}

	if v, ok := tfMap["singular"].(string); ok && v != "" {
		a.Singular = types.SingularConnectorOperator(v)
	}

	if v, ok := tfMap["slack"].(string); ok && v != "" {
		a.Slack = types.SlackConnectorOperator(v)
	}

	if v, ok := tfMap["trendmicro"].(string); ok && v != "" {
		a.Trendmicro = types.TrendmicroConnectorOperator(v)
	}

	if v, ok := tfMap["veeva"].(string); ok && v != "" {
		a.Veeva = types.VeevaConnectorOperator(v)
	}

	if v, ok := tfMap["zendesk"].(string); ok && v != "" {
		a.Zendesk = types.ZendeskConnectorOperator(v)
	}

	return a
}

func expandTriggerConfig(tfMap map[string]interface{}) *types.TriggerConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.TriggerConfig{}

	if v, ok := tfMap["trigger_properties"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.TriggerProperties = expandTriggerProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["trigger_type"].(string); ok && v != "" {
		a.TriggerType = types.TriggerType(v)
	}

	return a
}

func expandTriggerProperties(tfMap map[string]interface{}) *types.TriggerProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.TriggerProperties{}

	// Only return TriggerProperties if nested field is non-empty
	if v, ok := tfMap["scheduled"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.Scheduled = expandScheduledTriggerProperties(v[0].(map[string]interface{}))
		return a
	}

	return nil
}

func expandScheduledTriggerProperties(tfMap map[string]interface{}) *types.ScheduledTriggerProperties {
	if tfMap == nil {
		return nil
	}

	a := &types.ScheduledTriggerProperties{}

	if v, ok := tfMap["data_pull_mode"].(string); ok && v != "" {
		a.DataPullMode = types.DataPullMode(v)
	}

	if v, ok := tfMap["first_execution_from"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		a.FirstExecutionFrom = aws.Time(v)
	}

	if v, ok := tfMap["schedule_end_time"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		a.ScheduleEndTime = aws.Time(v)
	}

	if v, ok := tfMap[names.AttrScheduleExpression].(string); ok && v != "" {
		a.ScheduleExpression = aws.String(v)
	}

	if v, ok := tfMap["schedule_offset"].(int); ok && v != 0 {
		a.ScheduleOffset = aws.Int64(int64(v))
	}

	if v, ok := tfMap["schedule_start_time"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		a.ScheduleStartTime = aws.Time(v)
	}

	if v, ok := tfMap["timezone"].(string); ok && v != "" {
		a.Timezone = aws.String(v)
	}

	return a
}

func expandMetadataCatalogConfig(tfList []any) *types.MetadataCatalogConfig {
	if len(tfList) == 0 {
		return nil
	}

	m := tfList[0].(map[string]any)

	a := &types.MetadataCatalogConfig{}

	if v, ok := m["glue_data_catalog"].([]any); ok && len(v) > 0 && v[0] != nil {
		a.GlueDataCatalog = expandGlueDataCatalog(v[0].(map[string]any))
	}

	return a
}

func expandGlueDataCatalog(tfMap map[string]interface{}) *types.GlueDataCatalogConfig {
	if tfMap == nil {
		return nil
	}

	a := &types.GlueDataCatalogConfig{}

	if v, ok := tfMap[names.AttrDatabaseName].(string); ok && v != "" {
		a.DatabaseName = aws.String(v)
	}

	if v, ok := tfMap[names.AttrRoleARN].(string); ok && v != "" {
		a.RoleArn = aws.String(v)
	}

	if v, ok := tfMap["table_prefix"].(string); ok && v != "" {
		a.TablePrefix = aws.String(v)
	}

	return a
}

func flattenMetadataCatalogConfig(in *types.MetadataCatalogConfig) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		"glue_data_catalog": flattenGlueDataCatalog(in.GlueDataCatalog),
	}

	return []any{m}
}

func flattenGlueDataCatalog(in *types.GlueDataCatalogConfig) []any {
	if in == nil {
		return nil
	}

	m := map[string]any{
		names.AttrDatabaseName: in.DatabaseName,
		names.AttrRoleARN:      in.RoleArn,
		"table_prefix":         in.TablePrefix,
	}

	return []any{m}
}

func flattenErrorHandlingConfig(errorHandlingConfig *types.ErrorHandlingConfig) map[string]interface{} {
	if errorHandlingConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := errorHandlingConfig.BucketName; v != nil {
		m[names.AttrBucketName] = aws.ToString(v)
	}

	if v := errorHandlingConfig.BucketPrefix; v != nil {
		m[names.AttrBucketPrefix] = aws.ToString(v)
	}

	m["fail_on_first_destination_error"] = errorHandlingConfig.FailOnFirstDestinationError

	return m
}

func flattenPrefixConfig(prefixConfig *types.PrefixConfig) map[string]interface{} {
	if prefixConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["prefix_format"] = prefixConfig.PrefixFormat
	m["prefix_type"] = prefixConfig.PrefixType
	m["prefix_hierarchy"] = flex.FlattenStringyValueList(prefixConfig.PathPrefixHierarchy)

	return m
}

func flattenAggregationConfig(aggregationConfig *types.AggregationConfig) map[string]interface{} {
	if aggregationConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["aggregation_type"] = aggregationConfig.AggregationType
	m["target_file_size"] = aggregationConfig.TargetFileSize

	return m
}

func flattenDestinationFlowConfigs(destinationFlowConfigs []types.DestinationFlowConfig) []interface{} {
	if len(destinationFlowConfigs) == 0 {
		return nil
	}

	var l []interface{}

	for _, destinationFlowConfig := range destinationFlowConfigs {
		l = append(l, flattenDestinationFlowConfig(destinationFlowConfig))
	}

	return l
}

func flattenDestinationFlowConfig(destinationFlowConfig types.DestinationFlowConfig) map[string]interface{} {
	m := map[string]interface{}{}

	if v := destinationFlowConfig.ApiVersion; v != nil {
		m["api_version"] = aws.ToString(v)
	}

	if v := destinationFlowConfig.ConnectorProfileName; v != nil {
		m["connector_profile_name"] = aws.ToString(v)
	}

	m["connector_type"] = destinationFlowConfig.ConnectorType

	if v := destinationFlowConfig.DestinationConnectorProperties; v != nil {
		m["destination_connector_properties"] = []interface{}{flattenDestinationConnectorProperties(v)}
	}

	return m
}

func flattenDestinationConnectorProperties(destinationConnectorProperties *types.DestinationConnectorProperties) map[string]interface{} {
	if destinationConnectorProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := destinationConnectorProperties.CustomConnector; v != nil {
		m["custom_connector"] = []interface{}{flattenCustomConnectorDestinationProperties(v)}
	}

	if v := destinationConnectorProperties.CustomerProfiles; v != nil {
		m["customer_profiles"] = []interface{}{flattenCustomerProfilesDestinationProperties(v)}
	}

	if v := destinationConnectorProperties.EventBridge; v != nil {
		m["event_bridge"] = []interface{}{flattenEventBridgeDestinationProperties(v)}
	}

	if v := destinationConnectorProperties.Honeycode; v != nil {
		m["honeycode"] = []interface{}{flattenHoneycodeDestinationProperties(v)}
	}

	// API reference does not list valid attributes for LookoutMetricsDestinationProperties
	// https://docs.aws.amazon.com/appflow/1.0/APIReference/API_LookoutMetricsDestinationProperties.html
	if v := destinationConnectorProperties.LookoutMetrics; v != nil {
		m["lookout_metrics"] = []interface{}{map[string]interface{}{}}
	}

	if v := destinationConnectorProperties.Marketo; v != nil {
		m["marketo"] = []interface{}{flattenMarketoDestinationProperties(v)}
	}

	if v := destinationConnectorProperties.Redshift; v != nil {
		m["redshift"] = []interface{}{flattenRedshiftDestinationProperties(v)}
	}

	if v := destinationConnectorProperties.S3; v != nil {
		m["s3"] = []interface{}{flattenS3DestinationProperties(v)}
	}

	if v := destinationConnectorProperties.Salesforce; v != nil {
		m["salesforce"] = []interface{}{flattenSalesforceDestinationProperties(v)}
	}

	if v := destinationConnectorProperties.SAPOData; v != nil {
		m["sapo_data"] = []interface{}{flattenSAPODataDestinationProperties(v)}
	}

	if v := destinationConnectorProperties.Snowflake; v != nil {
		m["snowflake"] = []interface{}{flattenSnowflakeDestinationProperties(v)}
	}

	if v := destinationConnectorProperties.Upsolver; v != nil {
		m["upsolver"] = []interface{}{flattenUpsolverDestinationProperties(v)}
	}

	if v := destinationConnectorProperties.Zendesk; v != nil {
		m["zendesk"] = []interface{}{flattenZendeskDestinationProperties(v)}
	}

	return m
}

func flattenCustomConnectorDestinationProperties(customConnectorDestinationProperties *types.CustomConnectorDestinationProperties) map[string]interface{} {
	if customConnectorDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := customConnectorDestinationProperties.CustomProperties; v != nil {
		m["custom_properties"] = v
	}

	if v := customConnectorDestinationProperties.EntityName; v != nil {
		m["entity_name"] = aws.ToString(v)
	}

	if v := customConnectorDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := customConnectorDestinationProperties.IdFieldNames; v != nil {
		m["id_field_names"] = v
	}

	m["write_operation_type"] = customConnectorDestinationProperties.WriteOperationType

	return m
}

func flattenCustomerProfilesDestinationProperties(customerProfilesDestinationProperties *types.CustomerProfilesDestinationProperties) map[string]interface{} {
	if customerProfilesDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := customerProfilesDestinationProperties.DomainName; v != nil {
		m[names.AttrDomainName] = aws.ToString(v)
	}

	if v := customerProfilesDestinationProperties.ObjectTypeName; v != nil {
		m["object_type_name"] = aws.ToString(v)
	}

	return m
}

func flattenEventBridgeDestinationProperties(eventBridgeDestinationProperties *types.EventBridgeDestinationProperties) map[string]interface{} {
	if eventBridgeDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := eventBridgeDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := eventBridgeDestinationProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenHoneycodeDestinationProperties(honeycodeDestinationProperties *types.HoneycodeDestinationProperties) map[string]interface{} {
	if honeycodeDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := honeycodeDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := honeycodeDestinationProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenMarketoDestinationProperties(marketoDestinationProperties *types.MarketoDestinationProperties) map[string]interface{} {
	if marketoDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := marketoDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := marketoDestinationProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenRedshiftDestinationProperties(redshiftDestinationProperties *types.RedshiftDestinationProperties) map[string]interface{} {
	if redshiftDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := redshiftDestinationProperties.BucketPrefix; v != nil {
		m[names.AttrBucketPrefix] = aws.ToString(v)
	}

	if v := redshiftDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := redshiftDestinationProperties.IntermediateBucketName; v != nil {
		m["intermediate_bucket_name"] = aws.ToString(v)
	}

	if v := redshiftDestinationProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenS3DestinationProperties(s3DestinationProperties *types.S3DestinationProperties) map[string]interface{} {
	if s3DestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := s3DestinationProperties.BucketName; v != nil {
		m[names.AttrBucketName] = aws.ToString(v)
	}

	if v := s3DestinationProperties.BucketPrefix; v != nil {
		m[names.AttrBucketPrefix] = aws.ToString(v)
	}

	if v := s3DestinationProperties.S3OutputFormatConfig; v != nil {
		m["s3_output_format_config"] = []interface{}{flattenS3OutputFormatConfig(v)}
	}

	return m
}

func flattenS3OutputFormatConfig(s3OutputFormatConfig *types.S3OutputFormatConfig) map[string]interface{} {
	if s3OutputFormatConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := s3OutputFormatConfig.AggregationConfig; v != nil {
		m["aggregation_config"] = []interface{}{flattenAggregationConfig(v)}
	}

	m["file_type"] = s3OutputFormatConfig.FileType

	if v := s3OutputFormatConfig.PrefixConfig; v != nil {
		m["prefix_config"] = []interface{}{flattenPrefixConfig(v)}
	}

	m["preserve_source_data_typing"] = s3OutputFormatConfig.PreserveSourceDataTyping

	return m
}

func flattenSalesforceDestinationProperties(salesforceDestinationProperties *types.SalesforceDestinationProperties) map[string]interface{} {
	if salesforceDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := salesforceDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := salesforceDestinationProperties.IdFieldNames; v != nil {
		m["id_field_names"] = v
	}

	if v := salesforceDestinationProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	m["write_operation_type"] = salesforceDestinationProperties.WriteOperationType

	return m
}

func flattenSAPODataDestinationProperties(SAPODataDestinationProperties *types.SAPODataDestinationProperties) map[string]interface{} {
	if SAPODataDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := SAPODataDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := SAPODataDestinationProperties.IdFieldNames; v != nil {
		m["id_field_names"] = v
	}

	if v := SAPODataDestinationProperties.ObjectPath; v != nil {
		m["object_path"] = aws.ToString(v)
	}

	if v := SAPODataDestinationProperties.SuccessResponseHandlingConfig; v != nil {
		m["success_response_handling_config"] = []interface{}{flattenSuccessResponseHandlingConfig(v)}
	}

	m["write_operation_type"] = SAPODataDestinationProperties.WriteOperationType

	return m
}

func flattenSuccessResponseHandlingConfig(successResponseHandlingConfig *types.SuccessResponseHandlingConfig) map[string]interface{} {
	if successResponseHandlingConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := successResponseHandlingConfig.BucketName; v != nil {
		m[names.AttrBucketName] = aws.ToString(v)
	}

	if v := successResponseHandlingConfig.BucketPrefix; v != nil {
		m[names.AttrBucketPrefix] = aws.ToString(v)
	}

	return m
}

func flattenSnowflakeDestinationProperties(snowflakeDestinationProperties *types.SnowflakeDestinationProperties) map[string]interface{} {
	if snowflakeDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := snowflakeDestinationProperties.BucketPrefix; v != nil {
		m[names.AttrBucketPrefix] = aws.ToString(v)
	}

	if v := snowflakeDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := snowflakeDestinationProperties.IntermediateBucketName; v != nil {
		m["intermediate_bucket_name"] = aws.ToString(v)
	}

	if v := snowflakeDestinationProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenUpsolverDestinationProperties(upsolverDestinationProperties *types.UpsolverDestinationProperties) map[string]interface{} {
	if upsolverDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := upsolverDestinationProperties.BucketName; v != nil {
		m[names.AttrBucketName] = aws.ToString(v)
	}

	if v := upsolverDestinationProperties.BucketPrefix; v != nil {
		m[names.AttrBucketPrefix] = aws.ToString(v)
	}

	if v := upsolverDestinationProperties.S3OutputFormatConfig; v != nil {
		m["s3_output_format_config"] = []interface{}{flattenUpsolverS3OutputFormatConfig(v)}
	}

	return m
}

func flattenUpsolverS3OutputFormatConfig(upsolverS3OutputFormatConfig *types.UpsolverS3OutputFormatConfig) map[string]interface{} {
	if upsolverS3OutputFormatConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := upsolverS3OutputFormatConfig.AggregationConfig; v != nil {
		m["aggregation_config"] = []interface{}{flattenAggregationConfig(v)}
	}

	m["file_type"] = upsolverS3OutputFormatConfig.FileType

	if v := upsolverS3OutputFormatConfig.PrefixConfig; v != nil {
		m["prefix_config"] = []interface{}{flattenPrefixConfig(v)}
	}

	return m
}

func flattenZendeskDestinationProperties(zendeskDestinationProperties *types.ZendeskDestinationProperties) map[string]interface{} {
	if zendeskDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := zendeskDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := zendeskDestinationProperties.IdFieldNames; v != nil {
		m["id_field_names"] = v
	}

	if v := zendeskDestinationProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	m["write_operation_type"] = zendeskDestinationProperties.WriteOperationType

	return m
}

func flattenSourceFlowConfig(sourceFlowConfig *types.SourceFlowConfig) map[string]interface{} {
	if sourceFlowConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := sourceFlowConfig.ApiVersion; v != nil {
		m["api_version"] = aws.ToString(v)
	}

	if v := sourceFlowConfig.ConnectorProfileName; v != nil {
		m["connector_profile_name"] = aws.ToString(v)
	}

	m["connector_type"] = sourceFlowConfig.ConnectorType

	if v := sourceFlowConfig.IncrementalPullConfig; v != nil {
		m["incremental_pull_config"] = []interface{}{flattenIncrementalPullConfig(v)}
	}

	if v := sourceFlowConfig.SourceConnectorProperties; v != nil {
		m["source_connector_properties"] = []interface{}{flattenSourceConnectorProperties(v)}
	}

	return m
}

func flattenIncrementalPullConfig(incrementalPullConfig *types.IncrementalPullConfig) map[string]interface{} {
	if incrementalPullConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := incrementalPullConfig.DatetimeTypeFieldName; v != nil {
		m["datetime_type_field_name"] = aws.ToString(v)
	}

	return m
}

func flattenSourceConnectorProperties(sourceConnectorProperties *types.SourceConnectorProperties) map[string]interface{} {
	if sourceConnectorProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := sourceConnectorProperties.Amplitude; v != nil {
		m["amplitude"] = []interface{}{flattenAmplitudeSourceProperties(v)}
	}

	if v := sourceConnectorProperties.CustomConnector; v != nil {
		m["custom_connector"] = []interface{}{flattenCustomConnectorSourceProperties(v)}
	}

	if v := sourceConnectorProperties.Datadog; v != nil {
		m["datadog"] = []interface{}{flattenDatadogSourceProperties(v)}
	}

	if v := sourceConnectorProperties.Dynatrace; v != nil {
		m["dynatrace"] = []interface{}{flattenDynatraceSourceProperties(v)}
	}

	if v := sourceConnectorProperties.GoogleAnalytics; v != nil {
		m["google_analytics"] = []interface{}{flattenGoogleAnalyticsSourceProperties(v)}
	}

	if v := sourceConnectorProperties.InforNexus; v != nil {
		m["infor_nexus"] = []interface{}{flattenInforNexusSourceProperties(v)}
	}

	if v := sourceConnectorProperties.Marketo; v != nil {
		m["marketo"] = []interface{}{flattenMarketoSourceProperties(v)}
	}

	if v := sourceConnectorProperties.S3; v != nil {
		m["s3"] = []interface{}{flattenS3SourceProperties(v)}
	}

	if v := sourceConnectorProperties.Salesforce; v != nil {
		m["salesforce"] = []interface{}{flattenSalesforceSourceProperties(v)}
	}

	if v := sourceConnectorProperties.SAPOData; v != nil {
		m["sapo_data"] = []interface{}{flattenSAPODataSourceProperties(v)}
	}

	if v := sourceConnectorProperties.ServiceNow; v != nil {
		m["service_now"] = []interface{}{flattenServiceNowSourceProperties(v)}
	}

	if v := sourceConnectorProperties.Singular; v != nil {
		m["singular"] = []interface{}{flattenSingularSourceProperties(v)}
	}

	if v := sourceConnectorProperties.Slack; v != nil {
		m["slack"] = []interface{}{flattenSlackSourceProperties(v)}
	}

	if v := sourceConnectorProperties.Trendmicro; v != nil {
		m["trendmicro"] = []interface{}{flattenTrendmicroSourceProperties(v)}
	}

	if v := sourceConnectorProperties.Veeva; v != nil {
		m["veeva"] = []interface{}{flattenVeevaSourceProperties(v)}
	}

	if v := sourceConnectorProperties.Zendesk; v != nil {
		m["zendesk"] = []interface{}{flattenZendeskSourceProperties(v)}
	}

	return m
}

func flattenAmplitudeSourceProperties(amplitudeSourceProperties *types.AmplitudeSourceProperties) map[string]interface{} {
	if amplitudeSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := amplitudeSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenCustomConnectorSourceProperties(customConnectorSourceProperties *types.CustomConnectorSourceProperties) map[string]interface{} {
	if customConnectorSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := customConnectorSourceProperties.CustomProperties; v != nil {
		m["custom_properties"] = v
	}

	if v := customConnectorSourceProperties.EntityName; v != nil {
		m["entity_name"] = aws.ToString(v)
	}

	return m
}

func flattenDatadogSourceProperties(datadogSourceProperties *types.DatadogSourceProperties) map[string]interface{} {
	if datadogSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := datadogSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenDynatraceSourceProperties(dynatraceSourceProperties *types.DynatraceSourceProperties) map[string]interface{} {
	if dynatraceSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := dynatraceSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenGoogleAnalyticsSourceProperties(googleAnalyticsSourceProperties *types.GoogleAnalyticsSourceProperties) map[string]interface{} {
	if googleAnalyticsSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := googleAnalyticsSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenInforNexusSourceProperties(inforNexusSourceProperties *types.InforNexusSourceProperties) map[string]interface{} {
	if inforNexusSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := inforNexusSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenMarketoSourceProperties(marketoSourceProperties *types.MarketoSourceProperties) map[string]interface{} {
	if marketoSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := marketoSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenS3SourceProperties(s3SourceProperties *types.S3SourceProperties) map[string]interface{} {
	if s3SourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := s3SourceProperties.BucketName; v != nil {
		m[names.AttrBucketName] = aws.ToString(v)
	}

	if v := s3SourceProperties.BucketPrefix; v != nil {
		m[names.AttrBucketPrefix] = aws.ToString(v)
	}

	if v := s3SourceProperties.S3InputFormatConfig; v != nil {
		m["s3_input_format_config"] = []interface{}{flattenS3InputFormatConfig(v)}
	}

	return m
}

func flattenS3InputFormatConfig(s3InputFormatConfig *types.S3InputFormatConfig) map[string]interface{} {
	if s3InputFormatConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["s3_input_file_type"] = s3InputFormatConfig.S3InputFileType

	return m
}

func flattenSalesforceSourceProperties(salesforceSourceProperties *types.SalesforceSourceProperties) map[string]interface{} {
	if salesforceSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["enable_dynamic_field_update"] = salesforceSourceProperties.EnableDynamicFieldUpdate
	m["include_deleted_records"] = salesforceSourceProperties.IncludeDeletedRecords

	if v := salesforceSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenSAPODataSourceProperties(sapoDataSourceProperties *types.SAPODataSourceProperties) map[string]interface{} {
	if sapoDataSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := sapoDataSourceProperties.ObjectPath; v != nil {
		m["object_path"] = aws.ToString(v)
	}

	return m
}

func flattenServiceNowSourceProperties(serviceNowSourceProperties *types.ServiceNowSourceProperties) map[string]interface{} {
	if serviceNowSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := serviceNowSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenSingularSourceProperties(singularSourceProperties *types.SingularSourceProperties) map[string]interface{} {
	if singularSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := singularSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenSlackSourceProperties(slackSourceProperties *types.SlackSourceProperties) map[string]interface{} {
	if slackSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := slackSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenTrendmicroSourceProperties(trendmicroSourceProperties *types.TrendmicroSourceProperties) map[string]interface{} {
	if trendmicroSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := trendmicroSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenVeevaSourceProperties(veevaSourceProperties *types.VeevaSourceProperties) map[string]interface{} {
	if veevaSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := veevaSourceProperties.DocumentType; v != nil {
		m["document_type"] = aws.ToString(v)
	}

	m["include_all_versions"] = veevaSourceProperties.IncludeAllVersions
	m["include_renditions"] = veevaSourceProperties.IncludeRenditions
	m["include_source_files"] = veevaSourceProperties.IncludeSourceFiles

	if v := veevaSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenZendeskSourceProperties(zendeskSourceProperties *types.ZendeskSourceProperties) map[string]interface{} {
	if zendeskSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := zendeskSourceProperties.Object; v != nil {
		m["object"] = aws.ToString(v)
	}

	return m
}

func flattenTasks(tasks []types.Task) []interface{} {
	if len(tasks) == 0 {
		return nil
	}

	var l []interface{}

	t := slices.IndexFunc(tasks, func(t types.Task) bool {
		return t.TaskType == types.TaskTypeMapAll
	})

	if t != -1 {
		l = append(l, flattenTask(tasks[t]))
		return l
	}

	for _, task := range tasks {
		l = append(l, flattenTask(task))
	}

	return l
}

func flattenTask(task types.Task) map[string]interface{} {
	if itypes.IsZero(&task) {
		return nil
	}

	m := map[string]interface{}{}

	if v := task.ConnectorOperator; v != nil {
		m["connector_operator"] = []interface{}{flattenConnectorOperator(v)}
	}

	if v := task.DestinationField; v != nil {
		m["destination_field"] = aws.ToString(v)
	}

	if v := task.SourceFields; v != nil {
		m["source_fields"] = v
	}

	if v := task.TaskProperties; v != nil {
		m["task_properties"] = flex.FlattenStringValueMap(v)
	}

	m["task_type"] = task.TaskType

	return m
}

func flattenConnectorOperator(connectorOperator *types.ConnectorOperator) map[string]interface{} {
	if connectorOperator == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["amplitude"] = connectorOperator.Amplitude
	m["custom_connector"] = connectorOperator.CustomConnector
	m["datadog"] = connectorOperator.Datadog
	m["dynatrace"] = connectorOperator.Dynatrace
	m["google_analytics"] = connectorOperator.GoogleAnalytics
	m["infor_nexus"] = connectorOperator.InforNexus
	m["marketo"] = connectorOperator.Marketo
	m["s3"] = connectorOperator.S3
	m["salesforce"] = connectorOperator.Salesforce
	m["sapo_data"] = connectorOperator.SAPOData
	m["service_now"] = connectorOperator.ServiceNow
	m["singular"] = connectorOperator.Singular
	m["slack"] = connectorOperator.Slack
	m["trendmicro"] = connectorOperator.Trendmicro
	m["veeva"] = connectorOperator.Veeva
	m["zendesk"] = connectorOperator.Zendesk

	return m
}

func flattenTriggerConfig(triggerConfig *types.TriggerConfig) map[string]interface{} {
	if triggerConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := triggerConfig.TriggerProperties; v != nil {
		m["trigger_properties"] = []interface{}{flattenTriggerProperties(v)}
	}

	m["trigger_type"] = triggerConfig.TriggerType

	return m
}

func flattenTriggerProperties(triggerProperties *types.TriggerProperties) map[string]interface{} {
	if triggerProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := triggerProperties.Scheduled; v != nil {
		m["scheduled"] = []interface{}{flattenScheduled(v)}
	}

	return m
}

func flattenScheduled(scheduledTriggerProperties *types.ScheduledTriggerProperties) map[string]interface{} {
	if scheduledTriggerProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	m["data_pull_mode"] = scheduledTriggerProperties.DataPullMode

	if v := scheduledTriggerProperties.FirstExecutionFrom; v != nil {
		m["first_execution_from"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := scheduledTriggerProperties.ScheduleEndTime; v != nil {
		m["schedule_end_time"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := scheduledTriggerProperties.ScheduleExpression; v != nil {
		m[names.AttrScheduleExpression] = aws.ToString(v)
	}

	if v := scheduledTriggerProperties.ScheduleOffset; v != nil {
		m["schedule_offset"] = aws.ToInt64(v)
	}

	if v := scheduledTriggerProperties.ScheduleStartTime; v != nil {
		m["schedule_start_time"] = aws.ToTime(v).Format(time.RFC3339)
	}

	if v := scheduledTriggerProperties.Timezone; v != nil {
		m["timezone"] = aws.ToString(v)
	}

	return m
}
