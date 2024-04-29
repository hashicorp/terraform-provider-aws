// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appmesh_virtual_gateway", name="Virtual Gateway")
// @Tags(identifierAttribute="arn")
func resourceVirtualGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVirtualGatewayCreate,
		ReadWithoutTimeout:   resourceVirtualGatewayRead,
		UpdateWithoutTimeout: resourceVirtualGatewayUpdate,
		DeleteWithoutTimeout: resourceVirtualGatewayDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceVirtualGatewayImport,
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
				"spec":            resourceVirtualGatewaySpecSchema(),
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVirtualGatewaySpecSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"backend_defaults": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"client_policy": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"tls": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"certificate": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"file": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 0,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"certificate_chain": {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: validation.StringLenBetween(1, 255),
																			},
																			"private_key": {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: validation.StringLenBetween(1, 255),
																			},
																		},
																	},
																	ExactlyOneOf: []string{
																		"spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file",
																		"spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.sds",
																	},
																},
																"sds": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 0,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"secret_name": {
																				Type:     schema.TypeString,
																				Required: true,
																			},
																		},
																	},
																	ExactlyOneOf: []string{
																		"spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.file",
																		"spec.0.backend_defaults.0.client_policy.0.tls.0.certificate.0.sds",
																	},
																},
															},
														},
													},
													"enforce": {
														Type:     schema.TypeBool,
														Optional: true,
														Default:  true,
													},
													"ports": {
														Type:     schema.TypeSet,
														Optional: true,
														Elem: &schema.Schema{
															Type:         schema.TypeInt,
															ValidateFunc: validation.IsPortNumber,
														},
													},
													"validation": {
														Type:     schema.TypeList,
														Required: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"subject_alternative_names": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 0,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"match": {
																				Type:     schema.TypeList,
																				Required: true,
																				MinItems: 1,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"exact": {
																							Type:     schema.TypeSet,
																							Required: true,
																							Elem:     &schema.Schema{Type: schema.TypeString},
																						},
																					},
																				},
																			},
																		},
																	},
																},
																"trust": {
																	Type:     schema.TypeList,
																	Required: true,
																	MinItems: 1,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"acm": {
																				Type:     schema.TypeList,
																				Optional: true,
																				MinItems: 0,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"certificate_authority_arns": {
																							Type:     schema.TypeSet,
																							Required: true,
																							Elem: &schema.Schema{
																								Type:         schema.TypeString,
																								ValidateFunc: verify.ValidARN,
																							},
																						},
																					},
																				},
																				ExactlyOneOf: []string{
																					"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm",
																					"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file",
																					"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds",
																				},
																			},
																			"file": {
																				Type:     schema.TypeList,
																				Optional: true,
																				MinItems: 0,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"certificate_chain": {
																							Type:         schema.TypeString,
																							Required:     true,
																							ValidateFunc: validation.StringLenBetween(1, 255),
																						},
																					},
																				},
																				ExactlyOneOf: []string{
																					"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm",
																					"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file",
																					"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds",
																				},
																			},
																			"sds": {
																				Type:     schema.TypeList,
																				Optional: true,
																				MinItems: 0,
																				MaxItems: 1,
																				Elem: &schema.Resource{
																					Schema: map[string]*schema.Schema{
																						"secret_name": {
																							Type:         schema.TypeString,
																							Required:     true,
																							ValidateFunc: validation.StringLenBetween(1, 255),
																						},
																					},
																				},
																				ExactlyOneOf: []string{
																					"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.acm",
																					"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.file",
																					"spec.0.backend_defaults.0.client_policy.0.tls.0.validation.0.trust.0.sds",
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
					Required: true,
					MinItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"connection_pool": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"grpc": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"max_requests": {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
												},
											},
										},
										"http": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"max_connections": {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
													"max_pending_requests": {
														Type:         schema.TypeInt,
														Optional:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
												},
											},
										},
										"http2": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"max_requests": {
														Type:         schema.TypeInt,
														Required:     true,
														ValidateFunc: validation.IntAtLeast(1),
													},
												},
											},
										},
									},
								},
							},
							"health_check": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"healthy_threshold": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntBetween(2, 10),
										},
										"interval_millis": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntBetween(5000, 300000),
										},
										"path": {
											Type:     schema.TypeString,
											Optional: true,
										},
										"port": {
											Type:         schema.TypeInt,
											Optional:     true,
											Computed:     true,
											ValidateFunc: validation.IsPortNumber,
										},
										"protocol": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.VirtualGatewayPortProtocol](),
										},
										"timeout_millis": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntBetween(2000, 60000),
										},
										"unhealthy_threshold": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntBetween(2, 10),
										},
									},
								},
							},
							"port_mapping": {
								Type:     schema.TypeList,
								Required: true,
								MinItems: 1,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"port": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IsPortNumber,
										},
										"protocol": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.VirtualGatewayPortProtocol](),
										},
									},
								},
							},
							"tls": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"certificate": {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"acm": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"certificate_arn": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: verify.ValidARN,
																},
															},
														},
													},
													"file": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"certificate_chain": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringLenBetween(1, 255),
																},
																"private_key": {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringLenBetween(1, 255),
																},
															},
														},
													},
													"sds": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"secret_name": {
																	Type:     schema.TypeString,
																	Required: true,
																},
															},
														},
													},
												},
											},
										},
										"mode": {
											Type:             schema.TypeString,
											Required:         true,
											ValidateDiagFunc: enum.Validate[awstypes.VirtualGatewayListenerTlsMode](),
										},
										"validation": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"subject_alternative_names": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"match": {
																	Type:     schema.TypeList,
																	Required: true,
																	MinItems: 1,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"exact": {
																				Type:     schema.TypeSet,
																				Required: true,
																				Elem:     &schema.Schema{Type: schema.TypeString},
																			},
																		},
																	},
																},
															},
														},
													},
													"trust": {
														Type:     schema.TypeList,
														Required: true,
														MinItems: 1,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"file": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 0,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"certificate_chain": {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: validation.StringLenBetween(1, 255),
																			},
																		},
																	},
																},
																"sds": {
																	Type:     schema.TypeList,
																	Optional: true,
																	MinItems: 0,
																	MaxItems: 1,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"secret_name": {
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
									},
								},
							},
						},
					},
				},
				"logging": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"access_log": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"file": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"format": {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																"json": {
																	Type:     schema.TypeList,
																	Optional: true,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			"key": {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: validation.StringLenBetween(1, 100),
																			},
																			"value": {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: validation.StringLenBetween(1, 100),
																			},
																		},
																	},
																},
																"text": {
																	Type:         schema.TypeString,
																	Optional:     true,
																	ValidateFunc: validation.StringLenBetween(1, 1000),
																},
															},
														},
													},
													"path": {
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
			},
		},
	}
}

func resourceVirtualGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	name := d.Get("name").(string)
	input := &appmesh.CreateVirtualGatewayInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		Spec:               expandVirtualGatewaySpec(d.Get("spec").([]interface{})),
		Tags:               getTagsIn(ctx),
		VirtualGatewayName: aws.String(name),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	output, err := conn.CreateVirtualGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Virtual Gateway (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.VirtualGateway.Metadata.Uid))

	return append(diags, resourceVirtualGatewayRead(ctx, d, meta)...)
}

func resourceVirtualGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findVirtualGatewayByThreePartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), d.Get("name").(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Mesh Virtual Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Gateway (%s): %s", d.Id(), err)
	}

	virtualGateway := outputRaw.(*awstypes.VirtualGatewayData)

	arn := aws.ToString(virtualGateway.Metadata.Arn)
	d.Set("arn", arn)
	d.Set("created_date", virtualGateway.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", virtualGateway.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", virtualGateway.MeshName)
	d.Set("mesh_owner", virtualGateway.Metadata.MeshOwner)
	d.Set("name", virtualGateway.VirtualGatewayName)
	d.Set("resource_owner", virtualGateway.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenVirtualGatewaySpec(virtualGateway.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	return diags
}

func resourceVirtualGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	if d.HasChange("spec") {
		input := &appmesh.UpdateVirtualGatewayInput{
			MeshName:           aws.String(d.Get("mesh_name").(string)),
			Spec:               expandVirtualGatewaySpec(d.Get("spec").([]interface{})),
			VirtualGatewayName: aws.String(d.Get("name").(string)),
		}

		if v, ok := d.GetOk("mesh_owner"); ok {
			input.MeshOwner = aws.String(v.(string))
		}

		_, err := conn.UpdateVirtualGateway(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Virtual Gateway (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVirtualGatewayRead(ctx, d, meta)...)
}

func resourceVirtualGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	log.Printf("[DEBUG] Deleting App Mesh Virtual Gateway: %s", d.Id())
	input := &appmesh.DeleteVirtualGatewayInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualGatewayName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	_, err := conn.DeleteVirtualGateway(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh Virtual Gateway (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceVirtualGatewayImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-gateway-name'", d.Id())
	}

	meshName := parts[0]
	name := parts[1]

	conn := meta.(*conns.AWSClient).AppMeshClient(ctx)

	virtualGateway, err := findVirtualGatewayByThreePartKey(ctx, conn, meshName, "", name)

	if err != nil {
		return nil, err
	}

	d.SetId(aws.ToString(virtualGateway.Metadata.Uid))
	d.Set("mesh_name", virtualGateway.MeshName)
	d.Set("name", virtualGateway.VirtualGatewayName)

	return []*schema.ResourceData{d}, nil
}

func findVirtualGatewayByThreePartKey(ctx context.Context, conn *appmesh.Client, meshName, meshOwner, name string) (*awstypes.VirtualGatewayData, error) {
	input := &appmesh.DescribeVirtualGatewayInput{
		MeshName:           aws.String(meshName),
		VirtualGatewayName: aws.String(name),
	}
	if meshOwner != "" {
		input.MeshOwner = aws.String(meshOwner)
	}

	output, err := findVirtualGateway(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if output.Status.Status == awstypes.VirtualGatewayStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(output.Status.Status),
			LastRequest: input,
		}
	}

	return output, nil
}

func findVirtualGateway(ctx context.Context, conn *appmesh.Client, input *appmesh.DescribeVirtualGatewayInput) (*awstypes.VirtualGatewayData, error) {
	output, err := conn.DescribeVirtualGateway(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VirtualGateway == nil || output.VirtualGateway.Metadata == nil || output.VirtualGateway.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VirtualGateway, nil
}

func expandVirtualGatewaySpec(vSpec []interface{}) *awstypes.VirtualGatewaySpec {
	if len(vSpec) == 0 || vSpec[0] == nil {
		return nil
	}

	spec := &awstypes.VirtualGatewaySpec{}

	mSpec := vSpec[0].(map[string]interface{})

	if vBackendDefaults, ok := mSpec["backend_defaults"].([]interface{}); ok && len(vBackendDefaults) > 0 && vBackendDefaults[0] != nil {
		backendDefaults := &awstypes.VirtualGatewayBackendDefaults{}

		mBackendDefaults := vBackendDefaults[0].(map[string]interface{})

		if vClientPolicy, ok := mBackendDefaults["client_policy"].([]interface{}); ok {
			backendDefaults.ClientPolicy = expandVirtualGatewayClientPolicy(vClientPolicy)
		}

		spec.BackendDefaults = backendDefaults
	}

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []awstypes.VirtualGatewayListener{}

		for _, vListener := range vListeners {
			listener := awstypes.VirtualGatewayListener{}

			mListener := vListener.(map[string]interface{})

			if vHealthCheck, ok := mListener["health_check"].([]interface{}); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
				healthCheck := &awstypes.VirtualGatewayHealthCheckPolicy{}

				mHealthCheck := vHealthCheck[0].(map[string]interface{})

				if vHealthyThreshold, ok := mHealthCheck["healthy_threshold"].(int); ok && vHealthyThreshold > 0 {
					healthCheck.HealthyThreshold = aws.Int32(int32(vHealthyThreshold))
				}
				if vIntervalMillis, ok := mHealthCheck["interval_millis"].(int); ok && vIntervalMillis > 0 {
					healthCheck.IntervalMillis = aws.Int64(int64(vIntervalMillis))
				}
				if vPath, ok := mHealthCheck["path"].(string); ok && vPath != "" {
					healthCheck.Path = aws.String(vPath)
				}
				if vPort, ok := mHealthCheck["port"].(int); ok && vPort > 0 {
					healthCheck.Port = aws.Int32(int32(vPort))
				}
				if vProtocol, ok := mHealthCheck["protocol"].(string); ok && vProtocol != "" {
					healthCheck.Protocol = awstypes.VirtualGatewayPortProtocol(vProtocol)
				}
				if vTimeoutMillis, ok := mHealthCheck["timeout_millis"].(int); ok && vTimeoutMillis > 0 {
					healthCheck.TimeoutMillis = aws.Int64(int64(vTimeoutMillis))
				}
				if vUnhealthyThreshold, ok := mHealthCheck["unhealthy_threshold"].(int); ok && vUnhealthyThreshold > 0 {
					healthCheck.UnhealthyThreshold = aws.Int32(int32(vUnhealthyThreshold))
				}

				listener.HealthCheck = healthCheck
			}

			if vConnectionPool, ok := mListener["connection_pool"].([]interface{}); ok && len(vConnectionPool) > 0 && vConnectionPool[0] != nil {
				mConnectionPool := vConnectionPool[0].(map[string]interface{})

				if vGrpcConnectionPool, ok := mConnectionPool["grpc"].([]interface{}); ok && len(vGrpcConnectionPool) > 0 && vGrpcConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualGatewayConnectionPoolMemberGrpc{}
					mGrpcConnectionPool := vGrpcConnectionPool[0].(map[string]interface{})

					grpcConnectionPool := awstypes.VirtualGatewayGrpcConnectionPool{}

					if vMaxRequests, ok := mGrpcConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						grpcConnectionPool.MaxRequests = aws.Int32(int32(vMaxRequests))
					}

					connectionPool.Value = grpcConnectionPool
					listener.ConnectionPool = connectionPool
				}

				if vHttpConnectionPool, ok := mConnectionPool["http"].([]interface{}); ok && len(vHttpConnectionPool) > 0 && vHttpConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualGatewayConnectionPoolMemberHttp{}
					mHttpConnectionPool := vHttpConnectionPool[0].(map[string]interface{})

					httpConnectionPool := awstypes.VirtualGatewayHttpConnectionPool{}

					if vMaxConnections, ok := mHttpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						httpConnectionPool.MaxConnections = aws.Int32(int32(vMaxConnections))
					}
					if vMaxPendingRequests, ok := mHttpConnectionPool["max_pending_requests"].(int); ok && vMaxPendingRequests > 0 {
						httpConnectionPool.MaxPendingRequests = aws.Int32(int32(vMaxPendingRequests))
					}

					connectionPool.Value = httpConnectionPool
					listener.ConnectionPool = connectionPool
				}

				if vHttp2ConnectionPool, ok := mConnectionPool["http2"].([]interface{}); ok && len(vHttp2ConnectionPool) > 0 && vHttp2ConnectionPool[0] != nil {
					connectionPool := &awstypes.VirtualGatewayConnectionPoolMemberHttp2{}
					mHttp2ConnectionPool := vHttp2ConnectionPool[0].(map[string]interface{})

					http2ConnectionPool := awstypes.VirtualGatewayHttp2ConnectionPool{}

					if vMaxRequests, ok := mHttp2ConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						http2ConnectionPool.MaxRequests = aws.Int32(int32(vMaxRequests))
					}

					connectionPool.Value = http2ConnectionPool
					listener.ConnectionPool = connectionPool
				}
			}

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				portMapping := &awstypes.VirtualGatewayPortMapping{}

				mPortMapping := vPortMapping[0].(map[string]interface{})

				if vPort, ok := mPortMapping["port"].(int); ok && vPort > 0 {
					portMapping.Port = aws.Int32(int32(vPort))
				}
				if vProtocol, ok := mPortMapping["protocol"].(string); ok && vProtocol != "" {
					portMapping.Protocol = awstypes.VirtualGatewayPortProtocol(vProtocol)
				}

				listener.PortMapping = portMapping
			}

			if vTls, ok := mListener["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
				tls := &awstypes.VirtualGatewayListenerTls{}

				mTls := vTls[0].(map[string]interface{})

				if vMode, ok := mTls["mode"].(string); ok && vMode != "" {
					tls.Mode = awstypes.VirtualGatewayListenerTlsMode(vMode)
				}

				if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
					mCertificate := vCertificate[0].(map[string]interface{})

					if vAcm, ok := mCertificate["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
						certificate := &awstypes.VirtualGatewayListenerTlsCertificateMemberAcm{}
						acm := awstypes.VirtualGatewayListenerTlsAcmCertificate{}

						mAcm := vAcm[0].(map[string]interface{})

						if vCertificateArn, ok := mAcm["certificate_arn"].(string); ok && vCertificateArn != "" {
							acm.CertificateArn = aws.String(vCertificateArn)
						}

						certificate.Value = acm
						tls.Certificate = certificate
					}

					if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
						certificate := &awstypes.VirtualGatewayListenerTlsCertificateMemberFile{}
						file := awstypes.VirtualGatewayListenerTlsFileCertificate{}

						mFile := vFile[0].(map[string]interface{})

						if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
							file.CertificateChain = aws.String(vCertificateChain)
						}
						if vPrivateKey, ok := mFile["private_key"].(string); ok && vPrivateKey != "" {
							file.PrivateKey = aws.String(vPrivateKey)
						}

						certificate.Value = file
						tls.Certificate = certificate
					}

					if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
						certificate := &awstypes.VirtualGatewayListenerTlsCertificateMemberSds{}
						sds := awstypes.VirtualGatewayListenerTlsSdsCertificate{}

						mSds := vSds[0].(map[string]interface{})

						if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
							sds.SecretName = aws.String(vSecretName)
						}

						certificate.Value = sds
						tls.Certificate = certificate
					}
				}

				if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
					validation := &awstypes.VirtualGatewayListenerTlsValidationContext{}

					mValidation := vValidation[0].(map[string]interface{})

					if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
						subjectAlternativeNames := &awstypes.SubjectAlternativeNames{}

						mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

						if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
							match := &awstypes.SubjectAlternativeNameMatchers{}

							mMatch := vMatch[0].(map[string]interface{})

							if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
								match.Exact = flex.ExpandStringValueSet(vExact)
							}

							subjectAlternativeNames.Match = match
						}

						validation.SubjectAlternativeNames = subjectAlternativeNames
					}

					if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
						mTrust := vTrust[0].(map[string]interface{})

						if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
							trust := &awstypes.VirtualGatewayListenerTlsValidationContextTrustMemberFile{}
							file := awstypes.VirtualGatewayTlsValidationContextFileTrust{}

							mFile := vFile[0].(map[string]interface{})

							if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
								file.CertificateChain = aws.String(vCertificateChain)
							}

							trust.Value = file
							validation.Trust = trust
						}

						if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
							trust := &awstypes.VirtualGatewayListenerTlsValidationContextTrustMemberSds{}
							sds := awstypes.VirtualGatewayTlsValidationContextSdsTrust{}

							mSds := vSds[0].(map[string]interface{})

							if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
								sds.SecretName = aws.String(vSecretName)
							}

							trust.Value = sds
							validation.Trust = trust
						}
					}

					tls.Validation = validation
				}

				listener.Tls = tls
			}

			listeners = append(listeners, listener)
		}

		spec.Listeners = listeners
	}

	if vLogging, ok := mSpec["logging"].([]interface{}); ok && len(vLogging) > 0 && vLogging[0] != nil {
		logging := &awstypes.VirtualGatewayLogging{}

		mLogging := vLogging[0].(map[string]interface{})

		if vAccessLog, ok := mLogging["access_log"].([]interface{}); ok && len(vAccessLog) > 0 && vAccessLog[0] != nil {
			mAccessLog := vAccessLog[0].(map[string]interface{})

			if vFile, ok := mAccessLog["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				accessLog := &awstypes.VirtualGatewayAccessLogMemberFile{}
				file := awstypes.VirtualGatewayFileAccessLog{}

				mFile := vFile[0].(map[string]interface{})

				if vFormat, ok := mFile["format"].([]interface{}); ok && len(vFormat) > 0 && vFormat[0] != nil {
					mFormat := vFormat[0].(map[string]interface{})

					if vJsonFormatRefs, ok := mFormat["json"].([]interface{}); ok && len(vJsonFormatRefs) > 0 {
						format := &awstypes.LoggingFormatMemberJson{}
						jsonFormatRefs := []awstypes.JsonFormatRef{}
						for _, vJsonFormatRef := range vJsonFormatRefs {
							mJsonFormatRef := awstypes.JsonFormatRef{
								Key:   aws.String(vJsonFormatRef.(map[string]interface{})["key"].(string)),
								Value: aws.String(vJsonFormatRef.(map[string]interface{})["value"].(string)),
							}
							jsonFormatRefs = append(jsonFormatRefs, mJsonFormatRef)
						}
						format.Value = jsonFormatRefs
						file.Format = format
					}

					if vText, ok := mFormat["text"].(string); ok && vText != "" {
						format := &awstypes.LoggingFormatMemberText{}
						format.Value = vText
						file.Format = format
					}
				}

				if vPath, ok := mFile["path"].(string); ok && vPath != "" {
					file.Path = aws.String(vPath)
				}

				accessLog.Value = file
				logging.AccessLog = accessLog
			}
		}

		spec.Logging = logging
	}

	return spec
}

