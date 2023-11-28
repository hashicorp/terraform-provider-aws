// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/prometheusservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_prometheus_alert_manager_definition")
func ResourceAlertManagerDefinition() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAlertManagerDefinitionCreate,
		ReadWithoutTimeout:   resourceAlertManagerDefinitionRead,
		UpdateWithoutTimeout: resourceAlertManagerDefinitionUpdate,
		DeleteWithoutTimeout: resourceAlertManagerDefinitionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"definition": {
				Type:     schema.TypeString,
				Required: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAlertManagerDefinitionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	workspaceID := d.Get("workspace_id").(string)
	input := &prometheusservice.CreateAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(workspaceID),
	}

	_, err := conn.CreateAlertManagerDefinitionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Prometheus Alert Manager Definition (%s): %s", workspaceID, err)
	}

	d.SetId(workspaceID)

	if _, err := waitAlertManagerDefinitionCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Alert Manager Definition (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceAlertManagerDefinitionRead(ctx, d, meta)...)
}

func resourceAlertManagerDefinitionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	amd, err := FindAlertManagerDefinitionByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Prometheus Alert Manager Definition (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Alert Manager Definition (%s): %s", d.Id(), err)
	}

	d.Set("definition", string(amd.Data))
	d.Set("workspace_id", d.Id())

	return diags
}

func resourceAlertManagerDefinitionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	input := &prometheusservice.PutAlertManagerDefinitionInput{
		Data:        []byte(d.Get("definition").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	_, err := conn.PutAlertManagerDefinitionWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Prometheus Alert Manager Definition (%s): %s", d.Id(), err)
	}

	if _, err := waitAlertManagerDefinitionUpdated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Alert Manager Definition (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceAlertManagerDefinitionRead(ctx, d, meta)...)
}

func resourceAlertManagerDefinitionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	log.Printf("[DEBUG] Deleting Prometheus Alert Manager Definition: (%s)", d.Id())
	_, err := conn.DeleteAlertManagerDefinitionWithContext(ctx, &prometheusservice.DeleteAlertManagerDefinitionInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Prometheus Alert Manager Definition (%s): %s", d.Id(), err)
	}

	if _, err := waitAlertManagerDefinitionDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Alert Manager Definition (%s) delete: %s", d.Id(), err)
	}

	return diags
}
