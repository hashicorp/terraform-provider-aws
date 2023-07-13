// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apprunner_service", name="Service")
// @Tags(identifierAttribute="arn")
func ResourceService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceCreate,
		ReadWithoutTimeout:   resourceServiceRead,
		UpdateWithoutTimeout: resourceServiceUpdate,
		DeleteWithoutTimeout: resourceServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"auto_scaling_configuration_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},

			"encryption_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kms_key": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},

			"health_check_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"healthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"interval": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"path": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "/",
							ForceNew:     true,
							ValidateFunc: validation.StringLenBetween(0, 51200),
						},
						"protocol": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      apprunner.HealthCheckProtocolTcp,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(apprunner.HealthCheckProtocol_Values(), false),
						},
						"timeout": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      2,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"unhealthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 20),
						},
					},
				},
			},

			"instance_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "1024",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`256|512|1024|2048|4096|(0.25|0.5|1|2|4) vCPU`), ""),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// App Runner API always returns the amount in multiples of 1024 units
								return (old == "256" && new == "0.25 vCPU") || (old == "512" && new == "0.5 vCPU") || (old == "1024" && new == "1 vCPU") || (old == "2048" && new == "2 vCPU") || (old == "4096" && new == "4 vCPU")
							},
						},
						"instance_role_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"memory": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "2048",
							ValidateFunc: validation.StringMatch(regexp.MustCompile(`512|1024|2048|3072|4096|6144|8192|10240|12288|(0.5|1|2|3|4|6|8|10|12) GB`), ""),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// App Runner API always returns the amount in MB
								return (old == "512" && new == "0.5 GB") || (old == "1024" && new == "1 GB") || (old == "2048" && new == "2 GB") || (old == "3072" && new == "3 GB") || (old == "4096" && new == "4 GB") || (old == "6144" && new == "6 GB") || (old == "8192" && new == "8 GB") || (old == "10240" && new == "10 GB") || (old == "12288" && new == "12 GB")
							},
						},
					},
				},
			},

			"network_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ingress_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"is_publicly_accessible": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"egress_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"egress_type": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringInSlice(apprunner.EgressType_Values(), false),
									},
									"vpc_connector_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},

			"observability_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"observability_configuration_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"observability_enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
					},
				},
			},

			"service_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"service_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"service_url": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"source_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"authentication_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"access_role_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
									"connection_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"auto_deployments_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"code_repository": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"code_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"code_configuration_values": {
													Type:     schema.TypeList,
													Optional: true,
													MaxItems: 1,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"build_command": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 51200),
															},
															"port": {
																Type:         schema.TypeString,
																Optional:     true,
																Default:      "8080",
																ValidateFunc: validation.StringLenBetween(0, 51200),
															},
															"runtime": {
																Type:         schema.TypeString,
																Required:     true,
																ValidateFunc: validation.StringInSlice(apprunner.Runtime_Values(), false),
															},
															"runtime_environment_secrets": {
																Type:     schema.TypeMap,
																Optional: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringLenBetween(0, 2048),
																},
															},
															"runtime_environment_variables": {
																Type:     schema.TypeMap,
																Optional: true,
																Elem: &schema.Schema{
																	Type:         schema.TypeString,
																	ValidateFunc: validation.StringLenBetween(0, 51200),
																},
															},
															"start_command": {
																Type:         schema.TypeString,
																Optional:     true,
																ValidateFunc: validation.StringLenBetween(0, 51200),
															},
														},
													},
												},
												"configuration_source": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(apprunner.ConfigurationSource_Values(), false),
												},
											},
										},
									},
									"repository_url": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringLenBetween(0, 51200),
									},
									"source_code_version": {
										Type:     schema.TypeList,
										Required: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"type": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringInSlice(apprunner.SourceCodeVersionType_Values(), false),
												},
												"value": {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 51200),
												},
											},
										},
									},
								},
							},
							ExactlyOneOf: []string{"source_configuration.0.code_repository", "source_configuration.0.image_repository"},
						},
						"image_repository": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"image_configuration": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"port": {
													Type:         schema.TypeString,
													Optional:     true,
													Default:      "8080",
													ValidateFunc: validation.StringLenBetween(0, 51200),
												},
												"runtime_environment_secrets": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringLenBetween(0, 2048),
													},
												},
												"runtime_environment_variables": {
													Type:     schema.TypeMap,
													Optional: true,
													Elem: &schema.Schema{
														Type:         schema.TypeString,
														ValidateFunc: validation.StringLenBetween(0, 51200),
													},
												},
												"start_command": {
													Type:         schema.TypeString,
													Optional:     true,
													ValidateFunc: validation.StringLenBetween(0, 51200),
												},
											},
										},
									},
									"image_identifier": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringMatch(regexp.MustCompile(`([0-9]{12}\.dkr\.ecr\.[a-z\-]+-[0-9]{1}\.amazonaws\.com\/.*)|(^public\.ecr\.aws\/.+\/.+)`), ""),
									},
									"image_repository_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(apprunner.ImageRepositoryType_Values(), false),
									},
								},
							},
							ExactlyOneOf: []string{"source_configuration.0.image_repository", "source_configuration.0.code_repository"},
						},
					},
				},
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceServiceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	serviceName := d.Get("service_name").(string)
	input := &apprunner.CreateServiceInput{
		ServiceName:         aws.String(serviceName),
		SourceConfiguration: expandServiceSourceConfiguration(d.Get("source_configuration").([]interface{})),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk("auto_scaling_configuration_arn"); ok {
		input.AutoScalingConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("encryption_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EncryptionConfiguration = expandServiceEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("health_check_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.HealthCheckConfiguration = expandServiceHealthCheckConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("instance_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.InstanceConfiguration = expandServiceInstanceConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("network_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.NetworkConfiguration = expandNetworkConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("observability_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ObservabilityConfiguration = expandServiceObservabilityConfiguration(v.([]interface{}))
	}

	var output *apprunner.CreateServiceOutput

	err := retry.RetryContext(ctx, propagationTimeout, func() *retry.RetryError {
		var err error
		output, err = conn.CreateServiceWithContext(ctx, input)

		if tfawserr.ErrMessageContains(err, apprunner.ErrCodeInvalidRequestException, "Error in assuming instance role") {
			return retry.RetryableError(err)
		}

		if err != nil {
			return retry.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateServiceWithContext(ctx, input)
	}

	if err != nil {
		return diag.Errorf("creating App Runner Service (%s): %s", serviceName, err)
	}

	if output == nil || output.Service == nil {
		return diag.Errorf("creating App Runner Service (%s): empty output", serviceName)
	}

	d.SetId(aws.StringValue(output.Service.ServiceArn))

	if err := WaitServiceCreated(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for App Runner Service (%s) creation: %s", d.Id(), err)
	}

	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	input := &apprunner.DescribeServiceInput{
		ServiceArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeServiceWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] App Runner Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading App Runner Service (%s): %s", d.Id(), err)
	}

	if output == nil || output.Service == nil {
		return diag.Errorf("reading App Runner Service (%s): empty output", d.Id())
	}

	if aws.StringValue(output.Service.Status) == apprunner.ServiceStatusDeleted {
		if d.IsNewResource() {
			return diag.Errorf("reading App Runner Service (%s): %s after creation", d.Id(), aws.StringValue(output.Service.Status))
		}
		log.Printf("[WARN] App Runner Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	service := output.Service
	arn := aws.StringValue(service.ServiceArn)

	var autoScalingConfigArn string
	if service.AutoScalingConfigurationSummary != nil {
		autoScalingConfigArn = aws.StringValue(service.AutoScalingConfigurationSummary.AutoScalingConfigurationArn)
	}

	d.Set("arn", arn)
	d.Set("auto_scaling_configuration_arn", autoScalingConfigArn)
	d.Set("service_id", service.ServiceId)
	d.Set("service_name", service.ServiceName)
	d.Set("service_url", service.ServiceUrl)
	d.Set("status", service.Status)
	if err := d.Set("encryption_configuration", flattenServiceEncryptionConfiguration(service.EncryptionConfiguration)); err != nil {
		return diag.Errorf("setting encryption_configuration: %s", err)
	}

	if err := d.Set("health_check_configuration", flattenServiceHealthCheckConfiguration(service.HealthCheckConfiguration)); err != nil {
		return diag.Errorf("setting health_check_configuration: %s", err)
	}

	if err := d.Set("instance_configuration", flattenServiceInstanceConfiguration(service.InstanceConfiguration)); err != nil {
		return diag.Errorf("setting instance_configuration: %s", err)
	}

	if err := d.Set("network_configuration", flattenNetworkConfiguration(service.NetworkConfiguration)); err != nil {
		return diag.Errorf("setting network_configuration: %s", err)
	}

	if err := d.Set("observability_configuration", flattenServiceObservabilityConfiguration(service.ObservabilityConfiguration)); err != nil {
		return diag.Errorf("setting observability_configuration: %s", err)
	}

	if err := d.Set("source_configuration", flattenServiceSourceConfiguration(service.SourceConfiguration)); err != nil {
		return diag.Errorf("setting source_configuration: %s", err)
	}

	return nil
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	if d.HasChanges(
		"auto_scaling_configuration_arn",
		"instance_configuration",
		"network_configuration",
		"observability_configuration",
		"source_configuration",
	) {
		input := &apprunner.UpdateServiceInput{
			ServiceArn: aws.String(d.Id()),
		}

		if d.HasChange("auto_scaling_configuration_arn") {
			input.AutoScalingConfigurationArn = aws.String(d.Get("auto_scaling_configuration_arn").(string))
		}

		if d.HasChange("instance_configuration") {
			input.InstanceConfiguration = expandServiceInstanceConfiguration(d.Get("instance_configuration").([]interface{}))
		}

		if d.HasChange("network_configuration") {
			input.NetworkConfiguration = expandNetworkConfiguration(d.Get("network_configuration").([]interface{}))
		}

		if d.HasChange("observability_configuration") {
			input.ObservabilityConfiguration = expandServiceObservabilityConfiguration(d.Get("observability_configuration").([]interface{}))
		}

		if d.HasChange("source_configuration") {
			input.SourceConfiguration = expandServiceSourceConfiguration(d.Get("source_configuration").([]interface{}))
		}

		_, err := conn.UpdateServiceWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating App Runner Service (%s): %s", d.Id(), err)
		}

		if err := WaitServiceUpdated(ctx, conn, d.Id()); err != nil {
			return diag.Errorf("waiting for App Runner Service (%s) to update: %s", d.Id(), err)
		}
	}

	return resourceServiceRead(ctx, d, meta)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	input := &apprunner.DeleteServiceInput{
		ServiceArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteServiceWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting App Runner Service (%s): %s", d.Id(), err)
	}

	if err := WaitServiceDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}

		return diag.Errorf("waiting for App Runner Service (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func expandServiceEncryptionConfiguration(l []interface{}) *apprunner.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.EncryptionConfiguration{}

	if v, ok := tfMap["kms_key"].(string); ok && v != "" {
		result.KmsKey = aws.String(v)
	}

	return result
}

func expandServiceHealthCheckConfiguration(l []interface{}) *apprunner.HealthCheckConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.HealthCheckConfiguration{}

	if v, ok := tfMap["healthy_threshold"].(int); ok {
		result.HealthyThreshold = aws.Int64(int64(v))
	}

	if v, ok := tfMap["interval"].(int); ok {
		result.Interval = aws.Int64(int64(v))
	}

	if v, ok := tfMap["path"].(string); ok {
		result.Path = aws.String(v)
	}

	if v, ok := tfMap["protocol"].(string); ok {
		result.Protocol = aws.String(v)
	}

	if v, ok := tfMap["timeout"].(int); ok {
		result.Timeout = aws.Int64(int64(v))
	}

	if v, ok := tfMap["unhealthy_threshold"].(int); ok {
		result.UnhealthyThreshold = aws.Int64(int64(v))
	}

	return result
}

func expandServiceInstanceConfiguration(l []interface{}) *apprunner.InstanceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.InstanceConfiguration{}

	if v, ok := tfMap["cpu"].(string); ok {
		result.Cpu = aws.String(v)
	}

	if v, ok := tfMap["instance_role_arn"].(string); ok && v != "" {
		result.InstanceRoleArn = aws.String(v)
	}

	if v, ok := tfMap["memory"].(string); ok {
		result.Memory = aws.String(v)
	}

	return result
}

func expandNetworkConfiguration(l []interface{}) *apprunner.NetworkConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.NetworkConfiguration{}

	if v, ok := tfMap["ingress_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.IngressConfiguration = expandNetworkIngressConfiguration(v)
	}

	if v, ok := tfMap["egress_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.EgressConfiguration = expandNetworkEgressConfiguration(v)
	}

	return result
}

func expandServiceObservabilityConfiguration(l []interface{}) *apprunner.ServiceObservabilityConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.ServiceObservabilityConfiguration{}

	if v, ok := tfMap["observability_configuration_arn"].(string); ok && len(v) > 0 {
		result.ObservabilityConfigurationArn = aws.String(v)
	}

	if v, ok := tfMap["observability_enabled"].(bool); ok {
		result.ObservabilityEnabled = aws.Bool(v)
	}

	return result
}

func expandServiceSourceConfiguration(l []interface{}) *apprunner.SourceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.SourceConfiguration{}

	if v, ok := tfMap["authentication_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.AuthenticationConfiguration = expandServiceAuthenticationConfiguration(v)
	}

	if v, ok := tfMap["auto_deployments_enabled"].(bool); ok {
		result.AutoDeploymentsEnabled = aws.Bool(v)
	}

	if v, ok := tfMap["code_repository"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.CodeRepository = expandServiceCodeRepository(v)
	}

	if v, ok := tfMap["image_repository"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ImageRepository = expandServiceImageRepository(v)
	}

	return result
}

