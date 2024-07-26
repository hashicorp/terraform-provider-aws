// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
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

// @SDKResource("aws_apprunner_service", name="Service")
// @Tags(identifierAttribute="arn")
func resourceService() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceServiceCreate,
		ReadWithoutTimeout:   resourceServiceRead,
		UpdateWithoutTimeout: resourceServiceUpdate,
		DeleteWithoutTimeout: resourceServiceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_scaling_configuration_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrEncryptionConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrKMSKey: {
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
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"healthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						names.AttrInterval: {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						names.AttrPath: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "/",
							ValidateFunc: validation.StringLenBetween(0, 51200),
						},
						names.AttrProtocol: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.HealthCheckProtocolTcp,
							ValidateDiagFunc: enum.Validate[types.HealthCheckProtocol](),
						},
						names.AttrTimeout: {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      2,
							ValidateFunc: validation.IntBetween(1, 20),
						},
						"unhealthy_threshold": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
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
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`256|512|1024|2048|4096|(0.25|0.5|1|2|4) vCPU`), ""),
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
							ValidateFunc: validation.StringMatch(regexache.MustCompile(`512|1024|2048|3072|4096|6144|8192|10240|12288|(0.5|1|2|3|4|6|8|10|12) GB`), ""),
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// App Runner API always returns the amount in MB
								return (old == "512" && new == "0.5 GB") || (old == "1024" && new == "1 GB") || (old == "2048" && new == "2 GB") || (old == "3072" && new == "3 GB") || (old == "4096" && new == "4 GB") || (old == "6144" && new == "6 GB") || (old == "8192" && new == "8 GB") || (old == "10240" && new == "10 GB") || (old == "12288" && new == "12 GB")
							},
						},
					},
				},
			},
			names.AttrNetworkConfiguration: {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"egress_configuration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"egress_type": {
										Type:             schema.TypeString,
										Optional:         true,
										Computed:         true,
										ValidateDiagFunc: enum.Validate[types.EgressType](),
									},
									"vpc_connector_arn": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
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
						names.AttrIPAddressType: {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          types.IpAddressTypeIpv4,
							ValidateDiagFunc: enum.Validate[types.IpAddressType](),
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
			names.AttrServiceName: {
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
															names.AttrPort: {
																Type:         schema.TypeString,
																Optional:     true,
																Default:      "8080",
																ValidateFunc: validation.StringLenBetween(0, 51200),
															},
															"runtime": {
																Type:             schema.TypeString,
																Required:         true,
																ValidateDiagFunc: enum.Validate[types.Runtime](),
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
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.ConfigurationSource](),
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
												names.AttrType: {
													Type:             schema.TypeString,
													Required:         true,
													ValidateDiagFunc: enum.Validate[types.SourceCodeVersionType](),
												},
												names.AttrValue: {
													Type:         schema.TypeString,
													Required:     true,
													ValidateFunc: validation.StringLenBetween(0, 51200),
												},
											},
										},
									},
									"source_directory": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.StringLenBetween(0, 4096),
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
												names.AttrPort: {
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
										ValidateFunc: validation.StringMatch(regexache.MustCompile(`([0-9]{12}\.dkr\.ecr\.[a-z\-]+-[0-9]{1}\.amazonaws\.com\/.*)|(^public\.ecr\.aws\/.+\/.+)`), ""),
									},
									"image_repository_type": {
										Type:             schema.TypeString,
										Required:         true,
										ValidateDiagFunc: enum.Validate[types.ImageRepositoryType](),
									},
								},
							},
							ExactlyOneOf: []string{"source_configuration.0.image_repository", "source_configuration.0.code_repository"},
						},
					},
				},
			},
			names.AttrStatus: {
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
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	name := d.Get(names.AttrServiceName).(string)
	input := &apprunner.CreateServiceInput{
		ServiceName:         aws.String(name),
		SourceConfiguration: expandServiceSourceConfiguration(d.Get("source_configuration").([]interface{})),
		Tags:                getTagsIn(ctx),
	}

	if v, ok := d.GetOk("auto_scaling_configuration_arn"); ok {
		input.AutoScalingConfigurationArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrEncryptionConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.EncryptionConfiguration = expandServiceEncryptionConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("health_check_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.HealthCheckConfiguration = expandServiceHealthCheckConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("instance_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.InstanceConfiguration = expandServiceInstanceConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrNetworkConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.NetworkConfiguration = expandNetworkConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("observability_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ObservabilityConfiguration = expandServiceObservabilityConfiguration(v.([]interface{}))
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*types.InvalidRequestException](ctx, propagationTimeout, func() (interface{}, error) {
		return conn.CreateService(ctx, input)
	}, "Error in assuming instance role")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Runner Service (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*apprunner.CreateServiceOutput).Service.ServiceArn))

	if _, err := waitServiceCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner Service (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	service, err := findServiceByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Runner Service (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Runner Service (%s): %s", d.Id(), err)
	}

	serviceURL := aws.ToString(service.ServiceUrl)

	if serviceURL == "" {
		// Alternate lookup required for private services.
		input := &apprunner.DescribeCustomDomainsInput{
			ServiceArn: aws.String(d.Id()),
		}

		err := forEachCustomDomainPage(ctx, conn, input, func(page *apprunner.DescribeCustomDomainsOutput) {
			serviceURL = aws.ToString(page.DNSTarget)
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading App Runner Service (%s) custom domains: %s", d.Id(), err)
		}
	}

	d.Set(names.AttrARN, service.ServiceArn)
	if service.AutoScalingConfigurationSummary != nil {
		d.Set("auto_scaling_configuration_arn", service.AutoScalingConfigurationSummary.AutoScalingConfigurationArn)
	} else {
		d.Set("auto_scaling_configuration_arn", nil)
	}
	if err := d.Set(names.AttrEncryptionConfiguration, flattenServiceEncryptionConfiguration(service.EncryptionConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting encryption_configuration: %s", err)
	}
	if err := d.Set("health_check_configuration", flattenServiceHealthCheckConfiguration(service.HealthCheckConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting health_check_configuration: %s", err)
	}
	if err := d.Set("instance_configuration", flattenServiceInstanceConfiguration(service.InstanceConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_configuration: %s", err)
	}
	if err := d.Set(names.AttrNetworkConfiguration, flattenNetworkConfiguration(service.NetworkConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting network_configuration: %s", err)
	}
	if err := d.Set("observability_configuration", flattenServiceObservabilityConfiguration(service.ObservabilityConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting observability_configuration: %s", err)
	}
	d.Set("service_id", service.ServiceId)
	d.Set(names.AttrServiceName, service.ServiceName)
	d.Set("service_url", serviceURL)
	if err := d.Set("source_configuration", flattenServiceSourceConfiguration(service.SourceConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source_configuration: %s", err)
	}
	d.Set(names.AttrStatus, service.Status)

	return diags
}

func resourceServiceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &apprunner.UpdateServiceInput{
			ServiceArn: aws.String(d.Id()),
		}

		if d.HasChange("auto_scaling_configuration_arn") {
			input.AutoScalingConfigurationArn = aws.String(d.Get("auto_scaling_configuration_arn").(string))
		}

		if d.HasChange("health_check_configuration") {
			input.HealthCheckConfiguration = expandServiceHealthCheckConfiguration(d.Get("health_check_configuration").([]interface{}))
		}

		if d.HasChange("instance_configuration") {
			input.InstanceConfiguration = expandServiceInstanceConfiguration(d.Get("instance_configuration").([]interface{}))
		}

		if d.HasChange(names.AttrNetworkConfiguration) {
			input.NetworkConfiguration = expandNetworkConfiguration(d.Get(names.AttrNetworkConfiguration).([]interface{}))
		}

		if d.HasChange("observability_configuration") {
			input.ObservabilityConfiguration = expandServiceObservabilityConfiguration(d.Get("observability_configuration").([]interface{}))
		}

		if d.HasChange("source_configuration") {
			input.SourceConfiguration = expandServiceSourceConfiguration(d.Get("source_configuration").([]interface{}))
		}

		_, err := conn.UpdateService(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating App Runner Service (%s): %s", d.Id(), err)
		}

		if _, err := waitServiceUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for App Runner Service (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceServiceRead(ctx, d, meta)...)
}

func resourceServiceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	log.Printf("[INFO] Deleting App Runner Service: %s", d.Id())
	_, err := conn.DeleteService(ctx, &apprunner.DeleteServiceInput{
		ServiceArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Runner Service (%s): %s", d.Id(), err)
	}

	if _, err := waitServiceDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner Service (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findServiceByARN(ctx context.Context, conn *apprunner.Client, arn string) (*types.Service, error) {
	input := &apprunner.DescribeServiceInput{
		ServiceArn: aws.String(arn),
	}

	output, err := conn.DescribeService(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Service == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.Service.Status; status == types.ServiceStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output.Service, nil
}

func statusService(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findServiceByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitServiceCreated(ctx context.Context, conn *apprunner.Client, arn string) (*types.Service, error) {
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusOperationInProgress),
		Target:  enum.Slice(types.ServiceStatusRunning),
		Refresh: statusService(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Service); ok {
		return output, err
	}

	return nil, err
}

func waitServiceUpdated(ctx context.Context, conn *apprunner.Client, arn string) (*types.Service, error) {
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusOperationInProgress),
		Target:  enum.Slice(types.ServiceStatusRunning),
		Refresh: statusService(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Service); ok {
		return output, err
	}

	return nil, err
}

func waitServiceDeleted(ctx context.Context, conn *apprunner.Client, arn string) (*types.Service, error) {
	const (
		timeout = 20 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ServiceStatusRunning, types.ServiceStatusOperationInProgress),
		Target:  []string{},
		Refresh: statusService(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.Service); ok {
		return output, err
	}

	return nil, err
}

func expandServiceEncryptionConfiguration(l []interface{}) *types.EncryptionConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.EncryptionConfiguration{}

	if v, ok := tfMap[names.AttrKMSKey].(string); ok && v != "" {
		result.KmsKey = aws.String(v)
	}

	return result
}

func expandServiceHealthCheckConfiguration(l []interface{}) *types.HealthCheckConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.HealthCheckConfiguration{}

	if v, ok := tfMap["healthy_threshold"].(int); ok {
		result.HealthyThreshold = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrInterval].(int); ok {
		result.Interval = aws.Int32(int32(v))
	}

	if v, ok := tfMap[names.AttrPath].(string); ok {
		result.Path = aws.String(v)
	}

	if v, ok := tfMap[names.AttrProtocol].(string); ok {
		result.Protocol = types.HealthCheckProtocol(v)
	}

	if v, ok := tfMap[names.AttrTimeout].(int); ok {
		result.Timeout = aws.Int32(int32(v))
	}

	if v, ok := tfMap["unhealthy_threshold"].(int); ok {
		result.UnhealthyThreshold = aws.Int32(int32(v))
	}

	return result
}

func expandServiceInstanceConfiguration(l []interface{}) *types.InstanceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.InstanceConfiguration{}

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

func expandNetworkConfiguration(l []interface{}) *types.NetworkConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.NetworkConfiguration{}

	if v, ok := tfMap["ingress_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.IngressConfiguration = expandNetworkIngressConfiguration(v)
	}

	if v, ok := tfMap["egress_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.EgressConfiguration = expandNetworkEgressConfiguration(v)
	}

	if v, ok := tfMap[names.AttrIPAddressType].(string); ok && v != "" {
		result.IpAddressType = types.IpAddressType(v)
	}

	return result
}

func expandServiceObservabilityConfiguration(l []interface{}) *types.ServiceObservabilityConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.ServiceObservabilityConfiguration{}

	if v, ok := tfMap["observability_configuration_arn"].(string); ok && len(v) > 0 {
		result.ObservabilityConfigurationArn = aws.String(v)
	}

	if v, ok := tfMap["observability_enabled"].(bool); ok {
		result.ObservabilityEnabled = v
	}

	return result
}

func expandServiceSourceConfiguration(l []interface{}) *types.SourceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.SourceConfiguration{}

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

func expandServiceAuthenticationConfiguration(l []interface{}) *types.AuthenticationConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.AuthenticationConfiguration{}

	if v, ok := tfMap["access_role_arn"].(string); ok && v != "" {
		result.AccessRoleArn = aws.String(v)
	}

	if v, ok := tfMap["connection_arn"].(string); ok && v != "" {
		result.ConnectionArn = aws.String(v)
	}

	return result
}

func expandNetworkIngressConfiguration(l []interface{}) *types.IngressConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.IngressConfiguration{}

	if v, ok := tfMap["is_publicly_accessible"].(bool); ok {
		result.IsPubliclyAccessible = v
	}

	return result
}

func expandNetworkEgressConfiguration(l []interface{}) *types.EgressConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.EgressConfiguration{}

	if v, ok := tfMap["egress_type"].(string); ok {
		result.EgressType = types.EgressType(v)
	}

	if v, ok := tfMap["vpc_connector_arn"].(string); ok && v != "" {
		result.VpcConnectorArn = aws.String(v)
	}

	return result
}

func expandServiceImageConfiguration(l []interface{}) *types.ImageConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.ImageConfiguration{}

	if v, ok := tfMap[names.AttrPort].(string); ok && v != "" {
		result.Port = aws.String(v)
	}

	if v, ok := tfMap["runtime_environment_secrets"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentSecrets = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["runtime_environment_variables"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentVariables = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["start_command"].(string); ok && v != "" {
		result.StartCommand = aws.String(v)
	}

	return result
}

func expandServiceCodeRepository(l []interface{}) *types.CodeRepository {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.CodeRepository{}

	if v, ok := tfMap["source_code_version"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.SourceCodeVersion = expandServiceSourceCodeVersion(v)
	}

	if v, ok := tfMap["code_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.CodeConfiguration = expandServiceCodeConfiguration(v)
	}

	if v, ok := tfMap["repository_url"].(string); ok && v != "" {
		result.RepositoryUrl = aws.String(v)
	}

	if v, ok := tfMap["source_directory"].(string); ok && v != "" {
		result.SourceDirectory = aws.String(v)
	}

	return result
}

func expandServiceImageRepository(l []interface{}) *types.ImageRepository {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.ImageRepository{}

	if v, ok := tfMap["image_configuration"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.ImageConfiguration = expandServiceImageConfiguration(v)
	}

	if v, ok := tfMap["image_identifier"].(string); ok && v != "" {
		result.ImageIdentifier = aws.String(v)
	}

	if v, ok := tfMap["image_repository_type"].(string); ok && v != "" {
		result.ImageRepositoryType = types.ImageRepositoryType(v)
	}

	return result
}

func expandServiceCodeConfiguration(l []interface{}) *types.CodeConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.CodeConfiguration{}

	if v, ok := tfMap["configuration_source"].(string); ok && v != "" {
		result.ConfigurationSource = types.ConfigurationSource(v)
	}

	if v, ok := tfMap["code_configuration_values"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		result.CodeConfigurationValues = expandServiceCodeConfigurationValues(v)
	}

	return result
}

func expandServiceCodeConfigurationValues(l []interface{}) *types.CodeConfigurationValues {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.CodeConfigurationValues{}

	if v, ok := tfMap["build_command"].(string); ok && v != "" {
		result.BuildCommand = aws.String(v)
	}

	if v, ok := tfMap[names.AttrPort].(string); ok && v != "" {
		result.Port = aws.String(v)
	}

	if v, ok := tfMap["runtime"].(string); ok && v != "" {
		result.Runtime = types.Runtime(v)
	}

	if v, ok := tfMap["runtime_environment_secrets"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentSecrets = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["runtime_environment_variables"].(map[string]interface{}); ok && len(v) > 0 {
		result.RuntimeEnvironmentVariables = flex.ExpandStringValueMap(v)
	}

	if v, ok := tfMap["start_command"].(string); ok && v != "" {
		result.StartCommand = aws.String(v)
	}

	return result
}

func expandServiceSourceCodeVersion(l []interface{}) *types.SourceCodeVersion {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	tfMap, ok := l[0].(map[string]interface{})

	if !ok {
		return nil
	}

	result := &types.SourceCodeVersion{}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		result.Type = types.SourceCodeVersionType(v)
	}

	if v, ok := tfMap[names.AttrValue].(string); ok && v != "" {
		result.Value = aws.String(v)
	}

	return result
}

func flattenServiceEncryptionConfiguration(config *types.EncryptionConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrKMSKey: aws.ToString(config.KmsKey),
	}

	return []interface{}{m}
}

func flattenServiceHealthCheckConfiguration(config *types.HealthCheckConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"healthy_threshold":   config.HealthyThreshold,
		names.AttrInterval:    config.Interval,
		names.AttrPath:        aws.ToString(config.Path),
		names.AttrProtocol:    string(config.Protocol),
		names.AttrTimeout:     config.Timeout,
		"unhealthy_threshold": config.UnhealthyThreshold,
	}

	return []interface{}{m}
}

