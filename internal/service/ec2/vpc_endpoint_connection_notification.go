// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_vpc_endpoint_connection_notification", name="VPC Endpoint Connection Notification")
func resourceVPCEndpointConnectionNotification() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCEndpointConnectionNotificationCreate,
		ReadWithoutTimeout:   resourceVPCEndpointConnectionNotificationRead,
		UpdateWithoutTimeout: resourceVPCEndpointConnectionNotificationUpdate,
		DeleteWithoutTimeout: resourceVPCEndpointConnectionNotificationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"connection_events": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"connection_notification_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"notification_type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVPCEndpointID: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrVPCEndpointID, "vpc_endpoint_service_id"},
			},
			"vpc_endpoint_service_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{names.AttrVPCEndpointID, "vpc_endpoint_service_id"},
			},
		},
	}
}

func resourceVPCEndpointConnectionNotificationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateVpcEndpointConnectionNotificationInput{
		ClientToken:               aws.String(id.UniqueId()),
		ConnectionEvents:          flex.ExpandStringValueSet(d.Get("connection_events").(*schema.Set)),
		ConnectionNotificationArn: aws.String(d.Get("connection_notification_arn").(string)),
	}

	if v, ok := d.GetOk("vpc_endpoint_service_id"); ok {
		input.ServiceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrVPCEndpointID); ok {
		input.VpcEndpointId = aws.String(v.(string))
	}

	output, err := conn.CreateVpcEndpointConnectionNotification(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 VPC Endpoint Connection Notification: %s", err)
	}

	d.SetId(aws.ToString(output.ConnectionNotification.ConnectionNotificationId))

	return append(diags, resourceVPCEndpointConnectionNotificationRead(ctx, d, meta)...)
}

func resourceVPCEndpointConnectionNotificationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	cn, err := findVPCEndpointConnectionNotificationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 VPC Endpoint Connection Notification %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 VPC Endpoint Connection Notification (%s): %s", d.Id(), err)
	}

	d.Set("connection_events", cn.ConnectionEvents)
	d.Set("connection_notification_arn", cn.ConnectionNotificationArn)
	d.Set("notification_type", cn.ConnectionNotificationType)
	d.Set(names.AttrState, cn.ConnectionNotificationState)
	d.Set(names.AttrVPCEndpointID, cn.VpcEndpointId)
	d.Set("vpc_endpoint_service_id", cn.ServiceId)

	return diags
}

func resourceVPCEndpointConnectionNotificationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.ModifyVpcEndpointConnectionNotificationInput{
		ConnectionNotificationId: aws.String(d.Id()),
	}

	if d.HasChange("connection_events") {
		input.ConnectionEvents = flex.ExpandStringValueSet(d.Get("connection_events").(*schema.Set))
	}

	if d.HasChange("connection_notification_arn") {
		input.ConnectionNotificationArn = aws.String(d.Get("connection_notification_arn").(string))
	}

	_, err := conn.ModifyVpcEndpointConnectionNotification(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EC2 VPC Endpoint Connection Notification (%s): %s", d.Id(), err)
	}

	return append(diags, resourceVPCEndpointConnectionNotificationRead(ctx, d, meta)...)
}

func resourceVPCEndpointConnectionNotificationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	log.Printf("[DEBUG] Deleting EC2 VPC Endpoint Connection Notification: %s", d.Id())
	_, err := conn.DeleteVpcEndpointConnectionNotifications(ctx, &ec2.DeleteVpcEndpointConnectionNotificationsInput{
		ConnectionNotificationIds: []string{d.Id()},
	})

	if tfawserr.ErrCodeEquals(err, errCodeInvalidConnectionNotification) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 VPC Endpoint Connection Notification (%s): %s", d.Id(), err)
	}

	return diags
}
