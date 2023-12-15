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

// @SDKResource("aws_prometheus_rule_group_namespace")
func ResourceRuleGroupNamespace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceRuleGroupNamespaceCreate,
		ReadWithoutTimeout:   resourceRuleGroupNamespaceRead,
		UpdateWithoutTimeout: resourceRuleGroupNamespaceUpdate,
		DeleteWithoutTimeout: resourceRuleGroupNamespaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"data": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceRuleGroupNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	workspaceID := d.Get("workspace_id").(string)
	name := d.Get("name").(string)
	input := &prometheusservice.CreateRuleGroupsNamespaceInput{
		Data:        []byte(d.Get("data").(string)),
		Name:        aws.String(name),
		WorkspaceId: aws.String(workspaceID),
	}

	output, err := conn.CreateRuleGroupsNamespaceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Prometheus Rule Group Namespace (%s) for Workspace (%s): %s", name, workspaceID, err)
	}

	d.SetId(aws.StringValue(output.Arn))

	if _, err := waitRuleGroupNamespaceCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Rule Group Namespace (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceRuleGroupNamespaceRead(ctx, d, meta)...)
}

func resourceRuleGroupNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	rgn, err := FindRuleGroupNamespaceByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Prometheus Rule Group Namespace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Rule Group Namespace (%s): %s", d.Id(), err)
	}

	d.Set("data", string(rgn.Data))
	d.Set("name", rgn.Name)
	_, workspaceID, err := nameAndWorkspaceIDFromRuleGroupNamespaceARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("workspace_id", workspaceID)

	return diags
}

func resourceRuleGroupNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	input := &prometheusservice.PutRuleGroupsNamespaceInput{
		Data:        []byte(d.Get("data").(string)),
		Name:        aws.String(d.Get("name").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	_, err := conn.PutRuleGroupsNamespaceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Prometheus Rule Group Namespace (%s): %s", d.Id(), err)
	}

	if _, err := waitRuleGroupNamespaceUpdated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Rule Group Namespace (%s) update: %s", d.Id(), err)
	}

	return append(diags, resourceRuleGroupNamespaceRead(ctx, d, meta)...)
}

func resourceRuleGroupNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	log.Printf("[DEBUG] Deleting Prometheus Rule Group Namespace: (%s)", d.Id())
	_, err := conn.DeleteRuleGroupsNamespaceWithContext(ctx, &prometheusservice.DeleteRuleGroupsNamespaceInput{
		Name:        aws.String(d.Get("name").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Prometheus Rule Group Namespace (%s): %s", d.Id(), err)
	}

	if _, err := waitRuleGroupNamespaceDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Rule Group Namespace (%s) delete: %s", d.Id(), err)
	}

	return diags
}