func flattenServiceInstanceConfiguration(config *types.InstanceConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"cpu":               aws.ToString(config.Cpu),
		"instance_role_arn": aws.ToString(config.InstanceRoleArn),
		"memory":            aws.ToString(config.Memory),
	}

	return []interface{}{m}
}

func flattenNetworkConfiguration(config *types.NetworkConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"ingress_configuration": flattenNetworkIngressConfiguration(config.IngressConfiguration),
		"egress_configuration":  flattenNetworkEgressConfiguration(config.EgressConfiguration),
		names.AttrIPAddressType: config.IpAddressType,
	}

	return []interface{}{m}
}

func flattenNetworkIngressConfiguration(config *types.IngressConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"is_publicly_accessible": config.IsPubliclyAccessible,
	}

	return []interface{}{m}
}

func flattenNetworkEgressConfiguration(config *types.EgressConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"egress_type":       string(config.EgressType),
		"vpc_connector_arn": aws.ToString(config.VpcConnectorArn),
	}

	return []interface{}{m}
}

func flattenServiceObservabilityConfiguration(config *types.ServiceObservabilityConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"observability_configuration_arn": aws.ToString(config.ObservabilityConfigurationArn),
		"observability_enabled":           config.ObservabilityEnabled,
	}

	return []interface{}{m}
}

