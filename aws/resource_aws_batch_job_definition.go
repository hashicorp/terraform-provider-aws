package aws

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/private/protocol/json/jsonutil"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/equivalency"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/batch/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func resourceAwsBatchJobDefinition() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsBatchJobDefinitionCreate,
		Read:   resourceAwsBatchJobDefinitionRead,
		Update: resourceAwsBatchJobDefinitionUpdate,
		Delete: resourceAwsBatchJobDefinitionDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateBatchName,
			},
			"container_properties": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					equal, _ := equivalency.EquivalentBatchContainerPropertiesJSON(old, new)

					return equal
				},
				ValidateFunc: validateAwsBatchJobContainerProperties,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"platform_capabilities": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(batch.PlatformCapability_Values(), true),
				},
			},
			"retry_strategy": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attempts": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntBetween(1, 10),
						},
					},
				},
			},
			"tags": tagsSchema(),
			"propagate_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
				Default:  false,
			},
			"timeout": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"attempt_duration_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							ValidateFunc: validation.IntAtLeast(60),
						},
					},
				},
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{batch.JobDefinitionTypeContainer}, true),
			},
			"revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsBatchJobDefinitionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	name := d.Get("name").(string)

	input := &batch.RegisterJobDefinitionInput{
		JobDefinitionName: aws.String(name),
		Type:              aws.String(d.Get("type").(string)),
		PropagateTags:     aws.Bool(d.Get("propagate_tags").(bool)),
	}

	if v, ok := d.GetOk("container_properties"); ok {
		props, err := expandBatchJobContainerProperties(v.(string))
		if err != nil {
			return err
		}

		input.ContainerProperties = props
	}

	if v, ok := d.GetOk("parameters"); ok {
		input.Parameters = expandJobDefinitionParameters(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("platform_capabilities"); ok && v.(*schema.Set).Len() > 0 {
		input.PlatformCapabilities = expandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("retry_strategy"); ok {
		input.RetryStrategy = expandJobDefinitionRetryStrategy(v.([]interface{}))
	}

	if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
		input.Tags = keyvaluetags.New(v).IgnoreAws().BatchTags()
	}

	if v, ok := d.GetOk("timeout"); ok {
		input.Timeout = expandJobDefinitionTimeout(v.([]interface{}))
	}

	output, err := conn.RegisterJobDefinition(input)

	if err != nil {
		return fmt.Errorf("error creating Batch Job Definition (%s): %w", name, err)
	}

	d.SetId(aws.StringValue(output.JobDefinitionArn))

	return resourceAwsBatchJobDefinitionRead(d, meta)
}

func resourceAwsBatchJobDefinitionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	jobDefinition, err := finder.JobDefinitionByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Batch Job Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Batch Job Definition (%s): %w", d.Id(), err)
	}

	d.Set("arn", jobDefinition.JobDefinitionArn)

	containerProperties, err := flattenBatchContainerProperties(jobDefinition.ContainerProperties)

	if err != nil {
		return fmt.Errorf("error converting Batch Container Properties to JSON: %w", err)
	}

	if err := d.Set("container_properties", containerProperties); err != nil {
		return fmt.Errorf("error setting container_properties: %w", err)
	}

	d.Set("name", jobDefinition.JobDefinitionName)
	d.Set("parameters", aws.StringValueMap(jobDefinition.Parameters))
	d.Set("platform_capabilities", aws.StringValueSlice(jobDefinition.PlatformCapabilities))
	d.Set("propagate_tags", jobDefinition.PropagateTags)

	if err := d.Set("retry_strategy", flattenBatchRetryStrategy(jobDefinition.RetryStrategy)); err != nil {
		return fmt.Errorf("error setting retry_strategy: %w", err)
	}

	if err := d.Set("tags", keyvaluetags.BatchKeyValueTags(jobDefinition.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("timeout", flattenBatchJobTimeout(jobDefinition.Timeout)); err != nil {
		return fmt.Errorf("error setting timeout: %w", err)
	}

	d.Set("revision", jobDefinition.Revision)
	d.Set("type", jobDefinition.Type)

	return nil
}

func resourceAwsBatchJobDefinitionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := keyvaluetags.BatchUpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %w", err)
		}
	}

	return nil
}

func resourceAwsBatchJobDefinitionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).batchconn

	_, err := conn.DeregisterJobDefinition(&batch.DeregisterJobDefinitionInput{
		JobDefinition: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("error deleting Batch Job Definition (%s): %w", d.Id(), err)
	}

	return nil
}

func validateAwsBatchJobContainerProperties(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	_, err := expandBatchJobContainerProperties(value)
	if err != nil {
		errors = append(errors, fmt.Errorf("AWS Batch Job container_properties is invalid: %s", err))
	}
	return
}

func expandBatchJobContainerProperties(rawProps string) (*batch.ContainerProperties, error) {
	var props *batch.ContainerProperties

	err := json.Unmarshal([]byte(rawProps), &props)
	if err != nil {
		return nil, fmt.Errorf("Error decoding JSON: %s", err)
	}

	return props, nil
}

// Convert batch.ContainerProperties object into its JSON representation
func flattenBatchContainerProperties(containerProperties *batch.ContainerProperties) (string, error) {
	b, err := jsonutil.BuildJSON(containerProperties)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

func expandJobDefinitionParameters(params map[string]interface{}) map[string]*string {
	var jobParams = make(map[string]*string)
	for k, v := range params {
		jobParams[k] = aws.String(v.(string))
	}

	return jobParams
}

func expandJobDefinitionRetryStrategy(item []interface{}) *batch.RetryStrategy {
	retryStrategy := &batch.RetryStrategy{}
	data := item[0].(map[string]interface{})

	if v, ok := data["attempts"].(int); ok && v > 0 && v <= 10 {
		retryStrategy.Attempts = aws.Int64(int64(v))
	}

	return retryStrategy
}

func flattenBatchRetryStrategy(item *batch.RetryStrategy) []map[string]interface{} {
	data := []map[string]interface{}{}
	if item != nil && item.Attempts != nil {
		data = append(data, map[string]interface{}{
			"attempts": int(aws.Int64Value(item.Attempts)),
		})
	}
	return data
}

func expandJobDefinitionTimeout(item []interface{}) *batch.JobTimeout {
	timeout := &batch.JobTimeout{}
	data := item[0].(map[string]interface{})

	if v, ok := data["attempt_duration_seconds"].(int); ok && v >= 60 {
		timeout.AttemptDurationSeconds = aws.Int64(int64(v))
	}

	return timeout
}

func flattenBatchJobTimeout(item *batch.JobTimeout) []map[string]interface{} {
	data := []map[string]interface{}{}
	if item != nil && item.AttemptDurationSeconds != nil {
		data = append(data, map[string]interface{}{
			"attempt_duration_seconds": int(aws.Int64Value(item.AttemptDurationSeconds)),
		})
	}
	return data
}
