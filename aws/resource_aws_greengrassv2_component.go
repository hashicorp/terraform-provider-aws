package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrassv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsGreengrassv2Component() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassv2ComponentCreate,
		Read:   resourceAwsGreengrassv2ComponentRead,
		Update: resourceAwsGreengrassv2ComponentCreate,
		Delete: resourceAwsGreengrassv2ComponentDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"inline_recipe": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,

				ValidateFunc: validateStringIsJsonOrYaml,
				ExactlyOneOf: []string{"inline_recipe", "lambda_function"},
			},
			"lambda_function": {
				Type:     schema.TypeList,
				ForceNew: true,
				Optional: true,

				ExactlyOneOf: []string{"inline_recipe", "lambda_function"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"component_dependencies": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,

							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"component_name": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"dependency_type": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"version_requirement": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
								},
							},
						},
						"component_lambda_parameters": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"environment_variables": {
										Type:     schema.TypeMap,
										ForceNew: true,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"event_sources": {
										Type:     schema.TypeList,
										ForceNew: true,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"topic": {
													Type:     schema.TypeString,
													ForceNew: true,
													Required: true,
												},
												"type": {
													Type:     schema.TypeString,
													ForceNew: true,
													Required: true,
												},
											},
										},
									},
									"exec_args": {
										Type:     schema.TypeList,
										ForceNew: true,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"input_payload_encoding_type": {
										Type:         schema.TypeString,
										ForceNew:     true,
										Optional:     true,
										Default:      "json",
										ValidateFunc: validation.StringInSlice([]string{"binary", "json"}, false),
									},
									"linux_process_params": {
										Type:     schema.TypeList,
										ForceNew: true,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"container_params": {
													Type:     schema.TypeList,
													ForceNew: true,
													Optional: true,
													Elem: &schema.Resource{
														Schema: map[string]*schema.Schema{
															"devices": {
																Type:     schema.TypeList,
																ForceNew: true,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"add_group_owner": {
																			Type:     schema.TypeBool,
																			ForceNew: true,
																			Optional: true,
																			Default:  false,
																		},
																		"path": {
																			Type:     schema.TypeString,
																			ForceNew: true,
																			Required: true,
																		},
																		"permission": {
																			Type:     schema.TypeString,
																			ForceNew: true,
																			Optional: true,
																			Default:  "ro",
																		},
																	},
																},
															},
															"memory_size_in_kb": {
																Type:         schema.TypeInt,
																ForceNew:     true,
																Optional:     true,
																Default:      16384,
																ValidateFunc: validation.IntBetween(2048, 2147483647),
															},
															"mount_ro_sysfs": {
																Type:     schema.TypeBool,
																ForceNew: true,
																Optional: true,
																Default:  false,
															},
															"volumes": {
																Type:     schema.TypeList,
																ForceNew: true,
																Optional: true,
																Elem: &schema.Resource{
																	Schema: map[string]*schema.Schema{
																		"add_group_owner": {
																			Type:     schema.TypeBool,
																			ForceNew: true,
																			Optional: true,
																			Default:  false,
																		},
																		"destination_path": {
																			Type:     schema.TypeString,
																			ForceNew: true,
																			Required: true,
																		},
																		"permission": {
																			Type:     schema.TypeString,
																			ForceNew: true,
																			Optional: true,
																			Default:  "ro",
																		},
																		"source_path": {
																			Type:     schema.TypeString,
																			ForceNew: true,
																			Required: true,
																		},
																	},
																},
															},
														},
													},
												},
												"isolation_mode": {
													Type:         schema.TypeString,
													ForceNew:     true,
													Optional:     true,
													Default:      "GreengrassContainer",
													ValidateFunc: validation.StringInSlice([]string{"NoContainer", "GreengrassContainer"}, false),
												},
											},
										},
									},
									"max_idle_time_in_seconds": {
										Type:         schema.TypeInt,
										ForceNew:     true,
										Optional:     true,
										ValidateFunc: validation.IntBetween(30, 2147483647),
									},
									"max_instances_count": {
										Type:     schema.TypeInt,
										ForceNew: true,
										Optional: true,
									},
									"max_queue_size": {
										Type:     schema.TypeInt,
										ForceNew: true,
										Optional: true,
									},
									"pinned": {
										Type:     schema.TypeBool,
										ForceNew: true,
										Optional: true,
										Default:  true,
									},
									"status_timeout_in_seconds": {
										Type:         schema.TypeInt,
										ForceNew:     true,
										Optional:     true,
										ValidateFunc: validation.IntBetween(30, 2147483647),
									},
									"timeout_in_seconds": {
										Type:     schema.TypeInt,
										ForceNew: true,
										Optional: true,
									},
								},
							},
						},
						"component_name": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"component_platforms": {
							Type:     schema.TypeList,
							ForceNew: true,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"attributes": {
										Type:     schema.TypeList,
										ForceNew: true,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"os": {
													Type:     schema.TypeString,
													ForceNew: true,
													Optional: true,
												},
												"architecture": {
													Type:     schema.TypeString,
													ForceNew: true,
													Optional: true,
												},
											},
										},
									},
									"name": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
								},
							},
						},
						"component_version": {
							Type:     schema.TypeString,
							ForceNew: true,
							Optional: true,
						},
						"lambda_arn": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
					},
				},
			},
			"tags": tagsSchema(),
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Second),
			Update: schema.DefaultTimeout(10 * time.Second),
		},
	}
}