func flattenServiceCodeRepository(r *types.CodeRepository) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"code_configuration":  flattenServiceCodeConfiguration(r.CodeConfiguration),
		"repository_url":      aws.ToString(r.RepositoryUrl),
		"source_code_version": flattenServiceSourceCodeVersion(r.SourceCodeVersion),
		"source_directory":    aws.ToString(r.SourceDirectory),
	}

	return []interface{}{m}
}

func flattenServiceCodeConfiguration(config *types.CodeConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"code_configuration_values": flattenServiceCodeConfigurationValues(config.CodeConfigurationValues),
		"configuration_source":      string(config.ConfigurationSource),
	}

	return []interface{}{m}
}

func flattenServiceCodeConfigurationValues(values *types.CodeConfigurationValues) []interface{} {
	if values == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"build_command":                 aws.ToString(values.BuildCommand),
		names.AttrPort:                  aws.ToString(values.Port),
		"runtime":                       string(values.Runtime),
		"runtime_environment_secrets":   values.RuntimeEnvironmentSecrets,
		"runtime_environment_variables": values.RuntimeEnvironmentVariables,
		"start_command":                 aws.ToString(values.StartCommand),
	}

	return []interface{}{m}
}

func flattenServiceSourceCodeVersion(v *types.SourceCodeVersion) []interface{} {
	if v == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrType:  string(v.Type),
		names.AttrValue: aws.ToString(v.Value),
	}

	return []interface{}{m}
}

