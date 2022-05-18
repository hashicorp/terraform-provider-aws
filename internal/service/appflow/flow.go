package appflow

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appflow"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFlow() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFlowCreate,
		ReadWithoutTimeout:   resourceFlowRead,
		UpdateWithoutTimeout: resourceFlowUpdate,
		DeleteWithoutTimeout: resourceFlowDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`[a-zA-Z0-9][\w!@#.-]+`), "must contain only alphanumeric, exclamation point (!), at sign (@), number sign (#), period (.), and hyphen (-) characters"), validation.StringLenBetween(1, 256)),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[\w!@#\-.?,\s]*`), "must contain only alphanumeric, underscore (_), exclamation point (!), at sign (@), number sign (#), hyphen (-), period (.), question mark (?), comma (,), and whitespace characters"),
			},
			"destination_flow_config": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_version": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 256)),
						},
						"connector_profile_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`[\w\/!@#+=.-]+`), "must contain only alphanumeric, underscore (_), forward slash (/), exclamation point (!), at sign (@), number sign (#), plus sign (+), equals sign (=), period (.), and hyphen (-) characters"), validation.StringLenBetween(1, 256)),
						},
						"connector_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appflow.ConnectorType_Values(), false),
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
													ValidateDiagFunc: allDiagFunc(
														validation.MapKeyLenBetween(1, 128),
														validation.MapKeyMatch(regexp.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
													),
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 2048)),
													},
												},
												"entity_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 1024)),
												},
												"error_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
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
														ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 128)),
													},
												},
												"write_operation_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(appflow.WriteOperationType_Values(), false),
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
												"domain_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 64)),
												},
												"object_type_name": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 255)),
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
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
												"bucket_prefix": {
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
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"s3": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"s3_output_format_config": {
													Type:     schema.TypeList,
													Optional: true,
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
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.AggregationType_Values(), false),
																		},
																	},
																},
															},
															"file_type": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringInSlice(appflow.FileType_Values(), false),
															},
															"prefix_config": {
																Type:     schema.TypeList,
																Optional: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"prefix_format": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.PrefixFormat_Values(), false),
																		},
																		"prefix_type": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.PrefixType_Values(), false),
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
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
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
														ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 128)),
													},
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
												"write_operation_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(appflow.WriteOperationType_Values(), false),
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
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
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
														ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 128)),
													},
												},
												"object_path": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
												"success_response_handling_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 512),
															},
														},
													},
												},
												"write_operation_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(appflow.WriteOperationType_Values(), false),
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
												"bucket_prefix": {
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
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
												"bucket_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`^(upsolver-appflow)\S*`), "must start with 'upsolver-appflow' and can not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												"bucket_prefix": {
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
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.AggregationType_Values(), false),
																		},
																	},
																},
															},
															"file_type": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringInSlice(appflow.FileType_Values(), false),
															},
															"prefix_config": {
																Type:     schema.TypeList,
																Required: true,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"prefix_format": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringInSlice(appflow.PrefixFormat_Values(), false),
																		},
																		"prefix_type": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: validation.StringInSlice(appflow.PrefixType_Values(), false),
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
															"bucket_name": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
															},
															"bucket_prefix": {
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
														ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 128)),
													},
												},
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
												"write_operation_type": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(appflow.WriteOperationType_Values(), false),
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
			"kms_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`arn:.*:kms:.*:[0-9]+:.*`), "must be a valid ARN of a Key Management Services (KMS) key"),
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
							ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 256)),
						},
						"connector_profile_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`[\w\/!@#+=.-]+`), "must contain only alphanumeric, underscore (_), forward slash (/), exclamation point (!), at sign (@), number sign (#), plus sign (+), equals sign (=), period (.), and hyphen (-) characters"), validation.StringLenBetween(1, 256)),
						},
						"connector_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appflow.ConnectorType_Values(), false),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateDiagFunc: allDiagFunc(
														validation.MapKeyLenBetween(1, 128),
														validation.MapKeyMatch(regexp.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"),
													),
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 2048)),
													},
												},
												"entity_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 1024)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
												},
											},
										},
									},
									"s3": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"bucket_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(3, 63)),
												},
												"bucket_prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 512),
												},
												"s3_input_format_config": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"s3_input_file_type": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringInSlice(appflow.S3InputFileType_Values(), false),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
												"object": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`[\s\w_-]+`), "must contain only alphanumeric, underscore (_), and hyphen (-) characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
													ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(1, 512)),
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
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.AmplitudeConnectorOperator_Values(), false),
									},
									"custom_connector": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.Operator_Values(), false),
									},
									"datadog": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.DatadogConnectorOperator_Values(), false),
									},
									"dynatrace": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.DynatraceConnectorOperator_Values(), false),
									},
									"google_analytics": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.GoogleAnalyticsConnectorOperator_Values(), false),
									},
									"infor_nexus": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.InforNexusConnectorOperator_Values(), false),
									},
									"marketo": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.MarketoConnectorOperator_Values(), false),
									},
									"s3": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.S3ConnectorOperator_Values(), false),
									},
									"salesforce": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.SalesforceConnectorOperator_Values(), false),
									},
									"sapo_data": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.SAPODataConnectorOperator_Values(), false),
									},
									"service_now": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.ServiceNowConnectorOperator_Values(), false),
									},
									"singular": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.SingularConnectorOperator_Values(), false),
									},
									"slack": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.SlackConnectorOperator_Values(), false),
									},
									"trendmicro": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.TrendmicroConnectorOperator_Values(), false),
									},
									"veeva": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.VeevaConnectorOperator_Values(), false),
									},
									"zendesk": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice(appflow.ZendeskConnectorOperator_Values(), false),
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
							Required: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringLenBetween(0, 2048),
							},
						},
						"task_properties": {
							Type:         schema.TypeMap,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(appflow.OperatorPropertiesKeys_Values(), false),
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.All(validation.StringMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters"), validation.StringLenBetween(0, 2048)),
							},
						},
						"task_type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appflow.TaskType_Values(), false),
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
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringInSlice(appflow.DataPullMode_Values(), false),
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
												"schedule_expression": {
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
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(appflow.TriggerType_Values(), false),
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFlowCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppFlowConn

	in := &appflow.CreateFlowInput{
		FlowName:                  aws.String(d.Get("name").(string)),
		DestinationFlowConfigList: expandDestinationFlowConfigs(d.Get("destination_flow_config").(*schema.Set).List()),
		SourceFlowConfig:          expandSourceFlowConfig(d.Get("source_flow_config").([]interface{})[0].(map[string]interface{})),
		Tasks:                     expandTasks(d.Get("task").(*schema.Set).List()),
		TriggerConfig:             expandTriggerConfig(d.Get("trigger_config").([]interface{})[0].(map[string]interface{})),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("kms_arn"); ok {
		in.KmsArn = aws.String(v.(string))
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	out, err := conn.CreateFlowWithContext(ctx, in)

	if err != nil {
		return diag.Errorf("creating Appflow Flow (%s): %s", d.Get("name").(string), err)
	}

	if out == nil || out.FlowArn == nil {
		return diag.Errorf("creating Appflow Flow (%s): empty output", d.Get("name").(string))
	}

	d.SetId(aws.StringValue(out.FlowArn))

	return resourceFlowRead(ctx, d, meta)
}

func resourceFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppFlowConn

	out, err := FindFlowByArn(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppFlow Flow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("finding AppFlow Flow (%s): %s", d.Id(), err)
	}

	in := &appflow.DescribeFlowInput{
		FlowName: out.FlowName,
	}

	out2, err := conn.DescribeFlowWithContext(ctx, in)

	if err != nil {
		return diag.Errorf("reading AppFlow Flow (%s): %s", d.Id(), err)
	}

	d.Set("name", out.FlowName)
	d.Set("arn", out2.FlowArn)
	d.Set("description", out2.Description)

	if err := d.Set("destination_flow_config", flattenDestinationFlowConfigs(out2.DestinationFlowConfigList)); err != nil {
		return diag.Errorf("error setting destination_flow_config: %s", err)
	}

	d.Set("kms_arn", out2.KmsArn)

	if out2.SourceFlowConfig != nil {
		if err := d.Set("source_flow_config", []interface{}{flattenSourceFlowConfig(out2.SourceFlowConfig)}); err != nil {
			return diag.Errorf("error setting source_flow_config: %s", err)
		}
	} else {
		d.Set("source_flow_config", nil)
	}

	if err := d.Set("task", flattenTasks(out2.Tasks)); err != nil {
		return diag.Errorf("error setting task: %s", err)
	}

	if out2.TriggerConfig != nil {
		if err := d.Set("trigger_config", []interface{}{flattenTriggerConfig(out2.TriggerConfig)}); err != nil {
			return diag.Errorf("error setting trigger_config: %s", err)
		}
	} else {
		d.Set("trigger_config", nil)
	}

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return diag.Errorf("listing tags for AppFlow Flow (%s): %s", d.Id(), err)
	}

	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppFlowConn

	in := &appflow.UpdateFlowInput{
		FlowName:                  aws.String(d.Get("name").(string)),
		DestinationFlowConfigList: expandDestinationFlowConfigs(d.Get("destination_flow_config").(*schema.Set).List()),
		SourceFlowConfig:          expandSourceFlowConfig(d.Get("source_flow_config").([]interface{})[0].(map[string]interface{})),
		Tasks:                     expandTasks(d.Get("task").(*schema.Set).List()),
		TriggerConfig:             expandTriggerConfig(d.Get("trigger_config").([]interface{})[0].(map[string]interface{})),
	}

	if d.HasChange("description") {
		in.Description = aws.String(d.Get("description").(string))
	}

	log.Printf("[DEBUG] Updating AppFlow Flow (%s): %#v", d.Id(), in)
	_, err := conn.UpdateFlow(in)

	if err != nil {
		return diag.Errorf("updating AppFlow Flow (%s): %s", d.Id(), err)
	}

	arn := d.Get("arn").(string)

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, arn, o, n); err != nil {
			return diag.Errorf("error updating tags: %s", err)
		}
	}

	return resourceFlowRead(ctx, d, meta)
}

func resourceFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppFlowConn

	out, _ := FindFlowByArn(ctx, conn, d.Id())

	log.Printf("[INFO] Deleting AppFlow Flow %s", d.Id())

	_, err := conn.DeleteFlowWithContext(ctx, &appflow.DeleteFlowInput{
		FlowName: out.FlowName,
	})

	if tfawserr.ErrCodeEquals(err, appflow.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting AppFlow Flow (%s): %s", d.Id(), err)
	}

	if err := FlowDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for AppFlow Flow (%s) to be deleted: %s", d.Id(), err)
	}

	return nil
}

func expandErrorHandlingConfig(tfMap map[string]interface{}) *appflow.ErrorHandlingConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.ErrorHandlingConfig{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap["bucket_prefix"].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["fail_on_first_destination_error"].(bool); ok {
		a.FailOnFirstDestinationError = aws.Bool(v)
	}

	return a
}

func expandAggregationConfig(tfMap map[string]interface{}) *appflow.AggregationConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.AggregationConfig{}

	if v, ok := tfMap["aggregation_type"].(string); ok && v != "" {
		a.AggregationType = aws.String(v)
	}

	return a
}

func expandPrefixConfig(tfMap map[string]interface{}) *appflow.PrefixConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.PrefixConfig{}

	if v, ok := tfMap["prefix_format"].(string); ok && v != "" {
		a.PrefixFormat = aws.String(v)
	}

	if v, ok := tfMap["prefix_type"].(string); ok && v != "" {
		a.PrefixType = aws.String(v)
	}

	return a
}

func expandDestinationFlowConfigs(tfList []interface{}) []*appflow.DestinationFlowConfig {
	if len(tfList) == 0 {
		return nil
	}

	var s []*appflow.DestinationFlowConfig

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandDestinationFlowConfig(m)

		if a == nil {
			continue
		}

		s = append(s, a)
	}

	return s
}

func expandDestinationFlowConfig(tfMap map[string]interface{}) *appflow.DestinationFlowConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.DestinationFlowConfig{}

	if v, ok := tfMap["api_version"].(string); ok && v != "" {
		a.ApiVersion = aws.String(v)
	}

	if v, ok := tfMap["connector_profile_name"].(string); ok && v != "" {
		a.ConnectorProfileName = aws.String(v)
	}

	if v, ok := tfMap["connector_type"].(string); ok && v != "" {
		a.ConnectorType = aws.String(v)
	}

	if v, ok := tfMap["destination_connector_properties"].([]interface{}); ok && len(v) > 0 {
		a.DestinationConnectorProperties = expandDestinationConnectorProperties(v[0].(map[string]interface{}))
	}

	return a
}