func resourceAwsGreengrassv2ComponentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassv2conn
	input := greengrassv2.CreateComponentVersionInput{}

	if v, ok := d.GetOk("inline_recipe"); ok {
		input.InlineRecipe = []byte((v.(string)))
	}

	if v, ok := d.GetOk("lambda_function"); ok {
		config := v.([]interface{})[0].(map[string]interface{})
		input.LambdaFunction = setLambdaFunctionRecipeSource(config)
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().Greengrassv2Tags()
	}

	log.Printf("[DEBUG] Creating Greengrassv2 component : %s", input)
	output, err := conn.CreateComponentVersion(&input)
	if err != nil {
		return fmt.Errorf("Creating Greengrassv2 component failed: %s", err.Error())
	}

	d.SetId(aws.StringValue(output.Arn))
	return resourceAwsGreengrassv2ComponentRead(d, meta)
}

func resourceAwsGreengrassv2ComponentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassv2conn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	getComponentInput := greengrassv2.GetComponentInput{
		Arn:                aws.String(d.Id()),
		RecipeOutputFormat: aws.String("JSON"),
	}

	log.Printf("[DEBUG] Reading Greengrassv2 component : %s", getComponentInput)
	getComponentOutput, err := conn.GetComponent(&getComponentInput)
	if err != nil {
		return fmt.Errorf("Reading Greengrassv2 component '%s' failed: %s", d.Id(), err.Error())
	}
	log.Printf("Get Greengrassv2 component %q", getComponentOutput)

	d.Set("arn", d.Id())
	d.Set("tags", keyvaluetags.Greengrassv2KeyValueTags(getComponentOutput.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map())

	describeComponentInput := greengrassv2.DescribeComponentInput{
		Arn: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrassv2 component : %s", describeComponentInput)
	describeComponentOutput, err := conn.DescribeComponent(&describeComponentInput)
	if err != nil {
		return fmt.Errorf("Reading Greengrassv2 component '%s' failed: %s", d.Id(), err.Error())
	}
	log.Printf("Describe Greengrassv2 component %q", describeComponentOutput)

	return nil
}

func resourceAwsGreengrassv2ComponentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassv2conn

	input := greengrassv2.DeleteComponentInput{
		Arn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Greengrassv2 component : %s", input)
	gmo, err := conn.DeleteComponent(&input)
	if err != nil {
		return fmt.Errorf("Deleting Greengrassv2 component '%s' failed: %s", d.Id(), err.Error())
	}
	log.Printf("[WARN] Greengrassv2 component %q not found, removing from state", gmo)
	return nil
}

func setLambdaFunctionRecipeSource(m map[string]interface{}) *greengrassv2.LambdaFunctionRecipeSource {
	lambdaFunctionRecipeSource := &greengrassv2.LambdaFunctionRecipeSource{}

	if v, ok := m["component_dependencies"].([]interface{}); ok {
		params := expandLambdaComponentDependencies(v)
		lambdaFunctionRecipeSource.SetComponentDependencies(params)
	}
	if v, ok := m["component_lambda_parameters"].([]interface{}); ok {
		params := expandLambdaComponentLambdaParameters(v)
		lambdaFunctionRecipeSource.SetComponentLambdaParameters(params)
	}

	if v, ok := m["component_name"].(string); ok {
		lambdaFunctionRecipeSource.SetComponentName(v)
	}

	if v, ok := m["component_platforms"].([]interface{}); ok {
		params := expandLambdaComponentPlatforms(v)
		lambdaFunctionRecipeSource.SetComponentPlatforms(params)
	}

	if v, ok := m["component_version"].(string); ok {
		lambdaFunctionRecipeSource.SetComponentVersion(v)
	}

	if v, ok := m["lambda_arn"].(string); ok {
		lambdaFunctionRecipeSource.SetLambdaArn(v)
	}
	return lambdaFunctionRecipeSource
}

func expandLambdaComponentDependencies(lcdMaps []interface{}) map[string]*greengrassv2.ComponentDependencyRequirement {
	if lcdMaps == nil {
		return nil
	}
	componentDependencyRequiremnet := greengrassv2.ComponentDependencyRequirement{}
	componentDependencyRequiremnets := map[string]*greengrassv2.ComponentDependencyRequirement{}
	for _, l := range lcdMaps {
		m := l.(map[string]interface{})
		if v, ok := m["dependency_type"].(string); ok {
			componentDependencyRequiremnet.SetDependencyType(v)
		}
		if v, ok := m["version_requirement"].(string); ok {
			componentDependencyRequiremnet.SetVersionRequirement(v)
		}
		if v, ok := m["component_name"].(string); ok {
			componentDependencyRequiremnets[v] = &componentDependencyRequiremnet
		}
	}
	log.Printf("[DEBUG] component_dependency_requiremnet : %s", componentDependencyRequiremnets)
	return componentDependencyRequiremnets
}

func expandLambdaComponentLambdaParameters(lclpMaps []interface{}) *greengrassv2.LambdaExecutionParameters {
	if lclpMaps == nil {
		return nil
	}
	log.Printf("[DEBUG] expandLambdaComponentLambdaParameters input : %s", lclpMaps)
	lambdaExecutionParameter := greengrassv2.LambdaExecutionParameters{}

	m := lclpMaps[0].(map[string]interface{})
	log.Printf("[DEBUG] expandLambdaComponentLambdaParameters param : %s", m)
	if v, ok := m["environment_variables"].(map[string]interface{}); ok {
		lambdaExecutionParameter.SetEnvironmentVariables(stringMapToPointers(v))
	}
	if v, ok := m["event_sources"].([]interface{}); ok {
		lambdaExecutionParameter.SetEventSources(expandLambdaEventSources(v))
	}
	if v, ok := m["exec_args"].([]interface{}); ok {
		lambdaExecutionParameter.SetExecArgs(expandStringList(v))
	}
	if v, ok := m["input_payload_encoding_type"].(string); ok {
		lambdaExecutionParameter.SetInputPayloadEncodingType(v)
	}
	if v, ok := m["linux_process_params"].([]interface{}); ok {
		lambdaExecutionParameter.SetLinuxProcessParams(expandLinuxProcessParams(v))
	}
	if v, ok := m["max_idle_time_in_seconds"].(int); ok {
		lambdaExecutionParameter.SetMaxIdleTimeInSeconds(int64(v))
	}
	if v, ok := m["max_instances_count"].(int); ok {
		lambdaExecutionParameter.SetMaxInstancesCount(int64(v))
	}
	if v, ok := m["max_queue_size"].(int); ok {
		lambdaExecutionParameter.SetMaxQueueSize(int64(v))
	}
	if v, ok := m["pinned"].(bool); ok {
		lambdaExecutionParameter.SetPinned(v)
	}
	if v, ok := m["status_timeout_in_seconds"].(int); ok {
		lambdaExecutionParameter.SetStatusTimeoutInSeconds(int64(v))
	}
	if v, ok := m["timeout_in_seconds"].(int); ok {
		lambdaExecutionParameter.SetTimeoutInSeconds(int64(v))
	}
	log.Printf("[DEBUG] lambdaExecutionParameter is : %s", lambdaExecutionParameter)
	return &lambdaExecutionParameter
}

func expandLambdaComponentPlatforms(lcpMaps []interface{}) []*greengrassv2.ComponentPlatform {
	if lcpMaps == nil {
		return nil
	}
	log.Printf("[DEBUG] expandLambdaComponentPlatforms input : %s", lcpMaps)
	componentPlatforms := []*greengrassv2.ComponentPlatform{}
	for _, l := range lcpMaps {
		componentPlatform := greengrassv2.ComponentPlatform{}
		m := l.(map[string]interface{})
		if v, ok := m["attributes"].([]interface{})[0].(map[string]interface{}); ok {
			attribute := map[string]*string{
				"os":           aws.String(v["os"].(string)),
				"architecture": aws.String(v["architecture"].(string)),
			}
			componentPlatform.SetAttributes(attribute)
		}
		if v, ok := m["name"].(string); ok {
			componentPlatform.SetName(v)
		}
		componentPlatforms = append(componentPlatforms, &componentPlatform)
	}
	log.Printf("[DEBUG] component platform is : %s", componentPlatforms)
	return componentPlatforms
}

func expandLambdaEventSources(esMaps []interface{}) []*greengrassv2.LambdaEventSource {
	if esMaps == nil {
		return nil
	}
	log.Printf("[DEBUG] expandLambdaEventSources input : %s", esMaps)
	eventSources := []*greengrassv2.LambdaEventSource{}

	for _, l := range esMaps {
		eventSource := greengrassv2.LambdaEventSource{}
		m := l.(map[string]interface{})
		if v, ok := m["topic"].(interface{}).(string); ok {
			eventSource.SetTopic(v)
		}
		if v, ok := m["type"].(interface{}).(string); ok {
			eventSource.SetType(v)
		}
		eventSources = append(eventSources, &eventSource)
	}
	log.Printf("[DEBUG] event source is : %s", eventSources)
	return eventSources
}

func expandLinuxProcessParams(lppMaps []interface{}) *greengrassv2.LambdaLinuxProcessParams {
	if lppMaps == nil {
		return nil
	}
	log.Printf("[DEBUG] expandLinuxProcessParams input : %s", lppMaps)
	linuxProcessParams := &greengrassv2.LambdaLinuxProcessParams{}
	m := lppMaps[0].(map[string]interface{})
	log.Printf("[DEBUG] lppMaps input : %s", m)
	if v, ok := m["container_params"].([]interface{}); ok {
		linuxProcessParams.SetContainerParams(expandContainerParams(v))
	}
	if v, ok := m["isolation_mode"].(string); ok {
		linuxProcessParams.SetIsolationMode(v)
	}
	log.Printf("[DEBUG] linuxProcessParams is : %s", linuxProcessParams)
	return linuxProcessParams
}

func expandContainerParams(cpMaps []interface{}) *greengrassv2.LambdaContainerParams {
	if cpMaps == nil {
		return nil
	}
	log.Printf("[DEBUG] expandContainerParams input : %s", cpMaps)
	lambdaContainerParams := &greengrassv2.LambdaContainerParams{}
	m := cpMaps[0].(map[string]interface{})
	if v, ok := m["devices"].([]interface{}); ok {
		lambdaContainerParams.SetDevices(expandDevices(v))
	}
	if v, ok := m["memory_size_in_kb"].(int); ok {
		lambdaContainerParams.SetMemorySizeInKB(int64(v))
	}
	if v, ok := m["mount_ro_sysfs"].(bool); ok {
		lambdaContainerParams.SetMountROSysfs(v)
	}
	if v, ok := m["volumes"].([]interface{}); ok {
		lambdaContainerParams.SetVolumes(expandVolumes(v))
	}
	log.Printf("[DEBUG] linuxProcessParams is : %s", lambdaContainerParams)
	return lambdaContainerParams
}

func expandDevices(dMaps []interface{}) []*greengrassv2.LambdaDeviceMount {
	if dMaps == nil {
		return nil
	}
	log.Printf("[DEBUG] expandDevices input : %s", dMaps)
	lambdaDeviceMounts := []*greengrassv2.LambdaDeviceMount{}

	for _, l := range dMaps {
		lambdaDeviceMount := greengrassv2.LambdaDeviceMount{}
		m := l.(map[string]interface{})

		if v, ok := m["add_group_owner"].(bool); ok {
			lambdaDeviceMount.SetAddGroupOwner(v)
		}
		if v, ok := m["path"].(string); ok {
			lambdaDeviceMount.SetPath(v)
		}
		if v, ok := m["permission"].(string); ok {
			lambdaDeviceMount.SetPermission(v)
		}
		lambdaDeviceMounts = append(lambdaDeviceMounts, &lambdaDeviceMount)
	}
	log.Printf("[DEBUG] lambdaDeviceMounts is : %s", lambdaDeviceMounts)
	return lambdaDeviceMounts
}

func expandVolumes(vMaps []interface{}) []*greengrassv2.LambdaVolumeMount {
	if vMaps == nil {
		return nil
	}
	log.Printf("[DEBUG] expandVolumes input : %s", vMaps)
	lambdaVolumeMounts := []*greengrassv2.LambdaVolumeMount{}

	for _, l := range vMaps {
		lambdaVolumeMount := greengrassv2.LambdaVolumeMount{}
		m := l.(map[string]interface{})

		if v, ok := m["add_group_owner"].(bool); ok {
			lambdaVolumeMount.SetAddGroupOwner(v)
		}
		if v, ok := m["destination_path"].(string); ok {
			lambdaVolumeMount.SetDestinationPath(v)
		}
		if v, ok := m["permission"].(string); ok {
			lambdaVolumeMount.SetPermission(v)
		}
		if v, ok := m["source_path"].(string); ok {
			lambdaVolumeMount.SetSourcePath(v)
		}
		lambdaVolumeMounts = append(lambdaVolumeMounts, &lambdaVolumeMount)
	}
	log.Printf("[DEBUG] lambdaVolumeMounts is : %s", lambdaVolumeMounts)
	return lambdaVolumeMounts
}
