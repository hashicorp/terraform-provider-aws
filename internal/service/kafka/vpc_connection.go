package kafka

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kafka"
	"github.com/aws/aws-sdk-go-v2/service/kafka/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_msk_vpc_connection", name="Vpc Connection")
func ResourceVpcConnection() *schema.Resource {
	return &schema.Resource{

		CreateWithoutTimeout: resourceVpcConnectionCreate,
		ReadWithoutTimeout:   resourceVpcConnectionRead,
		DeleteWithoutTimeout: resourceVpcConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
				ForceNew: true,
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
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"target_cluster_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

const (
	ResNameVpcConnection = "Vpc Connection"
)

func resourceVpcConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	in := &kafka.CreateVpcConnectionInput{
		Authentication:   aws.String(d.Get("authentication").(string)),
		ClientSubnets:    flex.ExpandStringValueSet(d.Get("client_subnets").(*schema.Set)),
		TargetClusterArn: aws.String(d.Get("target_cluster_arn").(string)),
		VpcId:            aws.String(d.Get("vpc_id").(string)),
	}

	if v, ok := d.GetOk("security_groups"); ok {
		in.SecurityGroups = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	out, err := conn.CreateVpcConnection(ctx, in)
	if err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionCreating, ResNameVpcConnection, d.Get("arn").(string), err)...)
	}

	if out == nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionCreating, ResNameVpcConnection, d.Get("arn").(string), errors.New("empty output"))...)
	}

	d.SetId(aws.ToString(out.VpcConnectionArn))

	if _, err := waitVpcConnectionCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionWaitingForCreation, ResNameVpcConnection, d.Id(), err)...)
	}

	return append(diags, resourceVpcConnectionRead(ctx, d, meta)...)
}

func resourceVpcConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	out, err := FindVpcConnectionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Kafka VpcConnection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionReading, ResNameVpcConnection, d.Id(), err)...)
	}

	d.Set("authentication", out.Authentication)
	d.Set("arn", out.VpcConnectionArn)
	d.Set("vpc_id", out.VpcId)
	d.Set("target_cluster_arn", out.TargetClusterArn)

	if err := d.Set("client_subnets", flex.FlattenStringValueSet(out.Subnets)); err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionSetting, ResNameVpcConnection, d.Id(), err)...)
	}

	if err := d.Set("security_groups", flex.FlattenStringValueSet(out.SecurityGroups)); err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionSetting, ResNameVpcConnection, d.Id(), err)...)
	}

	return diags
}

func resourceVpcConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).KafkaClient(ctx)

	log.Printf("[INFO] Deleting Kafka VpcConnection %s", d.Id())

	_, err := conn.DeleteVpcConnection(ctx, &kafka.DeleteVpcConnectionInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}
	if err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionDeleting, ResNameVpcConnection, d.Id(), err)...)
	}

	if _, err := waitVpcConnectionDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return append(diags, create.DiagError(names.Kafka, create.ErrActionWaitingForDeletion, ResNameVpcConnection, d.Id(), err)...)
	}

	return diags
}

func waitVpcConnectionCreated(ctx context.Context, conn *kafka.Client, id string, timeout time.Duration) (*kafka.DescribeVpcConnectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(types.VpcConnectionStateCreating),
		Target:                    enum.Slice(types.VpcConnectionStateAvailable),
		Refresh:                   statusVpcConnection(ctx, conn, id),
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

func waitVpcConnectionDeleted(ctx context.Context, conn *kafka.Client, arn string, timeout time.Duration) (*kafka.DescribeVpcConnectionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VpcConnectionStateAvailable, types.VpcConnectionStateInactive, types.VpcConnectionStateDeactivating, types.VpcConnectionStateDeleting),
		Target:  []string{},
		Refresh: statusVpcConnection(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*kafka.DescribeVpcConnectionOutput); ok {
		return out, err
	}

	return nil, err
}

func statusVpcConnection(ctx context.Context, conn *kafka.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindVpcConnectionByARN(ctx, conn, arn)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.State), nil
	}
}

func FindVpcConnectionByARN(ctx context.Context, conn *kafka.Client, arn string) (*kafka.DescribeVpcConnectionOutput, error) {
	in := &kafka.DescribeVpcConnectionInput{
		Arn: aws.String(arn),
	}

	out, err := conn.DescribeVpcConnection(ctx, in)
	if err != nil {
		var nfe *types.NotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: in,
			}
		}

		return nil, err
	}

	if out == nil {
		return nil, tfresource.NewEmptyResultError(in)
	}

	return out, nil
}
