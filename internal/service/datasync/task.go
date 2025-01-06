// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datasync

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/datasync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/datasync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_datasync_task", name="Task")
// @Tags(identifierAttribute="id")
func resourceTask() *schema.Resource {
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
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCloudWatchLogGroupARN: {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.FilterType](),
						},
						names.AttrValue: {
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
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.FilterType](),
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrName: {
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
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.AtimeBestEffort,
							ValidateDiagFunc: enum.Validate[awstypes.Atime](),
						},
						"bytes_per_second": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      -1,
							ValidateFunc: validation.IntAtLeast(-1),
						},
						"gid": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.GidIntValue,
							ValidateDiagFunc: enum.Validate[awstypes.Gid](),
						},
						"log_level": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.LogLevelOff,
							ValidateDiagFunc: enum.Validate[awstypes.LogLevel](),
						},
						"mtime": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.MtimePreserve,
							ValidateDiagFunc: enum.Validate[awstypes.Mtime](),
						},
						"object_tags": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.ObjectTagsPreserve,
							ValidateDiagFunc: enum.Validate[awstypes.ObjectTags](),
						},
						"overwrite_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.OverwriteModeAlways,
							ValidateDiagFunc: enum.Validate[awstypes.OverwriteMode](),
						},
						"posix_permissions": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.PosixPermissionsPreserve,
							ValidateDiagFunc: enum.Validate[awstypes.PosixPermissions](),
						},
						"preserve_deleted_files": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.PreserveDeletedFilesPreserve,
							ValidateDiagFunc: enum.Validate[awstypes.PreserveDeletedFiles](),
						},
						"preserve_devices": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.PreserveDevicesNone,
							ValidateDiagFunc: enum.Validate[awstypes.PreserveDevices](),
						},
						"security_descriptor_copy_flags": {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.SmbSecurityDescriptorCopyFlags](),
						},
						"task_queueing": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.TaskQueueingEnabled,
							ValidateDiagFunc: enum.Validate[awstypes.TaskQueueing](),
						},
						"transfer_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.TransferModeChanged,
							ValidateDiagFunc: enum.Validate[awstypes.TransferMode](),
						},
						"uid": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.UidIntValue,
							ValidateDiagFunc: enum.Validate[awstypes.Uid](),
						},
						"verify_mode": {
							Type:             schema.TypeString,
							Optional:         true,
							Default:          awstypes.VerifyModePointInTimeConsistent,
							ValidateDiagFunc: enum.Validate[awstypes.VerifyMode](),
						},
					},
				},
			},
			names.AttrSchedule: {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrScheduleExpression: {
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
			"task_report_config": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_destination": {
							Type:     schema.TypeList,
							Required: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bucket_access_role_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"s3_bucket_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
									"subdirectory": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"s3_object_versioning": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ObjectVersionIds](),
						},
						"output_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ReportOutputType](),
						},
						"report_overrides": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"deleted_override": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ReportLevel](),
									},
									"skipped_override": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ReportLevel](),
									},
									"transferred_override": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ReportLevel](),
									},
									"verified_override": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.ReportLevel](),
									},
								},
							},
						},
						"report_level": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ReportLevel](),
						},
					},
				},
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTaskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	input := &datasync.CreateTaskInput{
		DestinationLocationArn: aws.String(d.Get("destination_location_arn").(string)),
		Options:                expandOptions(d.Get("options").([]interface{})),
		SourceLocationArn:      aws.String(d.Get("source_location_arn").(string)),
		Tags:                   getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrCloudWatchLogGroupARN); ok {
		input.CloudWatchLogGroupArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excludes"); ok {
		input.Excludes = expandFilterRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("includes"); ok {
		input.Includes = expandFilterRules(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("task_report_config"); ok {
		input.TaskReportConfig = expandTaskReportConfig(v.([]interface{}))
	}

	if v, ok := d.GetOk(names.AttrSchedule); ok {
		input.Schedule = expandTaskSchedule(v.([]interface{}))
	}

	output, err := conn.CreateTask(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataSync Task: %s", err)
	}

	d.SetId(aws.ToString(output.TaskArn))

	if _, err := waitTaskAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for DataSync Task (%s) creation: %s", d.Id(), err)
	}

	return append(diags, resourceTaskRead(ctx, d, meta)...)
}

func resourceTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	output, err := findTaskByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Task (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataSync Task (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.TaskArn)
	d.Set(names.AttrCloudWatchLogGroupARN, output.CloudWatchLogGroupArn)
	d.Set("destination_location_arn", output.DestinationLocationArn)
	if err := d.Set("excludes", flattenFilterRules(output.Excludes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting excludes: %s", err)
	}
	if err := d.Set("includes", flattenFilterRules(output.Includes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting includes: %s", err)
	}
	d.Set(names.AttrName, output.Name)
	if err := d.Set("options", flattenOptions(output.Options)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting options: %s", err)
	}
	if err := d.Set(names.AttrSchedule, flattenTaskSchedule(output.Schedule)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting schedule: %s", err)
	}
	if err := d.Set("task_report_config", flattenTaskReportConfig(output.TaskReportConfig)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting task_report_config: %s", err)
	}
	d.Set("source_location_arn", output.SourceLocationArn)

	return diags
}

func resourceTaskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &datasync.UpdateTaskInput{
			TaskArn: aws.String(d.Id()),
		}

		if d.HasChanges(names.AttrCloudWatchLogGroupARN) {
			input.CloudWatchLogGroupArn = aws.String(d.Get(names.AttrCloudWatchLogGroupARN).(string))
		}

		if d.HasChanges("excludes") {
			input.Excludes = expandFilterRules(d.Get("excludes").([]interface{}))
		}

		if d.HasChanges("includes") {
			input.Includes = expandFilterRules(d.Get("includes").([]interface{}))
		}

		if d.HasChanges(names.AttrName) {
			input.Name = aws.String(d.Get(names.AttrName).(string))
		}

		if d.HasChanges("options") {
			input.Options = expandOptions(d.Get("options").([]interface{}))
		}

		if d.HasChanges(names.AttrSchedule) {
			input.Schedule = expandTaskSchedule(d.Get(names.AttrSchedule).([]interface{}))
		}

		if d.HasChanges("task_report_config") {
			input.TaskReportConfig = expandTaskReportConfig(d.Get("task_report_config").([]interface{}))
		}

		if _, err := conn.UpdateTask(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating DataSync Task (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceTaskRead(ctx, d, meta)...)
}

func resourceTaskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataSyncClient(ctx)

	log.Printf("[DEBUG] Deleting DataSync Task: %s", d.Id())
	_, err := conn.DeleteTask(ctx, &datasync.DeleteTaskInput{
		TaskArn: aws.String(d.Id()),
	})

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting DataSync Task (%s): %s", d.Id(), err)
	}

	return diags
}

func findTaskByARN(ctx context.Context, conn *datasync.Client, arn string) (*datasync.DescribeTaskOutput, error) {
	input := &datasync.DescribeTaskInput{
		TaskArn: aws.String(arn),
	}

	output, err := conn.DescribeTask(ctx, input)

	if errs.IsAErrorMessageContains[*awstypes.InvalidRequestException](err, "not found") {
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

func statusTask(ctx context.Context, conn *datasync.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTaskByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitTaskAvailable(ctx context.Context, conn *datasync.Client, arn string, timeout time.Duration) (*datasync.DescribeTaskOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TaskStatusCreating, awstypes.TaskStatusUnavailable),
		Target:  enum.Slice(awstypes.TaskStatusAvailable, awstypes.TaskStatusRunning),
		Refresh: statusTask(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*datasync.DescribeTaskOutput); ok {
		if errorCode, errorDetail := aws.ToString(output.ErrorCode), aws.ToString(output.ErrorDetail); errorCode != "" && errorDetail != "" {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", errorCode, errorDetail))
		}

		return output, err
	}

	return nil, err
}

func flattenOptions(options *awstypes.Options) []interface{} {
	if options == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"atime":                          string(options.Atime),
		"bytes_per_second":               aws.ToInt64(options.BytesPerSecond),
		"gid":                            string(options.Gid),
		"log_level":                      string(options.LogLevel),
		"mtime":                          string(options.Mtime),
		"object_tags":                    string(options.ObjectTags),
		"overwrite_mode":                 string(options.OverwriteMode),
		"posix_permissions":              string(options.PosixPermissions),
		"preserve_deleted_files":         string(options.PreserveDeletedFiles),
		"preserve_devices":               string(options.PreserveDevices),
		"security_descriptor_copy_flags": string(options.SecurityDescriptorCopyFlags),
		"task_queueing":                  string(options.TaskQueueing),
		"transfer_mode":                  string(options.TransferMode),
		"uid":                            string(options.Uid),
		"verify_mode":                    string(options.VerifyMode),
	}

	return []interface{}{m}
}

func flattenTaskReportConfig(options *awstypes.TaskReportConfig) []interface{} {
	if options == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"s3_object_versioning": string(options.ObjectVersionIds),
		"output_type":          string(options.OutputType),
		"report_level":         string(options.ReportLevel),
		"s3_destination":       flattenTaskReportConfigS3Destination(options.Destination.S3),
		"report_overrides":     flattenTaskReportConfigReportOverrides(options.Overrides),
	}

	return []interface{}{m}
}

func flattenTaskReportConfigReportOverrides(options *awstypes.ReportOverrides) []interface{} {
	m := make(map[string]interface{})

	if options == nil {
		return []interface{}{m}
	}

	if options.Deleted != nil && options.Deleted.ReportLevel != "" {
		m["deleted_override"] = string(options.Deleted.ReportLevel)
	}

	if options.Skipped != nil && options.Skipped.ReportLevel != "" {
		m["skipped_override"] = string(options.Skipped.ReportLevel)
	}

	if options.Transferred != nil && options.Transferred.ReportLevel != "" {
		m["transferred_override"] = string(options.Transferred.ReportLevel)
	}

	if options.Verified != nil && options.Verified.ReportLevel != "" {
		m["verified_override"] = string(options.Verified.ReportLevel)
	}

	return []interface{}{m}
}

func flattenTaskReportConfigS3Destination(options *awstypes.ReportDestinationS3) []interface{} {
	if options == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"bucket_access_role_arn": aws.ToString(options.BucketAccessRoleArn),
		"s3_bucket_arn":          aws.ToString(options.S3BucketArn),
		"subdirectory":           aws.ToString(options.Subdirectory),
	}

	return []interface{}{m}
}

