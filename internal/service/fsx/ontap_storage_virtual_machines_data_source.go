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
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKDataSource("aws_fsx_ontap_storage_virtual_machines", name="Ontap Storage Virtual Machines")
func DataSourceOntapStorageVirtualMachines() *schema.Resource {
	return &schema.Resource{
		ReadWithoutTimeout: dataSourceOntapStorageVirtualMachinesRead,

		Schema: map[string]*schema.Schema{
			"filter": DataSourceStorageVirtualMachineFiltersSchema(),
			"ids": {
				Type:     schema.TypeList,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

const (
	DSNameOntapStorageVirtualMachines = "Ontap Storage Virtual Machines Data Source"
)

func dataSourceOntapStorageVirtualMachinesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).FSxConn(ctx)

	input := &fsx.DescribeStorageVirtualMachinesInput{}

	input.Filters = BuildStorageVirtualMachineFiltersDataSource(
		d.Get("filter").(*schema.Set),
	)

	if len(input.Filters) == 0 {
		input.Filters = nil
	}

	output, err := FindStorageVirtualMachines(ctx, conn, input)

	if err != nil {
		return sdkdiag.AppendFromErr(diags, tfresource.SingularDataSourceFindError("FSx StorageVirtualMachines", err))
	}

	var svmIDs []string

	for _, v := range output {
		svmIDs = append(svmIDs, aws.StringValue(v.StorageVirtualMachineId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	d.Set("ids", svmIDs)

	return diags
}
