// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dataexchange

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_dataexchange_event_action", name="Event Action")
func ResourceEventAction() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEventActionCreate,
		ReadWithoutTimeout:   resourceEventActionRead,
		UpdateWithoutTimeout: resourceEventActionUpdate,
		DeleteWithoutTimeout: resourceEventActionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedAt: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrLastUpdatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"action_export_revision_to_s3": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"bucket": {
							Type:     schema.TypeString,
							Required: true,
						},
						"key_pattern": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "${Revision.CreatedAt}/${Asset.Name}",
						},
						"s3_encryption_kms_key_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: verify.ValidARN,
						},
						"s3_encryption_type": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ServerSideEncryptionTypes](),
						},
					},
				},
			},
			"event_revision_published": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"data_set_id": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

// CRUD
func resourceEventActionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	out, err := conn.CreateEventAction(ctx, buildCreateEventActionInput(d))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating DataExchange EventAction: %s", err)
	}

	d.SetId(aws.ToString(out.Id))

	return append(diags, resourceEventActionRead(ctx, d, meta)...)
}

func resourceEventActionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	output, err := conn.GetEventAction(ctx, &dataexchange.GetEventActionInput{
		EventActionId: aws.String(d.Get(names.AttrID).(string)),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading DataExchange EventAction (%s): %s", d.Id(), err)
	}

	buildEventActionAttr(output, d)

	return diags
}

func resourceEventActionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	if d.HasChange("event_revision_published.data_set_id") {
		resourceEventActionDelete(ctx, d, meta)
		return resourceJobCreate(ctx, d, meta)
	}

	_, err := conn.UpdateEventAction(ctx, buildUpdateEventActionInput(d))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating DataExchange EventAction (%s): %s", d.Id(), err)
	}
	return append(diags, resourceEventActionRead(ctx, d, meta)...)
}

func resourceEventActionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	_, err := conn.DeleteEventAction(ctx, &dataexchange.DeleteEventActionInput{
		EventActionId: aws.String(d.Get(names.AttrID).(string)),
	})
	if err != nil {
		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting DataExchange EventAction: %s", err)
	}

	return diags
}

// Builders
func buildCreateEventActionInput(d *schema.ResourceData) *dataexchange.CreateEventActionInput {
	res := &dataexchange.CreateEventActionInput{
		Action: &awstypes.Action{},
		Event:  &awstypes.Event{},
	}

	if _, ok := d.GetOk("action_export_revision_to_s3"); ok {
		m := d.Get("action_export_revision_to_s3").([]interface{})[0].(map[string]interface{})
		res.Action.ExportRevisionToS3 = &awstypes.AutoExportRevisionToS3RequestDetails{
			RevisionDestination: &awstypes.AutoExportRevisionDestinationEntry{
				Bucket: aws.String(m["bucket"].(string)),
			},
		}

		if keyPattern, ok := m["key_pattern"]; ok && keyPattern != "" {
			res.Action.ExportRevisionToS3.RevisionDestination.KeyPattern = aws.String(keyPattern.(string))
		}

		if encType, ok := m["s3_encryption_type"]; ok && encType != "" {
			res.Action.ExportRevisionToS3.Encryption = &awstypes.ExportServerSideEncryption{
				Type: awstypes.ServerSideEncryptionTypes(encType.(string)),
			}

			if keyArn, ok := m["s3_encryption_kms_key_arn"]; ok && keyArn != "" {
				res.Action.ExportRevisionToS3.Encryption.KmsKeyArn = aws.String(keyArn.(string))
			}
		}
	}

	if _, ok := d.GetOk("event_revision_published"); ok {
		m := d.Get("event_revision_published").([]interface{})[0].(map[string]interface{})
		res.Event.RevisionPublished = &awstypes.RevisionPublished{
			DataSetId: aws.String(m["data_set_id"].(string)),
		}
	}

	return res
}

func buildUpdateEventActionInput(d *schema.ResourceData) *dataexchange.UpdateEventActionInput {
	res := &dataexchange.UpdateEventActionInput{
		EventActionId: aws.String(d.Get(names.AttrID).(string)),
		Action:        &awstypes.Action{},
	}

	if _, ok := d.GetOk("action_export_revision_to_s3"); ok {
		m := d.Get("action_export_revision_to_s3").([]interface{})[0].(map[string]interface{})
		res.Action.ExportRevisionToS3 = &awstypes.AutoExportRevisionToS3RequestDetails{
			RevisionDestination: &awstypes.AutoExportRevisionDestinationEntry{
				Bucket: aws.String(m["bucket"].(string)),
			},
		}

		if keyPattern, ok := m["key_pattern"]; ok && keyPattern != "" {
			res.Action.ExportRevisionToS3.RevisionDestination.KeyPattern = aws.String(keyPattern.(string))
		}

		if encType, ok := m["s3_encryption_type"]; ok && encType != "" {
			res.Action.ExportRevisionToS3.Encryption = &awstypes.ExportServerSideEncryption{
				Type: awstypes.ServerSideEncryptionTypes(encType.(string)),
			}

			if keyArn, ok := m["s3_encryption_kms_key_arn"]; ok && keyArn != "" {
				res.Action.ExportRevisionToS3.Encryption.KmsKeyArn = aws.String(keyArn.(string))
			}
		}
	}

	return res
}

func buildEventActionAttr(out *dataexchange.GetEventActionOutput, d *schema.ResourceData) {
	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrCreatedAt, out.CreatedAt.String())
	d.Set(names.AttrLastUpdatedTime, out.UpdatedAt.String())

	if out.Action != nil && out.Action.ExportRevisionToS3 != nil {
		actionM := map[string]any{}
		if out.Action.ExportRevisionToS3.RevisionDestination != nil {
			actionM["bucket"] = out.Action.ExportRevisionToS3.RevisionDestination.Bucket
			actionM["key_pattern"] = out.Action.ExportRevisionToS3.RevisionDestination.KeyPattern
		}

		if out.Action.ExportRevisionToS3.Encryption != nil {
			actionM["s3_encryption_type"] = out.Action.ExportRevisionToS3.Encryption.Type
			actionM["s3_encryption_kms_key_arn"] = out.Action.ExportRevisionToS3.Encryption.KmsKeyArn
		}

		d.Set("action_export_revision_to_s3", []interface{}{actionM})
	}

	if out.Event != nil && out.Event.RevisionPublished != nil {
		eventM := map[string]any{
			"data_set_id": out.Event.RevisionPublished.DataSetId,
		}
		d.Set("event_revision_published", []interface{}{eventM})
	}
}