func expandOptions(l []interface{}) *awstypes.Options {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	options := &awstypes.Options{
		Atime:                awstypes.Atime(m["atime"].(string)),
		Gid:                  awstypes.Gid(m["gid"].(string)),
		LogLevel:             awstypes.LogLevel(m["log_level"].(string)),
		Mtime:                awstypes.Mtime(m["mtime"].(string)),
		ObjectTags:           awstypes.ObjectTags(m["object_tags"].(string)),
		OverwriteMode:        awstypes.OverwriteMode(m["overwrite_mode"].(string)),
		PreserveDeletedFiles: awstypes.PreserveDeletedFiles(m["preserve_deleted_files"].(string)),
		PreserveDevices:      awstypes.PreserveDevices(m["preserve_devices"].(string)),
		PosixPermissions:     awstypes.PosixPermissions(m["posix_permissions"].(string)),
		TaskQueueing:         awstypes.TaskQueueing(m["task_queueing"].(string)),
		TransferMode:         awstypes.TransferMode(m["transfer_mode"].(string)),
		Uid:                  awstypes.Uid(m["uid"].(string)),
		VerifyMode:           awstypes.VerifyMode(m["verify_mode"].(string)),
	}

	if v, ok := m["bytes_per_second"].(int); ok && v != 0 {
		options.BytesPerSecond = aws.Int64(int64(v))
	}

	if v, ok := m["security_descriptor_copy_flags"].(string); ok && v != "" {
		options.SecurityDescriptorCopyFlags = awstypes.SmbSecurityDescriptorCopyFlags(v)
	}

	return options
}

