package lightsail

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	ResNameDatabase = "Database"
)

func ResourceDatabase() *schema.Resource {
	return &schema.Resource{
		Create: resourceDatabaseCreate,
		Read:   resourceDatabaseRead,
		Update: resourceDatabaseUpdate,
		Delete: resourceDatabaseDelete,
		Importer: &schema.ResourceImporter{
			State: ResourceDatabaseImport,
		},

		Schema: map[string]*schema.Schema{
			"apply_immediately": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"backup_retention_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"blueprint_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bundle_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"ca_certificate_identifier": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cpu_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_size": {
				Type:     schema.TypeFloat,
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
			"final_snapshot_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9-]+[A-Za-z0-9]$`), "Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number"),
				),
			},
			"master_database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z]`), "Must begin with a letter"),
					validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z_]+$`), "Subsequent characters can be letters, underscores, or digits (0- 9)"),
				),
			},
			"master_endpoint_address": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"master_endpoint_port": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"master_password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(8, 128),
					validation.StringMatch(regexp.MustCompile(`^[ -~][^@\/" ]+$`), "The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces."),
				),
			},
			"master_username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[A-Za-z]`), "Must begin with a letter"),
					validation.StringMatch(regexp.MustCompile(`^[0-9A-Za-z_]+$`), "Subsequent characters can be letters, underscores, or digits (0- 9)"),
				),
			},
			"preferred_backup_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceADayWindowFormat,
			},
			"preferred_maintenance_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			"publicly_accessible": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ram_size": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"relational_database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexp.MustCompile(`^[^._\-]+[0-9A-Za-z-]+[^._\-]$`), "Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number"),
				),
			},
			"secondary_availability_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"skip_final_snapshot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"support_code": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDatabaseCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	req := lightsail.CreateRelationalDatabaseInput{
		MasterDatabaseName:            aws.String(d.Get("master_database_name").(string)),
		MasterUsername:                aws.String(d.Get("master_username").(string)),
		RelationalDatabaseBlueprintId: aws.String(d.Get("blueprint_id").(string)),
		RelationalDatabaseBundleId:    aws.String(d.Get("bundle_id").(string)),
		RelationalDatabaseName:        aws.String(d.Get("relational_database_name").(string)),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		req.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("master_password"); ok {
		req.MasterUserPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_backup_window"); ok {
		req.PreferredBackupWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_maintenance_window"); ok {
		req.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("publicly_accessible"); ok {
		req.PubliclyAccessible = aws.Bool(v.(bool))
	}

	if len(tags) > 0 {
		req.Tags = Tags(tags.IgnoreAWS())
	}

	resp, err := conn.CreateRelationalDatabase(&req)
	if err != nil {
		return err
	}

	if len(resp.Operations) == 0 {
		return fmt.Errorf("No operations found for Create Relational Database request")
	}

	op := resp.Operations[0]
	d.SetId(d.Get("relational_database_name").(string))

	err = waitOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to become ready: %s", d.Id(), err)
	}

	// Backup Retention is not a value you can pass on creation and defaults to true.
	// Forcing an update of the value after creation if the backup_retention_enabled value is false.
	if v := d.Get("backup_retention_enabled"); v == false {
		log.Printf("[DEBUG] Lightsail Database (%s) backup_retention_enabled setting is false. Updating value.", d.Id())
		req := lightsail.UpdateRelationalDatabaseInput{
			ApplyImmediately:       aws.Bool(true),
			RelationalDatabaseName: aws.String(d.Id()),
			DisableBackupRetention: aws.Bool(true),
		}

		resp, err := conn.UpdateRelationalDatabase(&req)
		if err != nil {
			return err
		}

		if len(resp.Operations) == 0 {
			return fmt.Errorf("No operations found for Update Relational Database request")
		}

		op := resp.Operations[0]

		err = waitOperation(conn, op.Id)
		if err != nil {
			return fmt.Errorf("Error waiting for Relational Database (%s) to become ready: %s", d.Id(), err)
		}

		err = waitDatabaseBackupRetentionModified(conn, aws.String(d.Id()), aws.Bool(v.(bool)))
		if err != nil {
			return fmt.Errorf("Error waiting for Relational Database (%s) Backup Retention to be updated: %s", d.Id(), err)
		}

	}

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	_, err = waitDatabaseModified(conn, aws.String(d.Id()))
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to become available: %s", d.Id(), err)
	}

	return resourceDatabaseRead(d, meta)
}

func resourceDatabaseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	// This is to support importing a resource that is not in a ready state.
	database, err := waitDatabaseModified(conn, aws.String(d.Id()))

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		log.Printf("[WARN] Lightsail Relational Database (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading LightSail Relational Database (%s): %w", d.Id(), err)
	}

	rd := database.RelationalDatabase

	d.Set("arn", rd.Arn)
	d.Set("availability_zone", rd.Location.AvailabilityZone)
	d.Set("backup_retention_enabled", rd.BackupRetentionEnabled)
	d.Set("blueprint_id", rd.RelationalDatabaseBlueprintId)
	d.Set("bundle_id", rd.RelationalDatabaseBundleId)
	d.Set("ca_certificate_identifier", rd.CaCertificateIdentifier)
	d.Set("cpu_count", rd.Hardware.CpuCount)
	d.Set("created_at", rd.CreatedAt.Format(time.RFC3339))
	d.Set("disk_size", rd.Hardware.DiskSizeInGb)
	d.Set("engine", rd.Engine)
	d.Set("engine_version", rd.EngineVersion)
	d.Set("master_database_name", rd.MasterDatabaseName)
	d.Set("master_endpoint_address", rd.MasterEndpoint.Address)
	d.Set("master_endpoint_port", rd.MasterEndpoint.Port)
	d.Set("master_username", rd.MasterUsername)
	d.Set("preferred_backup_window", rd.PreferredBackupWindow)
	d.Set("preferred_maintenance_window", rd.PreferredMaintenanceWindow)
	d.Set("publicly_accessible", rd.PubliclyAccessible)
	d.Set("ram_size", rd.Hardware.RamSizeInGb)
	d.Set("relational_database_name", rd.Name)
	d.Set("secondary_availability_zone", rd.SecondaryAvailabilityZone)
	d.Set("support_code", rd.SupportCode)

	tags := KeyValueTags(rd.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceDatabaseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	_, err := waitDatabaseModified(conn, aws.String(d.Id()))
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to become available: %s", d.Id(), err)
	}

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)

	req := lightsail.DeleteRelationalDatabaseInput{
		RelationalDatabaseName: aws.String(d.Id()),
		SkipFinalSnapshot:      aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if name, present := d.GetOk("final_snapshot_name"); present {
			req.FinalRelationalDatabaseSnapshotName = aws.String(name.(string))
		} else {
			return fmt.Errorf("Lightsail Database FinalRelationalDatabaseSnapshotName is required when a final snapshot is required")
		}
	}

	resp, err := conn.DeleteRelationalDatabase(&req)

	if err != nil {
		return err
	}

	op := resp.Operations[0]

	err = waitOperation(conn, op.Id)
	if err != nil {
		return fmt.Errorf("Error waiting for Relational Database (%s) to Delete: %s", d.Id(), err)
	}

	return nil
}

func ResourceDatabaseImport(
	d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required
	d.Set("skip_final_snapshot", true)
	return []*schema.ResourceData{d}, nil
}

func resourceDatabaseUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).LightsailConn
	requestUpdate := false

	req := lightsail.UpdateRelationalDatabaseInput{
		ApplyImmediately:       aws.Bool(d.Get("apply_immediately").(bool)),
		RelationalDatabaseName: aws.String(d.Id()),
	}

	if d.HasChange("ca_certificate_identifier") {
		req.CaCertificateIdentifier = aws.String(d.Get("ca_certificate_identifier").(string))
		requestUpdate = true
	}

	if d.HasChange("backup_retention_enabled") {
		if d.Get("backup_retention_enabled").(bool) {
			req.EnableBackupRetention = aws.Bool(d.Get("backup_retention_enabled").(bool))
		} else {
			req.DisableBackupRetention = aws.Bool(true)
		}
		requestUpdate = true
	}

	if d.HasChange("master_password") {
		req.MasterUserPassword = aws.String(d.Get("master_password").(string))
		requestUpdate = true
	}

	if d.HasChange("preferred_backup_window") {
		req.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		requestUpdate = true
	}

	if d.HasChange("preferred_maintenance_window") {
		req.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		requestUpdate = true
	}

	if d.HasChange("publicly_accessible") {
		req.PubliclyAccessible = aws.Bool(d.Get("publicly_accessible").(bool))
		requestUpdate = true
	}

	if d.HasChange("tags") {
		o, n := d.GetChange("tags")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Database (%s) tags: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating Lightsail Database (%s) tags: %s", d.Id(), err)
		}
	}

	if requestUpdate {
		resp, err := conn.UpdateRelationalDatabase(&req)
		if err != nil {
			return err
		}

		if len(resp.Operations) == 0 {
			return fmt.Errorf("No operations found for Update Relational Database request")
		}

		op := resp.Operations[0]

		err = waitOperation(conn, op.Id)
		if err != nil {
			return fmt.Errorf("Error waiting for Relational Database (%s) to become ready: %s", d.Id(), err)
		}

		if d.HasChange("backup_retention_enabled") {
			err = waitDatabaseBackupRetentionModified(conn, aws.String(d.Id()), aws.Bool(d.Get("backup_retention_enabled").(bool)))
			if err != nil {
				return fmt.Errorf("Error waiting for Relational Database (%s) Backup Retention to be updated: %s", d.Id(), err)
			}
		}

		// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
		_, err = waitDatabaseModified(conn, aws.String(d.Id()))
		if err != nil {
			return fmt.Errorf("Error waiting for Relational Database (%s) to become available: %s", d.Id(), err)
		}
	}

	return resourceDatabaseRead(d, meta)
}
