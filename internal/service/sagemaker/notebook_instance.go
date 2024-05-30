// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_notebook_instance", name="Notebook Instance")
// @Tags(identifierAttribute="arn")
func ResourceNotebookInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNotebookInstanceCreate,
		ReadWithoutTimeout:   resourceNotebookInstanceRead,
		UpdateWithoutTimeout: resourceNotebookInstanceUpdate,
		DeleteWithoutTimeout: resourceNotebookInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange(names.AttrVolumeSize, func(_ context.Context, old, new, meta interface{}) bool {
				return new.(int) < old.(int)
			}),
			verify.SetTagsDiff,
		),

		Schema: map[string]*schema.Schema{
			"accelerator_types": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type:         schema.TypeString,
					ValidateFunc: validation.StringInSlice(sagemaker.NotebookInstanceAcceleratorType_Values(), false),
				},
			},
			"additional_code_repositories": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 3,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_code_repository": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"direct_internet_access": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      sagemaker.DirectInternetAccessEnabled,
				ValidateFunc: validation.StringInSlice(sagemaker.DirectInternetAccess_Values(), false),
			},
			"instance_metadata_service_configuration": {
				Type:             schema.TypeList,
				Optional:         true,
				DiffSuppressFunc: verify.SuppressMissingOptionalConfigurationBlock,
				MaxItems:         1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"minimum_instance_metadata_service_version": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice([]string{"1", "2"}, false),
						},
					},
				},
			},
			names.AttrInstanceType: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(sagemaker.InstanceType_Values(), false),
			},
			names.AttrKMSKeyID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"lifecycle_config_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			names.AttrName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validName,
			},
			names.AttrNetworkInterfaceID: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"platform_identifier": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^(notebook-al1-v1|notebook-al2-v1|notebook-al2-v2)$`), ""),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"root_access": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      sagemaker.RootAccessEnabled,
				ValidateFunc: validation.StringInSlice(sagemaker.RootAccess_Values(), false),
			},
			names.AttrSecurityGroups: {
				Type:     schema.TypeSet,
				MinItems: 1,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrSubnetID: {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrURL: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrVolumeSize: {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  5,
			},
		},
	}
}

func resourceNotebookInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get(names.AttrName).(string)
	input := &sagemaker.CreateNotebookInstanceInput{
		InstanceMetadataServiceConfiguration: expandNotebookInstanceMetadataServiceConfiguration(d.Get("instance_metadata_service_configuration").([]interface{})),
		InstanceType:                         aws.String(d.Get(names.AttrInstanceType).(string)),
		NotebookInstanceName:                 aws.String(name),
		RoleArn:                              aws.String(d.Get(names.AttrRoleARN).(string)),
		SecurityGroupIds:                     flex.ExpandStringSet(d.Get(names.AttrSecurityGroups).(*schema.Set)),
		Tags:                                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("accelerator_types"); ok && v.(*schema.Set).Len() > 0 {
		input.AcceleratorTypes = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("additional_code_repositories"); ok && v.(*schema.Set).Len() > 0 {
		input.AdditionalCodeRepositories = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("default_code_repository"); ok {
		input.DefaultCodeRepository = aws.String(v.(string))
	}

	if v, ok := d.GetOk("direct_internet_access"); ok {
		input.DirectInternetAccess = aws.String(v.(string))
	}

	if k, ok := d.GetOk(names.AttrKMSKeyID); ok {
		input.KmsKeyId = aws.String(k.(string))
	}

	if l, ok := d.GetOk("lifecycle_config_name"); ok {
		input.LifecycleConfigName = aws.String(l.(string))
	}

	if v, ok := d.GetOk("platform_identifier"); ok {
		input.PlatformIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("root_access"); ok {
		input.RootAccess = aws.String(v.(string))
	}

	if s, ok := d.GetOk(names.AttrSubnetID); ok {
		input.SubnetId = aws.String(s.(string))
	}

	if v, ok := d.GetOk(names.AttrVolumeSize); ok {
		input.VolumeSizeInGB = aws.Int64(int64(v.(int)))
	}

	log.Printf("[DEBUG] Creating SageMaker Notebook Instance: %s", input)
	_, err := conn.CreateNotebookInstanceWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker Notebook Instance: %s", err)
	}

	d.SetId(name)

	if _, err := WaitNotebookInstanceInService(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Notebook Instance (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceNotebookInstanceRead(ctx, d, meta)...)
}

func resourceNotebookInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	notebookInstance, err := FindNotebookInstanceByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] SageMaker Notebook Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	d.Set("accelerator_types", aws.StringValueSlice(notebookInstance.AcceleratorTypes))
	d.Set("additional_code_repositories", aws.StringValueSlice(notebookInstance.AdditionalCodeRepositories))
	d.Set(names.AttrARN, notebookInstance.NotebookInstanceArn)
	d.Set("default_code_repository", notebookInstance.DefaultCodeRepository)
	d.Set("direct_internet_access", notebookInstance.DirectInternetAccess)
	d.Set(names.AttrInstanceType, notebookInstance.InstanceType)
	d.Set(names.AttrKMSKeyID, notebookInstance.KmsKeyId)
	d.Set("lifecycle_config_name", notebookInstance.NotebookInstanceLifecycleConfigName)
	d.Set(names.AttrName, notebookInstance.NotebookInstanceName)
	d.Set(names.AttrNetworkInterfaceID, notebookInstance.NetworkInterfaceId)
	d.Set("platform_identifier", notebookInstance.PlatformIdentifier)
	d.Set(names.AttrRoleARN, notebookInstance.RoleArn)
	d.Set("root_access", notebookInstance.RootAccess)
	d.Set(names.AttrSecurityGroups, aws.StringValueSlice(notebookInstance.SecurityGroups))
	d.Set(names.AttrSubnetID, notebookInstance.SubnetId)
	d.Set(names.AttrURL, notebookInstance.Url)
	d.Set(names.AttrVolumeSize, notebookInstance.VolumeSizeInGB)

	if err := d.Set("instance_metadata_service_configuration", flattenNotebookInstanceMetadataServiceConfiguration(notebookInstance.InstanceMetadataServiceConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_metadata_service_configuration: %s", err)
	}

	return diags
}

func resourceNotebookInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateNotebookInstanceInput{
			NotebookInstanceName: aws.String(d.Get(names.AttrName).(string)),
		}

		if d.HasChange("accelerator_types") {
			if v, ok := d.GetOk("accelerator_types"); ok {
				input.AcceleratorTypes = flex.ExpandStringSet(v.(*schema.Set))
			} else {
				input.DisassociateAcceleratorTypes = aws.Bool(true)
			}
		}

		if d.HasChange("additional_code_repositories") {
			if v, ok := d.GetOk("additional_code_repositories"); ok {
				input.AdditionalCodeRepositories = flex.ExpandStringSet(v.(*schema.Set))
			} else {
				input.DisassociateAdditionalCodeRepositories = aws.Bool(true)
			}
		}

		if d.HasChange("default_code_repository") {
			if v, ok := d.GetOk("default_code_repository"); ok {
				input.DefaultCodeRepository = aws.String(v.(string))
			} else {
				input.DisassociateDefaultCodeRepository = aws.Bool(true)
			}
		}

		if d.HasChange("instance_metadata_service_configuration") {
			input.InstanceMetadataServiceConfiguration = expandNotebookInstanceMetadataServiceConfiguration(d.Get("instance_metadata_service_configuration").([]interface{}))
		}

		if d.HasChange(names.AttrInstanceType) {
			input.InstanceType = aws.String(d.Get(names.AttrInstanceType).(string))
		}

		if d.HasChange("lifecycle_config_name") {
			if v, ok := d.GetOk("lifecycle_config_name"); ok {
				input.LifecycleConfigName = aws.String(v.(string))
			} else {
				input.DisassociateLifecycleConfig = aws.Bool(true)
			}
		}

		if d.HasChange(names.AttrRoleARN) {
			input.RoleArn = aws.String(d.Get(names.AttrRoleARN).(string))
		}

		if d.HasChange("root_access") {
			input.RootAccess = aws.String(d.Get("root_access").(string))
		}

		if d.HasChange(names.AttrVolumeSize) {
			input.VolumeSizeInGB = aws.Int64(int64(d.Get(names.AttrVolumeSize).(int)))
		}

		// Stop notebook.
		notebook, err := FindNotebookInstanceByName(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Notebook Instance (%s): %s", d.Id(), err)
		}

		previousStatus := aws.StringValue(notebook.NotebookInstanceStatus)

		if previousStatus != sagemaker.NotebookInstanceStatusStopped {
			if err := StopNotebookInstance(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating SageMaker Notebook Instance (%s): %s", d.Id(), err)
			}
		}

		if _, err := conn.UpdateNotebookInstanceWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Notebook Instance (%s): %s", d.Id(), err)
		}

		if _, err := WaitNotebookInstanceStopped(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Notebook Instance (%s) to stop: %s", d.Id(), err)
		}

		// Restart if needed
		if previousStatus == sagemaker.NotebookInstanceStatusInService {
			if err := StartNotebookInstance(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating SageMaker Notebook Instance (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceNotebookInstanceRead(ctx, d, meta)...)
}

func resourceNotebookInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	notebook, err := FindNotebookInstanceByName(ctx, conn, d.Id())

	if tfresource.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	switch status := aws.StringValue(notebook.NotebookInstanceStatus); status {
	case sagemaker.NotebookInstanceStatusInService:
		if err := StopNotebookInstance(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "stopping SageMaker Notebook Instance (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting SageMaker Notebook Instance: %s", d.Id())
	_, err = conn.DeleteNotebookInstanceWithContext(ctx, &sagemaker.DeleteNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Notebook Instance (%s): %s", d.Id(), err)
	}

	if _, err := WaitNotebookInstanceDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Notebook Instance (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func StartNotebookInstance(ctx context.Context, conn *sagemaker.SageMaker, id string) error {
	startOpts := &sagemaker.StartNotebookInstanceInput{
		NotebookInstanceName: aws.String(id),
	}
	// StartNotebookInstance sometimes doesn't take so we'll check for a state change and if
	// it doesn't change we'll send another request
	err := retry.RetryContext(ctx, 5*time.Minute, func() *retry.RetryError {
		_, err := conn.StartNotebookInstanceWithContext(ctx, startOpts)
		if err != nil {
			return retry.NonRetryableError(fmt.Errorf("starting: %s", err))
		}

		_, err = WaitNotebookInstanceStarted(ctx, conn, id)
		if err != nil {
			return retry.RetryableError(fmt.Errorf("starting: waiting for completion: %s", err))
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.StartNotebookInstanceWithContext(ctx, startOpts)
		if err != nil {
			return fmt.Errorf("starting: %s", err)
		}

		_, err = WaitNotebookInstanceStarted(ctx, conn, id)
		if err != nil {
			return fmt.Errorf("starting: waiting for completion: %s", err)
		}
	}

	if _, err := WaitNotebookInstanceInService(ctx, conn, id); err != nil {
		return fmt.Errorf("starting: waiting to be in service: %s", err)
	}
	return nil
}

func StopNotebookInstance(ctx context.Context, conn *sagemaker.SageMaker, id string) error {
	notebook, err := FindNotebookInstanceByName(ctx, conn, id)

	if err != nil {
		if tfresource.NotFound(err) {
			return nil
		}
		return err
	}

	if aws.StringValue(notebook.NotebookInstanceStatus) == sagemaker.NotebookInstanceStatusStopped {
		return nil
	}

	stopOpts := &sagemaker.StopNotebookInstanceInput{
		NotebookInstanceName: aws.String(id),
	}

	if _, err := conn.StopNotebookInstanceWithContext(ctx, stopOpts); err != nil {
		return fmt.Errorf("stopping: %s", err)
	}

	if _, err := WaitNotebookInstanceStopped(ctx, conn, id); err != nil {
		return fmt.Errorf("stopping: waiting for completion: %s", err)
	}

	return nil
}

func expandNotebookInstanceMetadataServiceConfiguration(l []interface{}) *sagemaker.InstanceMetadataServiceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	config := &sagemaker.InstanceMetadataServiceConfiguration{
		MinimumInstanceMetadataServiceVersion: aws.String(m["minimum_instance_metadata_service_version"].(string)),
	}

	return config
}

func flattenNotebookInstanceMetadataServiceConfiguration(config *sagemaker.InstanceMetadataServiceConfiguration) []map[string]interface{} {
	if config == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"minimum_instance_metadata_service_version": aws.StringValue(config.MinimumInstanceMetadataServiceVersion),
	}

	return []map[string]interface{}{m}
}
