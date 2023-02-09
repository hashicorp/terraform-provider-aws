package glue

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceJob() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceJobCreate,
		ReadWithoutTimeout:   resourceJobRead,
		UpdateWithoutTimeout: resourceJobUpdate,
		DeleteWithoutTimeout: resourceJobDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
							ValidateFunc: validation.StringInSlice([]string{"2", "3", "3.9"}, true),
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
			"execution_class": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(glue.ExecutionClass_Values(), true),
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
			"non_overridable_arguments": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			"number_of_workers": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"max_capacity"},
				ValidateFunc:  validation.IntAtLeast(2),
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
		},
	}
}

func resourceJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &glue.CreateJobInput{
		Command: expandJobCommand(d.Get("command").([]interface{})),
		Name:    aws.String(name),
		Role:    aws.String(d.Get("role_arn").(string)),
		Tags:    Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("connections"); ok {
		input.Connections = &glue.ConnectionsList{
			Connections: flex.ExpandStringList(v.([]interface{})),
		}
	}

	if v, ok := d.GetOk("default_arguments"); ok {
		input.DefaultArguments = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_class"); ok {
		input.ExecutionClass = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_property"); ok {
		input.ExecutionProperty = expandExecutionProperty(v.([]interface{}))
	}

	if v, ok := d.GetOk("glue_version"); ok {
		input.GlueVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_capacity"); ok {
		input.MaxCapacity = aws.Float64(v.(float64))
	}

	if v, ok := d.GetOk("max_retries"); ok {
		input.MaxRetries = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("non_overridable_arguments"); ok {
		input.NonOverridableArguments = flex.ExpandStringMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("notification_property"); ok {
		input.NotificationProperty = expandNotificationProperty(v.([]interface{}))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("timeout"); ok {
		input.Timeout = aws.Int64(int64(v.(int)))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating Glue Job: %s", input)
	output, err := conn.CreateJobWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Job (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Name))

	return append(diags, resourceJobRead(ctx, d, meta)...)
}

func resourceJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	job, err := FindJobByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Glue Job (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Job (%s): %s", d.Id(), err)
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
		return sdkdiag.AppendErrorf(diags, "setting command: %s", err)
	}
	if err := d.Set("connections", flattenConnectionsList(job.Connections)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting connections: %s", err)
	}
	d.Set("default_arguments", aws.StringValueMap(job.DefaultArguments))
	d.Set("description", job.Description)
	d.Set("execution_class", job.ExecutionClass)
	if err := d.Set("execution_property", flattenExecutionProperty(job.ExecutionProperty)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting execution_property: %s", err)
	}
	d.Set("glue_version", job.GlueVersion)
	d.Set("max_capacity", job.MaxCapacity)
	d.Set("max_retries", job.MaxRetries)
	d.Set("name", job.Name)
	d.Set("non_overridable_arguments", aws.StringValueMap(job.NonOverridableArguments))
	if err := d.Set("notification_property", flattenNotificationProperty(job.NotificationProperty)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting notification_property: %s", err)
	}
	d.Set("number_of_workers", job.NumberOfWorkers)
	d.Set("role_arn", job.Role)
	d.Set("security_configuration", job.SecurityConfiguration)
	d.Set("timeout", job.Timeout)
	d.Set("worker_type", job.WorkerType)

	tags, err := ListTags(ctx, conn, jobARN)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing tags for Glue Job (%s): %s", jobARN, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceJobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	if d.HasChangesExcept("tags", "tags_all") {
		jobUpdate := &glue.JobUpdate{
			Command: expandJobCommand(d.Get("command").([]interface{})),
			Role:    aws.String(d.Get("role_arn").(string)),
		}

		if v, ok := d.GetOk("connections"); ok {
			jobUpdate.Connections = &glue.ConnectionsList{
				Connections: flex.ExpandStringList(v.([]interface{})),
			}
		}

		if kv, ok := d.GetOk("default_arguments"); ok {
			jobUpdate.DefaultArguments = flex.ExpandStringMap(kv.(map[string]interface{}))
		}

		if v, ok := d.GetOk("description"); ok {
			jobUpdate.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("execution_class"); ok {
			jobUpdate.ExecutionClass = aws.String(v.(string))
		}

		if v, ok := d.GetOk("execution_property"); ok {
			jobUpdate.ExecutionProperty = expandExecutionProperty(v.([]interface{}))
		}

		if v, ok := d.GetOk("glue_version"); ok {
			jobUpdate.GlueVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_retries"); ok {
			jobUpdate.MaxRetries = aws.Int64(int64(v.(int)))
		}

		if kv, ok := d.GetOk("non_overridable_arguments"); ok {
			jobUpdate.NonOverridableArguments = flex.ExpandStringMap(kv.(map[string]interface{}))
		}

		if v, ok := d.GetOk("notification_property"); ok {
			jobUpdate.NotificationProperty = expandNotificationProperty(v.([]interface{}))
		}

		if v, ok := d.GetOk("number_of_workers"); ok {
			jobUpdate.NumberOfWorkers = aws.Int64(int64(v.(int)))
		} else {
			if v, ok := d.GetOk("max_capacity"); ok {
				jobUpdate.MaxCapacity = aws.Float64(v.(float64))
			}
		}

		if v, ok := d.GetOk("security_configuration"); ok {
			jobUpdate.SecurityConfiguration = aws.String(v.(string))
		}

		if v, ok := d.GetOk("timeout"); ok {
			jobUpdate.Timeout = aws.Int64(int64(v.(int)))
		}

		if v, ok := d.GetOk("worker_type"); ok {
			jobUpdate.WorkerType = aws.String(v.(string))
		}

		input := &glue.UpdateJobInput{
			JobName:   aws.String(d.Id()),
			JobUpdate: jobUpdate,
		}

		log.Printf("[DEBUG] Updating Glue Job: %s", input)
		_, err := conn.UpdateJobWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Job (%s): %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating tags: %s", err)
		}
	}

	return append(diags, resourceJobRead(ctx, d, meta)...)
}

func resourceJobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueConn()

	log.Printf("[DEBUG] Deleting Glue Job: %s", d.Id())
	_, err := conn.DeleteJobWithContext(ctx, &glue.DeleteJobInput{
		JobName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, glue.ErrCodeEntityNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Job (%s): %s", d.Id(), err)
	}

	return diags
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
