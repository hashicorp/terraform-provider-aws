package aws

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/qldb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceAwsQLDBStream() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsQLDBStreamCreate,
		Read:   resourceAwsQLDBStreamRead,
		Update: resourceAwsQLDBStreamUpdate,
		Delete: resourceAwsQLDBStreamDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		// https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-qldb-stream.html
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"exlusive_end_time": {
				Type:     schema.TypeString,
				Required: false,
				ForceNew: true,
				ValidateFunc: validation.All(
					validateUTCTimestamp, // The ExclusiveEndTime must be in ISO 8601 date and time format and in Universal Coordinated Time (UTC). For example: 2019-06-13T21:36:34Z.
				),
			},
			"inclusive_start_time": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validateUTCTimestamp,
				),
			},

			"kinesis_configuration": {
				Type:     schema.TypeList,
				ForceNew: true,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aggregation_enabled": {
							Type:         schema.TypeBool,
							Required:     false,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},

						"role_arn": {
							Type:         schema.TypeString,
							Required:     false,
							ForceNew:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},

			"ledger_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`(?!^.*--)(?!^[0-9]+$)(?!^-)(?!.*-$)^[A-Za-z0-9_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),
				),
			},

			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateArn,
			},

			"stream_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`(?!^.*--)(?!^[0-9]+$)(?!^-)(?!.*-$)^[A-Za-z0-9_-]+`), "must contain only alphanumeric characters, underscores, and hyphens"),				)
			},

			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsQLDBStreamCreate(d *schema.ResourceData, meta interface{}) error {
	// conn := meta.(*AWSClient).qldbconn

	// var name string
	// if v, ok := d.GetOk("name"); ok {
	// 	name = v.(string)
	// } else {
	// 	name = resource.PrefixedUniqueId("tf")
	// }

	// if err := d.Set("name", name); err != nil {
	// 	return fmt.Errorf("error setting name: %s", err)
	// }

	// // Create the QLDB Ledger
	// // The qldb.PermissionsModeAllowAll is currently hardcoded because AWS doesn't support changing the mode.
	// createOpts := &qldb.CreateLedgerInput{
	// 	Name:               aws.String(d.Get("name").(string)),
	// 	PermissionsMode:    aws.String(qldb.PermissionsModeAllowAll),
	// 	DeletionProtection: aws.Bool(d.Get("deletion_protection").(bool)),
	// 	Tags:               keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().QldbTags(),
	// }

	// log.Printf("[DEBUG] QLDB Ledger create config: %#v", *createOpts)
	// qldbResp, err := conn.CreateLedger(createOpts)
	// if err != nil {
	// 	return fmt.Errorf("Error creating QLDB Ledger: %s", err)
	// }

	// // Set QLDB ledger name
	// d.SetId(aws.StringValue(qldbResp.Name))

	// log.Printf("[INFO] QLDB Ledger name: %s", d.Id())

	// stateConf := &resource.StateChangeConf{
	// 	Pending:    []string{qldb.LedgerStateCreating},
	// 	Target:     []string{qldb.LedgerStateActive},
	// 	Refresh:    qldbLedgerRefreshStatusFunc(conn, d.Id()),
	// 	Timeout:    8 * time.Minute,
	// 	MinTimeout: 3 * time.Second,
	// }

	// _, err = stateConf.WaitForState()
	// if err != nil {
	// 	return fmt.Errorf("Error waiting for QLDB Ledger status to be \"%s\": %s", qldb.LedgerStateActive, err)
	// }

	// // Update our attributes and return
	// return resourceAwsQLDBStreamRead(d, meta)
}

func resourceAwsQLDBStreamRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).qldbconn
	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

	// Refresh the QLDB Stream state
	input := &qldb.DescribeJournalKinesisStreamInput{
		LedgerName: aws.String(d.Get("ledger_name")),
		StreamId:   aws.String(d.Get("stream_id")),
	}

	qldbStream, err := conn.DescribeJournalKinesisStream(input)

	if isAWSErr(err, qldb.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] QLDB Stream (%s) not found, removing from state", d.Get("stream_id"))
		d.Set("stream_id", "")
		d.Set("ledger_name", "")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing QLDB Stream (%s): %s", d.Get("stream_id"), err)
	}

	if err := d.Set("arn", qldbStream.Stream.Arn); err != nil {
		return fmt.Errorf("error setting ARN: %s", err)
	}

	if err := d.Set("creation_time", qldbStream.Stream.CreationTime); err != nil {
		return fmt.Errorf("error setting Creation Time: %s", err)
	}

	if err := d.Set("error_cause", qldbStream.Stream.ErrorCause); err != nil {
		return fmt.Errorf("error setting Error Cause: %s", err)
	}

	if err := d.Set("exclusive_end_time", qldbStream.Stream.ExclusiveEndTime); err != nil {
		return fmt.Errorf("error setting Exclusive End Time: %s", err)
	}

	if err := d.Set("inclusive_start_time", qldbStream.Stream.InclusiveStartTime); err != nil {
		return fmt.Errorf("error setting Inclusive Start Time: %s", err)
	}

	if err := d.Set("kinesis_configuration", qldbStream.Stream.KinesisConfiguration); err != nil {
		return fmt.Errorf("error setting Kinesis Configuration: %s", err)
	}

	if err := d.Set("ledger_name", qldbStream.Stream.LedgerName); err != nil {
		return fmt.Errorf("error setting Ledger Name: %s", err)
	}

	if err := d.Set("role_arn", qldbStream.Stream.RoleArn); err != nil {
		return fmt.Errorf("error setting Role Arn: %s", err)
	}

	if err := d.Set("status", qldbStream.Stream.Status); err != nil {
		return fmt.Errorf("error setting Status: %s", err)
	}
	
	if err := d.Set("stream_id", qldbStream.Stream.StreamId); err != nil {
		return fmt.Errorf("error setting Stream Id: %s", err)
	}

	if err := d.Set("stream_name", qldbStream.Stream.StreamName); err != nil {
		return fmt.Errorf("error setting Stream Name: %s", err)
	}

	// TODO: Check this is working.  Seems like it should be the same regardless of resource type...
	// Tags
	log.Printf("[INFO] Fetching tags for %s", d.Id())
	tags, err := keyvaluetags.QldbListTags(conn, d.Get("arn").(string))
	if err != nil {
		return fmt.Errorf("Error listing tags for QLDB Ledger: %s", err)
	}

	if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %s", err)
	}

	return nil
}

