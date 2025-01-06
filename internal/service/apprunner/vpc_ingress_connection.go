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
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apprunner_vpc_ingress_connection", name="VPC Ingress Connection")
// @Tags(identifierAttribute="arn")
func resourceVPCIngressConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCIngressConnectionCreate,
		ReadWithoutTimeout:   resourceVPCIngressConnectionRead,
		UpdateWithoutTimeout: resourceVPCIngressConnectionUpdate,
		DeleteWithoutTimeout: resourceVPCIngressConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ingress_vpc_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrVPCEndpointID: {
							Type:     schema.TypeString,
							Optional: true,
						},
						names.AttrVPCID: {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			"service_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceVPCIngressConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &apprunner.CreateVpcIngressConnectionInput{
		ServiceArn:               aws.String(d.Get("service_arn").(string)),
		Tags:                     getTagsIn(ctx),
		VpcIngressConnectionName: aws.String(name),
	}

	if v, ok := d.GetOk("ingress_vpc_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.IngressVpcConfiguration = expandIngressVPCConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateVpcIngressConnection(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating App Runner VPC Ingress Connection (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.VpcIngressConnection.VpcIngressConnectionArn))

	if _, err := waitVPCIngressConnectionCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner VPC Ingress Connection (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceVPCIngressConnectionRead(ctx, d, meta)...)
}

func resourceVPCIngressConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	connection, err := findVPCIngressConnectionByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] App Runner VPC Ingress Connection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading App Runner VPC Ingress Connection (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, connection.VpcIngressConnectionArn)
	d.Set(names.AttrDomainName, connection.DomainName)
	if err := d.Set("ingress_vpc_configuration", flattenIngressVPCConfiguration(connection.IngressVpcConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ingress_vpc_configuration: %s", err)
	}
	d.Set(names.AttrName, connection.VpcIngressConnectionName)
	d.Set("service_arn", connection.ServiceArn)
	d.Set(names.AttrStatus, connection.Status)

	return diags
}

func resourceVPCIngressConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceVPCIngressConnectionRead(ctx, d, meta)
}

func resourceVPCIngressConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AppRunnerClient(ctx)

	log.Printf("[INFO] Deleting App Runner VPC Ingress Connection: %s", d.Id())
	_, err := conn.DeleteVpcIngressConnection(ctx, &apprunner.DeleteVpcIngressConnectionInput{
		VpcIngressConnectionArn: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting App Runner VPC Ingress Connection (%s): %s", d.Id(), err)
	}

	if _, err := waitVPCIngressConnectionDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for App Runner VPC Ingress Connection (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findVPCIngressConnectionByARN(ctx context.Context, conn *apprunner.Client, arn string) (*types.VpcIngressConnection, error) {
	input := &apprunner.DescribeVpcIngressConnectionInput{
		VpcIngressConnectionArn: aws.String(arn),
	}

	output, err := conn.DescribeVpcIngressConnection(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.VpcIngressConnection == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if status := output.VpcIngressConnection.Status; status == types.VpcIngressConnectionStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output.VpcIngressConnection, nil
}

func statusVPCIngressConnection(ctx context.Context, conn *apprunner.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findVPCIngressConnectionByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
func waitVPCIngressConnectionCreated(ctx context.Context, conn *apprunner.Client, arn string) (*types.VpcIngressConnection, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VpcIngressConnectionStatusPendingCreation),
		Target:  enum.Slice(types.VpcIngressConnectionStatusAvailable),
		Refresh: statusVPCIngressConnection(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcIngressConnection); ok {
		return output, err
	}

	return nil, err
}

func waitVPCIngressConnectionDeleted(ctx context.Context, conn *apprunner.Client, arn string) (*types.VpcIngressConnection, error) {
	const (
		timeout = 2 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.VpcIngressConnectionStatusAvailable, types.VpcIngressConnectionStatusPendingDeletion),
		Target:  []string{},
		Refresh: statusVPCIngressConnection(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.VpcIngressConnection); ok {
		return output, err
	}

	return nil, err
}

func expandIngressVPCConfiguration(l []interface{}) *types.IngressVpcConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &types.IngressVpcConfiguration{}

	if v, ok := m[names.AttrVPCID].(string); ok && v != "" {
		configuration.VpcId = aws.String(v)
	}

	if v, ok := m[names.AttrVPCEndpointID].(string); ok && v != "" {
		configuration.VpcEndpointId = aws.String(v)
	}

	return configuration
}

func flattenIngressVPCConfiguration(ingressVpcConfiguration *types.IngressVpcConfiguration) []interface{} {
	if ingressVpcConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		names.AttrVPCID:         aws.ToString(ingressVpcConfiguration.VpcId),
		names.AttrVPCEndpointID: aws.ToString(ingressVpcConfiguration.VpcEndpointId),
	}

	return []interface{}{m}
}
