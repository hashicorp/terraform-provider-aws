// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_notebook_instance_lifecycle_configuration", name="Notebook Instance Lifecycle Configuration")
func resourceNotebookInstanceLifeCycleConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNotebookInstanceLifeCycleConfigurationCreate,
		ReadWithoutTimeout:   resourceNotebookInstanceLifeCycleConfigurationRead,
		UpdateWithoutTimeout: resourceNotebookInstanceLifeCycleConfigurationUpdate,
		DeleteWithoutTimeout: resourceNotebookInstanceLifeCycleConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},

			names.AttrName: {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},

			"on_create": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},

			"on_start": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 16384),
			},
		},
	}
}

func resourceNotebookInstanceLifeCycleConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	var name string
	if v, ok := d.GetOk(names.AttrName); ok {
		name = v.(string)
	} else {
		name = id.UniqueId()
	}

	createOpts := &sagemaker.CreateNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(name),
	}

	// on_create is technically a list of NotebookInstanceLifecycleHook elements, but the list has to be length 1
	// (same for on_start)
	if v, ok := d.GetOk("on_create"); ok {
		hook := awstypes.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		createOpts.OnCreate = []awstypes.NotebookInstanceLifecycleHook{hook}
	}

	if v, ok := d.GetOk("on_start"); ok {
		hook := awstypes.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		createOpts.OnStart = []awstypes.NotebookInstanceLifecycleHook{hook}
	}

	log.Printf("[DEBUG] SageMaker AI notebook instance lifecycle configuration create config: %#v", *createOpts)
	_, err := conn.CreateNotebookInstanceLifecycleConfig(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI notebook instance lifecycle configuration: %s", err)
	}
	d.SetId(name)

	return append(diags, resourceNotebookInstanceLifeCycleConfigurationRead(ctx, d, meta)...)
}

func resourceNotebookInstanceLifeCycleConfigurationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	output, err := findNotebookInstanceLifecycleConfigByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[INFO] unable to find the SageMaker AI notebook instance lifecycle configuration (%s); therefore it is removed from the state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI notebook instance lifecycle configuration %s: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrName, output.NotebookInstanceLifecycleConfigName); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting name for SageMaker AI notebook instance lifecycle configuration (%s): %s", d.Id(), err)
	}

	if len(output.OnCreate) > 0 {
		if err := d.Set("on_create", output.OnCreate[0].Content); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting on_create for SageMaker AI notebook instance lifecycle configuration (%s): %s", d.Id(), err)
		}
	}

	if len(output.OnStart) > 0 {
		if err := d.Set("on_start", output.OnStart[0].Content); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting on_start for SageMaker AI notebook instance lifecycle configuration (%s): %s", d.Id(), err)
		}
	}

	if err := d.Set(names.AttrARN, output.NotebookInstanceLifecycleConfigArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arn for SageMaker AI notebook instance lifecycle configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceNotebookInstanceLifeCycleConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	updateOpts := &sagemaker.UpdateNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("on_create"); ok {
		onCreateHook := awstypes.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		updateOpts.OnCreate = []awstypes.NotebookInstanceLifecycleHook{onCreateHook}
	}

	if v, ok := d.GetOk("on_start"); ok {
		onStartHook := awstypes.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		updateOpts.OnStart = []awstypes.NotebookInstanceLifecycleHook{onStartHook}
	}

	_, err := conn.UpdateNotebookInstanceLifecycleConfig(ctx, updateOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Notebook Instance Lifecycle Configuration: %s", err)
	}
	return append(diags, resourceNotebookInstanceLifeCycleConfigurationRead(ctx, d, meta)...)
}

func resourceNotebookInstanceLifeCycleConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	deleteOpts := &sagemaker.DeleteNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker AI Notebook Instance Lifecycle Configuration: %s", d.Id())

	_, err := conn.DeleteNotebookInstanceLifecycleConfig(ctx, deleteOpts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ErrCodeValidationException) {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Notebook Instance Lifecycle Configuration: %s", err)
	}
	return diags
}

func findNotebookInstanceLifecycleConfigByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeNotebookInstanceLifecycleConfigOutput, error) {
	input := &sagemaker.DescribeNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(name),
	}

	output, err := conn.DescribeNotebookInstanceLifecycleConfig(ctx, input)

	if tfawserr.ErrCodeEquals(err, ErrCodeValidationException) {
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
