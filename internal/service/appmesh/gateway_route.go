package appmesh

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceGatewayRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceGatewayRouteCreate,
		Read:   resourceGatewayRouteRead,
		Update: resourceGatewayRouteUpdate,
		Delete: resourceGatewayRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceGatewayRouteImport,
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

			"virtual_gateway_name": {
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

						"http2_route": {
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
																	"spec.0.http2_route.0.action.0.rewrite.0.prefix",
																	"spec.0.http2_route.0.action.0.rewrite.0.hostname",
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
																				"spec.0.http2_route.0.action.0.rewrite.0.prefix.0.default_prefix",
																				"spec.0.http2_route.0.action.0.rewrite.0.prefix.0.value",
																			},
																		},
																		"value": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
																			ExactlyOneOf: []string{
																				"spec.0.http2_route.0.action.0.rewrite.0.prefix.0.default_prefix",
																				"spec.0.http2_route.0.action.0.rewrite.0.prefix.0.value",
																			},
																		},
																	},
																},
																AtLeastOneOf: []string{
																	"spec.0.http2_route.0.action.0.rewrite.0.prefix",
																	"spec.0.http2_route.0.action.0.rewrite.0.hostname",
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
												"prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
													AtLeastOneOf: []string{
														"spec.0.http2_route.0.match.0.prefix",
														"spec.0.http2_route.0.match.0.hostname",
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
																	"spec.0.http2_route.0.match.0.hostname.0.exact",
																	"spec.0.http2_route.0.match.0.hostname.0.suffix",
																},
															},
															"suffix": {
																Type:     schema.TypeString,
																Optional: true,
																ExactlyOneOf: []string{
																	"spec.0.http2_route.0.match.0.hostname.0.exact",
																	"spec.0.http2_route.0.match.0.hostname.0.suffix",
																},
															},
														},
													},
													AtLeastOneOf: []string{
														"spec.0.http2_route.0.match.0.prefix",
														"spec.0.http2_route.0.match.0.hostname",
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
						},

						"http_route": {
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
																	"spec.0.http_route.0.action.0.rewrite.0.prefix",
																	"spec.0.http_route.0.action.0.rewrite.0.hostname",
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
																				"spec.0.http_route.0.action.0.rewrite.0.prefix.0.default_prefix",
																				"spec.0.http_route.0.action.0.rewrite.0.prefix.0.value",
																			},
																		},
																		"value": {
																			Type:         schema.TypeString,
																			Optional:     true,
																			ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
																			ExactlyOneOf: []string{
																				"spec.0.http_route.0.action.0.rewrite.0.prefix.0.default_prefix",
																				"spec.0.http_route.0.action.0.rewrite.0.prefix.0.value",
																			},
																		},
																	},
																},
																AtLeastOneOf: []string{
																	"spec.0.http_route.0.action.0.rewrite.0.prefix",
																	"spec.0.http_route.0.action.0.rewrite.0.hostname",
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
												"prefix": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
													AtLeastOneOf: []string{
														"spec.0.http_route.0.match.0.prefix",
														"spec.0.http_route.0.match.0.hostname",
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
																	"spec.0.http_route.0.match.0.hostname.0.exact",
																	"spec.0.http_route.0.match.0.hostname.0.suffix",
																},
															},
															"suffix": {
																Type:     schema.TypeString,
																Optional: true,
																ExactlyOneOf: []string{
																	"spec.0.http_route.0.match.0.hostname.0.exact",
																	"spec.0.http_route.0.match.0.hostname.0.suffix",
																},
															},
														},
													},
													AtLeastOneOf: []string{
														"spec.0.http_route.0.match.0.prefix",
														"spec.0.http_route.0.match.0.hostname",
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

func resourceGatewayRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppMeshConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &appmesh.CreateGatewayRouteInput{
		GatewayRouteName:   aws.String(d.Get("name").(string)),
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		Spec:               expandGatewayRouteSpec(d.Get("spec").([]interface{})),
		Tags:               Tags(tags.IgnoreAWS()),
		VirtualGatewayName: aws.String(d.Get("virtual_gateway_name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating App Mesh gateway route: %s", input)
	output, err := conn.CreateGatewayRoute(input)

	if err != nil {
		return fmt.Errorf("error creating App Mesh gateway route: %w", err)
	}

	d.SetId(aws.StringValue(output.GatewayRoute.Metadata.Uid))

	return resourceGatewayRouteRead(d, meta)
}

func resourceGatewayRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppMeshConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var gatewayRoute *appmesh.GatewayRouteData

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		gatewayRoute, err = FindGatewayRoute(conn, d.Get("mesh_name").(string), d.Get("virtual_gateway_name").(string), d.Get("name").(string), d.Get("mesh_owner").(string))

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		gatewayRoute, err = FindGatewayRoute(conn, d.Get("mesh_name").(string), d.Get("virtual_gateway_name").(string), d.Get("name").(string), d.Get("mesh_owner").(string))
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		log.Printf("[WARN] App Mesh Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading App Mesh Gateway Route: %w", err)
	}

	if gatewayRoute == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading App Mesh Gateway Route: not found after creation")
		}

		log.Printf("[WARN] App Mesh Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(gatewayRoute.Status.Status) == appmesh.GatewayRouteStatusCodeDeleted {
		if d.IsNewResource() {
			return fmt.Errorf("error reading App Mesh Gateway Route: %s after creation", aws.StringValue(gatewayRoute.Status.Status))
		}

		log.Printf("[WARN] App Mesh Gateway Route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(gatewayRoute.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", gatewayRoute.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", gatewayRoute.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", gatewayRoute.MeshName)
	d.Set("mesh_owner", gatewayRoute.Metadata.MeshOwner)
	d.Set("name", gatewayRoute.GatewayRouteName)
	d.Set("resource_owner", gatewayRoute.Metadata.ResourceOwner)
	err = d.Set("spec", flattenGatewayRouteSpec(gatewayRoute.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %w", err)
	}
	d.Set("virtual_gateway_name", gatewayRoute.VirtualGatewayName)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for App Mesh gateway route (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceGatewayRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppMeshConn

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

		log.Printf("[DEBUG] Updating App Mesh gateway route: %s", input)
		_, err := conn.UpdateGatewayRoute(input)

		if err != nil {
			return fmt.Errorf("error updating App Mesh gateway route (%s): %w", d.Id(), err)
		}
	}

	arn := d.Get("arn").(string)
	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating App Mesh gateway route (%s) tags: %s", arn, err)
		}
	}

	return resourceGatewayRouteRead(d, meta)
}

func resourceGatewayRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppMeshConn

	log.Printf("[DEBUG] Deleting App Mesh gateway route (%s)", d.Id())
	_, err := conn.DeleteGatewayRoute(&appmesh.DeleteGatewayRouteInput{
		GatewayRouteName:   aws.String(d.Get("name").(string)),
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualGatewayName: aws.String(d.Get("virtual_gateway_name").(string)),
	})

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting App Mesh gateway route (%s) : %w", d.Id(), err)
	}

	return nil
}

func resourceGatewayRouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-gateway-name/gateway-route-name'", d.Id())
	}

	mesh := parts[0]
	vgName := parts[1]
	name := parts[2]
	log.Printf("[DEBUG] Importing App Mesh gateway route %s from mesh %s/virtual gateway %s ", name, mesh, vgName)

	conn := meta.(*conns.AWSClient).AppMeshConn

	gatewayRoute, err := FindGatewayRoute(conn, mesh, vgName, name, "")

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(gatewayRoute.Metadata.Uid))
	d.Set("mesh_name", gatewayRoute.MeshName)
	d.Set("name", gatewayRoute.GatewayRouteName)
	d.Set("virtual_gateway_name", gatewayRoute.VirtualGatewayName)

	return []*schema.ResourceData{d}, nil
}

func expandGatewayRouteSpec(vSpec []interface{}) *appmesh.GatewayRouteSpec {
	if len(vSpec) == 0 || vSpec[0] == nil {
		return nil
	}

	spec := &appmesh.GatewayRouteSpec{}

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

	return spec
}

func expandGatewayRouteTarget(vRouteTarget []interface{}) *appmesh.GatewayRouteTarget {
	if len(vRouteTarget) == 0 || vRouteTarget[0] == nil {
		return nil
	}

	routeTarget := &appmesh.GatewayRouteTarget{}

	mRouteTarget := vRouteTarget[0].(map[string]interface{})

	if vVirtualService, ok := mRouteTarget["virtual_service"].([]interface{}); ok && len(vVirtualService) > 0 && vVirtualService[0] != nil {
		virtualService := &appmesh.GatewayRouteVirtualService{}

		mVirtualService := vVirtualService[0].(map[string]interface{})

		if vVirtualServiceName, ok := mVirtualService["virtual_service_name"].(string); ok && vVirtualServiceName != "" {
			virtualService.VirtualServiceName = aws.String(vVirtualServiceName)
		}

		routeTarget.VirtualService = virtualService
	}

	return routeTarget
}

