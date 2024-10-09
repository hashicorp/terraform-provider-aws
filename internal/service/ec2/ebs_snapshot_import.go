// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ebs_snapshot_import", name="EBS Snapshot Import")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceEBSSnapshotImport() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEBSSnapshotImportCreate,
		ReadWithoutTimeout:   resourceEBSSnapshotImportRead,
		UpdateWithoutTimeout: resourceEBSSnapshotUpdate,
		DeleteWithoutTimeout: resourceEBSSnapshotDelete,

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(60 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"client_data": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrComment: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"upload_end": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsRFC3339Time,
						},
						"upload_size": {
							Type:     schema.TypeFloat,
							Optional: true,
							Computed: true,
						},
						"upload_start": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IsRFC3339Time,
						},
					},
				},
			},
			"data_encryption_key_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
				ForceNew: true,
			},
			"disk_container": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						names.AttrFormat: {
							Type:             schema.TypeString,
							Required:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.DiskImageFormat](),
						},
						names.AttrURL: {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							ExactlyOneOf: []string{"disk_container.0.user_bucket", "disk_container.0.url"},
						},
						"user_bucket": {
							Type:     schema.TypeList,
							Optional: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrS3Bucket: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
									"s3_key": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
							ExactlyOneOf: []string{"disk_container.0.user_bucket", "disk_container.0.url"},
						},
					},
				},
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"outpost_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"owner_alias": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"permanent_restore": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"role_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  defaultSnapshotImportRoleName,
			},
			"storage_tier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(enum.Slice(append(awstypes.TargetStorageTier.Values(""), targetStorageTierStandard)...), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"temporary_restore_days": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"volume_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVolumeSize: {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceEBSSnapshotImportCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.ImportSnapshotInput{
		ClientToken:       aws.String(id.UniqueId()),
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeImportSnapshotTask),
	}

	if v, ok := d.GetOk("client_data"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.ClientData = expandClientData(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("disk_container"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.DiskContainer = expandSnapshotDiskContainer(v.([]interface{})[0].(map[string]interface{}))
	}

	if v, ok := d.GetOk(names.AttrEncrypted); ok {
		input.Encrypted = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("role_name"); ok {
		input.RoleName = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenAWSErrMessageContains(ctx, iamPropagationTimeout,
		func() (interface{}, error) {
			return conn.ImportSnapshot(ctx, input)
		},
		errCodeInvalidParameter, "provided does not exist or does not have sufficient permissions")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EBS Snapshot Import: %s", err)
	}

	taskID := aws.ToString(outputRaw.(*ec2.ImportSnapshotOutput).ImportTaskId)
	output, err := waitEBSSnapshotImportComplete(ctx, conn, taskID, d.Timeout(schema.TimeoutCreate))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot Import (%s) create: %s", taskID, err)
	}

	d.SetId(aws.ToString(output.SnapshotId))

	if err := createTags(ctx, conn, d.Id(), getTagsIn(ctx)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting EBS Snapshot Import (%s) tags: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("storage_tier"); ok && v.(string) == string(awstypes.TargetStorageTierArchive) {
		_, err = conn.ModifySnapshotTier(ctx, &ec2.ModifySnapshotTierInput{
			SnapshotId:  aws.String(d.Id()),
			StorageTier: awstypes.TargetStorageTier(v.(string)),
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "setting EBS Snapshot Import (%s) Storage Tier: %s", d.Id(), err)
		}

		_, err = waitEBSSnapshotTierArchive(ctx, conn, d.Id(), ebsSnapshotArchivedTimeout)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EBS Snapshot Import (%s) Storage Tier archive: %s", d.Id(), err)
		}
	}

	return append(diags, resourceEBSSnapshotImportRead(ctx, d, meta)...)
}

func resourceEBSSnapshotImportRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	snapshot, err := findSnapshotByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EBS Snapshot %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EBS Snapshot (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("snapshot/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set("data_encryption_key_id", snapshot.DataEncryptionKeyId)
	d.Set(names.AttrDescription, snapshot.Description)
	d.Set(names.AttrEncrypted, snapshot.Encrypted)
	d.Set(names.AttrKMSKeyID, snapshot.KmsKeyId)
	d.Set("owner_alias", snapshot.OwnerAlias)
	d.Set(names.AttrOwnerID, snapshot.OwnerId)
	d.Set("storage_tier", snapshot.StorageTier)
	d.Set(names.AttrVolumeSize, snapshot.VolumeSize)

	setTagsOut(ctx, snapshot.Tags)

	return diags
}

func expandClientData(tfMap map[string]interface{}) *awstypes.ClientData {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ClientData{}

	if v, ok := tfMap[names.AttrComment].(string); ok && v != "" {
		apiObject.Comment = aws.String(v)
	}

	if v, ok := tfMap["upload_end"].(string); ok && v != "" {
		v, _ := time.Parse(time.RFC3339, v)

		apiObject.UploadEnd = aws.Time(v)
	}

	if v, ok := tfMap["upload_size"].(float64); ok && v != 0.0 {
		apiObject.UploadSize = aws.Float64(v)
	}

	if v, ok := tfMap["upload_start"].(string); ok {
		v, _ := time.Parse(time.RFC3339, v)

		apiObject.UploadStart = aws.Time(v)
	}

	return apiObject
}

func expandSnapshotDiskContainer(tfMap map[string]interface{}) *awstypes.SnapshotDiskContainer {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.SnapshotDiskContainer{}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap[names.AttrFormat].(string); ok && v != "" {
		apiObject.Format = aws.String(v)
	}

	if v, ok := tfMap[names.AttrURL].(string); ok && v != "" {
		apiObject.Url = aws.String(v)
	}

	if v, ok := tfMap["user_bucket"].([]interface{}); ok && len(v) > 0 && v[0] != nil {
		apiObject.UserBucket = expandUserBucket(v[0].(map[string]interface{}))
	}

	return apiObject
}

func expandUserBucket(tfMap map[string]interface{}) *awstypes.UserBucket {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.UserBucket{}

	if v, ok := tfMap[names.AttrS3Bucket].(string); ok && v != "" {
		apiObject.S3Bucket = aws.String(v)
	}

	if v, ok := tfMap["s3_key"].(string); ok && v != "" {
		apiObject.S3Key = aws.String(v)
	}

	return apiObject
}
