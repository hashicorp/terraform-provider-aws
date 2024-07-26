// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_notebook_instance_lifecycle_configuration")
func ResourceNotebookInstanceLifeCycleConfiguration() *schema.Resource {
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

func resourceNotebookInstanceLifeCycleConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

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
		hook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		createOpts.OnCreate = []*sagemaker.NotebookInstanceLifecycleHook{hook}
	}

	if v, ok := d.GetOk("on_start"); ok {
		hook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		createOpts.OnStart = []*sagemaker.NotebookInstanceLifecycleHook{hook}
	}

	log.Printf("[DEBUG] SageMaker notebook instance lifecycle configuration create config: %#v", *createOpts)
	_, err := conn.CreateNotebookInstanceLifecycleConfigWithContext(ctx, createOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker notebook instance lifecycle configuration: %s", err)
	}
	d.SetId(name)

	return append(diags, resourceNotebookInstanceLifeCycleConfigurationRead(ctx, d, meta)...)
}

func resourceNotebookInstanceLifeCycleConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	request := &sagemaker.DescribeNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Id()),
	}

	lifecycleConfig, err := conn.DescribeNotebookInstanceLifecycleConfigWithContext(ctx, request)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, "ValidationException") {
			log.Printf("[INFO] unable to find the SageMaker notebook instance lifecycle configuration (%s); therefore it is removed from the state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker notebook instance lifecycle configuration %s: %s", d.Id(), err)
	}

	if err := d.Set(names.AttrName, lifecycleConfig.NotebookInstanceLifecycleConfigName); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting name for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
	}

	if len(lifecycleConfig.OnCreate) > 0 && lifecycleConfig.OnCreate[0] != nil {
		if err := d.Set("on_create", lifecycleConfig.OnCreate[0].Content); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting on_create for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
		}
	}

	if len(lifecycleConfig.OnStart) > 0 && lifecycleConfig.OnStart[0] != nil {
		if err := d.Set("on_start", lifecycleConfig.OnStart[0].Content); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting on_start for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
		}
	}

	if err := d.Set(names.AttrARN, lifecycleConfig.NotebookInstanceLifecycleConfigArn); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting arn for SageMaker notebook instance lifecycle configuration (%s): %s", d.Id(), err)
	}

	return diags
}

func resourceNotebookInstanceLifeCycleConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	updateOpts := &sagemaker.UpdateNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("on_create"); ok {
		onCreateHook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		updateOpts.OnCreate = []*sagemaker.NotebookInstanceLifecycleHook{onCreateHook}
	}

	if v, ok := d.GetOk("on_start"); ok {
		onStartHook := &sagemaker.NotebookInstanceLifecycleHook{Content: aws.String(v.(string))}
		updateOpts.OnStart = []*sagemaker.NotebookInstanceLifecycleHook{onStartHook}
	}

	_, err := conn.UpdateNotebookInstanceLifecycleConfigWithContext(ctx, updateOpts)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating SageMaker Notebook Instance Lifecycle Configuration: %s", err)
	}
	return append(diags, resourceNotebookInstanceLifeCycleConfigurationRead(ctx, d, meta)...)
}

func resourceNotebookInstanceLifeCycleConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	deleteOpts := &sagemaker.DeleteNotebookInstanceLifecycleConfigInput{
		NotebookInstanceLifecycleConfigName: aws.String(d.Id()),
	}
	log.Printf("[INFO] Deleting SageMaker Notebook Instance Lifecycle Configuration: %s", d.Id())

	_, err := conn.DeleteNotebookInstanceLifecycleConfigWithContext(ctx, deleteOpts)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, "ValidationException") {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Notebook Instance Lifecycle Configuration: %s", err)
	}
	return diags
}
