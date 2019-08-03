package aws

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	appmesh "github.com/aws/aws-sdk-go/service/appmeshpreview"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsAppmeshRoute() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppmeshRouteCreate,
		Read:   resourceAwsAppmeshRouteRead,
		Update: resourceAwsAppmeshRouteUpdate,
		Delete: resourceAwsAppmeshRouteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppmeshRouteImport,
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
						"http_route": {
							Type:          schema.TypeList,
							Optional:      true,
							MinItems:      0,
							MaxItems:      1,
							ConflictsWith: []string{"spec.0.tcp_route"},
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
														},
													},
													Set: appmeshWeightedTargetHash,
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
																			Type:     schema.TypeString,
																			Optional: true,
																		},

																		"prefix": {
																			Type:     schema.TypeString,
																			Optional: true,
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
																			Type:     schema.TypeString,
																			Optional: true,
																		},

																		"suffix": {
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
													Set: appmeshHttpRouteHeaderHash,
												},

												"method": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														appmesh.HttpMethodConnect,
														appmesh.HttpMethodDelete,
														appmesh.HttpMethodGet,
														appmesh.HttpMethodHead,
														appmesh.HttpMethodOptions,
														appmesh.HttpMethodPatch,
														appmesh.HttpMethodPost,
														appmesh.HttpMethodPut,
														appmesh.HttpMethodTrace,
													}, false),
												},

												"prefix": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
												},

												"scheme": {
													Type:     schema.TypeString,
													Optional: true,
													ValidateFunc: validation.StringInSlice([]string{
														appmesh.HttpSchemeHttp,
														appmesh.HttpSchemeHttps,
													}, false),
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
													Optional: true,
													Default:  1,
												},

												// TODO The API default is 15000ms, but that cannot currently be expressed via 'Default:'.
												// TODO The API always returns results as ms. Should the attribute be 'per_retry_timeout_millis'?
												// TODO See https://github.com/aws/aws-app-mesh-roadmap/issues/7#issuecomment-518041427.
												"per_retry_timeout": {
													Type:     schema.TypeList,
													Optional: true,
													MinItems: 0,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"unit": {
																Type:     schema.TypeString,
																Required: true,
																ValidateFunc: validation.StringInSlice([]string{
																	appmesh.DurationUnitMs,
																	appmesh.DurationUnitS,
																}, false),
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
								},
							},
						},

						"tcp_route": {
							Type:          schema.TypeList,
							Optional:      true,
							MinItems:      0,
							MaxItems:      1,
							ConflictsWith: []string{"spec.0.http_route"},
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
														},
													},
													Set: appmeshWeightedTargetHash,
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

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAppmeshRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshpreviewconn

	req := &appmesh.CreateRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
		Spec:              expandAppmeshRouteSpec(d.Get("spec").([]interface{})),
		// Tags:              tagsFromMapAppmesh(d.Get("tags").(map[string]interface{})),
	}

	log.Printf("[DEBUG] Creating App Mesh route: %#v", req)
	resp, err := conn.CreateRoute(req)
	if err != nil {
		return fmt.Errorf("error creating App Mesh route: %s", err)
	}

	d.SetId(aws.StringValue(resp.Route.Metadata.Uid))

	return resourceAwsAppmeshRouteRead(d, meta)
}

func resourceAwsAppmeshRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshpreviewconn

	resp, err := conn.DescribeRoute(&appmesh.DescribeRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading App Mesh route: %s", err)
	}
	if aws.StringValue(resp.Route.Status.Status) == appmesh.RouteStatusCodeDeleted {
		log.Printf("[WARN] App Mesh route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("name", resp.Route.RouteName)
	d.Set("mesh_name", resp.Route.MeshName)
	d.Set("virtual_router_name", resp.Route.VirtualRouterName)
	d.Set("arn", resp.Route.Metadata.Arn)
	d.Set("created_date", resp.Route.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.Route.Metadata.LastUpdatedAt.Format(time.RFC3339))
	err = d.Set("spec", flattenAppmeshRouteSpec(resp.Route.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %s", err)
	}

	// err = saveTagsAppmesh(conn, d, aws.StringValue(resp.Route.Metadata.Arn))
	// if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
	// 	log.Printf("[WARN] App Mesh route (%s) not found, removing from state", d.Id())
	// 	d.SetId("")
	// 	return nil
	// }
	// if err != nil {
	// 	return fmt.Errorf("error saving tags: %s", err)
	// }

	return nil
}

func resourceAwsAppmeshRouteUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshpreviewconn

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateRouteInput{
			MeshName:          aws.String(d.Get("mesh_name").(string)),
			RouteName:         aws.String(d.Get("name").(string)),
			VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
			Spec:              expandAppmeshRouteSpec(v.([]interface{})),
		}

		log.Printf("[DEBUG] Updating App Mesh route: %#v", req)
		_, err := conn.UpdateRoute(req)
		if err != nil {
			return fmt.Errorf("error updating App Mesh route: %s", err)
		}
	}

	// err := setTagsAppmesh(conn, d, d.Get("arn").(string))
	// if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
	// 	log.Printf("[WARN] App Mesh route (%s) not found, removing from state", d.Id())
	// 	d.SetId("")
	// 	return nil
	// }
	// if err != nil {
	// 	return fmt.Errorf("error setting tags: %s", err)
	// }

	return resourceAwsAppmeshRouteRead(d, meta)
}

func resourceAwsAppmeshRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshpreviewconn

	log.Printf("[DEBUG] Deleting App Mesh route: %s", d.Id())
	_, err := conn.DeleteRoute(&appmesh.DeleteRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting App Mesh route: %s", err)
	}

	return nil
}

func resourceAwsAppmeshRouteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 3 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'mesh-name/virtual-router-name/route-name'", d.Id())
	}

	mesh := parts[0]
	vrName := parts[1]
	name := parts[2]
	log.Printf("[DEBUG] Importing App Mesh route %s from mesh %s/virtual router %s ", name, mesh, vrName)

	conn := meta.(*AWSClient).appmeshpreviewconn

	resp, err := conn.DescribeRoute(&appmesh.DescribeRouteInput{
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

func expandAppmeshRouteSpec(vSpec []interface{}) *appmesh.RouteSpec {
	spec := &appmesh.RouteSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vHttpRoute, ok := mSpec["http_route"].([]interface{}); ok && len(vHttpRoute) > 0 && vHttpRoute[0] != nil {
		mHttpRoute := vHttpRoute[0].(map[string]interface{})

		spec.HttpRoute = &appmesh.HttpRoute{}

		if vHttpRouteAction, ok := mHttpRoute["action"].([]interface{}); ok && len(vHttpRouteAction) > 0 && vHttpRouteAction[0] != nil {
			mHttpRouteAction := vHttpRouteAction[0].(map[string]interface{})

			if vWeightedTargets, ok := mHttpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
				weightedTargets := []*appmesh.WeightedTarget{}

				for _, vWeightedTarget := range vWeightedTargets.List() {
					weightedTarget := &appmesh.WeightedTarget{}

					mWeightedTarget := vWeightedTarget.(map[string]interface{})

					if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
						weightedTarget.VirtualNode = aws.String(vVirtualNode)
					}
					if vWeight, ok := mWeightedTarget["weight"].(int); ok && vWeight > 0 {
						weightedTarget.Weight = aws.Int64(int64(vWeight))
					}

					weightedTargets = append(weightedTargets, weightedTarget)
				}

				spec.HttpRoute.Action = &appmesh.HttpRouteAction{
					WeightedTargets: weightedTargets,
				}
			}
		}

		if vHttpRouteMatch, ok := mHttpRoute["match"].([]interface{}); ok && len(vHttpRouteMatch) > 0 && vHttpRouteMatch[0] != nil {
			httpRouteMatch := &appmesh.HttpRouteMatch{}

			mHttpRouteMatch := vHttpRouteMatch[0].(map[string]interface{})

			if vMethod, ok := mHttpRouteMatch["method"].(string); ok && vMethod != "" {
				httpRouteMatch.Method = aws.String(vMethod)
			}
			if vPrefix, ok := mHttpRouteMatch["prefix"].(string); ok && vPrefix != "" {
				httpRouteMatch.Prefix = aws.String(vPrefix)
			}
			if vScheme, ok := mHttpRouteMatch["scheme"].(string); ok && vScheme != "" {
				httpRouteMatch.Scheme = aws.String(vScheme)
			}

			if vHttpRouteHeaders, ok := mHttpRouteMatch["header"].(*schema.Set); ok && vHttpRouteHeaders.Len() > 0 {
				httpRouteHeaders := []*appmesh.HttpRouteHeader{}

				for _, vHttpRouteHeader := range vHttpRouteHeaders.List() {
					httpRouteHeader := &appmesh.HttpRouteHeader{}

					mHttpRouteHeader := vHttpRouteHeader.(map[string]interface{})

					if vInvert, ok := mHttpRouteHeader["invert"].(bool); ok {
						httpRouteHeader.Invert = aws.Bool(vInvert)
					}
					if vName, ok := mHttpRouteHeader["name"].(string); ok && vName != "" {
						httpRouteHeader.Name = aws.String(vName)
					}

					if vMatch, ok := mHttpRouteHeader["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
						httpRouteHeader.Match = &appmesh.HeaderMatchMethod{}

						mMatch := vMatch[0].(map[string]interface{})

						if vExact, ok := mMatch["exact"].(string); ok && vExact != "" {
							httpRouteHeader.Match.Exact = aws.String(vExact)
						}
						if vPrefix, ok := mMatch["prefix"].(string); ok && vPrefix != "" {
							httpRouteHeader.Match.Prefix = aws.String(vPrefix)
						}
						if vRegex, ok := mMatch["regex"].(string); ok && vRegex != "" {
							httpRouteHeader.Match.Regex = aws.String(vRegex)
						}
						if vSuffix, ok := mMatch["suffix"].(string); ok && vSuffix != "" {
							httpRouteHeader.Match.Suffix = aws.String(vSuffix)
						}

						if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
							httpRouteHeader.Match.Range = &appmesh.MatchRange{}

							mRange := vRange[0].(map[string]interface{})

							if vEnd, ok := mRange["end"].(int); ok && vEnd > 0 {
								httpRouteHeader.Match.Range.End = aws.Int64(int64(vEnd))
							}
							if vStart, ok := mRange["start"].(int); ok && vStart > 0 {
								httpRouteHeader.Match.Range.Start = aws.Int64(int64(vStart))
							}
						}
					}

					httpRouteHeaders = append(httpRouteHeaders, httpRouteHeader)
				}

				httpRouteMatch.Headers = httpRouteHeaders
			}

			spec.HttpRoute.Match = httpRouteMatch
		}

		if vHttpRetryPolicy, ok := mHttpRoute["retry_policy"].([]interface{}); ok && len(vHttpRetryPolicy) > 0 && vHttpRetryPolicy[0] != nil {
			httpRetryPolicy := &appmesh.HttpRetryPolicy{}

			mHttpRetryPolicy := vHttpRetryPolicy[0].(map[string]interface{})

			if vMaxRetries, ok := mHttpRetryPolicy["max_retries"].(int); ok && vMaxRetries > 0 {
				httpRetryPolicy.MaxRetries = aws.Int64(int64(vMaxRetries))
			}

			if vHttpRetryEvents, ok := mHttpRetryPolicy["http_retry_events"].(*schema.Set); ok && vHttpRetryEvents.Len() > 0 {
				httpRetryPolicy.HttpRetryEvents = expandStringSet(vHttpRetryEvents)
			}

			if vPerRetryTimeout, ok := mHttpRetryPolicy["per_retry_timeout"].([]interface{}); ok && len(vPerRetryTimeout) > 0 && vPerRetryTimeout[0] != nil {
				perRetryTimeout := &appmesh.Duration{}

				mPerRetryTimeout := vPerRetryTimeout[0].(map[string]interface{})

				if vUnit, ok := mPerRetryTimeout["unit"].(string); ok && vUnit != "" {
					perRetryTimeout.Unit = aws.String(vUnit)
				}
				if vValue, ok := mPerRetryTimeout["value"].(int); ok && vValue > 0 {
					perRetryTimeout.Value = aws.Int64(int64(vValue))
				}

				httpRetryPolicy.PerRetryTimeout = perRetryTimeout
			}

			if vTcpRetryEvents, ok := mHttpRetryPolicy["tcp_retry_events"].(*schema.Set); ok && vTcpRetryEvents.Len() > 0 {
				httpRetryPolicy.TcpRetryEvents = expandStringSet(vTcpRetryEvents)
			}

			spec.HttpRoute.RetryPolicy = httpRetryPolicy
		}
	}

	if vTcpRoute, ok := mSpec["tcp_route"].([]interface{}); ok && len(vTcpRoute) > 0 && vTcpRoute[0] != nil {
		mTcpRoute := vTcpRoute[0].(map[string]interface{})

		spec.TcpRoute = &appmesh.TcpRoute{}

		if vTcpRouteAction, ok := mTcpRoute["action"].([]interface{}); ok && len(vTcpRouteAction) > 0 && vTcpRouteAction[0] != nil {
			mTcpRouteAction := vTcpRouteAction[0].(map[string]interface{})

			if vWeightedTargets, ok := mTcpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
				weightedTargets := []*appmesh.WeightedTarget{}

				for _, vWeightedTarget := range vWeightedTargets.List() {
					weightedTarget := &appmesh.WeightedTarget{}

					mWeightedTarget := vWeightedTarget.(map[string]interface{})

					if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
						weightedTarget.VirtualNode = aws.String(vVirtualNode)
					}
					if vWeight, ok := mWeightedTarget["weight"].(int); ok && vWeight > 0 {
						weightedTarget.Weight = aws.Int64(int64(vWeight))
					}

					weightedTargets = append(weightedTargets, weightedTarget)
				}

				spec.TcpRoute.Action = &appmesh.TcpRouteAction{
					WeightedTargets: weightedTargets,
				}
			}
		}
	}

	return spec
}

