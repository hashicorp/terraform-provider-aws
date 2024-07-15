// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dms_replication_task", name="Replication Task")
// @Tags(identifierAttribute="replication_task_arn")
func ResourceReplicationTask() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceReplicationTaskCreate,
		ReadWithoutTimeout:   resourceReplicationTaskRead,
		UpdateWithoutTimeout: resourceReplicationTaskUpdate,
		DeleteWithoutTimeout: resourceReplicationTaskDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"cdc_start_position": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"cdc_start_time"},
			},
			"cdc_start_time": {
				Type:          schema.TypeString,
				Optional:      true,
				ValidateFunc:  verify.ValidStringDateOrPositiveInt,
				ConflictsWith: []string{"cdc_start_position"},
			},
			"migration_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(dms.MigrationTypeValue_Values(), false),
			},
			"replication_instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"replication_task_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"replication_task_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validReplicationTaskID,
			},
			// "replication_task_settings" is equivalent to "replication_settings" on "aws_dms_replication_config"
			// All changes to this field and supporting tests should be mirrored in "aws_dms_replication_config"
			"replication_task_settings": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateDiagFunc: validation.AllDiag(
					validation.ToDiagFunc(validation.StringIsJSON),
					validateReplicationSettings,
				),
				DiffSuppressFunc:      suppressEquivalentTaskSettings,
				DiffSuppressOnRefresh: true,
			},
			"resource_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 31),
					validation.StringMatch(regexache.MustCompile("^[A-Za-z][0-9A-Za-z-]+$"), "must start with a letter, only contain alphanumeric characters and hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`--`), "cannot contain two consecutive hyphens"),
					validation.StringDoesNotMatch(regexache.MustCompile(`-$`), "cannot end in a hyphen"),
				),
			},
			"source_endpoint_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"start_replication_task": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"table_mappings": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_endpoint_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceReplicationTaskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	taskID := d.Get("replication_task_id").(string)
	input := &dms.CreateReplicationTaskInput{
		MigrationType:             aws.String(d.Get("migration_type").(string)),
		ReplicationInstanceArn:    aws.String(d.Get("replication_instance_arn").(string)),
		ReplicationTaskIdentifier: aws.String(taskID),
		SourceEndpointArn:         aws.String(d.Get("source_endpoint_arn").(string)),
		TableMappings:             aws.String(d.Get("table_mappings").(string)),
		Tags:                      getTagsIn(ctx),
		TargetEndpointArn:         aws.String(d.Get("target_endpoint_arn").(string)),
	}

	if v, ok := d.GetOk("cdc_start_position"); ok {
		input.CdcStartPosition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cdc_start_time"); ok {
		v := v.(string)
		if t, err := time.Parse(time.RFC3339, v); err != nil {
			input.CdcStartTime = aws.Time(time.Unix(flex.StringValueToInt64Value(v), 0))
		} else {
			input.CdcStartTime = aws.Time(t)
		}
	}

	if v, ok := d.GetOk("replication_task_settings"); ok {
		input.ReplicationTaskSettings = aws.String(v.(string))
	}

	if v, ok := d.GetOk("resource_identifier"); ok {
		input.ResourceIdentifier = aws.String(v.(string))
	}

	_, err := conn.CreateReplicationTaskWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Replication Task (%s): %s", taskID, err)
	}

	d.SetId(taskID)

	if _, err := waitReplicationTaskReady(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Task (%s) create: %s", d.Id(), err)
	}

	if d.Get("start_replication_task").(bool) {
		if err := startReplicationTask(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceReplicationTaskRead(ctx, d, meta)...)
}

func resourceReplicationTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	task, err := FindReplicationTaskByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Replication Task (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Task (%s): %s", d.Id(), err)
	}

	d.Set("cdc_start_position", task.CdcStartPosition)
	d.Set("migration_type", task.MigrationType)
	d.Set("replication_instance_arn", task.ReplicationInstanceArn)
	d.Set("replication_task_arn", task.ReplicationTaskArn)
	d.Set("replication_task_id", task.ReplicationTaskIdentifier)
	d.Set("replication_task_settings", task.ReplicationTaskSettings)
	d.Set("source_endpoint_arn", task.SourceEndpointArn)
	d.Set(names.AttrStatus, task.Status)
	d.Set("table_mappings", task.TableMappings)
	d.Set("target_endpoint_arn", task.TargetEndpointArn)

	return diags
}

func resourceReplicationTaskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll, "replication_instance_arn", "start_replication_task") {
		if err := stopReplicationTask(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &dms.ModifyReplicationTaskInput{
			MigrationType:      aws.String(d.Get("migration_type").(string)),
			ReplicationTaskArn: aws.String(d.Get("replication_task_arn").(string)),
			TableMappings:      aws.String(d.Get("table_mappings").(string)),
		}

		if d.HasChange("cdc_start_position") {
			input.CdcStartPosition = aws.String(d.Get("cdc_start_position").(string))
		}

		if d.HasChange("cdc_start_time") {
			if v, ok := d.GetOk("cdc_start_time"); ok {
				v := v.(string)
				if t, err := time.Parse(time.RFC3339, v); err != nil {
					input.CdcStartTime = aws.Time(time.Unix(flex.StringValueToInt64Value(v), 0))
				} else {
					input.CdcStartTime = aws.Time(t)
				}
			}
		}

		if d.HasChange("replication_task_settings") {
			if v, ok := d.GetOk("replication_task_settings"); ok {
				s, err := normalizeReplicationSettings(v.(string))
				if err != nil {
					return sdkdiag.AppendErrorf(diags, "updating DMS Replication Task (%s): %s", d.Id(), err)
				}
				input.ReplicationTaskSettings = aws.String(s)
			}
		}

		_, err := conn.ModifyReplicationTaskWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying DMS Replication Task (%s): %s", d.Id(), err)
		}

		if _, err := waitReplicationTaskModified(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Task (%s) update: %s", d.Id(), err)
		}

		if d.Get("start_replication_task").(bool) {
			if err := startReplicationTask(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChange("replication_instance_arn") {
		if err := stopReplicationTask(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &dms.MoveReplicationTaskInput{
			ReplicationTaskArn:           aws.String(d.Get("replication_task_arn").(string)),
			TargetReplicationInstanceArn: aws.String(d.Get("replication_instance_arn").(string)),
		}

		_, err := conn.MoveReplicationTaskWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "moving DMS Replication Task (%s): %s", d.Id(), err)
		}

		if _, err := waitReplicationTaskMoved(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Task (%s) update: %s", d.Id(), err)
		}

		if d.Get("start_replication_task").(bool) {
			if err := startReplicationTask(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChanges("start_replication_task") {
		var f func(context.Context, *dms.DatabaseMigrationService, string) error
		if d.Get("start_replication_task").(bool) {
			f = startReplicationTask
		} else {
			f = stopReplicationTask
		}
		if err := f(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceReplicationTaskRead(ctx, d, meta)...)
}

func resourceReplicationTaskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if err := stopReplicationTask(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting DMS Replication Task: %s", d.Id())
	_, err := conn.DeleteReplicationTaskWithContext(ctx, &dms.DeleteReplicationTaskInput{
		ReplicationTaskArn: aws.String(d.Get("replication_task_arn").(string)),
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Replication Task (%s): %s", d.Id(), err)
	}

	if _, err := waitReplicationTaskDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Task (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func FindReplicationTaskByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationTask, error) {
	input := &dms.DescribeReplicationTasksInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-task-id"),
				Values: aws.StringSlice([]string{id}),
			},
		},
	}

	return findReplicationTask(ctx, conn, input)
}

func findReplicationTask(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeReplicationTasksInput) (*dms.ReplicationTask, error) {
	output, err := findReplicationTasks(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSinglePtrResult(output)
}

func findReplicationTasks(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeReplicationTasksInput) ([]*dms.ReplicationTask, error) {
	var output []*dms.ReplicationTask

	err := conn.DescribeReplicationTasksPagesWithContext(ctx, input, func(page *dms.DescribeReplicationTasksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ReplicationTasks {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func statusReplicationTask(ctx context.Context, conn *dms.DatabaseMigrationService, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindReplicationTaskByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func setLastReplicationTaskError(err error, replication *dms.ReplicationTask) {
	var errs []error

	if v := aws.StringValue(replication.LastFailureMessage); v != "" {
		errs = append(errs, errors.New(v))
	}
	if v := aws.StringValue(replication.StopReason); v != "" {
		errs = append(errs, errors.New(v))
	}

	tfresource.SetLastError(err, errors.Join(errs...))
}

func waitReplicationTaskDeleted(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) (*dms.ReplicationTask, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusDeleting},
		Target:     []string{},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationTask); ok {
		setLastReplicationTaskError(err, output)
		return output, err
	}

	return nil, err
}

func waitReplicationTaskModified(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) (*dms.ReplicationTask, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusModifying},
		Target:     []string{replicationTaskStatusReady, replicationTaskStatusStopped, replicationTaskStatusFailed},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationTask); ok {
		setLastReplicationTaskError(err, output)
		return output, err
	}

	return nil, err
}

func waitReplicationTaskMoved(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) (*dms.ReplicationTask, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusModifying, replicationTaskStatusMoving},
		Target:     []string{replicationTaskStatusReady, replicationTaskStatusStopped, replicationTaskStatusFailed},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationTask); ok {
		setLastReplicationTaskError(err, output)
		return output, err
	}

	return nil, err
}

func waitReplicationTaskReady(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) (*dms.ReplicationTask, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusCreating},
		Target:     []string{replicationTaskStatusReady},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationTask); ok {
		setLastReplicationTaskError(err, output)
		return output, err
	}

	return nil, err
}

func waitReplicationTaskRunning(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationTask, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusStarting},
		Target:     []string{replicationTaskStatusRunning},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationTask); ok {
		setLastReplicationTaskError(err, output)
		return output, err
	}

	return nil, err
}

