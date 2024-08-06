// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_storagegateway_nfs_file_share", name="NFS File Share")
// @Tags(identifierAttribute="arn")
func resourceNFSFileShare() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNFSFileShareCreate,
		ReadWithoutTimeout:   resourceNFSFileShareRead,
		UpdateWithoutTimeout: resourceNFSFileShareUpdate,
		DeleteWithoutTimeout: resourceNFSFileShareDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"audit_destination_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"bucket_region": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				RequiredWith: []string{"vpc_endpoint_dns_name"},
			},
			"cache_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_stale_timeout_in_seconds": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntBetween(300, 2592000),
						},
					},
				},
			},
			"client_list": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				MaxItems: 100,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.Any(
						verify.ValidIPv4CIDRNetworkAddress,
						validation.IsIPv4Address,
					),
				},
			},
			"default_storage_class": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      defaultStorageClassS3Standard,
				ValidateFunc: validation.StringInSlice(defaultStorageClass_Values(), false),
			},
			"file_share_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
			"fileshare_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"guess_mime_type_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"kms_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrKMSKeyARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"nfs_file_share_defaults": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"directory_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "0777",
							ValidateFunc: validLinuxFileMode,
						},
						"file_mode": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "0666",
							ValidateFunc: validLinuxFileMode,
						},
						"group_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "65534",
							ValidateFunc: verify.Valid4ByteASN,
						},
						names.AttrOwnerID: {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "65534",
							ValidateFunc: verify.Valid4ByteASN,
						},
					},
				},
			},
			"notification_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "{}",
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^\{[\w\s:\{\}\[\]"]*}$`), ""),
					validation.StringLenBetween(2, 100),
				),
			},
			"object_acl": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.ObjectACLPrivate,
				ValidateDiagFunc: enum.Validate[awstypes.ObjectACL](),
			},
			names.AttrPath: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"read_only": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"requester_pays": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"squash": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      squashRootSquash,
				ValidateFunc: validation.StringInSlice(squash_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_endpoint_dns_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNFSFileShareCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	fileShareDefaults, err := expandNFSFileShareDefaults(d.Get("nfs_file_share_defaults").([]interface{}))
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := &storagegateway.CreateNFSFileShareInput{
		ClientList:           flex.ExpandStringValueSet(d.Get("client_list").(*schema.Set)),
		ClientToken:          aws.String(id.UniqueId()),
		DefaultStorageClass:  aws.String(d.Get("default_storage_class").(string)),
		GatewayARN:           aws.String(d.Get("gateway_arn").(string)),
		GuessMIMETypeEnabled: aws.Bool(d.Get("guess_mime_type_enabled").(bool)),
		KMSEncrypted:         aws.Bool(d.Get("kms_encrypted").(bool)),
		LocationARN:          aws.String(d.Get("location_arn").(string)),
		NFSFileShareDefaults: fileShareDefaults,
		ObjectACL:            awstypes.ObjectACL(d.Get("object_acl").(string)),
		ReadOnly:             aws.Bool(d.Get("read_only").(bool)),
		RequesterPays:        aws.Bool(d.Get("requester_pays").(bool)),
		Role:                 aws.String(d.Get(names.AttrRoleARN).(string)),
		Squash:               aws.String(d.Get("squash").(string)),
		Tags:                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("audit_destination_arn"); ok {
		input.AuditDestinationARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("bucket_region"); ok {
		input.BucketRegion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cache_attributes"); ok {
		input.CacheAttributes = expandNFSFileShareCacheAttributes(v.([]interface{}))
	}

	if v, ok := d.GetOk("file_share_name"); ok {
		input.FileShareName = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KMSKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("notification_policy"); ok {
		input.NotificationPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("vpc_endpoint_dns_name"); ok {
		input.VPCEndpointDNSName = aws.String(v.(string))
	}

	output, err := conn.CreateNFSFileShare(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway NFS File Share: %s", err)
	}

	d.SetId(aws.ToString(output.FileShareARN))

	if _, err = waitNFSFileShareCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway NFS File Share (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceNFSFileShareRead(ctx, d, meta)...)
}

func resourceNFSFileShareRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	fileshare, err := findNFSFileShareByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway NFS File Share (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway NFS File Share (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, fileshare.FileShareARN)
	d.Set("audit_destination_arn", fileshare.AuditDestinationARN)
	d.Set("bucket_region", fileshare.BucketRegion)
	if err := d.Set("cache_attributes", flattenNFSFileShareCacheAttributes(fileshare.CacheAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cache_attributes: %s", err)
	}
	if err := d.Set("client_list", flex.FlattenStringValueSet(fileshare.ClientList)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting client_list: %s", err)
	}
	d.Set("default_storage_class", fileshare.DefaultStorageClass)
	d.Set("file_share_name", fileshare.FileShareName)
	d.Set("fileshare_id", fileshare.FileShareId)
	d.Set("gateway_arn", fileshare.GatewayARN)
	d.Set("guess_mime_type_enabled", fileshare.GuessMIMETypeEnabled)
	d.Set("kms_encrypted", fileshare.KMSEncrypted)
	d.Set(names.AttrKMSKeyARN, fileshare.KMSKey)
	d.Set("location_arn", fileshare.LocationARN)
	if err := d.Set("nfs_file_share_defaults", flattenNFSFileShareDefaults(fileshare.NFSFileShareDefaults)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting nfs_file_share_defaults: %s", err)
	}
	d.Set("notification_policy", fileshare.NotificationPolicy)
	d.Set("object_acl", fileshare.ObjectACL)
	d.Set(names.AttrPath, fileshare.Path)
	d.Set("read_only", fileshare.ReadOnly)
	d.Set("requester_pays", fileshare.RequesterPays)
	d.Set(names.AttrRoleARN, fileshare.Role)
	d.Set("squash", fileshare.Squash)
	d.Set("vpc_endpoint_dns_name", fileshare.VPCEndpointDNSName)

	setTagsOut(ctx, fileshare.Tags)

	return diags
}

func resourceNFSFileShareUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		fileShareDefaults, err := expandNFSFileShareDefaults(d.Get("nfs_file_share_defaults").([]interface{}))
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		input := &storagegateway.UpdateNFSFileShareInput{
			ClientList:           flex.ExpandStringValueSet(d.Get("client_list").(*schema.Set)),
			DefaultStorageClass:  aws.String(d.Get("default_storage_class").(string)),
			FileShareARN:         aws.String(d.Id()),
			GuessMIMETypeEnabled: aws.Bool(d.Get("guess_mime_type_enabled").(bool)),
			KMSEncrypted:         aws.Bool(d.Get("kms_encrypted").(bool)),
			NFSFileShareDefaults: fileShareDefaults,
			ObjectACL:            awstypes.ObjectACL(d.Get("object_acl").(string)),
			ReadOnly:             aws.Bool(d.Get("read_only").(bool)),
			RequesterPays:        aws.Bool(d.Get("requester_pays").(bool)),
			Squash:               aws.String(d.Get("squash").(string)),
		}

		if v, ok := d.GetOk("audit_destination_arn"); ok {
			input.AuditDestinationARN = aws.String(v.(string))
		}

		if v, ok := d.GetOk("cache_attributes"); ok {
			input.CacheAttributes = expandNFSFileShareCacheAttributes(v.([]interface{}))
		}

		if v, ok := d.GetOk("file_share_name"); ok {
			input.FileShareName = aws.String(v.(string))
		}

		if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
			input.KMSKey = aws.String(v.(string))
		}

		if v, ok := d.GetOk("notification_policy"); ok {
			input.NotificationPolicy = aws.String(v.(string))
		}

		_, err = conn.UpdateNFSFileShare(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway NFS File Share (%s): %s", d.Id(), err)
		}

		if _, err = waitNFSFileShareUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway NFS File Share (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceNFSFileShareRead(ctx, d, meta)...)
}

func findNFSFileShareByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*awstypes.NFSFileShareInfo, error) {
	input := &storagegateway.DescribeNFSFileSharesInput{
		FileShareARNList: []string{arn},
	}

	return findNFSFileShare(ctx, conn, input)
}

func findNFSFileShare(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeNFSFileSharesInput) (*awstypes.NFSFileShareInfo, error) {
	output, err := findNFSFileShares(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findNFSFileShares(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeNFSFileSharesInput) ([]awstypes.NFSFileShareInfo, error) {
	output, err := conn.DescribeNFSFileShares(ctx, input)

	if operationErrorCode(err) == operationErrCodeFileShareNotFound {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.NFSFileShareInfoList, nil
}

func statusNFSFileShare(ctx context.Context, conn *storagegateway.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findNFSFileShareByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.FileShareStatus), nil
	}
}

func waitNFSFileShareCreated(ctx context.Context, conn *storagegateway.Client, arn string, timeout time.Duration) (*awstypes.NFSFileShareInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fileShareStatusCreating},
		Target:  []string{fileShareStatusAvailable},
		Refresh: statusNFSFileShare(ctx, conn, arn),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NFSFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func waitNFSFileShareUpdated(ctx context.Context, conn *storagegateway.Client, arn string, timeout time.Duration) (*awstypes.NFSFileShareInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fileShareStatusUpdating},
		Target:  []string{fileShareStatusAvailable},
		Refresh: statusNFSFileShare(ctx, conn, arn),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NFSFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func waitNFSFileShareDeleted(ctx context.Context, conn *storagegateway.Client, arn string, timeout time.Duration) (*awstypes.NFSFileShareInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{fileShareStatusAvailable, fileShareStatusDeleting, fileShareStatusForceDeleting},
		Target:         []string{},
		Refresh:        statusNFSFileShare(ctx, conn, arn),
		Timeout:        timeout,
		Delay:          5 * time.Second,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.NFSFileShareInfo); ok {
		return output, err
	}

	return nil, err
}

func resourceNFSFileShareDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting Storage Gateway NFS File Share: %s", d.Id())
	_, err := conn.DeleteFileShare(ctx, &storagegateway.DeleteFileShareInput{
		FileShareARN: aws.String(d.Id()),
	})

	if operationErrorCode(err) == operationErrCodeFileShareNotFound {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway NFS File Share (%s): %s", d.Id(), err)
	}

	if _, err = waitNFSFileShareDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway NFS File Share (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func expandNFSFileShareDefaults(tfList []interface{}) (*awstypes.NFSFileShareDefaults, error) {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil, nil
	}

	tfMap := tfList[0].(map[string]interface{})

	groupID, err := strconv.ParseInt(tfMap["group_id"].(string), 10, 64)
	if err != nil {
		return nil, err
	}

	ownerID, err := strconv.ParseInt(tfMap[names.AttrOwnerID].(string), 10, 64)
	if err != nil {
		return nil, err
	}

	apiObject := &awstypes.NFSFileShareDefaults{
		DirectoryMode: aws.String(tfMap["directory_mode"].(string)),
		FileMode:      aws.String(tfMap["file_mode"].(string)),
		GroupId:       aws.Int64(groupID),
		OwnerId:       aws.Int64(ownerID),
	}

	return apiObject, nil
}

func flattenNFSFileShareDefaults(apiObject *awstypes.NFSFileShareDefaults) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"directory_mode":  aws.ToString(apiObject.DirectoryMode),
		"file_mode":       aws.ToString(apiObject.FileMode),
		"group_id":        strconv.Itoa(int(aws.ToInt64(apiObject.GroupId))),
		names.AttrOwnerID: strconv.Itoa(int(aws.ToInt64(apiObject.OwnerId))),
	}

	return []interface{}{tfMap}
}

func expandNFSFileShareCacheAttributes(tfList []interface{}) *awstypes.CacheAttributes {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &awstypes.CacheAttributes{
		CacheStaleTimeoutInSeconds: aws.Int32(int32(tfMap["cache_stale_timeout_in_seconds"].(int))),
	}

	return apiObject
}

func flattenNFSFileShareCacheAttributes(apiObject *awstypes.CacheAttributes) []interface{} {
	if apiObject == nil {
		return []interface{}{}
	}

	tfMap := map[string]interface{}{
		"cache_stale_timeout_in_seconds": aws.ToInt32(apiObject.CacheStaleTimeoutInSeconds),
	}

	return []interface{}{tfMap}
}