func expandDestinationConnectorProperties(tfMap map[string]interface{}) *appflow.DestinationConnectorProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.DestinationConnectorProperties{}

	if v, ok := tfMap["custom_connector"].([]interface{}); ok && len(v) > 0 {
		a.CustomConnector = expandCustomConnectorDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["customer_profiles"].([]interface{}); ok && len(v) > 0 {
		a.CustomerProfiles = expandCustomerProfilesDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["event_bridge"].([]interface{}); ok && len(v) > 0 {
		a.EventBridge = expandEventBridgeDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["honeycode"].([]interface{}); ok && len(v) > 0 {
		a.Honeycode = expandHoneycodeDestinationProperties(v[0].(map[string]interface{}))
	}

	// API reference does not list valid attributes for LookoutMetricsDestinationProperties
	// https://docs.aws.amazon.com/appflow/1.0/APIReference/API_LookoutMetricsDestinationProperties.html
	if v, ok := tfMap["lookout_metrics"].([]interface{}); ok && len(v) > 0 {
		a.LookoutMetrics = v[0].(*appflow.LookoutMetricsDestinationProperties)
	}

	if v, ok := tfMap["marketo"].([]interface{}); ok && len(v) > 0 {
		a.Marketo = expandMarketoDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["redshift"].([]interface{}); ok && len(v) > 0 {
		a.Redshift = expandRedshiftDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		a.S3 = expandS3DestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["salesforce"].([]interface{}); ok && len(v) > 0 {
		a.Salesforce = expandSalesforceDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sapo_data"].([]interface{}); ok && len(v) > 0 {
		a.SAPOData = expandSAPODataDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["snowflake"].([]interface{}); ok && len(v) > 0 {
		a.Snowflake = expandSnowflakeDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["upsolver"].([]interface{}); ok && len(v) > 0 {
		a.Upsolver = expandUpsolverDestinationProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["zendesk"].([]interface{}); ok && len(v) > 0 {
		a.Zendesk = expandZendeskDestinationProperties(v[0].(map[string]interface{}))
	}

	return a
}

func expandCustomConnectorDestinationProperties(tfMap map[string]interface{}) *appflow.CustomConnectorDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.CustomConnectorDestinationProperties{}

	if v, ok := tfMap["custom_properties"].(map[string]interface{}); ok && len(v) > 0 {
		a.CustomProperties = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["entity_name"].(string); ok && v != "" {
		a.EntityName = aws.String(v)
	}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["id_field_names"].([]interface{}); ok && len(v) > 0 {
		a.IdFieldNames = flex.ExpandStringList(v)
	}

	if v, ok := tfMap["write_operation_type"].(string); ok && v != "" {
		a.WriteOperationType = aws.String(v)
	}

	return a
}

func expandCustomerProfilesDestinationProperties(tfMap map[string]interface{}) *appflow.CustomerProfilesDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.CustomerProfilesDestinationProperties{}

	if v, ok := tfMap["domain_name"].(string); ok && v != "" {
		a.DomainName = aws.String(v)
	}

	if v, ok := tfMap["object_type_name"].(string); ok && v != "" {
		a.ObjectTypeName = aws.String(v)
	}

	return a
}

func expandEventBridgeDestinationProperties(tfMap map[string]interface{}) *appflow.EventBridgeDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.EventBridgeDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandHoneycodeDestinationProperties(tfMap map[string]interface{}) *appflow.HoneycodeDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.HoneycodeDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandMarketoDestinationProperties(tfMap map[string]interface{}) *appflow.MarketoDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.MarketoDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandRedshiftDestinationProperties(tfMap map[string]interface{}) *appflow.RedshiftDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.RedshiftDestinationProperties{}

	if v, ok := tfMap["bucket_prefix"].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 {
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

func expandS3DestinationProperties(tfMap map[string]interface{}) *appflow.S3DestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.S3DestinationProperties{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap["bucket_prefix"].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["s3_output_format_config"].([]interface{}); ok && len(v) > 0 {
		a.S3OutputFormatConfig = expandS3OutputFormatConfig(v[0].(map[string]interface{}))
	}

	return a
}

func expandS3OutputFormatConfig(tfMap map[string]interface{}) *appflow.S3OutputFormatConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.S3OutputFormatConfig{}

	if v, ok := tfMap["aggregation_config"].([]interface{}); ok && len(v) > 0 {
		a.AggregationConfig = expandAggregationConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["file_type"].(string); ok && v != "" {
		a.FileType = aws.String(v)
	}

	if v, ok := tfMap["prefix_config"].([]interface{}); ok && len(v) > 0 {
		a.PrefixConfig = expandPrefixConfig(v[0].(map[string]interface{}))
	}

	return a
}

func expandSalesforceDestinationProperties(tfMap map[string]interface{}) *appflow.SalesforceDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SalesforceDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["id_field_names"].([]interface{}); ok && len(v) > 0 {
		a.IdFieldNames = flex.ExpandStringList(v)
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	if v, ok := tfMap["write_operation_type"].(string); ok && v != "" {
		a.WriteOperationType = aws.String(v)
	}

	return a
}

func expandSAPODataDestinationProperties(tfMap map[string]interface{}) *appflow.SAPODataDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SAPODataDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["id_field_names"].([]interface{}); ok && len(v) > 0 {
		a.IdFieldNames = flex.ExpandStringList(v)
	}

	if v, ok := tfMap["object_path"].(string); ok && v != "" {
		a.ObjectPath = aws.String(v)
	}

	if v, ok := tfMap["success_response_handling_config"].([]interface{}); ok && len(v) > 0 {
		a.SuccessResponseHandlingConfig = expandSuccessResponseHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["write_operation_type"].(string); ok && v != "" {
		a.WriteOperationType = aws.String(v)
	}

	return a
}

func expandSuccessResponseHandlingConfig(tfMap map[string]interface{}) *appflow.SuccessResponseHandlingConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SuccessResponseHandlingConfig{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap["bucket_prefix"].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	return a
}

func expandSnowflakeDestinationProperties(tfMap map[string]interface{}) *appflow.SnowflakeDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SnowflakeDestinationProperties{}

	if v, ok := tfMap["bucket_prefix"].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 {
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

func expandUpsolverDestinationProperties(tfMap map[string]interface{}) *appflow.UpsolverDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.UpsolverDestinationProperties{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap["bucket_prefix"].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["s3_output_format_config"].([]interface{}); ok && len(v) > 0 {
		a.S3OutputFormatConfig = expandUpsolverS3OutputFormatConfig(v[0].(map[string]interface{}))
	}

	return a
}

func expandUpsolverS3OutputFormatConfig(tfMap map[string]interface{}) *appflow.UpsolverS3OutputFormatConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.UpsolverS3OutputFormatConfig{}

	if v, ok := tfMap["aggregation_config"].([]interface{}); ok && len(v) > 0 {
		a.AggregationConfig = expandAggregationConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["file_type"].(string); ok && v != "" {
		a.FileType = aws.String(v)
	}

	if v, ok := tfMap["prefix_config"].([]interface{}); ok && len(v) > 0 {
		a.PrefixConfig = expandPrefixConfig(v[0].(map[string]interface{}))
	}

	return a
}

func expandZendeskDestinationProperties(tfMap map[string]interface{}) *appflow.ZendeskDestinationProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.ZendeskDestinationProperties{}

	if v, ok := tfMap["error_handling_config"].([]interface{}); ok && len(v) > 0 {
		a.ErrorHandlingConfig = expandErrorHandlingConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["id_field_names"].([]interface{}); ok && len(v) > 0 {
		a.IdFieldNames = flex.ExpandStringList(v)
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	if v, ok := tfMap["write_operation_type"].(string); ok && v != "" {
		a.WriteOperationType = aws.String(v)
	}

	return a
}

func expandSourceFlowConfig(tfMap map[string]interface{}) *appflow.SourceFlowConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SourceFlowConfig{}

	if v, ok := tfMap["api_version"].(string); ok && v != "" {
		a.ApiVersion = aws.String(v)
	}

	if v, ok := tfMap["connector_profile_name"].(string); ok && v != "" {
		a.ConnectorProfileName = aws.String(v)
	}

	if v, ok := tfMap["connector_type"].(string); ok && v != "" {
		a.ConnectorType = aws.String(v)
	}

	if v, ok := tfMap["incremental_pull_config"].([]interface{}); ok && len(v) > 0 {
		a.IncrementalPullConfig = expandIncrementalPullConfig(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["source_connector_properties"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		a.SourceConnectorProperties = expandSourceConnectorProperties(v[0].(map[string]interface{}))
	}

	return a
}

func expandIncrementalPullConfig(tfMap map[string]interface{}) *appflow.IncrementalPullConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.IncrementalPullConfig{}

	if v, ok := tfMap["datetime_type_field_name"].(string); ok && v != "" {
		a.DatetimeTypeFieldName = aws.String(v)
	}

	return a
}

func expandSourceConnectorProperties(tfMap map[string]interface{}) *appflow.SourceConnectorProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SourceConnectorProperties{}

	if v, ok := tfMap["amplitude"].([]interface{}); ok && len(v) > 0 {
		a.Amplitude = expandAmplitudeSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["custom_connector"].([]interface{}); ok && len(v) > 0 {
		a.CustomConnector = expandCustomConnectorSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["datadog"].([]interface{}); ok && len(v) > 0 {
		a.Datadog = expandDatadogSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["dynatrace"].([]interface{}); ok && len(v) > 0 {
		a.Dynatrace = expandDynatraceSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["google_analytics"].([]interface{}); ok && len(v) > 0 {
		a.GoogleAnalytics = expandGoogleAnalyticsSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["infor_nexus"].([]interface{}); ok && len(v) > 0 {
		a.InforNexus = expandInforNexusSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["marketo"].([]interface{}); ok && len(v) > 0 {
		a.Marketo = expandMarketoSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["s3"].([]interface{}); ok && len(v) > 0 {
		a.S3 = expandS3SourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["sapo_data"].([]interface{}); ok && len(v) > 0 {
		a.SAPOData = expandSAPODataSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["salesforce"].([]interface{}); ok && len(v) > 0 {
		a.Salesforce = expandSalesforceSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["service_now"].([]interface{}); ok && len(v) > 0 {
		a.ServiceNow = expandServiceNowSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["singular"].([]interface{}); ok && len(v) > 0 {
		a.Singular = expandSingularSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["slack"].([]interface{}); ok && len(v) > 0 {
		a.Slack = expandSlackSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["trendmicro"].([]interface{}); ok && len(v) > 0 {
		a.Trendmicro = expandTrendmicroSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["veeva"].([]interface{}); ok && len(v) > 0 {
		a.Veeva = expandVeevaSourceProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["zendesk"].([]interface{}); ok && len(v) > 0 {
		a.Zendesk = expandZendeskSourceProperties(v[0].(map[string]interface{}))
	}

	return a
}

func expandAmplitudeSourceProperties(tfMap map[string]interface{}) *appflow.AmplitudeSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.AmplitudeSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandCustomConnectorSourceProperties(tfMap map[string]interface{}) *appflow.CustomConnectorSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.CustomConnectorSourceProperties{}

	if v, ok := tfMap["custom_properties"].(map[string]interface{}); ok && len(v) > 0 {
		a.CustomProperties = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["entity_name"].(string); ok && v != "" {
		a.EntityName = aws.String(v)
	}

	return a
}

func expandDatadogSourceProperties(tfMap map[string]interface{}) *appflow.DatadogSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.DatadogSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandDynatraceSourceProperties(tfMap map[string]interface{}) *appflow.DynatraceSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.DynatraceSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandGoogleAnalyticsSourceProperties(tfMap map[string]interface{}) *appflow.GoogleAnalyticsSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.GoogleAnalyticsSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandInforNexusSourceProperties(tfMap map[string]interface{}) *appflow.InforNexusSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.InforNexusSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandMarketoSourceProperties(tfMap map[string]interface{}) *appflow.MarketoSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.MarketoSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandS3SourceProperties(tfMap map[string]interface{}) *appflow.S3SourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.S3SourceProperties{}

	if v, ok := tfMap["bucket_name"].(string); ok && v != "" {
		a.BucketName = aws.String(v)
	}

	if v, ok := tfMap["bucket_prefix"].(string); ok && v != "" {
		a.BucketPrefix = aws.String(v)
	}

	if v, ok := tfMap["s3_input_format_config"].([]interface{}); ok && len(v) > 0 {
		a.S3InputFormatConfig = expandS3InputFormatConfig(v[0].(map[string]interface{}))
	}

	return a
}

func expandS3InputFormatConfig(tfMap map[string]interface{}) *appflow.S3InputFormatConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.S3InputFormatConfig{}

	if v, ok := tfMap["s3_input_file_type"].(string); ok && v != "" {
		a.S3InputFileType = aws.String(v)
	}

	return a
}

func expandSalesforceSourceProperties(tfMap map[string]interface{}) *appflow.SalesforceSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SalesforceSourceProperties{}

	if v, ok := tfMap["enable_dynamic_field_update"].(bool); ok {
		a.EnableDynamicFieldUpdate = aws.Bool(v)
	}

	if v, ok := tfMap["include_deleted_records"].(bool); ok {
		a.IncludeDeletedRecords = aws.Bool(v)
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandSAPODataSourceProperties(tfMap map[string]interface{}) *appflow.SAPODataSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SAPODataSourceProperties{}

	if v, ok := tfMap["object_path"].(string); ok && v != "" {
		a.ObjectPath = aws.String(v)
	}

	return a
}

func expandServiceNowSourceProperties(tfMap map[string]interface{}) *appflow.ServiceNowSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.ServiceNowSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandSingularSourceProperties(tfMap map[string]interface{}) *appflow.SingularSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SingularSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandSlackSourceProperties(tfMap map[string]interface{}) *appflow.SlackSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.SlackSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandTrendmicroSourceProperties(tfMap map[string]interface{}) *appflow.TrendmicroSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.TrendmicroSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandVeevaSourceProperties(tfMap map[string]interface{}) *appflow.VeevaSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.VeevaSourceProperties{}

	if v, ok := tfMap["document_type"].(string); ok && v != "" {
		a.DocumentType = aws.String(v)
	}

	if v, ok := tfMap["include_all_versions"].(bool); ok {
		a.IncludeAllVersions = aws.Bool(v)
	}

	if v, ok := tfMap["include_renditions"].(bool); ok {
		a.IncludeRenditions = aws.Bool(v)
	}

	if v, ok := tfMap["include_source_files"].(bool); ok {
		a.IncludeSourceFiles = aws.Bool(v)
	}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandZendeskSourceProperties(tfMap map[string]interface{}) *appflow.ZendeskSourceProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.ZendeskSourceProperties{}

	if v, ok := tfMap["object"].(string); ok && v != "" {
		a.Object = aws.String(v)
	}

	return a
}

func expandTasks(tfList []interface{}) []*appflow.Task {
	if len(tfList) == 0 {
		return nil
	}

	var s []*appflow.Task

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandTask(m)

		if a == nil {
			continue
		}

		s = append(s, a)
	}

	return s
}

func expandTask(tfMap map[string]interface{}) *appflow.Task {
	if tfMap == nil {
		return nil
	}

	a := &appflow.Task{}

	if v, ok := tfMap["connector_operator"].([]interface{}); ok && len(v) > 0 {
		a.ConnectorOperator = expandConnectorOperator(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["destination_field"].(string); ok && v != "" {
		a.DestinationField = aws.String(v)
	}

	if v, ok := tfMap["source_fields"].([]interface{}); ok && len(v) > 0 {
		a.SourceFields = flex.ExpandStringList(v)
	}

	if v, ok := tfMap["task_properties"].(map[string]interface{}); ok && len(v) > 0 {
		a.TaskProperties = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["task_type"].(string); ok && v != "" {
		a.TaskType = aws.String(v)
	}

	return a
}

func expandConnectorOperator(tfMap map[string]interface{}) *appflow.ConnectorOperator {
	if tfMap == nil {
		return nil
	}

	a := &appflow.ConnectorOperator{}

	if v, ok := tfMap["amplitude"].(string); ok && v != "" {
		a.Amplitude = aws.String(v)
	}

	if v, ok := tfMap["custom_connector"].(string); ok && v != "" {
		a.CustomConnector = aws.String(v)
	}

	if v, ok := tfMap["datadog"].(string); ok && v != "" {
		a.Datadog = aws.String(v)
	}

	if v, ok := tfMap["dynatrace"].(string); ok && v != "" {
		a.Dynatrace = aws.String(v)
	}

	if v, ok := tfMap["google_analytics"].(string); ok && v != "" {
		a.GoogleAnalytics = aws.String(v)
	}

	if v, ok := tfMap["infor_nexus"].(string); ok && v != "" {
		a.InforNexus = aws.String(v)
	}

	if v, ok := tfMap["marketo"].(string); ok && v != "" {
		a.Marketo = aws.String(v)
	}

	if v, ok := tfMap["s3"].(string); ok && v != "" {
		a.S3 = aws.String(v)
	}

	if v, ok := tfMap["sapo_data"].(string); ok && v != "" {
		a.SAPOData = aws.String(v)
	}

	if v, ok := tfMap["salesforce"].(string); ok && v != "" {
		a.Salesforce = aws.String(v)
	}

	if v, ok := tfMap["service_now"].(string); ok && v != "" {
		a.ServiceNow = aws.String(v)
	}

	if v, ok := tfMap["singular"].(string); ok && v != "" {
		a.Singular = aws.String(v)
	}

	if v, ok := tfMap["slack"].(string); ok && v != "" {
		a.Slack = aws.String(v)
	}

	if v, ok := tfMap["trendmicro"].(string); ok && v != "" {
		a.Trendmicro = aws.String(v)
	}

	if v, ok := tfMap["veeva"].(string); ok && v != "" {
		a.Veeva = aws.String(v)
	}

	if v, ok := tfMap["zendesk"].(string); ok && v != "" {
		a.Zendesk = aws.String(v)
	}

	return a
}

func expandTriggerConfig(tfMap map[string]interface{}) *appflow.TriggerConfig {
	if tfMap == nil {
		return nil
	}

	a := &appflow.TriggerConfig{}

	if v, ok := tfMap["trigger_properties"].([]interface{}); ok && len(v) > 0 {
		a.TriggerProperties = expandTriggerProperties(v[0].(map[string]interface{}))
	}

	if v, ok := tfMap["trigger_type"].(string); ok && v != "" {
		a.TriggerType = aws.String(v)
	}

	return a
}

func expandTriggerProperties(tfMap map[string]interface{}) *appflow.TriggerProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.TriggerProperties{}

	// Only return TriggerProperties if nested field is non-empty
	if v, ok := tfMap["scheduled"].([]interface{}); ok && len(v) > 0 {
		a.Scheduled = expandScheduledTriggerProperties(v[0].(map[string]interface{}))
		return a
	}

	return nil
}

func expandScheduledTriggerProperties(tfMap map[string]interface{}) *appflow.ScheduledTriggerProperties {
	if tfMap == nil {
		return nil
	}

	a := &appflow.ScheduledTriggerProperties{}

	if v, ok := tfMap["data_pull_mode"].(string); ok && v != "" {
		a.DataPullMode = aws.String(v)
	}

	if v, ok := tfMap["first_execution_from"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		a.FirstExecutionFrom = aws.Time(v)
	}

	if v, ok := tfMap["schedule_end_time"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		a.ScheduleEndTime = aws.Time(v)
	}

	if v, ok := tfMap["schedule_expression"].(string); ok && v != "" {
		a.ScheduleExpression = aws.String(v)
	}

	if v, ok := tfMap["schedule_offset"].(int); ok && v != 0 {
		a.ScheduleOffset = aws.Int64(int64(v))
	}

	if v, ok := tfMap["schedule_start_time"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		a.ScheduleEndTime = aws.Time(v)
	}

	if v, ok := tfMap["timezone"].(string); ok && v != "" {
		a.Timezone = aws.String(v)
	}

	return a
}

func flattenErrorHandlingConfig(errorHandlingConfig *appflow.ErrorHandlingConfig) map[string]interface{} {
	if errorHandlingConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := errorHandlingConfig.BucketName; v != nil {
		m["bucket_name"] = aws.StringValue(v)
	}

	if v := errorHandlingConfig.BucketPrefix; v != nil {
		m["bucket_prefix"] = aws.StringValue(v)
	}

	if v := errorHandlingConfig.FailOnFirstDestinationError; v != nil {
		m["fail_on_first_destination_error"] = aws.BoolValue(v)
	}

	return m
}

func flattenPrefixConfig(prefixConfig *appflow.PrefixConfig) map[string]interface{} {
	if prefixConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := prefixConfig.PrefixFormat; v != nil {
		m["prefix_format"] = aws.StringValue(v)
	}

	if v := prefixConfig.PrefixType; v != nil {
		m["prefix_type"] = aws.StringValue(v)
	}

	return m
}

func flattenAggregationConfig(aggregationConfig *appflow.AggregationConfig) map[string]interface{} {
	if aggregationConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := aggregationConfig.AggregationType; v != nil {
		m["aggregation_type"] = aws.StringValue(v)
	}

	return m
}

func flattenDestinationFlowConfigs(destinationFlowConfigs []*appflow.DestinationFlowConfig) []interface{} {
	if len(destinationFlowConfigs) == 0 {
		return nil
	}

	var l []interface{}

	for _, destinationFlowConfig := range destinationFlowConfigs {
		if destinationFlowConfig == nil {
			continue
		}

		l = append(l, flattenDestinationFlowConfig(destinationFlowConfig))
	}

	return l
}

func flattenDestinationFlowConfig(destinationFlowConfig *appflow.DestinationFlowConfig) map[string]interface{} {
	if destinationFlowConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := destinationFlowConfig.ApiVersion; v != nil {
		m["api_version"] = aws.StringValue(v)
	}

	if v := destinationFlowConfig.ConnectorProfileName; v != nil {
		m["connector_profile_name"] = aws.StringValue(v)
	}

	if v := destinationFlowConfig.ConnectorType; v != nil {
		m["connector_type"] = aws.StringValue(v)
	}

	if v := destinationFlowConfig.DestinationConnectorProperties; v != nil {
		m["destination_connector_properties"] = []interface{}{flattenDestinationConnectorProperties(v)}
	}

	return m
}

func flattenDestinationConnectorProperties(destinationConnectorProperties *appflow.DestinationConnectorProperties) map[string]interface{} {
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

func flattenCustomConnectorDestinationProperties(customConnectorDestinationProperties *appflow.CustomConnectorDestinationProperties) map[string]interface{} {
	if customConnectorDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := customConnectorDestinationProperties.CustomProperties; v != nil {
		m["custom_properties"] = aws.StringValueMap(v)
	}

	if v := customConnectorDestinationProperties.EntityName; v != nil {
		m["entity_name"] = aws.StringValue(v)
	}

	if v := customConnectorDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := customConnectorDestinationProperties.IdFieldNames; v != nil {
		m["id_field_names"] = aws.StringValueSlice(v)
	}

	if v := customConnectorDestinationProperties.WriteOperationType; v != nil {
		m["write_operation_type"] = aws.StringValue(v)
	}

	return m
}

func flattenCustomerProfilesDestinationProperties(customerProfilesDestinationProperties *appflow.CustomerProfilesDestinationProperties) map[string]interface{} {
	if customerProfilesDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := customerProfilesDestinationProperties.DomainName; v != nil {
		m["domain_name"] = aws.StringValue(v)
	}

	if v := customerProfilesDestinationProperties.ObjectTypeName; v != nil {
		m["object_type_name"] = aws.StringValue(v)
	}

	return m
}

func flattenEventBridgeDestinationProperties(eventBridgeDestinationProperties *appflow.EventBridgeDestinationProperties) map[string]interface{} {
	if eventBridgeDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := eventBridgeDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := eventBridgeDestinationProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenHoneycodeDestinationProperties(honeycodeDestinationProperties *appflow.HoneycodeDestinationProperties) map[string]interface{} {
	if honeycodeDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := honeycodeDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := honeycodeDestinationProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenMarketoDestinationProperties(marketoDestinationProperties *appflow.MarketoDestinationProperties) map[string]interface{} {
	if marketoDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := marketoDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := marketoDestinationProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenRedshiftDestinationProperties(redshiftDestinationProperties *appflow.RedshiftDestinationProperties) map[string]interface{} {
	if redshiftDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := redshiftDestinationProperties.BucketPrefix; v != nil {
		m["bucket_prefix"] = aws.StringValue(v)
	}

	if v := redshiftDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := redshiftDestinationProperties.IntermediateBucketName; v != nil {
		m["intermediate_bucket_name"] = aws.StringValue(v)
	}

	if v := redshiftDestinationProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenS3DestinationProperties(s3DestinationProperties *appflow.S3DestinationProperties) map[string]interface{} {
	if s3DestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := s3DestinationProperties.BucketName; v != nil {
		m["bucket_name"] = aws.StringValue(v)
	}

	if v := s3DestinationProperties.BucketPrefix; v != nil {
		m["bucket_prefix"] = aws.StringValue(v)
	}

	if v := s3DestinationProperties.S3OutputFormatConfig; v != nil {
		m["s3_output_format_config"] = []interface{}{flattenS3OutputFormatConfig(v)}
	}

	return m
}

func flattenS3OutputFormatConfig(s3OutputFormatConfig *appflow.S3OutputFormatConfig) map[string]interface{} {
	if s3OutputFormatConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := s3OutputFormatConfig.AggregationConfig; v != nil {
		m["aggregation_config"] = []interface{}{flattenAggregationConfig(v)}
	}

	if v := s3OutputFormatConfig.FileType; v != nil {
		m["file_type"] = aws.StringValue(v)
	}

	if v := s3OutputFormatConfig.PrefixConfig; v != nil {
		m["prefix_config"] = []interface{}{flattenPrefixConfig(v)}
	}

	return m
}

func flattenSalesforceDestinationProperties(salesforceDestinationProperties *appflow.SalesforceDestinationProperties) map[string]interface{} {
	if salesforceDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := salesforceDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := salesforceDestinationProperties.IdFieldNames; v != nil {
		m["id_field_names"] = aws.StringValueSlice(v)
	}

	if v := salesforceDestinationProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	if v := salesforceDestinationProperties.WriteOperationType; v != nil {
		m["write_operation_type"] = aws.StringValue(v)
	}

	return m
}

func flattenSAPODataDestinationProperties(SAPODataDestinationProperties *appflow.SAPODataDestinationProperties) map[string]interface{} {
	if SAPODataDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := SAPODataDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := SAPODataDestinationProperties.IdFieldNames; v != nil {
		m["id_field_names"] = aws.StringValueSlice(v)
	}

	if v := SAPODataDestinationProperties.ObjectPath; v != nil {
		m["object_path"] = aws.StringValue(v)
	}

	if v := SAPODataDestinationProperties.SuccessResponseHandlingConfig; v != nil {
		m["success_response_handling_config"] = []interface{}{flattenSuccessResponseHandlingConfig(v)}
	}

	if v := SAPODataDestinationProperties.WriteOperationType; v != nil {
		m["write_operation_type"] = aws.StringValue(v)
	}

	return m
}

func flattenSuccessResponseHandlingConfig(successResponseHandlingConfig *appflow.SuccessResponseHandlingConfig) map[string]interface{} {
	if successResponseHandlingConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := successResponseHandlingConfig.BucketName; v != nil {
		m["bucket_name"] = aws.StringValue(v)
	}

	if v := successResponseHandlingConfig.BucketPrefix; v != nil {
		m["bucket_prefix"] = aws.StringValue(v)
	}

	return m
}

func flattenSnowflakeDestinationProperties(snowflakeDestinationProperties *appflow.SnowflakeDestinationProperties) map[string]interface{} {
	if snowflakeDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := snowflakeDestinationProperties.BucketPrefix; v != nil {
		m["bucket_prefix"] = aws.StringValue(v)
	}

	if v := snowflakeDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := snowflakeDestinationProperties.IntermediateBucketName; v != nil {
		m["intermediate_bucket_name"] = aws.StringValue(v)
	}

	if v := snowflakeDestinationProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenUpsolverDestinationProperties(upsolverDestinationProperties *appflow.UpsolverDestinationProperties) map[string]interface{} {
	if upsolverDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := upsolverDestinationProperties.BucketName; v != nil {
		m["bucket_name"] = aws.StringValue(v)
	}

	if v := upsolverDestinationProperties.BucketPrefix; v != nil {
		m["bucket_prefix"] = aws.StringValue(v)
	}

	if v := upsolverDestinationProperties.S3OutputFormatConfig; v != nil {
		m["s3_output_format_config"] = []interface{}{flattenUpsolverS3OutputFormatConfig(v)}
	}

	return m
}

func flattenUpsolverS3OutputFormatConfig(upsolverS3OutputFormatConfig *appflow.UpsolverS3OutputFormatConfig) map[string]interface{} {
	if upsolverS3OutputFormatConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := upsolverS3OutputFormatConfig.AggregationConfig; v != nil {
		m["aggregation_config"] = []interface{}{flattenAggregationConfig(v)}
	}

	if v := upsolverS3OutputFormatConfig.FileType; v != nil {
		m["file_type"] = aws.StringValue(v)
	}

	if v := upsolverS3OutputFormatConfig.PrefixConfig; v != nil {
		m["prefix_config"] = []interface{}{flattenPrefixConfig(v)}
	}

	return m
}

func flattenZendeskDestinationProperties(zendeskDestinationProperties *appflow.ZendeskDestinationProperties) map[string]interface{} {
	if zendeskDestinationProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := zendeskDestinationProperties.ErrorHandlingConfig; v != nil {
		m["error_handling_config"] = []interface{}{flattenErrorHandlingConfig(v)}
	}

	if v := zendeskDestinationProperties.IdFieldNames; v != nil {
		m["id_field_names"] = aws.StringValueSlice(v)
	}

	if v := zendeskDestinationProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	if v := zendeskDestinationProperties.WriteOperationType; v != nil {
		m["write_operation_type"] = aws.StringValue(v)
	}

	return m
}

func flattenSourceFlowConfig(sourceFlowConfig *appflow.SourceFlowConfig) map[string]interface{} {
	if sourceFlowConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := sourceFlowConfig.ApiVersion; v != nil {
		m["api_version"] = aws.StringValue(v)
	}

	if v := sourceFlowConfig.ConnectorProfileName; v != nil {
		m["connector_profile_name"] = aws.StringValue(v)
	}

	if v := sourceFlowConfig.ConnectorType; v != nil {
		m["connector_type"] = aws.StringValue(v)
	}

	if v := sourceFlowConfig.IncrementalPullConfig; v != nil {
		m["incremental_pull_config"] = []interface{}{flattenIncrementalPullConfig(v)}
	}

	if v := sourceFlowConfig.SourceConnectorProperties; v != nil {
		m["source_connector_properties"] = []interface{}{flattenSourceConnectorProperties(v)}
	}

	return m
}

func flattenIncrementalPullConfig(incrementalPullConfig *appflow.IncrementalPullConfig) map[string]interface{} {
	if incrementalPullConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := incrementalPullConfig.DatetimeTypeFieldName; v != nil {
		m["datetime_type_field_name"] = aws.StringValue(v)
	}

	return m
}

func flattenSourceConnectorProperties(sourceConnectorProperties *appflow.SourceConnectorProperties) map[string]interface{} {
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

func flattenAmplitudeSourceProperties(amplitudeSourceProperties *appflow.AmplitudeSourceProperties) map[string]interface{} {
	if amplitudeSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := amplitudeSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenCustomConnectorSourceProperties(customConnectorSourceProperties *appflow.CustomConnectorSourceProperties) map[string]interface{} {
	if customConnectorSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := customConnectorSourceProperties.CustomProperties; v != nil {
		m["custom_properties"] = aws.StringValueMap(v)
	}

	if v := customConnectorSourceProperties.EntityName; v != nil {
		m["entity_name"] = aws.StringValue(v)
	}

	return m
}

func flattenDatadogSourceProperties(datadogSourceProperties *appflow.DatadogSourceProperties) map[string]interface{} {
	if datadogSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := datadogSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenDynatraceSourceProperties(dynatraceSourceProperties *appflow.DynatraceSourceProperties) map[string]interface{} {
	if dynatraceSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := dynatraceSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenGoogleAnalyticsSourceProperties(googleAnalyticsSourceProperties *appflow.GoogleAnalyticsSourceProperties) map[string]interface{} {
	if googleAnalyticsSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := googleAnalyticsSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenInforNexusSourceProperties(inforNexusSourceProperties *appflow.InforNexusSourceProperties) map[string]interface{} {
	if inforNexusSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := inforNexusSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenMarketoSourceProperties(marketoSourceProperties *appflow.MarketoSourceProperties) map[string]interface{} {
	if marketoSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := marketoSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenS3SourceProperties(s3SourceProperties *appflow.S3SourceProperties) map[string]interface{} {
	if s3SourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := s3SourceProperties.BucketName; v != nil {
		m["bucket_name"] = aws.StringValue(v)
	}

	if v := s3SourceProperties.BucketPrefix; v != nil {
		m["bucket_prefix"] = aws.StringValue(v)
	}

	if v := s3SourceProperties.S3InputFormatConfig; v != nil {
		m["s3_input_format_config"] = []interface{}{flattenS3InputFormatConfig(v)}
	}

	return m
}

func flattenS3InputFormatConfig(s3InputFormatConfig *appflow.S3InputFormatConfig) map[string]interface{} {
	if s3InputFormatConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := s3InputFormatConfig.S3InputFileType; v != nil {
		m["s3_input_file_type"] = aws.StringValue(v)
	}

	return m
}

func flattenSalesforceSourceProperties(salesforceSourceProperties *appflow.SalesforceSourceProperties) map[string]interface{} {
	if salesforceSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := salesforceSourceProperties.EnableDynamicFieldUpdate; v != nil {
		m["enable_dynamic_field_update"] = aws.BoolValue(v)
	}

	if v := salesforceSourceProperties.IncludeDeletedRecords; v != nil {
		m["include_deleted_records"] = aws.BoolValue(v)
	}

	if v := salesforceSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenSAPODataSourceProperties(sapoDataSourceProperties *appflow.SAPODataSourceProperties) map[string]interface{} {
	if sapoDataSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := sapoDataSourceProperties.ObjectPath; v != nil {
		m["object_path"] = aws.StringValue(v)
	}

	return m
}

func flattenServiceNowSourceProperties(serviceNowSourceProperties *appflow.ServiceNowSourceProperties) map[string]interface{} {
	if serviceNowSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := serviceNowSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenSingularSourceProperties(singularSourceProperties *appflow.SingularSourceProperties) map[string]interface{} {
	if singularSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := singularSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenSlackSourceProperties(slackSourceProperties *appflow.SlackSourceProperties) map[string]interface{} {
	if slackSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := slackSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenTrendmicroSourceProperties(trendmicroSourceProperties *appflow.TrendmicroSourceProperties) map[string]interface{} {
	if trendmicroSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := trendmicroSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenVeevaSourceProperties(veevaSourceProperties *appflow.VeevaSourceProperties) map[string]interface{} {
	if veevaSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := veevaSourceProperties.DocumentType; v != nil {
		m["document_type"] = aws.StringValue(v)
	}

	if v := veevaSourceProperties.IncludeAllVersions; v != nil {
		m["include_all_versions"] = aws.BoolValue(v)
	}

	if v := veevaSourceProperties.IncludeRenditions; v != nil {
		m["include_renditions"] = aws.BoolValue(v)
	}

	if v := veevaSourceProperties.IncludeSourceFiles; v != nil {
		m["include_source_files"] = aws.BoolValue(v)
	}

	if v := veevaSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenZendeskSourceProperties(zendeskSourceProperties *appflow.ZendeskSourceProperties) map[string]interface{} {
	if zendeskSourceProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := zendeskSourceProperties.Object; v != nil {
		m["object"] = aws.StringValue(v)
	}

	return m
}

func flattenTasks(tasks []*appflow.Task) []interface{} {
	if len(tasks) == 0 {
		return nil
	}

	var l []interface{}

	for _, task := range tasks {
		if task == nil {
			continue
		}

		l = append(l, flattenTask(task))
	}

	return l
}

func flattenTask(task *appflow.Task) map[string]interface{} {
	if task == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := task.ConnectorOperator; v != nil {
		m["connector_operator"] = []interface{}{flattenConnectorOperator(v)}
	}

	if v := task.DestinationField; v != nil {
		m["destination_field"] = aws.StringValue(v)
	}

	if v := task.SourceFields; v != nil {
		m["source_fields"] = aws.StringValueSlice(v)
	}

	if v := task.TaskProperties; v != nil {
		m["task_properties"] = aws.StringValueMap(v)
	}

	if v := task.TaskType; v != nil {
		m["task_type"] = aws.StringValue(v)
	}

	return m
}

func flattenConnectorOperator(connectorOperator *appflow.ConnectorOperator) map[string]interface{} {
	if connectorOperator == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := connectorOperator.Amplitude; v != nil {
		m["amplitude"] = aws.StringValue(v)
	}

	if v := connectorOperator.CustomConnector; v != nil {
		m["custom_connector"] = aws.StringValue(v)
	}

	if v := connectorOperator.Datadog; v != nil {
		m["datadog"] = aws.StringValue(v)
	}

	if v := connectorOperator.Dynatrace; v != nil {
		m["dynatrace"] = aws.StringValue(v)
	}

	if v := connectorOperator.GoogleAnalytics; v != nil {
		m["google_analytics"] = aws.StringValue(v)
	}

	if v := connectorOperator.InforNexus; v != nil {
		m["infor_nexus"] = aws.StringValue(v)
	}

	if v := connectorOperator.Marketo; v != nil {
		m["marketo"] = aws.StringValue(v)
	}

	if v := connectorOperator.S3; v != nil {
		m["s3"] = aws.StringValue(v)
	}

	if v := connectorOperator.Salesforce; v != nil {
		m["salesforce"] = aws.StringValue(v)
	}

	if v := connectorOperator.SAPOData; v != nil {
		m["sapo_data"] = aws.StringValue(v)
	}

	if v := connectorOperator.ServiceNow; v != nil {
		m["service_now"] = aws.StringValue(v)
	}

	if v := connectorOperator.Singular; v != nil {
		m["singular"] = aws.StringValue(v)
	}

	if v := connectorOperator.Slack; v != nil {
		m["slack"] = aws.StringValue(v)
	}

	if v := connectorOperator.Trendmicro; v != nil {
		m["trendmicro"] = aws.StringValue(v)
	}

	if v := connectorOperator.Veeva; v != nil {
		m["veeva"] = aws.StringValue(v)
	}

	if v := connectorOperator.Zendesk; v != nil {
		m["zendesk"] = aws.StringValue(v)
	}

	return m
}

func flattenTriggerConfig(triggerConfig *appflow.TriggerConfig) map[string]interface{} {
	if triggerConfig == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := triggerConfig.TriggerProperties; v != nil {
		m["trigger_properties"] = []interface{}{flattenTriggerProperties(v)}
	}

	if v := triggerConfig.TriggerType; v != nil {
		m["trigger_type"] = aws.StringValue(v)
	}

	return m
}

func flattenTriggerProperties(triggerProperties *appflow.TriggerProperties) map[string]interface{} {
	if triggerProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := triggerProperties.Scheduled; v != nil {
		m["trigger_properties"] = []interface{}{flattenScheduled(v)}
	}

	return m
}

func flattenScheduled(scheduledTriggerProperties *appflow.ScheduledTriggerProperties) map[string]interface{} {
	if scheduledTriggerProperties == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := scheduledTriggerProperties.DataPullMode; v != nil {
		m["data_pull_mode"] = aws.StringValue(v)
	}

	if v := scheduledTriggerProperties.FirstExecutionFrom; v != nil {
		m["first_execution_from"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := scheduledTriggerProperties.ScheduleEndTime; v != nil {
		m["schedule_end_time"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := scheduledTriggerProperties.ScheduleExpression; v != nil {
		m["schedule_expression"] = aws.StringValue(v)
	}

	if v := scheduledTriggerProperties.ScheduleOffset; v != nil {
		m["schedule_offset"] = aws.Int64Value(v)
	}

	if v := scheduledTriggerProperties.ScheduleStartTime; v != nil {
		m["schedule_start_time"] = aws.TimeValue(v).Format(time.RFC3339)
	}

	if v := scheduledTriggerProperties.Timezone; v != nil {
		m["timezone"] = aws.StringValue(v)
	}

	return m
}
