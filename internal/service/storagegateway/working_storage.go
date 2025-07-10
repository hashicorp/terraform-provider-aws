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
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_storagegateway_working_storage", name="Working Storage")
func ResourceWorkingStorage() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkingStorageCreate,
		ReadWithoutTimeout:   resourceWorkingStorageRead,
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

func resourceWorkingStorageCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	diskID := d.Get("disk_id").(string)
	gatewayARN := d.Get("gateway_arn").(string)
	id := workingStorageCreateResourceID(gatewayARN, diskID)
	input := &storagegateway.AddWorkingStorageInput{
		DiskIds:    []string{diskID},
		GatewayARN: aws.String(gatewayARN),
	}

	_, err := conn.AddWorkingStorage(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Storage Gateway Working Storage (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceWorkingStorageRead(ctx, d, meta)...)
}

func resourceWorkingStorageRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	gatewayARN, diskID, err := workingStorageParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	_, err = findWorkingStorageDiskIDByTwoPartKey(ctx, conn, gatewayARN, diskID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Storage Gateway Working Storage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Working Storage (%s): %s", d.Id(), err)
	}

	d.Set("disk_id", diskID)
	d.Set("gateway_arn", gatewayARN)

	return diags
}

const workingStorageResourceIDSeparator = ":"

func workingStorageCreateResourceID(gatewayARN, diskID string) string {
	parts := []string{gatewayARN, diskID}
	id := strings.Join(parts, workingStorageResourceIDSeparator)

	return id
}

func workingStorageParseResourceID(id string) (string, string, error) {
	// id = arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	idFormatErr := fmt.Errorf("unexpected format for ID (%[1]s), expected GatewayARN%[2]sDiskID", id, workingStorageResourceIDSeparator)
	gatewayARNAndDisk, err := arn.Parse(id)
	if err != nil {
		return "", "", idFormatErr
	}
	// gatewayARNAndDisk.Resource = gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	resourceParts := strings.SplitN(gatewayARNAndDisk.Resource, workingStorageResourceIDSeparator, 2)
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

func findWorkingStorageDiskIDByTwoPartKey(ctx context.Context, conn *storagegateway.Client, gatewayARN string, diskID string) (*string, error) {
	input := &storagegateway.DescribeWorkingStorageInput{
		GatewayARN: aws.String(gatewayARN),
	}
	output, err := findWorkingStorage(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	return tfresource.AssertSingleValueResult(tfslices.Filter(output.DiskIds, func(v string) bool {
		return v == diskID
	}))
}

func findWorkingStorage(ctx context.Context, conn *storagegateway.Client, input *storagegateway.DescribeWorkingStorageInput) (*storagegateway.DescribeWorkingStorageOutput, error) {
	output, err := conn.DescribeWorkingStorage(ctx, input)

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
