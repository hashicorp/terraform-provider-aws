// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_mlflow_app", name="Mlflow App")
// @Tags(identifierAttribute="arn")
func resourceMlflowApp() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMlflowAppCreate,
		ReadWithoutTimeout:   resourceMlflowAppRead,
		UpdateWithoutTimeout: resourceMlflowAppUpdate,
		DeleteWithoutTimeout: resourceMlflowAppDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_default_status": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateDiagFunc: enum.Validate[awstypes.AccountDefaultStatus](),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"artifact_store_uri": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validModelDataURL,
			},
			"default_domain_id_list": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"model_registration_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          "AutoModelRegistrationDisabled",
				ValidateDiagFunc: enum.Validate[awstypes.ModelRegistrationMode](),
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			"weekly_maintenance_window_start": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceMlflowAppCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get(names.AttrName).(string)
	log.Printf("[DEBUG] Creating SageMaker Mlflow App: %s", name)

	input := &sagemaker.CreateMlflowAppInput{
		ArtifactStoreUri: aws.String(d.Get("artifact_store_uri").(string)),
		Name:             aws.String(name),
		RoleArn:          aws.String(d.Get(names.AttrRoleARN).(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("account_default_status"); ok {
		input.AccountDefaultStatus = awstypes.AccountDefaultStatus(v.(string))
		log.Printf("[DEBUG] Setting account_default_status: %s", v.(string))
	}

	if v, ok := d.GetOk("default_domain_id_list"); ok {
		input.DefaultDomainIdList = flex.ExpandStringValueSet(v.(*schema.Set))
		log.Printf("[DEBUG] Setting default_domain_id_list: %v", input.DefaultDomainIdList)
	}

	if v, ok := d.GetOk("model_registration_mode"); ok {
		input.ModelRegistrationMode = awstypes.ModelRegistrationMode(v.(string))
		log.Printf("[DEBUG] Setting model_registration_mode: %s", v.(string))
	}

	if v, ok := d.GetOk("weekly_maintenance_window_start"); ok {
		input.WeeklyMaintenanceWindowStart = aws.String(v.(string))
		log.Printf("[DEBUG] Setting weekly_maintenance_window_start: %s", v.(string))
	}

	log.Printf("[DEBUG] SageMaker Mlflow App create config: %#v", *input)
	output, err := conn.CreateMlflowApp(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Mlflow App (%s): %s", name, err)
	}

	log.Printf("[DEBUG] Created SageMaker Mlflow App with ARN: %s", aws.ToString(output.Arn))
	d.SetId(aws.ToString(output.Arn))

	// Wait for the MLflow App to be created
	log.Printf("[DEBUG] Waiting for SageMaker Mlflow App (%s) to be created", d.Id())
	if _, err := waitMlflowAppCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Mlflow App (%s) create: %s", d.Id(), err)
	}
	log.Printf("[DEBUG] SageMaker Mlflow App (%s) creation completed", d.Id())

	return append(diags, resourceMlflowAppRead(ctx, d, meta)...)
}

func resourceMlflowAppRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[DEBUG] Reading SageMaker Mlflow App: %s", d.Id())
	output, err := findMlflowAppByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SageMaker Mlflow App (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Mlflow App (%s): %s", d.Id(), err)
	}

	// If the MLflow App is deleted, remove from state
	if !d.IsNewResource() && output.Status == "Deleted" {
		log.Printf("[WARN] SageMaker Mlflow App (%s) is deleted, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	log.Printf("[DEBUG] SageMaker Mlflow App status: %s", output.Status)
	d.Set("account_default_status", output.AccountDefaultStatus)
	d.Set(names.AttrARN, output.Arn)
	d.Set("artifact_store_uri", output.ArtifactStoreUri)
	d.Set("default_domain_id_list", flex.FlattenStringValueSet(output.DefaultDomainIdList))
	d.Set("model_registration_mode", output.ModelRegistrationMode)
	d.Set(names.AttrName, output.Name)
	d.Set(names.AttrRoleARN, output.RoleArn)
	d.Set("weekly_maintenance_window_start", output.WeeklyMaintenanceWindowStart)

	return diags
}

func resourceMlflowAppUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateMlflowAppInput{
			Arn: aws.String(d.Id()),
		}

		if d.HasChange("account_default_status") {
			input.AccountDefaultStatus = awstypes.AccountDefaultStatus(d.Get("account_default_status").(string))
		}

		if d.HasChange("artifact_store_uri") {
			input.ArtifactStoreUri = aws.String(d.Get("artifact_store_uri").(string))
		}

		if d.HasChange("default_domain_id_list") {
			input.DefaultDomainIdList = flex.ExpandStringValueSet(d.Get("default_domain_id_list").(*schema.Set))
		}

		if d.HasChange("model_registration_mode") {
			input.ModelRegistrationMode = awstypes.ModelRegistrationMode(d.Get("model_registration_mode").(string))
		}

		if d.HasChange("weekly_maintenance_window_start") {
			input.WeeklyMaintenanceWindowStart = aws.String(d.Get("weekly_maintenance_window_start").(string))
		}

		_, err := conn.UpdateMlflowApp(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Mlflow App (%s): %s", d.Id(), err)
		}

		if _, err := waitMlflowAppUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Mlflow App (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceMlflowAppRead(ctx, d, meta)...)
}

func resourceMlflowAppDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[DEBUG] Deleting SageMaker Mlflow App: %s", d.Id())

	// Check current status first
	output, err := findMlflowAppByARN(ctx, conn, d.Id())
	if retry.NotFound(err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Mlflow App (%s) before delete: %s", d.Id(), err)
	}

	// If already deleted or deleting, just wait for completion
	if output.Status == "Deleted" {
		return diags
	}
	if output.Status == "Deleting" {
		if err := waitMlflowAppDeleted(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Mlflow App (%s) delete: %s", d.Id(), err)
		}
		return diags
	}

	_, err = conn.DeleteMlflowApp(ctx, &sagemaker.DeleteMlflowAppInput{
		Arn: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Mlflow App (%s): %s", d.Id(), err)
	}

	if err := waitMlflowAppDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Mlflow App (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findMlflowAppByARN(ctx context.Context, conn *sagemaker.Client, arn string) (*sagemaker.DescribeMlflowAppOutput, error) {
	input := &sagemaker.DescribeMlflowAppInput{
		Arn: aws.String(arn),
	}

	output, err := conn.DescribeMlflowApp(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
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
