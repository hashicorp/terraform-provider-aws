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
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appmesh_gateway_route", name="Gateway Route")
// @Tags(identifierAttribute="arn")
func resourceGatewayRoute() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGatewayRouteCreate,
		ReadWithoutTimeout:   resourceGatewayRouteRead,
		UpdateWithoutTimeout: resourceGatewayRouteUpdate,
		DeleteWithoutTimeout: resourceGatewayRouteDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceGatewayRouteImport,
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
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
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
				"resource_owner": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"spec":            resourceGatewayRouteSpecSchema(),
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				"virtual_gateway_name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 255),
				},
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceGatewayRouteSpecSchema() *schema.Schema {
	// httpRouteSchema returns the schema for `http_route` and `http2_route` attributes.
	httpRouteSchema := func(attrName string) *schema.Schema {
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
								"rewrite": {
									Type:     schema.TypeList,
									Optional: true,
									MinItems: 1,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"hostname": {
												Type:     schema.TypeList,
												Optional: true,
												MinItems: 1,
												MaxItems: 1,
												Elem: &schema.Resource{
													Schema: map[string]*schema.Schema{
														"default_target_hostname": {
															Type:         schema.TypeString,
															Required:     true,
															ValidateFunc: validation.StringInSlice([]string{"ENABLED", "DISABLED"}, false),
														},
													},
												},
												AtLeastOneOf: []string{
													fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.hostname", attrName),
													fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.path", attrName),
													fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.prefix", attrName),
												},
											},
											"path": {
												Type:     schema.TypeList,
												Optional: true,
												MinItems: 1,
												MaxItems: 1,
												Elem: &schema.Resource{
													Schema: map[string]*schema.Schema{
														"exact": {
															Type:         schema.TypeString,
															Required:     true,
															ValidateFunc: validation.StringLenBetween(1, 255),
														},
													},
												},
												AtLeastOneOf: []string{
													fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.hostname", attrName),
													fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.path", attrName),
													fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.prefix", attrName),
												},
											},
											"prefix": {
												Type:     schema.TypeList,
												Optional: true,
												MinItems: 1,
												MaxItems: 1,
												Elem: &schema.Resource{
													Schema: map[string]*schema.Schema{
														"default_prefix": {
															Type:         schema.TypeString,
															Optional:     true,
															ValidateFunc: validation.StringInSlice([]string{"ENABLED", "DISABLED"}, false),
															ExactlyOneOf: []string{
																fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.prefix.0.default_prefix", attrName),
																fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.prefix.0.value", attrName),
															},
														},
														"value": {
															Type:         schema.TypeString,
															Optional:     true,
															ValidateFunc: validation.StringMatch(regexache.MustCompile(`^/`), "must start with /"),
															ExactlyOneOf: []string{
																fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.prefix.0.default_prefix", attrName),
																fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.prefix.0.value", attrName),
															},
														},
													},
												},
												AtLeastOneOf: []string{
													fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.hostname", attrName),
													fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.path", attrName),
													fmt.Sprintf("spec.0.%s.0.action.0.rewrite.0.prefix", attrName),
												},
											},
										},
									},
								},
								"target": {
									Type:     schema.TypeList,
									Required: true,
									MinItems: 1,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"port": {
												Type:         schema.TypeInt,
												Optional:     true,
												ValidateFunc: validation.IsPortNumber,
											},
											"virtual_service": {
												Type:     schema.TypeList,
												Required: true,
												MinItems: 1,
												MaxItems: 1,
												Elem: &schema.Resource{
													Schema: map[string]*schema.Schema{
														"virtual_service_name": {
															Type:         schema.TypeString,
															Required:     true,
															ValidateFunc: validation.StringLenBetween(1, 255),
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
								"hostname": {
									Type:     schema.TypeList,
									Optional: true,
									MinItems: 1,
									MaxItems: 1,
									Elem: &schema.Resource{
										Schema: map[string]*schema.Schema{
											"exact": {
												Type:     schema.TypeString,
												Optional: true,
												ExactlyOneOf: []string{
													fmt.Sprintf("spec.0.%s.0.match.0.hostname.0.exact", attrName),
													fmt.Sprintf("spec.0.%s.0.match.0.hostname.0.suffix", attrName),
												},
											},
											"suffix": {
												Type:     schema.TypeString,
												Optional: true,
												ExactlyOneOf: []string{
													fmt.Sprintf("spec.0.%s.0.match.0.hostname.0.exact", attrName),
													fmt.Sprintf("spec.0.%s.0.match.0.hostname.0.suffix", attrName),
												},
											},
										},
									},
									AtLeastOneOf: []string{
										fmt.Sprintf("spec.0.%s.0.match.0.hostname", attrName),
										fmt.Sprintf("spec.0.%s.0.match.0.path", attrName),
										fmt.Sprintf("spec.0.%s.0.match.0.prefix", attrName),
									},
								},
								"path": {
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
									AtLeastOneOf: []string{
										fmt.Sprintf("spec.0.%s.0.match.0.hostname", attrName),
										fmt.Sprintf("spec.0.%s.0.match.0.path", attrName),
										fmt.Sprintf("spec.0.%s.0.match.0.prefix", attrName),
									},
								},
								"port": {
									Type:         schema.TypeInt,
									Optional:     true,
									ValidateFunc: validation.IsPortNumber,
								},
								"prefix": {
									Type:         schema.TypeString,
									Optional:     true,
									ValidateFunc: validation.StringMatch(regexache.MustCompile(`^/`), "must start with /"),
									AtLeastOneOf: []string{
										fmt.Sprintf("spec.0.%s.0.match.0.hostname", attrName),
										fmt.Sprintf("spec.0.%s.0.match.0.path", attrName),
										fmt.Sprintf("spec.0.%s.0.match.0.prefix", attrName),
									},
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
											"name": {
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
			},
			ExactlyOneOf: []string{
				"spec.0.grpc_route",
				"spec.0.http2_route",
				"spec.0.http_route",
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
										"target": {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"port": {
														Type:         schema.TypeInt,
														Optional:     true,
														ValidateFunc: validation.IsPortNumber,
													},
													"virtual_service": {
														Type:     schema.TypeList,
														Required: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"virtual_service_name": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringLenBetween(1, 255),
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
							"match": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"port": {
											Type:         schema.TypeInt,
											Optional:     true,
											ValidateFunc: validation.IsPortNumber,
										},
										"service_name": {
											Type:     schema.TypeString,
											Required: true,
										},
									},
								},
							},
						},
					},
					ExactlyOneOf: []string{
						"spec.0.grpc_route",
						"spec.0.http2_route",
						"spec.0.http_route",
					},
				},
				"http_route":  httpRouteSchema("http_route"),
				"http2_route": httpRouteSchema("http2_route"),
				"priority": {
					Type:         schema.TypeInt,
					Optional:     true,
					ValidateFunc: validation.IntBetween(0, 1000),
				},
			},
		},
	}
}

func resourceGatewayRouteCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	name := d.Get("name").(string)
	input := &appmesh.CreateGatewayRouteInput{
		GatewayRouteName:   aws.String(name),
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		Spec:               expandGatewayRouteSpec(d.Get("spec").([]interface{})),
		Tags:               getTagsIn(ctx),
		VirtualGatewayName: aws.String(d.Get("virtual_gateway_name").(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	output, err := conn.CreateGatewayRoute(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Gateway Route (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.GatewayRoute.Metadata.Uid))

	return append(diags, resourceGatewayRouteRead(ctx, d, meta)...)
}

func resourceGatewayRouteRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findGatewayRouteByFourPartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), d.Get("virtual_gateway_name").(string), d.Get("name").(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Mesh Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Gateway Route (%s): %s", d.Id(), err)
	}

	gatewayRoute := outputRaw.(*awstypes.GatewayRouteData)

	arn := aws.ToString(gatewayRoute.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", gatewayRoute.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", gatewayRoute.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", gatewayRoute.MeshName)
	d.Set("mesh_owner", gatewayRoute.Metadata.MeshOwner)
	d.Set("name", gatewayRoute.GatewayRouteName)
	d.Set("resource_owner", gatewayRoute.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenGatewayRouteSpec(gatewayRoute.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}
	d.Set("virtual_gateway_name", gatewayRoute.VirtualGatewayName)

	return diags
}

func resourceGatewayRouteUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	if d.HasChange("spec") {
		input := &appmesh.UpdateGatewayRouteInput{
			GatewayRouteName:   aws.String(d.Get("name").(string)),
			MeshName:           aws.String(d.Get("mesh_name").(string)),
			Spec:               expandGatewayRouteSpec(d.Get("spec").([]interface{})),
			VirtualGatewayName: aws.String(d.Get("virtual_gateway_name").(string)),
		}

		if v, ok := d.GetOk("mesh_owner"); ok {
			input.MeshOwner = aws.String(v.(string))
		}

		_, err := conn.UpdateGatewayRoute(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Gateway Route (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGatewayRouteRead(ctx, d, meta)...)
}

func resourceGatewayRouteDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	log.Printf("[DEBUG] Deleting App Mesh Gateway Route: %s", d.Id())
	input := &appmesh.DeleteGatewayRouteInput{
		GatewayRouteName:   aws.String(d.Get("name").(string)),
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualGatewayName: aws.String(d.Get("virtual_gateway_name").(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	_, err := conn.DeleteGatewayRoute(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh Gateway Route (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceGatewayRouteImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-gateway-name/gateway-route-name'", d.Id())
	}

	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)
	meshName := parts[0]
	virtualGatewayName := parts[1]
	name := parts[2]

	gatewayRoute, err := findGatewayRouteByFourPartKey(ctx, conn, meshName, "", virtualGatewayName, name)

	if err != nil {
		return nil, err
	}

	d.SetId(aws.ToString(gatewayRoute.Metadata.Uid))
	d.Set("mesh_name", gatewayRoute.MeshName)
	d.Set("name", gatewayRoute.GatewayRouteName)
	d.Set("virtual_gateway_name", gatewayRoute.VirtualGatewayName)

	return []*schema.ResourceData{d}, nil
}

func findGatewayRouteByFourPartKey(ctx context.Context, conn *appmesh.Client, meshName, meshOwner, virtualGatewayName, name string) (*awstypes.GatewayRouteData, error) {
	input := &appmesh.DescribeGatewayRouteInput{
		GatewayRouteName:   aws.String(name),
		MeshName:           aws.String(meshName),
		VirtualGatewayName: aws.String(virtualGatewayName),
	}
	if meshOwner != "" {
		input.MeshOwner = aws.String(meshOwner)
	}

	output, err := findGatewayRoute(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.Status.Status == awstypes.GatewayRouteStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(output.Status.Status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findGatewayRoute(ctx context.Context, conn *appmesh.Client, input *appmesh.DescribeGatewayRouteInput) (*awstypes.GatewayRouteData, error) {
	output, err := conn.DescribeGatewayRoute(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.GatewayRoute == nil || output.GatewayRoute.Metadata == nil || output.GatewayRoute.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.GatewayRoute, nil
}

func expandGatewayRouteSpec(vSpec []interface{}) *awstypes.GatewayRouteSpec {
	if len(vSpec) == 0 || vSpec[0] == nil {
		return nil
	}

	spec := &awstypes.GatewayRouteSpec{}

	mSpec := vSpec[0].(map[string]interface{})

	if vGrpcRoute, ok := mSpec["grpc_route"].([]interface{}); ok {
		spec.GrpcRoute = expandGRPCGatewayRoute(vGrpcRoute)
	}

	if vHttp2Route, ok := mSpec["http2_route"].([]interface{}); ok {
		spec.Http2Route = expandHTTPGatewayRoute(vHttp2Route)
	}

	if vHttpRoute, ok := mSpec["http_route"].([]interface{}); ok {
		spec.HttpRoute = expandHTTPGatewayRoute(vHttpRoute)
	}

	if vPriority, ok := mSpec["priority"].(int); ok && vPriority > 0 {
		spec.Priority = aws.Int32(int32(vPriority))
	}

	return spec
}

func expandGatewayRouteTarget(vRouteTarget []interface{}) *awstypes.GatewayRouteTarget {
	if len(vRouteTarget) == 0 || vRouteTarget[0] == nil {
		return nil
	}

	routeTarget := &awstypes.GatewayRouteTarget{}

	mRouteTarget := vRouteTarget[0].(map[string]interface{})

	if vVirtualService, ok := mRouteTarget["virtual_service"].([]interface{}); ok && len(vVirtualService) > 0 && vVirtualService[0] != nil {
		virtualService := &awstypes.GatewayRouteVirtualService{}

		mVirtualService := vVirtualService[0].(map[string]interface{})

		if vVirtualServiceName, ok := mVirtualService["virtual_service_name"].(string); ok && vVirtualServiceName != "" {
			virtualService.VirtualServiceName = aws.String(vVirtualServiceName)
		}

		routeTarget.VirtualService = virtualService
	}

	if vPort, ok := mRouteTarget["port"].(int); ok && vPort > 0 {
		routeTarget.Port = aws.Int32(int32(vPort))
	}

	return routeTarget
}

func expandGRPCGatewayRoute(vGrpcRoute []interface{}) *awstypes.GrpcGatewayRoute {
	if len(vGrpcRoute) == 0 || vGrpcRoute[0] == nil {
		return nil
	}

	route := &awstypes.GrpcGatewayRoute{}

	mGrpcRoute := vGrpcRoute[0].(map[string]interface{})

	if vRouteAction, ok := mGrpcRoute["action"].([]interface{}); ok && len(vRouteAction) > 0 && vRouteAction[0] != nil {
		routeAction := &awstypes.GrpcGatewayRouteAction{}

		mRouteAction := vRouteAction[0].(map[string]interface{})

		if vRouteTarget, ok := mRouteAction["target"].([]interface{}); ok {
			routeAction.Target = expandGatewayRouteTarget(vRouteTarget)
		}

		route.Action = routeAction
	}

	if vRouteMatch, ok := mGrpcRoute["match"].([]interface{}); ok && len(vRouteMatch) > 0 && vRouteMatch[0] != nil {
		routeMatch := &awstypes.GrpcGatewayRouteMatch{}

		mRouteMatch := vRouteMatch[0].(map[string]interface{})

		if vServiceName, ok := mRouteMatch["service_name"].(string); ok && vServiceName != "" {
			routeMatch.ServiceName = aws.String(vServiceName)
		}

		if vPort, ok := mRouteMatch["port"].(int); ok && vPort > 0 {
			routeMatch.Port = aws.Int32(int32(vPort))
		}

		route.Match = routeMatch
	}

	return route
}

func expandHTTPGatewayRouteRewrite(vHttpRouteRewrite []interface{}) *awstypes.HttpGatewayRouteRewrite {
	if len(vHttpRouteRewrite) == 0 || vHttpRouteRewrite[0] == nil {
		return nil
	}
	mRouteRewrite := vHttpRouteRewrite[0].(map[string]interface{})
	routeRewrite := &awstypes.HttpGatewayRouteRewrite{}

	if vRouteHostnameRewrite, ok := mRouteRewrite["hostname"].([]interface{}); ok && len(vRouteHostnameRewrite) > 0 && vRouteHostnameRewrite[0] != nil {
		mRouteHostnameRewrite := vRouteHostnameRewrite[0].(map[string]interface{})
		routeHostnameRewrite := &awstypes.GatewayRouteHostnameRewrite{}
		if vDefaultTargetHostname, ok := mRouteHostnameRewrite["default_target_hostname"].(string); ok && vDefaultTargetHostname != "" {
			routeHostnameRewrite.DefaultTargetHostname = awstypes.DefaultGatewayRouteRewrite(vDefaultTargetHostname)
		}
		routeRewrite.Hostname = routeHostnameRewrite
	}

	if vRoutePathRewrite, ok := mRouteRewrite["path"].([]interface{}); ok && len(vRoutePathRewrite) > 0 && vRoutePathRewrite[0] != nil {
		mRoutePathRewrite := vRoutePathRewrite[0].(map[string]interface{})
		routePathRewrite := &awstypes.HttpGatewayRoutePathRewrite{}
		if vExact, ok := mRoutePathRewrite["exact"].(string); ok && vExact != "" {
			routePathRewrite.Exact = aws.String(vExact)
		}
		routeRewrite.Path = routePathRewrite
	}

	if vRoutePrefixRewrite, ok := mRouteRewrite["prefix"].([]interface{}); ok && len(vRoutePrefixRewrite) > 0 && vRoutePrefixRewrite[0] != nil {
		mRoutePrefixRewrite := vRoutePrefixRewrite[0].(map[string]interface{})
		routePrefixRewrite := &awstypes.HttpGatewayRoutePrefixRewrite{}
		if vDefaultPrefix, ok := mRoutePrefixRewrite["default_prefix"].(string); ok && vDefaultPrefix != "" {
			routePrefixRewrite.DefaultPrefix = awstypes.DefaultGatewayRouteRewrite(vDefaultPrefix)
		}
		if vValue, ok := mRoutePrefixRewrite["value"].(string); ok && vValue != "" {
			routePrefixRewrite.Value = aws.String(vValue)
		}
		routeRewrite.Prefix = routePrefixRewrite
	}

	return routeRewrite
}

func expandHTTPGatewayRouteMatch(vHttpRouteMatch []interface{}) *awstypes.HttpGatewayRouteMatch {
	if len(vHttpRouteMatch) == 0 || vHttpRouteMatch[0] == nil {
		return nil
	}

	routeMatch := &awstypes.HttpGatewayRouteMatch{}

	mRouteMatch := vHttpRouteMatch[0].(map[string]interface{})

	if vPort, ok := mRouteMatch["port"].(int); ok && vPort > 0 {
		routeMatch.Port = aws.Int32(int32(vPort))
	}

	if vPrefix, ok := mRouteMatch["prefix"].(string); ok && vPrefix != "" {
		routeMatch.Prefix = aws.String(vPrefix)
	}

	if vHeaders, ok := mRouteMatch["header"].(*schema.Set); ok && vHeaders.Len() > 0 {
		headers := []awstypes.HttpGatewayRouteHeader{}

		for _, vHeader := range vHeaders.List() {
			header := awstypes.HttpGatewayRouteHeader{}

			mHeader := vHeader.(map[string]interface{})

			if vInvert, ok := mHeader["invert"].(bool); ok {
				header.Invert = aws.Bool(vInvert)
			}
			if vName, ok := mHeader["name"].(string); ok && vName != "" {
				header.Name = aws.String(vName)
			}

			if vMatch, ok := mHeader["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
				mMatch := vMatch[0].(map[string]interface{})

				if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
					header.Match = &awstypes.HeaderMatchMethodMemberExact{Value: vExact}
				}
				if vPrefix, ok := mMatch["prefix"].(string); ok && vPrefix != "" {
					header.Match = &awstypes.HeaderMatchMethodMemberPrefix{Value: vPrefix}
				}
				if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
					header.Match = &awstypes.HeaderMatchMethodMemberRegex{Value: vRegex}
				}
				if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
					header.Match = &awstypes.HeaderMatchMethodMemberSuffix{Value: vSuffix}
				}

				if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
					memberRange := awstypes.MatchRange{}

					mRange := vRange[0].(map[string]interface{})

					if vEnd, ok := mRange["end"].(int); ok && vEnd > 0 {
						memberRange.End = aws.Int64(int64(vEnd))
					}
					if vStart, ok := mRange["start"].(int); ok && vStart > 0 {
						memberRange.Start = aws.Int64(int64(vStart))
					}
					header.Match = &awstypes.HeaderMatchMethodMemberRange{Value: memberRange}
				}
			}

			headers = append(headers, header)
		}

		routeMatch.Headers = headers
	}

	if vHostname, ok := mRouteMatch["hostname"].([]interface{}); ok && len(vHostname) > 0 && vHostname[0] != nil {
		hostnameMatch := &awstypes.GatewayRouteHostnameMatch{}

		mHostname := vHostname[0].(map[string]interface{})

		if vExact, ok := mHostname["exact"].(string); ok && vExact != "" {
			hostnameMatch.Exact = aws.String(vExact)
		}
		if vSuffix, ok := mHostname["suffix"].(string); ok && vSuffix != "" {
			hostnameMatch.Suffix = aws.String(vSuffix)
		}

		routeMatch.Hostname = hostnameMatch
	}

	if vPath, ok := mRouteMatch["path"].([]interface{}); ok && len(vPath) > 0 && vPath[0] != nil {
		pathMatch := &awstypes.HttpPathMatch{}

		mHostname := vPath[0].(map[string]interface{})

		if vExact, ok := mHostname["exact"].(string); ok && vExact != "" {
			pathMatch.Exact = aws.String(vExact)
		}
		if vRegex, ok := mHostname["regex"].(string); ok && vRegex != "" {
			pathMatch.Regex = aws.String(vRegex)
		}

		routeMatch.Path = pathMatch
	}

	if vQueryParameters, ok := mRouteMatch["query_parameter"].(*schema.Set); ok && vQueryParameters.Len() > 0 {
		queryParameters := []awstypes.HttpQueryParameter{}

		for _, vQueryParameter := range vQueryParameters.List() {
			queryParameter := awstypes.HttpQueryParameter{}

			mQueryParameter := vQueryParameter.(map[string]interface{})

			if vName, ok := mQueryParameter["name"].(string); ok && vName != "" {
				queryParameter.Name = aws.String(vName)
			}

			if vMatch, ok := mQueryParameter["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
				queryParameter.Match = &awstypes.QueryParameterMatch{}

				mMatch := vMatch[0].(map[string]interface{})

				if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
					queryParameter.Match.Exact = aws.String(vExact)
				}
			}

			queryParameters = append(queryParameters, queryParameter)
		}

		routeMatch.QueryParameters = queryParameters
	}

	return routeMatch
}

func expandHTTPGatewayRoute(vHttpRoute []interface{}) *awstypes.HttpGatewayRoute {
	if len(vHttpRoute) == 0 || vHttpRoute[0] == nil {
		return nil
	}

	route := &awstypes.HttpGatewayRoute{}

	mHttpRoute := vHttpRoute[0].(map[string]interface{})

	if vRouteAction, ok := mHttpRoute["action"].([]interface{}); ok && len(vRouteAction) > 0 && vRouteAction[0] != nil {
		routeAction := &awstypes.HttpGatewayRouteAction{}

		mRouteAction := vRouteAction[0].(map[string]interface{})

		if vRouteTarget, ok := mRouteAction["target"].([]interface{}); ok {
			routeAction.Target = expandGatewayRouteTarget(vRouteTarget)
		}

		if vRouteRewrite, ok := mRouteAction["rewrite"].([]interface{}); ok {
			routeAction.Rewrite = expandHTTPGatewayRouteRewrite(vRouteRewrite)
		}

		route.Action = routeAction
	}

	if vRouteMatch, ok := mHttpRoute["match"].([]interface{}); ok && len(vRouteMatch) > 0 && vRouteMatch[0] != nil {
		route.Match = expandHTTPGatewayRouteMatch(vRouteMatch)
	}

	return route
}

func flattenGatewayRouteSpec(spec *awstypes.GatewayRouteSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{
		"grpc_route":  flattenGRPCGatewayRoute(spec.GrpcRoute),
		"http2_route": flattenHTTPGatewayRoute(spec.Http2Route),
		"http_route":  flattenHTTPGatewayRoute(spec.HttpRoute),
		"priority":    int(aws.ToInt32(spec.Priority)),
	}

	return []interface{}{mSpec}
}

func flattenGatewayRouteTarget(routeTarget *awstypes.GatewayRouteTarget) []interface{} {
	if routeTarget == nil {
		return []interface{}{}
	}

	mRouteTarget := map[string]interface{}{
		"port": int(aws.ToInt32(routeTarget.Port)),
	}

	if virtualService := routeTarget.VirtualService; virtualService != nil {
		mVirtualService := map[string]interface{}{
			"virtual_service_name": aws.ToString(virtualService.VirtualServiceName),
		}

		mRouteTarget["virtual_service"] = []interface{}{mVirtualService}
	}

	return []interface{}{mRouteTarget}
}

func flattenGRPCGatewayRoute(grpcRoute *awstypes.GrpcGatewayRoute) []interface{} {
	if grpcRoute == nil {
		return []interface{}{}
	}

	mGrpcRoute := map[string]interface{}{}

	if routeAction := grpcRoute.Action; routeAction != nil {
		mRouteAction := map[string]interface{}{
			"target": flattenGatewayRouteTarget(routeAction.Target),
		}

		mGrpcRoute["action"] = []interface{}{mRouteAction}
	}

	if routeMatch := grpcRoute.Match; routeMatch != nil {
		mRouteMatch := map[string]interface{}{
			"service_name": aws.ToString(routeMatch.ServiceName),
		}
		if routeMatch.Port != nil {
			mRouteMatch["port"] = int(aws.ToInt32(routeMatch.Port))
		}

		mGrpcRoute["match"] = []interface{}{mRouteMatch}
	}

	return []interface{}{mGrpcRoute}
}

func flattenHTTPGatewayRouteMatch(routeMatch *awstypes.HttpGatewayRouteMatch) []interface{} {
	if routeMatch == nil {
		return []interface{}{}
	}

	mRouteMatch := map[string]interface{}{}

	if routeMatch.Port != nil {
		mRouteMatch["port"] = int(aws.ToInt32(routeMatch.Port))
	}

	if routeMatch.Prefix != nil {
		mRouteMatch["prefix"] = aws.ToString(routeMatch.Prefix)
	}

	vHeaders := []interface{}{}

	for _, header := range routeMatch.Headers {
		mHeader := map[string]interface{}{
			"invert": aws.ToBool(header.Invert),
			"name":   aws.ToString(header.Name),
		}

		mMatch := map[string]interface{}{}

		if match := header.Match; match != nil {
			switch v := match.(type) {
			case *awstypes.HeaderMatchMethodMemberExact:
				mMatch["exact"] = v.Value
			case *awstypes.HeaderMatchMethodMemberPrefix:
				mMatch["prefix"] = v.Value
			case *awstypes.HeaderMatchMethodMemberRegex:
				mMatch["regex"] = v.Value
			case *awstypes.HeaderMatchMethodMemberSuffix:
				mMatch["suffix"] = v.Value
			case *awstypes.HeaderMatchMethodMemberRange:
				mRange := map[string]interface{}{
					"end":   int(aws.ToInt64(v.Value.End)),
					"start": int(aws.ToInt64(v.Value.Start)),
				}
				mMatch["range"] = []interface{}{mRange}
			}

			mHeader["match"] = []interface{}{mMatch}
		}

		vHeaders = append(vHeaders, mHeader)
	}

	mRouteMatch["header"] = vHeaders

	if hostname := routeMatch.Hostname; hostname != nil {
		mHostname := map[string]interface{}{}

		if hostname.Exact != nil {
			mHostname["exact"] = aws.ToString(hostname.Exact)
		}
		if hostname.Suffix != nil {
			mHostname["suffix"] = aws.ToString(hostname.Suffix)
		}

		mRouteMatch["hostname"] = []interface{}{mHostname}
	}

	if path := routeMatch.Path; path != nil {
		mPath := map[string]interface{}{}

		if path.Exact != nil {
			mPath["exact"] = aws.ToString(path.Exact)
		}
		if path.Regex != nil {
			mPath["regex"] = aws.ToString(path.Regex)
		}

		mRouteMatch["path"] = []interface{}{mPath}
	}

	vQueryParameters := []interface{}{}

	for _, queryParameter := range routeMatch.QueryParameters {
		mQueryParameter := map[string]interface{}{
			"name": aws.ToString(queryParameter.Name),
		}

		if match := queryParameter.Match; match != nil {
			mMatch := map[string]interface{}{
				"exact": aws.ToString(match.Exact),
			}

			mQueryParameter["match"] = []interface{}{mMatch}
		}

		vQueryParameters = append(vQueryParameters, mQueryParameter)
	}

	mRouteMatch["query_parameter"] = vQueryParameters

	return []interface{}{mRouteMatch}
}

func flattenHTTPGatewayRouteRewrite(routeRewrite *awstypes.HttpGatewayRouteRewrite) []interface{} {
	if routeRewrite == nil {
		return []interface{}{}
	}

	mRouteRewrite := map[string]interface{}{}

	if rewriteHostname := routeRewrite.Hostname; rewriteHostname != nil {
		mRewriteHostname := map[string]interface{}{
			"default_target_hostname": string(rewriteHostname.DefaultTargetHostname),
		}
		mRouteRewrite["hostname"] = []interface{}{mRewriteHostname}
	}

	if rewritePath := routeRewrite.Path; rewritePath != nil {
		mRewritePath := map[string]interface{}{
			"exact": aws.ToString(rewritePath.Exact),
		}
		mRouteRewrite["path"] = []interface{}{mRewritePath}
	}

	if rewritePrefix := routeRewrite.Prefix; rewritePrefix != nil {
		mRewritePrefix := map[string]interface{}{
			"default_prefix": string(rewritePrefix.DefaultPrefix),
		}
		if rewritePrefixValue := rewritePrefix.Value; rewritePrefixValue != nil {
			mRewritePrefix["value"] = aws.ToString(rewritePrefix.Value)
		}
		mRouteRewrite["prefix"] = []interface{}{mRewritePrefix}
	}

	return []interface{}{mRouteRewrite}
}

func flattenHTTPGatewayRoute(httpRoute *awstypes.HttpGatewayRoute) []interface{} {
	if httpRoute == nil {
		return []interface{}{}
	}

	mHttpRoute := map[string]interface{}{}

	if routeAction := httpRoute.Action; routeAction != nil {
		mRouteAction := map[string]interface{}{
			"target":  flattenGatewayRouteTarget(routeAction.Target),
			"rewrite": flattenHTTPGatewayRouteRewrite(routeAction.Rewrite),
		}

		mHttpRoute["action"] = []interface{}{mRouteAction}
	}

	if routeMatch := httpRoute.Match; routeMatch != nil {
		mHttpRoute["match"] = flattenHTTPGatewayRouteMatch(routeMatch)
	}

	return []interface{}{mHttpRoute}
}
