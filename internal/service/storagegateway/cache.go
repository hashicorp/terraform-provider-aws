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

// @SDKResource("aws_storagegateway_cache", name="Cache")
func resourceCache() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCacheCreate,
		ReadWithoutTimeout:   resourceCacheRead,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"disk_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

func resourceCacheCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	diskID := d.Get("disk_id").(string)
	gatewayARN := d.Get("gateway_arn").(string)
	id := cacheCreateResourceID(gatewayARN, diskID)
	inputAC := &storagegateway.AddCacheInput{
		DiskIds:    []string{diskID},
		GatewayARN: aws.String(gatewayARN),
	}

	_, err := conn.AddCache(ctx, inputAC)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway Cache (%s): %s", id, err)
	}

	d.SetId(id)

	// Depending on the Storage Gateway software, it will sometimes relabel a local DiskId
	// with a UUID if previously unlabeled, e.g.
	//   Before conn.AddCache(): "DiskId": "/dev/xvdb",
	//   After conn.AddCache(): "DiskId": "112764d7-7e83-42ce-9af3-d482985a31cc",
	// This prevents us from successfully reading the disk after creation.
	// Here we try to refresh the local disks to see if we can find a new DiskId.
	inputLLD := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(gatewayARN),
	}
	disk, err := findLocalDisk(ctx, conn, inputLLD, func(v awstypes.Disk) bool {
		return aws.ToString(v.DiskId) == diskID || aws.ToString(v.DiskNode) == diskID || aws.ToString(v.DiskPath) == diskID
	})

	switch {
	case tfresource.NotFound(err):
	case err != nil:
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway local disk: %s", err)
	default:
		id = cacheCreateResourceID(gatewayARN, aws.ToString(disk.DiskId))
		d.SetId(id)
	}

	return append(diags, resourceCacheRead(ctx, d, meta)...)
}

func resourceCacheRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	gatewayARN, diskID, err := cacheParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	err = findCacheByTwoPartKey(ctx, conn, gatewayARN, diskID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway Cache (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Cache (%s): %s", d.Id(), err)
	}

	d.Set("disk_id", diskID)
	d.Set("gateway_arn", gatewayARN)

	return diags
}

const cacheResourceIDSeparator = ":"

func cacheCreateResourceID(gatewayARN, diskID string) string {
	parts := []string{gatewayARN, diskID}
	id := strings.Join(parts, cacheResourceIDSeparator)

	return id
}

func cacheParseResourceID(id string) (string, string, error) {
	// id = arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	idFormatErr := fmt.Errorf("unexpected format for ID (%[1]s), expected GatewayARN%[2]sDiskID", id, cacheResourceIDSeparator)
	gatewayARNAndDisk, err := arn.Parse(id)
	if err != nil {
		return "", "", idFormatErr
	}
	// gatewayARNAndDisk.Resource = gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	resourceParts := strings.SplitN(gatewayARNAndDisk.Resource, cacheResourceIDSeparator, 2)
	if len(resourceParts) != 2 {
		return "", "", idFormatErr
	}
	// resourceParts = ["gateway/sgw-12345678", "pci-0000:03:00.0-scsi-0:0:0:0"]
	gatewayARN := &arn.ARN{
		Partition: gatewayARNAndDisk.Partition,
		Service:   gatewayARNAndDisk.Service,
		Region:    gatewayARNAndDisk.Region,
		AccountID: gatewayARNAndDisk.AccountID,
		Resource:  resourceParts[0],
	}
	return gatewayARN.String(), resourceParts[1], nil
}

func findCacheByTwoPartKey(ctx context.Context, conn *storagegateway.Client, gatewayARN, diskID string) error {
	input := &storagegateway.DescribeCacheInput{
		GatewayARN: aws.String(gatewayARN),
	}
	output, err := findCache(ctx, conn, input)

	if err != nil {
		return err
	}

	_, err = tfresource.AssertSingleValueResult(tfslices.Filter(output.DiskIds, func(v string) bool {
		return v == diskID
	}))

	return err
}

func findCache(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeCacheInput) (*storagegateway.DescribeCacheOutput, error) {
	output, err := conn.DescribeCache(ctx, input)

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

	return output, err
}
