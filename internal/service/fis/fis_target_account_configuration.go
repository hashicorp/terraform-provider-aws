// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fis

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/YakDriver/smarterr"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/fis"
	"github.com/aws/aws-sdk-go-v2/service/fis/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	sdkid "github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/smerr"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// Function annotations are used for resource registration to the Provider. DO NOT EDIT.
// @SDKResource("aws_fis_target_account_configuration", name="Target Account Configuration")
func ResourceTargetAccountConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceTargetAccountConfigurationCreate,
		ReadWithoutTimeout:   resourceTargetAccountConfigurationRead,
		UpdateWithoutTimeout: resourceTargetAccountConfigurationUpdate,
		DeleteWithoutTimeout: resourceTargetAccountConfigurationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidAwsAccountID,
				ForceNew:     true,
			},
			"experiment_template_id": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
				ForceNew:     true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
				ForceNew:     true,
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512),
			},
		},
	}
}

const (
	ResNameTargetAccountConfiguration = "Target Account Configuration"
)

func resourceTargetAccountConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	in := &fis.CreateTargetAccountConfigurationInput{
		AccountId:            aws.String(d.Get("account_id").(string)),
		ExperimentTemplateId: aws.String(d.Get("experiment_template_id").(string)),
		RoleArn:              aws.String(d.Get("role_arn").(string)),
		ClientToken:          aws.String(sdkid.UniqueId()),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	out, err := conn.CreateTargetAccountConfiguration(ctx, in)
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Get("experiment_template_id").(string))
	}

	if out == nil || out.TargetAccountConfiguration == nil {
		return smerr.Append(ctx, diags, tfresource.NewEmptyResultError(in), smerr.ID, d.Get("experiment_template_id").(string))
	}

	tac := out.TargetAccountConfiguration
	templateID := aws.ToString(tac.ExperimentTemplateId)
	if templateID == "" {
		templateID = d.Get("experiment_template_id").(string)
	}
	accountID := aws.ToString(tac.AccountId)
	if accountID == "" {
		accountID = d.Get("account_id").(string)
	}

	d.SetId(fmt.Sprintf("%s/%s", templateID, accountID))
	d.Set("description", aws.ToString(tac.Description))
	d.Set("role_arn", aws.ToString(tac.RoleArn))

	return append(diags, resourceTargetAccountConfigurationRead(ctx, d, meta)...)
}

func resourceTargetAccountConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	experimentTemplateId, accountId, err := parseTargetAccountConfigurationID(d.Id())
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	in := &fis.GetTargetAccountConfigurationInput{
		AccountId:            aws.String(accountId),
		ExperimentTemplateId: aws.String(experimentTemplateId),
	}

	out, err := conn.GetTargetAccountConfiguration(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err) {
		if !d.IsNewResource() {
			log.Printf("[WARN] FIS TargetAccountConfiguration (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	if out == nil || out.TargetAccountConfiguration == nil {
		return smerr.Append(ctx, diags, tfresource.NewEmptyResultError(in), smerr.ID, d.Id())
	}

	tac := out.TargetAccountConfiguration
	accountValue := accountId
	if v := aws.ToString(tac.AccountId); v != "" {
		accountValue = v
	}

	roleARN := d.Get("role_arn").(string)
	if v := aws.ToString(tac.RoleArn); v != "" {
		roleARN = v
	}

	d.Set("account_id", accountValue)
	d.Set("experiment_template_id", experimentTemplateId)
	d.Set("role_arn", roleARN)
	d.Set("description", aws.ToString(tac.Description))

	return diags
}

func resourceTargetAccountConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	if !d.HasChange("description") {
		return diags
	}

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	experimentTemplateId, accountId, err := parseTargetAccountConfigurationID(d.Id())
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	in := &fis.CreateTargetAccountConfigurationInput{
		AccountId:            aws.String(accountId),
		ExperimentTemplateId: aws.String(experimentTemplateId),
		RoleArn:              aws.String(d.Get("role_arn").(string)),
		ClientToken:          aws.String(sdkid.UniqueId()),
	}

	if v, ok := d.GetOk("description"); ok {
		in.Description = aws.String(v.(string))
	}

	_, err = conn.CreateTargetAccountConfiguration(ctx, in)
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	return append(diags, resourceTargetAccountConfigurationRead(ctx, d, meta)...)
}

func resourceTargetAccountConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).FISClient(ctx)

	experimentTemplateId, accountId, err := parseTargetAccountConfigurationID(d.Id())
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	log.Printf("[INFO] Deleting FIS TargetAccountConfiguration %s", d.Id())

	_, err = conn.DeleteTargetAccountConfiguration(ctx, &fis.DeleteTargetAccountConfigurationInput{
		AccountId:            aws.String(accountId),
		ExperimentTemplateId: aws.String(experimentTemplateId),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		d.SetId("")
		return diags
	}
	if err != nil {
		return smerr.Append(ctx, diags, err, smerr.ID, d.Id())
	}

	d.SetId("")

	return diags
}

// parseTargetAccountConfigurationID parses the composite ID into experiment template ID and account ID
func parseTargetAccountConfigurationID(id string) (experimentTemplateId, accountId string, err error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 {
		err = fmt.Errorf("invalid ID format: expected experiment_template_id/account_id, got %s", id)
		return
	}
	experimentTemplateId = parts[0]
	accountId = parts[1]
	return
}

const (
	statusChangePending = "Pending"
	statusDeleting      = "Deleting"
	statusNormal        = "Normal"
	statusUpdated       = "Updated"
)

func waitTargetAccountConfigurationCreated(ctx context.Context, conn *fis.Client, id string, timeout time.Duration) (*types.TargetAccountConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{},
		Target:                    []string{statusNormal},
		Refresh:                   statusTargetAccountConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.TargetAccountConfiguration); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitTargetAccountConfigurationUpdated(ctx context.Context, conn *fis.Client, id string, timeout time.Duration) (*types.TargetAccountConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{statusChangePending},
		Target:                    []string{statusUpdated},
		Refresh:                   statusTargetAccountConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.TargetAccountConfiguration); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func waitTargetAccountConfigurationDeleted(ctx context.Context, conn *fis.Client, id string, timeout time.Duration) (*types.TargetAccountConfiguration, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{statusDeleting, statusNormal},
		Target:  []string{},
		Refresh: statusTargetAccountConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*types.TargetAccountConfiguration); ok {
		return out, smarterr.NewError(err)
	}

	return nil, smarterr.NewError(err)
}

func statusTargetAccountConfiguration(ctx context.Context, conn *fis.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		out, err := findTargetAccountConfigurationByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", smarterr.NewError(err)
		}

		status := ""
		if out != nil {
			status = string(out.Status)
		}

		return out, status, nil
	}
}

func findTargetAccountConfigurationByID(ctx context.Context, conn *fis.Client, id string) (*types.TargetAccountConfiguration, error) {
	experimentTemplateID, accountID, err := parseTargetAccountConfigurationID(id)
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	in := &fis.GetTargetAccountConfigurationInput{
		AccountId:            aws.String(accountID),
		ExperimentTemplateId: aws.String(experimentTemplateID),
	}

	out, err := conn.GetTargetAccountConfiguration(ctx, in)
	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, smarterr.NewError(&retry.NotFoundError{
			LastError:   err,
			LastRequest: in,
		})
	}
	if err != nil {
		return nil, smarterr.NewError(err)
	}

	if out == nil || out.TargetAccountConfiguration == nil {
		return nil, smarterr.NewError(tfresource.NewEmptyResultError(in))
	}

	return out.TargetAccountConfiguration, nil
}

func flattenComplexArgument(apiObject *fis.ComplexArgument) map[string]any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{}

	if v := apiObject.SubFieldOne; v != nil {
		m["sub_field_one"] = aws.ToString(v)
	}

	if v := apiObject.SubFieldTwo; v != nil {
		m["sub_field_two"] = aws.ToString(v)
	}

	return m
}

func flattenComplexArguments(apiObjects []*fis.ComplexArgument) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	var l []any

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		l = append(l, flattenComplexArgument(apiObject))
	}

	return l
}

func expandComplexArgument(tfMap map[string]any) *fis.ComplexArgument {
	if tfMap == nil {
		return nil
	}

	a := &fis.ComplexArgument{}

	if v, ok := tfMap["sub_field_one"].(string); ok && v != "" {
		a.SubFieldOne = aws.String(v)
	}

	if v, ok := tfMap["sub_field_two"].(string); ok && v != "" {
		a.SubFieldTwo = aws.String(v)
	}

	return a
}

func expandComplexArguments(tfList []any) []*fis.ComplexArgument {
	if len(tfList) == 0 {
		return nil
	}

	var s []*fis.ComplexArgument

	for _, r := range tfList {
		m, ok := r.(map[string]any)

		if !ok {
			continue
		}

		a := expandComplexArgument(m)

		if a == nil {
			continue
		}

		s = append(s, a)
	}

	return s
}
