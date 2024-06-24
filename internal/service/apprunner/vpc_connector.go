// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apprunner"
	"github.com/aws/aws-sdk-go-v2/service/apprunner/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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

// @SDKResource("aws_apprunner_vpc_connector", name="VPC Connector")
// @Tags(identifierAttribute="arn")
func resourceVPCConnector() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCConnectorCreate,
		ReadWithoutTimeout:   resourceVPCConnectorRead,
		UpdateWithoutTimeout: resourceVPCConnectorUpdate,
		DeleteWithoutTimeout: resourceVPCConnectorDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrSubnets: {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"vpc_connector_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(4, 40),
			},
			"vpc_connector_revision": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	name := d.Get("vpc_connector_name").(string)
	input := &apprunner.CreateVpcConnectorInput{
		SecurityGroups:   flex.ExpandStringValueSet(d.Get(names.AttrSecurityGroups).(*schema.Set)),
		Subnets:          flex.ExpandStringValueSet(d.Get(names.AttrSubnets).(*schema.Set)),
		Tags:             getTagsIn(ctx),
		VpcConnectorName: aws.String(name),
	}

	output, err := conn.CreateVpcConnector(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Runner VPC Connector (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.VpcConnector.VpcConnectorArn))

	if _, err := waitVPCConnectorCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner VPC Connector (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCConnectorRead(ctx, d, meta)...)
}

func resourceVPCConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	vpcConnector, err := findVPCConnectorByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Runner VPC Connector (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Runner VPC Connector (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, vpcConnector.VpcConnectorArn)
	d.Set(names.AttrSecurityGroups, vpcConnector.SecurityGroups)
	d.Set(names.AttrStatus, vpcConnector.Status)
	d.Set(names.AttrSubnets, vpcConnector.Subnets)
	d.Set("vpc_connector_name", vpcConnector.VpcConnectorName)
	d.Set("vpc_connector_revision", vpcConnector.VpcConnectorRevision)

	return diags
}

func resourceVPCConnectorUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceVPCConnectorRead(ctx, d, meta)
}

func resourceVPCConnectorDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	log.Printf("[DEBUG] Deleting App Runner VPC Connector: %s", d.Id())
	_, err := conn.DeleteVpcConnector(ctx, &apprunner.DeleteVpcConnectorInput{
		VpcConnectorArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Runner VPC Connector (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCConnectorDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner VPC Connector (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findVPCConnectorByARN(ctx context.Context, conn *apprunner.Client, arn string) (*types.VpcConnector, error) {
	input := &apprunner.DescribeVpcConnectorInput{
		VpcConnectorArn: aws.String(arn),
	}

	output, err := conn.DescribeVpcConnector(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VpcConnector == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.VpcConnector.Status; status == types.VpcConnectorStatusInactive {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output.VpcConnector, nil
}

func statusVPCConnector(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCConnectorByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}

func waitVPCConnectorCreated(ctx context.Context, conn *apprunner.Client, arn string) (*types.VpcConnector, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Target:  enum.Slice(types.VpcConnectorStatusActive),
		Refresh: statusVPCConnector(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcConnector); ok {
		return output, err
	}

	return nil, err
}

func waitVPCConnectorDeleted(ctx context.Context, conn *apprunner.Client, arn string) (*types.VpcConnector, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VpcConnectorStatusActive),
		Target:  []string{},
		Refresh: statusVPCConnector(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcConnector); ok {
		return output, err
	}

	return nil, err
}
