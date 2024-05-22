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

func resourceWorkingStorageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	diskID := d.Get("disk_id").(string)
	gatewayARN := d.Get("gateway_arn").(string)

	input := &storagegateway.AddWorkingStorageInput{
		DiskIds:    []*string{aws.String(diskID)},
		GatewayARN: aws.String(gatewayARN),
	}

	log.Printf("[DEBUG] Adding Storage Gateway Working Storage: %s", input)
	_, err := conn.AddWorkingStorageWithContext(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "adding Storage Gateway Working Storage: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", gatewayARN, diskID))

	return append(diags, resourceWorkingStorageRead(ctx, d, meta)...)
}

func resourceWorkingStorageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayConn(ctx)

	gatewayARN, diskID, err := DecodeWorkingStorageID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Working Storage (%s): %s", d.Id(), err)
	}

	input := &storagegateway.DescribeWorkingStorageInput{
		GatewayARN: aws.String(gatewayARN),
	}

	output, err := conn.DescribeWorkingStorageWithContext(ctx, input)
	if err != nil {
		if IsErrGatewayNotFound(err) {
			log.Printf("[WARN] Storage Gateway Working Storage (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Storage Gateway Working Storage (%s): %s", d.Id(), err)
	}

	if output == nil || len(output.DiskIds) == 0 {
		log.Printf("[WARN] Storage Gateway Working Storage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	found := false
	for _, existingDiskID := range output.DiskIds {
		if aws.StringValue(existingDiskID) == diskID {
			found = true
			break
		}
	}

	if !found {
		log.Printf("[WARN] Storage Gateway Working Storage (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	d.Set("disk_id", diskID)
	d.Set("gateway_arn", gatewayARN)

	return diags
}

func DecodeWorkingStorageID(id string) (string, string, error) {
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
