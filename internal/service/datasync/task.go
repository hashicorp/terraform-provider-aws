// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
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

// @SDKResource("aws_datasync_task", name="Task")
// @Tags(identifierAttribute="id")
func ResourceTask() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTaskCreate,
		ReadWithoutTimeout:   resourceTaskRead,
		UpdateWithoutTimeout: resourceTaskUpdate,
		DeleteWithoutTimeout: resourceTaskDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudwatch_log_group_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"destination_location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"excludes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(datasync.FilterType_Values(), false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"includes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"filter_type": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice(datasync.FilterType_Values(), false),
						},
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"options": {
				Type:             schema.TypeList,
				Optional:         true,
				MaxItems:         1,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"atime": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.AtimeBestEffort,
							ValidateFunc: validation.StringInSlice(datasync.Atime_Values(), false),
						},
						"bytes_per_second": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      -1,
							ValidateFunc: validation.IntAtLeast(-1),
						},
						"gid": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.GidIntValue,
							ValidateFunc: validation.StringInSlice(datasync.Gid_Values(), false),
						},
						"log_level": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.LogLevelOff,
							ValidateFunc: validation.StringInSlice(datasync.LogLevel_Values(), false),
						},
						"mtime": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.MtimePreserve,
							ValidateFunc: validation.StringInSlice(datasync.Mtime_Values(), false),
						},
						"object_tags": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.ObjectTagsPreserve,
							ValidateFunc: validation.StringInSlice(datasync.ObjectTags_Values(), false),
						},
						"overwrite_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.OverwriteModeAlways,
							ValidateFunc: validation.StringInSlice(datasync.OverwriteMode_Values(), false),
						},
						"posix_permissions": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.PosixPermissionsPreserve,
							ValidateFunc: validation.StringInSlice(datasync.PosixPermissions_Values(), false),
						},
						"preserve_deleted_files": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.PreserveDeletedFilesPreserve,
							ValidateFunc: validation.StringInSlice(datasync.PreserveDeletedFiles_Values(), false),
						},
						"preserve_devices": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.PreserveDevicesNone,
							ValidateFunc: validation.StringInSlice(datasync.PreserveDevices_Values(), false),
						},
						"security_descriptor_copy_flags": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(datasync.SmbSecurityDescriptorCopyFlags_Values(), false),
						},
						"task_queueing": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.TaskQueueingEnabled,
							ValidateFunc: validation.StringInSlice(datasync.TaskQueueing_Values(), false),
						},
						"transfer_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.TransferModeChanged,
							ValidateFunc: validation.StringInSlice(datasync.TransferMode_Values(), false),
						},
						"uid": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.UidIntValue,
							ValidateFunc: validation.StringInSlice(datasync.Uid_Values(), false),
						},
						"verify_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      datasync.VerifyModePointInTimeConsistent,
							ValidateFunc: validation.StringInSlice(datasync.VerifyMode_Values(), false),
						},
					},
				},
			},
			"schedule": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schedule_expression": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.All(
								validation.StringLenBetween(1, 256),
								validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_\s #()*+,/?^|-]*$`),
									"Schedule expressions must have the following syntax: rate(<number>\\\\s?(minutes?|hours?|days?)), cron(<cron_expression>) or at(yyyy-MM-dd'T'HH:mm:ss)."),
							),
						},
					},
				},
			},
			"source_location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTaskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	input := &datasync.CreateTaskInput{
		DestinationLocationArn: aws.String(d.Get("destination_location_arn").(string)),
		Options:                expandOptions(d.Get("options").([]interface{})),
		SourceLocationArn:      aws.String(d.Get("source_location_arn").(string)),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk("cloudwatch_log_group_arn"); ok {
		input.CloudWatchLogGroupArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excludes"); ok {
		input.Excludes = expandFilterRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("includes"); ok {
		input.Includes = expandFilterRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = expandTaskSchedule(v.([]interface{}))
	}

	output, err := conn.CreateTaskWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Task: %s", err)
	}

	d.SetId(aws.StringValue(output.TaskArn))

	if _, err := waitTaskAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DataSync Task (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceTaskRead(ctx, d, meta)...)
}

func resourceTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	output, err := FindTaskByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Task (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Task (%s): %s", d.Id(), err)
	}

	d.Set("arn", output.TaskArn)
	d.Set("cloudwatch_log_group_arn", output.CloudWatchLogGroupArn)
	d.Set("destination_location_arn", output.DestinationLocationArn)
	if err := d.Set("excludes", flattenFilterRules(output.Excludes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting excludes: %s", err)
	}
	if err := d.Set("includes", flattenFilterRules(output.Includes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting includes: %s", err)
	}
	d.Set("name", output.Name)
	if err := d.Set("options", flattenOptions(output.Options)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting options: %s", err)
	}
	if err := d.Set("schedule", flattenTaskSchedule(output.Schedule)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting schedule: %s", err)
	}
	d.Set("source_location_arn", output.SourceLocationArn)

	return diags
}

func resourceTaskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &datasync.UpdateTaskInput{
			TaskArn: aws.String(d.Id()),
		}

		if d.HasChanges("cloudwatch_log_group_arn") {
			input.CloudWatchLogGroupArn = aws.String(d.Get("cloudwatch_log_group_arn").(string))
		}

		if d.HasChanges("excludes") {
			input.Excludes = expandFilterRules(d.Get("excludes").([]interface{}))
		}

		if d.HasChanges("includes") {
			input.Includes = expandFilterRules(d.Get("includes").([]interface{}))
		}

		if d.HasChanges("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChanges("options") {
			input.Options = expandOptions(d.Get("options").([]interface{}))
		}

		if d.HasChanges("schedule") {
			input.Schedule = expandTaskSchedule(d.Get("schedule").([]interface{}))
		}

		_, err := conn.UpdateTaskWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Task (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTaskRead(ctx, d, meta)...)
}

func resourceTaskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncConn(ctx)

	log.Printf("[DEBUG] Deleting DataSync Task: %s", d.Id())
	_, err := conn.DeleteTaskWithContext(ctx, &datasync.DeleteTaskInput{
		TaskArn: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Task (%s): %s", d.Id(), err)
	}

	return diags
}

func FindTaskByARN(ctx context.Context, conn *datasync.DataSync, arn string) (*datasync.DescribeTaskOutput, error) {
	input := &datasync.DescribeTaskInput{
		TaskArn: aws.String(arn),
	}

	output, err := conn.DescribeTaskWithContext(ctx, input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func statusTask(ctx context.Context, conn *datasync.DataSync, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindTaskByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func waitTaskAvailable(ctx context.Context, conn *datasync.DataSync, arn string, timeout time.Duration) (*datasync.DescribeTaskOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{datasync.TaskStatusCreating, datasync.TaskStatusUnavailable},
		Target:  []string{datasync.TaskStatusAvailable, datasync.TaskStatusRunning},
		Refresh: statusTask(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*datasync.DescribeTaskOutput); ok {
		if errorCode, errorDetail := aws.StringValue(output.ErrorCode), aws.StringValue(output.ErrorDetail); errorCode != "" && errorDetail != "" {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", errorCode, errorDetail))
		}

		return output, err
	}

	return nil, err
}

func flattenOptions(options *datasync.Options) []interface{} {
	if options == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"atime":                          aws.StringValue(options.Atime),
		"bytes_per_second":               aws.Int64Value(options.BytesPerSecond),
		"gid":                            aws.StringValue(options.Gid),
		"log_level":                      aws.StringValue(options.LogLevel),
		"mtime":                          aws.StringValue(options.Mtime),
		"object_tags":                    aws.StringValue(options.ObjectTags),
		"overwrite_mode":                 aws.StringValue(options.OverwriteMode),
		"posix_permissions":              aws.StringValue(options.PosixPermissions),
		"preserve_deleted_files":         aws.StringValue(options.PreserveDeletedFiles),
		"preserve_devices":               aws.StringValue(options.PreserveDevices),
		"security_descriptor_copy_flags": aws.StringValue(options.SecurityDescriptorCopyFlags),
		"task_queueing":                  aws.StringValue(options.TaskQueueing),
		"transfer_mode":                  aws.StringValue(options.TransferMode),
		"uid":                            aws.StringValue(options.Uid),
		"verify_mode":                    aws.StringValue(options.VerifyMode),
	}

	return []interface{}{m}
}

func expandOptions(l []interface{}) *datasync.Options {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	options := &datasync.Options{
		Atime:                aws.String(m["atime"].(string)),
		Gid:                  aws.String(m["gid"].(string)),
		LogLevel:             aws.String(m["log_level"].(string)),
		Mtime:                aws.String(m["mtime"].(string)),
		ObjectTags:           aws.String(m["object_tags"].(string)),
		OverwriteMode:        aws.String(m["overwrite_mode"].(string)),
		PreserveDeletedFiles: aws.String(m["preserve_deleted_files"].(string)),
		PreserveDevices:      aws.String(m["preserve_devices"].(string)),
		PosixPermissions:     aws.String(m["posix_permissions"].(string)),
		TaskQueueing:         aws.String(m["task_queueing"].(string)),
		TransferMode:         aws.String(m["transfer_mode"].(string)),
		Uid:                  aws.String(m["uid"].(string)),
		VerifyMode:           aws.String(m["verify_mode"].(string)),
	}

	if v, ok := m["bytes_per_second"].(int); ok && v != 0 {
		options.BytesPerSecond = aws.Int64(int64(v))
	}

	if v, ok := m["security_descriptor_copy_flags"].(string); ok && v != "" {
		options.SecurityDescriptorCopyFlags = aws.String(v)
	}

	return options
}

func expandTaskSchedule(l []interface{}) *datasync.TaskSchedule {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	schedule := &datasync.TaskSchedule{
		ScheduleExpression: aws.String(m["schedule_expression"].(string)),
	}

	return schedule
}

func flattenTaskSchedule(schedule *datasync.TaskSchedule) []interface{} {
	if schedule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"schedule_expression": aws.StringValue(schedule.ScheduleExpression),
	}

	return []interface{}{m}
}

func expandFilterRules(l []interface{}) []*datasync.FilterRule {
	filterRules := []*datasync.FilterRule{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}
		m := mRaw.(map[string]interface{})
		filterRule := &datasync.FilterRule{
			FilterType: aws.String(m["filter_type"].(string)),
			Value:      aws.String(m["value"].(string)),
		}
		filterRules = append(filterRules, filterRule)
	}

	return filterRules
}

func flattenFilterRules(filterRules []*datasync.FilterRule) []interface{} {
	l := []interface{}{}

	for _, filterRule := range filterRules {
		m := map[string]interface{}{
			"filter_type": aws.StringValue(filterRule.FilterType),
			"value":       aws.StringValue(filterRule.Value),
		}
		l = append(l, m)
	}

	return l
}
