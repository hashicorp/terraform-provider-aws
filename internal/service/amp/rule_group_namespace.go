// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_prometheus_rule_group_namespace", name="Rule Group Namespace")
func resourceRuleGroupNamespace() *schema.Resource {
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
			names.AttrName: {
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
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	workspaceID := d.Get("workspace_id").(string)
	name := d.Get(names.AttrName).(string)
	input := &amp.CreateRuleGroupsNamespaceInput{
		Data:        []byte(d.Get("data").(string)),
		Name:        aws.String(name),
		WorkspaceId: aws.String(workspaceID),
	}

	output, err := conn.CreateRuleGroupsNamespace(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Prometheus Rule Group Namespace (%s) for Workspace (%s): %s", name, workspaceID, err)
	}

	d.SetId(aws.ToString(output.Arn))

	if _, err := waitRuleGroupNamespaceCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Rule Group Namespace (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceRuleGroupNamespaceRead(ctx, d, meta)...)
}

func resourceRuleGroupNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	rgn, err := findRuleGroupNamespaceByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Prometheus Rule Group Namespace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Rule Group Namespace (%s): %s", d.Id(), err)
	}

	d.Set("data", string(rgn.Data))
	d.Set(names.AttrName, rgn.Name)
	_, workspaceID, err := nameAndWorkspaceIDFromRuleGroupNamespaceARN(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	d.Set("workspace_id", workspaceID)

	return diags
}

func resourceRuleGroupNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	input := &amp.PutRuleGroupsNamespaceInput{
		Data:        []byte(d.Get("data").(string)),
		Name:        aws.String(d.Get(names.AttrName).(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	_, err := conn.PutRuleGroupsNamespace(ctx, input)

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
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	log.Printf("[DEBUG] Deleting Prometheus Rule Group Namespace: (%s)", d.Id())
	_, err := conn.DeleteRuleGroupsNamespace(ctx, &amp.DeleteRuleGroupsNamespaceInput{
		Name:        aws.String(d.Get(names.AttrName).(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

func nameAndWorkspaceIDFromRuleGroupNamespaceARN(v string) (string, string, error) {
	p, err := arn.Parse(v)
	if err != nil {
		return "", "", err
	}

	v = p.Resource
	parts := strings.Split(v, "/")

	// https://docs.aws.amazon.com/service-authorization/latest/reference/list_amazonmanagedserviceforprometheus.html#amazonmanagedserviceforprometheus-resources-for-iam-policies.
	if len(parts) == 3 && parts[0] == "rulegroupsnamespace" && parts[1] != "" && parts[2] != "" {
		return parts[2], parts[1], nil
	}

	return "", "", fmt.Errorf("unexpected format for resource (%[1]s), expected rulegroupsnamespace/${WorkspaceId}/${Namespace}", v)
}

func findRuleGroupNamespaceByARN(ctx context.Context, conn *amp.Client, arn string) (*types.RuleGroupsNamespaceDescription, error) {
	name, workspaceID, err := nameAndWorkspaceIDFromRuleGroupNamespaceARN(arn)
	if err != nil {
		return nil, err
	}

	input := &amp.DescribeRuleGroupsNamespaceInput{
		Name:        aws.String(name),
		WorkspaceId: aws.String(workspaceID),
	}

	output, err := conn.DescribeRuleGroupsNamespace(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.RuleGroupsNamespace == nil || output.RuleGroupsNamespace.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.RuleGroupsNamespace, nil
}

func statusRuleGroupNamespace(ctx context.Context, conn *amp.Client, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRuleGroupNamespaceByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func waitRuleGroupNamespaceCreated(ctx context.Context, conn *amp.Client, id string) (*types.RuleGroupsNamespaceDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RuleGroupsNamespaceStatusCodeCreating),
		Target:  enum.Slice(types.RuleGroupsNamespaceStatusCodeActive),
		Refresh: statusRuleGroupNamespace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RuleGroupsNamespaceDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRuleGroupNamespaceUpdated(ctx context.Context, conn *amp.Client, id string) (*types.RuleGroupsNamespaceDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RuleGroupsNamespaceStatusCodeUpdating),
		Target:  enum.Slice(types.RuleGroupsNamespaceStatusCodeActive),
		Refresh: statusRuleGroupNamespace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RuleGroupsNamespaceDescription); ok {
		return output, err
	}

	return nil, err
}

func waitRuleGroupNamespaceDeleted(ctx context.Context, conn *amp.Client, id string) (*types.RuleGroupsNamespaceDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.RuleGroupsNamespaceStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusRuleGroupNamespace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.RuleGroupsNamespaceDescription); ok {
		return output, err
	}

	return nil, err
}
