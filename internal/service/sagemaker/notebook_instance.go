// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sagemaker

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_notebook_instance", name="Notebook Instance")
// @Tags(identifierAttribute="arn")
func resourceNotebookInstance() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceNotebookInstanceCreate,
		ReadWithoutTimeout:   resourceNotebookInstanceRead,
		UpdateWithoutTimeout: resourceNotebookInstanceUpdate,
		DeleteWithoutTimeout: resourceNotebookInstanceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: customdiff.Sequence(
			customdiff.ForceNewIfChange(names.AttrVolumeSize, func(_ context.Context, old, new, meta any) bool {
				return new.(int) < old.(int)
			}),
		),

		Schema: map[string]*schema.Schema{
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
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          awstypes.DirectInternetAccessEnabled,
				ValidateDiagFunc: enum.Validate[awstypes.DirectInternetAccess](),
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
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.InstanceType](),
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
				ValidateFunc: validation.StringMatch(regexache.MustCompile(`^(notebook-al1-v1|notebook-al2-v1|notebook-al2-v2|notebook-al2-v3|notebook-al2023-v1)$`), ""),
			},
			names.AttrRoleARN: {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"root_access": {
				Type:             schema.TypeString,
				Optional:         true,
				Default:          awstypes.RootAccessEnabled,
				ValidateDiagFunc: enum.Validate[awstypes.RootAccess](),
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

func resourceNotebookInstanceCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &sagemaker.CreateNotebookInstanceInput{
		InstanceMetadataServiceConfiguration: expandNotebookInstanceMetadataServiceConfiguration(d.Get("instance_metadata_service_configuration").([]any)),
		InstanceType:                         awstypes.InstanceType(d.Get(names.AttrInstanceType).(string)),
		NotebookInstanceName:                 aws.String(name),
		RoleArn:                              aws.String(d.Get(names.AttrRoleARN).(string)),
		SecurityGroupIds:                     flex.ExpandStringValueSet(d.Get(names.AttrSecurityGroups).(*schema.Set)),
		Tags:                                 getTagsIn(ctx),
	}

	if v, ok := d.GetOk("additional_code_repositories"); ok && v.(*schema.Set).Len() > 0 {
		input.AdditionalCodeRepositories = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("default_code_repository"); ok {
		input.DefaultCodeRepository = aws.String(v.(string))
	}

	if v, ok := d.GetOk("direct_internet_access"); ok {
		input.DirectInternetAccess = awstypes.DirectInternetAccess(v.(string))
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
		input.RootAccess = awstypes.RootAccess(v.(string))
	}

	if s, ok := d.GetOk(names.AttrSubnetID); ok {
		input.SubnetId = aws.String(s.(string))
	}

	if v, ok := d.GetOk(names.AttrVolumeSize); ok {
		input.VolumeSizeInGB = aws.Int32(int32(v.(int)))
	}

	log.Printf("[DEBUG] Creating SageMaker AI Notebook Instance: %#v", input)
	_, err := conn.CreateNotebookInstance(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI Notebook Instance: %s", err)
	}

	d.SetId(name)

	if err := waitNotebookInstanceInService(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Notebook Instance (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceNotebookInstanceRead(ctx, d, meta)...)
}

func resourceNotebookInstanceRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	notebookInstance, err := findNotebookInstanceByName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] SageMaker AI Notebook Instance (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Notebook Instance (%s): %s", d.Id(), err)
	}

	d.Set("additional_code_repositories", notebookInstance.AdditionalCodeRepositories)
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
	d.Set(names.AttrSecurityGroups, notebookInstance.SecurityGroups)
	d.Set(names.AttrSubnetID, notebookInstance.SubnetId)
	d.Set(names.AttrURL, notebookInstance.Url)
	d.Set(names.AttrVolumeSize, notebookInstance.VolumeSizeInGB)

	if err := d.Set("instance_metadata_service_configuration", flattenNotebookInstanceMetadataServiceConfiguration(notebookInstance.InstanceMetadataServiceConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting instance_metadata_service_configuration: %s", err)
	}

	return diags
}

func resourceNotebookInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateNotebookInstanceInput{
			NotebookInstanceName: aws.String(d.Get(names.AttrName).(string)),
		}

		if d.HasChange("additional_code_repositories") {
			if v, ok := d.GetOk("additional_code_repositories"); ok {
				input.AdditionalCodeRepositories = flex.ExpandStringValueSet(v.(*schema.Set))
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
			input.InstanceMetadataServiceConfiguration = expandNotebookInstanceMetadataServiceConfiguration(d.Get("instance_metadata_service_configuration").([]any))
		}

		if d.HasChange(names.AttrInstanceType) {
			input.InstanceType = awstypes.InstanceType(d.Get(names.AttrInstanceType).(string))
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
			input.RootAccess = awstypes.RootAccess(d.Get("root_access").(string))
		}

		if d.HasChange(names.AttrVolumeSize) {
			input.VolumeSizeInGB = aws.Int32(int32(d.Get(names.AttrVolumeSize).(int)))
		}

		// Stop notebook.
		notebook, err := findNotebookInstanceByName(ctx, conn, d.Id())

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Notebook Instance (%s): %s", d.Id(), err)
		}

		previousStatus := notebook.NotebookInstanceStatus

		if previousStatus != awstypes.NotebookInstanceStatusStopped {
			if err := stopNotebookInstance(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Notebook Instance (%s): %s", d.Id(), err)
			}
		}

		if _, err := conn.UpdateNotebookInstance(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Notebook Instance (%s): %s", d.Id(), err)
		}

		if err := waitNotebookInstanceStopped(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Notebook Instance (%s) to stop: %s", d.Id(), err)
		}

		// Restart if needed
		if previousStatus == awstypes.NotebookInstanceStatusInService {
			if err := startNotebookInstance(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Notebook Instance (%s): %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceNotebookInstanceRead(ctx, d, meta)...)
}

func resourceNotebookInstanceDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	notebook, err := findNotebookInstanceByName(ctx, conn, d.Id())

	if retry.NotFound(err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Notebook Instance (%s): %s", d.Id(), err)
	}

	switch notebook.NotebookInstanceStatus {
	case awstypes.NotebookInstanceStatusInService:
		if err := stopNotebookInstance(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "stopping SageMaker AI Notebook Instance (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[DEBUG] Deleting SageMaker AI Notebook Instance: %s", d.Id())
	_, err = conn.DeleteNotebookInstance(ctx, &sagemaker.DeleteNotebookInstanceInput{
		NotebookInstanceName: aws.String(d.Id()),
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Notebook Instance (%s): %s", d.Id(), err)
	}

	if _, err := waitNotebookInstanceDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Notebook Instance (%s) delete: %s", d.Id(), err)
	}

	return diags
}

func findNotebookInstanceByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeNotebookInstanceOutput, error) {
	input := &sagemaker.DescribeNotebookInstanceInput{
		NotebookInstanceName: aws.String(name),
	}

	output, err := conn.DescribeNotebookInstance(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "RecordNotFound") {
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

func startNotebookInstance(ctx context.Context, conn *sagemaker.Client, id string) error {
	startOpts := &sagemaker.StartNotebookInstanceInput{
		NotebookInstanceName: aws.String(id),
	}
	// StartNotebookInstance sometimes doesn't take so we'll check for a state change and if
	// it doesn't change we'll send another request
	err := tfresource.Retry(ctx, 5*time.Minute, func(ctx context.Context) *tfresource.RetryError {
		_, err := conn.StartNotebookInstance(ctx, startOpts)
		if err != nil {
			return tfresource.NonRetryableError(fmt.Errorf("starting: %w", err))
		}

		err = waitNotebookInstanceStarted(ctx, conn, id)
		if err != nil {
			return tfresource.RetryableError(fmt.Errorf("starting: waiting for completion: %w", err))
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("starting: %w", err)
	}

	if err := waitNotebookInstanceInService(ctx, conn, id); err != nil {
		return fmt.Errorf("starting: waiting to be in service: %w", err)
	}
	return nil
}

func stopNotebookInstance(ctx context.Context, conn *sagemaker.Client, id string) error {
	notebook, err := findNotebookInstanceByName(ctx, conn, id)

	if err != nil {
		if retry.NotFound(err) {
			return nil
		}
		return err
	}

	if notebook.NotebookInstanceStatus == awstypes.NotebookInstanceStatusStopped {
		return nil
	}

	stopOpts := &sagemaker.StopNotebookInstanceInput{
		NotebookInstanceName: aws.String(id),
	}

	if _, err := conn.StopNotebookInstance(ctx, stopOpts); err != nil {
		return fmt.Errorf("stopping: %w", err)
	}

	if err := waitNotebookInstanceStopped(ctx, conn, id); err != nil {
		return fmt.Errorf("stopping: waiting for completion: %w", err)
	}

	return nil
}

func expandNotebookInstanceMetadataServiceConfiguration(l []any) *awstypes.InstanceMetadataServiceConfiguration {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]any)

	config := &awstypes.InstanceMetadataServiceConfiguration{
		MinimumInstanceMetadataServiceVersion: aws.String(m["minimum_instance_metadata_service_version"].(string)),
	}

	return config
}

func flattenNotebookInstanceMetadataServiceConfiguration(config *awstypes.InstanceMetadataServiceConfiguration) []map[string]any {
	if config == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"minimum_instance_metadata_service_version": aws.ToString(config.MinimumInstanceMetadataServiceVersion),
	}

	return []map[string]any{m}
}
