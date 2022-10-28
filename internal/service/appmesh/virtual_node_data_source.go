package appmesh

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appmesh"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
)

func DataSourceVirtualNode() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVirtualNodeRead,

		Schema: map[string]*schema.Schema{
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
				Type:     schema.TypeString,
				Required: true,
			},

			"mesh_owner": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"name": {
				Type:     schema.TypeString,
				Required: true,
			},

			"resource_owner": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backend": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"virtual_service": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"virtual_service_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 255),
												},

												"client_policy": VirtualNodeClientPolicySchema(),
											},
										},
									},
								},
							},
						},

						"backend_defaults": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_policy": VirtualNodeClientPolicySchema(),
								},
							},
						},

						"listener": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection_pool": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"grpc": {
													Type:     schema.
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_requests": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.connection_pool.0.grpc",
														"spec.0.listener.0.connection_pool.0.http",
														"spec.0.listener.0.connection_pool.0.http2",
														"spec.0.listener.0.connection_pool.0.tcp",
													},
												},

												"http": {
													Type:     schema.TypeList,
													Optional: true,
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
													ExactlyOneOf: []string{
														"spec.0.listener.0.connection_pool.0.grpc",
														"spec.0.listener.0.connection_pool.0.http",
														"spec.0.listener.0.connection_pool.0.http2",
														"spec.0.listener.0.connection_pool.0.tcp",
													},
												},

												"http2": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_requests": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.connection_pool.0.grpc",
														"spec.0.listener.0.connection_pool.0.http",
														"spec.0.listener.0.connection_pool.0.http2",
														"spec.0.listener.0.connection_pool.0.tcp",
													},
												},

												"tcp": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"max_connections": {
																Type:         schema.TypeInt,
																Required:     true,
																ValidateFunc: validation.IntAtLeast(1),
															},
														},
													},
													ExactlyOneOf: []string{
														"spec.0.listener.0.connection_pool.0.grpc",
														"spec.0.listener.0.connection_pool.0.http",
														"spec.0.listener.0.connection_pool.0.http2",
														"spec.0.listener.0.connection_pool.0.tcp",
													},
												},
											},
										},
									},

									"health_check": {
										Type:     schema.TypeList,
										Optional: true,
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
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"base_ejection_duration": {
													Type:     schema.TypeList,
													Required: true,
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

												"interval": {
													Type:     schema.TypeList,
													Required: true,
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
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:         schema.TypeInt,
													Required:     true,
													ValidateFunc: validation.IsPortNumber,
												},

												"protocol": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.PortProtocol_Values(), false),
												},
											},
										},
									},

									"timeout": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"grpc": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
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
													ExactlyOneOf: []string{
														"spec.0.listener.0.timeout.0.grpc",
														"spec.0.listener.0.timeout.0.http",
														"spec.0.listener.0.timeout.0.http2",
														"spec.0.listener.0.timeout.0.tcp",
													},
												},

												"http": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
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
													ExactlyOneOf: []string{
														"spec.0.listener.0.timeout.0.grpc",
														"spec.0.listener.0.timeout.0.http",
														"spec.0.listener.0.timeout.0.http2",
														"spec.0.listener.0.timeout.0.tcp",
													},
												},

												"http2": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
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
													ExactlyOneOf: []string{
														"spec.0.listener.0.timeout.0.grpc",
														"spec.0.listener.0.timeout.0.http",
														"spec.0.listener.0.timeout.0.http2",
														"spec.0.listener.0.timeout.0.tcp",
													},
												},

												"tcp": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"idle": {
																Type:     schema.TypeList,
																Optional: true,
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
													ExactlyOneOf: []string{
														"spec.0.listener.0.timeout.0.grpc",
														"spec.0.listener.0.timeout.0.http",
														"spec.0.listener.0.timeout.0.http2",
														"spec.0.listener.0.timeout.0.tcp",
													},
												},
											},
										},
									},

									"tls": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"certificate": {
													Type:     schema.TypeList,
													Required: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"acm": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"certificate_arn": {
																			Type:         schema.TypeString,
																			Required:     true,
																			ValidateFunc: verify.ValidARN,
																		},
																	},
																},
																ExactlyOneOf: []string{
																	"spec.0.listener.0.tls.0.certificate.0.acm",
																	"spec.0.listener.0.tls.0.certificate.0.file",
																	"spec.0.listener.0.tls.0.certificate.0.sds",
																},
															},

															"file": {
																Type:     schema.TypeList,
																Optional: true,
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
																	"spec.0.listener.0.tls.0.certificate.0.acm",
																	"spec.0.listener.0.tls.0.certificate.0.file",
																	"spec.0.listener.0.tls.0.certificate.0.sds",
																},
															},

															"sds": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"secret_name": {
																			Type:     schema.TypeString,
																			Required: true,
																		},
																	},
																},
																ExactlyOneOf: []string{
																	"spec.0.listener.0.tls.0.certificate.0.acm",
																	"spec.0.listener.0.tls.0.certificate.0.file",
																	"spec.0.listener.0.tls.0.certificate.0.sds",
																},
															},
														},
													},
												},

												"mode": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(appmesh.ListenerTlsMode_Values(), false),
												},

												"validation": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"subject_alternative_names": {
																Type:     schema.TypeList,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"match": {
																			Type:     schema.TypeList,
																			Required: true,
																			Elem: &schema.Resource{
																				Schema: map[string]*schema.Schema{
																					"exact": {
																						Type:     schema.TypeSet,
																						Required: true,
																						Elem:     &schema.Schema{Type: schema.TypeString},
																						Set:      schema.HashString,
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
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"file": {
																			Type:     schema.TypeList,
																			Optional: true,
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
																				"spec.0.listener.0.tls.0.validation.0.trust.0.file",
																				"spec.0.listener.0.tls.0.validation.0.trust.0.sds",
																			},
																		},

																		"sds": {
																			Type:     schema.TypeList,
																			Optional: true,
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
																				"spec.0.listener.0.tls.0.validation.0.trust.0.file",
																				"spec.0.listener.0.tls.0.validation.0.trust.0.sds",
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
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_log": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"file": {
													Type:     schema.TypeList,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
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

						"service_discovery": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"aws_cloud_map": {
										Type:          schema.TypeList,
										Optional:      true,
										ConflictsWith: []string{"spec.0.service_discovery.0.dns"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"attributes": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem:     &schema.Schema{Type: schema.TypeString},
												},

												"namespace_name": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(1, 1024),
												},

												"service_name": {
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
										ConflictsWith: []string{"spec.0.service_discovery.0.aws_cloud_map"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"hostname": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.NoZeroValues,
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

			"tags": tftags.TagsSchema(),
		},
	}
}

func dataSourceVirtualNodeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppMeshConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	var resp *appmesh.DescribeVirtualNodeOutput

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.DescribeVirtualNode(req)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.DescribeVirtualNode(req)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, appmesh.ErrCodeNotFoundException) {
		log.Printf("[WARN] App Mesh Virtual Node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading App Mesh Virtual Node: %w", err)
	}

	if resp == nil || resp.VirtualNode == nil {
		return fmt.Errorf("reading App Mesh Virtual Node: empty response")
	}

	if aws.StringValue(resp.VirtualNode.Status.Status) == appmesh.VirtualNodeStatusCodeDeleted {
		if d.IsNewResource() {
			return fmt.Errorf("reading App Mesh Virtual Node: %s after creation", aws.StringValue(resp.VirtualNode.Status.Status))
		}

		log.Printf("[WARN] App Mesh Virtual Node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(resp.VirtualNode.Metadata.Arn)
	d.Set("name", resp.VirtualNode.VirtualNodeName)
	d.Set("mesh_name", resp.VirtualNode.MeshName)
	d.Set("mesh_owner", resp.VirtualNode.Metadata.MeshOwner)
	d.Set("arn", arn)
	d.Set("created_date", resp.VirtualNode.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualNode.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("resource_owner", resp.VirtualNode.Metadata.ResourceOwner)
	err = d.Set("spec", flattenVirtualNodeSpec(resp.VirtualNode.Spec))
	if err != nil {
		return fmt.Errorf("setting spec: %w", err)
	}

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("listing tags for App Mesh virtual node (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}
