// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fsx"
	awstypes "github.com/aws/aws-sdk-go-v2/service/fsx/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKDataSource("aws_fsx_ontap_storage_virtual_machines", name="ONTAP Storage Virtual Machines")
func dataSourceONTAPStorageVirtualMachines() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceONTAPStorageVirtualMachinesRead,

		Schema: map[string]*schema.Schema{
			names.AttrFilter: storageVirtualMachineFiltersSchema(),
			names.AttrIDs: {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceONTAPStorageVirtualMachinesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxClient(ctx)

	input := &fsx.DescribeStorageVirtualMachinesInput{}

	input.Filters = newStorageVirtualMachineFilterList(
		d.Get(names.AttrFilter).(*schema.Set),
	)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	svms, err := findStorageVirtualMachines(ctx, conn, input, tfslices.PredicateTrue[*awstypes.StorageVirtualMachine]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx ONTAP Storage Virtual Machines: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set(names.AttrIDs, tfslices.ApplyToAll(svms, func(svm awstypes.StorageVirtualMachine) string {
		return aws.ToString(svm.StorageVirtualMachineId)
	}))

	return diags
}
