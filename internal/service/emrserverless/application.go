package emrserverless

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emrserverless"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceApplicationCreate,
		Read:   resourceApplicationRead,
		Update: resourceApplicationUpdate,
		Delete: resourceApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"auto_start_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"auto_stop_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"idle_timeout_minutes": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      15,
							ValidateFunc: validation.IntBetween(1, 10080),
						},
					},
				},
			},
			"initial_capacity": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"initial_capacity_config": {
							Type:             schema.TypeList,
							Optional:         true,
							DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
							MaxItems:         1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"worker_configuration": {
										Type:             schema.TypeList,
										Optional:         true,
										DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
										MaxItems:         1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"cpu": {
													Type:     schema.TypeString,
													Required: true,
												},
												"disk": {
													Type:     schema.TypeString,
													Optional: true,
													Computed: true,
												},
												"memory": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
									"worker_count": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntBetween(1, 1000000),
									},
								},
							},
						},
						"initial_capacity_type": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"maximum_capacity": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cpu": {
							Type:     schema.TypeString,
							Required: true,
						},
						"disk": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"memory": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"network_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"security_group_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"subnet_ids": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"release_label": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(val interface{}) string {
					return strings.ToLower(val.(string))
				},
			},
		},
	}
}

func resourceApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRServerlessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &emrserverless.CreateApplicationInput{
		ClientToken:  aws.String(resource.UniqueId()),
		ReleaseLabel: aws.String(d.Get("release_label").(string)),
		Name:         aws.String(name),
		Type:         aws.String(d.Get("type").(string)),
	}

	if v, ok := d.GetOk("auto_start_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AutoStartConfiguration = expandAutoStartConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("auto_stop_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.AutoStopConfiguration = expandAutoStopConfig(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("initial_capacity"); ok && v.(*schema.Set).Len() > 0 {
		input.InitialCapacity = expandInitialCapacity(v.(*schema.Set))
	}

	if v, ok := d.GetOk("maximum_capacity"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.MaximumCapacity = expandMaximumCapacity(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("network_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.NetworkConfiguration = expandNetworkConfiguration(v.([]interface{})[0].(map[string]interface{}))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating EMR Serveless Application: %s", input)
	result, err := conn.CreateApplication(input)

	if err != nil {
		return fmt.Errorf("creating EMR Serveless Application (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(result.ApplicationId))

	if _, err := waitApplicationCreated(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EMR Serveless Application (%s) create: %w", d.Id(), err)
	}

	return resourceApplicationRead(d, meta)
}

func resourceApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRServerlessConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	application, err := FindApplicationByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EMR Serverless Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading EMR Serverless Application (%s): %w", d.Id(), err)
	}

	d.Set("arn", application.Arn)
	d.Set("name", application.Name)
	d.Set("release_label", application.ReleaseLabel)
	d.Set("type", strings.ToLower(aws.StringValue(application.Type)))

	if err := d.Set("auto_start_configuration", []interface{}{flattenAutoStartConfig(application.AutoStartConfiguration)}); err != nil {
		return fmt.Errorf("setting auto_start_configuration: %w", err)
	}

	if err := d.Set("auto_stop_configuration", []interface{}{flattenAutoStopConfig(application.AutoStopConfiguration)}); err != nil {
		return fmt.Errorf("setting auto_stop_configuration: %w", err)
	}

	if err := d.Set("initial_capacity", flattenInitialCapacity(application.InitialCapacity)); err != nil {
		return fmt.Errorf("setting initial_capacity: %w", err)
	}

	if err := d.Set("maximum_capacity", []interface{}{flattenMaximumCapacity(application.MaximumCapacity)}); err != nil {
		return fmt.Errorf("setting maximum_capacity: %w", err)
	}

	if err := d.Set("network_configuration", []interface{}{flattenNetworkConfiguration(application.NetworkConfiguration)}); err != nil {
		return fmt.Errorf("setting network_configuration: %w", err)
	}

	tags := KeyValueTags(application.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourceApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRServerlessConn

	if d.HasChangesExcept("tags", "tags_all") {
		input := &emrserverless.UpdateApplicationInput{
			ApplicationId: aws.String(d.Id()),
			ClientToken:   aws.String(resource.UniqueId()),
		}

		if v, ok := d.GetOk("auto_start_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.AutoStartConfiguration = expandAutoStartConfig(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("auto_stop_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.AutoStopConfiguration = expandAutoStopConfig(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("initial_capacity"); ok && v.(*schema.Set).Len() > 0 {
			input.InitialCapacity = expandInitialCapacity(v.(*schema.Set))
		}

		if v, ok := d.GetOk("maximum_capacity"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.MaximumCapacity = expandMaximumCapacity(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("network_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input.NetworkConfiguration = expandNetworkConfiguration(v.([]interface{})[0].(map[string]interface{}))
		}

		log.Printf("[DEBUG] Updating EMR Serveless Application: %s", input)
		_, err := conn.UpdateApplication(input)

		if err != nil {
			return fmt.Errorf("updating EMR Serveless Application (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("updating EMR Serverless Application (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceApplicationRead(d, meta)
}

func resourceApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EMRServerlessConn

	log.Printf("[INFO] Deleting EMR Serverless Application: %s", d.Id())
	_, err := conn.DeleteApplication(&emrserverless.DeleteApplicationInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, emrserverless.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting EMR Serverless Application (%s): %w", d.Id(), err)
	}

	if _, err := waitApplicationTerminated(conn, d.Id()); err != nil {
		return fmt.Errorf("waiting for EMR Serveless Application (%s) delete: %w", d.Id(), err)
	}

	return nil
}

func expandAutoStartConfig(tfMap map[string]interface{}) *emrserverless.AutoStartConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrserverless.AutoStartConfig{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	return apiObject
}

func flattenAutoStartConfig(apiObject *emrserverless.AutoStartConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	return tfMap
}

func expandAutoStopConfig(tfMap map[string]interface{}) *emrserverless.AutoStopConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrserverless.AutoStopConfig{}

	if v, ok := tfMap["enabled"].(bool); ok {
		apiObject.Enabled = aws.Bool(v)
	}

	if v, ok := tfMap["idle_timeout_minutes"].(int); ok {
		apiObject.IdleTimeoutMinutes = aws.Int64(int64(v))
	}

	return apiObject
}

func flattenAutoStopConfig(apiObject *emrserverless.AutoStopConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Enabled; v != nil {
		tfMap["enabled"] = aws.BoolValue(v)
	}

	if v := apiObject.IdleTimeoutMinutes; v != nil {
		tfMap["idle_timeout_minutes"] = aws.Int64Value(v)
	}

	return tfMap
}

func expandMaximumCapacity(tfMap map[string]interface{}) *emrserverless.MaximumAllowedResources {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrserverless.MaximumAllowedResources{}

	if v, ok := tfMap["cpu"].(string); ok && v != "" {
		apiObject.Cpu = aws.String(v)
	}

	if v, ok := tfMap["disk"].(string); ok && v != "" {
		apiObject.Disk = aws.String(v)
	}

	if v, ok := tfMap["memory"].(string); ok && v != "" {
		apiObject.Memory = aws.String(v)
	}

	return apiObject
}

func flattenMaximumCapacity(apiObject *emrserverless.MaximumAllowedResources) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Cpu; v != nil {
		tfMap["cpu"] = aws.StringValue(v)
	}

	if v := apiObject.Disk; v != nil {
		tfMap["disk"] = aws.StringValue(v)
	}

	if v := apiObject.Memory; v != nil {
		tfMap["memory"] = aws.StringValue(v)
	}

	return tfMap
}

func expandNetworkConfiguration(tfMap map[string]interface{}) *emrserverless.NetworkConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrserverless.NetworkConfiguration{}

	if v, ok := tfMap["security_group_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SecurityGroupIds = flex.ExpandStringSet(v)
	}

	if v, ok := tfMap["subnet_ids"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.SubnetIds = flex.ExpandStringSet(v)
	}

	return apiObject
}

func flattenNetworkConfiguration(apiObject *emrserverless.NetworkConfiguration) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.SecurityGroupIds; v != nil {
		tfMap["security_group_ids"] = flex.FlattenStringSet(v)
	}

	if v := apiObject.SubnetIds; v != nil {
		tfMap["subnet_ids"] = flex.FlattenStringSet(v)
	}

	return tfMap
}

func expandInitialCapacity(tfMap *schema.Set) map[string]*emrserverless.InitialCapacityConfig {
	if tfMap == nil {
		return nil
	}

	configs := make(map[string]*emrserverless.InitialCapacityConfig)

	for _, tfMapRaw := range tfMap.List() {

		config, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		if v, ok := config["initial_capacity_type"].(string); ok && v != "" {
			if conf, ok := config["initial_capacity_config"].([]interface{}); ok && len(conf) > 0 {
				configs[v] = expandInitialCapacityConfig(conf[0].(map[string]interface{}))
			}
		}
	}

	return configs
}

func flattenInitialCapacity(apiObject map[string]*emrserverless.InitialCapacityConfig) []interface{} {
	if apiObject == nil {
		return nil
	}

	var tfList []interface{}

	for capacityType, config := range apiObject {

		if config == nil {
			continue
		}

		tfList = append(tfList, map[string]interface{}{
			"initial_capacity_type":   capacityType,
			"initial_capacity_config": []interface{}{flattenInitialCapacityConfig(config)},
		})
	}

	return tfList
}

func expandInitialCapacityConfig(tfMap map[string]interface{}) *emrserverless.InitialCapacityConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrserverless.InitialCapacityConfig{}

	if v, ok := tfMap["worker_count"].(int); ok {
		apiObject.WorkerCount = aws.Int64(int64(v))
	}

	if v, ok := tfMap["worker_configuration"].([]interface{}); ok && v[0] != nil {
		apiObject.WorkerConfiguration = expandWorkerResourceConfig(v[0].(map[string]interface{}))
	}

	return apiObject
}

func flattenInitialCapacityConfig(apiObject *emrserverless.InitialCapacityConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.WorkerCount; v != nil {
		tfMap["worker_count"] = aws.Int64Value(v)
	}

	if v := apiObject.WorkerConfiguration; v != nil {
		tfMap["worker_configuration"] = []interface{}{flattenWorkerResourceConfig(v)}
	}

	return tfMap
}

func expandWorkerResourceConfig(tfMap map[string]interface{}) *emrserverless.WorkerResourceConfig {
	if tfMap == nil {
		return nil
	}

	apiObject := &emrserverless.WorkerResourceConfig{}

	if v, ok := tfMap["cpu"].(string); ok && v != "" {
		apiObject.Cpu = aws.String(v)
	}

	if v, ok := tfMap["disk"].(string); ok && v != "" {
		apiObject.Disk = aws.String(v)
	}

	if v, ok := tfMap["memory"].(string); ok && v != "" {
		apiObject.Memory = aws.String(v)
	}

	return apiObject
}

func flattenWorkerResourceConfig(apiObject *emrserverless.WorkerResourceConfig) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.Cpu; v != nil {
		tfMap["cpu"] = aws.StringValue(v)
	}

	if v := apiObject.Disk; v != nil {
		tfMap["disk"] = aws.StringValue(v)
	}

	if v := apiObject.Memory; v != nil {
		tfMap["memory"] = aws.StringValue(v)
	}

	return tfMap
}
