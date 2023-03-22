package appmesh

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

// @SDKDataSource("aws_appmesh_route")
func DataSourceRoute() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceRouteRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"mesh_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"mesh_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"virtual_router_name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"spec": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grpc_route": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"weighted_target": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"virtual_node": {
																Type:     schema.TypeString,
																Computed: true,
															},

															"weight": {
																Type:     schema.TypeInt,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},

									"match": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"metadata": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"invert": {
																Type:     schema.TypeBool,
																Computed: true,
															},

															"match": {
																Type:     schema.TypeList,
																Computed: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"prefix": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"range": {
																			Type:     schema.TypeList,
																			Computed: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"end": {
																						Type:     schema.TypeInt,
																						Computed: true,
																					},

																					"start": {
																						Type:     schema.TypeInt,
																						Computed: true,
																					},
																				},
																			},
																		},

																		"regex": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},

																		"suffix": {
																			Type:     schema.TypeString,
																			Computed: true,
																		},
																	},
																},
															},

															"name": {
																Type:     schema.TypeString,
																Computed: true,
															},
														},
													},
												},

												"method_name": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"prefix": {
													Type:     schema.TypeString,
													Computed: true,
												},

												"service_name": {
													Type:     schema.TypeString,
													Computed: true,
												},
											},
										},
									},

									"retry_policy": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"grpc_retry_events": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},

												"http_retry_events": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},

												"max_retries": {
													Type:     schema.TypeInt,
													Computed: true,
												},

												"per_retry_timeout": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:     schema.TypeString,
																Computed: true,
															},

															"value": {
																Type:     schema.TypeInt,
																Computed: true,
															},
														},
													},
												},

												"tcp_retry_events": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},
											},
										},
									},

									"timeout": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:     schema.TypeString,
																Computed: true,
															},

															"value": {
																Type:     schema.TypeInt,
																Computed: true,
															},
														},
													},
												},

												"per_request": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:     schema.TypeString,
																Computed: true,
															},

															"value": {
																Type:     schema.TypeInt,
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

						"http2_route": func() *schema.Schema {
							schema := DataSourceRouteHTTPRouteSchema()
							return schema
						}(),

						"http_route": func() *schema.Schema {
							schema := DataSourceRouteHTTPRouteSchema()
							return schema
						}(),

						"priority": {
							Type:     schema.TypeInt,
							Computed: true,
						},

						"tcp_route": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"weighted_target": {
													Type:     schema.TypeSet,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"virtual_node": {
																Type:     schema.TypeString,
																Computed: true,
															},

															"weight": {
																Type:     schema.TypeInt,
																Computed: true,
															},
														},
													},
												},
											},
										},
									},

									"timeout": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle": {
													Type:     schema.TypeList,
													Computed: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:     schema.TypeString,
																Computed: true,
															},

															"value": {
																Type:     schema.TypeInt,
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

			"tags": tftags.TagsSchemaComputed(),
		},
	}
}

func DataSourceRouteHTTPRouteSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Computed: true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"action": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"weighted_target": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"virtual_node": {
											Type:     schema.TypeString,
											Computed: true,
										},

										"weight": {
											Type:     schema.TypeInt,
											Computed: true,
										},
									},
								},
							},
						},
					},
				},

				"match": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"header": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"invert": {
											Type:     schema.TypeBool,
											Computed: true,
										},

										"match": {
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"exact": {
														Type:     schema.TypeString,
														Computed: true,
													},

													"prefix": {
														Type:     schema.TypeString,
														Computed: true,
													},

													"range": {
														Type:     schema.TypeList,
														Computed: true,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"end": {
																	Type:     schema.TypeInt,
																	Computed: true,
																},

																"start": {
																	Type:     schema.TypeInt,
																	Computed: true,
																},
															},
														},
													},

													"regex": {
														Type:     schema.TypeString,
														Computed: true,
													},

													"suffix": {
														Type:     schema.TypeString,
														Computed: true,
													},
												},
											},
										},

										"name": {
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},

							"method": {
								Type:     schema.TypeString,
								Computed: true,
							},

							"prefix": {
								Type:     schema.TypeString,
								Computed: true,
							},

							"scheme": {
								Type:     schema.TypeString,
								Computed: true,
							},
						},
					},
				},

				"retry_policy": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"http_retry_events": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},

							"max_retries": {
								Type:     schema.TypeInt,
								Computed: true,
							},

							"per_retry_timeout": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"unit": {
											Type:     schema.TypeString,
											Computed: true,
										},

										"value": {
											Type:     schema.TypeInt,
											Computed: true,
										},
									},
								},
							},

							"tcp_retry_events": {
								Type:     schema.TypeSet,
								Computed: true,
								Elem:     &schema.Schema{Type: schema.TypeString},
							},
						},
					},
				},

				"timeout": {
					Type:     schema.TypeList,
					Computed: true,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"idle": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"unit": {
											Type:     schema.TypeString,
											Computed: true,
										},

										"value": {
											Type:     schema.TypeInt,
											Computed: true,
										},
									},
								},
							},

							"per_request": {
								Type:     schema.TypeList,
								Computed: true,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"unit": {
											Type:     schema.TypeString,
											Computed: true,
										},

										"value": {
											Type:     schema.TypeInt,
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
	}
}

func dataSourceRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	resp, err := conn.DescribeRoute(req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Route: %s", err)
	}

	arn := aws.StringValue(resp.Route.Metadata.Arn)

	d.SetId(aws.StringValue(resp.Route.RouteName))

	d.Set("name", resp.Route.RouteName)
	d.Set("mesh_name", resp.Route.MeshName)
	d.Set("mesh_owner", resp.Route.Metadata.MeshOwner)
	d.Set("virtual_router_name", resp.Route.VirtualRouterName)
	d.Set("arn", arn)
	d.Set("created_date", resp.Route.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.Route.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("resource_owner", resp.Route.Metadata.ResourceOwner)

	err = d.Set("spec", flattenRouteSpec(resp.Route.Spec))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh Route (%s): %s", arn, err)
	}

	if err := d.Set("tags", tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	return diags
}
