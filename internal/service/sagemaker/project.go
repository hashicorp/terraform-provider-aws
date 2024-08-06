// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sagemaker

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sagemaker"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sagemaker_project", name="Project")
// @Tags(identifierAttribute="arn")
func ResourceProject() *schema.Resource {
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

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	name := d.Get("project_name").(string)
	input := &sagemaker.CreateProjectInput{
		ProjectName:                       aws.String(name),
		ServiceCatalogProvisioningDetails: expandProjectServiceCatalogProvisioningDetails(d.Get("service_catalog_provisioning_details").([]interface{})),
		Tags:                              getTagsIn(ctx),
	}

	if v, ok := d.GetOk("project_description"); ok {
		input.ProjectDescription = aws.String(v.(string))
	}

	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, 2*time.Minute, func() (interface{}, error) {
		return conn.CreateProjectWithContext(ctx, input)
	}, "ValidationException")
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating SageMaker project: %s", err)
	}

	d.SetId(name)

	if _, err := WaitProjectCreated(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Project (%s) to be created: %s", d.Id(), err)
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	project, err := FindProjectByName(ctx, conn, d.Id())
	if err != nil {
		if !d.IsNewResource() && tfresource.NotFound(err) {
			d.SetId("")
			log.Printf("[WARN] Unable to find SageMaker Project (%s); removing from state", d.Id())
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading SageMaker Project (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(project.ProjectArn)
	d.Set("project_name", project.ProjectName)
	d.Set("project_id", project.ProjectId)
	d.Set(names.AttrARN, arn)
	d.Set("project_description", project.ProjectDescription)

	if err := d.Set("service_catalog_provisioning_details", flattenProjectServiceCatalogProvisioningDetails(project.ServiceCatalogProvisioningDetails)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting service_catalog_provisioning_details: %s", err)
	}

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &sagemaker.UpdateProjectInput{
			ProjectName: aws.String(d.Id()),
		}

		if d.HasChange("project_description") {
			input.ProjectDescription = aws.String(d.Get("project_description").(string))
		}

		if d.HasChange("service_catalog_provisioning_details") {
			input.ServiceCatalogProvisioningUpdateDetails = expandProjectServiceCatalogProvisioningDetailsUpdate(d.Get("service_catalog_provisioning_details").([]interface{}))
		}

		_, err := conn.UpdateProjectWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating SageMaker Project (%s): %s", d.Id(), err)
		}

		if _, err := WaitProjectUpdated(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Project (%s) to be updated: %s", d.Id(), err)
		}
	}

	return append(diags, resourceProjectRead(ctx, d, meta)...)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SageMakerConn(ctx)

	log.Printf("[DEBUG] Deleting SageMaker Project: %s", d.Id())
	_, err := conn.DeleteProjectWithContext(ctx, &sagemaker.DeleteProjectInput{
		ProjectName: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, "ValidationException", "does not exist") ||
		tfawserr.ErrMessageContains(err, "ValidationException", "Cannot delete Project in DeleteCompleted status") {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting SageMaker Project (%s): %s", d.Id(), err)
	}

	if _, err := WaitProjectDeleted(ctx, conn, d.Id()); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for SageMaker Project (%s) to delete: %s", d.Id(), err)
	}

	return diags
}

func expandProjectServiceCatalogProvisioningDetails(l []interface{}) *sagemaker.ServiceCatalogProvisioningDetails {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	scpd := &sagemaker.ServiceCatalogProvisioningDetails{
		ProductId: aws.String(m["product_id"].(string)),
	}

	if v, ok := m["path_id"].(string); ok && v != "" {
		scpd.PathId = aws.String(v)
	}

	if v, ok := m["provisioning_artifact_id"].(string); ok && v != "" {
		scpd.ProvisioningArtifactId = aws.String(v)
	}

	if v, ok := m["provisioning_parameter"].([]interface{}); ok && len(v) > 0 {
		scpd.ProvisioningParameters = expandProjectProvisioningParameters(v)
	}

	return scpd
}

func expandProjectServiceCatalogProvisioningDetailsUpdate(l []interface{}) *sagemaker.ServiceCatalogProvisioningUpdateDetails {
	if len(l) == 0 {
		return nil
	}

	m := l[0].(map[string]interface{})

	scpd := &sagemaker.ServiceCatalogProvisioningUpdateDetails{}

	if v, ok := m["provisioning_artifact_id"].(string); ok && v != "" {
		scpd.ProvisioningArtifactId = aws.String(v)
	}

	if v, ok := m["provisioning_parameter"].([]interface{}); ok && len(v) > 0 {
		scpd.ProvisioningParameters = expandProjectProvisioningParameters(v)
	}

	return scpd
}

func expandProjectProvisioningParameters(l []interface{}) []*sagemaker.ProvisioningParameter {
	if len(l) == 0 {
		return nil
	}

	params := make([]*sagemaker.ProvisioningParameter, 0, len(l))

	for _, lRaw := range l {
		data := lRaw.(map[string]interface{})

		scpd := &sagemaker.ProvisioningParameter{
			Key: aws.String(data[names.AttrKey].(string)),
		}

		if v, ok := data[names.AttrValue].(string); ok && v != "" {
			scpd.Value = aws.String(v)
		}

		params = append(params, scpd)
	}

	return params
}

func flattenProjectServiceCatalogProvisioningDetails(scpd *sagemaker.ServiceCatalogProvisioningDetails) []map[string]interface{} {
	if scpd == nil {
		return []map[string]interface{}{}
	}

	m := map[string]interface{}{
		"product_id": aws.StringValue(scpd.ProductId),
	}

	if scpd.PathId != nil {
		m["path_id"] = aws.StringValue(scpd.PathId)
	}

	if scpd.ProvisioningArtifactId != nil {
		m["provisioning_artifact_id"] = aws.StringValue(scpd.ProvisioningArtifactId)
	}

	if scpd.ProvisioningParameters != nil {
		m["provisioning_parameter"] = flattenProjectProvisioningParameters(scpd.ProvisioningParameters)
	}

	return []map[string]interface{}{m}
}

func flattenProjectProvisioningParameters(scpd []*sagemaker.ProvisioningParameter) []map[string]interface{} {
	params := make([]map[string]interface{}, 0, len(scpd))

	for _, lRaw := range scpd {
		param := make(map[string]interface{})
		param[names.AttrKey] = aws.StringValue(lRaw.Key)

		if lRaw.Value != nil {
			param[names.AttrValue] = lRaw.Value
		}

		params = append(params, param)
	}

	return params
}
