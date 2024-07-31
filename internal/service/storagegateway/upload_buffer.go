// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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

func resourceUploadBufferCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	input := &storagegateway.AddUploadBufferInput{}

	if v, ok := d.GetOk("disk_id"); ok {
		input.DiskIds = aws.StringSlice([]string{v.(string)})
	}

	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/17809
	if v, ok := d.GetOk("disk_path"); ok {
		input.DiskIds = aws.StringSlice([]string{v.(string)})
	}

	if v, ok := d.GetOk("gateway_arn"); ok {
		input.GatewayARN = aws.String(v.(string))
	}

	output, err := conn.AddUploadBufferWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding Storage Gateway upload buffer: %s", err)
	}

	if output == nil {
		return sdkdiag.AppendErrorf(diags, "adding Storage Gateway upload buffer: empty response")
	}

	if v, ok := d.GetOk("disk_id"); ok {
		d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(output.GatewayARN), v.(string)))

		return append(diags, resourceUploadBufferRead(ctx, d, meta)...)
	}

	disk, err := findLocalDiskByGatewayARNAndDiskPath(ctx, conn, aws.StringValue(output.GatewayARN), aws.StringValue(input.DiskIds[0]))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "listing Storage Gateway Local Disks after creating Upload Buffer: %s", err)
	}

	if disk == nil {
		return sdkdiag.AppendErrorf(diags, "listing Storage Gateway Local Disks after creating Upload Buffer: disk not found")
	}

	d.SetId(fmt.Sprintf("%s:%s", aws.StringValue(output.GatewayARN), aws.StringValue(disk.DiskId)))

	return append(diags, resourceUploadBufferRead(ctx, d, meta)...)
}

func resourceUploadBufferRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	gatewayARN, diskID, err := DecodeUploadBufferID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Upload Buffer (%s): %s", d.Id(), err)
	}

	foundDiskID, err := FindUploadBufferDisk(ctx, conn, gatewayARN, diskID)

	if !d.IsNewResource() && IsErrGatewayNotFound(err) {
		log.Printf("[WARN] Storage Gateway Upload Buffer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Upload Buffer (%s): %s", d.Id(), err)
	}

	if foundDiskID == nil {
		if d.IsNewResource() {
			return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Upload Buffer (%s): not found", d.Id())
		}

		log.Printf("[WARN] Storage Gateway Upload Buffer (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("disk_id", foundDiskID)
	d.Set("gateway_arn", gatewayARN)

	if _, ok := d.GetOk("disk_path"); !ok {
		disk, err := findLocalDiskByGatewayARNAndDiskID(ctx, conn, gatewayARN, aws.StringValue(foundDiskID))

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "listing Storage Gateway Local Disks: %s", err)
		}

		if disk == nil {
			return sdkdiag.AppendErrorf(diags, "listing Storage Gateway Local Disks: disk not found")
		}

		d.Set("disk_path", disk.DiskPath)
	}

	return diags
}

func DecodeUploadBufferID(id string) (string, string, error) {
	// id = arn:aws:storagegateway:us-east-1:123456789012:gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	idFormatErr := fmt.Errorf("expected ID in form of GatewayARN:DiskId, received: %s", id)
	gatewayARNAndDisk, err := arn.Parse(id)
	if err != nil {
		return "", "", idFormatErr
	}
	// gatewayARNAndDisk.Resource = gateway/sgw-12345678:pci-0000:03:00.0-scsi-0:0:0:0
	resourceParts := strings.SplitN(gatewayARNAndDisk.Resource, ":", 2)
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
