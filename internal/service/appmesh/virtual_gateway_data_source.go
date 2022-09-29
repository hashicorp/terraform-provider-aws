package appmesh

// Remember to register this new data source in the provider
// (internal/provider/provider.go) once you finish. Otherwise, Terraform won't
// know about it.

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"time"
)

// DataSourceVirtualGateway 1. Package declaration
// 2. Imports
// 3. Main data source function with schema
// 4. Create, read, update, delete functions (in that order)
// 5. Other functions (flatteners, expanders, waiters, finders, etc.)
func DataSourceVirtualGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVirtualGatewayRead,

		// TIP: ==== SCHEMA ====
		// In the schema, add each of the arguments and attributes in snake
		// case (e.g., delete_automated_backups).
		// * Alphabetize arguments to make them easier to find.
		// * Do not add a blank line between arguments/attributes.
		//
		// Users can configure argument values while attribute values cannot be
		// configured and are used as output. Arguments have either:
		// Required: true,
		// Optional: true,
		//
		// All attributes will be computed and some arguments. If users w
		// ant to read updated information or detect drift for an argument,
		// it should be computed:
		// Computed: true,
		//
		// You will typically find arguments in the input struct
		// (e.g., CreateDBInstanceInput) for the create operation. Sometimes
		// they are only in the input struct (e.g., ModifyDBInstanceInput) for
		// the modify operation.
		//
		// For more about schema options, visit
		// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema#Schema
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"mesh_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"mesh_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"spec": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backend_defaults": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_policy": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"tls": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"certificate": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"file": {
																			Type:     schema.TypeList,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"certificate_chain": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"private_key": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},
																				},
																			},
																		},

																		"sds": {
																			Type:     schema.TypeList,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"secret_name": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},
																				},
																			},
																		},
																	},
																},
															},

															"enforce": {
																Type:     schema.TypeBool,
																Computed: true,
															},

															"ports": {
																Type:     schema.TypeList,
																Computed: true,
															},

															"validation": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"subject_alternative_names": {
																			Type:     schema.TypeList,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"match": {
																						Type:     schema.TypeList,
																						Computed: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"exact": {
																									Type:     schema.TypeList,
																									Computed: true,
																								},
																							},
																						},
																					},
																				},
																			},
																		},

																		"trust": {
																			Type:     schema.TypeList,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"acm": {
																						Type:     schema.TypeList,
																						Computed: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"certificate_authority_arns": {
																									Type:     schema.TypeList,
																									Computed: true,
																								},
																							},
																						},
																					},

																					"file": {
																						Type:     schema.TypeList,
																						Computed: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"certificate_chain": {
																									Type:     schema.TypeString,
																									Computed: true,
																								},
																							},
																						},
																					},

																					"sds": {
																						Type:     schema.TypeList,
																						Computed: true,
																						Elem: &schema.Resource{
																							Schema: map[string]*schema.Schema{
																								"secret_name": {
																									Type:     schema.TypeString,
																									Computed: true,
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
										},
									},
								},
							},
						},

						"listeners": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection_pool": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"grpc": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_requests": {
																Type:     schema.TypeInt,
																Computed: true,
															},
														},
													},
												},

												"http": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_connections": {
																Type:     schema.TypeInt,
																Computed: true,
															},

															"max_pending_requests": {
																Type:     schema.TypeInt,
																Computed: true,
															},
														},
													},
												},

												"http2": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_requests": {
																Type:     schema.TypeInt,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},

									"health_check": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"healthy_threshold": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"interval_millis": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"path": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"port": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"protocol": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"timeout_millis": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"unhealthy_threshold": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},

									"port_mapping": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"protocol": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},

									"tls": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"certificate": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"acm": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"certificate_arn": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
															},

															"file": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"certificate_chain": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"private_key": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
															},

															"sds": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"secret_name": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
															},
														},
													},
												},

												"validation": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"subject_alternative_names": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"match": {
																			Type:     schema.TypeList,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"exact": {
																						Type:     schema.TypeList,
																						Computed: true,
																					},
																				},
																			},
																		},
																	},
																},
															},

															"trust": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"file": {
																			Type:     schema.TypeList,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"certificate_chain": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},
																				},
																			},
																		},

																		"sds": {
																			Type:     schema.TypeList,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"secret_name": {
																						Type:     schema.TypeString,
																						Computed: true,
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

												"mode": {
													Type:         schema.TypeString,
													Computed:     true,
													ValidateFunc: validation.StringInSlice([]string{"STRICT", "PERMISSIVE", "DISABLED"}, false),
												},
											},
										},
									},
								},
							},
						},

						"logging": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_log": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"file": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"format": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"json": {
																			Type:     schema.TypeList,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"key": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},

																					"value": {
																						Type:     schema.TypeString,
																						Computed: true,
																					},
																				},
																			},
																		},

																		"text": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
															},

															"path": {
																Type:     schema.TypeString,
																Computed: true,
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
			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

const (
	DSNameVirtualGateway = "Virtual Gateway Data Source"
)

func dataSourceVirtualGatewayRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*conns.AWSClient).AppMeshConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeVirtualGatewayInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualGatewayName: aws.String(d.Get("virtual_gateway_name").(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	resp, err := conn.DescribeVirtualGateway(req)
	if err != nil {
		return fmt.Errorf("error reading App Mesh Virtual Gateway: %s", err)
	}

	arn := aws.StringValue(resp.VirtualGateway.Metadata.Arn)

	d.SetId(aws.StringValue(resp.VirtualGateway.VirtualGatewayName))

	d.Set("name", resp.VirtualGateway.VirtualGatewayName)
	d.Set("mesh_name", resp.VirtualGateway.MeshName)
	d.Set("mesh_owner", resp.VirtualGateway.Metadata.MeshOwner)
	d.Set("resource_owner", resp.VirtualGateway.Metadata.ResourceOwner)
	d.Set("arn", arn)
	d.Set("created_date", resp.VirtualGateway.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualGateway.Metadata.LastUpdatedAt.Format(time.RFC3339))

	err = d.Set("spec", flattenVirtualGatewaySpec(resp.VirtualGateway.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %s", err)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for App Mesh Virtual Service (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}