func expandVirtualGatewayClientPolicy(vClientPolicy []interface{}) *awstypes.VirtualGatewayClientPolicy {
	if len(vClientPolicy) == 0 || vClientPolicy[0] == nil {
		return nil
	}

	clientPolicy := &awstypes.VirtualGatewayClientPolicy{}

	mClientPolicy := vClientPolicy[0].(map[string]interface{})

	if vTls, ok := mClientPolicy["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
		tls := &awstypes.VirtualGatewayClientPolicyTls{}

		mTls := vTls[0].(map[string]interface{})

		if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
			mCertificate := vCertificate[0].(map[string]interface{})

			if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				certificate := &awstypes.VirtualGatewayClientTlsCertificateMemberFile{}
				file := awstypes.VirtualGatewayListenerTlsFileCertificate{}

				mFile := vFile[0].(map[string]interface{})

				if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
					file.CertificateChain = aws.String(vCertificateChain)
				}
				if vPrivateKey, ok := mFile["private_key"].(string); ok && vPrivateKey != "" {
					file.PrivateKey = aws.String(vPrivateKey)
				}

				certificate.Value = file
				tls.Certificate = certificate
			}

			if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
				certificate := &awstypes.VirtualGatewayClientTlsCertificateMemberSds{}
				sds := awstypes.VirtualGatewayListenerTlsSdsCertificate{}

				mSds := vSds[0].(map[string]interface{})

				if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
					sds.SecretName = aws.String(vSecretName)
				}

				certificate.Value = sds
				tls.Certificate = certificate
			}
		}

		if vEnforce, ok := mTls["enforce"].(bool); ok {
			tls.Enforce = aws.Bool(vEnforce)
		}

		if vPorts, ok := mTls["ports"].(*schema.Set); ok && vPorts.Len() > 0 {
			tls.Ports = flex.ExpandInt32ValueSet(vPorts)
		}

		if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
			validation := &awstypes.VirtualGatewayTlsValidationContext{}

			mValidation := vValidation[0].(map[string]interface{})

			if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
				subjectAlternativeNames := &awstypes.SubjectAlternativeNames{}

				mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

				if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					match := &awstypes.SubjectAlternativeNameMatchers{}

					mMatch := vMatch[0].(map[string]interface{})

					if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
						match.Exact = flex.ExpandStringValueSet(vExact)
					}

					subjectAlternativeNames.Match = match
				}

				validation.SubjectAlternativeNames = subjectAlternativeNames
			}

			if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
				mTrust := vTrust[0].(map[string]interface{})

				if vAcm, ok := mTrust["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
					trust := &awstypes.VirtualGatewayTlsValidationContextTrustMemberAcm{}
					acm := awstypes.VirtualGatewayTlsValidationContextAcmTrust{}

					mAcm := vAcm[0].(map[string]interface{})

					if vCertificateAuthorityArns, ok := mAcm["certificate_authority_arns"].(*schema.Set); ok && vCertificateAuthorityArns.Len() > 0 {
						acm.CertificateAuthorityArns = flex.ExpandStringValueSet(vCertificateAuthorityArns)
					}

					trust.Value = acm
					validation.Trust = trust
				}

				if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
					trust := &awstypes.VirtualGatewayTlsValidationContextTrustMemberFile{}
					file := awstypes.VirtualGatewayTlsValidationContextFileTrust{}

					mFile := vFile[0].(map[string]interface{})

					if vCertificateChain, ok := mFile["certificate_chain"].(string); ok && vCertificateChain != "" {
						file.CertificateChain = aws.String(vCertificateChain)
					}

					trust.Value = file
					validation.Trust = trust
				}

				if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
					trust := &awstypes.VirtualGatewayTlsValidationContextTrustMemberSds{}
					sds := awstypes.VirtualGatewayTlsValidationContextSdsTrust{}

					mSds := vSds[0].(map[string]interface{})

					if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
						sds.SecretName = aws.String(vSecretName)
					}

					trust.Value = sds
					validation.Trust = trust
				}
			}

			tls.Validation = validation
		}

		clientPolicy.Tls = tls
	}

	return clientPolicy
}

