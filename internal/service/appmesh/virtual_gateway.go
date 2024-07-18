// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appmesh

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appmesh_virtual_gateway", name="Virtual Gateway")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/appmesh;appmesh.VirtualGatewayData")
// @Testing(serialize=true)
// @Testing(importStateIdFunc=testAccVirtualGatewayImportStateIdFunc)
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
													names.AttrCertificate: {
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
																			names.AttrCertificateChain: {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: validation.StringLenBetween(1, 255),
																			},
																			names.AttrPrivateKey: {
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
																						names.AttrCertificateChain: {
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
							names.AttrHealthCheck: {
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
										names.AttrPath: {
											Type:     schema.TypeString,
											Optional: true,
										},
										names.AttrPort: {
											Type:         schema.TypeInt,
											Optional:     true,
											Computed:     true,
											ValidateFunc: validation.IsPortNumber,
										},
										names.AttrProtocol: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(appmesh.VirtualGatewayPortProtocol_Values(), false),
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
										names.AttrPort: {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IsPortNumber,
										},
										names.AttrProtocol: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(appmesh.VirtualGatewayPortProtocol_Values(), false),
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
										names.AttrCertificate: {
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
																names.AttrCertificateARN: {
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
																names.AttrCertificateChain: {
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringLenBetween(1, 255),
																},
																names.AttrPrivateKey: {
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
										names.AttrMode: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringInSlice(appmesh.VirtualGatewayListenerTlsMode_Values(), false),
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
																			names.AttrCertificateChain: {
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
													names.AttrFormat: {
														Type:     schema.TypeList,
														Optional: true,
														MinItems: 0,
														MaxItems: 1,
														Elem: &schema.Resource{
															Schema: map[string]*schema.Schema{
																names.AttrJSON: {
																	Type:     schema.TypeList,
																	Optional: true,
																	Elem: &schema.Resource{
																		Schema: map[string]*schema.Schema{
																			names.AttrKey: {
																				Type:         schema.TypeString,
																				Required:     true,
																				ValidateFunc: validation.StringLenBetween(1, 100),
																			},
																			names.AttrValue: {
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
													names.AttrPath: {
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
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appmesh.CreateVirtualGatewayInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		Spec:               expandVirtualGatewaySpec(d.Get("spec").([]interface{})),
		Tags:               getTagsIn(ctx),
		VirtualGatewayName: aws.String(name),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	output, err := conn.CreateVirtualGatewayWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Virtual Gateway (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.VirtualGateway.Metadata.Uid))

	return append(diags, resourceVirtualGatewayRead(ctx, d, meta)...)
}

func resourceVirtualGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findVirtualGatewayByThreePartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), d.Get(names.AttrName).(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Mesh Virtual Gateway (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Gateway (%s): %s", d.Id(), err)
	}

	virtualGateway := outputRaw.(*appmesh.VirtualGatewayData)

	arn := aws.StringValue(virtualGateway.Metadata.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, virtualGateway.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedDate, virtualGateway.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", virtualGateway.MeshName)
	d.Set("mesh_owner", virtualGateway.Metadata.MeshOwner)
	d.Set(names.AttrName, virtualGateway.VirtualGatewayName)
	d.Set(names.AttrResourceOwner, virtualGateway.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenVirtualGatewaySpec(virtualGateway.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	return diags
}

func resourceVirtualGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	if d.HasChange("spec") {
		input := &appmesh.UpdateVirtualGatewayInput{
			MeshName:           aws.String(d.Get("mesh_name").(string)),
			Spec:               expandVirtualGatewaySpec(d.Get("spec").([]interface{})),
			VirtualGatewayName: aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk("mesh_owner"); ok {
			input.MeshOwner = aws.String(v.(string))
		}

		_, err := conn.UpdateVirtualGatewayWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Virtual Gateway (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVirtualGatewayRead(ctx, d, meta)...)
}

func resourceVirtualGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	log.Printf("[DEBUG] Deleting App Mesh Virtual Gateway: %s", d.Id())
	input := &appmesh.DeleteVirtualGatewayInput{
		MeshName:           aws.String(d.Get("mesh_name").(string)),
		VirtualGatewayName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	_, err := conn.DeleteVirtualGatewayWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
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

	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	virtualGateway, err := findVirtualGatewayByThreePartKey(ctx, conn, meshName, "", name)

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(virtualGateway.Metadata.Uid))
	d.Set("mesh_name", virtualGateway.MeshName)
	d.Set(names.AttrName, virtualGateway.VirtualGatewayName)

	return []*schema.ResourceData{d}, nil
}

func findVirtualGatewayByThreePartKey(ctx context.Context, conn *appmesh.AppMesh, meshName, meshOwner, name string) (*appmesh.VirtualGatewayData, error) {
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

	if status := aws.StringValue(output.Status.Status); status == appmesh.VirtualGatewayStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findVirtualGateway(ctx context.Context, conn *appmesh.AppMesh, input *appmesh.DescribeVirtualGatewayInput) (*appmesh.VirtualGatewayData, error) {
	output, err := conn.DescribeVirtualGatewayWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
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

func expandVirtualGatewaySpec(vSpec []interface{}) *appmesh.VirtualGatewaySpec {
	if len(vSpec) == 0 || vSpec[0] == nil {
		return nil
	}

	spec := &appmesh.VirtualGatewaySpec{}

	mSpec := vSpec[0].(map[string]interface{})

	if vBackendDefaults, ok := mSpec["backend_defaults"].([]interface{}); ok && len(vBackendDefaults) > 0 && vBackendDefaults[0] != nil {
		backendDefaults := &appmesh.VirtualGatewayBackendDefaults{}

		mBackendDefaults := vBackendDefaults[0].(map[string]interface{})

		if vClientPolicy, ok := mBackendDefaults["client_policy"].([]interface{}); ok {
			backendDefaults.ClientPolicy = expandVirtualGatewayClientPolicy(vClientPolicy)
		}

		spec.BackendDefaults = backendDefaults
	}

	if vListeners, ok := mSpec["listener"].([]interface{}); ok && len(vListeners) > 0 && vListeners[0] != nil {
		listeners := []*appmesh.VirtualGatewayListener{}

		for _, vListener := range vListeners {
			listener := &appmesh.VirtualGatewayListener{}

			mListener := vListener.(map[string]interface{})

			if vHealthCheck, ok := mListener[names.AttrHealthCheck].([]interface{}); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
				healthCheck := &appmesh.VirtualGatewayHealthCheckPolicy{}

				mHealthCheck := vHealthCheck[0].(map[string]interface{})

				if vHealthyThreshold, ok := mHealthCheck["healthy_threshold"].(int); ok && vHealthyThreshold > 0 {
					healthCheck.HealthyThreshold = aws.Int64(int64(vHealthyThreshold))
				}
				if vIntervalMillis, ok := mHealthCheck["interval_millis"].(int); ok && vIntervalMillis > 0 {
					healthCheck.IntervalMillis = aws.Int64(int64(vIntervalMillis))
				}
				if vPath, ok := mHealthCheck[names.AttrPath].(string); ok && vPath != "" {
					healthCheck.Path = aws.String(vPath)
				}
				if vPort, ok := mHealthCheck[names.AttrPort].(int); ok && vPort > 0 {
					healthCheck.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mHealthCheck[names.AttrProtocol].(string); ok && vProtocol != "" {
					healthCheck.Protocol = aws.String(vProtocol)
				}
				if vTimeoutMillis, ok := mHealthCheck["timeout_millis"].(int); ok && vTimeoutMillis > 0 {
					healthCheck.TimeoutMillis = aws.Int64(int64(vTimeoutMillis))
				}
				if vUnhealthyThreshold, ok := mHealthCheck["unhealthy_threshold"].(int); ok && vUnhealthyThreshold > 0 {
					healthCheck.UnhealthyThreshold = aws.Int64(int64(vUnhealthyThreshold))
				}

				listener.HealthCheck = healthCheck
			}

			if vConnectionPool, ok := mListener["connection_pool"].([]interface{}); ok && len(vConnectionPool) > 0 && vConnectionPool[0] != nil {
				mConnectionPool := vConnectionPool[0].(map[string]interface{})

				connectionPool := &appmesh.VirtualGatewayConnectionPool{}

				if vGrpcConnectionPool, ok := mConnectionPool["grpc"].([]interface{}); ok && len(vGrpcConnectionPool) > 0 && vGrpcConnectionPool[0] != nil {
					mGrpcConnectionPool := vGrpcConnectionPool[0].(map[string]interface{})

					grpcConnectionPool := &appmesh.VirtualGatewayGrpcConnectionPool{}

					if vMaxRequests, ok := mGrpcConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						grpcConnectionPool.MaxRequests = aws.Int64(int64(vMaxRequests))
					}

					connectionPool.Grpc = grpcConnectionPool
				}

				if vHttpConnectionPool, ok := mConnectionPool["http"].([]interface{}); ok && len(vHttpConnectionPool) > 0 && vHttpConnectionPool[0] != nil {
					mHttpConnectionPool := vHttpConnectionPool[0].(map[string]interface{})

					httpConnectionPool := &appmesh.VirtualGatewayHttpConnectionPool{}

					if vMaxConnections, ok := mHttpConnectionPool["max_connections"].(int); ok && vMaxConnections > 0 {
						httpConnectionPool.MaxConnections = aws.Int64(int64(vMaxConnections))
					}
					if vMaxPendingRequests, ok := mHttpConnectionPool["max_pending_requests"].(int); ok && vMaxPendingRequests > 0 {
						httpConnectionPool.MaxPendingRequests = aws.Int64(int64(vMaxPendingRequests))
					}

					connectionPool.Http = httpConnectionPool
				}

				if vHttp2ConnectionPool, ok := mConnectionPool["http2"].([]interface{}); ok && len(vHttp2ConnectionPool) > 0 && vHttp2ConnectionPool[0] != nil {
					mHttp2ConnectionPool := vHttp2ConnectionPool[0].(map[string]interface{})

					http2ConnectionPool := &appmesh.VirtualGatewayHttp2ConnectionPool{}

					if vMaxRequests, ok := mHttp2ConnectionPool["max_requests"].(int); ok && vMaxRequests > 0 {
						http2ConnectionPool.MaxRequests = aws.Int64(int64(vMaxRequests))
					}

					connectionPool.Http2 = http2ConnectionPool
				}

				listener.ConnectionPool = connectionPool
			}

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				portMapping := &appmesh.VirtualGatewayPortMapping{}

				mPortMapping := vPortMapping[0].(map[string]interface{})

				if vPort, ok := mPortMapping[names.AttrPort].(int); ok && vPort > 0 {
					portMapping.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mPortMapping[names.AttrProtocol].(string); ok && vProtocol != "" {
					portMapping.Protocol = aws.String(vProtocol)
				}

				listener.PortMapping = portMapping
			}

			if vTls, ok := mListener["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
				tls := &appmesh.VirtualGatewayListenerTls{}

				mTls := vTls[0].(map[string]interface{})

				if vMode, ok := mTls[names.AttrMode].(string); ok && vMode != "" {
					tls.Mode = aws.String(vMode)
				}

				if vCertificate, ok := mTls[names.AttrCertificate].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
					certificate := &appmesh.VirtualGatewayListenerTlsCertificate{}

					mCertificate := vCertificate[0].(map[string]interface{})

					if vAcm, ok := mCertificate["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
						acm := &appmesh.VirtualGatewayListenerTlsAcmCertificate{}

						mAcm := vAcm[0].(map[string]interface{})

						if vCertificateArn, ok := mAcm[names.AttrCertificateARN].(string); ok && vCertificateArn != "" {
							acm.CertificateArn = aws.String(vCertificateArn)
						}

						certificate.Acm = acm
					}

					if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
						file := &appmesh.VirtualGatewayListenerTlsFileCertificate{}

						mFile := vFile[0].(map[string]interface{})

						if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
							file.CertificateChain = aws.String(vCertificateChain)
						}
						if vPrivateKey, ok := mFile[names.AttrPrivateKey].(string); ok && vPrivateKey != "" {
							file.PrivateKey = aws.String(vPrivateKey)
						}

						certificate.File = file
					}

					if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
						sds := &appmesh.VirtualGatewayListenerTlsSdsCertificate{}

						mSds := vSds[0].(map[string]interface{})

						if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
							sds.SecretName = aws.String(vSecretName)
						}

						certificate.Sds = sds
					}

					tls.Certificate = certificate
				}

				if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
					validation := &appmesh.VirtualGatewayListenerTlsValidationContext{}

					mValidation := vValidation[0].(map[string]interface{})

					if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
						subjectAlternativeNames := &appmesh.SubjectAlternativeNames{}

						mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

						if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
							match := &appmesh.SubjectAlternativeNameMatchers{}

							mMatch := vMatch[0].(map[string]interface{})

							if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
								match.Exact = flex.ExpandStringSet(vExact)
							}

							subjectAlternativeNames.Match = match
						}

						validation.SubjectAlternativeNames = subjectAlternativeNames
					}

					if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
						trust := &appmesh.VirtualGatewayListenerTlsValidationContextTrust{}

						mTrust := vTrust[0].(map[string]interface{})

						if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
							file := &appmesh.VirtualGatewayTlsValidationContextFileTrust{}

							mFile := vFile[0].(map[string]interface{})

							if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
								file.CertificateChain = aws.String(vCertificateChain)
							}

							trust.File = file
						}

						if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
							sds := &appmesh.VirtualGatewayTlsValidationContextSdsTrust{}

							mSds := vSds[0].(map[string]interface{})

							if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
								sds.SecretName = aws.String(vSecretName)
							}

							trust.Sds = sds
						}

						validation.Trust = trust
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
		logging := &appmesh.VirtualGatewayLogging{}

		mLogging := vLogging[0].(map[string]interface{})

		if vAccessLog, ok := mLogging["access_log"].([]interface{}); ok && len(vAccessLog) > 0 && vAccessLog[0] != nil {
			accessLog := &appmesh.VirtualGatewayAccessLog{}

			mAccessLog := vAccessLog[0].(map[string]interface{})

			if vFile, ok := mAccessLog["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				file := &appmesh.VirtualGatewayFileAccessLog{}

				mFile := vFile[0].(map[string]interface{})

				if vFormat, ok := mFile[names.AttrFormat].([]interface{}); ok && len(vFormat) > 0 && vFormat[0] != nil {
					format := &appmesh.LoggingFormat{}

					mFormat := vFormat[0].(map[string]interface{})

					if vJsonFormatRefs, ok := mFormat[names.AttrJSON].([]interface{}); ok && len(vJsonFormatRefs) > 0 {
						jsonFormatRefs := []*appmesh.JsonFormatRef{}
						for _, vJsonFormatRef := range vJsonFormatRefs {
							mJsonFormatRef := &appmesh.JsonFormatRef{
								Key:   aws.String(vJsonFormatRef.(map[string]interface{})[names.AttrKey].(string)),
								Value: aws.String(vJsonFormatRef.(map[string]interface{})[names.AttrValue].(string)),
							}
							jsonFormatRefs = append(jsonFormatRefs, mJsonFormatRef)
						}
						format.Json = jsonFormatRefs
					}

					if vText, ok := mFormat["text"].(string); ok && vText != "" {
						format.Text = aws.String(vText)
					}

					file.Format = format
				}

				if vPath, ok := mFile[names.AttrPath].(string); ok && vPath != "" {
					file.Path = aws.String(vPath)
				}

				accessLog.File = file
			}

			logging.AccessLog = accessLog
		}

		spec.Logging = logging
	}

	return spec
}

func expandVirtualGatewayClientPolicy(vClientPolicy []interface{}) *appmesh.VirtualGatewayClientPolicy {
	if len(vClientPolicy) == 0 || vClientPolicy[0] == nil {
		return nil
	}

	clientPolicy := &appmesh.VirtualGatewayClientPolicy{}

	mClientPolicy := vClientPolicy[0].(map[string]interface{})

	if vTls, ok := mClientPolicy["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
		tls := &appmesh.VirtualGatewayClientPolicyTls{}

		mTls := vTls[0].(map[string]interface{})

		if vCertificate, ok := mTls[names.AttrCertificate].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
			certificate := &appmesh.VirtualGatewayClientTlsCertificate{}

			mCertificate := vCertificate[0].(map[string]interface{})

			if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				file := &appmesh.VirtualGatewayListenerTlsFileCertificate{}

				mFile := vFile[0].(map[string]interface{})

				if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
					file.CertificateChain = aws.String(vCertificateChain)
				}
				if vPrivateKey, ok := mFile[names.AttrPrivateKey].(string); ok && vPrivateKey != "" {
					file.PrivateKey = aws.String(vPrivateKey)
				}

				certificate.File = file
			}

			if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
				sds := &appmesh.VirtualGatewayListenerTlsSdsCertificate{}

				mSds := vSds[0].(map[string]interface{})

				if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
					sds.SecretName = aws.String(vSecretName)
				}

				certificate.Sds = sds
			}

			tls.Certificate = certificate
		}

		if vEnforce, ok := mTls["enforce"].(bool); ok {
			tls.Enforce = aws.Bool(vEnforce)
		}

		if vPorts, ok := mTls["ports"].(*schema.Set); ok && vPorts.Len() > 0 {
			tls.Ports = flex.ExpandInt64Set(vPorts)
		}

		if vValidation, ok := mTls["validation"].([]interface{}); ok && len(vValidation) > 0 && vValidation[0] != nil {
			validation := &appmesh.VirtualGatewayTlsValidationContext{}

			mValidation := vValidation[0].(map[string]interface{})

			if vSubjectAlternativeNames, ok := mValidation["subject_alternative_names"].([]interface{}); ok && len(vSubjectAlternativeNames) > 0 && vSubjectAlternativeNames[0] != nil {
				subjectAlternativeNames := &appmesh.SubjectAlternativeNames{}

				mSubjectAlternativeNames := vSubjectAlternativeNames[0].(map[string]interface{})

				if vMatch, ok := mSubjectAlternativeNames["match"].([]interface{}); ok && len(vMatch) > 0 && vMatch[0] != nil {
					match := &appmesh.SubjectAlternativeNameMatchers{}

					mMatch := vMatch[0].(map[string]interface{})

					if vExact, ok := mMatch["exact"].(*schema.Set); ok && vExact.Len() > 0 {
						match.Exact = flex.ExpandStringSet(vExact)
					}

					subjectAlternativeNames.Match = match
				}

				validation.SubjectAlternativeNames = subjectAlternativeNames
			}

			if vTrust, ok := mValidation["trust"].([]interface{}); ok && len(vTrust) > 0 && vTrust[0] != nil {
				trust := &appmesh.VirtualGatewayTlsValidationContextTrust{}

				mTrust := vTrust[0].(map[string]interface{})

				if vAcm, ok := mTrust["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
					acm := &appmesh.VirtualGatewayTlsValidationContextAcmTrust{}

					mAcm := vAcm[0].(map[string]interface{})

					if vCertificateAuthorityArns, ok := mAcm["certificate_authority_arns"].(*schema.Set); ok && vCertificateAuthorityArns.Len() > 0 {
						acm.CertificateAuthorityArns = flex.ExpandStringSet(vCertificateAuthorityArns)
					}

					trust.Acm = acm
				}

				if vFile, ok := mTrust["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
					file := &appmesh.VirtualGatewayTlsValidationContextFileTrust{}

					mFile := vFile[0].(map[string]interface{})

					if vCertificateChain, ok := mFile[names.AttrCertificateChain].(string); ok && vCertificateChain != "" {
						file.CertificateChain = aws.String(vCertificateChain)
					}

					trust.File = file
				}

				if vSds, ok := mTrust["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
					sds := &appmesh.VirtualGatewayTlsValidationContextSdsTrust{}

					mSds := vSds[0].(map[string]interface{})

					if vSecretName, ok := mSds["secret_name"].(string); ok && vSecretName != "" {
						sds.SecretName = aws.String(vSecretName)
					}

					trust.Sds = sds
				}

				validation.Trust = trust
			}

			tls.Validation = validation
		}

		clientPolicy.Tls = tls
	}

	return clientPolicy
}

func flattenVirtualGatewaySpec(spec *appmesh.VirtualGatewaySpec) []interface{} {
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

				if grpcConnectionPool := connectionPool.Grpc; grpcConnectionPool != nil {
					mGrpcConnectionPool := map[string]interface{}{
						"max_requests": int(aws.Int64Value(grpcConnectionPool.MaxRequests)),
					}
					mConnectionPool["grpc"] = []interface{}{mGrpcConnectionPool}
				}

				if httpConnectionPool := connectionPool.Http; httpConnectionPool != nil {
					mHttpConnectionPool := map[string]interface{}{
						"max_connections":      int(aws.Int64Value(httpConnectionPool.MaxConnections)),
						"max_pending_requests": int(aws.Int64Value(httpConnectionPool.MaxPendingRequests)),
					}
					mConnectionPool["http"] = []interface{}{mHttpConnectionPool}
				}

				if http2ConnectionPool := connectionPool.Http2; http2ConnectionPool != nil {
					mHttp2ConnectionPool := map[string]interface{}{
						"max_requests": int(aws.Int64Value(http2ConnectionPool.MaxRequests)),
					}
					mConnectionPool["http2"] = []interface{}{mHttp2ConnectionPool}
				}

				mListener["connection_pool"] = []interface{}{mConnectionPool}
			}

			if healthCheck := listener.HealthCheck; healthCheck != nil {
				mHealthCheck := map[string]interface{}{
					"healthy_threshold":   int(aws.Int64Value(healthCheck.HealthyThreshold)),
					"interval_millis":     int(aws.Int64Value(healthCheck.IntervalMillis)),
					names.AttrPath:        aws.StringValue(healthCheck.Path),
					names.AttrPort:        int(aws.Int64Value(healthCheck.Port)),
					names.AttrProtocol:    aws.StringValue(healthCheck.Protocol),
					"timeout_millis":      int(aws.Int64Value(healthCheck.TimeoutMillis)),
					"unhealthy_threshold": int(aws.Int64Value(healthCheck.UnhealthyThreshold)),
				}
				mListener[names.AttrHealthCheck] = []interface{}{mHealthCheck}
			}

			if portMapping := listener.PortMapping; portMapping != nil {
				mPortMapping := map[string]interface{}{
					names.AttrPort:     int(aws.Int64Value(portMapping.Port)),
					names.AttrProtocol: aws.StringValue(portMapping.Protocol),
				}
				mListener["port_mapping"] = []interface{}{mPortMapping}
			}

			if tls := listener.Tls; tls != nil {
				mTls := map[string]interface{}{
					names.AttrMode: aws.StringValue(tls.Mode),
				}

				if certificate := tls.Certificate; certificate != nil {
					mCertificate := map[string]interface{}{}

					if acm := certificate.Acm; acm != nil {
						mAcm := map[string]interface{}{
							names.AttrCertificateARN: aws.StringValue(acm.CertificateArn),
						}

						mCertificate["acm"] = []interface{}{mAcm}
					}

					if file := certificate.File; file != nil {
						mFile := map[string]interface{}{
							names.AttrCertificateChain: aws.StringValue(file.CertificateChain),
							names.AttrPrivateKey:       aws.StringValue(file.PrivateKey),
						}

						mCertificate["file"] = []interface{}{mFile}
					}

					if sds := certificate.Sds; sds != nil {
						mSds := map[string]interface{}{
							"secret_name": aws.StringValue(sds.SecretName),
						}

						mCertificate["sds"] = []interface{}{mSds}
					}

					mTls[names.AttrCertificate] = []interface{}{mCertificate}
				}

				if validation := tls.Validation; validation != nil {
					mValidation := map[string]interface{}{}

					if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
						mSubjectAlternativeNames := map[string]interface{}{}

						if match := subjectAlternativeNames.Match; match != nil {
							mMatch := map[string]interface{}{
								"exact": flex.FlattenStringSet(match.Exact),
							}

							mSubjectAlternativeNames["match"] = []interface{}{mMatch}
						}

						mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
					}

					if trust := validation.Trust; trust != nil {
						mTrust := map[string]interface{}{}

						if file := trust.File; file != nil {
							mFile := map[string]interface{}{
								names.AttrCertificateChain: aws.StringValue(file.CertificateChain),
							}

							mTrust["file"] = []interface{}{mFile}
						}

						if sds := trust.Sds; sds != nil {
							mSds := map[string]interface{}{
								"secret_name": aws.StringValue(sds.SecretName),
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

			if file := accessLog.File; file != nil {
				mFile := map[string]interface{}{}

				if format := file.Format; format != nil {
					mFormat := map[string]interface{}{}

					if jsons := format.Json; jsons != nil {
						vJsons := []interface{}{}

						for _, j := range format.Json {
							mJson := map[string]interface{}{
								names.AttrKey:   aws.StringValue(j.Key),
								names.AttrValue: aws.StringValue(j.Value),
							}

							vJsons = append(vJsons, mJson)
						}

						mFormat[names.AttrJSON] = vJsons
					}

					if text := format.Text; text != nil {
						mFormat["text"] = aws.StringValue(text)
					}

					mFile[names.AttrFormat] = []interface{}{mFormat}
				}

				mFile[names.AttrPath] = aws.StringValue(file.Path)

				mAccessLog["file"] = []interface{}{mFile}
			}

			mLogging["access_log"] = []interface{}{mAccessLog}
		}

		mSpec["logging"] = []interface{}{mLogging}
	}

	return []interface{}{mSpec}
}

func flattenVirtualGatewayClientPolicy(clientPolicy *appmesh.VirtualGatewayClientPolicy) []interface{} {
	if clientPolicy == nil {
		return []interface{}{}
	}

	mClientPolicy := map[string]interface{}{}

	if tls := clientPolicy.Tls; tls != nil {
		mTls := map[string]interface{}{
			"enforce": aws.BoolValue(tls.Enforce),
			"ports":   flex.FlattenInt64Set(tls.Ports),
		}

		if certificate := tls.Certificate; certificate != nil {
			mCertificate := map[string]interface{}{}

			if file := certificate.File; file != nil {
				mFile := map[string]interface{}{
					names.AttrCertificateChain: aws.StringValue(file.CertificateChain),
					names.AttrPrivateKey:       aws.StringValue(file.PrivateKey),
				}

				mCertificate["file"] = []interface{}{mFile}
			}

			if sds := certificate.Sds; sds != nil {
				mSds := map[string]interface{}{
					"secret_name": aws.StringValue(sds.SecretName),
				}

				mCertificate["sds"] = []interface{}{mSds}
			}

			mTls[names.AttrCertificate] = []interface{}{mCertificate}
		}

		if validation := tls.Validation; validation != nil {
			mValidation := map[string]interface{}{}

			if subjectAlternativeNames := validation.SubjectAlternativeNames; subjectAlternativeNames != nil {
				mSubjectAlternativeNames := map[string]interface{}{}

				if match := subjectAlternativeNames.Match; match != nil {
					mMatch := map[string]interface{}{
						"exact": flex.FlattenStringSet(match.Exact),
					}

					mSubjectAlternativeNames["match"] = []interface{}{mMatch}
				}

				mValidation["subject_alternative_names"] = []interface{}{mSubjectAlternativeNames}
			}

			if trust := validation.Trust; trust != nil {
				mTrust := map[string]interface{}{}

				if acm := trust.Acm; acm != nil {
					mAcm := map[string]interface{}{
						"certificate_authority_arns": flex.FlattenStringSet(acm.CertificateAuthorityArns),
					}

					mTrust["acm"] = []interface{}{mAcm}
				}

				if file := trust.File; file != nil {
					mFile := map[string]interface{}{
						names.AttrCertificateChain: aws.StringValue(file.CertificateChain),
					}

					mTrust["file"] = []interface{}{mFile}
				}

				if sds := trust.Sds; sds != nil {
					mSds := map[string]interface{}{
						"secret_name": aws.StringValue(sds.SecretName),
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