func expandServiceAuthenticationConfiguration(l []interface{}) *apprunner.AuthenticationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.AuthenticationConfiguration{}

	if v, ok := tfMap["access_role_arn"].(string); ok && v != "" {
		result.AccessRoleArn = aws.String(v)
	}

	if v, ok := tfMap["connection_arn"].(string); ok && v != "" {
		result.ConnectionArn = aws.String(v)
	}

	return result
}

func expandNetworkIngressConfiguration(l []interface{}) *apprunner.IngressConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.IngressConfiguration{}

	if v, ok := tfMap["is_publicly_accessible"].(bool); ok {
		result.IsPubliclyAccessible = aws.Bool(v)
	}

	return result
}

func expandNetworkEgressConfiguration(l []interface{}) *apprunner.EgressConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.EgressConfiguration{}

	if v, ok := tfMap["egress_type"].(string); ok {
		result.EgressType = aws.String(v)
	}

	if v, ok := tfMap["vpc_connector_arn"].(string); ok && v != "" {
		result.VpcConnectorArn = aws.String(v)
	}

	return result
}

func expandServiceImageConfiguration(l []interface{}) *apprunner.ImageConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.ImageConfiguration{}

	if v, ok := tfMap["port"].(string); ok && v != "" {
		result.Port = aws.String(v)
	}

	if v, ok := tfMap["runtime_environment_secrets"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentSecrets = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["runtime_environment_variables"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentVariables = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["start_command"].(string); ok && v != "" {
		result.StartCommand = aws.String(v)
	}

	return result
}

func expandServiceCodeRepository(l []interface{}) *apprunner.CodeRepository {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.CodeRepository{}

	if v, ok := tfMap["source_code_version"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.SourceCodeVersion = expandServiceSourceCodeVersion(v)
	}

	if v, ok := tfMap["code_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.CodeConfiguration = expandServiceCodeConfiguration(v)
	}

	if v, ok := tfMap["repository_url"].(string); ok && v != "" {
		result.RepositoryUrl = aws.String(v)
	}

	return result
}

func expandServiceImageRepository(l []interface{}) *apprunner.ImageRepository {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.ImageRepository{}

	if v, ok := tfMap["image_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ImageConfiguration = expandServiceImageConfiguration(v)
	}

	if v, ok := tfMap["image_identifier"].(string); ok && v != "" {
		result.ImageIdentifier = aws.String(v)
	}

	if v, ok := tfMap["image_repository_type"].(string); ok && v != "" {
		result.ImageRepositoryType = aws.String(v)
	}

	return result
}

func expandServiceCodeConfiguration(l []interface{}) *apprunner.CodeConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.CodeConfiguration{}

	if v, ok := tfMap["configuration_source"].(string); ok && v != "" {
		result.ConfigurationSource = aws.String(v)
	}

	if v, ok := tfMap["code_configuration_values"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.CodeConfigurationValues = expandServiceCodeConfigurationValues(v)
	}

	return result
}

