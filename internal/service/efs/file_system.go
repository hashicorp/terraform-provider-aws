// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_efs_file_system", name="File System")
// @Tags(identifierAttribute="id")
func resourceFileSystem() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFileSystemCreate,
		ReadWithoutTimeout:   resourceFileSystemRead,
		UpdateWithoutTimeout: resourceFileSystemUpdate,
		DeleteWithoutTimeout: resourceFileSystemDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"availability_zone_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"creation_token": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(0, 64),
			},
			names.AttrDNSName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrEncrypted: {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			names.AttrKMSKeyID: {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"lifecycle_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 3,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"transition_to_archive": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TransitionToArchiveRules](),
						},
						"transition_to_ia": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TransitionToIARules](),
						},
						"transition_to_primary_storage_class": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.TransitionToPrimaryStorageClassRules](),
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"number_of_mount_targets": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"performance_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.PerformanceMode](),
			},
			"protection": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"replication_overwrite": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice(enum.Slice(
								awstypes.ReplicationOverwriteProtectionEnabled,
								awstypes.ReplicationOverwriteProtectionDisabled,
							), false),
						},
					},
				},
			},
			"provisioned_throughput_in_mibps": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"size_in_bytes": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrValue: {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"value_in_ia": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"value_in_standard": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"throughput_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ThroughputModeBursting,
				ValidateDiagFunc: enum.Validate[awstypes.ThroughputMode](),
			},
		},
	}
}

func resourceFileSystemCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	var creationToken string
	if v, ok := d.GetOk("creation_token"); ok {
		creationToken = v.(string)
	} else {
		creationToken = id.UniqueId()
	}
	throughputMode := awstypes.ThroughputMode(d.Get("throughput_mode").(string))
	input := &efs.CreateFileSystemInput{
		CreationToken:  aws.String(creationToken),
		Tags:           getTagsIn(ctx),
		ThroughputMode: throughputMode,
	}

	if v, ok := d.GetOk("availability_zone_name"); ok {
		input.AvailabilityZoneName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("performance_mode"); ok {
		input.PerformanceMode = awstypes.PerformanceMode(v.(string))
	}

	if throughputMode == awstypes.ThroughputModeProvisioned {
		input.ProvisionedThroughputInMibps = aws.Float64(d.Get("provisioned_throughput_in_mibps").(float64))
	}

	encrypted, hasEncrypted := d.GetOk(names.AttrEncrypted)
	if hasEncrypted {
		input.Encrypted = aws.Bool(encrypted.(bool))
	}

	kmsKeyID, hasKmsKeyID := d.GetOk(names.AttrKMSKeyID)
	if hasKmsKeyID {
		input.KmsKeyId = aws.String(kmsKeyID.(string))
	}

	if encrypted == false && hasKmsKeyID {
		return sdkdiag.AppendFromErr(diags, errors.New("encrypted must be set to true when kms_key_id is specified"))
	}

	output, err := conn.CreateFileSystem(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EFS File System: %s", err)
	}

	d.SetId(aws.ToString(output.FileSystemId))

	if _, err := waitFileSystemAvailable(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS File System (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("lifecycle_policy"); ok {
		input := &efs.PutLifecycleConfigurationInput{
			FileSystemId:      aws.String(d.Id()),
			LifecyclePolicies: expandLifecyclePolicies(v.([]interface{})),
		}

		_, err := conn.PutLifecycleConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting EFS File System (%s) lifecycle configuration: %s", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("protection"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input := expandUpdateFileSystemProtectionInput(d.Id(), v.([]interface{})[0].(map[string]interface{}))

		_, err := conn.UpdateFileSystemProtection(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EFS File System (%s) protection: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFileSystemRead(ctx, d, meta)...)
}

func resourceFileSystemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	fs, err := findFileSystemByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS File System (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS File System (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, fs.FileSystemArn)
	d.Set("availability_zone_id", fs.AvailabilityZoneId)
	d.Set("availability_zone_name", fs.AvailabilityZoneName)
	d.Set("creation_token", fs.CreationToken)
	d.Set(names.AttrDNSName, meta.(*conns.AWSClient).RegionalHostname(ctx, d.Id()+".efs"))
	d.Set(names.AttrEncrypted, fs.Encrypted)
	d.Set(names.AttrKMSKeyID, fs.KmsKeyId)
	d.Set(names.AttrName, fs.Name)
	d.Set("number_of_mount_targets", fs.NumberOfMountTargets)
	d.Set(names.AttrOwnerID, fs.OwnerId)
	d.Set("performance_mode", fs.PerformanceMode)
	if err := d.Set("protection", flattenFileSystemProtectionDescription(fs.FileSystemProtection)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting protection: %s", err)
	}
	d.Set("provisioned_throughput_in_mibps", fs.ProvisionedThroughputInMibps)
	if err := d.Set("size_in_bytes", flattenFileSystemSize(fs.SizeInBytes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting size_in_bytes: %s", err)
	}
	d.Set("throughput_mode", fs.ThroughputMode)

	setTagsOut(ctx, fs.Tags)

	output, err := conn.DescribeLifecycleConfiguration(ctx, &efs.DescribeLifecycleConfigurationInput{
		FileSystemId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS File System (%s) lifecycle configuration: %s", d.Id(), err)
	}

	if err := d.Set("lifecycle_policy", flattenLifecyclePolicies(output.LifecyclePolicies)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting lifecycle_policy: %s", err)
	}

	return diags
}

func resourceFileSystemUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	if d.HasChanges("provisioned_throughput_in_mibps", "throughput_mode") {
		throughputMode := awstypes.ThroughputMode(d.Get("throughput_mode").(string))
		input := &efs.UpdateFileSystemInput{
			FileSystemId:   aws.String(d.Id()),
			ThroughputMode: throughputMode,
		}

		if throughputMode == awstypes.ThroughputModeProvisioned {
			input.ProvisionedThroughputInMibps = aws.Float64(d.Get("provisioned_throughput_in_mibps").(float64))
		}

		_, err := conn.UpdateFileSystem(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EFS File System (%s): %s", d.Id(), err)
		}

		if _, err := waitFileSystemAvailable(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for EFS File System (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("lifecycle_policy") {
		input := &efs.PutLifecycleConfigurationInput{
			FileSystemId:      aws.String(d.Id()),
			LifecyclePolicies: expandLifecyclePolicies(d.Get("lifecycle_policy").([]interface{})),
		}

		// Prevent the following error during removal:
		// InvalidParameter: 1 validation error(s) found.
		// - missing required field, PutLifecycleConfigurationInput.LifecyclePolicies.
		if input.LifecyclePolicies == nil {
			input.LifecyclePolicies = []awstypes.LifecyclePolicy{}
		}

		_, err := conn.PutLifecycleConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "putting EFS File System (%s) lifecycle configuration: %s", d.Id(), err)
		}
	}

	if d.HasChanges("protection") {
		if v, ok := d.GetOk("protection"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			input := expandUpdateFileSystemProtectionInput(d.Id(), v.([]interface{})[0].(map[string]interface{}))

			_, err := conn.UpdateFileSystemProtection(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EFS File System (%s) protection: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceFileSystemRead(ctx, d, meta)...)
}

func resourceFileSystemDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	log.Printf("[DEBUG] Deleting EFS File System: %s", d.Id())
	_, err := conn.DeleteFileSystem(ctx, &efs.DeleteFileSystemInput{
		FileSystemId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS File System (%s): %s", d.Id(), err)
	}

	if _, err := waitFileSystemDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS File System (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFileSystem(ctx context.Context, conn *efs.Client, input *efs.DescribeFileSystemsInput, filter tfslices.Predicate[*awstypes.FileSystemDescription]) (*awstypes.FileSystemDescription, error) {
	output, err := findFileSystems(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFileSystems(ctx context.Context, conn *efs.Client, input *efs.DescribeFileSystemsInput, filter tfslices.Predicate[*awstypes.FileSystemDescription]) ([]awstypes.FileSystemDescription, error) {
	var output []awstypes.FileSystemDescription

	pages := efs.NewDescribeFileSystemsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.FileSystemNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.FileSystems {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findFileSystemByID(ctx context.Context, conn *efs.Client, id string) (*awstypes.FileSystemDescription, error) {
	input := &efs.DescribeFileSystemsInput{
		FileSystemId: aws.String(id),
	}

	output, err := findFileSystem(ctx, conn, input, tfslices.PredicateTrue[*awstypes.FileSystemDescription]())

	if err != nil {
		return nil, err
	}

	if state := output.LifeCycleState; state == awstypes.LifeCycleStateDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(state),
			LastRequest: input,
		}
	}

	return output, nil
}

func statusFileSystemLifeCycleState(ctx context.Context, conn *efs.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFileSystemByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.LifeCycleState), nil
	}
}

func waitFileSystemAvailable(ctx context.Context, conn *efs.Client, fileSystemID string) (*awstypes.FileSystemDescription, error) { //nolint:unparam
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.LifeCycleStateCreating, awstypes.LifeCycleStateUpdating),
		Target:     enum.Slice(awstypes.LifeCycleStateAvailable),
		Refresh:    statusFileSystemLifeCycleState(ctx, conn, fileSystemID),
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileSystemDescription); ok {
		return output, err
	}

	return nil, err
}

func waitFileSystemDeleted(ctx context.Context, conn *efs.Client, fileSystemID string) (*awstypes.FileSystemDescription, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateDeleting),
		Target:     []string{},
		Refresh:    statusFileSystemLifeCycleState(ctx, conn, fileSystemID),
		Timeout:    timeout,
		Delay:      2 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileSystemDescription); ok {
		return output, err
	}

	return nil, err
}

func flattenLifecyclePolicies(apiObjects []awstypes.LifecyclePolicy) []interface{} {
	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := make(map[string]interface{})

		tfMap["transition_to_archive"] = apiObject.TransitionToArchive
		tfMap["transition_to_ia"] = apiObject.TransitionToIA
		tfMap["transition_to_primary_storage_class"] = apiObject.TransitionToPrimaryStorageClass

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func expandLifecyclePolicies(tfList []interface{}) []awstypes.LifecyclePolicy {
	var apiObjects []awstypes.LifecyclePolicy

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}

		apiObject := awstypes.LifecyclePolicy{}

		if v, ok := tfMap["transition_to_archive"].(string); ok && v != "" {
			apiObject.TransitionToArchive = awstypes.TransitionToArchiveRules(v)
		}

		if v, ok := tfMap["transition_to_ia"].(string); ok && v != "" {
			apiObject.TransitionToIA = awstypes.TransitionToIARules(v)
		}

		if v, ok := tfMap["transition_to_primary_storage_class"].(string); ok && v != "" {
			apiObject.TransitionToPrimaryStorageClass = awstypes.TransitionToPrimaryStorageClassRules(v)
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenFileSystemSize(apiObject *awstypes.FileSystemSize) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrValue: apiObject.Value,
	}

	if apiObject.ValueInIA != nil {
		m["value_in_ia"] = aws.ToInt64(apiObject.ValueInIA)
	}

	if apiObject.ValueInStandard != nil {
		m["value_in_standard"] = aws.ToInt64(apiObject.ValueInStandard)
	}

	return []interface{}{m}
}

func expandUpdateFileSystemProtectionInput(fileSystemID string, tfMap map[string]interface{}) *efs.UpdateFileSystemProtectionInput {
	if tfMap == nil {
		return nil
	}

	apiObject := &efs.UpdateFileSystemProtectionInput{
		FileSystemId: aws.String(fileSystemID),
	}

	if v, ok := tfMap["replication_overwrite"].(string); ok && v != "" {
		apiObject.ReplicationOverwriteProtection = awstypes.ReplicationOverwriteProtection(v)
	}

	return apiObject
}

func flattenFileSystemProtectionDescription(apiObject *awstypes.FileSystemProtectionDescription) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{}

	tfMap["replication_overwrite"] = apiObject.ReplicationOverwriteProtection

	return []interface{}{tfMap}
}
