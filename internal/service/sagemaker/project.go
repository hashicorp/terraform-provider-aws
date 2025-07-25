// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sagemaker"
	awstypes "github.com/aws/aws-sdk-go-v2/service/sagemaker/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_project", name="Project")
// @Tags(identifierAttribute="arn")
func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProjectCreate,
		ReadWithoutTimeout:   resourceProjectRead,
		UpdateWithoutTimeout: resourceProjectUpdate,
		DeleteWithoutTimeout: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexache.MustCompile(`^[0-9A-Za-z](-*[0-9A-Za-z]){0,31}$`),
						"Valid characters are a-z, A-Z, 0-9, and - (hyphen)."),
				),
			},
			"project_description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(1, 1024),
			},
			"service_catalog_provisioning_details": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"path_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"product_id": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"provisioning_artifact_id": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"provisioning_parameter": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrKey: {
										Type:     schema.TypeString,
										Required: true,
									},
									names.AttrValue: {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	name := d.Get("project_name").(string)
	input := &sagemaker.CreateProjectInput{
		ProjectName:                       aws.String(name),
		ServiceCatalogProvisioningDetails: expandProjectServiceCatalogProvisioningDetails(d.Get("service_catalog_provisioning_details").([]any)),
		Tags:                              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("project_description"); ok {
		input.ProjectDescription = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (any, error) {
		return conn.CreateProject(ctx, input)
	}, ErrCodeValidationException)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker AI project: %s", err)
	}

	d.SetId(name)

	if _, err := waitProjectCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Project (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	project, err := findProjectByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		d.SetId("")
		log.Printf("[WARN] Unable to find SageMaker AI Project (%s); removing from state", d.Id())
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading SageMaker AI Project (%s): %s", d.Id(), err)
	}

	d.Set("project_name", project.ProjectName)
	d.Set("project_id", project.ProjectId)
	d.Set(names.AttrARN, project.ProjectArn)
	d.Set("project_description", project.ProjectDescription)

	if err := d.Set("service_catalog_provisioning_details", flattenProjectServiceCatalogProvisioningDetails(project.ServiceCatalogProvisioningDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_catalog_provisioning_details: %s", err)
	}

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateProjectInput{
			ProjectName: aws.String(d.Id()),
		}

		if d.HasChange("project_description") {
			input.ProjectDescription = aws.String(d.Get("project_description").(string))
		}

		if d.HasChange("service_catalog_provisioning_details") {
			input.ServiceCatalogProvisioningUpdateDetails = expandProjectServiceCatalogProvisioningDetailsUpdate(d.Get("service_catalog_provisioning_details").([]any))
		}

		_, err := conn.UpdateProject(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker AI Project (%s): %s", d.Id(), err)
		}

		if _, err := waitProjectUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Project (%s) to be updated: %s", d.Id(), err)
		}
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerClient(ctx)

	log.Printf("[DEBUG] Deleting SageMaker AI Project: %s", d.Id())
	_, err := conn.DeleteProject(ctx, &sagemaker.DeleteProjectInput{
		ProjectName: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") ||
		tfawserr.ErrMessageContains(err, "ValidationException", "Cannot delete Project in DeleteCompleted status") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker AI Project (%s): %s", d.Id(), err)
	}

	if _, err := waitProjectDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker AI Project (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func findProjectByName(ctx context.Context, conn *sagemaker.Client, name string) (*sagemaker.DescribeProjectOutput, error) {
	input := &sagemaker.DescribeProjectInput{
		ProjectName: aws.String(name),
	}

	output, err := conn.DescribeProject(ctx, input)

	if tfawserr.ErrMessageContains(err, ErrCodeValidationException, "does not exist") {
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

	if status := output.ProjectStatus; status == awstypes.ProjectStatusDeleteCompleted {
		return nil, &retry.NotFoundError{
			Message:     string(status),
			LastRequest: input,
		}
	}

	return output, nil
}

func expandProjectServiceCatalogProvisioningDetails(l []any) *awstypes.ServiceCatalogProvisioningDetails {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]any)

	scpd := &awstypes.ServiceCatalogProvisioningDetails{
		ProductId: aws.String(m["product_id"].(string)),
	}

	if v, ok := m["path_id"].(string); ok && v != "" {
		scpd.PathId = aws.String(v)
	}

	if v, ok := m["provisioning_artifact_id"].(string); ok && v != "" {
		scpd.ProvisioningArtifactId = aws.String(v)
	}

	if v, ok := m["provisioning_parameter"].([]any); ok && len(v) > 0 {
		scpd.ProvisioningParameters = expandProjectProvisioningParameters(v)
	}

	return scpd
}

func expandProjectServiceCatalogProvisioningDetailsUpdate(l []any) *awstypes.ServiceCatalogProvisioningUpdateDetails {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]any)

	scpd := &awstypes.ServiceCatalogProvisioningUpdateDetails{}

	if v, ok := m["provisioning_artifact_id"].(string); ok && v != "" {
		scpd.ProvisioningArtifactId = aws.String(v)
	}

	if v, ok := m["provisioning_parameter"].([]any); ok && len(v) > 0 {
		scpd.ProvisioningParameters = expandProjectProvisioningParameters(v)
	}

	return scpd
}

func expandProjectProvisioningParameters(l []any) []awstypes.ProvisioningParameter {
	if len(l) == 0 {
		return nil
	}

	params := make([]awstypes.ProvisioningParameter, 0, len(l))

	for _, lRaw := range l {
		data := lRaw.(map[string]any)

		scpd := awstypes.ProvisioningParameter{
			Key: aws.String(data[names.AttrKey].(string)),
		}

		if v, ok := data[names.AttrValue].(string); ok && v != "" {
			scpd.Value = aws.String(v)
		}

		params = append(params, scpd)
	}

	return params
}

func flattenProjectServiceCatalogProvisioningDetails(scpd *awstypes.ServiceCatalogProvisioningDetails) []map[string]any {
	if scpd == nil {
		return []map[string]any{}
	}

	m := map[string]any{
		"product_id": aws.ToString(scpd.ProductId),
	}

	if scpd.PathId != nil {
		m["path_id"] = aws.ToString(scpd.PathId)
	}

	if scpd.ProvisioningArtifactId != nil {
		m["provisioning_artifact_id"] = aws.ToString(scpd.ProvisioningArtifactId)
	}

	if scpd.ProvisioningParameters != nil {
		m["provisioning_parameter"] = flattenProjectProvisioningParameters(scpd.ProvisioningParameters)
	}

	return []map[string]any{m}
}

func flattenProjectProvisioningParameters(scpd []awstypes.ProvisioningParameter) []map[string]any {
	params := make([]map[string]any, 0, len(scpd))

	for _, lRaw := range scpd {
		param := make(map[string]any)
		param[names.AttrKey] = aws.ToString(lRaw.Key)

		if lRaw.Value != nil {
			param[names.AttrValue] = lRaw.Value
		}

		params = append(params, param)
	}

	return params
}
