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
func ResourceVPCConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCConnectionCreate,
		ReadWithoutTimeout:   resourceVPCConnectionRead,
		UpdateWithoutTimeout: resourceVPCConnectionUpdate,
		DeleteWithoutTimeout: resourceVPCConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
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
			"security_groups": {
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
			"vpc_id": {
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

	in := &kafka.CreateVpcConnectionInput{
		Authentication:   aws.String(d.Get("authentication").(string)),
		ClientSubnets:    flex.ExpandStringValueSet(d.Get("client_subnets").(*schema.Set)),
		SecurityGroups:   flex.ExpandStringValueSet(d.Get("security_groups").(*schema.Set)),
		Tags:             getTagsInV2(ctx),
		TargetClusterArn: aws.String(d.Get("target_cluster_arn").(string)),
		VpcId:            aws.String(d.Get("vpc_id").(string)),
	}

	out, err := conn.CreateVpcConnection(ctx, in)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating MSK VPC Connection: %s", err)
	}

	d.SetId(aws.ToString(out.VpcConnectionArn))

	if _, err := waitVPCConnectionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for MSK VPC Connection (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCConnectionRead(ctx, d, meta)...)
}

func resourceVPCConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	out, err := FindVPCConnectionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] MSK VPC Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading MSK VPC Connection (%s): %s", d.Id(), err)
	}

	d.Set("arn", out.VpcConnectionArn)
	d.Set("authentication", out.Authentication)
	d.Set("client_subnets", flex.FlattenStringValueSet(out.Subnets))
	d.Set("security_groups", flex.FlattenStringValueSet(out.SecurityGroups))
	d.Set("target_cluster_arn", out.TargetClusterArn)
	d.Set("vpc_id", out.VpcId)

	setTagsOutV2(ctx, out.Tags)

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
	if out, ok := outputRaw.(*kafka.DescribeVpcConnectionOutput); ok {
		return out, err
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
	if out, ok := outputRaw.(*kafka.DescribeVpcConnectionOutput); ok {
		return out, err
	}

	return nil, err
}

func statusVPCConnection(ctx context.Context, conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindVPCConnectionByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindVPCConnectionByARN(ctx context.Context, conn *kafka.Client, arn string) (*kafka.DescribeVpcConnectionOutput, error) {
	in := &kafka.DescribeVpcConnectionInput{
		Arn: aws.String(arn),
	}

	out, err := conn.DescribeVpcConnection(ctx, in)

	if errs.IsA[*types.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		}
	}

	if err != nil {
		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
