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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_prometheus_workspace", name="Workspace")
// @Tags(identifierAttribute="arn")
func ResourceWorkspace() *schema.Resource {
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
			customdiff.ForceNewIfChange("alias", func(_ context.Context, old, new, meta interface{}) bool {
				return old.(string) != "" && new.(string) == ""
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"logging_configuration": {
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

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	input := &prometheusservice.CreateWorkspaceInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("alias"); ok {
		input.Alias = aws.String(v.(string))
	}

	result, err := conn.CreateWorkspaceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Prometheus Workspace: %s", err)
	}

	d.SetId(aws.StringValue(result.WorkspaceId))

	if _, err := waitWorkspaceCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) create: %s", d.Id(), err)
	}

	if v, ok := d.GetOk("logging_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
		tfMap := v.([]interface{})[0].(map[string]interface{})
		input := &prometheusservice.CreateLoggingConfigurationInput{
			LogGroupArn: aws.String(tfMap["log_group_arn"].(string)),
			WorkspaceId: aws.String(d.Id()),
		}

		_, err := conn.CreateLoggingConfigurationWithContext(ctx, input)

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

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	ws, err := FindWorkspaceByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Prometheus Workspace (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Workspace (%s): %s", d.Id(), err)
	}

	d.Set("alias", ws.Alias)
	arn := aws.StringValue(ws.Arn)
	d.Set("arn", arn)
	d.Set("prometheus_endpoint", ws.PrometheusEndpoint)

	loggingConfiguration, err := FindLoggingConfigurationByWorkspaceID(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		d.Set("logging_configuration", nil)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
	} else {
		if err := d.Set("logging_configuration", []interface{}{flattenLoggingConfigurationMetadata(loggingConfiguration)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting logging_configuration: %s", err)
		}
	}

	return diags
}

func resourceWorkspaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	if d.HasChange("alias") {
		input := &prometheusservice.UpdateWorkspaceAliasInput{
			Alias:       aws.String(d.Get("alias").(string)),
			WorkspaceId: aws.String(d.Id()),
		}

		_, err := conn.UpdateWorkspaceAliasWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Prometheus Workspace alias (%s): %s", d.Id(), err)
		}

		if _, err := waitWorkspaceUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("logging_configuration") {
		if v, ok := d.GetOk("logging_configuration"); ok && len(v.([]interface{})) > 0 && v.([]interface{})[0] != nil {
			tfMap := v.([]interface{})[0].(map[string]interface{})

			if o, _ := d.GetChange("logging_configuration"); o == nil || len(o.([]interface{})) == 0 || o.([]interface{})[0] == nil {
				input := &prometheusservice.CreateLoggingConfigurationInput{
					LogGroupArn: aws.String(tfMap["log_group_arn"].(string)),
					WorkspaceId: aws.String(d.Id()),
				}

				if _, err := conn.CreateLoggingConfigurationWithContext(ctx, input); err != nil {
					return sdkdiag.AppendErrorf(diags, "creating Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
				}

				if _, err := waitLoggingConfigurationCreated(ctx, conn, d.Id()); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) logging configuration create: %s", d.Id(), err)
				}
			} else {
				input := &prometheusservice.UpdateLoggingConfigurationInput{
					LogGroupArn: aws.String(tfMap["log_group_arn"].(string)),
					WorkspaceId: aws.String(d.Id()),
				}

				if _, err := conn.UpdateLoggingConfigurationWithContext(ctx, input); err != nil {
					return sdkdiag.AppendErrorf(diags, "updating Prometheus Workspace (%s) logging configuration: %s", d.Id(), err)
				}

				if _, err := waitLoggingConfigurationUpdated(ctx, conn, d.Id()); err != nil {
					return sdkdiag.AppendErrorf(diags, "waiting for Prometheus Workspace (%s) logging configuration update: %s", d.Id(), err)
				}
			}
		} else {
			_, err := conn.DeleteLoggingConfigurationWithContext(ctx, &prometheusservice.DeleteLoggingConfigurationInput{
				WorkspaceId: aws.String(d.Id()),
			})

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

	conn := meta.(*conns.AWSClient).AMPConn(ctx)

	log.Printf("[INFO] Deleting Prometheus Workspace: %s", d.Id())
	_, err := conn.DeleteWorkspaceWithContext(ctx, &prometheusservice.DeleteWorkspaceInput{
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, prometheusservice.ErrCodeResourceNotFoundException) {
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

func flattenLoggingConfigurationMetadata(apiObject *prometheusservice.LoggingConfigurationMetadata) map[string]interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.LogGroupArn; v != nil {
		tfMap["log_group_arn"] = aws.StringValue(v)
	}

	return tfMap
}