func expandTaskSchedule(l []interface{}) *awstypes.TaskSchedule {
	if len(l) == 0 || l[0] == nil {
		return &awstypes.TaskSchedule{ScheduleExpression: aws.String("")} // explicitly set empty object if schedule is nil
	}

	m := l[0].(map[string]interface{})

	schedule := &awstypes.TaskSchedule{
		ScheduleExpression: aws.String(m[names.AttrScheduleExpression].(string)),
	}

	return schedule
}

func flattenTaskSchedule(schedule *awstypes.TaskSchedule) []interface{} {
	if schedule == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrScheduleExpression: aws.ToString(schedule.ScheduleExpression),
	}

	return []interface{}{m}
}

func expandTaskReportConfig(l []interface{}) *awstypes.TaskReportConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	reportConfig := &awstypes.TaskReportConfig{}

	m := l[0].(map[string]interface{})

	dest := m["s3_destination"].([]interface{})
	reportConfig.Destination = expandTaskReportDestination(dest)
	reportConfig.ObjectVersionIds = awstypes.ObjectVersionIds(m["s3_object_versioning"].(string))
	reportConfig.OutputType = awstypes.ReportOutputType(m["output_type"].(string))
	reportConfig.ReportLevel = awstypes.ReportLevel(m["report_level"].(string))
	o := m["report_overrides"].([]interface{})
	reportConfig.Overrides = expandTaskReportOverrides(o)

	return reportConfig
}