func expandGRPCGatewayRoute(vGrpcRoute []interface{}) *appmesh.GrpcGatewayRoute {
	if len(vGrpcRoute) == 0 || vGrpcRoute[0] == nil {
		return nil
	}

	route := &appmesh.GrpcGatewayRoute{}

	mGrpcRoute := vGrpcRoute[0].(map[string]interface{})

	if vRouteAction, ok := mGrpcRoute["action"].([]interface{}); ok && len(vRouteAction) > 0 && vRouteAction[0] != nil {
		routeAction := &appmesh.GrpcGatewayRouteAction{}

		mRouteAction := vRouteAction[0].(map[string]interface{})

		if vRouteTarget, ok := mRouteAction["target"].([]interface{}); ok {
			routeAction.Target = expandGatewayRouteTarget(vRouteTarget)
		}

		route.Action = routeAction
	}

	if vRouteMatch, ok := mGrpcRoute["match"].([]interface{}); ok && len(vRouteMatch) > 0 && vRouteMatch[0] != nil {
		routeMatch := &appmesh.GrpcGatewayRouteMatch{}

		mRouteMatch := vRouteMatch[0].(map[string]interface{})

		if vServiceName, ok := mRouteMatch["service_name"].(string); ok && vServiceName != "" {
			routeMatch.ServiceName = aws.String(vServiceName)
		}

		route.Match = routeMatch
	}

	return route
}

func expandHTTPGatewayRouteRewrite(vHttpRouteRewrite []interface{}) *appmesh.HttpGatewayRouteRewrite {
	if len(vHttpRouteRewrite) == 0 || vHttpRouteRewrite[0] == nil {
		return nil
	}
	mRouteRewrite := vHttpRouteRewrite[0].(map[string]interface{})
	routeRewrite := &appmesh.HttpGatewayRouteRewrite{}

	if vRouteHostnameRewrite, ok := mRouteRewrite["hostname"].([]interface{}); ok && len(vRouteHostnameRewrite) > 0 && vRouteHostnameRewrite[0] != nil {
		mRouteHostnameRewrite := vRouteHostnameRewrite[0].(map[string]interface{})
		routeHostnameRewrite := &appmesh.GatewayRouteHostnameRewrite{}
		if vDefaultTargetHostname, ok := mRouteHostnameRewrite["default_target_hostname"].(string); ok && vDefaultTargetHostname != "" {
			routeHostnameRewrite.DefaultTargetHostname = aws.String(vDefaultTargetHostname)
		}
		routeRewrite.Hostname = routeHostnameRewrite
	}

	if vRoutePrefixRewrite, ok := mRouteRewrite["prefix"].([]interface{}); ok && len(vRoutePrefixRewrite) > 0 && vRoutePrefixRewrite[0] != nil {
		mRoutePrefixRewrite := vRoutePrefixRewrite[0].(map[string]interface{})
		routePrefixRewrite := &appmesh.HttpGatewayRoutePrefixRewrite{}
		if vDefaultPrefix, ok := mRoutePrefixRewrite["default_prefix"].(string); ok && vDefaultPrefix != "" {
			routePrefixRewrite.DefaultPrefix = aws.String(vDefaultPrefix)
		}
		if vValue, ok := mRoutePrefixRewrite["value"].(string); ok && vValue != "" {
			routePrefixRewrite.Value = aws.String(vValue)
		}
		routeRewrite.Prefix = routePrefixRewrite
	}

	return routeRewrite
}

func expandHTTPGatewayRouteMatch(vHttpRouteMatch []interface{}) *appmesh.HttpGatewayRouteMatch {
	if len(vHttpRouteMatch) == 0 || vHttpRouteMatch[0] == nil {
		return nil
	}

	routeMatch := &appmesh.HttpGatewayRouteMatch{}

	mRouteMatch := vHttpRouteMatch[0].(map[string]interface{})

	if vPrefix, ok := mRouteMatch["prefix"].(string); ok && vPrefix != "" {
		routeMatch.Prefix = aws.String(vPrefix)
	}

	if vHostnameMatch, ok := mRouteMatch["hostname"].([]interface{}); ok && len(vHostnameMatch) > 0 && vHostnameMatch[0] != nil {
		hostnameMatch := &appmesh.GatewayRouteHostnameMatch{}

		mHostnameMatch := vHostnameMatch[0].(map[string]interface{})
		if vExact, ok := mHostnameMatch["exact"].(string); ok && vExact != "" {
			hostnameMatch.Exact = aws.String(vExact)
		}
		if vSuffix, ok := mHostnameMatch["suffix"].(string); ok && vSuffix != "" {
			hostnameMatch.Suffix = aws.String(vSuffix)
		}
		routeMatch.Hostname = hostnameMatch
	}

	return routeMatch
}

