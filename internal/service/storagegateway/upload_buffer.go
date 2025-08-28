// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_storagegateway_upload_buffer", name="Upload Buffer")
func resourceUploadBuffer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceUploadBufferCreate,
		ReadWithoutTimeout:   resourceUploadBufferRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"disk_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"disk_id", "disk_path"},
			},
			"disk_path": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"disk_id", "disk_path"},
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceUploadBufferCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	diskID := d.Get("disk_id").(string)
	gatewayARN := d.Get("gateway_arn").(string)
	input := &storagegateway.AddUploadBufferInput{
		GatewayARN: aws.String(gatewayARN),
	}

	if diskID != "" {
		input.DiskIds = []string{diskID}
	}

	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17809.
	var diskPath string
	if v, ok := d.GetOk("disk_path"); ok {
		diskPath = v.(string)
		input.DiskIds = []string{diskPath}
	}

	output, err := conn.AddUploadBuffer(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway Upload Buffer: %s", err)
	}

	gatewayARN = aws.ToString(output.GatewayARN)

	if diskID != "" {
		d.SetId(uploadBufferCreateResourceID(gatewayARN, diskID))

		return append(diags, resourceUploadBufferRead(ctx, d, meta)...)
	}

	disk, err := findLocalDiskByGatewayARNAndDiskPath(ctx, conn, aws.ToString(output.GatewayARN), diskPath)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Local Disk (%s): %s", diskPath, err)
	}

	diskID = aws.ToString(disk.DiskId)
	d.SetId(uploadBufferCreateResourceID(gatewayARN, diskID))

	return append(diags, resourceUploadBufferRead(ctx, d, meta)...)
}

func resourceUploadBufferRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	gatewayARN, diskID, err := uploadBufferParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	foundDiskID, err := findUploadBufferDiskIDByTwoPartKey(ctx, conn, gatewayARN, diskID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway Upload Buffer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Upload Buffer (%s): %s", d.Id(), err)
	}

	d.Set("disk_id", foundDiskID)
	d.Set("gateway_arn", gatewayARN)

	if _, ok := d.GetOk("disk_path"); !ok {
		diskID := aws.ToString(foundDiskID)
		disk, err := findLocalDiskByGatewayARNAndDiskID(ctx, conn, gatewayARN, diskID)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Local Disk (%s): %s", diskID, err)
		}

		d.Set("disk_path", disk.DiskPath)
	}

	return diags
}

const uploadBufferResourceIDSeparator = ":"

func uploadBufferCreateResourceID(gatewayARN, diskID string) string {
	parts := []string{gatewayARN, diskID}
	id := strings.Join(parts, uploadBufferResourceIDSeparator)

	return id
}

func uploadBufferParseResourceID(id string) (string, string, error) {
	// id = arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	idFormatErr := fmt.Errorf("unexpected format for ID (%[1]s), expected GatewayARN%[2]sDiskID", id, uploadBufferResourceIDSeparator)
	gatewayARNAndDisk, err := arn.Parse(id)
	if err != nil {
		return "", "", idFormatErr
	}
	// gatewayARNAndDisk.Resource = gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	resourceParts := strings.SplitN(gatewayARNAndDisk.Resource, uploadBufferResourceIDSeparator, 2)
	if len(resourceParts) != 2 {
		return "", "", idFormatErr
	}
	// resourceParts = ["gateway/sgw-12345678", "pci-0000:03:00.0-scsi-0:0:0:0"]
	gatewayARN := &arn.ARN{
		AccountID: gatewayARNAndDisk.AccountID,
		Partition: gatewayARNAndDisk.Partition,
		Region:    gatewayARNAndDisk.Region,
		Service:   gatewayARNAndDisk.Service,
		Resource:  resourceParts[0],
	}
	return gatewayARN.String(), resourceParts[1], nil
}

func findLocalDiskByGatewayARNAndDiskID(ctx context.Context, conn *storagegateway.Client, gatewayARN, diskID string) (*awstypes.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	return findLocalDisk(ctx, conn, input, func(v awstypes.Disk) bool {
		return aws.ToString(v.DiskId) == diskID
	})
}

func findLocalDiskByGatewayARNAndDiskPath(ctx context.Context, conn *storagegateway.Client, gatewayARN, diskPath string) (*awstypes.Disk, error) {
	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}

	return findLocalDisk(ctx, conn, input, func(v awstypes.Disk) bool {
		return aws.ToString(v.DiskPath) == diskPath
	})
}

func findLocalDisk(ctx context.Context, conn *storagegateway.Client, input *storagegateway.ListLocalDisksInput, filter tfslices.Predicate[awstypes.Disk]) (*awstypes.Disk, error) {
	output, err := findLocalDisks(ctx, conn, input, filter)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(output)
}

func findLocalDisks(ctx context.Context, conn *storagegateway.Client, input *storagegateway.ListLocalDisksInput, filter tfslices.Predicate[awstypes.Disk]) ([]awstypes.Disk, error) {
	output, err := conn.ListLocalDisks(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return tfslices.Filter(output.Disks, filter), nil
}

func findUploadBufferDiskIDByTwoPartKey(ctx context.Context, conn *storagegateway.Client, gatewayARN string, diskID string) (*string, error) {
	input := &storagegateway.DescribeUploadBufferInput{
		GatewayARN: aws.String(gatewayARN),
	}
	output, err := findUploadBuffer(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.DiskIds, func(v string) bool {
		return v == diskID
	}))
}

func findUploadBuffer(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeUploadBufferInput) (*storagegateway.DescribeUploadBufferOutput, error) {
	output, err := conn.DescribeUploadBuffer(ctx, input)

	if isGatewayNotFoundErr(err) {
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

	return output, nil
}
