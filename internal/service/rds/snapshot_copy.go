// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_db_snapshot_copy", name="DB Snapshot")
// @Tags(identifierAttribute="db_snapshot_arn")
// @Testing(tagsTest=false)
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
			names.AttrAllocatedStorage: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrAvailabilityZone: {
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
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
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
			names.AttrIOPS: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrKMSKeyID: {
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
			names.AttrPort: {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"presigned_url": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"shared_accounts": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
			names.AttrStorageType: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
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
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z][\w-]+`), "must contain only alphanumeric, and hyphen (-) characters"),
				),
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceSnapshotCopyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	targetDBSnapshotID := d.Get("target_db_snapshot_identifier").(string)
	input := &rds.CopyDBSnapshotInput{
		SourceDBSnapshotIdentifier: aws.String(d.Get("source_db_snapshot_identifier").(string)),
		Tags:                       getTagsIn(ctx),
		TargetDBSnapshotIdentifier: aws.String(targetDBSnapshotID),
	}

	if v, ok := d.GetOk("copy_tags"); ok {
		input.CopyTags = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("destination_region"); ok {
		input.DestinationRegion = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("option_group_name"); ok {
		input.OptionGroupName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("presigned_url"); ok {
		input.PreSignedUrl = aws.String(v.(string))
	}

	output, err := conn.CopyDBSnapshotWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RDS DB Snapshot Copy (%s): %s", targetDBSnapshotID, err)
	}

	d.SetId(aws.StringValue(output.DBSnapshot.DBSnapshotIdentifier))

	if err := waitDBSnapshotCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RDS DB Snapshot Copy (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("shared_accounts"); ok && v.(*schema.Set).Len() > 0 {
		input := &rds.ModifyDBSnapshotAttributeInput{
			AttributeName:        aws.String("restore"),
			DBSnapshotIdentifier: aws.String(d.Id()),
			ValuesToAdd:          flex.ExpandStringSet(v.(*schema.Set)),
		}

		_, err := conn.ModifyDBSnapshotAttributeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying RDS DB Snapshot (%s) attribute: %s", d.Id(), err)
		}
	}

	return append(diags, resourceSnapshotCopyRead(ctx, d, meta)...)
}

func resourceSnapshotCopyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	snapshot, err := FindDBSnapshotByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] RDS DB Snapshot (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Snapshot Copy (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(snapshot.DBSnapshotArn)
	d.Set(names.AttrAllocatedStorage, snapshot.AllocatedStorage)
	d.Set(names.AttrAvailabilityZone, snapshot.AvailabilityZone)
	d.Set("db_snapshot_arn", arn)
	d.Set(names.AttrEncrypted, snapshot.Encrypted)
	d.Set(names.AttrEngine, snapshot.Engine)
	d.Set(names.AttrEngineVersion, snapshot.EngineVersion)
	d.Set(names.AttrIOPS, snapshot.Iops)
	d.Set(names.AttrKMSKeyID, snapshot.KmsKeyId)
	d.Set("license_model", snapshot.LicenseModel)
	d.Set("option_group_name", snapshot.OptionGroupName)
	d.Set(names.AttrPort, snapshot.Port)
	d.Set("snapshot_type", snapshot.SnapshotType)
	d.Set("source_db_snapshot_identifier", snapshot.SourceDBSnapshotIdentifier)
	d.Set("source_region", snapshot.SourceRegion)
	d.Set(names.AttrStorageType, snapshot.StorageType)
	d.Set("target_db_snapshot_identifier", snapshot.DBSnapshotIdentifier)
	d.Set(names.AttrVPCID, snapshot.VpcId)

	input := &rds.DescribeDBSnapshotAttributesInput{
		DBSnapshotIdentifier: aws.String(d.Id()),
	}

	output, err := conn.DescribeDBSnapshotAttributesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RDS DB Snapshot (%s) attributes: %s", d.Id(), err)
	}

	d.Set("shared_accounts", flex.FlattenStringSet(output.DBSnapshotAttributesResult.DBSnapshotAttributes[0].AttributeValues))

	return diags
}

func resourceSnapshotCopyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	if d.HasChange("shared_accounts") {
		o, n := d.GetChange("shared_accounts")
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		additionList := ns.Difference(os)
		removalList := os.Difference(ns)

		input := &rds.ModifyDBSnapshotAttributeInput{
			AttributeName:        aws.String("restore"),
			DBSnapshotIdentifier: aws.String(d.Id()),
			ValuesToAdd:          flex.ExpandStringSet(additionList),
			ValuesToRemove:       flex.ExpandStringSet(removalList),
		}

		_, err := conn.ModifyDBSnapshotAttributeWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "modifying RDS DB Snapshot (%s) attributes: %s", d.Id(), err)
		}
	}

	return resourceSnapshotCopyRead(ctx, d, meta)
}

func resourceSnapshotCopyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RDSConn(ctx)

	log.Printf("[DEBUG] Deleting RDS DB Snapshot Copy: %s", d.Id())
	_, err := conn.DeleteDBSnapshotWithContext(ctx, &rds.DeleteDBSnapshotInput{
		DBSnapshotIdentifier: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBSnapshotNotFoundFault) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting RDS DB Snapshot Copy (%s): %s", d.Id(), err)
	}

	return diags
}
