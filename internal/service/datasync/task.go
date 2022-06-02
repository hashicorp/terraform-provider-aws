package datasync

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/datasync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceTaskCreate,
		Read:   resourceTaskRead,
		Update: resourceTaskUpdate,
		Delete: resourceTaskDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
								validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\ \_\*\?\,\|\^\-\/\#\s\(\)\+]*$`),
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceTaskCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &datasync.CreateTaskInput{
		DestinationLocationArn: aws.String(d.Get("destination_location_arn").(string)),
		Options:                expandOptions(d.Get("options").([]interface{})),
		SourceLocationArn:      aws.String(d.Get("source_location_arn").(string)),
		Tags:                   Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("cloudwatch_log_group_arn"); ok {
		input.CloudWatchLogGroupArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("excludes"); ok {
		input.Excludes = expandFilterRules(v.([]interface{}))
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = expandTaskSchedule(v.([]interface{}))
	}

	log.Printf("[DEBUG] Creating DataSync Task: %s", input)
	output, err := conn.CreateTask(input)

	if err != nil {
		return fmt.Errorf("error creating DataSync Task: %w", err)
	}

	d.SetId(aws.StringValue(output.TaskArn))

	if _, err := waitTaskAvailable(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for DataSync Task (%s) creation: %w", d.Id(), err)
	}

	return resourceTaskRead(d, meta)
}

func resourceTaskRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindTaskByARN(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DataSync Task (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DataSync Task (%s): %w", d.Id(), err)
	}

	d.Set("arn", output.TaskArn)
	d.Set("cloudwatch_log_group_arn", output.CloudWatchLogGroupArn)
	d.Set("destination_location_arn", output.DestinationLocationArn)
	if err := d.Set("excludes", flattenFilterRules(output.Excludes)); err != nil {
		return fmt.Errorf("error setting excludes: %w", err)
	}
	d.Set("name", output.Name)
	if err := d.Set("options", flattenOptions(output.Options)); err != nil {
		return fmt.Errorf("error setting options: %w", err)
	}
	if err := d.Set("schedule", flattenTaskSchedule(output.Schedule)); err != nil {
		return fmt.Errorf("error setting schedule: %w", err)
	}
	d.Set("source_location_arn", output.SourceLocationArn)

	tags, err := ListTags(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error listing tags for DataSync Task (%s): %w", d.Id(), err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

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

		if d.HasChanges("name") {
			input.Name = aws.String(d.Get("name").(string))
		}

		if d.HasChanges("options") {
			input.Options = expandOptions(d.Get("options").([]interface{}))
		}

		if d.HasChanges("schedule") {
			input.Schedule = expandTaskSchedule(d.Get("schedule").([]interface{}))
		}

		log.Printf("[DEBUG] Updating DataSync Task: %s", input)
		if _, err := conn.UpdateTask(input); err != nil {
			return fmt.Errorf("error updating DataSync Task (%s): %w", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating DataSync Task (%s) tags: %w", d.Id(), err)
		}
	}

	return resourceTaskRead(d, meta)
}

func resourceTaskDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DataSyncConn

	input := &datasync.DeleteTaskInput{
		TaskArn: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting DataSync Task: %s", input)
	_, err := conn.DeleteTask(input)

	if tfawserr.ErrMessageContains(err, datasync.ErrCodeInvalidRequestException, "not found") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DataSync Task (%s): %w", d.Id(), err)
	}

	return nil
}

func flattenOptions(options *datasync.Options) []interface{} {
	if options == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"atime":                  aws.StringValue(options.Atime),
		"bytes_per_second":       aws.Int64Value(options.BytesPerSecond),
		"gid":                    aws.StringValue(options.Gid),
		"log_level":              aws.StringValue(options.LogLevel),
		"mtime":                  aws.StringValue(options.Mtime),
		"overwrite_mode":         aws.StringValue(options.OverwriteMode),
		"posix_permissions":      aws.StringValue(options.PosixPermissions),
		"preserve_deleted_files": aws.StringValue(options.PreserveDeletedFiles),
		"preserve_devices":       aws.StringValue(options.PreserveDevices),
		"task_queueing":          aws.StringValue(options.TaskQueueing),
		"transfer_mode":          aws.StringValue(options.TransferMode),
		"uid":                    aws.StringValue(options.Uid),
		"verify_mode":            aws.StringValue(options.VerifyMode),
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
		OverwriteMode:        aws.String(m["overwrite_mode"].(string)),
		PreserveDeletedFiles: aws.String(m["preserve_deleted_files"].(string)),
		PreserveDevices:      aws.String(m["preserve_devices"].(string)),
		PosixPermissions:     aws.String(m["posix_permissions"].(string)),
		TaskQueueing:         aws.String(m["task_queueing"].(string)),
		TransferMode:         aws.String(m["transfer_mode"].(string)),
		Uid:                  aws.String(m["uid"].(string)),
		VerifyMode:           aws.String(m["verify_mode"].(string)),
	}

	if v, ok := m["bytes_per_second"]; ok && v.(int) > 0 {
		options.BytesPerSecond = aws.Int64(int64(v.(int)))
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
