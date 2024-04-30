// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// @SDKDataSource("aws_fsx_ontap_storage_virtual_machines", name="ONTAP Storage Virtual Machines")
func dataSourceONTAPStorageVirtualMachines() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceONTAPStorageVirtualMachinesRead,

		Schema: map[string]*schema.Schema{
			"filter": storageVirtualMachineFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceONTAPStorageVirtualMachinesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	input := &fsx.DescribeStorageVirtualMachinesInput{}

	input.Filters = newStorageVirtualMachineFilterList(
		d.Get("filter").(*schema.Set),
	)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	svms, err := findStorageVirtualMachines(ctx, conn, input, tfslices.PredicateTrue[*fsx.StorageVirtualMachine]())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading FSx ONTAP Storage Virtual Machines: %s", err)
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", tfslices.ApplyToAll(svms, func(svm *fsx.StorageVirtualMachine) string {
		return aws.StringValue(svm.StorageVirtualMachineId)
	}))

	return diags
}
