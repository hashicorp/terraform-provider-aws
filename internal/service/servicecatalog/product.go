// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog

import (
	"context"
	"log"
	"slices"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/servicecatalog"
	awstypes "github.com/aws/aws-sdk-go-v2/service/servicecatalog/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_servicecatalog_product", name="Product")
// @Tags
// @Testing(skipEmptyTags=true, importIgnore="accept_language;provisioning_artifact_parameters.0.disable_template_validation")
// @Testing(tagsIdentifierAttribute="id", tagsResourceType="Product")
func resourceProduct() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProductCreate,
		ReadWithoutTimeout:   resourceProductRead,
		UpdateWithoutTimeout: resourceProductUpdate,
		DeleteWithoutTimeout: resourceProductDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceProductImport,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(ProductReadyTimeout),
			Read:   schema.DefaultTimeout(ProductReadTimeout),
			Update: schema.DefaultTimeout(ProductUpdateTimeout),
			Delete: schema.DefaultTimeout(ProductDeleteTimeout),
		},

		Schema: map[string]*schema.Schema{
			"accept_language": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      acceptLanguageEnglish,
				ValidateFunc: validation.StringInSlice(acceptLanguage_Values(), false),
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrCreatedTime: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDescription: {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"distributor": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"has_default_path": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrOwner: {
				Type:     schema.TypeString,
				Required: true,
			},
			"provisioning_artifact_parameters": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrDescription: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"disable_template_validation": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
							Default:  false,
						},
						names.AttrName: {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
						},
						"template_physical_id": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ExactlyOneOf: []string{
								"provisioning_artifact_parameters.0.template_url",
								"provisioning_artifact_parameters.0.template_physical_id",
							},
						},
						"template_url": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ExactlyOneOf: []string{
								"provisioning_artifact_parameters.0.template_url",
								"provisioning_artifact_parameters.0.template_physical_id",
							},
						},
						names.AttrType: {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ProvisioningArtifactType](),
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"support_description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"support_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"support_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
			names.AttrType: {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: enum.Validate[awstypes.ProductType](),
			},
		},
	}
}

func resourceProductCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	name := d.Get(names.AttrName).(string)
	input := &servicecatalog.CreateProductInput{
		IdempotencyToken: aws.String(id.UniqueId()),
		Name:             aws.String(name),
		Owner:            aws.String(d.Get(names.AttrOwner).(string)),
		ProductType:      awstypes.ProductType(d.Get(names.AttrType).(string)),
		ProvisioningArtifactParameters: expandProvisioningArtifactParameters(
			d.Get("provisioning_artifact_parameters").([]any)[0].(map[string]any),
		),
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distributor"); ok {
		input.Distributor = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_description"); ok {
		input.SupportDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_email"); ok {
		input.SupportEmail = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_url"); ok {
		input.SupportUrl = aws.String(v.(string))
	}

	outputRaw, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParametersException](ctx, d.Timeout(schema.TimeoutCreate), func() (any, error) {
		return conn.CreateProduct(ctx, input)
	}, "profile does not exist")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Service Catalog Product (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*servicecatalog.CreateProductOutput).ProductViewDetail.ProductViewSummary.ProductId))

	if _, err := waitProductReady(ctx, conn, aws.ToString(input.AcceptLanguage), d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Product (%s) to be ready: %s", d.Id(), err)
	}

	return append(diags, resourceProductRead(ctx, d, meta)...)
}

func resourceProductRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	output, err := waitProductReady(ctx, conn, d.Get("accept_language").(string), d.Id(), d.Timeout(schema.TimeoutRead))

	if !d.IsNewResource() && errs.IsA[*awstypes.ResourceNotFoundException](err) {
		log.Printf("[WARN] Service Catalog Product (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "describing Service Catalog Product (%s): %s", d.Id(), err)
	}

	if output == nil || output.ProductViewDetail == nil || output.ProductViewDetail.ProductViewSummary == nil {
		return sdkdiag.AppendErrorf(diags, "getting Service Catalog Product (%s): empty response", d.Id())
	}

	pvs := output.ProductViewDetail.ProductViewSummary

	d.Set(names.AttrARN, output.ProductViewDetail.ProductARN)
	if output.ProductViewDetail.CreatedTime != nil {
		d.Set(names.AttrCreatedTime, output.ProductViewDetail.CreatedTime.Format(time.RFC3339))
	}
	d.Set(names.AttrDescription, pvs.ShortDescription)
	d.Set("distributor", pvs.Distributor)
	d.Set("has_default_path", pvs.HasDefaultPath)
	d.Set(names.AttrName, pvs.Name)
	d.Set(names.AttrOwner, pvs.Owner)
	d.Set(names.AttrStatus, output.ProductViewDetail.Status)
	d.Set("support_description", pvs.SupportDescription)
	d.Set("support_email", pvs.SupportEmail)
	d.Set("support_url", pvs.SupportUrl)
	d.Set(names.AttrType, pvs.Type)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceProductUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	input := &servicecatalog.UpdateProductInput{
		Id: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("distributor"); ok {
		input.Distributor = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrName); ok {
		input.Name = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrOwner); ok {
		input.Owner = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_description"); ok {
		input.SupportDescription = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_email"); ok {
		input.SupportEmail = aws.String(v.(string))
	}

	if v, ok := d.GetOk("support_url"); ok {
		input.SupportUrl = aws.String(v.(string))
	}

	if d.HasChange(names.AttrTagsAll) {
		o, n := d.GetChange(names.AttrTagsAll)
		oldTags := tftags.New(ctx, o)
		newTags := tftags.New(ctx, n)

		if removedTags := oldTags.Removed(newTags).IgnoreSystem(names.ServiceCatalog); len(removedTags) > 0 {
			input.RemoveTags = removedTags.Keys()
		}

		if updatedTags := oldTags.Updated(newTags).IgnoreSystem(names.ServiceCatalog); len(updatedTags) > 0 {
			input.AddTags = svcTags(updatedTags)
		}
	}

	_, err := tfresource.RetryWhenIsAErrorMessageContains[*awstypes.InvalidParametersException](ctx, d.Timeout(schema.TimeoutUpdate), func() (any, error) {
		return conn.UpdateProduct(ctx, input)
	}, "profile does not exist")

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "updating Service Catalog Product (%s): %s", d.Id(), err)
	}

	return append(diags, resourceProductRead(ctx, d, meta)...)
}

func resourceProductDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	input := &servicecatalog.DeleteProductInput{
		Id: aws.String(d.Id()),
	}

	if v, ok := d.GetOk("accept_language"); ok {
		input.AcceptLanguage = aws.String(v.(string))
	}

	_, err := conn.DeleteProduct(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Service Catalog Product (%s): %s", d.Id(), err)
	}

	if _, err := waitProductDeleted(ctx, conn, d.Get("accept_language").(string), d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Service Catalog Product (%s) to be deleted: %s", d.Id(), err)
	}

	return diags
}

func resourceProductImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	conn := meta.(*conns.AWSClient).ServiceCatalogClient(ctx)

	productData, err := findProductByID(ctx, conn, d.Id())

	if err != nil {
		return []*schema.ResourceData{d}, err
	}

	// import the last entry in the summary
	if len(productData.ProvisioningArtifactSummaries) > 0 {
		provisioningArtifact := slices.MaxFunc(productData.ProvisioningArtifactSummaries, func(a, b awstypes.ProvisioningArtifactSummary) int {
			return aws.ToTime(a.CreatedTime).Compare(aws.ToTime(b.CreatedTime))
		})
		in := &servicecatalog.DescribeProvisioningArtifactInput{
			ProductId:              aws.String(d.Id()),
			ProvisioningArtifactId: provisioningArtifact.Id,
		}

		// Find additional artifact details.
		artifactData, err := conn.DescribeProvisioningArtifact(ctx, in)

		if err != nil {
			return []*schema.ResourceData{d}, err
		}

		d.Set("provisioning_artifact_parameters", flattenProvisioningArtifactParameters(artifactData))
	}

	return []*schema.ResourceData{d}, nil
}

func expandProvisioningArtifactParameters(tfMap map[string]any) *awstypes.ProvisioningArtifactProperties {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ProvisioningArtifactProperties{}

	if v, ok := tfMap[names.AttrDescription].(string); ok && v != "" {
		apiObject.Description = aws.String(v)
	}

	if v, ok := tfMap["disable_template_validation"].(bool); ok {
		apiObject.DisableTemplateValidation = v
	}

	info := make(map[string]string)

	// schema will enforce that one of these is present
	if v, ok := tfMap["template_physical_id"].(string); ok && v != "" {
		info["ImportFromPhysicalId"] = v
	}

	if v, ok := tfMap["template_url"].(string); ok && v != "" {
		info["LoadTemplateFromURL"] = v
	}

	apiObject.Info = info

	if v, ok := tfMap[names.AttrName].(string); ok && v != "" {
		apiObject.Name = aws.String(v)
	}

	if v, ok := tfMap[names.AttrType].(string); ok && v != "" {
		apiObject.Type = awstypes.ProvisioningArtifactType(v)
	}

	return apiObject
}

func flattenProvisioningArtifactParameters(apiObject *servicecatalog.DescribeProvisioningArtifactOutput) []any {
	if apiObject == nil {
		return nil
	}

	m := map[string]any{
		names.AttrDescription:         aws.ToString(apiObject.ProvisioningArtifactDetail.Description),
		"disable_template_validation": false, // set default because it cannot be read
		names.AttrName:                aws.ToString(apiObject.ProvisioningArtifactDetail.Name),
		names.AttrType:                apiObject.ProvisioningArtifactDetail.Type,
	}

	if apiObject.Info != nil {
		if v, ok := apiObject.Info["TemplateUrl"]; ok {
			m["template_url"] = v
		}

		if v, ok := apiObject.Info["PhysicalId"]; ok {
			m["template_physical_id"] = v
		}
	}

	return []any{m}
}
