// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package glue

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_glue_job", name="Job")
// @Tags(identifierAttribute="arn")
func resourceJob() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceJobCreate,
		ReadWithoutTimeout:   resourceJobRead,
		UpdateWithoutTimeout: resourceJobUpdate,
		DeleteWithoutTimeout: resourceJobDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: resourceJobCustomizeDiff,

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
							Default:  jobCommandNameApacheSparkETL,
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
			"job_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[awstypes.JobMode](),
			},
			"job_run_queuing_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
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
			"security_configuration": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"source_control_details": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth_strategy": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SourceControlAuthStrategy](),
						},
						"auth_token": {
							Type:      schema.TypeString,
							Optional:  true,
							Sensitive: true,
						},
						"branch": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"folder": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"last_commit_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrOwner: {
							Type:     schema.TypeString,
							Optional: true,
						},
						"provider": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SourceControlProvider](),
						},
						"repository": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrTimeout: {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"worker_type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{names.AttrMaxCapacity},
				ValidateFunc:  validation.StringInSlice(workerType_Values(), false),
			},
		},
	}
}

func resourceJobCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := glue.CreateJobInput{
		Command: expandJobCommand(d.Get("command").([]any)),
		Name:    aws.String(name),
		Role:    aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:    getTagsIn(ctx),
	}

	if v, ok := d.GetOk("connections"); ok {
		input.Connections = &awstypes.ConnectionsList{
			Connections: flex.ExpandStringValueList(v.([]any)),
		}
	}

	if v, ok := d.GetOk("default_arguments"); ok {
		input.DefaultArguments = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("execution_class"); ok {
		input.ExecutionClass = awstypes.ExecutionClass(v.(string))
	}

	if v, ok := d.GetOk("execution_property"); ok {
		input.ExecutionProperty = expandExecutionProperty(v.([]any))
	}

	if v, ok := d.GetOk("glue_version"); ok {
		input.GlueVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("job_mode"); ok {
		input.JobMode = awstypes.JobMode(v.(string))
	}

	if v, ok := d.GetOk("job_run_queuing_enabled"); ok {
		input.JobRunQueuingEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("maintenance_window"); ok {
		input.MaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrMaxCapacity); ok {
		input.MaxCapacity = aws.Float64(v.(float64))
	}

	if v, ok := d.GetOk("max_retries"); ok {
		input.MaxRetries = int32(v.(int))
	}

	if v, ok := d.GetOk("non_overridable_arguments"); ok {
		input.NonOverridableArguments = flex.ExpandStringValueMap(v.(map[string]any))
	}

	if v, ok := d.GetOk("notification_property"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.NotificationProperty = expandNotificationProperty(v.([]any)[0].(map[string]any))
	}

	if v, ok := d.GetOk("number_of_workers"); ok {
		input.NumberOfWorkers = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("security_configuration"); ok {
		input.SecurityConfiguration = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_control_details"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.SourceControlDetails = expandSourceControlDetails(v.([]any))
	}

	if v, ok := d.GetOk(names.AttrTimeout); ok {
		input.Timeout = aws.Int32(int32(v.(int)))
	}

	if v, ok := d.GetOk("worker_type"); ok {
		input.WorkerType = awstypes.WorkerType(v.(string))
	}

	output, err := conn.CreateJob(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Glue Job (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Name))

	return append(diags, resourceJobRead(ctx, d, meta)...)
}

func resourceJobRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	job, err := findJobByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Glue Job (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Glue Job (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, jobARN(ctx, meta.(*conns.AWSClient), d.Id()))
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
	d.Set("job_mode", job.JobMode)
	d.Set("job_run_queuing_enabled", job.JobRunQueuingEnabled)
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
	if err := d.Set("source_control_details", flattenSourceControlDetails(job.SourceControlDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting source_control_details: %s", err)
	}
	d.Set(names.AttrTimeout, job.Timeout)
	d.Set("worker_type", job.WorkerType)

	return diags
}

func resourceJobUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		jobUpdate := &awstypes.JobUpdate{
			Command: expandJobCommand(d.Get("command").([]any)),
			Role:    aws.String(d.Get(names.AttrRoleARN).(string)),
		}

		if v, ok := d.GetOk("connections"); ok {
			jobUpdate.Connections = &awstypes.ConnectionsList{
				Connections: flex.ExpandStringValueList(v.([]any)),
			}
		}

		if kv, ok := d.GetOk("default_arguments"); ok {
			jobUpdate.DefaultArguments = flex.ExpandStringValueMap(kv.(map[string]any))
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			jobUpdate.Description = aws.String(v.(string))
		}

		if v, ok := d.GetOk("execution_class"); ok {
			jobUpdate.ExecutionClass = awstypes.ExecutionClass(v.(string))
		}

		if v, ok := d.GetOk("execution_property"); ok {
			jobUpdate.ExecutionProperty = expandExecutionProperty(v.([]any))
		}

		if v, ok := d.GetOk("glue_version"); ok {
			jobUpdate.GlueVersion = aws.String(v.(string))
		}

		if v, ok := d.GetOk("job_mode"); ok {
			jobUpdate.JobMode = awstypes.JobMode(v.(string))
		}

		if v, ok := d.GetOk("job_run_queuing_enabled"); ok {
			jobUpdate.JobRunQueuingEnabled = aws.Bool(v.(bool))
		}

		if v, ok := d.GetOk("maintenance_window"); ok {
			jobUpdate.MaintenanceWindow = aws.String(v.(string))
		}

		if v, ok := d.GetOk("max_retries"); ok {
			jobUpdate.MaxRetries = int32(v.(int))
		}

		if kv, ok := d.GetOk("non_overridable_arguments"); ok {
			jobUpdate.NonOverridableArguments = flex.ExpandStringValueMap(kv.(map[string]any))
		}

		if v, ok := d.GetOk("notification_property"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			jobUpdate.NotificationProperty = expandNotificationProperty(v.([]any)[0].(map[string]any))
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

		if v, ok := d.GetOk("source_control_details"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			jobUpdate.SourceControlDetails = expandSourceControlDetails(v.([]any))
		}

		if !strings.EqualFold(jobCommandNameRay, aws.ToString(jobUpdate.Command.Name)) {
			if v, ok := d.GetOk(names.AttrTimeout); ok {
				jobUpdate.Timeout = aws.Int32(int32(v.(int)))
			}
		}

		if v, ok := d.GetOk("worker_type"); ok {
			jobUpdate.WorkerType = awstypes.WorkerType(v.(string))
		}

		input := glue.UpdateJobInput{
			JobName:   aws.String(d.Id()),
			JobUpdate: jobUpdate,
		}

		_, err := conn.UpdateJob(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Glue Job (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceJobRead(ctx, d, meta)...)
}

func resourceJobDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GlueClient(ctx)

	log.Printf("[DEBUG] Deleting Glue Job: %s", d.Id())
	input := glue.DeleteJobInput{
		JobName: aws.String(d.Id()),
	}
	_, err := conn.DeleteJob(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Glue Job (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceJobCustomizeDiff(_ context.Context, diff *schema.ResourceDiff, v any) error {
	if command := expandJobCommand(diff.Get("command").([]any)); command != nil {
		commandName := aws.ToString(command.Name)
		// Allow 0 timeout for streaming jobs.
		var minVal int64
		if !strings.EqualFold(jobCommandNameApacheSparkStreamingETL, commandName) {
			minVal = 1
		}

		key := names.AttrTimeout
		if v := diff.GetRawConfig().GetAttr(key); v.IsKnown() && !v.IsNull() {
			// InvalidInputException: Timeout not supported for Ray jobs.
			if strings.EqualFold(jobCommandNameRay, commandName) {
				return fmt.Errorf("%s must not be configured for Ray jobs", key)
			}

			if v, _ := v.AsBigFloat().Int64(); v < minVal {
				return fmt.Errorf("expected %s to be at least (%d), got %d", key, minVal, v)
			}
		}
	}

	return nil
}

func findJobByName(ctx context.Context, conn *glue.Client, name string) (*awstypes.Job, error) {
	input := glue.GetJobInput{
		JobName: aws.String(name),
	}

	output, err := conn.GetJob(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Job == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output.Job, nil
}

func expandExecutionProperty(l []any) *awstypes.ExecutionProperty {
	m := l[0].(map[string]any)

	executionProperty := &awstypes.ExecutionProperty{
		MaxConcurrentRuns: int32(m["max_concurrent_runs"].(int)),
	}

	return executionProperty
}

func expandJobCommand(l []any) *awstypes.JobCommand {
	m := l[0].(map[string]any)

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

func expandNotificationProperty(tfMap map[string]any) *awstypes.NotificationProperty {
	if tfMap == nil {
		return nil
	}

	notificationProperty := &awstypes.NotificationProperty{}

	if v, ok := tfMap["notify_delay_after"].(int); ok && v != 0 {
		notificationProperty.NotifyDelayAfter = aws.Int32(int32(v))
	}

	return notificationProperty
}

func flattenConnectionsList(connectionsList *awstypes.ConnectionsList) []any {
	if connectionsList == nil {
		return []any{}
	}

	return flex.FlattenStringValueList(connectionsList.Connections)
}

func flattenExecutionProperty(executionProperty *awstypes.ExecutionProperty) []map[string]any {
	if executionProperty == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"max_concurrent_runs": int(executionProperty.MaxConcurrentRuns),
	}

	return []map[string]any{m}
}

func flattenJobCommand(jobCommand *awstypes.JobCommand) []map[string]any {
	if jobCommand == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		names.AttrName:    aws.ToString(jobCommand.Name),
		"script_location": aws.ToString(jobCommand.ScriptLocation),
		"python_version":  aws.ToString(jobCommand.PythonVersion),
		"runtime":         aws.ToString(jobCommand.Runtime),
	}

	return []map[string]any{m}
}

func flattenNotificationProperty(notificationProperty *awstypes.NotificationProperty) []map[string]any {
	if notificationProperty == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"notify_delay_after": int(aws.ToInt32(notificationProperty.NotifyDelayAfter)),
	}

	return []map[string]any{m}
}

func expandSourceControlDetails(l []any) *awstypes.SourceControlDetails {
	m := l[0].(map[string]any)

	sourceControlDetails := &awstypes.SourceControlDetails{}

	if v, ok := m["auth_token"].(string); ok && v != "" {
		sourceControlDetails.AuthToken = aws.String(v)
	}

	if v, ok := m["folder"].(string); ok && v != "" {
		sourceControlDetails.Folder = aws.String(v)
	}
	if v, ok := m["auth_strategy"].(string); ok && v != "" {
		sourceControlDetails.AuthStrategy = awstypes.SourceControlAuthStrategy(v)
	}
	if v, ok := m["branch"].(string); ok && v != "" {
		sourceControlDetails.Branch = aws.String(v)
	}
	if v, ok := m["last_commit_id"].(string); ok && v != "" {
		sourceControlDetails.LastCommitId = aws.String(v)
	}
	if v, ok := m[names.AttrOwner].(string); ok && v != "" {
		sourceControlDetails.Owner = aws.String(v)
	}
	if v, ok := m["provider"].(string); ok && v != "" {
		sourceControlDetails.Provider = awstypes.SourceControlProvider(v)
	}
	if v, ok := m["repository"].(string); ok && v != "" {
		sourceControlDetails.Repository = aws.String(v)
	}

	return sourceControlDetails
}

func flattenSourceControlDetails(sourceControlDetails *awstypes.SourceControlDetails) []map[string]any {
	if sourceControlDetails == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"auth_strategy":  sourceControlDetails.AuthStrategy,
		"branch":         aws.ToString(sourceControlDetails.Branch),
		"folder":         aws.ToString(sourceControlDetails.Folder),
		"last_commit_id": aws.ToString(sourceControlDetails.LastCommitId),
		names.AttrOwner:  aws.ToString(sourceControlDetails.Owner),
		"provider":       sourceControlDetails.Provider,
		"repository":     aws.ToString(sourceControlDetails.Repository),
	}

	return []map[string]any{m}
}

func workerType_Values() []string {
	return tfslices.AppendUnique(enum.Values[awstypes.WorkerType](), "G.12X", "G.16X", "R.1X", "R.2X", "R.4X", "R.8X")
}

func jobARN(ctx context.Context, c *conns.AWSClient, jobID string) string {
	return c.RegionalARN(ctx, "glue", "job/"+jobID)
}