func expandServiceCodeConfigurationValues(l []interface{}) *apprunner.CodeConfigurationValues {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.CodeConfigurationValues{}

	if v, ok := tfMap["build_command"].(string); ok && v != "" {
		result.BuildCommand = aws.String(v)
	}

	if v, ok := tfMap["port"].(string); ok && v != "" {
		result.Port = aws.String(v)
	}

	if v, ok := tfMap["runtime"].(string); ok && v != "" {
		result.Runtime = aws.String(v)
	}

	if v, ok := tfMap["runtime_environment_secrets"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentSecrets = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["runtime_environment_variables"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentVariables = flex.ExpandStringMap(v)
	}

	if v, ok := tfMap["start_command"].(string); ok && v != "" {
		result.StartCommand = aws.String(v)
	}

	return result
}

func expandServiceSourceCodeVersion(l []interface{}) *apprunner.SourceCodeVersion {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &apprunner.SourceCodeVersion{}

	if v, ok := tfMap["type"].(string); ok && v != "" {
		result.Type = aws.String(v)
	}

	if v, ok := tfMap["value"].(string); ok && v != "" {
		result.Value = aws.String(v)
	}

	return result
}

func flattenServiceEncryptionConfiguration(config *apprunner.EncryptionConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"kms_key": aws.StringValue(config.KmsKey),
	}

	return []interface{}{m}
}

