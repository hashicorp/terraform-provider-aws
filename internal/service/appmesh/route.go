package appmesh

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceRoute() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteCreate,
		ReadWithoutTimeout:   resourceRouteRead,
		UpdateWithoutTimeout: resourceRouteUpdate,
		DeleteWithoutTimeout: resourceRouteDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRouteImport,
		},

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

			"virtual_router_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"grpc_route": {
							Type:          schema.TypeList,
							Optional:      true,
							MinItems:      0,
							MaxItems:      1,
							ConflictsWith: []string{"spec.0.http2_route", "spec.0.http_route", "spec.0.tcp_route"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"weighted_target": {
													Type:     schema.TypeSet,
													Required: true,
													MinItems: 1,
													MaxItems: 10,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"virtual_node": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 255),
															},

															"weight": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(0, 100),
															},

															"port": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IsPortNumber,
															},
														},
													},
												},
											},
										},
									},

									"match": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"metadata": {
													Type:     schema.TypeSet,
													Optional: true,
													MinItems: 0,
													MaxItems: 10,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"invert": {
																Type:     schema.TypeBool,
																Optional: true,
																Default:  false,
															},

															"match": {
																Type:     schema.TypeList,
																Optional: true,
																MinItems: 0,
																MaxItems: 1,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"exact": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},

																		"prefix": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},

																		"range": {
																			Type:     schema.TypeList,
																			Optional: true,
																			MinItems: 0,
																			MaxItems: 1,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"end": {
																						Type:     schema.TypeInt,
																						Required: true,
																					},

																					"start": {
																						Type:     schema.TypeInt,
																						Required: true,
																					},
																				},
																			},
																		},

																		"regex": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},

																		"suffix": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringLenBetween(1, 255),
																		},
																	},
																},
															},

															"name": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 50),
															},
														},
													},
												},

												"method_name": {
													Type:         schema.TypeString,
													Optional:     true,
													RequiredWith: []string{"spec.0.grpc_route.0.match.0.service_name"},
												},

												"prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 50),
												},

												"service_name": {
													Type:     schema.TypeString,
													Optional: true,
												},

												"port": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IsPortNumber,
												},
											},
										},
									},

									"retry_policy": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"grpc_retry_events": {
													Type:     schema.TypeSet,
													Optional: true,
													MinItems: 0,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
												},

												"http_retry_events": {
													Type:     schema.TypeSet,
													Optional: true,
													MinItems: 0,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
												},

												"max_retries": {
													Type:     schema.TypeInt,
													Required: true,
												},

												"per_retry_timeout": {
													Type:     schema.TypeList,
													Required: true,
													MinItems: 1,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
															},

															"value": {
																Type:     schema.TypeInt,
																Required: true,
															},
														},
													},
												},

												"tcp_retry_events": {
													Type:     schema.TypeSet,
													Optional: true,
													MinItems: 0,
													Elem:     &schema.Schema{Type: schema.TypeString},
													Set:      schema.HashString,
												},
											},
										},
									},

									"timeout": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
															},

															"value": {
																Type:     schema.TypeInt,
																Required: true,
															},
														},
													},
												},

												"per_request": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
															},

															"value": {
																Type:     schema.TypeInt,
																Required: true,
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
							schema := RouteHTTPRouteSchema()
							schema.ConflictsWith = []string{"spec.0.grpc_route", "spec.0.http_route", "spec.0.tcp_route"}
							return schema
						}(),

						"http_route": func() *schema.Schema {
							schema := RouteHTTPRouteSchema()
							schema.ConflictsWith = []string{"spec.0.grpc_route", "spec.0.http2_route", "spec.0.tcp_route"}
							return schema
						}(),

						"priority": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(0, 1000),
						},

						"tcp_route": {
							Type:          schema.TypeList,
							Optional:      true,
							MinItems:      0,
							MaxItems:      1,
							ConflictsWith: []string{"spec.0.grpc_route", "spec.0.http2_route", "spec.0.http_route"},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"action": {
										Type:     schema.TypeList,
										Required: true,
										MinItems: 1,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"weighted_target": {
													Type:     schema.TypeSet,
													Required: true,
													MinItems: 1,
													MaxItems: 10,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"virtual_node": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringLenBetween(1, 255),
															},

															"weight": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntBetween(0, 100),
															},

															"port": {
																Type:         schema.TypeInt,
																Optional:     true,
																ValidateFunc: validation.IsPortNumber,
															},
														},
													},
												},
											},
										},
									},

									"match": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:         schema.TypeInt,
													Optional:     true,
													ValidateFunc: validation.IsPortNumber,
												},
											},
										},
									},

									"timeout": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"idle": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
															},

															"value": {
																Type:     schema.TypeInt,
																Required: true,
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

			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

// RouteHTTPRouteSchema returns the schema for `http2_route` and `http_route` attributes.
func RouteHTTPRouteSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MinItems: 0,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"action": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"weighted_target": {
								Type:     schema.TypeSet,
								Required: true,
								MinItems: 1,
								MaxItems: 10,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"virtual_node": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 255),
										},

										"weight": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntBetween(0, 100),
										},

										"port": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IsPortNumber,
										},
									},
								},
							},
						},
					},
				},

				"match": {
					Type:     schema.TypeList,
					Required: true,
					MinItems: 1,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"header": {
								Type:     schema.TypeSet,
								Optional: true,
								MinItems: 0,
								MaxItems: 10,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"invert": {
											Type:     schema.TypeBool,
											Optional: true,
											Default:  false,
										},

										"match": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"exact": {
														Type:         schema.TypeString,
														Optional:     true,
														ValidateFunc: validation.StringLenBetween(1, 255),
													},

													"prefix": {
														Type:         schema.TypeString,
														Optional:     true,
														ValidateFunc: validation.StringLenBetween(1, 255),
													},

													"range": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"end": {
																	Type:     schema.TypeInt,
																	Required: true,
																},

																"start": {
																	Type:     schema.TypeInt,
																	Required: true,
																},
															},
														},
													},

													"regex": {
														Type:         schema.TypeString,
														Optional:     true,
														ValidateFunc: validation.StringLenBetween(1, 255),
													},

													"suffix": {
														Type:         schema.TypeString,
														Optional:     true,
														ValidateFunc: validation.StringLenBetween(1, 255),
													},
												},
											},
										},

										"name": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 50),
										},
									},
								},
							},

							"method": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringInSlice(appmesh.HttpMethod_Values(), false),
							},

							"prefix": {
								Type:         schema.TypeString,
								Required:     true,
								ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
							},

							"scheme": {
								Type:         schema.TypeString,
								Optional:     true,
								ValidateFunc: validation.StringInSlice(appmesh.HttpScheme_Values(), false),
							},

							"port": {
								Type:         schema.TypeInt,
								Optional:     true,
								ValidateFunc: validation.IsPortNumber,
							},
						},
					},
				},

				"retry_policy": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"http_retry_events": {
								Type:     schema.TypeSet,
								Optional: true,
								MinItems: 0,
								Elem:     &schema.Schema{Type: schema.TypeString},
								Set:      schema.HashString,
							},

							"max_retries": {
								Type:     schema.TypeInt,
								Required: true,
							},

							"per_retry_timeout": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"unit": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
										},

										"value": {
											Type:     schema.TypeInt,
											Required: true,
										},
									},
								},
							},

							"tcp_retry_events": {
								Type:     schema.TypeSet,
								Optional: true,
								MinItems: 0,
								Elem:     &schema.Schema{Type: schema.TypeString},
								Set:      schema.HashString,
							},
						},
					},
				},

				"timeout": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"idle": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"unit": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
										},

										"value": {
											Type:     schema.TypeInt,
											Required: true,
										},
									},
								},
							},

							"per_request": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"unit": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
										},

										"value": {
											Type:     schema.TypeInt,
											Required: true,
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

func resourceRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := &appmesh.CreateRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
		Spec:              expandRouteSpec(d.Get("spec").([]interface{})),
		Tags:              Tags(tags.IgnoreAWS()),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating App Mesh route: %#v", req)
	resp, err := conn.CreateRouteWithContext(ctx, req)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh route: %s", err)
	}

	d.SetId(aws.StringValue(resp.Route.Metadata.Uid))

	return append(diags, resourceRouteRead(ctx, d, meta)...)
}

func resourceRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	var resp *appmesh.DescribeRouteOutput

	err := resource.RetryContext(ctx, propagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.DescribeRouteWithContext(ctx, req)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.DescribeRouteWithContext(ctx, req)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		log.Printf("[WARN] App Mesh Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Route: %s", err)
	}

	if resp == nil || resp.Route == nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Route: empty response")
	}

	if aws.StringValue(resp.Route.Status.Status) == appmesh.RouteStatusCodeDeleted {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading App Mesh Route: %s after creation", aws.StringValue(resp.Route.Status.Status))
		}

		log.Printf("[WARN] App Mesh Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	arn := aws.StringValue(resp.Route.Metadata.Arn)
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
		return sdkdiag.AppendErrorf(diags, "listing tags for App Mesh route (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateRouteInput{
			MeshName:          aws.String(d.Get("mesh_name").(string)),
			RouteName:         aws.String(d.Get("name").(string)),
			VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
			Spec:              expandRouteSpec(v.([]interface{})),
		}
		if v, ok := d.GetOk("mesh_owner"); ok {
			req.MeshOwner = aws.String(v.(string))
		}

		log.Printf("[DEBUG] Updating App Mesh route: %#v", req)
		_, err := conn.UpdateRouteWithContext(ctx, req)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh route: %s", err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, arn, o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh route (%s) tags: %s", arn, err)
		}
	}

	return append(diags, resourceRouteRead(ctx, d, meta)...)
}

func resourceRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn()

	log.Printf("[DEBUG] Deleting App Mesh Route: %s", d.Id())
	_, err := conn.DeleteRouteWithContext(ctx, &appmesh.DeleteRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	})
	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh route: %s", err)
	}

	return diags
}

func resourceRouteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-router-name/route-name'", d.Id())
	}

	mesh := parts[0]
	vrName := parts[1]
	name := parts[2]
	log.Printf("[DEBUG] Importing App Mesh route %s from mesh %s/virtual router %s ", name, mesh, vrName)

	conn := meta.(*conns.AWSClient).AppMeshConn()

	resp, err := conn.DescribeRouteWithContext(ctx, &appmesh.DescribeRouteInput{
		MeshName:          aws.String(mesh),
		RouteName:         aws.String(name),
		VirtualRouterName: aws.String(vrName),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.Route.Metadata.Uid))
	d.Set("name", resp.Route.RouteName)
	d.Set("mesh_name", resp.Route.MeshName)
	d.Set("virtual_router_name", resp.Route.VirtualRouterName)

	return []*schema.ResourceData{d}, nil
}