func resourceAwsQLDBStreamUpdate(d *schema.ResourceData, meta interface{}) error {
	// conn := meta.(*AWSClient).qldbconn

	// if d.HasChange("deletion_protection") {
	// 	val := d.Get("deletion_protection").(bool)
	// 	modifyOpts := &qldb.UpdateLedgerInput{
	// 		Name:               aws.String(d.Id()),
	// 		DeletionProtection: aws.Bool(val),
	// 	}
	// 	log.Printf(
	// 		"[INFO] Modifying deletion_protection QLDB attribute for %s: %#v",
	// 		d.Id(), modifyOpts)
	// 	if _, err := conn.UpdateLedger(modifyOpts); err != nil {

	// 		return err
	// 	}
	// }

	// if d.HasChange("tags") {
	// 	o, n := d.GetChange("tags")
	// 	if err := keyvaluetags.QldbUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
	// 		return fmt.Errorf("error updating tags: %s", err)
	// 	}
	// }

	// return resourceAwsQLDBLedgerRead(d, meta)
}

// TODO: You cannot actually "delete" a stream, it can only be "cancelled".  Not sure about naming preferences here...
func resourceAwsQLDBStreamDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).qldbconn
	deleteQLDBStreamOpts := &qldb.CancelJournalKinesisStreamInput{
		LedgerName: aws.String(d.Get("ledger_name")),
		StreamId:   aws.String(d.Get("stream_id")),
	}
	log.Printf("[INFO] Cancelling QLDB Ledger: %s %s", d.Get("ledger_name"), d.Get("stream_id"))

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.CancelJournalKinesisStream(deleteQLDBStreamOpts)

		// TODO:  Confirm which errors to be checking for here
		if isAWSErr(err, qldb.ErrCodeResourceInUseException, "") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CancelJournalKinesisStream(deleteQLDBStreamOpts)
	}

	if isAWSErr(err, qldb.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error cancelling QLDB Stream (%s, %s): %s", d.Get("ledger_name"), d.Get("stream_id"), err)
	}

	if err := waitForQLDBStreamDeletion(conn, d.LedgerName(), d.StreamID()); err != nil {
		return fmt.Errorf("error waiting for QLDB Stream (%s, %s) deletion: %s", d.Get("ledger_name"), d.Get("stream_id"), err)
	}

	return nil
}

func qldbStreamRefreshStatusFunc(conn *qldb.QLDB, ledgerName string, streamID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &qldb.DescribeJournalKinesisStreamInput{
			LedgerName: aws.String(ledgerName),
			StreamId:   aws.String(streamID),
		}
		resp, err := conn.DescribeJournalKinesisStream(input)
		if err != nil {
			return nil, "failed", err
		}
		return resp, aws.StringValue(resp.Stream.Status), nil
	}
}

func waitForQLDBStreamDeletion(conn *qldb.QLDB, ledgerName string, streamID string) error {
	stateConf := resource.StateChangeConf{
		Pending: []string{
			qldb.StreamStatusActive,
			qldb.StreamStatusCanceled,
			qldb.StreamStatusCompleted,
			qldb.StreamStatusFailed,
			qldb.StreamStatusImpaired,
		},
		Target:     []string{""},
		Timeout:    5 * time.Minute,
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			resp, err := conn.DescribeJournalKinesisStream(&qldb.DescribeJournalKinesisStreamInput{
				LedgerName: aws.String(ledgerName),
				StreamId:   aws.String(streamID),
			})

			if isAWSErr(err, qldb.ErrCodeResourceNotFoundException, "") {
				return 1, "", nil
			}

			if err != nil {
				return nil, qldb.ErrCodeResourceInUseException, err
			}

			return resp, aws.StringValue(resp.Stream.Status), nil
		},
	}

	_, err := stateConf.WaitForState()

	return err
}