func flattenServiceHealthCheckConfiguration(config *apprunner.HealthCheckConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"healthy_threshold":   aws.Int64Value(config.HealthyThreshold),
		"interval":            aws.Int64Value(config.Interval),
		"path":                aws.StringValue(config.Path),
		"protocol":            aws.StringValue(config.Protocol),
		"timeout":             aws.Int64Value(config.Timeout),
		"unhealthy_threshold": aws.Int64Value(config.UnhealthyThreshold),
	}

	return []interface{}{m}
}

func flattenServiceInstanceConfiguration(config *apprunner.InstanceConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cpu":               aws.StringValue(config.Cpu),
		"instance_role_arn": aws.StringValue(config.InstanceRoleArn),
		"memory":            aws.StringValue(config.Memory),
	}

	return []interface{}{m}
}

func flattenNetworkConfiguration(config *apprunner.NetworkConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"ingress_configuration": flattenNetworkIngressConfiguration(config.IngressConfiguration),
		"egress_configuration":  flattenNetworkEgressConfiguration(config.EgressConfiguration),
	}

	return []interface{}{m}
}

func flattenNetworkIngressConfiguration(config *apprunner.IngressConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"is_publicly_accessible": aws.BoolValue(config.IsPubliclyAccessible),
	}

	return []interface{}{m}
}