func expandHTTPGatewayRoute(vHttpRoute []interface{}) *appmesh.HttpGatewayRoute {
	if len(vHttpRoute) == 0 || vHttpRoute[0] == nil {
		return nil
	}

	route := &appmesh.HttpGatewayRoute{}

	mHttpRoute := vHttpRoute[0].(map[string]interface{})

	if vRouteAction, ok := mHttpRoute["action"].([]interface{}); ok && len(vRouteAction) > 0 && vRouteAction[0] != nil {
		routeAction := &appmesh.HttpGatewayRouteAction{}

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

func flattenGatewayRouteSpec(spec *appmesh.GatewayRouteSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{
		"grpc_route":  flattenGRPCGatewayRoute(spec.GrpcRoute),
		"http2_route": flattenHTTPGatewayRoute(spec.Http2Route),
		"http_route":  flattenHTTPGatewayRoute(spec.HttpRoute),
	}

	return []interface{}{mSpec}
}

func flattenGatewayRouteTarget(routeTarget *appmesh.GatewayRouteTarget) []interface{} {
	if routeTarget == nil {
		return []interface{}{}
	}

	mRouteTarget := map[string]interface{}{}

	if virtualService := routeTarget.VirtualService; virtualService != nil {
		mVirtualService := map[string]interface{}{
			"virtual_service_name": aws.StringValue(virtualService.VirtualServiceName),
		}

		mRouteTarget["virtual_service"] = []interface{}{mVirtualService}
	}

	return []interface{}{mRouteTarget}
}

func flattenGRPCGatewayRoute(grpcRoute *appmesh.GrpcGatewayRoute) []interface{} {
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
			"service_name": aws.StringValue(routeMatch.ServiceName),
		}

		mGrpcRoute["match"] = []interface{}{mRouteMatch}
	}

	return []interface{}{mGrpcRoute}
}

func flattenHTTPGatewayRouteMatch(routeMatch *appmesh.HttpGatewayRouteMatch) []interface{} {
	if routeMatch == nil {
		return []interface{}{}
	}

	mRouteMatch := map[string]interface{}{}

	if routeMatch.Prefix != nil {
		mRouteMatch["prefix"] = aws.StringValue(routeMatch.Prefix)
	}

	if hostnameMatch := routeMatch.Hostname; hostnameMatch != nil {
		mHostnameMatch := map[string]interface{}{}
		if hostnameMatch.Exact != nil {
			mHostnameMatch["exact"] = aws.StringValue(hostnameMatch.Exact)
		}
		if hostnameMatch.Suffix != nil {
			mHostnameMatch["suffix"] = aws.StringValue(hostnameMatch.Suffix)
		}

		mRouteMatch["hostname"] = []interface{}{mHostnameMatch}
	}
	return []interface{}{mRouteMatch}
}

func flattenHTTPGatewayRouteRewrite(routeRewrite *appmesh.HttpGatewayRouteRewrite) []interface{} {
	if routeRewrite == nil {
		return []interface{}{}
	}

	mRouteRewrite := map[string]interface{}{}

	if rewriteHostname := routeRewrite.Hostname; rewriteHostname != nil {
		mRewriteHostname := map[string]interface{}{
			"default_target_hostname": aws.StringValue(rewriteHostname.DefaultTargetHostname),
		}
		mRouteRewrite["hostname"] = []interface{}{mRewriteHostname}
	}

	if rewritePrefix := routeRewrite.Prefix; rewritePrefix != nil {
		mRewritePrefix := map[string]interface{}{
			"default_prefix": aws.StringValue(rewritePrefix.DefaultPrefix),
		}
		if rewritePrefixValue := rewritePrefix.Value; rewritePrefixValue != nil {
			mRewritePrefix["value"] = aws.StringValue(rewritePrefix.Value)
		}
		mRouteRewrite["prefix"] = []interface{}{mRewritePrefix}
	}

	return []interface{}{mRouteRewrite}
}

func flattenHTTPGatewayRoute(httpRoute *appmesh.HttpGatewayRoute) []interface{} {
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
