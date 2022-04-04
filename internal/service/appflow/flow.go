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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
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
															Type:             schema.TypeMap,
															Optional:         true,
															MaxItems:         50,
															ValidateDiagFunc: validation.All(validation.MapKeyLenBetween(1, 128), validation.MapKeyMatch(regexp.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"), validation.MapValueLenBetween(0, 2048), validation.MapValueMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters")),
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
																		Type:     schema.TypeBoolean,
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
																		Type:     schema.TypeBoolean,
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
																		Type:     schema.TypeBoolean,
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
												Elem:     &schema.Resource,
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
																		Type:     schema.TypeBoolean,
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
																		Type:     schema.TypeBoolean,
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
																			"aggregation_type": {
																				Type:         schema.TypeString,
																				Optional:     true,
																				ValidateFunc: validation.StringInSlice(appflow.AggregationType_Values(), false),
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
																		Type:     schema.TypeBoolean,
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
																		Type:     schema.TypeBoolean,
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
																		Type:     schema.TypeBoolean,
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
																			"aggregation_type": {
																				Type:         schema.TypeString,
																				Optional:     true,
																				ValidateFunc: validation.StringInSlice(appflow.AggregationType_Values(), false),
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
																		Type:     schema.TypeBoolean,
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
				},
			},
			"kms_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`arn:aws:kms:.*:[0-9]+:.*`), "must be a valid ARN of a Key Management Services (KMS) key"),
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
											Type:             schema.TypeMap,
											Optional:         true,
											MaxItems:         50,
											ValidateDiagFunc: validation.All(validation.MapKeyLenBetween(1, 128), validation.MapKeyMatch(regexp.MustCompile(`[\w]+`), "must contain only alphanumeric and underscore (_) characters"), validation.MapValueLenBetween(0, 2048), validation.MapValueMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters")),
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
										ValidateFunc: validation.StringInSlice(appflow.SAPODataOperator_Values(), false),
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
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(0, 2048),
						},
						"task_properties": {
							Type:             schema.TypeMap,
							Optional:         true,
							ValidateDiagFunc: validation.All(validation.MapValueLenBetween(0, 2048), validation.MapValueMatch(regexp.MustCompile(`\S+`), "must not contain any whitespace characters")),
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
													Type:         schema.TypeInteger,
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
					},
					"trigger_type": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice(appflow.TriggerType_Values()),
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
	// TIP: Generally, the Create function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 2. Populate a create input structure
	// 3. Call the AWS create/put function
	// 4. Using the output from the create function, set the minimum arguments
	//    and attributes for the Read function to work. At a minimum, set the
	//    resource ID. E.g., d.SetId(<Identifier, such as AWS ID or ARN>)
	// 5. Use a waiter to wait for create to complete
	// 6. Call the Read function in the Create return

	conn := meta.(*conns.AWSClient).AppFlowConn

	// TIP: 2. Populate a create input structure
	in := &appflow.CreateFlowInput{
		// TIP: Mandatory or fields that will always be present can be set when
		// you create the Input structure. (Replace these with real fields.)
		FlowName: aws.String(d.Get("name").(string)),
		FlowType: aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("max_size"); ok {
		// TIP: Optional fields should be set based on whether or not they are
		// used.
		in.MaxSize = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("complex_argument"); ok && len(v.([]interface{})) > 0 {
		// TIP: Use an expander to assign a complex argument.
		in.ComplexArguments = expandComplexArguments(v.([]interface{}))
	}

	// TIP: Not all resources support tags and tags don't always make sense. If
	// your resource doesn't need tags, you can remove the tags lines here and
	// below. Many resources do include tags so this a reminder to include them
	// where possible.
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	if len(tags) > 0 {
		in.Tags = Tags(tags.IgnoreAWS())
	}

	// TIP: 3. Call the AWS create function
	out, err := conn.CreateFlowWithContext(ctx, in)

	if err != nil {
		return diag.Errorf("creating Appflow Flow (%s): %s", d.Get("name").(string), err)
	}

	if out == nil || out.Flow == nil {
		return diag.Errorf("creating Appflow Flow (%s): empty output", d.Get("name").(string))
	}

	// TIP: 4. Set the minimum arguments and/or attributes for the Read function to
	// work.
	d.SetId(aws.ToString(out.Flow.FlowID))

	// TIP: 5. Use a waiter to wait for create to complete
	if _, err := waitFlowCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.Errorf("waiting for Appflow Flow (%s) create: %s", d.Id(), err)
	}

	// TIP: 6. Call the Read function in the Create return
	return resourceFlowRead(ctx, d, meta)
}

func resourceFlowRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TIP: Generally, the Read function should do the following things. Make
	// sure there is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Get the resource from AWS
	// 3. Set ID to empty where resource is not new and not found
	// 4. Set the arguments and attributes
	// 5. Set the tags
	// 6. Return nil

	// TIP: 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).AppFlowConn

	// TIP: 2. Get the resource from AWS using an API Get, List, or Describe-
	// type function, or, better yet, using a finder.
	out, err := findFlowByID(ctx, conn, d.Id())

	// TIP: 3. Set ID to empty where resource is not new and not found
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppFlow Flow (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading AppFlow Flow (%s): %s", d.Id(), err)
	}

	// TIP: 4. Set the arguments and attributes
	//
	// For simple data types (i.e., schema.TypeString, schema.TypeBool,
	// schema.TypeInt, and schema.TypeFloat), a simple Set call (e.g.,
	// d.Set("arn", out.Arn) is sufficient. No error or nil checking is
	// necessary.
	//
	// However, there are some situations where more handling is needed.
	// a. Complex data types (e.g., schema.TypeList, schema.TypeSet)
	// b. Where errorneous diffs occur. For example, a schema.TypeString may be
	//    a JSON. AWS may return the JSON in a slightly different order but it
	//    is equivalent to what is already set. In that case, you may check if
	//    it is equivalent before setting the different JSON.
	d.Set("arn", out.Arn)
	d.Set("name", out.Name)

	// TIP: Setting a complex type.
	// For more information, see:
	// https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/data-handling-and-conversion.md
	// https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/data-handling-and-conversion.md#flatten-functions-for-blocks
	// https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/data-handling-and-conversion.md#root-typeset-of-resource-and-aws-list-of-structure
	if err := d.Set("complex_argument", flattenComplexArguments(out.ComplexArguments)); err != nil {
		return diag.Errorf("setting complex argument: %s", err)
	}

	// TIP: Setting a JSON string to avoid errorneous diffs.
	p, err := verify.SecondJSONUnlessEquivalent(d.Get("policy").(string), aws.ToString(out.Policy))

	if err != nil {
		return diag.Errorf("while setting policy (%s), encountered: %s", p, err)
	}

	p, err = structure.NormalizeJsonString(p)

	if err != nil {
		return diag.Errorf("policy (%s) is invalid JSON: %s", p, err)
	}

	d.Set("policy", p)

	// TIP: 5. Set the tags
	//
	// TIP: Not all resources support tags and tags don't always make sense. If
	// your resource doesn't need tags, you can remove the tags lines here and
	// below. Many resources do include tags so this a reminder to include them
	// where possible.
	tags, err := ListTags(ctx, conn, d.Id())

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

	// TIP: 6. Return nil
	return nil
}

func resourceFlowUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TIP: Not all resources have Update functions. There are a few reasons:
	// a. The AWS API does not support changing a resource
	// b. All arguments have ForceNew: true, set
	// c. The AWS API uses a create call to modify an existing resource
	//
	// In the cases of a. and b., the main resource function will not have a
	// UpdateWithoutTimeout defined. In the case of c., Update and Create are
	// the same.
	//
	// The rest of the time, there should be an Update function and it should
	// do the following things. Make sure there is a good reason if you don't
	// do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a modify input structure and check for changes
	// 3. Call the AWS modify/update function
	// 4. Use a waiter to wait for update to complete
	// 5. Call the Read function in the Update return

	// TIP: 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).AppFlowConn

	// TIP: 2. Populate a modify input structure and check for changes
	//
	// When creating the input structure, only include mandatory fields. Other
	// fields are set as needed. You can use a flag, such as update below, to
	// determine if a certain portion of arguments have been changed and
	// whether to call the AWS update function.
	update := false

	in := &appflow.UpdateFlowInput{
		Id: aws.String(d.Id()),
	}

	if d.HasChanges("an_argument") {
		in.AnArgument = d.Get("an_argument").(string)
		update = true
	}

	if !update {
		// If update doesn't do anything at all, which is rare, you can return
		// nil. Otherwise, return a read call, as below.
		return nil
	}

	log.Printf("[DEBUG] Updating AppFlow Flow (%s): %#v", d.Id(), in)
	out, err := conn.UpdateFlow(ctx, in)

	if err != nil {
		return diag.Errorf("updating AppFlow Flow (%s): %s", d.Id(), err)
	}

	if _, err := waitFlowUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.Errorf("waiting for AppFlow Flow (%s) update: %s", d.Id(), err)
	}

	return resourceFlowRead(ctx, d, meta)
}

func resourceFlowDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// TIP: Most resources have Delete functions. There are rare situations
	// where you might not need a delete:
	// a. The AWS API does not provide a way to delete the resource
	// b. The point of your resource is to perform an action (e.g., reboot a
	//    server) and deleting serves no purpose.
	//
	// The Delete function and should do the following things. Make sure there
	// is a good reason if you don't do one of these.
	//
	// 1. Get a client connection to the relevant service
	// 2. Populate a delete input structure
	// 3. Call the AWS delete function
	// 4. Use a waiter to wait for delete to complete
	// 5. Return nil

	// TIP: 1. Get a client connection to the relevant service
	conn := meta.(*conns.AWSClient).AppFlowConn

	// TIP: 2. Populate a delete input structure
	log.Printf("[INFO] Deleting AppFlow Flow %s", d.Id())

	// TIP: 3. Call the AWS delete function
	_, err := conn.DeleteFlowWithContext(ctx, &appflow.DeleteFlowInput{
		Id: aws.String(d.Id()),
	})

	// On rare occassions, the API returns a not found error after deleting a
	// resource. If that happens, we don't want it to show up as an error.
	if tfawserr.ErrCodeEquals(err, appflow.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting AppFlow Flow (%s): %s", d.Id(), err)
	}

	// TIP: 4. Use a waiter to wait for delete to complete
	if _, err := waitFlowDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.Errorf("waiting for AppFlow Flow (%s) to be deleted: %s", d.Id(), err)
	}

	// TIP: 5. Return nil
	return nil
}

// TIP: Create constants for states and statuses if the service does not
// already have suitable constants. We prefer that you use the constants
// provided in the service if available (e.g., amp.WorkspaceStatusCodeActive).
const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