func flattenNetworkEgressConfiguration(config *apprunner.EgressConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"egress_type":       aws.StringValue(config.EgressType),
		"vpc_connector_arn": aws.StringValue(config.VpcConnectorArn),
	}

	return []interface{}{m}
}

func flattenServiceObservabilityConfiguration(config *apprunner.ServiceObservabilityConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"observability_configuration_arn": aws.StringValue(config.ObservabilityConfigurationArn),
		"observability_enabled":           aws.BoolValue(config.ObservabilityEnabled),
	}

	return []interface{}{m}
}

func flattenServiceCodeRepository(r *apprunner.CodeRepository) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"code_configuration":  flattenServiceCodeConfiguration(r.CodeConfiguration),
		"repository_url":      aws.StringValue(r.RepositoryUrl),
		"source_code_version": flattenServiceSourceCodeVersion(r.SourceCodeVersion),
	}

	return []interface{}{m}
}

func flattenServiceCodeConfiguration(config *apprunner.CodeConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"code_configuration_values": flattenServiceCodeConfigurationValues(config.CodeConfigurationValues),
		"configuration_source":      aws.StringValue(config.ConfigurationSource),
	}

	return []interface{}{m}
}

func flattenServiceCodeConfigurationValues(values *apprunner.CodeConfigurationValues) []interface{} {
	if values == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"build_command":                 aws.StringValue(values.BuildCommand),
		"port":                          aws.StringValue(values.Port),
		"runtime":                       aws.StringValue(values.Runtime),
		"runtime_environment_secrets":   aws.StringValueMap(values.RuntimeEnvironmentSecrets),
		"runtime_environment_variables": aws.StringValueMap(values.RuntimeEnvironmentVariables),
		"start_command":                 aws.StringValue(values.StartCommand),
	}

	return []interface{}{m}
}

