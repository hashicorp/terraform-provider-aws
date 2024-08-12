// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_internet_gateway", name="Internet Gateway")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceInternetGateway() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceInternetGatewayCreate,
		ReadWithoutTimeout:   resourceInternetGatewayRead,
		UpdateWithoutTimeout: resourceInternetGatewayUpdate,
		DeleteWithoutTimeout: resourceInternetGatewayDelete,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(20 * time.Minute),
			Update: schema.DefaultTimeout(20 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrOwnerID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceInternetGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	input := &ec2.CreateInternetGatewayInput{
		TagSpecifications: getTagSpecificationsIn(ctx, awstypes.ResourceTypeInternetGateway),
	}

	log.Printf("[DEBUG] Creating EC2 Internet Gateway: %#v", input)
	output, err := conn.CreateInternetGateway(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating EC2 Internet Gateway: %s", err)
	}

	d.SetId(aws.ToString(output.InternetGateway.InternetGatewayId))

	if v, ok := d.GetOk(names.AttrVPCID); ok {
		if err := attachInternetGateway(ctx, conn, d.Id(), v.(string), d.Timeout(schema.TimeoutCreate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "creating EC2 Internet Gateway: %s", err)
		}
	}

	return append(diags, resourceInternetGatewayRead(ctx, d, meta)...)
}

func resourceInternetGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	outputRaw, err := tfresource.RetryWhenNewResourceNotFound(ctx, ec2PropagationTimeout, func() (interface{}, error) {
		return findInternetGatewayByID(ctx, conn, d.Id())
	}, d.IsNewResource())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] EC2 Internet Gateway %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Internet Gateway (%s): %s", d.Id(), err)
	}

	ig := outputRaw.(*awstypes.InternetGateway)

	ownerID := aws.ToString(ig.OwnerId)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   names.EC2,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: ownerID,
		Resource:  fmt.Sprintf("internet-gateway/%s", d.Id()),
	}.String()
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrOwnerID, ownerID)
	if len(ig.Attachments) == 0 {
		// Gateway exists but not attached to the VPC.
		d.Set(names.AttrVPCID, "")
	} else {
		d.Set(names.AttrVPCID, ig.Attachments[0].VpcId)
	}

	setTagsOut(ctx, ig.Tags)

	return diags
}

func resourceInternetGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	if d.HasChange(names.AttrVPCID) {
		o, n := d.GetChange(names.AttrVPCID)

		if v := o.(string); v != "" {
			if err := detachInternetGateway(ctx, conn, d.Id(), v, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Internet Gateway (%s): %s", d.Id(), err)
			}
		}

		if v := n.(string); v != "" {
			if err := attachInternetGateway(ctx, conn, d.Id(), v, d.Timeout(schema.TimeoutUpdate)); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating EC2 Internet Gateway (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceInternetGatewayRead(ctx, d, meta)...)
}

func resourceInternetGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// Detach if it is attached.
	if v, ok := d.GetOk(names.AttrVPCID); ok {
		if err := detachInternetGateway(ctx, conn, d.Id(), v.(string), d.Timeout(schema.TimeoutDelete)); err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting EC2 Internet Gateway (%s): %s", d.Id(), err)
		}
	}

	input := &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(d.Id()),
	}

	log.Printf("[INFO] Deleting Internet Gateway: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, d.Timeout(schema.TimeoutDelete), func() (interface{}, error) {
		return conn.DeleteInternetGateway(ctx, input)
	}, errCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, errCodeInvalidInternetGatewayIDNotFound) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting EC2 Internet Gateway (%s): %s", d.Id(), err)
	}

	return diags
}

func attachInternetGateway(ctx context.Context, conn *ec2.Client, internetGatewayID, vpcID string, timeout time.Duration) error {
	input := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(internetGatewayID),
		VpcId:             aws.String(vpcID),
	}

	log.Printf("[INFO] Attaching EC2 Internet Gateway: %#v", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.AttachInternetGateway(ctx, input)
	}, errCodeInvalidInternetGatewayIDNotFound)

	if err != nil {
		return fmt.Errorf("attaching EC2 Internet Gateway (%s) to VPC (%s): %w", internetGatewayID, vpcID, err)
	}

	_, err = waitInternetGatewayAttached(ctx, conn, internetGatewayID, vpcID, timeout)

	if err != nil {
		return fmt.Errorf("waiting for EC2 Internet Gateway (%s) to attach to VPC (%s): %w", internetGatewayID, vpcID, err)
	}

	return nil
}

func detachInternetGateway(ctx context.Context, conn *ec2.Client, internetGatewayID, vpcID string, timeout time.Duration) error {
	input := &ec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(internetGatewayID),
		VpcId:             aws.String(vpcID),
	}

	log.Printf("[INFO] Detaching EC2 Internet Gateway: %#v", input)
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, timeout, func() (interface{}, error) {
		return conn.DetachInternetGateway(ctx, input)
	}, errCodeDependencyViolation)

	if tfawserr.ErrCodeEquals(err, errCodeGatewayNotAttached) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("detaching EC2 Internet Gateway (%s) from VPC (%s): %w", internetGatewayID, vpcID, err)
	}

	_, err = waitInternetGatewayDetached(ctx, conn, internetGatewayID, vpcID, timeout)

	if err != nil {
		return fmt.Errorf("waiting for EC2 Internet Gateway (%s) to detach from VPC (%s): %w", internetGatewayID, vpcID, err)
	}

	return nil
}
