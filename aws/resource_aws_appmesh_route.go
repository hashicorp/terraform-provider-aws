package aws

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmeshpreview"
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
												"prefix": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringMatch(regexp.MustCompile(`^/`), "must start with /"),
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

	req := &appmeshpreview.CreateRouteInput{
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

	resp, err := conn.DescribeRoute(&appmeshpreview.DescribeRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	})
	if isAWSErr(err, appmeshpreview.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh route (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading App Mesh route: %s", err)
	}
	if aws.StringValue(resp.Route.Status.Status) == appmeshpreview.RouteStatusCodeDeleted {
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
	// if isAWSErr(err, appmeshpreview.ErrCodeNotFoundException, "") {
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
		req := &appmeshpreview.UpdateRouteInput{
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
	// if isAWSErr(err, appmeshpreview.ErrCodeNotFoundException, "") {
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
	_, err := conn.DeleteRoute(&appmeshpreview.DeleteRouteInput{
		MeshName:          aws.String(d.Get("mesh_name").(string)),
		RouteName:         aws.String(d.Get("name").(string)),
		VirtualRouterName: aws.String(d.Get("virtual_router_name").(string)),
	})
	if isAWSErr(err, appmeshpreview.ErrCodeNotFoundException, "") {
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

	resp, err := conn.DescribeRoute(&appmeshpreview.DescribeRouteInput{
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

func expandAppmeshRouteSpec(vSpec []interface{}) *appmeshpreview.RouteSpec {
	spec := &appmeshpreview.RouteSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vHttpRoute, ok := mSpec["http_route"].([]interface{}); ok && len(vHttpRoute) > 0 && vHttpRoute[0] != nil {
		mHttpRoute := vHttpRoute[0].(map[string]interface{})

		spec.HttpRoute = &appmeshpreview.HttpRoute{}

		if vHttpRouteAction, ok := mHttpRoute["action"].([]interface{}); ok && len(vHttpRouteAction) > 0 && vHttpRouteAction[0] != nil {
			mHttpRouteAction := vHttpRouteAction[0].(map[string]interface{})

			if vWeightedTargets, ok := mHttpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
				weightedTargets := []*appmeshpreview.WeightedTarget{}

				for _, vWeightedTarget := range vWeightedTargets.List() {
					weightedTarget := &appmeshpreview.WeightedTarget{}

					mWeightedTarget := vWeightedTarget.(map[string]interface{})

					if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
						weightedTarget.VirtualNode = aws.String(vVirtualNode)
					}
					if vWeight, ok := mWeightedTarget["weight"].(int); ok {
						weightedTarget.Weight = aws.Int64(int64(vWeight))
					}

					weightedTargets = append(weightedTargets, weightedTarget)
				}

				spec.HttpRoute.Action = &appmeshpreview.HttpRouteAction{
					WeightedTargets: weightedTargets,
				}
			}
		}

		if vHttpRouteMatch, ok := mHttpRoute["match"].([]interface{}); ok && len(vHttpRouteMatch) > 0 && vHttpRouteMatch[0] != nil {
			mHttpRouteMatch := vHttpRouteMatch[0].(map[string]interface{})

			if vPrefix, ok := mHttpRouteMatch["prefix"].(string); ok && vPrefix != "" {
				spec.HttpRoute.Match = &appmeshpreview.HttpRouteMatch{
					Prefix: aws.String(vPrefix),
				}
			}
		}
	}

	if vTcpRoute, ok := mSpec["tcp_route"].([]interface{}); ok && len(vTcpRoute) > 0 && vTcpRoute[0] != nil {
		mTcpRoute := vTcpRoute[0].(map[string]interface{})

		spec.TcpRoute = &appmeshpreview.TcpRoute{}

		if vTcpRouteAction, ok := mTcpRoute["action"].([]interface{}); ok && len(vTcpRouteAction) > 0 && vTcpRouteAction[0] != nil {
			mTcpRouteAction := vTcpRouteAction[0].(map[string]interface{})

			if vWeightedTargets, ok := mTcpRouteAction["weighted_target"].(*schema.Set); ok && vWeightedTargets.Len() > 0 {
				weightedTargets := []*appmeshpreview.WeightedTarget{}

				for _, vWeightedTarget := range vWeightedTargets.List() {
					weightedTarget := &appmeshpreview.WeightedTarget{}

					mWeightedTarget := vWeightedTarget.(map[string]interface{})

					if vVirtualNode, ok := mWeightedTarget["virtual_node"].(string); ok && vVirtualNode != "" {
						weightedTarget.VirtualNode = aws.String(vVirtualNode)
					}
					if vWeight, ok := mWeightedTarget["weight"].(int); ok {
						weightedTarget.Weight = aws.Int64(int64(vWeight))
					}

					weightedTargets = append(weightedTargets, weightedTarget)
				}

				spec.TcpRoute.Action = &appmeshpreview.TcpRouteAction{
					WeightedTargets: weightedTargets,
				}
			}
		}
	}

	return spec
}

func flattenAppmeshRouteSpec(spec *appmeshpreview.RouteSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if spec.HttpRoute != nil {
		mHttpRoute := map[string]interface{}{}

		if spec.HttpRoute.Action != nil && spec.HttpRoute.Action.WeightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range spec.HttpRoute.Action.WeightedTargets {
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

		if spec.HttpRoute.Match != nil {
			mHttpRoute["match"] = []interface{}{
				map[string]interface{}{
					"prefix": aws.StringValue(spec.HttpRoute.Match.Prefix),
				},
			}
		}

		mSpec["http_route"] = []interface{}{mHttpRoute}
	}

	if spec.TcpRoute != nil {
		mTcpRoute := map[string]interface{}{}

		if spec.TcpRoute.Action != nil && spec.TcpRoute.Action.WeightedTargets != nil {
			vWeightedTargets := []interface{}{}

			for _, weightedTarget := range spec.TcpRoute.Action.WeightedTargets {
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

		mSpec["tcp_route"] = []interface{}{mTcpRoute}
	}

	return []interface{}{mSpec}
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
