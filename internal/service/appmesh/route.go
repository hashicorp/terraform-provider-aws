// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appmesh"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appmesh/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appmesh_route", name="Route")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/appmesh/types;types.RouteData")
// @Testing(serialize=true)
// @Testing(importStateIdFunc=testAccRouteImportStateIdFunc)
func resourceRoute() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRouteCreate,
		ReadWithoutTimeout:   resourceRouteRead,
		UpdateWithoutTimeout: resourceRouteUpdate,
		DeleteWithoutTimeout: resourceRouteDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceRouteImport,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrCreatedDate: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrLastUpdatedDate: {
					Type:     schema.TypeString,
					Computed: true,
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
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				names.AttrResourceOwner: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"spec":            resourceRouteSpecSchema(),
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"virtual_router_name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
			}
		},
	}
}

func resourceRouteSpecSchema() *schema.Schema {
	// httpRouteSchema returns the schema for `http_route` and `http2_route` attributes.
	httpRouteSchema := func() *schema.Schema {
		return &schema.Schema{
			Type:     schema.TypeList,
			Optional: true,
			MinItems: 0,
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					names.AttrAction: {
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
											names.AttrPort: {
												Type:         schema.TypeInt,
												Optional:     true,
												Computed:     true,
												ValidateFunc: validation.IsPortNumber,
											},
											"virtual_node": {
												Type:         schema.TypeString,
												Required:     true,
												ValidateFunc: validation.StringLenBetween(1, 255),
											},
											names.AttrWeight: {
												Type:         schema.TypeInt,
												Required:     true,
												ValidateFunc: validation.IntBetween(0, 100),
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
								names.AttrHeader: {
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
														names.AttrPrefix: {
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
											names.AttrName: {
												Type:         schema.TypeString,
												Required:     true,
												ValidateFunc: validation.StringLenBetween(1, 50),
											},
										},
									},
								},
								"method": {
									Type:             schema.TypeString,
									Optional:         true,
									ValidateDiagFunc: enum.Validate[awstypes.HttpMethod](),
								},
								names.AttrPath: {
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
											"regex": {
												Type:         schema.TypeString,
												Optional:     true,
												ValidateFunc: validation.StringLenBetween(1, 255),
											},
										},
									},
								},
								names.AttrPort: {
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IsPortNumber,
								},
								names.AttrPrefix: {
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: validation.StringMatch(regexache.MustCompile(`^/`), "must start with /"),
								},
								"query_parameter": {
									Type:     schema.TypeSet,
									Optional: true,
									MinItems: 0,
									MaxItems: 10,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"match": {
												Type:     schema.TypeList,
												Optional: true,
												MinItems: 0,
												MaxItems: 1,
												Elem: &schema.Resource{
													Schema: map[string]*schema.Schema{
														"exact": {
															Type:     schema.TypeString,
															Optional: true,
														},
													},
												},
											},
											names.AttrName: {
												Type:     schema.TypeString,
												Required: true,
											},
										},
									},
								},
								"scheme": {
									Type:             schema.TypeString,
									Optional:         true,
									ValidateDiagFunc: enum.Validate[awstypes.HttpScheme](),
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
											names.AttrUnit: {
												Type:             schema.TypeString,
												Required:         true,
												ValidateDiagFunc: enum.Validate[awstypes.DurationUnit](),
											},
											names.AttrValue: {
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
								},
							},
						},
					},
					names.AttrTimeout: {
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
											names.AttrUnit: {
												Type:             schema.TypeString,
												Required:         true,
												ValidateDiagFunc: enum.Validate[awstypes.DurationUnit](),
											},
											names.AttrValue: {
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
											names.AttrUnit: {
												Type:             schema.TypeString,
												Required:         true,
												ValidateDiagFunc: enum.Validate[awstypes.DurationUnit](),
											},
											names.AttrValue: {
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

	return &schema.Schema{
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
							names.AttrAction: {
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
													names.AttrPort: {
														Type:         schema.TypeInt,
														Optional:     true,
														Computed:     true,
														ValidateFunc: validation.IsPortNumber,
													},
													"virtual_node": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringLenBetween(1, 255),
													},
													names.AttrWeight: {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntBetween(0, 100),
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
																names.AttrPrefix: {
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
													names.AttrName: {
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
										names.AttrPort: {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IsPortNumber,
										},
										names.AttrPrefix: {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringLenBetween(0, 50),
										},
										names.AttrServiceName: {
											Type:     schema.TypeString,
											Optional: true,
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
										},
										"http_retry_events": {
											Type:     schema.TypeSet,
											Optional: true,
											MinItems: 0,
											Elem:     &schema.Schema{Type: schema.TypeString},
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
													names.AttrUnit: {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.DurationUnit](),
													},
													names.AttrValue: {
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
										},
									},
								},
							},
							names.AttrTimeout: {
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
													names.AttrUnit: {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.DurationUnit](),
													},
													names.AttrValue: {
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
													names.AttrUnit: {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.DurationUnit](),
													},
													names.AttrValue: {
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
				"http_route": func() *schema.Schema {
					schema := httpRouteSchema()
					schema.ConflictsWith = []string{"spec.0.grpc_route", "spec.0.http2_route", "spec.0.tcp_route"}
					return schema
				}(),
				"http2_route": func() *schema.Schema {
					schema := httpRouteSchema()
					schema.ConflictsWith = []string{"spec.0.grpc_route", "spec.0.http_route", "spec.0.tcp_route"}
					return schema
				}(),
				names.AttrPriority: {
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
							names.AttrAction: {
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
													names.AttrPort: {
														Type:         schema.TypeInt,
														Optional:     true,
														Computed:     true,
														ValidateFunc: validation.IsPortNumber,
													},
													"virtual_node": {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringLenBetween(1, 255),
													},
													names.AttrWeight: {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntBetween(0, 100),
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
										names.AttrPort: {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IsPortNumber,
										},
									},
								},
							},
							names.AttrTimeout: {
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
													names.AttrUnit: {
														Type:             schema.TypeString,
														Required:         true,
														ValidateDiagFunc: enum.Validate[awstypes.DurationUnit](),
													},
													names.AttrValue: {
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
	}
}

func resourceRouteCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appmesh.CreateRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(name),
		Spec:              expandRouteSpec(d.Get("spec").([]any)),
		Tags:              getTagsIn(ctx),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	output, err := conn.CreateRoute(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Route (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Route.Metadata.Uid))

	return append(diags, resourceRouteRead(ctx, d, meta)...)
}

func resourceRouteRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (any, error) {
		return findRouteByFourPartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), d.Get("virtual_router_name").(string), d.Get(names.AttrName).(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Mesh Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Route (%s): %s", d.Id(), err)
	}

	route := outputRaw.(*awstypes.RouteData)

	d.Set(names.AttrARN, route.Metadata.Arn)
	d.Set(names.AttrCreatedDate, route.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedDate, route.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", route.MeshName)
	d.Set("mesh_owner", route.Metadata.MeshOwner)
	d.Set(names.AttrName, route.RouteName)
	d.Set(names.AttrResourceOwner, route.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenRouteSpec(route.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}
	d.Set("virtual_router_name", route.VirtualRouterName)

	return diags
}

func resourceRouteUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	if d.HasChange("spec") {
		input := &appmesh.UpdateRouteInput{
			MeshName:          aws.String(d.Get("mesh_name").(string)),
			RouteName:         aws.String(d.Get(names.AttrName).(string)),
			Spec:              expandRouteSpec(d.Get("spec").([]any)),
			VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
		}

		if v, ok := d.GetOk("mesh_owner"); ok {
			input.MeshOwner = aws.String(v.(string))
		}

		_, err := conn.UpdateRoute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Route (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceRouteRead(ctx, d, meta)...)
}

func resourceRouteDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	log.Printf("[DEBUG] Deleting App Mesh Route: %s", d.Id())
	input := &appmesh.DeleteRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get(names.AttrName).(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	_, err := conn.DeleteRoute(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh Route (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceRouteImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-router-name/route-name'", d.Id())
	}

	meshName := parts[0]
	virtualRouterName := parts[1]
	name := parts[2]

	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	route, err := findRouteByFourPartKey(ctx, conn, meshName, "", virtualRouterName, name)

	if err != nil {
		return nil, err
	}

	d.SetId(aws.ToString(route.Metadata.Uid))
	d.Set("mesh_name", route.MeshName)
	d.Set(names.AttrName, route.RouteName)
	d.Set("virtual_router_name", route.VirtualRouterName)

	return []*schema.ResourceData{d}, nil
}

func findRouteByFourPartKey(ctx context.Context, conn *appmesh.Client, meshName, meshOwner, virtualRouterName, name string) (*awstypes.RouteData, error) {
	input := &appmesh.DescribeRouteInput{
		MeshName:          aws.String(meshName),
		RouteName:         aws.String(name),
		VirtualRouterName: aws.String(virtualRouterName),
	}
	if meshOwner != "" {
		input.MeshOwner = aws.String(meshOwner)
	}

	output, err := findRoute(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.Status.Status == awstypes.RouteStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(output.Status.Status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findRoute(ctx context.Context, conn *appmesh.Client, input *appmesh.DescribeRouteInput) (*awstypes.RouteData, error) {
	output, err := conn.DescribeRoute(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Route == nil || output.Route.Metadata == nil || output.Route.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Route, nil
}
