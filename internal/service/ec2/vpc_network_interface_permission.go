// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_network_interface_permission", name="Network Interface Permission")
func resourceNetworkInterfacePermission() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNetworkInterfacePermissionCreate,
		DeleteWithoutTimeout: resourceNetworkInterfacePermissionDelete,
		ReadWithoutTimeout:   resourceNetworkInterfacePermissionRead,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"permission": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.InterfacePermissionType](),
			},
		},
	}
}

func resourceNetworkInterfacePermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	accountID := d.Get(names.AttrAccountID).(string)
	networkInterfaceID := d.Get(names.AttrNetworkInterfaceID).(string)
	permission := d.Get("permission").(string)

	input := &ec2.CreateNetworkInterfacePermissionInput{
		AwsAccountId:       aws.String(accountID),
		NetworkInterfaceId: aws.String(networkInterfaceID),
		Permission:         awstypes.InterfacePermissionType(permission),
	}

	output, err := conn.CreateNetworkInterfacePermission(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Network Interface Permission: %s", err)
	}

	d.SetId(aws.ToString(output.InterfacePermission.NetworkInterfacePermissionId))

	return append(diags, resourceNetworkInterfacePermissionRead(ctx, d, meta)...)
}

func resourceNetworkInterfacePermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findNetworkInterfacePermissionByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Network Interface Permission (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network Interface Permission (%s): %s", d.Id(), err)
	}
	enip := outputRaw.(*types.NetworkInterfacePermission)

	d.Set(names.AttrNetworkInterfaceID, enip.NetworkInterfaceId)
	d.Set(names.AttrAccountID, enip.AwsAccountId)
	d.Set("permission", enip.Permission)

	return diags
}

func resourceNetworkInterfacePermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	tflog.Info(ctx, "Deleting EC2 Network Interface Permission", map[string]any{
		names.AttrNetworkInterfaceID: d.Id(),
	})

	_, err := conn.DeleteNetworkInterfacePermission(ctx, &ec2.DeleteNetworkInterfacePermissionInput{
		NetworkInterfacePermissionId: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}