func flattenVirtualGatewaySpec(spec *awstypes.VirtualGatewaySpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if backendDefaults := spec.BackendDefaults; backendDefaults != nil {
		mBackendDefaults := map[string]interface{}{
			"client_policy": flattenVirtualGatewayClientPolicy(backendDefaults.ClientPolicy),
		}

		mSpec["backend_defaults"] = []interface{}{mBackendDefaults}
	}

	if spec.Listeners != nil && len(spec.Listeners) > 0 {
		var mListeners []interface{}
		for _, listener := range spec.Listeners {
			mListener := map[string]interface{}{}

			if connectionPool := listener.ConnectionPool; connectionPool != nil {
				mConnectionPool := map[string]interface{}{}

				switch v := connectionPool.(type) {
				case *awstypes.VirtualGatewayConnectionPoolMemberGrpc:
					mGrpcConnectionPool := map[string]interface{}{
						"max_requests": int(aws.ToInt32(v.Value.MaxRequests)),
					}
					mConnectionPool["grpc"] = []interface{}{mGrpcConnectionPool}
				case *awstypes.VirtualGatewayConnectionPoolMemberHttp:
					mHttpConnectionPool := map[string]interface{}{
						"max_connections":      int(aws.ToInt32(v.Value.MaxConnections)),
						"max_pending_requests": int(aws.ToInt32(v.Value.MaxPendingRequests)),
					}
					mConnectionPool["http"] = []interface{}{mHttpConnectionPool}
				case *awstypes.VirtualGatewayConnectionPoolMemberHttp2:
					mHttp2ConnectionPool := map[string]interface{}{
						"max_requests": int(aws.ToInt32(v.Value.MaxRequests)),
					}
					mConnectionPool["http2"] = []interface{}{mHttp2ConnectionPool}
				}

				mListener["connection_pool"] = []interface{}{mConnectionPool}
			}

			if healthCheck := listener.HealthCheck; healthCheck != nil {
				mHealthCheck := map[string]interface{}{
					"healthy_threshold":   int(aws.ToInt32(healthCheck.HealthyThreshold)),
					"interval_millis":     int(aws.ToInt64(healthCheck.IntervalMillis)),
					"path":                aws.ToString(healthCheck.Path),
					"port":                int(aws.ToInt32(healthCheck.Port)),
					"protocol":            string(healthCheck.Protocol),
					"timeout_millis":      int(aws.ToInt64(healthCheck.TimeoutMillis)),
					"unhealthy_threshold": int(aws.ToInt32(healthCheck.UnhealthyThreshold)),
				}
				mListener["health_check"] = []interface{}{mHealthCheck}
			}

			if portMapping := listener.PortMapping; portMapping != nil {
				mPortMapping := map[string]interface{}{
					"port":     int(aws.ToInt32(portMapping.Port)),
					"protocol": string(portMapping.Protocol),
				}
				mListener["port_mapping"] = []interface{}{mPortMapping}
			}

			if tls := listener.Tls; tls != nil {
				mTls := map[string]interface{}{
					"mode": string(tls.Mode),
				}

				if certificate := tls.Certificate; certificate != nil {
					mCertificate := map[string]interface{}{}

					switch v := certificate.(type) {
					case *awstypes.VirtualGatewayListenerTlsCertificateMemberAcm:
						mAcm := map[string]interface{}{
							"certificate_arn": aws.ToString(v.Value.CertificateArn),
						}

						mCertificate["acm"] = []interface{}{mAcm}
					case *awstypes.VirtualGatewayListenerTlsCertificateMemberFile:
						mFile := map[string]interface{}{
							"certificate_chain": aws.ToString(v.Value.CertificateChain),
							"private_key":       aws.ToString(v.Value.PrivateKey),
						}

						mCertificate["file"] = []interface{}{mFile}
					case *awstypes.VirtualGatewayListenerTlsCertificateMemberSds:
						mSds := map[string]interface{}{
							"secret_name": aws.ToString(v.Value.SecretName),
						}

						mCertificate["sds"] = []interface{}{mSds}
					}

					mTls["certificate"] = []interface{}{mCertificate}
				}

				if validation := tls.Validation; validation != nil {
					mValidation := map[string]interface{}{}

					if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
						mSubjectAlternativeNames := map[string]interface{}{}

						if match := subjectAlternativeNames.Match; match != nil {
							mMatch := map[string]interface{}{
								"exact": flex.FlattenStringValueSet(match.Exact),
							}

							mSubjectAlternativeNames["match"] = []interface{}{mMatch}
						}

						mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
					}

					if trust := validation.Trust; trust != nil {
						mTrust := map[string]interface{}{}

						switch v := trust.(type) {
						case *awstypes.VirtualGatewayListenerTlsValidationContextTrustMemberFile:
							mFile := map[string]interface{}{
								"certificate_chain": aws.ToString(v.Value.CertificateChain),
							}

							mTrust["file"] = []interface{}{mFile}
						case *awstypes.VirtualGatewayListenerTlsValidationContextTrustMemberSds:
							mSds := map[string]interface{}{
								"secret_name": aws.ToString(v.Value.SecretName),
							}

							mTrust["sds"] = []interface{}{mSds}
						}

						mValidation["trust"] = []interface{}{mTrust}
					}

					mTls["validation"] = []interface{}{mValidation}
				}

				mListener["tls"] = []interface{}{mTls}
			}
			mListeners = append(mListeners, mListener)
		}
		mSpec["listener"] = mListeners
	}

	if logging := spec.Logging; logging != nil {
		mLogging := map[string]interface{}{}

		if accessLog := logging.AccessLog; accessLog != nil {
			mAccessLog := map[string]interface{}{}

			switch v := accessLog.(type) {
			case *awstypes.VirtualGatewayAccessLogMemberFile:
				mFile := map[string]interface{}{}

				if format := v.Value.Format; format != nil {
					mFormat := map[string]interface{}{}

					switch v := format.(type) {
					case *awstypes.LoggingFormatMemberJson:
						vJsons := []interface{}{}

						for _, j := range v.Value {
							mJson := map[string]interface{}{
								"key":   aws.ToString(j.Key),
								"value": aws.ToString(j.Value),
							}

							vJsons = append(vJsons, mJson)
						}

						mFormat["json"] = vJsons
					case *awstypes.LoggingFormatMemberText:
						mFormat["text"] = v.Value
						mFile["format"] = []interface{}{mFormat}
					}
				}

				mFile["path"] = aws.ToString(v.Value.Path)

				mAccessLog["file"] = []interface{}{mFile}
			}

			mLogging["access_log"] = []interface{}{mAccessLog}
		}

		mSpec["logging"] = []interface{}{mLogging}
	}

	return []interface{}{mSpec}
}

