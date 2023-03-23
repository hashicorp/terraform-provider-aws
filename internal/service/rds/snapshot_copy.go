package rds

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceSnapshotCopy() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceSnapshotCopyCreate,
		ReadWithoutTimeout:   resourceSnapshotCopyRead,
		UpdateWithoutTimeout: resourceSnapshotCopyUpdate,
		DeleteWithoutTimeout: resourceSnapshotCopyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"allocated_storage": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"copy_tags": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"db_snapshot_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"destination_region": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"encrypted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"engine": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"engine_version": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"iops": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"kms_key_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"license_model": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"option_group_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"presigned_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"source_db_snapshot_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"source_region": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"snapshot_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"storage_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"target_custom_availability_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"target_db_snapshot_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 255),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9][\w-]+`), "must contain only alphanumeric, and hyphen (-) characters"),
				),
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSnapshotCopyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	in := &rds.CopyDBSnapshotInput{
		SourceDBSnapshotIdentifier: aws.String(d.Get("source_db_snapshot_identifier").(string)),
		TargetDBSnapshotIdentifier: aws.String(d.Get("target_db_snapshot_identifier").(string)),
		Tags:                       Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("copy_tags"); ok {
		in.CopyTags = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("kms_key_id"); ok {
		in.KmsKeyId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("option_group_name"); ok {
		in.OptionGroupName = aws.String(v.(string))
	}
	if v, ok := d.GetOk("destination_region"); ok {
		in.DestinationRegion = aws.String(v.(string))
	}
	if v, ok := d.GetOk("presigned_url"); ok {
		in.PreSignedUrl = aws.String(v.(string))
	}

	out, err := conn.CopyDBSnapshotWithContext(ctx, in)
	if err != nil {
		return diag.Errorf("error creating RDS DB Snapshot Copy %s", err)
	}

	d.SetId(aws.StringValue(out.DBSnapshot.DBSnapshotIdentifier))

	err = waitSnapshotCopyAvailable(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceSnapshotCopyRead(ctx, d, meta)
}

func resourceSnapshotCopyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	snapshot, err := FindSnapshot(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error describing RDS DB snapshot (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(snapshot.DBSnapshotArn)

	d.Set("allocated_storage", snapshot.AllocatedStorage)
	d.Set("availability_zone", snapshot.AvailabilityZone)
	d.Set("db_snapshot_arn", snapshot.DBSnapshotArn)
	d.Set("encrypted", snapshot.Encrypted)
	d.Set("engine", snapshot.Engine)
	d.Set("engine_version", snapshot.EngineVersion)
	d.Set("iops", snapshot.Iops)
	d.Set("kms_key_id", snapshot.KmsKeyId)
	d.Set("license_model", snapshot.LicenseModel)
	d.Set("option_group_name", snapshot.OptionGroupName)
	d.Set("port", snapshot.Port)
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set("source_db_snapshot_identifier", snapshot.SourceDBSnapshotIdentifier)
	d.Set("source_region", snapshot.SourceRegion)
	d.Set("storage_type", snapshot.StorageType)
	d.Set("target_db_snapshot_identifier", snapshot.DBSnapshotIdentifier)
	d.Set("vpc_id", snapshot.VpcId)

	tags, err := ListTags(conn, arn)

	if err != nil {
		return diag.Errorf("error listing tags for RDS DB Snapshot (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("error setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("error setting tags_all: %s", err)
	}

	return nil
}

func resourceSnapshotCopyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("db_snapshot_arn").(string), o, n); err != nil {
			return diag.Errorf("error updating RDS DB Snapshot (%s) tags: %s", d.Get("db_snapshot_arn").(string), err)
		}
	}

	return resourceSnapshotCopyRead(ctx, d, meta)
}

func resourceSnapshotCopyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).RDSConn

	log.Printf("[INFO] Deleting RDS DB Snapshot %s", d.Id())

	in := &rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: aws.String(d.Id()),
	}

	_, err := conn.DeleteDBSnapshotWithContext(ctx, in)
	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBSnapshotNotFoundFault) {
		return nil
	}

	if err != nil {
		return diag.Errorf("error deleting RDS DB Snapshot (%s): %s", d.Id(), err)
	}

	return nil
}

func FindSnapshot(ctx context.Context, conn *rds.RDS, id string) (*rds.DBSnapshot, error) {
	in := &rds.DescribeDBSnapshotsInput{
		DBSnapshotIdentifier: aws.String(id),
	}
	out, err := conn.DescribeDBSnapshotsWithContext(ctx, in)

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBSnapshotNotFoundFault) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if out == nil || len(out.DBSnapshots) == 0 || out.DBSnapshots[0] == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.DBSnapshots[0], nil
}
func waitSnapshotCopyAvailable(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Waiting for Snapshot %s to become available...", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating"},
		Target:     []string{"available"},
		Refresh:    resourceSnapshotStateRefreshFunc(d, meta),
		Timeout:    d.Timeout(schema.TimeoutCreate),
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second, // Wait 30 secs before starting
	}

	// Wait, catching any errors
	_, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return err
	}

	return nil
}
