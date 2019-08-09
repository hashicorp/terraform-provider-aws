package aws

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	appmesh "github.com/aws/aws-sdk-go/service/appmeshpreview"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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

			"name": {
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
							Set: appmeshBackendHash,
						},

						"backends": {
							Type:     schema.TypeSet,
							Removed:  "Use `backend` configuration blocks instead",
							Optional: true,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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
							Set: appmeshListenerHash,
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
													ValidateFunc: validateServiceDiscoveryHttpNamespaceName,
												},

												"service_name": {
													Type:     schema.TypeString,
													Required: true,
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

												"service_name": {
													Type:     schema.TypeString,
													Removed:  "Use `hostname` argument instead",
													Optional: true,
													Computed: true,
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

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsAppmeshVirtualNodeCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshpreviewconn

	req := &appmesh.CreateVirtualNodeInput{
		MeshName: aws.String(d.Get("mesh_name").(string)),
		Spec:     expandAppmeshVirtualNodeSpec(d.Get("spec").([]interface{})),
		// Tags:            tagsFromMapAppmesh(d.Get("tags").(map[string]interface{})),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Creating App Mesh virtual node: %s", req)
	resp, err := conn.CreateVirtualNode(req)
	if err != nil {
		return fmt.Errorf("error creating App Mesh virtual node: %s", err)
	}

	d.SetId(aws.StringValue(resp.VirtualNode.Metadata.Uid))

	return resourceAwsAppmeshVirtualNodeRead(d, meta)
}

func resourceAwsAppmeshVirtualNodeRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshpreviewconn

	resp, err := conn.DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] App Mesh virtual node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading App Mesh virtual node (%s): %s", d.Id(), err)
	}
	if aws.StringValue(resp.VirtualNode.Status.Status) == appmesh.VirtualNodeStatusCodeDeleted {
		log.Printf("[WARN] App Mesh virtual node (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("arn", resp.VirtualNode.Metadata.Arn)
	d.Set("created_date", resp.VirtualNode.Metadata.CreatedAt.Format(time.RFC3339))
	d.Set("last_updated_date", resp.VirtualNode.Metadata.LastUpdatedAt.Format(time.RFC3339))
	d.Set("mesh_name", resp.VirtualNode.MeshName)
	d.Set("name", resp.VirtualNode.VirtualNodeName)
	err = d.Set("spec", flattenAppmeshVirtualNodeSpec(resp.VirtualNode.Spec))
	if err != nil {
		return fmt.Errorf("error setting spec: %s", err)
	}

	// err = saveTagsAppmesh(conn, d, aws.StringValue(resp.VirtualNode.Metadata.Arn))
	// if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
	// 	log.Printf("[WARN] App Mesh virtual node (%s) not found, removing from state", d.Id())
	// 	d.SetId("")
	// 	return nil
	// }
	// if err != nil {
	// 	return fmt.Errorf("error saving tags: %s", err)
	// }

	return nil
}

func resourceAwsAppmeshVirtualNodeUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshpreviewconn

	if d.HasChange("spec") {
		_, v := d.GetChange("spec")
		req := &appmesh.UpdateVirtualNodeInput{
			MeshName:        aws.String(d.Get("mesh_name").(string)),
			Spec:            expandAppmeshVirtualNodeSpec(v.([]interface{})),
			VirtualNodeName: aws.String(d.Get("name").(string)),
		}

		log.Printf("[DEBUG] Updating App Mesh virtual node: %s", req)
		_, err := conn.UpdateVirtualNode(req)
		if err != nil {
			return fmt.Errorf("error updating App Mesh virtual node (%s): %s", d.Id(), err)
		}
	}

	// err := setTagsAppmesh(conn, d, d.Get("arn").(string))
	// if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
	// 	log.Printf("[WARN] App Mesh virtual node (%s) not found, removing from state", d.Id())
	// 	d.SetId("")
	// 	return nil
	// }
	// if err != nil {
	// 	return fmt.Errorf("error setting tags: %s", err)
	// }

	return resourceAwsAppmeshVirtualNodeRead(d, meta)
}

func resourceAwsAppmeshVirtualNodeDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appmeshpreviewconn

	log.Printf("[DEBUG] Deleting App Mesh virtual node: %s", d.Id())
	_, err := conn.DeleteVirtualNode(&appmesh.DeleteVirtualNodeInput{
		MeshName:        aws.String(d.Get("mesh_name").(string)),
		VirtualNodeName: aws.String(d.Get("name").(string)),
	})
	if isAWSErr(err, appmesh.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting App Mesh virtual node (%s): %s", d.Id(), err)
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

	conn := meta.(*AWSClient).appmeshpreviewconn

	resp, err := conn.DescribeVirtualNode(&appmesh.DescribeVirtualNodeInput{
		MeshName:        aws.String(mesh),
		VirtualNodeName: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	d.SetId(aws.StringValue(resp.VirtualNode.Metadata.Uid))
	d.Set("mesh_name", resp.VirtualNode.MeshName)
	d.Set("name", resp.VirtualNode.VirtualNodeName)

	return []*schema.ResourceData{d}, nil
}

func expandAppmeshVirtualNodeSpec(vSpec []interface{}) *appmesh.VirtualNodeSpec {
	spec := &appmesh.VirtualNodeSpec{}

	if len(vSpec) == 0 || vSpec[0] == nil {
		// Empty Spec is allowed.
		return spec
	}
	mSpec := vSpec[0].(map[string]interface{})

	if vBackends, ok := mSpec["backend"].(*schema.Set); ok && vBackends.Len() > 0 {
		backends := []*appmesh.Backend{}

		for _, vBackend := range vBackends.List() {
			backend := &appmesh.Backend{}

			mBackend := vBackend.(map[string]interface{})

			if vVirtualService, ok := mBackend["virtual_service"].([]interface{}); ok && len(vVirtualService) > 0 && vVirtualService[0] != nil {
				mVirtualService := vVirtualService[0].(map[string]interface{})

				backend.VirtualService = &appmesh.VirtualServiceBackend{}

				if vVirtualServiceName, ok := mVirtualService["virtual_service_name"].(string); ok && vVirtualServiceName != "" {
					backend.VirtualService.VirtualServiceName = aws.String(vVirtualServiceName)
				}
			}

			backends = append(backends, backend)
		}

		spec.Backends = backends
	}

	if vListeners, ok := mSpec["listener"].(*schema.Set); ok && vListeners.Len() > 0 {
		listeners := []*appmesh.Listener{}

		for _, vListener := range vListeners.List() {
			listener := &appmesh.Listener{}

			mListener := vListener.(map[string]interface{})

			if vHealthCheck, ok := mListener["health_check"].([]interface{}); ok && len(vHealthCheck) > 0 && vHealthCheck[0] != nil {
				mHealthCheck := vHealthCheck[0].(map[string]interface{})

				healthCheck := &appmesh.HealthCheckPolicy{}

				if vHealthyThreshold, ok := mHealthCheck["healthy_threshold"].(int); ok && vHealthyThreshold > 0 {
					healthCheck.HealthyThreshold = aws.Int64(int64(vHealthyThreshold))
				}
				if vIntervalMillis, ok := mHealthCheck["interval_millis"].(int); ok && vIntervalMillis > 0 {
					healthCheck.IntervalMillis = aws.Int64(int64(vIntervalMillis))
				}
				if vPath, ok := mHealthCheck["path"].(string); ok && vPath != "" {
					healthCheck.Path = aws.String(vPath)
				}
				if vPort, ok := mHealthCheck["port"].(int); ok && vPort > 0 {
					healthCheck.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mHealthCheck["protocol"].(string); ok && vProtocol != "" {
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

			if vPortMapping, ok := mListener["port_mapping"].([]interface{}); ok && len(vPortMapping) > 0 && vPortMapping[0] != nil {
				mPortMapping := vPortMapping[0].(map[string]interface{})

				portMapping := &appmesh.PortMapping{}

				if vPort, ok := mPortMapping["port"].(int); ok && vPort > 0 {
					portMapping.Port = aws.Int64(int64(vPort))
				}
				if vProtocol, ok := mPortMapping["protocol"].(string); ok && vProtocol != "" {
					portMapping.Protocol = aws.String(vProtocol)
				}

				listener.PortMapping = portMapping
			}

			if vTls, ok := mListener["tls"].([]interface{}); ok && len(vTls) > 0 && vTls[0] != nil {
				mTls := vTls[0].(map[string]interface{})

				tls := &appmesh.ListenerTls{}

				if vMode, ok := mTls["mode"].(string); ok && vMode != "" {
					tls.Mode = aws.String(vMode)
				}

				if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
					mCertificate := vCertificate[0].(map[string]interface{})

					if vAcm, ok := mCertificate["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
						mAcm := vAcm[0].(map[string]interface{})

						if vCertificateArn, ok := mAcm["certificate_arn"].(string); ok && vCertificateArn != "" {
							tls.Certificate = &appmesh.ListenerTlsCertificate{
								Acm: &appmesh.ListenerTlsAcmCertificate{
									CertificateArn: aws.String(vCertificateArn),
								},
							}
						}
					}
				}

				listener.Tls = tls
			}

			listeners = append(listeners, listener)
		}

		spec.Listeners = listeners
	}

	if vLogging, ok := mSpec["logging"].([]interface{}); ok && len(vLogging) > 0 && vLogging[0] != nil {
		mLogging := vLogging[0].(map[string]interface{})

		if vAccessLog, ok := mLogging["access_log"].([]interface{}); ok && len(vAccessLog) > 0 && vAccessLog[0] != nil {
			mAccessLog := vAccessLog[0].(map[string]interface{})

			if vFile, ok := mAccessLog["file"].([]interface{}); ok && len(vFile) > 0 && vFile[0] != nil {
				mFile := vFile[0].(map[string]interface{})

				if vPath, ok := mFile["path"].(string); ok && vPath != "" {
					spec.Logging = &appmesh.Logging{
						AccessLog: &appmesh.AccessLog{
							File: &appmesh.FileAccessLog{
								Path: aws.String(vPath),
							},
						},
					}
				}
			}
		}
	}

	if vServiceDiscovery, ok := mSpec["service_discovery"].([]interface{}); ok && len(vServiceDiscovery) > 0 && vServiceDiscovery[0] != nil {
		spec.ServiceDiscovery = &appmesh.ServiceDiscovery{}

		mServiceDiscovery := vServiceDiscovery[0].(map[string]interface{})

		if vAwsCloudMap, ok := mServiceDiscovery["aws_cloud_map"].([]interface{}); ok && len(vAwsCloudMap) > 0 && vAwsCloudMap[0] != nil {
			awsCloudMap := &appmesh.AwsCloudMapServiceDiscovery{}

			mAwsCloudMap := vAwsCloudMap[0].(map[string]interface{})

			if vNamespaceName, ok := mAwsCloudMap["namespace_name"].(string); ok && vNamespaceName != "" {
				awsCloudMap.NamespaceName = aws.String(vNamespaceName)
			}
			if vServiceName, ok := mAwsCloudMap["service_name"].(string); ok && vServiceName != "" {
				awsCloudMap.ServiceName = aws.String(vServiceName)
			}

			if vAttributes, ok := mAwsCloudMap["attributes"].(map[string]interface{}); ok && len(vAttributes) > 0 {
				attributes := []*appmesh.AwsCloudMapInstanceAttribute{}

				for k, v := range vAttributes {
					attributes = append(attributes, &appmesh.AwsCloudMapInstanceAttribute{
						Key:   aws.String(k),
						Value: aws.String(v.(string)),
					})
				}

				awsCloudMap.Attributes = attributes
			}

			spec.ServiceDiscovery.AwsCloudMap = awsCloudMap
		}

		if vDns, ok := mServiceDiscovery["dns"].([]interface{}); ok && len(vDns) > 0 && vDns[0] != nil {
			mDns := vDns[0].(map[string]interface{})

			if vHostname, ok := mDns["hostname"].(string); ok && vHostname != "" {
				spec.ServiceDiscovery.Dns = &appmesh.DnsServiceDiscovery{
					Hostname: aws.String(vHostname),
				}
			}
		}
	}

	return spec
}

func flattenAppmeshVirtualNodeSpec(spec *appmesh.VirtualNodeSpec) []interface{} {
	if spec == nil {
		return []interface{}{}
	}

	mSpec := map[string]interface{}{}

	if backends := spec.Backends; backends != nil {
		vBackends := []interface{}{}

		for _, backend := range backends {
			mBackend := map[string]interface{}{}

			if virtualService := backend.VirtualService; virtualService != nil {
				mBackend["virtual_service"] = []interface{}{
					map[string]interface{}{
						"virtual_service_name": aws.StringValue(virtualService.VirtualServiceName),
					},
				}
			}

			vBackends = append(vBackends, mBackend)
		}

		mSpec["backend"] = schema.NewSet(appmeshBackendHash, vBackends)
	}

	if listeners := spec.Listeners; listeners != nil {
		vListeners := []interface{}{}

		for _, listener := range listeners {
			mListener := map[string]interface{}{}

			if healthCheck := listener.HealthCheck; healthCheck != nil {
				mListener["health_check"] = []interface{}{
					map[string]interface{}{
						"healthy_threshold":   int(aws.Int64Value(healthCheck.HealthyThreshold)),
						"interval_millis":     int(aws.Int64Value(healthCheck.IntervalMillis)),
						"path":                aws.StringValue(healthCheck.Path),
						"port":                int(aws.Int64Value(healthCheck.Port)),
						"protocol":            aws.StringValue(healthCheck.Protocol),
						"timeout_millis":      int(aws.Int64Value(healthCheck.TimeoutMillis)),
						"unhealthy_threshold": int(aws.Int64Value(healthCheck.UnhealthyThreshold)),
					},
				}
			}

			if portMapping := listener.PortMapping; portMapping != nil {
				mListener["port_mapping"] = []interface{}{
					map[string]interface{}{
						"port":     int(aws.Int64Value(portMapping.Port)),
						"protocol": aws.StringValue(portMapping.Protocol),
					},
				}
			}

			if tls := listener.Tls; tls != nil {
				mTls := map[string]interface{}{
					"mode": aws.StringValue(tls.Mode),
				}

				if certificate := tls.Certificate; certificate != nil {
					if acm := certificate.Acm; acm != nil {
						mTls["certificate"] = []interface{}{
							map[string]interface{}{
								"acm": []interface{}{
									map[string]interface{}{
										"certificate_arn": aws.StringValue(acm.CertificateArn),
									},
								},
							},
						}
					}
				}

				mListener["tls"] = []interface{}{mTls}
			}

			vListeners = append(vListeners, mListener)
		}

		mSpec["listener"] = schema.NewSet(appmeshListenerHash, vListeners)
	}

	if logging := spec.Logging; logging != nil {
		if accessLog := logging.AccessLog; accessLog != nil {
			if file := accessLog.File; file != nil {
				mSpec["logging"] = []interface{}{
					map[string]interface{}{
						"access_log": []interface{}{
							map[string]interface{}{
								"file": []interface{}{
									map[string]interface{}{
										"path": aws.StringValue(file.Path),
									},
								},
							},
						},
					},
				}
			}
		}
	}

	if serviceDiscovery := spec.ServiceDiscovery; serviceDiscovery != nil {
		mServiceDiscovery := map[string]interface{}{}

		if awsCloudMap := serviceDiscovery.AwsCloudMap; awsCloudMap != nil {
			vAttributes := map[string]interface{}{}

			for _, attribute := range awsCloudMap.Attributes {
				vAttributes[aws.StringValue(attribute.Key)] = aws.StringValue(attribute.Value)
			}

			mServiceDiscovery["aws_cloud_map"] = []interface{}{
				map[string]interface{}{
					"attributes":     vAttributes,
					"namespace_name": aws.StringValue(awsCloudMap.NamespaceName),
					"service_name":   aws.StringValue(awsCloudMap.ServiceName),
				},
			}
		}

		if dns := serviceDiscovery.Dns; dns != nil {
			mServiceDiscovery["dns"] = []interface{}{
				map[string]interface{}{
					"hostname": aws.StringValue(dns.Hostname),
				},
			}
		}

		mSpec["service_discovery"] = []interface{}{mServiceDiscovery}
	}

	return []interface{}{mSpec}
}

func appmeshBackendHash(vBackend interface{}) int {
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

func appmeshListenerHash(vListener interface{}) int {
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
		if vCertificate, ok := mTls["certificate"].([]interface{}); ok && len(vCertificate) > 0 && vCertificate[0] != nil {
			mCertificate := vCertificate[0].(map[string]interface{})
			if vAcm, ok := mCertificate["acm"].([]interface{}); ok && len(vAcm) > 0 && vAcm[0] != nil {
				mAcm := vAcm[0].(map[string]interface{})
				if v, ok := mAcm["certificate_arn"].(string); ok {
					buf.WriteString(fmt.Sprintf("%s-", v))
				}
			}
		}
		if v, ok := mTls["mode"].(string); ok {
			buf.WriteString(fmt.Sprintf("%s-", v))
		}
	}
	return hashcode.String(buf.String())
}
