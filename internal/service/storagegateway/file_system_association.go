// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"errors"
	"log"
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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_storagegateway_file_system_association", name="File System Association")
// @Tags(identifierAttribute="arn")
func resourceFileSystemAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFileSystemAssociationCreate,
		ReadWithoutTimeout:   resourceFileSystemAssociationRead,
		UpdateWithoutTimeout: resourceFileSystemAssociationUpdate,
		DeleteWithoutTimeout: resourceFileSystemAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
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
				Default:      "",
			},
			"cache_attributes": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cache_stale_timeout_in_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
							ValidateFunc: validation.Any(
								validation.IntInSlice([]int{0}),
								validation.IntBetween(300, 2592000),
							),
						},
					},
				},
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"location_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrPassword: {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^[ -~]+$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrUsername: {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringMatch(regexache.MustCompile(`^\w[\w\.\- ]*$`), ""),
					validation.StringLenBetween(1, 1024),
				),
			},
		},
	}
}

func resourceFileSystemAssociationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	gatewayARN := d.Get("gateway_arn").(string)
	input := &storagegateway.AssociateFileSystemInput{
		ClientToken: aws.String(id.UniqueId()),
		GatewayARN:  aws.String(gatewayARN),
		LocationARN: aws.String(d.Get("location_arn").(string)),
		Password:    aws.String(d.Get(names.AttrPassword).(string)),
		Tags:        getTagsIn(ctx),
		UserName:    aws.String(d.Get(names.AttrUsername).(string)),
	}

	if v, ok := d.GetOk("audit_destination_arn"); ok {
		input.AuditDestinationARN = aws.String(v.(string))
	}

	if v, ok := d.GetOk("cache_attributes"); ok {
		input.CacheAttributes = expandFileSystemAssociationCacheAttributes(v.([]any))
	}

	output, err := conn.AssociateFileSystem(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway (%s) File System Association: %s", gatewayARN, err)
	}

	d.SetId(aws.ToString(output.FileSystemAssociationARN))

	if _, err = waitFileSystemAssociationAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway File System Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceFileSystemAssociationRead(ctx, d, meta)...)
}

func resourceFileSystemAssociationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	filesystem, err := findFileSystemAssociationByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway File System Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway File System Association (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, filesystem.FileSystemAssociationARN)
	d.Set("audit_destination_arn", filesystem.AuditDestinationARN)
	if err := d.Set("cache_attributes", flattenFileSystemAssociationCacheAttributes(filesystem.CacheAttributes)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting cache_attributes: %s", err)
	}
	d.Set("gateway_arn", filesystem.GatewayARN)
	d.Set("location_arn", filesystem.LocationARN)

	setTagsOut(ctx, filesystem.Tags)

	return diags
}

func resourceFileSystemAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	if d.HasChangesExcept(names.AttrTagsAll) {
		input := &storagegateway.UpdateFileSystemAssociationInput{
			AuditDestinationARN:      aws.String(d.Get("audit_destination_arn").(string)),
			FileSystemAssociationARN: aws.String(d.Id()),
			Password:                 aws.String(d.Get(names.AttrPassword).(string)),
			UserName:                 aws.String(d.Get(names.AttrUsername).(string)),
		}

		if v, ok := d.GetOk("cache_attributes"); ok {
			input.CacheAttributes = expandFileSystemAssociationCacheAttributes(v.([]any))
		}

		_, err := conn.UpdateFileSystemAssociation(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Storage Gateway File System Association (%s): %s", d.Id(), err)
		}

		if _, err = waitFileSystemAssociationAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway File System Association (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceFileSystemAssociationRead(ctx, d, meta)...)
}

func resourceFileSystemAssociationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	log.Printf("[DEBUG] Deleting Storage Gateway File System Association: %s", d.Id())
	_, err := conn.DisassociateFileSystem(ctx, &storagegateway.DisassociateFileSystemInput{
		FileSystemAssociationARN: aws.String(d.Id()),
	})

	if operationErrorCode(err) == operationErrCodeFileSystemAssociationNotFound {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Storage Gateway File System Association (%s): %s", d.Id(), err)
	}

	if _, err = waitFileSystemAssociationDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Storage Gateway File System Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findFileSystemAssociationByARN(ctx context.Context, conn *storagegateway.Client, arn string) (*awstypes.FileSystemAssociationInfo, error) {
	input := &storagegateway.DescribeFileSystemAssociationsInput{
		FileSystemAssociationARNList: []string{arn},
	}

	return findFileSystemAssociation(ctx, conn, input)
}

func findFileSystemAssociation(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeFileSystemAssociationsInput) (*awstypes.FileSystemAssociationInfo, error) {
	output, err := findFileSystemAssociations(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findFileSystemAssociations(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeFileSystemAssociationsInput) ([]awstypes.FileSystemAssociationInfo, error) {
	output, err := conn.DescribeFileSystemAssociations(ctx, input)

	if operationErrorCode(err) == operationErrCodeFileSystemAssociationNotFound {
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

	return output.FileSystemAssociationInfoList, nil
}

func statusFileSystemAssociation(ctx context.Context, conn *storagegateway.Client, arn string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findFileSystemAssociationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.FileSystemAssociationStatus), nil
	}
}

func waitFileSystemAssociationAvailable(ctx context.Context, conn *storagegateway.Client, fileSystemArn string, timeout time.Duration) (*awstypes.FileSystemAssociationInfo, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{fileSystemAssociationStatusCreating, fileSystemAssociationStatusUpdating},
		Target:  []string{fileSystemAssociationStatusAvailable},
		Refresh: statusFileSystemAssociation(ctx, conn, fileSystemArn),
		Timeout: timeout,
		Delay:   5 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileSystemAssociationInfo); ok {
		tfresource.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FileSystemAssociationStatusDetails, fileSystemAssociationStatusDetailError)...))

		return output, err
	}

	return nil, err
}

func waitFileSystemAssociationDeleted(ctx context.Context, conn *storagegateway.Client, fileSystemArn string, timeout time.Duration) (*awstypes.FileSystemAssociationInfo, error) {
	stateConf := &retry.StateChangeConf{
		Pending:        []string{fileSystemAssociationStatusAvailable, fileSystemAssociationStatusDeleting, fileSystemAssociationStatusForceDeleting},
		Target:         []string{},
		Refresh:        statusFileSystemAssociation(ctx, conn, fileSystemArn),
		Timeout:        timeout,
		Delay:          5 * time.Second,
		NotFoundChecks: 1,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.FileSystemAssociationInfo); ok {
		tfresource.SetLastError(err, errors.Join(tfslices.ApplyToAll(output.FileSystemAssociationStatusDetails, fileSystemAssociationStatusDetailError)...))

		return output, err
	}

	return nil, err
}

func fileSystemAssociationStatusDetailError(v awstypes.FileSystemAssociationStatusDetail) error {
	return errors.New(aws.ToString(v.ErrorCode))
}

func expandFileSystemAssociationCacheAttributes(tfList []any) *awstypes.CacheAttributes {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &awstypes.CacheAttributes{
		CacheStaleTimeoutInSeconds: aws.Int32(int32(tfMap["cache_stale_timeout_in_seconds"].(int))),
	}

	return apiObject
}

func flattenFileSystemAssociationCacheAttributes(apiObject *awstypes.CacheAttributes) []any {
	if apiObject == nil {
		return []any{}
	}

	tfMap := map[string]any{
		"cache_stale_timeout_in_seconds": aws.ToInt32(apiObject.CacheStaleTimeoutInSeconds),
	}

	return []any{tfMap}
}
