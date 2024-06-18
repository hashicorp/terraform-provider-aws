// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/shield"
	awstypes "github.com/aws/aws-sdk-go-v2/service/shield/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// @SDKResource("aws_shield_protection_health_check_association")
func ResourceProtectionHealthCheckAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: ResourceProtectionHealthCheckAssociationCreate,
		ReadWithoutTimeout:   ResourceProtectionHealthCheckAssociationRead,
		DeleteWithoutTimeout: ResourceProtectionHealthCheckAssociationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"shield_protection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"health_check_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func ResourceProtectionHealthCheckAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	protectionId := d.Get("shield_protection_id").(string)
	healthCheckArn := d.Get("health_check_arn").(string)
	id := ProtectionHealthCheckAssociationCreateResourceID(protectionId, healthCheckArn)

	input := &shield.AssociateHealthCheckInput{
		ProtectionId:   aws.String(protectionId),
		HealthCheckArn: aws.String(healthCheckArn),
	}

	_, err := conn.AssociateHealthCheck(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "associating Route53 Health Check (%s) with Shield Protected resource (%s): %s", d.Get("health_check_arn"), d.Get("shield_protection_id"), err)
	}
	d.SetId(id)
	return append(diags, ResourceProtectionHealthCheckAssociationRead(ctx, d, meta)...)
}

func ResourceProtectionHealthCheckAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	protectionId, healthCheckArn, err := ProtectionHealthCheckAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Shield Protection and Route53 Health Check Association ID: %s", err)
	}

	input := &shield.DescribeProtectionInput{
		ProtectionId: aws.String(protectionId),
	}

	resp, err := conn.DescribeProtection(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Shield Protection itself (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Shield Protection Health Check Association (%s): %s", d.Id(), err)
	}

	isHealthCheck := stringInSlice(strings.Split(healthCheckArn, "/")[1], resp.Protection.HealthCheckIds)
	if !isHealthCheck {
		log.Printf("[WARN] Shield Protection Health Check Association (%s) not found, removing from state", d.Id())
		d.SetId("")
	}

	d.Set("health_check_arn", healthCheckArn)
	d.Set("shield_protection_id", resp.Protection.Id)

	return diags
}

func ResourceProtectionHealthCheckAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ShieldClient(ctx)

	protectionId, healthCheckId, err := ProtectionHealthCheckAssociationParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Shield Protection and Route53 Health Check Association ID: %s", err)
	}

	input := &shield.DisassociateHealthCheckInput{
		ProtectionId:   aws.String(protectionId),
		HealthCheckArn: aws.String(healthCheckId),
	}

	_, err = conn.DisassociateHealthCheck(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Route53 Health Check (%s) from Shield Protected resource (%s): %s", d.Get("health_check_arn"), d.Get("shield_protection_id"), err)
	}
	return diags
}

func stringInSlice(expected string, list []string) bool {
	for _, item := range list {
		if item == expected {
			return true
		}
	}
	return false
}