func flattenVirtualGatewayClientPolicy(clientPolicy *awstypes.VirtualGatewayClientPolicy) []interface{} {
	if clientPolicy == nil {
		return []interface{}{}
	}

	mClientPolicy := map[string]interface{}{}

	if tls := clientPolicy.Tls; tls != nil {
		mTls := map[string]interface{}{
			"enforce": aws.ToBool(tls.Enforce),
			"ports":   flex.FlattenInt32ValueSet(tls.Ports),
		}

		if certificate := tls.Certificate; certificate != nil {
			mCertificate := map[string]interface{}{}

			switch v := certificate.(type) {
			case *awstypes.VirtualGatewayClientTlsCertificateMemberFile:
				mFile := map[string]interface{}{
					"certificate_chain": aws.ToString(v.Value.CertificateChain),
					"private_key":       aws.ToString(v.Value.PrivateKey),
				}

				mCertificate["file"] = []interface{}{mFile}
			case *awstypes.VirtualGatewayClientTlsCertificateMemberSds:
				mSds := map[string]interface{}{
					"secret_name": aws.ToString(v.Value.SecretName),
				}

				mCertificate["sds"] = []interface{}{mSds}
			}

			mTls["certificate"] = []interface{}{mCertificate}
		}

		if validation := tls.Validation; validation != nil {
			mValidation := map[string]interface{}{}

			if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
				mSubjectAlternativeNames := map[string]interface{}{}

				if match := subjectAlternativeNames.Match; match != nil {
					mMatch := map[string]interface{}{
						"exact": flex.FlattenStringValueSet(match.Exact),
					}

					mSubjectAlternativeNames["match"] = []interface{}{mMatch}
				}

				mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
			}

			if trust := validation.Trust; trust != nil {
				mTrust := map[string]interface{}{}

				switch v := trust.(type) {
				case *awstypes.VirtualGatewayTlsValidationContextTrustMemberAcm:
					mAcm := map[string]interface{}{
						"certificate_authority_arns": flex.FlattenStringValueSet(v.Value.CertificateAuthorityArns),
					}

					mTrust["acm"] = []interface{}{mAcm}
				case *awstypes.VirtualGatewayTlsValidationContextTrustMemberFile:
					mFile := map[string]interface{}{
						"certificate_chain": aws.ToString(v.Value.CertificateChain),
					}

					mTrust["file"] = []interface{}{mFile}
				case *awstypes.VirtualGatewayTlsValidationContextTrustMemberSds:
					mSds := map[string]interface{}{
						"secret_name": aws.ToString(v.Value.SecretName),
					}

					mTrust["sds"] = []interface{}{mSds}
				}

				mValidation["trust"] = []interface{}{mTrust}
			}

			mTls["validation"] = []interface{}{mValidation}
		}

		mClientPolicy["tls"] = []interface{}{mTls}
	}

	return []interface{}{mClientPolicy}
}