func flattenAppmeshRouteSpec(spec *appmesh.RouteSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if httpRoute := spec.HttpRoute; httpRoute != nil {
		mHttpRoute := map[string]interface{}{}

		if action := httpRoute.Action; action != nil {
			if weightedTargets := action.WeightedTargets; weightedTargets != nil {
				vWeightedTargets := []interface{}{}

				for _, weightedTarget := range weightedTargets {
					mWeightedTarget := map[string]interface{}{
						"virtual_node": aws.StringValue(weightedTarget.VirtualNode),
						"weight":       int(aws.Int64Value(weightedTarget.Weight)),
					}

					vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
				}

				mHttpRoute["action"] = []interface{}{
					map[string]interface{}{
						"weighted_target": schema.NewSet(appmeshWeightedTargetHash, vWeightedTargets),
					},
				}
			}
		}

		if httpRouteMatch := httpRoute.Match; httpRouteMatch != nil {
			vHttpRouteHeaders := []interface{}{}

			for _, httpRouteHeader := range httpRouteMatch.Headers {
				mHttpRouteHeader := map[string]interface{}{
					"invert": aws.BoolValue(httpRouteHeader.Invert),
					"name":   aws.StringValue(httpRouteHeader.Name),
				}

				if match := httpRouteHeader.Match; match != nil {
					mMatch := map[string]interface{}{
						"exact":  aws.StringValue(match.Exact),
						"prefix": aws.StringValue(match.Prefix),
						"regex":  aws.StringValue(match.Regex),
						"suffix": aws.StringValue(match.Suffix),
					}

					if r := match.Range; r != nil {
						mRange := map[string]interface{}{
							"end":   int(aws.Int64Value(r.End)),
							"start": int(aws.Int64Value(r.Start)),
						}

						mMatch["range"] = []interface{}{mRange}
					}

					mHttpRouteHeader["match"] = []interface{}{mMatch}
				}

				vHttpRouteHeaders = append(vHttpRouteHeaders, mHttpRouteHeader)
			}

			mHttpRoute["match"] = []interface{}{
				map[string]interface{}{
					"header": schema.NewSet(appmeshHttpRouteHeaderHash, vHttpRouteHeaders),
					"method": aws.StringValue(httpRouteMatch.Method),
					"prefix": aws.StringValue(httpRouteMatch.Prefix),
					"scheme": aws.StringValue(httpRouteMatch.Scheme),
				},
			}
		}

		if httpRetryPolicy := httpRoute.RetryPolicy; httpRetryPolicy != nil {
			mHttpRetryPolicy := map[string]interface{}{
				"http_retry_events": flattenStringSet(httpRetryPolicy.HttpRetryEvents),
				"max_retries":       int(aws.Int64Value(httpRetryPolicy.MaxRetries)),
				"tcp_retry_events":  flattenStringSet(httpRetryPolicy.TcpRetryEvents),
			}

			if perRetryTimeout := httpRetryPolicy.PerRetryTimeout; perRetryTimeout != nil {
				mPerRetryTimeout := map[string]interface{}{
					"unit":  aws.StringValue(perRetryTimeout.Unit),
					"value": int(aws.Int64Value(perRetryTimeout.Value)),
				}

				mHttpRetryPolicy["per_retry_timeout"] = []interface{}{mPerRetryTimeout}
			}

			mHttpRoute["retry_policy"] = []interface{}{mHttpRetryPolicy}
		}

		mSpec["http_route"] = []interface{}{mHttpRoute}
	}

	if tcpRoute := spec.TcpRoute; tcpRoute != nil {
		mTcpRoute := map[string]interface{}{}

		if action := tcpRoute.Action; action != nil {
			if weightedTargets := action.WeightedTargets; weightedTargets != nil {
				vWeightedTargets := []interface{}{}

				for _, weightedTarget := range weightedTargets {
					mWeightedTarget := map[string]interface{}{
						"virtual_node": aws.StringValue(weightedTarget.VirtualNode),
						"weight":       int(aws.Int64Value(weightedTarget.Weight)),
					}

					vWeightedTargets = append(vWeightedTargets, mWeightedTarget)
				}

				mTcpRoute["action"] = []interface{}{
					map[string]interface{}{
						"weighted_target": schema.NewSet(appmeshWeightedTargetHash, vWeightedTargets),
					},
				}
			}
		}

		mSpec["tcp_route"] = []interface{}{mTcpRoute}
	}

	return []interface{}{mSpec}
}

