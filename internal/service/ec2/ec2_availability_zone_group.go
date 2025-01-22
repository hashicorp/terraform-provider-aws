// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_ec2_availability_zone_group", name="Availability Zone Group")
func resourceAvailabilityZoneGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAvailabilityZoneGroupCreate,
		ReadWithoutTimeout:   resourceAvailabilityZoneGroupRead,
		UpdateWithoutTimeout: resourceAvailabilityZoneGroupUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrGroupName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"opt_in_status": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice(enum.Slice(
					awstypes.AvailabilityZoneOptInStatusOptedIn,
					awstypes.AvailabilityZoneOptInStatusNotOptedIn,
				), false),
			},
		},
	}
}

func resourceAvailabilityZoneGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	groupName := d.Get(names.AttrGroupName).(string)
	availabilityZone, err := findAvailabilityZoneGroupByName(ctx, conn, groupName)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Availability Zone Group (%s): %s", groupName, err)
	}

	if v := d.Get("opt_in_status").(string); v != string(availabilityZone.OptInStatus) {
		if err := modifyAvailabilityZoneOptInStatus(ctx, conn, groupName, v); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Availability Zone Group (%s): %s", groupName, err)
		}
	}

	d.SetId(groupName)

	return append(diags, resourceAvailabilityZoneGroupRead(ctx, d, meta)...)
}

func resourceAvailabilityZoneGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	availabilityZone, err := findAvailabilityZoneGroupByName(ctx, conn, d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Availability Zone Group (%s): %s", d.Id(), err)
	}

	if availabilityZone.OptInStatus == awstypes.AvailabilityZoneOptInStatusOptInNotRequired {
		return sdkdiag.AppendErrorf(diags, "unnecessary handling of EC2 Availability Zone Group (%s), status: %s", d.Id(), awstypes.AvailabilityZoneOptInStatusOptInNotRequired)
	}

	d.Set(names.AttrGroupName, availabilityZone.GroupName)
	d.Set("opt_in_status", availabilityZone.OptInStatus)

	return diags
}

func resourceAvailabilityZoneGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if err := modifyAvailabilityZoneOptInStatus(ctx, conn, d.Id(), d.Get("opt_in_status").(string)); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EC2 Availability Zone Group (%s): %s", d.Id(), err)
	}

	return append(diags, resourceAvailabilityZoneGroupRead(ctx, d, meta)...)
}

func modifyAvailabilityZoneOptInStatus(ctx context.Context, conn *ec2.Client, groupName, optInStatus string) error {
	input := &ec2.ModifyAvailabilityZoneGroupInput{
		GroupName:   aws.String(groupName),
		OptInStatus: awstypes.ModifyAvailabilityZoneOptInStatus(optInStatus),
	}

	if _, err := conn.ModifyAvailabilityZoneGroup(ctx, input); err != nil {
		return err
	}

	waiter := waitAvailabilityZoneGroupOptedIn
	if optInStatus == string(awstypes.AvailabilityZoneOptInStatusNotOptedIn) {
		waiter = waitAvailabilityZoneGroupNotOptedIn
	}

	if _, err := waiter(ctx, conn, groupName); err != nil {
		return fmt.Errorf("waiting for completion: %w", err)
	}

	return nil
}
