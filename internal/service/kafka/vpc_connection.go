// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_msk_vpc_connection", name="VPC Connection")
// @Tags(identifierAttribute="id")
func resourceVPCConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCConnectionCreate,
		ReadWithoutTimeout:   resourceVPCConnectionRead,
		UpdateWithoutTimeout: resourceVPCConnectionUpdate,
		DeleteWithoutTimeout: resourceVPCConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"authentication": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"client_subnets": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"target_cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrVPCID: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	input := &kafka.CreateVpcConnectionInput{
		Authentication:   aws.String(d.Get("authentication").(string)),
		ClientSubnets:    flex.ExpandStringValueSet(d.Get("client_subnets").(*schema.Set)),
		SecurityGroups:   flex.ExpandStringValueSet(d.Get(names.AttrSecurityGroups).(*schema.Set)),
		Tags:             getTagsIn(ctx),
		TargetClusterArn: aws.String(d.Get("target_cluster_arn").(string)),
		VpcId:            aws.String(d.Get(names.AttrVPCID).(string)),
	}

	output, err := conn.CreateVpcConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK VPC Connection: %s", err)
	}

	d.SetId(aws.ToString(output.VpcConnectionArn))

	if _, err := waitVPCConnectionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK VPC Connection (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCConnectionRead(ctx, d, meta)...)
}

func resourceVPCConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	output, err := findVPCConnectionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK VPC Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK VPC Connection (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.VpcConnectionArn)
	d.Set("authentication", output.Authentication)
	d.Set("client_subnets", flex.FlattenStringValueSet(output.Subnets))
	d.Set(names.AttrSecurityGroups, flex.FlattenStringValueSet(output.SecurityGroups))
	d.Set("target_cluster_arn", output.TargetClusterArn)
	d.Set(names.AttrVPCID, output.VpcId)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceVPCConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Tags only.

	return append(diags, resourceVPCConnectionRead(ctx, d, meta)...)
}

func resourceVPCConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	log.Printf("[INFO] Deleting MSK VPC Connection: %s", d.Id())
	_, err := conn.DeleteVpcConnection(ctx, &kafka.DeleteVpcConnectionInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting MSK VPC Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCConnectionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK VPC Connection (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func waitVPCConnectionCreated(ctx context.Context, conn *kafka.Client, id string, timeout time.Duration) (*kafka.DescribeVpcConnectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.VpcConnectionStateCreating),
		Target:                    enum.Slice(types.VpcConnectionStateAvailable),
		Refresh:                   statusVPCConnection(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeVpcConnectionOutput); ok {
		return output, err
	}

	return nil, err
}

func waitVPCConnectionDeleted(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*kafka.DescribeVpcConnectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VpcConnectionStateAvailable, types.VpcConnectionStateInactive, types.VpcConnectionStateDeactivating, types.VpcConnectionStateDeleting),
		Target:  []string{},
		Refresh: statusVPCConnection(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if output, ok := outputRaw.(*kafka.DescribeVpcConnectionOutput); ok {
		return output, err
	}

	return nil, err
}

func statusVPCConnection(ctx context.Context, conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCConnectionByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func findVPCConnectionByARN(ctx context.Context, conn *kafka.Client, arn string) (*kafka.DescribeVpcConnectionOutput, error) {
	input := &kafka.DescribeVpcConnectionInput{
		Arn: aws.String(arn),
	}

	output, err := conn.DescribeVpcConnection(ctx, input)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
