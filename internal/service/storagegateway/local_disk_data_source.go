// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/storagegateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKDataSource("aws_storagegateway_local_disk", name="Local Disk")
func dataSourceLocalDisk() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceLocalDiskRead,

		Schema: map[string]*schema.Schema{
			"disk_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disk_node": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"disk_path": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"gateway_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func dataSourceLocalDiskRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).StorageGatewayClient(ctx)

	input := &storagegateway.ListLocalDisksInput{
		GatewayARN: aws.String(d.Get("gateway_arn").(string)),
	}

	disk, err := findLocalDisk(ctx, conn, input, func(disk awstypes.Disk) bool {
		if v, ok := d.GetOk("disk_node"); ok && v.(string) == aws.ToString(disk.DiskNode) {
			return true
		}
		if v, ok := d.GetOk("disk_path"); ok && v.(string) == aws.ToString(disk.DiskPath) {
			return true
		}
		return false
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("Storage Gateway Local Disk", err))
	}

	d.SetId(aws.ToString(disk.DiskId))
	d.Set("disk_id", disk.DiskId)
	d.Set("disk_node", disk.DiskNode)
	d.Set("disk_path", disk.DiskPath)

	return diags
}
