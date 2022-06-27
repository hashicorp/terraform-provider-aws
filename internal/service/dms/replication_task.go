package dms

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceReplicationTask() *schema.Resource {
	return &schema.Resource{
		Create: resourceReplicationTaskCreate,
		Read:   resourceReplicationTaskRead,
		Update: resourceReplicationTaskUpdate,
		Delete: resourceReplicationTaskDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
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

func resourceReplicationTaskCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	taskId := d.Get("replication_task_id").(string)

	request := &dms.CreateReplicationTaskInput{
		MigrationType:             aws.String(d.Get("migration_type").(string)),
		ReplicationInstanceArn:    aws.String(d.Get("replication_instance_arn").(string)),
		ReplicationTaskIdentifier: aws.String(taskId),
		SourceEndpointArn:         aws.String(d.Get("source_endpoint_arn").(string)),
		TableMappings:             aws.String(d.Get("table_mappings").(string)),
		Tags:                      Tags(tags.IgnoreAWS()),
		TargetEndpointArn:         aws.String(d.Get("target_endpoint_arn").(string)),
	}

	if v, ok := d.GetOk("cdc_start_position"); ok {
		request.CdcStartPosition = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cdc_start_time"); ok {
		seconds, err := strconv.ParseInt(v.(string), 10, 64)
		if err != nil {
			return fmt.Errorf("DMS create replication task. Invalid CDC Unix timestamp: %s", err)
		}
		request.CdcStartTime = aws.Time(time.Unix(seconds, 0))
	}

	if v, ok := d.GetOk("replication_task_settings"); ok {
		request.ReplicationTaskSettings = aws.String(v.(string))
	}

	log.Println("[DEBUG] DMS create replication task:", request)

	_, err := conn.CreateReplicationTask(request)
	if err != nil {
		return fmt.Errorf("error creating DMS Replication Task (%s): %w", taskId, err)
	}

	d.SetId(taskId)

	if err := waitReplicationTaskReady(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for DMS Replication Task (%s) to become available: %w", d.Id(), err)
	}

	if d.Get("start_replication_task").(bool) {
		if err := startReplicationTask(d.Id(), conn); err != nil {
			return err
		}
	}

	return resourceReplicationTaskRead(d, meta)
}

func resourceReplicationTaskRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	task, err := FindReplicationTaskByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] DMS Replication Task (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading DMS Replication Task (%s): %w", d.Id(), err)
	}

	if task == nil {
		return fmt.Errorf("error reading DMS Replication Task (%s): empty output", d.Id())
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
		return err
	}

	d.Set("replication_task_settings", settings)

	tags, err := ListTags(conn, d.Get("replication_task_arn").(string))

	if err != nil {
		return fmt.Errorf("error listing tags for DMS Replication Task (%s): %s", d.Get("replication_task_arn").(string), err)
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

func resourceReplicationTaskUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn

	if d.HasChangesExcept("tags", "tags_all", "start_replication_task") {
		input := &dms.ModifyReplicationTaskInput{
			ReplicationTaskArn: aws.String(d.Get("replication_task_arn").(string)),
		}

		if d.HasChange("cdc_start_position") {
			input.CdcStartPosition = aws.String(d.Get("cdc_start_position").(string))
		}

		if d.HasChange("cdc_start_time") {
			seconds, err := strconv.ParseInt(d.Get("cdc_start_time").(string), 10, 64)
			if err != nil {
				return fmt.Errorf("DMS update replication task. Invalid CRC Unix timestamp: %s", err)
			}
			input.CdcStartTime = aws.Time(time.Unix(seconds, 0))
		}

		if d.HasChange("migration_type") {
			input.MigrationType = aws.String(d.Get("migration_type").(string))
		}

		if d.HasChange("replication_task_settings") {
			input.ReplicationTaskSettings = aws.String(d.Get("replication_task_settings").(string))
		}

		if d.HasChange("table_mappings") {
			input.TableMappings = aws.String(d.Get("table_mappings").(string))
		}

		status := d.Get("status").(string)
		if status == replicationTaskStatusRunning {
			log.Println("[DEBUG] stopping DMS replication task:", input)
			if err := stopReplicationTask(d.Id(), conn); err != nil {
				return err
			}
		}

		log.Println("[DEBUG] updating DMS replication task:", input)
		_, err := conn.ModifyReplicationTask(input)
		if err != nil {
			return fmt.Errorf("error updating DMS Replication Task (%s): %w", d.Id(), err)
		}

		if err := waitReplicationTaskModified(conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return fmt.Errorf("error waiting for DMS Replication Task (%s) update: %s", d.Id(), err)
		}

		if d.Get("start_replication_task").(bool) {
			err := startReplicationTask(d.Id(), conn)
			if err != nil {
				return err
			}
		}
	}

	if d.HasChanges("start_replication_task") {
		status := d.Get("status").(string)
		if d.Get("start_replication_task").(bool) {
			if status != replicationTaskStatusRunning {
				if err := startReplicationTask(d.Id(), conn); err != nil {
					return err
				}
			}
		} else {
			if status == replicationTaskStatusRunning {
				if err := stopReplicationTask(d.Id(), conn); err != nil {
					return err
				}
			}
		}
	}

	if d.HasChange("tags_all") {
		arn := d.Get("replication_task_arn").(string)
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, arn, o, n); err != nil {
			return fmt.Errorf("error updating DMS Replication Task (%s) tags: %s", arn, err)
		}
	}

	return resourceReplicationTaskRead(d, meta)
}

func resourceReplicationTaskDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DMSConn

	if status := d.Get("status").(string); status == replicationTaskStatusRunning {
		if err := stopReplicationTask(d.Id(), conn); err != nil {
			return err
		}
	}

	input := &dms.DeleteReplicationTaskInput{
		ReplicationTaskArn: aws.String(d.Get("replication_task_arn").(string)),
	}

	log.Printf("[DEBUG] DMS delete replication task: %#v", input)

	_, err := conn.DeleteReplicationTask(input)

	if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting DMS Replication Task (%s): %w", d.Id(), err)
	}

	if err := waitReplicationTaskDeleted(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		if tfawserr.ErrCodeEquals(err, dms.ErrCodeResourceNotFoundFault) {
			return nil
		}
		return fmt.Errorf("error waiting for DMS Replication Task (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
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

func startReplicationTask(id string, conn *dms.DatabaseMigrationService) error {
	log.Printf("[DEBUG] Starting DMS Replication Task: (%s)", id)

	task, err := FindReplicationTaskByID(conn, id)
	if err != nil {
		return fmt.Errorf("error reading DMS Replication Task (%s): %w", id, err)
	}

	if task == nil {
		return fmt.Errorf("error reading DMS Replication Task (%s): empty output", id)
	}

	startReplicationTaskType := dms.StartReplicationTaskTypeValueStartReplication
	if aws.StringValue(task.Status) != replicationTaskStatusReady {
		startReplicationTaskType = dms.StartReplicationTaskTypeValueResumeProcessing
	}

	_, err = conn.StartReplicationTask(&dms.StartReplicationTaskInput{
		ReplicationTaskArn:       task.ReplicationTaskArn,
		StartReplicationTaskType: aws.String(startReplicationTaskType),
	})

	if err != nil {
		return fmt.Errorf("error starting DMS Replication Task (%s): %w", id, err)
	}

	err = waitReplicationTaskRunning(conn, id)
	if err != nil {
		return fmt.Errorf("error waiting for DMS Replication Task (%s) start: %w", id, err)
	}

	return nil
}

func stopReplicationTask(id string, conn *dms.DatabaseMigrationService) error {
	log.Printf("[DEBUG] Stopping DMS Replication Task: (%s)", id)

	task, err := FindReplicationTaskByID(conn, id)
	if err != nil {
		return fmt.Errorf("error reading DMS Replication Task (%s): %w", id, err)
	}

	if task == nil {
		return fmt.Errorf("error reading DMS Replication Task (%s): empty output", id)
	}

	_, err = conn.StopReplicationTask(&dms.StopReplicationTaskInput{
		ReplicationTaskArn: task.ReplicationTaskArn,
	})

	if err != nil {
		return fmt.Errorf("error stopping DMS Replication Task (%s): %w", id, err)
	}

	err = waitReplicationTaskStopped(conn, id)
	if err != nil {
		return fmt.Errorf("error waiting for DMS Replication Task (%s) stop: %w", id, err)
	}

	return nil
}
