// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_efs_access_point", name="Access Point")
// @Tags(identifierAttribute="id")
func resourceAccessPoint() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccessPointCreate,
		ReadWithoutTimeout:   resourceAccessPointRead,
		UpdateWithoutTimeout: resourceAccessPointUpdate,
		DeleteWithoutTimeout: resourceAccessPointDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"file_system_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrFileSystemID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"posix_user": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gid": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
						"secondary_gids": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeInt},
							Optional: true,
							ForceNew: true,
						},
						"uid": {
							Type:     schema.TypeInt,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"root_directory": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ForceNew: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"creation_info": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							ForceNew: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"owner_gid": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									"owner_uid": {
										Type:     schema.TypeInt,
										Required: true,
										ForceNew: true,
									},
									names.AttrPermissions: {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
									},
								},
							},
						},
						names.AttrPath: {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ForceNew: true,
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceAccessPointCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	fsID := d.Get(names.AttrFileSystemID).(string)
	input := &efs.CreateAccessPointInput{
		FileSystemId: aws.String(fsID),
		Tags:         getTagsIn(ctx),
	}

	if v, ok := d.GetOk("posix_user"); ok {
		input.PosixUser = expandAccessPointPOSIXUser(v.([]interface{}))
	}

	if v, ok := d.GetOk("root_directory"); ok {
		input.RootDirectory = expandAccessPointRootDirectory(v.([]interface{}))
	}

	output, err := conn.CreateAccessPoint(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EFS Access Point for File System (%s): %s", fsID, err)
	}

	d.SetId(aws.ToString(output.AccessPointId))

	if _, err := waitAccessPointCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS Access Point (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAccessPointRead(ctx, d, meta)...)
}

func resourceAccessPointRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	ap, err := findAccessPointByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EFS Access Point (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EFS Access Point (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, ap.AccessPointArn)
	fsID := aws.ToString(ap.FileSystemId)
	fsARN := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  "file-system/" + fsID,
		Service:   "elasticfilesystem",
	}.String()
	d.Set("file_system_arn", fsARN)
	d.Set(names.AttrFileSystemID, fsID)
	d.Set(names.AttrOwnerID, ap.OwnerId)
	if err := d.Set("posix_user", flattenAccessPointPOSIXUser(ap.PosixUser)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting posix_user: %s", err)
	}
	if err := d.Set("root_directory", flattenAccessPointRootDirectory(ap.RootDirectory)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting root_directory: %s", err)
	}

	setTagsOut(ctx, ap.Tags)

	return diags
}

func resourceAccessPointUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceAccessPointRead(ctx, d, meta)...)
}

func resourceAccessPointDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EFSClient(ctx)

	log.Printf("[DEBUG] Deleting EFS Access Point: %s", d.Id())
	_, err := conn.DeleteAccessPoint(ctx, &efs.DeleteAccessPointInput{
		AccessPointId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.AccessPointNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EFS Access Point (%s): %s", d.Id(), err)
	}

	if _, err := waitAccessPointDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for EFS Access Point (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findAccessPoint(ctx context.Context, conn *efs.Client, input *efs.DescribeAccessPointsInput, filter tfslices.Predicate[*awstypes.AccessPointDescription]) (*awstypes.AccessPointDescription, error) {
	output, err := findAccessPoints(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findAccessPoints(ctx context.Context, conn *efs.Client, input *efs.DescribeAccessPointsInput, filter tfslices.Predicate[*awstypes.AccessPointDescription]) ([]awstypes.AccessPointDescription, error) {
	var output []awstypes.AccessPointDescription

	pages := efs.NewDescribeAccessPointsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.AccessPointNotFound](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		for _, v := range page.AccessPoints {
			if filter(&v) {
				output = append(output, v)
			}
		}
	}

	return output, nil
}

func findAccessPointByID(ctx context.Context, conn *efs.Client, id string) (*awstypes.AccessPointDescription, error) {
	input := &efs.DescribeAccessPointsInput{
		AccessPointId: aws.String(id),
	}

	output, err := findAccessPoint(ctx, conn, input, tfslices.PredicateTrue[*awstypes.AccessPointDescription]())

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

func statusAccessPointLifeCycleState(ctx context.Context, conn *efs.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findAccessPointByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.LifeCycleState), nil
	}
}

func waitAccessPointCreated(ctx context.Context, conn *efs.Client, id string) (*awstypes.AccessPointDescription, error) {
	const (
		timeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateCreating),
		Target:  enum.Slice(awstypes.LifeCycleStateAvailable),
		Refresh: statusAccessPointLifeCycleState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}

func waitAccessPointDeleted(ctx context.Context, conn *efs.Client, id string) (*awstypes.AccessPointDescription, error) {
	const (
		accessPointCreatedTimeout = 10 * time.Minute
		accessPointDeletedTimeout = 10 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateDeleting),
		Target:  []string{},
		Refresh: statusAccessPointLifeCycleState(ctx, conn, id),
		Timeout: accessPointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}

func expandAccessPointPOSIXUser(tfList []interface{}) *awstypes.PosixUser {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.PosixUser{
		Gid: aws.Int64(int64(tfMap["gid"].(int))),
		Uid: aws.Int64(int64(tfMap["uid"].(int))),
	}

	if v, ok := tfMap["secondary_gids"].(*schema.Set); ok && len(v.List()) > 0 {
		apiObject.SecondaryGids = flex.ExpandInt64ValueSet(v)
	}

	return apiObject
}

func expandAccessPointRootDirectory(tfList []interface{}) *awstypes.RootDirectory {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.RootDirectory{}

	if v, ok := tfMap[names.AttrPath]; ok {
		apiObject.Path = aws.String(v.(string))
	}

	if v, ok := tfMap["creation_info"]; ok {
		apiObject.CreationInfo = expandAccessPointRootDirectoryCreationInfo(v.([]interface{}))
	}

	return apiObject
}

func expandAccessPointRootDirectoryCreationInfo(tfList []interface{}) *awstypes.CreationInfo {
	if len(tfList) < 1 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})
	apiObject := &awstypes.CreationInfo{
		OwnerGid:    aws.Int64(int64(tfMap["owner_gid"].(int))),
		OwnerUid:    aws.Int64(int64(tfMap["owner_uid"].(int))),
		Permissions: aws.String(tfMap[names.AttrPermissions].(string)),
	}

	return apiObject
}

func flattenAccessPointPOSIXUser(apiObject *awstypes.PosixUser) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"gid":            aws.ToInt64(apiObject.Gid),
		"uid":            aws.ToInt64(apiObject.Uid),
		"secondary_gids": apiObject.SecondaryGids,
	}

	return []interface{}{tfMap}
}

func flattenAccessPointRootDirectory(apiObject *awstypes.RootDirectory) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"creation_info": flattenAccessPointRootDirectoryCreationInfo(apiObject.CreationInfo),
		names.AttrPath:  aws.ToString(apiObject.Path),
	}

	return []interface{}{tfMap}
}

func flattenAccessPointRootDirectoryCreationInfo(apiObject *awstypes.CreationInfo) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"owner_gid":           aws.ToInt64(apiObject.OwnerGid),
		"owner_uid":           aws.ToInt64(apiObject.OwnerUid),
		names.AttrPermissions: aws.ToString(apiObject.Permissions),
	}

	return []interface{}{tfMap}
}