func flattenServiceSourceCodeVersion(v *apprunner.SourceCodeVersion) []interface{} {
	if v == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"type":  aws.StringValue(v.Type),
		"value": aws.StringValue(v.Value),
	}

	return []interface{}{m}
}

func flattenServiceSourceConfiguration(config *apprunner.SourceConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authentication_configuration": flattenServiceAuthenticationConfiguration(config.AuthenticationConfiguration),
		"auto_deployments_enabled":     aws.BoolValue(config.AutoDeploymentsEnabled),
		"code_repository":              flattenServiceCodeRepository(config.CodeRepository),
		"image_repository":             flattenServiceImageRepository(config.ImageRepository),
	}

	return []interface{}{m}
}

func flattenServiceAuthenticationConfiguration(config *apprunner.AuthenticationConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"access_role_arn": aws.StringValue(config.AccessRoleArn),
		"connection_arn":  aws.StringValue(config.ConnectionArn),
	}

	return []interface{}{m}
}

func flattenServiceImageConfiguration(config *apprunner.ImageConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"port":                          aws.StringValue(config.Port),
		"runtime_environment_secrets":   aws.StringValueMap(config.RuntimeEnvironmentSecrets),
		"runtime_environment_variables": aws.StringValueMap(config.RuntimeEnvironmentVariables),
		"start_command":                 aws.StringValue(config.StartCommand),
	}

	return []interface{}{m}
}

func flattenServiceImageRepository(r *apprunner.ImageRepository) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"image_configuration":   flattenServiceImageConfiguration(r.ImageConfiguration),
		"image_identifier":      aws.StringValue(r.ImageIdentifier),
		"image_repository_type": aws.StringValue(r.ImageRepositoryType),
	}

	return []interface{}{m}
}
