// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
				Type:     schema.TypeString,
				Optional: true,
				// Requires a Unix timestamp in seconds. Example 1484346880
				ConflictsWith: []string{"cdc_start_position"},
			},
			"migration_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					dms.MigrationTypeValueFullLoad,
					dms.MigrationTypeValueCdc,
					dms.MigrationTypeValueFullLoadAndCdc,
				}, false),
			},
			"replication_instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
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
			"replication_task_settings": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
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
			"status": {
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

	taskId := d.Get("replication_task_id").(string)

	request := &dms.CreateReplicationTaskInput{
		MigrationType:             aws.String(d.Get("migration_type").(string)),
		ReplicationInstanceArn:    aws.String(d.Get("replication_instance_arn").(string)),
		ReplicationTaskIdentifier: aws.String(taskId),
		SourceEndpointArn:         aws.String(d.Get("source_endpoint_arn").(string)),
		TableMappings:             aws.String(d.Get("table_mappings").(string)),
		Tags:                      getTagsIn(ctx),
		TargetEndpointArn:         aws.String(d.Get("target_endpoint_arn").(string)),
	}

	if v, ok := d.GetOk("cdc_start_position"); ok {
		request.CdcStartPosition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cdc_start_time"); ok {
		seconds, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "DMS create replication task. Invalid CDC Unix timestamp: %s", err)
		}
		request.CdcStartTime = aws.Time(time.Unix(seconds, 0))
	}

	if v, ok := d.GetOk("replication_task_settings"); ok {
		request.ReplicationTaskSettings = aws.String(v.(string))
	}

	log.Println("[DEBUG] DMS create replication task:", request)

	_, err := conn.CreateReplicationTaskWithContext(ctx, request)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DMS Replication Task (%s): %s", taskId, err)
	}

	d.SetId(taskId)

	if err := waitReplicationTaskReady(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Task (%s) to become available: %s", d.Id(), err)
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

	if task == nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Task (%s): empty output", d.Id())
	}

	d.Set("cdc_start_position", task.CdcStartPosition)
	d.Set("migration_type", task.MigrationType)
	d.Set("replication_instance_arn", task.ReplicationInstanceArn)
	d.Set("replication_task_arn", task.ReplicationTaskArn)
	d.Set("replication_task_id", task.ReplicationTaskIdentifier)
	d.Set("source_endpoint_arn", task.SourceEndpointArn)
	d.Set("status", task.Status)
	d.Set("table_mappings", task.TableMappings)
	d.Set("target_endpoint_arn", task.TargetEndpointArn)

	settings, err := replicationTaskRemoveReadOnlySettings(aws.StringValue(task.ReplicationTaskSettings))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DMS Replication Task (%s): %s", d.Id(), err)
	}

	d.Set("replication_task_settings", settings)

	return diags
}

func resourceReplicationTaskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if d.HasChangesExcept("tags", "tags_all", "start_replication_task") {
		input := &dms.ModifyReplicationTaskInput{
			ReplicationTaskArn: aws.String(d.Get("replication_task_arn").(string)),
			MigrationType:      aws.String(d.Get("migration_type").(string)),
			TableMappings:      aws.String(d.Get("table_mappings").(string)),
		}

		if d.HasChange("cdc_start_position") {
			input.CdcStartPosition = aws.String(d.Get("cdc_start_position").(string))
		}

		if d.HasChange("cdc_start_time") {
			seconds, err := strconv.ParseInt(d.Get("cdc_start_time").(string), 10, 64)
			if err != nil {
				return sdkdiag.AppendErrorf(diags, "DMS update replication task. Invalid CRC Unix timestamp: %s", err)
			}
			input.CdcStartTime = aws.Time(time.Unix(seconds, 0))
		}

		if d.HasChange("replication_task_settings") {
			if v, ok := d.Get("replication_task_settings").(string); ok && v != "" {
				input.ReplicationTaskSettings = aws.String(v)
			} else {
				input.ReplicationTaskSettings = nil
			}
		}

		status := d.Get("status").(string)
		if status == replicationTaskStatusRunning {
			log.Println("[DEBUG] stopping DMS replication task:", input)
			if err := stopReplicationTask(ctx, d.Id(), conn); err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}

		log.Println("[DEBUG] updating DMS replication task:", input)
		_, err := conn.ModifyReplicationTaskWithContext(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DMS Replication Task (%s): %s", d.Id(), err)
		}

		if err := waitReplicationTaskModified(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Task (%s) update: %s", d.Id(), err)
		}

		if d.Get("start_replication_task").(bool) {
			err := startReplicationTask(ctx, conn, d.Id())
			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}
		}
	}

	if d.HasChanges("start_replication_task") {
		status := d.Get("status").(string)
		if d.Get("start_replication_task").(bool) {
			if status != replicationTaskStatusRunning {
				if err := startReplicationTask(ctx, conn, d.Id()); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		} else {
			if status == replicationTaskStatusRunning {
				if err := stopReplicationTask(ctx, d.Id(), conn); err != nil {
					return sdkdiag.AppendFromErr(diags, err)
				}
			}
		}
	}

	return append(diags, resourceReplicationTaskRead(ctx, d, meta)...)
}

func resourceReplicationTaskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DMSConn(ctx)

	if status := d.Get("status").(string); status == replicationTaskStatusRunning {
		if err := stopReplicationTask(ctx, d.Id(), conn); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	input := &dms.DeleteReplicationTaskInput{
		ReplicationTaskArn: aws.String(d.Get("replication_task_arn").(string)),
	}

	log.Printf("[DEBUG] DMS delete replication task: %#v", input)

	_, err := conn.DeleteReplicationTaskWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DMS Replication Task (%s): %s", d.Id(), err)
	}

	if err := waitReplicationTaskDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "waiting for DMS Replication Task (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func replicationTaskRemoveReadOnlySettings(settings string) (*string, error) {
	var settingsData map[string]interface{}
	if err := json.Unmarshal([]byte(settings), &settingsData); err != nil {
		return nil, err
	}

	controlTablesSettings, ok := settingsData["ControlTablesSettings"].(map[string]interface{})
	if ok {
		delete(controlTablesSettings, "historyTimeslotInMinutes")
	}

	logging, ok := settingsData["Logging"].(map[string]interface{})
	if ok {
		delete(logging, "EnableLogContext")
		delete(logging, "CloudWatchLogGroup")
		delete(logging, "CloudWatchLogStream")
	}

	cleanedSettings, err := json.Marshal(settingsData)
	if err != nil {
		return nil, err
	}

	cleanedSettingsString := string(cleanedSettings)
	return &cleanedSettingsString, nil
}

func startReplicationTask(ctx context.Context, conn *dms.DatabaseMigrationService, id string) error {
	log.Printf("[DEBUG] Starting DMS Replication Task: (%s)", id)

	task, err := FindReplicationTaskByID(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("reading DMS Replication Task (%s): %w", id, err)
	}

	if task == nil {
		return fmt.Errorf("reading DMS Replication Task (%s): empty output", id)
	}

	startReplicationTaskType := dms.StartReplicationTaskTypeValueStartReplication
	if aws.StringValue(task.Status) != replicationTaskStatusReady {
		startReplicationTaskType = dms.StartReplicationTaskTypeValueResumeProcessing
	}

	_, err = conn.StartReplicationTaskWithContext(ctx, &dms.StartReplicationTaskInput{
		ReplicationTaskArn:       task.ReplicationTaskArn,
		StartReplicationTaskType: aws.String(startReplicationTaskType),
	})

	if err != nil {
		return fmt.Errorf("starting DMS Replication Task (%s): %w", id, err)
	}

	err = waitReplicationTaskRunning(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("waiting for DMS Replication Task (%s) start: %w", id, err)
	}

	return nil
}

