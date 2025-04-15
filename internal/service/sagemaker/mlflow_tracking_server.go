// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
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

// @SDKResource("aws_sagemaker_mlflow_tracking_server", name="Mlflow Tracking Server")
// @Tags(identifierAttribute="arn")
func resourceMlflowTrackingServer() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMlflowTrackingServerCreate,
		ReadWithoutTimeout:   resourceMlflowTrackingServerRead,
		UpdateWithoutTimeout: resourceMlflowTrackingServerUpdate,
		DeleteWithoutTimeout: resourceMlflowTrackingServerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"artifact_store_uri": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validModelDataURL,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"tracking_server_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mlflow_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"tracking_server_url": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"automatic_model_registration": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"tracking_server_size": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.TrackingServerSizeS,
				ValidateDiagFunc: enum.Validate[awstypes.TrackingServerSize](),
			},
			"weekly_maintenance_window_start": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceMlflowTrackingServerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("tracking_server_name").(string)
	input := &sagemaker.CreateMlflowTrackingServerInput{
		TrackingServerName:         aws.String(name),
		ArtifactStoreUri:           aws.String(d.Get("artifact_store_uri").(string)),
		RoleArn:                    aws.String(d.Get(names.AttrRoleARN).(string)),
		AutomaticModelRegistration: aws.Bool(d.Get("automatic_model_registration").(bool)),
		TrackingServerSize:         awstypes.TrackingServerSize(d.Get("tracking_server_size").(string)),
		Tags:                       getTagsIn(ctx),
	}

	if v, ok := d.GetOk("mlflow_version"); ok {
		input.MlflowVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("weekly_maintenance_window_start"); ok {
		input.WeeklyMaintenanceWindowStart = aws.String(v.(string))
	}

	_, err := conn.CreateMlflowTrackingServer(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Mlflow Tracking Server %s: %s", name, err)
	}

	d.SetId(name)

	if _, err := waitMlflowTrackingServerCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Mlflow Tracking Server (%s) to delete: %s", d.Id(), err)
	}

	return append(diags, resourceMlflowTrackingServerRead(ctx, d, meta)...)
}

func resourceMlflowTrackingServerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	output, err := findMlflowTrackingServerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI Mlflow Tracking Server (%s); removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Mlflow Tracking Server (%s): %s", d.Id(), err)
	}

	d.Set("tracking_server_name", output.TrackingServerName)
	d.Set(names.AttrARN, output.TrackingServerArn)
	d.Set("artifact_store_uri", output.ArtifactStoreUri)
	d.Set(names.AttrRoleARN, output.RoleArn)
	d.Set("mlflow_version", output.MlflowVersion)
	d.Set("tracking_server_size", output.TrackingServerSize)
	d.Set("weekly_maintenance_window_start", output.WeeklyMaintenanceWindowStart)
	d.Set("tracking_server_url", output.TrackingServerUrl)
	d.Set("automatic_model_registration", output.AutomaticModelRegistration)

	return diags
}

func resourceMlflowTrackingServerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateMlflowTrackingServerInput{
			TrackingServerName: aws.String(d.Id()),
		}

		if d.HasChange("artifact_store_uri") {
			if v, ok := d.GetOk("artifact_store_uri"); ok {
				input.ArtifactStoreUri = aws.String(v.(string))
			}
		}

		if d.HasChange("automatic_model_registration") {
			if v, ok := d.GetOk("automatic_model_registration"); ok {
				input.AutomaticModelRegistration = aws.Bool(v.(bool))
			}
		}

		if d.HasChange("tracking_server_size") {
			if v, ok := d.GetOk("tracking_server_size"); ok {
				input.TrackingServerSize = awstypes.TrackingServerSize(v.(string))
			}
		}

		if d.HasChange("weekly_maintenance_window_start") {
			if v, ok := d.GetOk("weekly_maintenance_window_start"); ok {
				input.WeeklyMaintenanceWindowStart = aws.String(v.(string))
			}
		}

		log.Printf("[DEBUG] SageMaker AI Mlflow Tracking Server update config: %#v", *input)
		_, err := conn.UpdateMlflowTrackingServer(ctx, input)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Mlflow Tracking Server: %s", err)
		}

		if _, err := waitMlflowTrackingServerUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Mlflow Tracking Server (%s) to update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMlflowTrackingServerRead(ctx, d, meta)...)
}

func resourceMlflowTrackingServerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	input := &sagemaker.DeleteMlflowTrackingServerInput{
		TrackingServerName: aws.String(d.Id()),
	}

	if _, err := conn.DeleteMlflowTrackingServer(ctx, input); err != nil {
		if errs.IsA[*awstypes.ResourceNotFound](err) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Mlflow Tracking Server (%s): %s", d.Id(), err)
	}

	if _, err := waitMlflowTrackingServerDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Mlflow Tracking Server (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findMlflowTrackingServerByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeMlflowTrackingServerOutput, error) {
	input := &sagemaker.DescribeMlflowTrackingServerInput{
		TrackingServerName: aws.String(name),
	}

	output, err := conn.DescribeMlflowTrackingServer(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
