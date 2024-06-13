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
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_appmesh_virtual_node", name="Virtual Node")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go/service/appmesh;appmesh.VirtualNodeData")
// @Testing(serialize=true)
// @Testing(importStateIdFunc=testAccVirtualNodeImportStateIdFunc)
func resourceVirtualNode() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: resourceVirtualNodeCreate,
		ReadWithoutTimeout:   resourceVirtualNodeRead,
		UpdateWithoutTimeout: resourceVirtualNodeUpdate,
		DeleteWithoutTimeout: resourceVirtualNodeDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceVirtualNodeImport,
		},

		SchemaVersion: 1,
		MigrateState:  resourceVirtualNodeMigrateState,

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
				"spec":            resourceVirtualNodeSpecSchema(),
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVirtualNodeSpecSchema() *schema.Schema {
	clientPolicySchema := func() *schema.Schema {
		return &schema.Schema{
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
								"enforce": {
									Type:     schema.TypeBool,
									Optional: true,
									Default:  true,
								},
								"ports": {
									Type:     schema.TypeSet,
									Optional: true,
									Elem:     &schema.Schema{Type: schema.TypeInt},
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
																		Elem:     &schema.Schema{Type: schema.TypeString},
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
		}
	}

	return &schema.Schema{
		Type:     schema.TypeList,
		Required: true,
		MinItems: 1,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"backend": {
					Type:     schema.TypeSet,
					Optional: true,
					MinItems: 0,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"virtual_service": {
								Type:     schema.TypeList,
								Required: true,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"client_policy": clientPolicySchema(),
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
				"backend_defaults": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"client_policy": clientPolicySchema(),
						},
					},
				},
				"listener": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
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
										"tcp": {
											Type:     schema.TypeList,
											Optional: true,
											MinItems: 0,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"max_connections": {
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
											ValidateFunc: validation.StringInSlice(appmesh.PortProtocol_Values(), false),
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
							"outlier_detection": {
								Type:     schema.TypeList,
								Optional: true,
								MinItems: 0,
								MaxItems: 1,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"base_ejection_duration": {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrUnit: {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
													},
													names.AttrValue: {
														Type:     schema.TypeInt,
														Required: true,
													},
												},
											},
										},
										names.AttrInterval: {
											Type:     schema.TypeList,
											Required: true,
											MinItems: 1,
											MaxItems: 1,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													names.AttrUnit: {
														Type:         schema.TypeString,
														Required:     true,
														ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
													},
													names.AttrValue: {
														Type:     schema.TypeInt,
														Required: true,
													},
												},
											},
										},
										"max_ejection_percent": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntBetween(0, 100),
										},
										"max_server_errors": {
											Type:         schema.TypeInt,
											Required:     true,
											ValidateFunc: validation.IntAtLeast(1),
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
											ValidateFunc: validation.StringInSlice(appmesh.PortProtocol_Values(), false),
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
										"grpc": {
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
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
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
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
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
										"http": {
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
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
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
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
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
										"http2": {
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
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
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
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
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
										"tcp": {
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
																	Type:         schema.TypeString,
																	Required:     true,
																	ValidateFunc: validation.StringInSlice(appmesh.DurationUnit_Values(), false),
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
											ValidateFunc: validation.StringInSlice(appmesh.ListenerTlsMode_Values(), false),
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
				"service_discovery": {
					Type:     schema.TypeList,
					Optional: true,
					MinItems: 0,
					MaxItems: 1,
					Elem: &schema.Resource{
						Schema: map[string]*schema.Schema{
							"aws_cloud_map": {
								Type:          schema.TypeList,
								Optional:      true,
								MinItems:      0,
								MaxItems:      1,
								ConflictsWith: []string{"spec.0.service_discovery.0.dns"},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										names.AttrAttributes: {
											Type:     schema.TypeMap,
											Optional: true,
											Elem:     &schema.Schema{Type: schema.TypeString},
										},
										"namespace_name": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 1024),
										},
										names.AttrServiceName: {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.StringLenBetween(1, 1024),
										},
									},
								},
							},
							"dns": {
								Type:          schema.TypeList,
								Optional:      true,
								MinItems:      0,
								MaxItems:      1,
								ConflictsWith: []string{"spec.0.service_discovery.0.aws_cloud_map"},
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"hostname": {
											Type:         schema.TypeString,
											Required:     true,
											ValidateFunc: validation.NoZeroValues,
										},
										"ip_preference": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringInSlice(appmesh.IpPreference_Values(), false),
										},
										"response_type": {
											Type:         schema.TypeString,
											Optional:     true,
											ValidateFunc: validation.StringInSlice(appmesh.DnsResponseType_Values(), false),
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

func resourceVirtualNodeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &appmesh.CreateVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		Spec:            expandVirtualNodeSpec(d.Get("spec").([]interface{})),
		Tags:            getTagsIn(ctx),
		VirtualNodeName: aws.String(name),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	resp, err := conn.CreateVirtualNodeWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Mesh Virtual Node (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(resp.VirtualNode.Metadata.Uid))

	return append(diags, resourceVirtualNodeRead(ctx, d, meta)...)
}

func resourceVirtualNodeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return findVirtualNodeByThreePartKey(ctx, conn, d.Get("mesh_name").(string), d.Get("mesh_owner").(string), d.Get(names.AttrName).(string))
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Mesh Virtual Node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Mesh Virtual Node (%s): %s", d.Id(), err)
	}

	vn := outputRaw.(*appmesh.VirtualNodeData)

	arn := aws.StringValue(vn.Metadata.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrCreatedDate, vn.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set(names.AttrLastUpdatedDate, vn.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", vn.MeshName)
	d.Set("mesh_owner", vn.Metadata.MeshOwner)
	d.Set(names.AttrName, vn.VirtualNodeName)
	d.Set(names.AttrResourceOwner, vn.Metadata.ResourceOwner)
	if err := d.Set("spec", flattenVirtualNodeSpec(vn.Spec)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting spec: %s", err)
	}

	return diags
}

func resourceVirtualNodeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	if d.HasChange("spec") {
		input := &appmesh.UpdateVirtualNodeInput{
			MeshName:        aws.String(d.Get("mesh_name").(string)),
			Spec:            expandVirtualNodeSpec(d.Get("spec").([]interface{})),
			VirtualNodeName: aws.String(d.Get(names.AttrName).(string)),
		}

		if v, ok := d.GetOk("mesh_owner"); ok {
			input.MeshOwner = aws.String(v.(string))
		}

		_, err := conn.UpdateVirtualNodeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Mesh Virtual Node (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceVirtualNodeRead(ctx, d, meta)...)
}

func resourceVirtualNodeDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)

	log.Printf("[DEBUG] Deleting App Mesh Virtual Node: %s", d.Id())
	input := &appmesh.DeleteVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("mesh_owner"); ok {
		input.MeshOwner = aws.String(v.(string))
	}

	_, err := conn.DeleteVirtualNodeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Mesh Virtual Node (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceVirtualNodeImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("wrong format of import ID (%s), use: 'mesh-name/virtual-node-name'", d.Id())
	}

	conn := meta.(*conns.AWSClient).AppMeshConn(ctx)
	meshName := parts[0]
	name := parts[1]

	vn, err := findVirtualNodeByThreePartKey(ctx, conn, meshName, "", name)

	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(vn.Metadata.Uid))
	d.Set("mesh_name", vn.MeshName)
	d.Set(names.AttrName, vn.VirtualNodeName)

	return []*schema.ResourceData{d}, nil
}

func findVirtualNodeByThreePartKey(ctx context.Context, conn *appmesh.AppMesh, meshName, meshOwner, name string) (*appmesh.VirtualNodeData, error) {
	input := &appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(meshName),
		VirtualNodeName: aws.String(name),
	}
	if meshOwner != "" {
		input.MeshOwner = aws.String(meshOwner)
	}

	output, err := findVirtualNode(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.Status.Status); status == appmesh.VirtualNodeStatusCodeDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return output, nil
}

func findVirtualNode(ctx context.Context, conn *appmesh.AppMesh, input *appmesh.DescribeVirtualNodeInput) (*appmesh.VirtualNodeData, error) {
	output, err := conn.DescribeVirtualNodeWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VirtualNode == nil || output.VirtualNode.Metadata == nil || output.VirtualNode.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.VirtualNode, nil
}
