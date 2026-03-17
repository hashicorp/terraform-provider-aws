// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/semver"
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
				ValidateFunc: validHTTPSOrS3URI,
			},
			"automatic_model_registration": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"mlflow_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
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
			"tracking_server_size": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.TrackingServerSizeS,
				ValidateDiagFunc: enum.Validate[awstypes.TrackingServerSize](),
			},
			"tracking_server_url": {
				Type:     schema.TypeString,
				Computed: true,
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
	input := sagemaker.CreateMlflowTrackingServerInput{
		ArtifactStoreUri:           aws.String(d.Get("artifact_store_uri").(string)),
		AutomaticModelRegistration: aws.Bool(d.Get("automatic_model_registration").(bool)),
		RoleArn:                    aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:                       getTagsIn(ctx),
		TrackingServerName:         aws.String(name),
		TrackingServerSize:         awstypes.TrackingServerSize(d.Get("tracking_server_size").(string)),
	}

	if v, ok := d.GetOk("mlflow_version"); ok {
		input.MlflowVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("weekly_maintenance_window_start"); ok {
		input.WeeklyMaintenanceWindowStart = aws.String(v.(string))
	}

	_, err := conn.CreateMlflowTrackingServer(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Mlflow Tracking Server (%s): %s", name, err)
	}

	d.SetId(name)

	if _, err := waitMlflowTrackingServerCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Mlflow Tracking Server (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceMlflowTrackingServerRead(ctx, d, meta)...)
}

func resourceMlflowTrackingServerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	output, err := findMlflowTrackingServerByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI Mlflow Tracking Server (%s); removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Mlflow Tracking Server (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrARN, output.TrackingServerArn)
	d.Set("artifact_store_uri", output.ArtifactStoreUri)
	d.Set("automatic_model_registration", output.AutomaticModelRegistration)
	d.Set("mlflow_version", normalizeMlflowVersion(aws.ToString(output.MlflowVersion)))
	d.Set(names.AttrRoleARN, output.RoleArn)
	d.Set("tracking_server_name", output.TrackingServerName)
	d.Set("tracking_server_size", output.TrackingServerSize)
	d.Set("tracking_server_url", output.TrackingServerUrl)
	d.Set("weekly_maintenance_window_start", output.WeeklyMaintenanceWindowStart)

	return diags
}

func resourceMlflowTrackingServerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := sagemaker.UpdateMlflowTrackingServerInput{
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

		_, err := conn.UpdateMlflowTrackingServer(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Mlflow Tracking Server (%s): %s", d.Id(), err)
		}

		if _, err := waitMlflowTrackingServerUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Mlflow Tracking Server (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMlflowTrackingServerRead(ctx, d, meta)...)
}

func resourceMlflowTrackingServerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[DEBUG] Deleting SageMaker AI Mlflow Tracking Server: %s", d.Id())
	input := sagemaker.DeleteMlflowTrackingServerInput{
		TrackingServerName: aws.String(d.Id()),
	}
	_, err := conn.DeleteMlflowTrackingServer(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Mlflow Tracking Server (%s): %s", d.Id(), err)
	}

	if _, err := waitMlflowTrackingServerDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Mlflow Tracking Server (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findMlflowTrackingServerByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeMlflowTrackingServerOutput, error) {
	input := sagemaker.DescribeMlflowTrackingServerInput{
		TrackingServerName: aws.String(name),
	}

	output, err := conn.DescribeMlflowTrackingServer(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

// normalizeMlflowVersion normalizes the MLflow version to major/minor
// format (e.g., "1.26" instead of "1.26.0").
func normalizeMlflowVersion(version string) string {
	s, _ := semver.MajorMinor(version)
	return s
}