func flattenServiceSourceConfiguration(config *types.SourceConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"authentication_configuration": flattenServiceAuthenticationConfiguration(config.AuthenticationConfiguration),
		"auto_deployments_enabled":     aws.ToBool(config.AutoDeploymentsEnabled),
		"code_repository":              flattenServiceCodeRepository(config.CodeRepository),
		"image_repository":             flattenServiceImageRepository(config.ImageRepository),
	}

	return []interface{}{m}
}

func flattenServiceAuthenticationConfiguration(config *types.AuthenticationConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"access_role_arn": aws.ToString(config.AccessRoleArn),
		"connection_arn":  aws.ToString(config.ConnectionArn),
	}

	return []interface{}{m}
}

func flattenServiceImageConfiguration(config *types.ImageConfiguration) []interface{} {
	if config == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrPort:                  aws.ToString(config.Port),
		"runtime_environment_secrets":   config.RuntimeEnvironmentSecrets,
		"runtime_environment_variables": config.RuntimeEnvironmentVariables,
		"start_command":                 aws.ToString(config.StartCommand),
	}

	return []interface{}{m}
}

func flattenServiceImageRepository(r *types.ImageRepository) []interface{} {
	if r == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"image_configuration":   flattenServiceImageConfiguration(r.ImageConfiguration),
		"image_identifier":      aws.ToString(r.ImageIdentifier),
		"image_repository_type": string(r.ImageRepositoryType),
	}

	return []interface{}{m}
}
