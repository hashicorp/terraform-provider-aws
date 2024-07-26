// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

// @SDKResource("aws_glue_job", name="Job")
// @Tags(identifierAttribute="arn")
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"command": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "glueetl",
						},
						"python_version": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"2", "3", "3.9"}, true),
						},
						"runtime": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"Ray2.4"}, true),
						},
						"script_location": {
							Type:     schema.TypeString,
							Required: true,
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
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
			},
			"execution_class": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ExecutionClass](),
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
			"glue_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrMaxCapacity: {
				Type:          schema.TypeFloat,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"number_of_workers", "worker_type"},
			},
			"maintenance_window": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"max_retries": {
				Type:         schema.TypeInt,
				Optional:     true,
				ValidateFunc: validation.IntBetween(0, 10),
			},
			names.AttrName: {
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
				Computed:      true,
				ConflictsWith: []string{names.AttrMaxCapacity},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTimeout: {
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
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ConflictsWith:    []string{names.AttrMaxCapacity},
				ValidateDiagFunc: enum.Validate[awstypes.WorkerType](),
			},
		},
	}
}

func resourceJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &glue.CreateJobInput{
		Command: expandJobCommand(d.Get("command").([]interface{})),
		Name:    aws.String(name),
		Role:    aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("connections"); ok {
		input.Connections = &awstypes.ConnectionsList{
			Connections: flex.ExpandStringValueList(v.([]interface{})),
		}
	}

	if v, ok := d.GetOk("default_arguments"); ok {
		input.DefaultArguments = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_class"); ok {
		input.ExecutionClass = awstypes.ExecutionClass(v.(string))
	}

	if v, ok := d.GetOk("execution_property"); ok {
		input.ExecutionProperty = expandExecutionProperty(v.([]interface{}))
	}

	if v, ok := d.GetOk("glue_version"); ok {
		input.GlueVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrMaxCapacity); ok {
		input.MaxCapacity = aws.Float64(v.(float64))
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		input.MaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_retries"); ok {
		input.MaxRetries = int32(v.(int))
	}

	if v, ok := d.GetOk("non_overridable_arguments"); ok {
		input.NonOverridableArguments = flex.ExpandStringValueMap(v.(map[string]interface{}))
	}

	if v, ok := d.GetOk("notification_property"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.NotificationProperty = expandNotificationProperty(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrTimeout); ok {
		input.Timeout = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = awstypes.WorkerType(v.(string))
	}

	output, err := conn.CreateJob(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Job (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Name))

	return append(diags, resourceJobRead(ctx, d, meta)...)
}

func resourceJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

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
	d.Set(names.AttrARN, jobARN)
	if err := d.Set("command", flattenJobCommand(job.Command)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting command: %s", err)
	}
	if err := d.Set("connections", flattenConnectionsList(job.Connections)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting connections: %s", err)
	}
	d.Set("default_arguments", job.DefaultArguments)
	d.Set(names.AttrDescription, job.Description)
	d.Set("execution_class", job.ExecutionClass)
	if err := d.Set("execution_property", flattenExecutionProperty(job.ExecutionProperty)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting execution_property: %s", err)
	}
	d.Set("glue_version", job.GlueVersion)
	d.Set("maintenance_window", job.MaintenanceWindow)
	d.Set(names.AttrMaxCapacity, job.MaxCapacity)
	d.Set("max_retries", job.MaxRetries)
	d.Set(names.AttrName, job.Name)
	d.Set("non_overridable_arguments", job.NonOverridableArguments)
	if err := d.Set("notification_property", flattenNotificationProperty(job.NotificationProperty)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting notification_property: %s", err)
	}
	d.Set("number_of_workers", job.NumberOfWorkers)
	d.Set(names.AttrRoleARN, job.Role)
	d.Set("security_configuration", job.SecurityConfiguration)
	d.Set(names.AttrTimeout, job.Timeout)
	d.Set("worker_type", job.WorkerType)

	return diags
}

func resourceJobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		jobUpdate := &awstypes.JobUpdate{
			Command: expandJobCommand(d.Get("command").([]interface{})),
			Role:    aws.String(d.Get(names.AttrRoleARN).(string)),
		}

		if v, ok := d.GetOk("connections"); ok {
			jobUpdate.Connections = &awstypes.ConnectionsList{
				Connections: flex.ExpandStringValueList(v.([]interface{})),
			}
		}

		if kv, ok := d.GetOk("default_arguments"); ok {
			jobUpdate.DefaultArguments = flex.ExpandStringValueMap(kv.(map[string]interface{}))
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			jobUpdate.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("execution_class"); ok {
			jobUpdate.ExecutionClass = awstypes.ExecutionClass(v.(string))
		}

		if v, ok := d.GetOk("execution_property"); ok {
			jobUpdate.ExecutionProperty = expandExecutionProperty(v.([]interface{}))
		}

		if v, ok := d.GetOk("glue_version"); ok {
			jobUpdate.GlueVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("maintenance_window"); ok {
			jobUpdate.MaintenanceWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_retries"); ok {
			jobUpdate.MaxRetries = int32(v.(int))
		}

		if kv, ok := d.GetOk("non_overridable_arguments"); ok {
			jobUpdate.NonOverridableArguments = flex.ExpandStringValueMap(kv.(map[string]interface{}))
		}

		if v, ok := d.GetOk("notification_property"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			jobUpdate.NotificationProperty = expandNotificationProperty(v.([]interface{})[0].(map[string]interface{}))
		}

		if v, ok := d.GetOk("number_of_workers"); ok {
			jobUpdate.NumberOfWorkers = aws.Int32(int32(v.(int)))
		} else {
			if v, ok := d.GetOk(names.AttrMaxCapacity); ok {
				jobUpdate.MaxCapacity = aws.Float64(v.(float64))
			}
		}

		if v, ok := d.GetOk("security_configuration"); ok {
			jobUpdate.SecurityConfiguration = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrTimeout); ok {
			jobUpdate.Timeout = aws.Int32(int32(v.(int)))
		}

		if v, ok := d.GetOk("worker_type"); ok {
			jobUpdate.WorkerType = awstypes.WorkerType(v.(string))
		}

		input := &glue.UpdateJobInput{
			JobName:   aws.String(d.Id()),
			JobUpdate: jobUpdate,
		}

		_, err := conn.UpdateJob(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Job (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceJobRead(ctx, d, meta)...)
}

func resourceJobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Deleting Glue Job: %s", d.Id())
	_, err := conn.DeleteJob(ctx, &glue.DeleteJobInput{
		JobName: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Job (%s): %s", d.Id(), err)
	}

	return diags
}

func expandExecutionProperty(l []interface{}) *awstypes.ExecutionProperty {
	m := l[0].(map[string]interface{})

	executionProperty := &awstypes.ExecutionProperty{
		MaxConcurrentRuns: int32(m["max_concurrent_runs"].(int)),
	}

	return executionProperty
}

func expandJobCommand(l []interface{}) *awstypes.JobCommand {
	m := l[0].(map[string]interface{})

	jobCommand := &awstypes.JobCommand{
		Name:           aws.String(m[names.AttrName].(string)),
		ScriptLocation: aws.String(m["script_location"].(string)),
	}

	if v, ok := m["python_version"].(string); ok && v != "" {
		jobCommand.PythonVersion = aws.String(v)
	}

	if v, ok := m["runtime"].(string); ok && v != "" {
		jobCommand.Runtime = aws.String(v)
	}

	return jobCommand
}

func expandNotificationProperty(tfMap map[string]interface{}) *awstypes.NotificationProperty {
	if tfMap == nil {
		return nil
	}

	notificationProperty := &awstypes.NotificationProperty{}

	if v, ok := tfMap["notify_delay_after"].(int); ok && v != 0 {
		notificationProperty.NotifyDelayAfter = aws.Int32(int32(v))
	}

	return notificationProperty
}

func flattenConnectionsList(connectionsList *awstypes.ConnectionsList) []interface{} {
	if connectionsList == nil {
		return []interface{}{}
	}

	return flex.FlattenStringValueList(connectionsList.Connections)
}

func flattenExecutionProperty(executionProperty *awstypes.ExecutionProperty) []map[string]interface{} {
	if executionProperty == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"max_concurrent_runs": int(executionProperty.MaxConcurrentRuns),
	}

	return []map[string]interface{}{m}
}

func flattenJobCommand(jobCommand *awstypes.JobCommand) []map[string]interface{} {
	if jobCommand == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		names.AttrName:    aws.ToString(jobCommand.Name),
		"script_location": aws.ToString(jobCommand.ScriptLocation),
		"python_version":  aws.ToString(jobCommand.PythonVersion),
		"runtime":         aws.ToString(jobCommand.Runtime),
	}

	return []map[string]interface{}{m}
}

func flattenNotificationProperty(notificationProperty *awstypes.NotificationProperty) []map[string]interface{} {
	if notificationProperty == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"notify_delay_after": int(aws.ToInt32(notificationProperty.NotifyDelayAfter)),
	}

	return []map[string]interface{}{m}
}
