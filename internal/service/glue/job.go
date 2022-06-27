package glue

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceJob() *schema.Resource {
	return &schema.Resource{
		Create: resourceJobCreate,
		Read:   resourceJobRead,
		Update: resourceJobUpdate,
		Delete: resourceJobDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"command": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "glueetl",
						},
						"script_location": {
							Type:     schema.TypeString,
							Required: true,
						},
						"python_version": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"2", "3"}, true),
						},
					},
				},
			},
			"connections": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"default_arguments": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"glue_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"execution_property": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_concurrent_runs": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      1,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"max_capacity": {
				Type:          schema.TypeFloat,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"number_of_workers", "worker_type"},
			},
			"max_retries": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.NoZeroValues,
			},
			"notification_property": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"notify_delay_after": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
						},
					},
				},
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"timeout": {
				Type:         schema.TypeInt,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.IntAtLeast(1),
			},
			"security_configuration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"worker_type": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"max_capacity"},
				ValidateFunc:  validation.StringInSlice(glue.WorkerType_Values(), false),
			},
			"number_of_workers": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"max_capacity"},
				ValidateFunc:  validation.IntAtLeast(2),
			},
			"non_overridable_arguments": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceJobCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)

	input := &glue.CreateJobInput{
		Command: expandJobCommand(d.Get("command").([]interface{})),
		Name:    aws.String(name),
		Role:    aws.String(d.Get("role_arn").(string)),
		Tags:    Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("timeout"); ok {
		input.Timeout = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("max_capacity"); ok {
		input.MaxCapacity = aws.Float64(v.(float64))
	}

	if v, ok := d.GetOk("connections"); ok {
		input.Connections = &glue.ConnectionsList{
			Connections: flex.ExpandStringList(v.([]interface{})),
		}
	}

	if kv, ok := d.GetOk("default_arguments"); ok {
		input.DefaultArguments = flex.ExpandStringMap(kv.(map[string]interface{}))
	}

	if kv, ok := d.GetOk("non_overridable_arguments"); ok {
		input.NonOverridableArguments = flex.ExpandStringMap(kv.(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("glue_version"); ok {
		input.GlueVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_property"); ok {
		input.ExecutionProperty = expandExecutionProperty(v.([]interface{}))
	}

	if v, ok := d.GetOk("max_retries"); ok {
		input.MaxRetries = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("notification_property"); ok {
		input.NotificationProperty = expandNotificationProperty(v.([]interface{}))
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = aws.String(v.(string))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating Glue Job: %s", input)
	_, err := conn.CreateJob(input)
	if err != nil {
		return fmt.Errorf("error creating Glue Job (%s): %s", name, err)
	}

	d.SetId(name)

	return resourceJobRead(d, meta)
}

func resourceJobRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &glue.GetJobInput{
		JobName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Glue Job: %s", input)
	output, err := conn.GetJob(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			log.Printf("[WARN] Glue Job (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Glue Job (%s): %s", d.Id(), err)
	}

	job := output.Job
	if job == nil {
		log.Printf("[WARN] Glue Job (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	jobARN := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "glue",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("job/%s", d.Id()),
	}.String()
	d.Set("arn", jobARN)

	if err := d.Set("command", flattenJobCommand(job.Command)); err != nil {
		return fmt.Errorf("error setting command: %s", err)
	}
	if err := d.Set("connections", flattenConnectionsList(job.Connections)); err != nil {
		return fmt.Errorf("error setting connections: %s", err)
	}
	if err := d.Set("default_arguments", aws.StringValueMap(job.DefaultArguments)); err != nil {
		return fmt.Errorf("error setting default_arguments: %s", err)
	}
	if err := d.Set("non_overridable_arguments", aws.StringValueMap(job.NonOverridableArguments)); err != nil {
		return fmt.Errorf("error setting non_overridable_arguments: %w", err)
	}
	d.Set("description", job.Description)
	d.Set("glue_version", job.GlueVersion)
	if err := d.Set("execution_property", flattenExecutionProperty(job.ExecutionProperty)); err != nil {
		return fmt.Errorf("error setting execution_property: %s", err)
	}
	d.Set("max_capacity", job.MaxCapacity)
	d.Set("max_retries", job.MaxRetries)
	if err := d.Set("notification_property", flattenNotificationProperty(job.NotificationProperty)); err != nil {
		return fmt.Errorf("error setting notification_property: #{err}")
	}
	d.Set("name", job.Name)
	d.Set("role_arn", job.Role)

	tags, err := ListTags(conn, jobARN)

	if err != nil {
		return fmt.Errorf("error listing tags for Glue Job (%s): %s", jobARN, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("timeout", job.Timeout)
	if err := d.Set("security_configuration", job.SecurityConfiguration); err != nil {
		return fmt.Errorf("error setting security_configuration: %s", err)
	}

	d.Set("worker_type", job.WorkerType)
	d.Set("number_of_workers", job.NumberOfWorkers)

	return nil
}

func resourceJobUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	if d.HasChanges("command", "connections", "default_arguments", "description",
		"execution_property", "glue_version", "max_capacity", "max_retries", "notification_property", "number_of_workers",
		"role_arn", "security_configuration", "timeout", "worker_type", "non_overridable_arguments") {
		jobUpdate := &glue.JobUpdate{
			Command: expandJobCommand(d.Get("command").([]interface{})),
			Role:    aws.String(d.Get("role_arn").(string)),
		}

		if v, ok := d.GetOk("timeout"); ok {
			jobUpdate.Timeout = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("number_of_workers"); ok {
			jobUpdate.NumberOfWorkers = aws.Int64(int64(v.(int)))
		} else {
			if v, ok := d.GetOk("max_capacity"); ok {
				jobUpdate.MaxCapacity = aws.Float64(v.(float64))
			}
		}

		if v, ok := d.GetOk("connections"); ok {
			jobUpdate.Connections = &glue.ConnectionsList{
				Connections: flex.ExpandStringList(v.([]interface{})),
			}
		}

		if kv, ok := d.GetOk("default_arguments"); ok {
			jobUpdate.DefaultArguments = flex.ExpandStringMap(kv.(map[string]interface{}))
		}

		if kv, ok := d.GetOk("non_overridable_arguments"); ok {
			jobUpdate.NonOverridableArguments = flex.ExpandStringMap(kv.(map[string]interface{}))
		}

		if v, ok := d.GetOk("description"); ok {
			jobUpdate.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("glue_version"); ok {
			jobUpdate.GlueVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("execution_property"); ok {
			jobUpdate.ExecutionProperty = expandExecutionProperty(v.([]interface{}))
		}

		if v, ok := d.GetOk("max_retries"); ok {
			jobUpdate.MaxRetries = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("notification_property"); ok {
			jobUpdate.NotificationProperty = expandNotificationProperty(v.([]interface{}))
		}

		if v, ok := d.GetOk("security_configuration"); ok {
			jobUpdate.SecurityConfiguration = aws.String(v.(string))
		}

		if v, ok := d.GetOk("worker_type"); ok {
			jobUpdate.WorkerType = aws.String(v.(string))
		}

		input := &glue.UpdateJobInput{
			JobName:   aws.String(d.Id()),
			JobUpdate: jobUpdate,
		}

		log.Printf("[DEBUG] Updating Glue Job: %s", input)
		_, err := conn.UpdateJob(input)
		if err != nil {
			return fmt.Errorf("error updating Glue Job (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceJobRead(d, meta)
}

func resourceJobDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GlueConn

	log.Printf("[DEBUG] Deleting Glue Job: %s", d.Id())
	err := DeleteJob(conn, d.Id())
	if err != nil {
		return fmt.Errorf("error deleting Glue Job (%s): %s", d.Id(), err)
	}

	return nil
}

func DeleteJob(conn *glue.Glue, jobName string) error {
	input := &glue.DeleteJobInput{
		JobName: aws.String(jobName),
	}

	_, err := conn.DeleteJob(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
			return nil
		}
		return err
	}

	return nil
}

func expandExecutionProperty(l []interface{}) *glue.ExecutionProperty {
	m := l[0].(map[string]interface{})

	executionProperty := &glue.ExecutionProperty{
		MaxConcurrentRuns: aws.Int64(int64(m["max_concurrent_runs"].(int))),
	}

	return executionProperty
}

func expandJobCommand(l []interface{}) *glue.JobCommand {
	m := l[0].(map[string]interface{})

	jobCommand := &glue.JobCommand{
		Name:           aws.String(m["name"].(string)),
		ScriptLocation: aws.String(m["script_location"].(string)),
	}

	if v, ok := m["python_version"].(string); ok && v != "" {
		jobCommand.PythonVersion = aws.String(v)
	}

	return jobCommand
}

func expandNotificationProperty(l []interface{}) *glue.NotificationProperty {
	m := l[0].(map[string]interface{})

	notificationProperty := &glue.NotificationProperty{
		NotifyDelayAfter: aws.Int64(int64(m["notify_delay_after"].(int))),
	}

	return notificationProperty
}

func flattenConnectionsList(connectionsList *glue.ConnectionsList) []interface{} {
	if connectionsList == nil {
		return []interface{}{}
	}

	return flex.FlattenStringList(connectionsList.Connections)
}

func flattenExecutionProperty(executionProperty *glue.ExecutionProperty) []map[string]interface{} {
	if executionProperty == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"max_concurrent_runs": int(aws.Int64Value(executionProperty.MaxConcurrentRuns)),
	}

	return []map[string]interface{}{m}
}

func flattenJobCommand(jobCommand *glue.JobCommand) []map[string]interface{} {
	if jobCommand == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"name":            aws.StringValue(jobCommand.Name),
		"script_location": aws.StringValue(jobCommand.ScriptLocation),
		"python_version":  aws.StringValue(jobCommand.PythonVersion),
	}

	return []map[string]interface{}{m}
}

func flattenNotificationProperty(notificationProperty *glue.NotificationProperty) []map[string]interface{} {
	if notificationProperty == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"notify_delay_after": int(aws.Int64Value(notificationProperty.NotifyDelayAfter)),
	}

	return []map[string]interface{}{m}
}
