package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	appmesh "github.com/aws/aws-sdk-go/service/appmeshpreview"
	"github.com/hashicorp/terraform-plugin-sdk/helper/hashcode"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	// "github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsAppmeshVirtualNode() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppmeshVirtualNodeCreate,
		Read:   resourceAwsAppmeshVirtualNodeRead,
		Update: resourceAwsAppmeshVirtualNodeUpdate,
		Delete: resourceAwsAppmeshVirtualNodeDelete,
		Importer: &schema.ResourceImporter{
			State: resourceAwsAppmeshVirtualNodeImport,
		},

		SchemaVersion: 1,
		MigrateState:  resourceAwsAppmeshVirtualNodeMigrateState,

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
				ValidateFunc: validateAwsAccountId,
			},

			"spec": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backends": {
							Type:     schema.TypeSet,
							Removed:  "Use `backend` configuration blocks instead",
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},

						"backend": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 0,
							MaxItems: 25,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"virtual_service": {
										Type:     schema.TypeList,
										Optional: true,
										MinItems: 0,
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
							Set: appmeshVirtualNodeBackendHash,
						},

						"listener": {
							Type:     schema.TypeSet,
							Optional: true,
							MinItems: 0,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
													ValidateFunc: validation.IntBetween(1, 65535),
												},

												"protocol": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														appmesh.PortProtocolHttp,
														appmesh.PortProtocolTcp,
													}, false),
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
													ValidateFunc: validation.IntBetween(1, 65535),
												},

												"protocol": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														appmesh.PortProtocolHttp,
														appmesh.PortProtocolTcp,
													}, false),
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
																			ValidateFunc: validateArn,
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

															// ForbiddenException: TLS Certificates from SDS are not supported.
															// "sds": {
															// 	Type:     schema.TypeList,
															// 	Optional: true,
															// 	MinItems: 0,
															// 	MaxItems: 1,
															// 	Elem: &schema.Resource{
															// 		Schema: map[string]*schema.Schema{
															// 			"secret_name": {
															// 				Type:     schema.TypeString,
															// 				Required: true,
															// 			},

															// 			"source": {
															// 				Type:     schema.TypeList,
															// 				Required: true,
															// 				MinItems: 1,
															// 				MaxItems: 1,
															// 				Elem: &schema.Resource{
															// 					Schema: map[string]*schema.Schema{
															// 						"unix_domain_socket": {
															// 							Type:     schema.TypeList,
															// 							Required: true,
															// 							MinItems: 1,
															// 							MaxItems: 1,
															// 							Elem: &schema.Resource{
															// 								Schema: map[string]*schema.Schema{
															// 									"path": {
															// 										Type:     schema.TypeString,
															// 										Required: true,
															// 									},
															// 								},
															// 							},
															// 						},
															// 					},
															// 				},
															// 			},
															// 		},
															// 	},
															// },
														},
													},
												},

												"mode": {
													Type:     schema.TypeString,
													Required: true,
													ValidateFunc: validation.StringInSlice([]string{
														appmesh.ListenerTlsModeDisabled,
														appmesh.ListenerTlsModePermissive,
														appmesh.ListenerTlsModeStrict,
													}, false),
												},
											},
										},
									},
								},
							},
							Set: appmeshVirtualNodeListenerHash,
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
										MinItems:      0,
										MaxItems:      1,
										ConflictsWith: []string{"spec.0.service_discovery.0.aws_cloud_map"},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"service_name": {
													Type:     schema.TypeString,
													Removed:  "Use `hostname` argument instead",
													Optional: true,
													Computed: true,
												},

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

			// "tags": tagsSchema(),
		},
	}
}

func resourceAwsAppmeshVirtualNodeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	req := &appmesh.CreateVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
		Spec:            expandAppmeshVirtualNodeSpec(d.Get("spec").([]interface{})),
		// Tags:            keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().AppmeshTags(),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating App Mesh virtual node: %#v", req)
	resp, err := conn.CreateVirtualNode(req)
	if err != nil {
		return fmt.Errorf("error creating App Mesh virtual node: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualNode.Metadata.Uid))

	return resourceAwsAppmeshVirtualNodeRead(d, meta)
}

func resourceAwsAppmeshVirtualNodeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	req := &appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("mesh_owner"); ok {
		req.MeshOwner = aws.String(v.(string))
	}

	resp, err := conn.DescribeVirtualNode(req)
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh virtual node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading App Mesh virtual node: %s", err)
	}
	if aws.StringValue(resp.VirtualNode.Status.Status) == appmesh.VirtualNodeStatusCodeDeleted {
		log.Printf("[WARN] App Mesh virtual node (%s) not found, removing from state", d.Id())
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
	err = d.Set("spec", flattenAppmeshVirtualNodeSpec(resp.VirtualNode.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %s", err)
	}

	// tags, err := keyvaluetags.AppmeshListTags(conn, arn)

	// if err != nil {
	// 	return fmt.Errorf("error listing tags for App Mesh virtual node (%s): %s", arn, err)
	// }

	// if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
	// 	return fmt.Errorf("error setting tags: %s", err)
	// }

	return nil
}

func resourceAwsAppmeshVirtualNodeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateVirtualNodeInput{
			MeshName:        aws.String(d.Get("mesh_name").(string)),
			MeshOwner:       aws.String(d.Get("mesh_owner").(string)),
			VirtualNodeName: aws.String(d.Get("name").(string)),
			Spec:            expandAppmeshVirtualNodeSpec(v.([]interface{})),
		}

		log.Printf("[DEBUG] Updating App Mesh virtual node: %#v", req)
		_, err := conn.UpdateVirtualNode(req)
		if err != nil {
			return fmt.Errorf("error updating App Mesh virtual node: %s", err)
		}
	}

	// arn := d.Get("arn").(string)
	// if d.HasChange("tags") {
	// 	o, n := d.GetChange("tags")

	// 	if err := keyvaluetags.AppmeshUpdateTags(conn, arn, o, n); err != nil {
	// 		return fmt.Errorf("error updating App Mesh virtual node (%s) tags: %s", arn, err)
	// 	}
	// }

	return resourceAwsAppmeshVirtualNodeRead(d, meta)
}

func resourceAwsAppmeshVirtualNodeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshconn

	log.Printf("[DEBUG] Deleting App Mesh virtual node: %s", d.Id())
	_, err := conn.DeleteVirtualNode(&appmesh.DeleteVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting App Mesh virtual node: %s", err)
	}

	return nil
}

func resourceAwsAppmeshVirtualNodeImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	parts := strings.Split(d.Id(), "/")
	if len(parts) != 2 {
		return []*schema.ResourceData{}, fmt.Errorf("Wrong format of resource: %s. Please follow 'mesh-name/virtual-node-name'", d.Id())
	}

	mesh := parts[0]
	name := parts[1]
	log.Printf("[DEBUG] Importing App Mesh virtual node %s from mesh %s", name, mesh)

	conn := meta.(*AWSClient).appmeshconn

	resp, err := conn.DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(mesh),
		VirtualNodeName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.VirtualNode.Metadata.Uid))
	d.Set("name", resp.VirtualNode.VirtualNodeName)
	d.Set("mesh_name", resp.VirtualNode.MeshName)

	return []*schema.ResourceData{d}, nil
}

func appmeshVirtualNodeBackendHash(vBackend interface{}) int {
	var buf bytes.Buffer
	mBackend := vBackend.(map[string]interface{})
	if vVirtualService, ok := mBackend["virtual_service"].([]interface{}); ok && len(vVirtualService) > 0 && vVirtualService[0] != nil {
		mVirtualService := vVirtualService[0].(map[string]interface{})
		if v, ok := mVirtualService["virtual_service_name"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	return hashcode.String(buf.String())
}

func appmeshVirtualNodeListenerHash(vListener interface{}) int {
	var buf bytes.Buffer
	mListener := vListener.(map[string]interface{})
	if vHealthCheck, ok := mListener["health_check"].([]interface{}); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
		mHealthCheck := vHealthCheck[0].(map[string]interface{})
		if v, ok := mHealthCheck["healthy_threshold"].(int); ok {
			buf.WriteString(fmt.Sprintf("%d-", v))
		}
		if v, ok := mHealthCheck["interval_millis"].(int); ok {
			buf.WriteString(fmt.Sprintf("%d-", v))
		}
		if v, ok := mHealthCheck["path"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
		// Don't include "port" in the hash as it's Optional/Computed.
		// If present it must match the "port_mapping.port" value, so changes will be detected.
		if v, ok := mHealthCheck["protocol"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
		if v, ok := mHealthCheck["timeout_millis"].(int); ok {
			buf.WriteString(fmt.Sprintf("%d-", v))
		}
		if v, ok := mHealthCheck["unhealthy_threshold"].(int); ok {
			buf.WriteString(fmt.Sprintf("%d-", v))
		}
	}
	if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
		mPortMapping := vPortMapping[0].(map[string]interface{})
		if v, ok := mPortMapping["port"].(int); ok {
			buf.WriteString(fmt.Sprintf("%d-", v))
		}
		if v, ok := mPortMapping["protocol"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	if vTls, ok := mListener["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
		mTls := vTls[0].(map[string]interface{})
		if v, ok := mTls["mode"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
		if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
			mCertificate := vCertificate[0].(map[string]interface{})
			if vAcm, ok := mCertificate["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
				mAcm := vAcm[0].(map[string]interface{})
				if v, ok := mAcm["certificate_arn"].(string); ok {
					buf.WriteString(fmt.Sprintf("%s-", v))
				}
			}
			if vFile, ok := mCertificate["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				mFile := vFile[0].(map[string]interface{})
				if v, ok := mFile["certificate_chain"].(string); ok {
					buf.WriteString(fmt.Sprintf("%s-", v))
				}
				if v, ok := mFile["private_key"].(string); ok {
					buf.WriteString(fmt.Sprintf("%s-", v))
				}
			}
			// ForbiddenException: TLS Certificates from SDS are not supported.
			// if vSds, ok := mCertificate["sds"].([]interface{}); ok && len(vSds) > 0 && vSds[0] != nil {
			// 	mSds := vSds[0].(map[string]interface{})
			// 	if v, ok := mSds["secret_name"].(string); ok {
			// 		buf.WriteString(fmt.Sprintf("%s-", v))
			// 	}
			// 	if vSource, ok := mSds["source"].([]interface{}); ok && len(vSource) > 0 && vSource[0] != nil {
			// 		mSource := vSource[0].(map[string]interface{})
			// 		if vUnixDomainSocket, ok := mSource["unix_domain_socket"].([]interface{}); ok && len(vUnixDomainSocket) > 0 && vUnixDomainSocket[0] != nil {
			// 			mUnixDomainSocket := vUnixDomainSocket[0].(map[string]interface{})
			// 			if v, ok := mUnixDomainSocket["path"].(string); ok {
			// 				buf.WriteString(fmt.Sprintf("%s-", v))
			// 			}
			// 		}
			// 	}
			// }
		}
	}
	return hashcode.String(buf.String())
}