// TIP: Some resources of some services have waiters provided by the AWS API.
// Unless they do not work properly, use them rather than defining new ones
// here.
//
// Sometimes we define the wait, status, and find functions in separate
// files, wait.go, status.go, and find.go. Follow the pattern set out in the
// service and define these where it makes the most sense.
//
// If these functions are used in the _test.go file, they will need to be
// exported (i.e., capitalized).
//
// You will need to adjust the parameters and names to fit the service.
func waitFlowCreated(ctx context.Context, conn *appflow.AppFlow, id string, timeout time.Duration) (*appflow.Flow, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusFlow(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForState()

	if out, ok := outputRaw.(*appflow.Flow); ok {
		return out, err
	}

	return nil, err
}

// TIP: It is easier to determine whether a resource is updated for some
// resources than others. The best case is a status flag that tells you when
// the update has been fully realized. Other times, you can check to see if a
// key resource argument is updated to a new value or not.
func waitFlowUpdated(ctx context.Context, conn *appflow.AppFlow, id string, timeout time.Duration) (*appflow.Flow, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusFlow(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForState()

	if out, ok := outputRaw.(*appflow.Flow); ok {
		return out, err
	}

	return nil, err
}

// TIP: A deleted waiter is almost like a backwards created waiter. There may
// be additional pending states, however.
func waitFlowDeleted(ctx context.Context, conn *appflow.AppFlow, id string, timeout time.Duration) (*appflow.Flow, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusFlow(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if out, ok := outputRaw.(*appflow.Flow); ok {
		return out, err
	}

	return nil, err
}

// TIP: The status function can return an actual status when that field is
// available from the API (e.g., out.Status). Otherwise, you can use custom
// statuses to communicate the states of the resource.
//
// Design status so that it can be reused by a create, update, and delete
// waiter, if possible.
func statusFlow(ctx context.Context, conn *appflow.AppFlow, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findFlowByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, out.Status, nil
	}
}

// TIP: The find function is not strictly necessary. You could do the API
// request from the status function. However, we have found that find often
// comes in handy in other places besides the status function. As a result, it
// is good practice to define it separately.
func findFlowByID(ctx context.Context, conn *appflow.AppFlow, id string) (*appflow.Flow, error) {
	in := &appflow.GetFlowInput{
		Id: aws.String(id),
	}

	out, err := conn.GetFlowWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, appflow.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.Flow == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.Flow, nil
}

// TIP: Flatteners and expanders ("flex" functions) help handle complex data
// types. Flatteners take an API data type and return something you can use in
// a d.Set() call. In other words, flatteners translate from AWS -> Terraform.
//
// On the other hand, expanders take a Terraform data structure and return
// something that you can send to the AWS API. In other words, expanders
// translate from Terraform -> AWS.
//
// See more:
// https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/data-handling-and-conversion.md
func flattenComplexArgument(apiObject *appflow.ComplexArgument) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	m := map[string]interface{}{}

	if v := apiObject.SubFieldOne; v != nil {
		m["sub_field_one"] = aws.ToString(v)
	}

	if v := apiObject.SubFieldTwo; v != nil {
		m["sub_field_two"] = aws.ToString(v)
	}

	return m
}

// TIP: Often the AWS API will return a slice of structures in response to a
// request for information. Sometimes you will have set criteria (e.g., the ID)
// that means you'll get back a one-length slice. This plural function works
// brilliantly for that situation too.
func flattenComplexArguments(apiObjects []*appflow.ComplexArgument) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []interface{}

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		l = append(l, flattenComplexArgument(apiObject))
	}

	return l
}

// TIP: Remember, as mentioned above, expanders take a Terraform data structure
// and return something that you can send to the AWS API. In other words,
// expanders translate from Terraform -> AWS.
//
// See more:
// https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/contributing/data-handling-and-conversion.md
func expandComplexArgument(tfMap map[string]interface{}) *appflow.ComplexArgument {
	if tfMap == nil {
		return nil
	}

	a := &appflow.ComplexArgument{}

	if v, ok := tfMap["sub_field_one"].(string); ok && v != "" {
		a.SubFieldOne = aws.String(v)
	}

	if v, ok := tfMap["sub_field_two"].(string); ok && v != "" {
		a.SubFieldTwo = aws.String(v)
	}

	return a
}

// TIP: Even when you have a list with max length of 1, this plural function
// works brilliantly. However, if the AWS API takes a structure rather than a
// slice of structures, you will not need it.
func expandComplexArguments(tfList []interface{}) []*appflow.ComplexArgument {
	// TIP: The AWS API can be picky about whether you send a nil or zero-
	// length for an argument that should be cleared. For example, in some
	// cases, if you send a nil value, the AWS API interprets that as "make no
	// changes" when what you want to say is "remove everything." Sometimes
	// using a zero-length list will cause an error.
	//
	// As a result, here are two options. Usually, option 1, nil, will work as
	// expected, clearing the field. But, test going from something to nothing
	// to make sure it works. If not, try the second option.
	// Option 1: Returning nil for zero-length list
	if len(tfList) == 0 {
		return nil
	}

	var s []*appflow.ComplexArgument

	// Option 2: Return zero-length list for zero-length list. If option 1 does
	// not work, after testing going from something to nothing (if that is
	// possible), uncomment out the next line and remove option 1.
	// s := make([]*appflow.ComplexArgument, 0)

	for _, r := range tfList {
		m, ok := r.(map[string]interface{})

		if !ok {
			continue
		}

		a := expandComplexArgument(m)

		if a == nil {
			continue
		}

		s = append(s, a)
	}

	return s
}