func appmeshHttpRouteHeaderHash(vHttpRouteHeader interface{}) int {
	var buf bytes.Buffer
	mHttpRouteHeader := vHttpRouteHeader.(map[string]interface{})
	if v, ok := mHttpRouteHeader["invert"].(bool); ok {
		buf.WriteString(fmt.Sprintf("%t-", v))
	}
	if vMatch, ok := mHttpRouteHeader["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
		mMatch := vMatch[0].(map[string]interface{})
		if v, ok := mMatch["exact"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
		if v, ok := mMatch["prefix"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
		if vRange, ok := mMatch["range"].([]interface{}); ok && len(vRange) > 0 && vRange[0] != nil {
			mRange := vRange[0].(map[string]interface{})
			if v, ok := mRange["end"].(int); ok {
				buf.WriteString(fmt.Sprintf("%d-", v))
			}
			if v, ok := mRange["start"].(int); ok {
				buf.WriteString(fmt.Sprintf("%d-", v))
			}
		}
		if v, ok := mMatch["regex"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
		if v, ok := mMatch["suffix"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if v, ok := mHttpRouteHeader["name"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	return hashcode.String(buf.String())
}

func appmeshWeightedTargetHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	if v, ok := m["virtual_node"].(string); ok {
		buf.WriteString(fmt.Sprintf("%s-", v))
	}
	if v, ok := m["weight"].(int); ok {
		buf.WriteString(fmt.Sprintf("%d-", v))
	}
	return hashcode.String(buf.String())
}
