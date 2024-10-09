// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lightsail

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lightsail"
	"github.com/aws/aws-sdk-go-v2/service/lightsail/types"
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

const (
	ResNameDatabase = "Database"
)

// @SDKResource("aws_lightsail_database", name="Database")
// @Tags(identifierAttribute="id", resourceType="Database")
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
			names.AttrApplyImmediately: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
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
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_size": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			names.AttrEngine: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEngineVersion: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"final_snapshot_name": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 255),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z][0-9A-Za-z-]+[0-9A-Za-z]$`), "Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number"),
				),
			},
			"master_database_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 64),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "Must begin with a letter"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_]+$`), "Subsequent characters can be letters, underscores, or digits (0- 9)"),
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
					validation.StringMatch(regexache.MustCompile(`^[ -~][^@\/" ]+$`), "The password can include any printable ASCII character except \"/\", \"\"\", or \"@\". It cannot contain spaces."),
				),
			},
			"master_username": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexache.MustCompile(`^[A-Za-z]`), "Must begin with a letter"),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z_]+$`), "Subsequent characters can be letters, underscores, or digits (0- 9)"),
				),
			},
			"preferred_backup_window": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceADayWindowFormat,
			},
			names.AttrPreferredMaintenanceWindow: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidOnceAWeekWindowFormat,
			},
			names.AttrPubliclyAccessible: {
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
					validation.StringMatch(regexache.MustCompile(`^[^_.-]+[0-9A-Za-z-]+[^_.-]$`), "Must contain from 2 to 255 alphanumeric characters, or hyphens. The first and last character must be a letter or number"),
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
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	relationalDatabaseName := d.Get("relational_database_name").(string)
	input := &lightsail.CreateRelationalDatabaseInput{
		MasterDatabaseName:            aws.String(d.Get("master_database_name").(string)),
		MasterUsername:                aws.String(d.Get("master_username").(string)),
		RelationalDatabaseBlueprintId: aws.String(d.Get("blueprint_id").(string)),
		RelationalDatabaseBundleId:    aws.String(d.Get("bundle_id").(string)),
		RelationalDatabaseName:        aws.String(relationalDatabaseName),
		Tags:                          getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrAvailabilityZone); ok {
		input.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := d.GetOk("master_password"); ok {
		input.MasterUserPassword = aws.String(v.(string))
	}

	if v, ok := d.GetOk("preferred_backup_window"); ok {
		input.PreferredBackupWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPreferredMaintenanceWindow); ok {
		input.PreferredMaintenanceWindow = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPubliclyAccessible); ok {
		input.PubliclyAccessible = aws.Bool(v.(bool))
	}

	output, err := conn.CreateRelationalDatabase(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Lightsail Relational Database (%s): %s", relationalDatabaseName, err)
	}

	diagError := expandOperations(ctx, conn, output.Operations, types.OperationTypeCreateRelationalDatabase, ResNameDatabase, relationalDatabaseName)

	if diagError != nil {
		return diagError
	}

	d.SetId(relationalDatabaseName)

	// Backup Retention is not a value you can pass on creation and defaults to true.
	// Forcing an update of the value after creation if the backup_retention_enabled value is false.
	if !d.Get("backup_retention_enabled").(bool) {
		input := &lightsail.UpdateRelationalDatabaseInput{
			ApplyImmediately:       aws.Bool(true),
			DisableBackupRetention: aws.Bool(true),
			RelationalDatabaseName: aws.String(d.Id()),
		}

		output, err := conn.UpdateRelationalDatabase(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lightsail Relational Database (%s) backup retention: %s", d.Id(), err)
		}

		diagError := expandOperations(ctx, conn, output.Operations, types.OperationTypeUpdateRelationalDatabase, ResNameDatabase, relationalDatabaseName)

		if diagError != nil {
			return diagError
		}

		if err := waitDatabaseBackupRetentionModified(ctx, conn, aws.String(d.Id()), false); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Lightsail Relational Database (%s) backup retention update: %s", d.Id(), err)
		}
	}

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	if _, err = waitDatabaseModified(ctx, conn, aws.String(d.Id())); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Lightsail Relational Database (%s) to become available: %s", d.Id(), err)
	}

	return append(diags, resourceDatabaseRead(ctx, d, meta)...)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	// This is to support importing a resource that is not in a ready state.
	database, err := waitDatabaseModified(ctx, conn, aws.String(d.Id()))

	if !d.IsNewResource() && IsANotFoundError(err) {
		log.Printf("[WARN] Lightsail Relational Database (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Lightsail Relational Database (%s): %s", d.Id(), err)
	}

	rd := database.RelationalDatabase

	d.Set(names.AttrARN, rd.Arn)
	d.Set(names.AttrAvailabilityZone, rd.Location.AvailabilityZone)
	d.Set("backup_retention_enabled", rd.BackupRetentionEnabled)
	d.Set("blueprint_id", rd.RelationalDatabaseBlueprintId)
	d.Set("bundle_id", rd.RelationalDatabaseBundleId)
	d.Set("ca_certificate_identifier", rd.CaCertificateIdentifier)
	d.Set("cpu_count", rd.Hardware.CpuCount)
	d.Set(names.AttrCreatedAt, rd.CreatedAt.Format(time.RFC3339))
	d.Set("disk_size", rd.Hardware.DiskSizeInGb)
	d.Set(names.AttrEngine, rd.Engine)
	d.Set(names.AttrEngineVersion, rd.EngineVersion)
	d.Set("master_database_name", rd.MasterDatabaseName)
	d.Set("master_endpoint_address", rd.MasterEndpoint.Address)
	d.Set("master_endpoint_port", rd.MasterEndpoint.Port)
	d.Set("master_username", rd.MasterUsername)
	d.Set("preferred_backup_window", rd.PreferredBackupWindow)
	d.Set(names.AttrPreferredMaintenanceWindow, rd.PreferredMaintenanceWindow)
	d.Set(names.AttrPubliclyAccessible, rd.PubliclyAccessible)
	d.Set("ram_size", rd.Hardware.RamSizeInGb)
	d.Set("relational_database_name", rd.Name)
	d.Set("secondary_availability_zone", rd.SecondaryAvailabilityZone)
	d.Set("support_code", rd.SupportCode)

	setTagsOut(ctx, rd.Tags)

	return diags
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	if d.HasChangesExcept(names.AttrApplyImmediately, "final_snapshot_name", "skip_final_snapshot", names.AttrTags, names.AttrTagsAll) {
		input := &lightsail.UpdateRelationalDatabaseInput{
			ApplyImmediately:       aws.Bool(d.Get(names.AttrApplyImmediately).(bool)),
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

		if d.HasChange(names.AttrPreferredMaintenanceWindow) {
			input.PreferredMaintenanceWindow = aws.String(d.Get(names.AttrPreferredMaintenanceWindow).(string))
		}

		if d.HasChange(names.AttrPubliclyAccessible) {
			input.PubliclyAccessible = aws.Bool(d.Get(names.AttrPubliclyAccessible).(bool))
		}

		output, err := conn.UpdateRelationalDatabase(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Lightsail Relational Database (%s): %s", d.Id(), err)
		}

		diagError := expandOperations(ctx, conn, output.Operations, types.OperationTypeUpdateRelationalDatabase, ResNameDatabase, d.Id())

		if diagError != nil {
			return diagError
		}

		if d.HasChange("backup_retention_enabled") {
			if err := waitDatabaseBackupRetentionModified(ctx, conn, aws.String(d.Id()), d.Get("backup_retention_enabled").(bool)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Lightsail Relational Database (%s) backup retention update: %s", d.Id(), err)
			}
		}

		if d.HasChange(names.AttrPubliclyAccessible) {
			if err := waitDatabasePubliclyAccessibleModified(ctx, conn, aws.String(d.Id()), d.Get(names.AttrPubliclyAccessible).(bool)); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Lightsail Relational Database (%s) publicly accessible update: %s", d.Id(), err)
			}
		}

		// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
		if _, err = waitDatabaseModified(ctx, conn, aws.String(d.Id())); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Lightsail Relational Database (%s) to become available: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDatabaseRead(ctx, d, meta)...)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).LightsailClient(ctx)

	// Some Operations can complete before the Database enters the Available state. Added a waiter to make sure the Database is available before continuing.
	if _, err := waitDatabaseModified(ctx, conn, aws.String(d.Id())); err != nil {
		if IsANotFoundError(err) {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "waiting for Lightsail Relational Database (%s) to become available: %s", d.Id(), err)
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
			return sdkdiag.AppendErrorf(diags, "Lightsail Database FinalRelationalDatabaseSnapshotName is required when a final snapshot is required")
		}
	}

	output, err := conn.DeleteRelationalDatabase(ctx, input)

	if IsANotFoundError(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Lightsail Relational Database (%s): %s", d.Id(), err)
	}

	diagError := expandOperations(ctx, conn, output.Operations, types.OperationTypeDeleteRelationalDatabase, ResNameDatabase, d.Id())

	if diagError != nil {
		return diagError
	}

	return diags
}

func resourceDatabaseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	// Neither skip_final_snapshot nor final_snapshot_identifier can be fetched
	// from any API call, so we need to default skip_final_snapshot to true so
	// that final_snapshot_identifier is not required
	d.Set("skip_final_snapshot", true)
	return []*schema.ResourceData{d}, nil
}

func FindDatabaseById(ctx context.Context, conn *lightsail.Client, id string) (*types.RelationalDatabase, error) {
	in := &lightsail.GetRelationalDatabaseInput{
		RelationalDatabaseName: aws.String(id),
	}

	out, err := conn.GetRelationalDatabase(ctx, in)

	if IsANotFoundError(err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil || out.RelationalDatabase == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out.RelationalDatabase, nil
}