func stopReplicationTask(ctx context.Context, id string, conn *dms.DatabaseMigrationService) error {
	log.Printf("[DEBUG] Stopping DMS Replication Task: (%s)", id)

	task, err := FindReplicationTaskByID(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("reading DMS Replication Task (%s): %w", id, err)
	}

	if task == nil {
		return fmt.Errorf("reading DMS Replication Task (%s): empty output", id)
	}

	_, err = conn.StopReplicationTaskWithContext(ctx, &dms.StopReplicationTaskInput{
		ReplicationTaskArn: task.ReplicationTaskArn,
	})

	if err != nil {
		return fmt.Errorf("stopping DMS Replication Task (%s): %w", id, err)
	}

	err = waitReplicationTaskStopped(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("waiting for DMS Replication Task (%s) stop: %w", id, err)
	}

	return nil
}

func FindReplicationTaskByID(ctx context.Context, conn *dms.DatabaseMigrationService, id string) (*dms.ReplicationTask, error) {
	input := &dms.DescribeReplicationTasksInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("replication-task-id"),
				Values: []*string{aws.String(id)}, // Must use d.Id() to work with import.
			},
		},
	}
	return FindReplicationTask(ctx, conn, input)
}

func FindReplicationTask(ctx context.Context, conn *dms.DatabaseMigrationService, input *dms.DescribeReplicationTasksInput) (*dms.ReplicationTask, error) {
	var results []*dms.ReplicationTask

	err := conn.DescribeReplicationTasksPagesWithContext(ctx, input, func(page *dms.DescribeReplicationTasksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, task := range page.ReplicationTasks {
			if task == nil {
				continue
			}
			results = append(results, task)
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

	if len(results) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(results); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return results[0], nil
}

func FindReplicationTasksByEndpointARN(ctx context.Context, conn *dms.DatabaseMigrationService, arn string) ([]*dms.ReplicationTask, error) {
	input := &dms.DescribeReplicationTasksInput{
		Filters: []*dms.Filter{
			{
				Name:   aws.String("endpoint-arn"),
				Values: []*string{aws.String(arn)},
			},
		},
	}

	var results []*dms.ReplicationTask

	err := conn.DescribeReplicationTasksPagesWithContext(ctx, input, func(page *dms.DescribeReplicationTasksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, task := range page.ReplicationTasks {
			if task == nil {
				continue
			}

			switch aws.StringValue(task.Status) {
			case replicationTaskStatusRunning, replicationTaskStatusStarting:
				results = append(results, task)
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

	return results, nil
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

const (
	replicationTaskRunningTimeout = 5 * time.Minute
)

func waitReplicationTaskDeleted(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusDeleting},
		Target:     []string{},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskModified(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusModifying},
		Target:     []string{replicationTaskStatusReady, replicationTaskStatusStopped, replicationTaskStatusFailed},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskReady(ctx context.Context, conn *dms.DatabaseMigrationService, id string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusCreating},
		Target:     []string{replicationTaskStatusReady},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskRunning(ctx context.Context, conn *dms.DatabaseMigrationService, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{replicationTaskStatusStarting},
		Target:     []string{replicationTaskStatusRunning},
		Refresh:    statusReplicationTask(ctx, conn, id),
		Timeout:    replicationTaskRunningTimeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskStopped(ctx context.Context, conn *dms.DatabaseMigrationService, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{replicationTaskStatusStopping, replicationTaskStatusRunning},
		Target:                    []string{replicationTaskStatusStopped},
		Refresh:                   statusReplicationTask(ctx, conn, id),
		Timeout:                   replicationTaskRunningTimeout,
		MinTimeout:                10 * time.Second,
		Delay:                     60 * time.Second, // Wait 60 secs before starting
		ContinuousTargetOccurence: 2,
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicationTaskSteady(ctx context.Context, conn *dms.DatabaseMigrationService, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{replicationTaskStatusCreating, replicationTaskStatusDeleting, replicationTaskStatusModifying, replicationTaskStatusStopping, replicationTaskStatusStarting},
		Target:                    []string{replicationTaskStatusFailed, replicationTaskStatusReady, replicationTaskStatusStopped, replicationTaskStatusRunning},
		Refresh:                   statusReplicationTask(ctx, conn, id),
		Timeout:                   replicationTaskRunningTimeout,
		MinTimeout:                10 * time.Second,
		Delay:                     60 * time.Second, // Wait 60 secs before starting
		ContinuousTargetOccurence: 2,
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