func expandTaskReportDestination(l []interface{}) *awstypes.ReportDestination {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	m := l[0].(map[string]interface{})
	return &awstypes.ReportDestination{
		S3: &awstypes.ReportDestinationS3{
			BucketAccessRoleArn: aws.String(m["bucket_access_role_arn"].(string)),
			S3BucketArn:         aws.String(m["s3_bucket_arn"].(string)),
			Subdirectory:        aws.String(m["subdirectory"].(string)),
		},
	}
}

func expandTaskReportOverrides(l []interface{}) *awstypes.ReportOverrides {
	var overrides = &awstypes.ReportOverrides{}

	if len(l) == 0 || l[0] == nil {
		return overrides
	}

	m := l[0].(map[string]interface{})

	deleteOverride := m["deleted_override"].(string)
	if deleteOverride != "" {
		overrides.Deleted = &awstypes.ReportOverride{
			ReportLevel: awstypes.ReportLevel(deleteOverride),
		}
	}

	skippedOverride := m["skipped_override"].(string)
	if skippedOverride != "" {
		overrides.Skipped = &awstypes.ReportOverride{
			ReportLevel: awstypes.ReportLevel(skippedOverride),
		}
	}

	transferredOverride := m["transferred_override"].(string)
	if transferredOverride != "" {
		overrides.Transferred = &awstypes.ReportOverride{
			ReportLevel: awstypes.ReportLevel(transferredOverride),
		}
	}

	verifiedOverride := m["verified_override"].(string)
	if verifiedOverride != "" {
		overrides.Verified = &awstypes.ReportOverride{
			ReportLevel: awstypes.ReportLevel(verifiedOverride),
		}
	}

	return overrides
}

func expandFilterRules(l []interface{}) []awstypes.FilterRule {
	filterRules := []awstypes.FilterRule{}

	for _, mRaw := range l {
		if mRaw == nil {
			continue
		}
		m := mRaw.(map[string]interface{})
		filterRule := awstypes.FilterRule{
			FilterType: awstypes.FilterType(m["filter_type"].(string)),
			Value:      aws.String(m[names.AttrValue].(string)),
		}
		filterRules = append(filterRules, filterRule)
	}

	return filterRules
}

func flattenFilterRules(filterRules []awstypes.FilterRule) []interface{} {
	l := []interface{}{}

	for _, filterRule := range filterRules {
		m := map[string]interface{}{
			"filter_type":   string(filterRule.FilterType),
			names.AttrValue: aws.ToString(filterRule.Value),
		}
		l = append(l, m)
	}

	return l
}
