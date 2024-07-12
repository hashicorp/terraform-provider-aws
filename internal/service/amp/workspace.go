// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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

// @SDKResource("aws_prometheus_workspace", name="Workspace")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/amp/types;types.WorkspaceDescription")
func resourceWorkspace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceWorkspaceCreate,
		ReadWithoutTimeout:   resourceWorkspaceRead,
		UpdateWithoutTimeout: resourceWorkspaceUpdate,
		DeleteWithoutTimeout: resourceWorkspaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			// Once set, alias cannot be unset.
			customdiff.ForceNewIfChange(names.AttrAlias, func(_ context.Context, old, new, meta interface{}) bool {
				return old.(string) != "" && new.(string) == ""
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			names.AttrAlias: {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrKMSKeyARN: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrLoggingConfiguration: {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"log_group_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
					},
				},
			},
			"prometheus_endpoint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceWorkspaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	input := &amp.CreateWorkspaceInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrAlias); ok {
		input.Alias = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrKMSKeyARN); ok {
		input.KmsKeyArn = aws.String(v.(string))
	}

	output, err := conn.CreateWorkspace(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Prometheus Workspace: %s", err)
	}

	d.SetId(aws.ToString(output.WorkspaceId))

	if _, err := waitWorkspaceCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk(names.AttrLoggingConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input := &amp.CreateLoggingConfigurationInput{
			LogGroupArn: aws.String(tfMap["log_group_arn"].(string)),
			WorkspaceId: aws.String(d.Id()),
		}

		_, err := conn.CreateLoggingConfiguration(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "creating Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
		}

		if _, err := waitLoggingConfigurationCreated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) logging configuration create: %s", d.Id(), err)
		}
	}

	return append(diags, resourceWorkspaceRead(ctx, d, meta)...)
}

func resourceWorkspaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	ws, err := findWorkspaceByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Prometheus Workspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Workspace (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAlias, ws.Alias)
	arn := aws.ToString(ws.Arn)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrKMSKeyARN, ws.KmsKeyArn)
	d.Set("prometheus_endpoint", ws.PrometheusEndpoint)

	loggingConfiguration, err := findLoggingConfigurationByWorkspaceID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		d.Set(names.AttrLoggingConfiguration, nil)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
	} else {
		if err := d.Set(names.AttrLoggingConfiguration, []interface{}{flattenLoggingConfigurationMetadata(loggingConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging_configuration: %s", err)
		}
	}

	return diags
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	if d.HasChange(names.AttrAlias) {
		input := &amp.UpdateWorkspaceAliasInput{
			Alias:       aws.String(d.Get(names.AttrAlias).(string)),
			WorkspaceId: aws.String(d.Id()),
		}

		_, err := conn.UpdateWorkspaceAlias(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Prometheus Workspace alias (%s): %s", d.Id(), err)
		}

		if _, err := waitWorkspaceUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange(names.AttrLoggingConfiguration) {
		if v, ok := d.GetOk(names.AttrLoggingConfiguration); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if o, _ := d.GetChange(names.AttrLoggingConfiguration); o == nil || len(o.([]interface{})) == 0 || o.([]interface{})[0] == nil {
				input := &amp.CreateLoggingConfigurationInput{
					LogGroupArn: aws.String(tfMap["log_group_arn"].(string)),
					WorkspaceId: aws.String(d.Id()),
				}

				if _, err := conn.CreateLoggingConfiguration(ctx, input); err != nil {
					return sdkdiag.AppendErrorf(diags, "creating Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
				}

				if _, err := waitLoggingConfigurationCreated(ctx, conn, d.Id()); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) logging configuration create: %s", d.Id(), err)
				}
			} else {
				input := &amp.UpdateLoggingConfigurationInput{
					LogGroupArn: aws.String(tfMap["log_group_arn"].(string)),
					WorkspaceId: aws.String(d.Id()),
				}

				if _, err := conn.UpdateLoggingConfiguration(ctx, input); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
				}

				if _, err := waitLoggingConfigurationUpdated(ctx, conn, d.Id()); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) logging configuration update: %s", d.Id(), err)
				}
			}
		} else {
			input := &amp.DeleteLoggingConfigurationInput{
				WorkspaceId: aws.String(d.Id()),
			}

			_, err := conn.DeleteLoggingConfiguration(ctx, input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "deleting Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
			}

			if _, err := waitLoggingConfigurationDeleted(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) logging configuration delete: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceWorkspaceRead(ctx, d, meta)...)
}

func resourceWorkspaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).AMPClient(ctx)

	log.Printf("[INFO] Deleting Prometheus Workspace: %s", d.Id())
	_, err := conn.DeleteWorkspace(ctx, &amp.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Prometheus Workspace (%s): %s", d.Id(), err)
	}

	if _, err := waitWorkspaceDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func flattenLoggingConfigurationMetadata(apiObject *types.LoggingConfigurationMetadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogGroupArn; v != nil {
		tfMap["log_group_arn"] = aws.ToString(v)
	}

	return tfMap
}

func findWorkspaceByID(ctx context.Context, conn *amp.Client, id string) (*types.WorkspaceDescription, error) {
	input := &amp.DescribeWorkspaceInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeWorkspace(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Workspace == nil || output.Workspace.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Workspace, nil
}

func statusWorkspace(ctx context.Context, conn *amp.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findWorkspaceByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func waitWorkspaceCreated(ctx context.Context, conn *amp.Client, id string) (*types.WorkspaceDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WorkspaceStatusCodeCreating),
		Target:  enum.Slice(types.WorkspaceStatusCodeActive),
		Refresh: statusWorkspace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.WorkspaceDescription); ok {
		return output, err
	}

	return nil, err
}

func waitWorkspaceUpdated(ctx context.Context, conn *amp.Client, id string) (*types.WorkspaceDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WorkspaceStatusCodeUpdating),
		Target:  enum.Slice(types.WorkspaceStatusCodeActive),
		Refresh: statusWorkspace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.WorkspaceDescription); ok {
		return output, err
	}

	return nil, err
}

func waitWorkspaceDeleted(ctx context.Context, conn *amp.Client, id string) (*types.WorkspaceDescription, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WorkspaceStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusWorkspace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.WorkspaceDescription); ok {
		return output, err
	}

	return nil, err
}

func findLoggingConfigurationByWorkspaceID(ctx context.Context, conn *amp.Client, id string) (*types.LoggingConfigurationMetadata, error) {
	input := &amp.DescribeLoggingConfigurationInput{
		WorkspaceId: aws.String(id),
	}

	output, err := conn.DescribeLoggingConfiguration(ctx, input)

	if errs.IsA[*types.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.LoggingConfiguration == nil || output.LoggingConfiguration.Status == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.LoggingConfiguration, nil
}

func statusLoggingConfiguration(ctx context.Context, conn *amp.Client, workspaceID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findLoggingConfigurationByWorkspaceID(ctx, conn, workspaceID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status.StatusCode), nil
	}
}

func waitLoggingConfigurationCreated(ctx context.Context, conn *amp.Client, workspaceID string) (*types.LoggingConfigurationMetadata, error) { //nolint:unparam
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.LoggingConfigurationStatusCodeCreating),
		Target:  enum.Slice(types.LoggingConfigurationStatusCodeActive),
		Refresh: statusLoggingConfiguration(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.LoggingConfigurationMetadata); ok {
		if statusCode := output.Status.StatusCode; statusCode == types.LoggingConfigurationStatusCodeCreationFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitLoggingConfigurationUpdated(ctx context.Context, conn *amp.Client, workspaceID string) (*types.LoggingConfigurationMetadata, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.LoggingConfigurationStatusCodeUpdating),
		Target:  enum.Slice(types.LoggingConfigurationStatusCodeActive),
		Refresh: statusLoggingConfiguration(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.LoggingConfigurationMetadata); ok {
		if statusCode := output.Status.StatusCode; statusCode == types.LoggingConfigurationStatusCodeUpdateFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.Status.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitLoggingConfigurationDeleted(ctx context.Context, conn *amp.Client, workspaceID string) (*types.LoggingConfigurationMetadata, error) {
	const (
		timeout = 5 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.LoggingConfigurationStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusLoggingConfiguration(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.LoggingConfigurationMetadata); ok {
		return output, err
	}

	return nil, err
}
