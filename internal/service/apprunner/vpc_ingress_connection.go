// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apprunner

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apprunner_vpc_ingress_connection", name="VPC Ingress Connection")
// @Tags(identifierAttribute="arn")
func ResourceVPCIngressConnection() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceVPCIngressConnectionCreate,
		ReadWithoutTimeout:   resourceVPCIngressConnectionRead,
		UpdateWithoutTimeout: resourceVPCIngressConnectionUpdate,
		DeleteWithoutTimeout: resourceVPCIngressConnectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ingress_vpc_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vpc_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vpc_endpoint_id": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
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
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	name := d.Get("name").(string)
	input := &apprunner.CreateVpcIngressConnectionInput{
		ServiceArn:               aws.String(d.Get("service_arn").(string)),
		Tags:                     getTagsIn(ctx),
		VpcIngressConnectionName: aws.String(name),
	}

	if v, ok := d.GetOk("ingress_vpc_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		input.IngressVpcConfiguration = expandIngressVPCConfiguration(v.([]interface{}))
	}

	output, err := conn.CreateVpcIngressConnectionWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating App Runner VPC Ingress Configuration (%s): %s", name, err)
	}

	if output == nil || output.VpcIngressConnection == nil {
		return diag.Errorf("creating App Runner VPC Ingress Configuration (%s): empty output", name)
	}

	d.SetId(aws.StringValue(output.VpcIngressConnection.VpcIngressConnectionArn))

	if err := WaitVPCIngressConnectionActive(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for App Runner VPC Ingress Configuration (%s) creation: %s", d.Id(), err)
	}

	return resourceVPCIngressConnectionRead(ctx, d, meta)
}

func resourceVPCIngressConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	input := &apprunner.DescribeVpcIngressConnectionInput{
		VpcIngressConnectionArn: aws.String(d.Id()),
	}

	output, err := conn.DescribeVpcIngressConnectionWithContext(ctx, input)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] App Runner VPC Ingress Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading App Runner VPC Ingress Configuration (%s): %s", d.Id(), err)
	}

	if output == nil || output.VpcIngressConnection == nil {
		return diag.Errorf("reading App Runner VPC Ingress Configuration (%s): empty output", d.Id())
	}

	if aws.StringValue(output.VpcIngressConnection.Status) == VPCIngressConnectionStatusDeleted {
		if d.IsNewResource() {
			return diag.Errorf("reading App Runner VPC Ingress Configuration (%s): %s after creation", d.Id(), aws.StringValue(output.VpcIngressConnection.Status))
		}
		log.Printf("[WARN] App Runner VPC Ingress Configuration (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	config := output.VpcIngressConnection
	arn := aws.StringValue(config.VpcIngressConnectionArn)

	d.Set("arn", arn)
	d.Set("service_arn", config.ServiceArn)
	d.Set("name", config.VpcIngressConnectionName)
	d.Set("status", config.Status)
	d.Set("domain_name", config.DomainName)

	if err := d.Set("ingress_vpc_configuration", flattenIngressVPCConfiguration(config.IngressVpcConfiguration)); err != nil {
		return diag.Errorf("setting ingress_vpc_configuration: %s", err)
	}

	return nil
}

func resourceVPCIngressConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Tags only.
	return resourceVPCIngressConnectionRead(ctx, d, meta)
}

func resourceVPCIngressConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).AppRunnerConn(ctx)

	input := &apprunner.DeleteVpcIngressConnectionInput{
		VpcIngressConnectionArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteVpcIngressConnectionWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting App Runner VPC Ingress Configuration (%s): %s", d.Id(), err)
	}

	if err := WaitVPCIngressConnectionDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, apprunner.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("waiting for App Runner VPC Ingress Configuration (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func expandIngressVPCConfiguration(l []interface{}) *apprunner.IngressVpcConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	configuration := &apprunner.IngressVpcConfiguration{}

	if v, ok := m["vpc_id"].(string); ok && v != "" {
		configuration.VpcId = aws.String(v)
	}

	if v, ok := m["vpc_endpoint_id"].(string); ok && v != "" {
		configuration.VpcEndpointId = aws.String(v)
	}

	return configuration
}

func flattenIngressVPCConfiguration(ingressVpcConfiguration *apprunner.IngressVpcConfiguration) []interface{} {
	if ingressVpcConfiguration == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"vpc_id":          aws.StringValue(ingressVpcConfiguration.VpcId),
		"vpc_endpoint_id": aws.StringValue(ingressVpcConfiguration.VpcEndpointId),
	}

	return []interface{}{m}
}