func waitReplicationTaskStopped(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationTask, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{replicationTaskStatusStopping, replicationTaskStatusRunning},
		Target:                    []string{replicationTaskStatusStopped},
		Refresh:                   statusReplicationTask(ctx, conn, id),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		Delay:                     60 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationTask); ok {
		setLastReplicationTaskError(err, output)
		return output, err
	}

	return nil, err
}

func waitReplicationTaskSteady(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationTask, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{replicationTaskStatusCreating, replicationTaskStatusDeleting, replicationTaskStatusModifying, replicationTaskStatusStopping, replicationTaskStatusStarting},
		Target:                    []string{replicationTaskStatusFailed, replicationTaskStatusReady, replicationTaskStatusStopped, replicationTaskStatusRunning},
		Refresh:                   statusReplicationTask(ctx, conn, id),
		Timeout:                   timeout,
		MinTimeout:                10 * time.Second,
		Delay:                     60 * time.Second,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dms.ReplicationTask); ok {
		setLastReplicationTaskError(err, output)
		return output, err
	}

	return nil, err
}

func startReplicationTask(ctx context.Context, conn *dms.DatabaseMigrationService, id string) error {
	task, err := FindReplicationTaskByID(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("reading DMS Replication Task (%s): %w", id, err)
	}

	taskStatus := aws.StringValue(task.Status)
	if taskStatus == replicationTaskStatusRunning {
		return nil
	}

	startReplicationTaskType := dms.StartReplicationTaskTypeValueStartReplication
	if taskStatus != replicationTaskStatusReady {
		startReplicationTaskType = dms.StartReplicationTaskTypeValueResumeProcessing
	}
	input := &dms.StartReplicationTaskInput{
		ReplicationTaskArn:       task.ReplicationTaskArn,
		StartReplicationTaskType: aws.String(startReplicationTaskType),
	}

	_, err = conn.StartReplicationTaskWithContext(ctx, input)

	if err != nil {
		return fmt.Errorf("starting DMS Replication Task (%s): %w", id, err)
	}

	if _, err := waitReplicationTaskRunning(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for DMS Replication Task (%s) start: %w", id, err)
	}

	return nil
}

func stopReplicationTask(ctx context.Context, conn *dms.DatabaseMigrationService, id string) error {
	task, err := FindReplicationTaskByID(ctx, conn, id)

	if tfresource.NotFound(err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading DMS Replication Task (%s): %w", id, err)
	}

	taskStatus := aws.StringValue(task.Status)
	if taskStatus != replicationTaskStatusRunning {
		return nil
	}

	input := &dms.StopReplicationTaskInput{
		ReplicationTaskArn: task.ReplicationTaskArn,
	}

	_, err = conn.StopReplicationTaskWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, dms.ErrCodeInvalidResourceStateFault, "is currently not running") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("stopping DMS Replication Task (%s): %w", id, err)
	}

	if _, err := waitReplicationTaskStopped(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for DMS Replication Task (%s) stop: %w", id, err)
	}

	return nil
}
