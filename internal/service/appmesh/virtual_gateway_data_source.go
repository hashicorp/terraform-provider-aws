package appmesh

// Remember to register this new data source in the provider
// (internal/provider/provider.go) once you finish. Otherwise, Terraform won't
// know about it.

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func DataSourceVirtualGateway() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVirtualGatewayRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"mesh_owner": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
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
																Type:     schema.TypeSet,
																Computed: true,
																Elem: &schema.Schema{
																	Type: schema.TypeInt,
																},
																Set: schema.HashInt,
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
																									Type:     schema.TypeSet,
																									Computed: true,
																									Elem:     &schema.Schema{Type: schema.TypeString},
																									Set:      schema.HashString,
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
																									Type:     schema.TypeSet,
																									Computed: true,
																									Elem: &schema.Schema{
																										Type: schema.TypeString,
																									},
																									Set: schema.HashString,
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

						"listener": {
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

												"mode": {
													Type:     schema.TypeString,
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
																						Type:     schema.TypeSet,
																						Computed: true,
																						Elem:     &schema.Schema{Type: schema.TypeString},
																						Set:      schema.HashString,
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

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"last_updated_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchema(),
		},
	}

}

func dataSourceVirtualGatewayRead(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*conns.AWSClient).AppMeshConn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeVirtualGatewayInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualGatewayName: aws.String(d.Get("name").(string)),
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
