package lightsail

import (
	"context"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
		CreateWithoutTimeout: resourceDatabaseCreate,
		ReadWithoutTimeout:   resourceDatabaseRead,
		UpdateWithoutTimeout: resourceDatabaseUpdate,
		DeleteWithoutTimeout: resourceDatabaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseImport,
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
				Optional: true,
				Computed: true,
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

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	relationalDatabaseName := d.Get("relational_database_name").(string)
	input := &lightsail.CreateRelationalDatabaseInput{
		MasterDatabaseName:            aws.String(d.Get("master_database_name").(string)),
		MasterUsername:                aws.String(d.Get("master_username").(string)),
		RelationalDatabaseBlueprintId: aws.String(d.Get("blueprint_id").(string)),
		RelationalDatabaseBundleId:    aws.String(d.Get("bundle_id").(string)),
		RelationalDatabaseName:        aws.String(relationalDatabaseName),
	}

	if v, ok := d.GetOk("availability_zone"); ok {
		input.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("master_password"); ok {
		input.MasterUserPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_backup_window"); ok {
		input.PreferredBackupWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_maintenance_window"); ok {
		input.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk("publicly_accessible"); ok {
		input.PubliclyAccessible = aws.Bool(v.(bool))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateRelationalDatabaseWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Lightsail Relational Database (%s): %s", relationalDatabaseName, err)
	}

	d.SetId(relationalDatabaseName)

	if err := waitOperationWithContext(ctx, conn, output.Operations[0].Id); err != nil {
		return diag.Errorf("waiting for Lightsail Relational Database (%s) create: %s", d.Id(), err)
	}

	// Backup Retention is not a value you can pass on creation and defaults to true.
	// Forcing an update of the value after creation if the backup_retention_enabled value is false.
	if !d.Get("backup_retention_enabled").(bool) {
		input := &lightsail.UpdateRelationalDatabaseInput{
			ApplyImmediately:       aws.Bool(true),
			DisableBackupRetention: aws.Bool(true),
			RelationalDatabaseName: aws.String(d.Id()),
		}

		output, err := conn.UpdateRelationalDatabaseWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Lightsail Relational Database (%s) backup retention: %s", d.Id(), err)
		}

		if err := waitOperationWithContext(ctx, conn, output.Operations[0].Id); err != nil {
			return diag.Errorf("waiting for Lightsail Relational Database (%s) update: %s", d.Id(), err)
		}

		if err := waitDatabaseBackupRetentionModified(ctx, conn, aws.String(d.Id()), false); err != nil {
			return diag.Errorf("waiting for Lightsail Relational Database (%s) backup retention update: %s", d.Id(), err)
		}
	}

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	if _, err = waitDatabaseModified(ctx, conn, aws.String(d.Id())); err != nil {
		return diag.Errorf("waiting for Lightsail Relational Database (%s) to become available: %s", d.Id(), err)
	}

	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	// This is to support importing a resource that is not in a ready state.
	database, err := waitDatabaseModified(ctx, conn, aws.String(d.Id()))

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, lightsail.ErrCodeNotFoundException) {
		log.Printf("[WARN] Lightsail Relational Database (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Lightsail Relational Database (%s): %s", d.Id(), err)
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
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	if d.HasChangesExcept("apply_immediately", "final_snapshot_name", "skip_final_snapshot", "tags", "tags_all") {
		input := &lightsail.UpdateRelationalDatabaseInput{
			ApplyImmediately:       aws.Bool(d.Get("apply_immediately").(bool)),
			RelationalDatabaseName: aws.String(d.Id()),
		}

		if d.HasChange("backup_retention_enabled") {
			if d.Get("backup_retention_enabled").(bool) {
				input.EnableBackupRetention = aws.Bool(d.Get("backup_retention_enabled").(bool))
			} else {
				input.DisableBackupRetention = aws.Bool(true)
			}
		}

		if d.HasChange("ca_certificate_identifier") {
			input.CaCertificateIdentifier = aws.String(d.Get("ca_certificate_identifier").(string))
		}

		if d.HasChange("master_password") {
			input.MasterUserPassword = aws.String(d.Get("master_password").(string))
		}

		if d.HasChange("preferred_backup_window") {
			input.PreferredBackupWindow = aws.String(d.Get("preferred_backup_window").(string))
		}

		if d.HasChange("preferred_maintenance_window") {
			input.PreferredMaintenanceWindow = aws.String(d.Get("preferred_maintenance_window").(string))
		}

		if d.HasChange("publicly_accessible") {
			input.PubliclyAccessible = aws.Bool(d.Get("publicly_accessible").(bool))
		}

		output, err := conn.UpdateRelationalDatabaseWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating Lightsail Relational Database (%s): %s", d.Id(), err)
		}

		if err := waitOperationWithContext(ctx, conn, output.Operations[0].Id); err != nil {
			return diag.Errorf("waiting for Lightsail Relational Database (%s) update: %s", d.Id(), err)
		}

		if d.HasChange("backup_retention_enabled") {
			if err := waitDatabaseBackupRetentionModified(ctx, conn, aws.String(d.Id()), d.Get("backup_retention_enabled").(bool)); err != nil {
				return diag.Errorf("waiting for Lightsail Relational Database (%s) backup retention update: %s", d.Id(), err)
			}
		}

		if d.HasChange("publicly_accessible") {
			if err := waitDatabasePubliclyAccessibleModified(ctx, conn, aws.String(d.Id()), d.Get("publicly_accessible").(bool)); err != nil {
				return diag.Errorf("waiting for Lightsail Relational Database (%s) publicly accessible update: %s", d.Id(), err)
			}
		}

		// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
		if _, err = waitDatabaseModified(ctx, conn, aws.String(d.Id())); err != nil {
			return diag.Errorf("waiting for Lightsail Relational Database (%s) to become available: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Id(), o, n); err != nil {
			return diag.Errorf("updating Lightsail Relational Database (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).LightsailConn()

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	if _, err := waitDatabaseModified(ctx, conn, aws.String(d.Id())); err != nil {
		return diag.Errorf("waiting for Lightsail Relational Database (%s) to become available: %s", d.Id(), err)
	}

	skipFinalSnapshot := d.Get("skip_final_snapshot").(bool)

	input := &lightsail.DeleteRelationalDatabaseInput{
		RelationalDatabaseName: aws.String(d.Id()),
		SkipFinalSnapshot:      aws.Bool(skipFinalSnapshot),
	}

	if !skipFinalSnapshot {
		if name, present := d.GetOk("final_snapshot_name"); present {
			input.FinalRelationalDatabaseSnapshotName = aws.String(name.(string))
		} else {
			return diag.Errorf("Lightsail Database FinalRelationalDatabaseSnapshotName is required when a final snapshot is required")
		}
	}

	output, err := conn.DeleteRelationalDatabaseWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("deleting Lightsail Relational Database (%s): %s", d.Id(), err)
	}

	if err := waitOperationWithContext(ctx, conn, output.Operations[0].Id); err != nil {
		return diag.Errorf("waiting for Lightsail Relational Database (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func resourceDatabaseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required
	d.Set("skip_final_snapshot", true)
	return []*schema.ResourceData{d}, nil
}
