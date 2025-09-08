// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	awstypes "github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	ret "github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_cloudfront_distribution_tenant", name="Distribution Tenant")
// @Tags(identifierAttribute="arn")
func resourceDistributionTenant() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDistributionTenantCreate,
		ReadWithoutTimeout:   resourceDistributionTenantRead,
		UpdateWithoutTimeout: resourceDistributionTenantUpdate,
		DeleteWithoutTimeout: resourceDistributionTenantDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				// Set non API attributes to their Default settings in the schema
				d.Set("wait_for_deployment", true)
				return []*schema.ResourceData{d}, nil
			}},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"connection_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"customizations": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"geo_restriction": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"locations": {
										Type:     schema.TypeSet,
										Optional: true,
										Computed: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
									},
									"restriction_type": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.GeoRestrictionType](),
									},
								},
							},
						},
						"certificate": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrARN: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"web_acl": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									names.AttrAction: {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateDiagFunc: enum.Validate[awstypes.CustomizationActionType](),
									},
									names.AttrARN: {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
			"distribution_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"domains": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			names.AttrEnabled: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"etag": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"last_modified_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"managed_certificate_request": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate_transparency_logging_preference": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.CertificateTransparencyLoggingPreference](),
						},
						"primary_domain_name": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"validation_token_host": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: enum.Validate[awstypes.ValidationTokenHost](),
						},
					},
				},
			},
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
			},
			names.AttrParameters: {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrName: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			"wait_for_deployment": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDistributionTenantCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	input := cloudfront.CreateDistributionTenantInput{
		DistributionId: aws.String(d.Get("distribution_id").(string)),
		Domains:        expandDomains(d.Get("domains").(*schema.Set)),
		Enabled:        aws.Bool(d.Get(names.AttrEnabled).(bool)),
		Name:           aws.String(d.Get(names.AttrName).(string)),
	}

	if v, ok := d.GetOk("connection_group_id"); ok {
		input.ConnectionGroupId = aws.String(v.(string))
	}
	if v, ok := d.GetOk("customizations"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.Customizations = expandCustomizations(v.([]any)[0].(map[string]any))
	}
	if v, ok := d.GetOk("managed_certificate_request"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
		input.ManagedCertificateRequest = expandManagedCertificateRequest(v.([]any)[0].(map[string]any))
	}
	if v, ok := d.GetOk(names.AttrParameters); ok {
		input.Parameters = expandParameters(v.([]any))
	}
	if tags := getTagsIn(ctx); len(tags) > 0 {
		input.Tags = &awstypes.Tags{Items: tags}
	}

	// ACM and IAM certificate eventual consistency.
	// InvalidViewerCertificate: The specified SSL certificate doesn't exist, isn't in us-east-1 region, isn't valid, or doesn't include a valid certificate chain.
	const (
		timeout = 1 * time.Minute
	)

	outputRaw, err := tfresource.RetryWhenIsA[any, *awstypes.InvalidViewerCertificate](ctx, timeout, func(ctx context.Context) (any, error) {
		return conn.CreateDistributionTenant(ctx, &input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating CloudFront Distribution Tenant: %s", err)
	}

	d.SetId(aws.ToString(outputRaw.(*cloudfront.CreateDistributionTenantOutput).DistributionTenant.Id))

	if d.Get("wait_for_deployment").(bool) {
		if _, err := waitDistributionTenantDeployed(ctx, conn, d.Id()); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for CloudFront Distribution Tenant (%s) deploy: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDistributionTenantRead(ctx, d, meta)...)
}

func resourceDistributionTenantRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	output, err := findDistributionTenantByID(ctx, conn, d.Id())

	if !d.IsNewResource() && ret.NotFound(err) {
		log.Printf("[WARN] CloudFront Distribution Tenant (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading CloudFront Distribution Tenant (%s): %s", d.Id(), err)
	}

	tenant := output.DistributionTenant
	d.Set(names.AttrARN, tenant.Arn)
	d.Set("connection_group_id", tenant.ConnectionGroupId)
	if tenant.Customizations != nil {
		if err := d.Set("customizations", []any{flattenCustomizations(tenant.Customizations)}); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting customizations: %s", err)
		}
	}
	d.Set("distribution_id", tenant.DistributionId)
	d.Set("domains", flattenDomains(tenant.Domains))
	d.Set("etag", output.ETag)
	d.Set(names.AttrEnabled, tenant.Enabled)
	d.Set("last_modified_time", aws.String(tenant.LastModifiedTime.String()))
	d.Set(names.AttrName, tenant.Name)
	if tenant.Parameters != nil {
		if err := d.Set(names.AttrParameters, flattenParameters(tenant.Parameters)); err != nil {
			return sdkdiag.AppendErrorf(diags, "setting parameters: %s", err)
		}
	}
	d.Set(names.AttrStatus, tenant.Status)

	return diags
}

func resourceDistributionTenantUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := cloudfront.UpdateDistributionTenantInput{
			Id:             aws.String(d.Id()),
			IfMatch:        aws.String(d.Get("etag").(string)),
			DistributionId: aws.String(d.Get("distribution_id").(string)),
			Domains:        expandDomains(d.Get("domains").(*schema.Set)),
			Enabled:        aws.Bool(d.Get(names.AttrEnabled).(bool)),
		}

		if v, ok := d.GetOk("connection_group_id"); ok {
			input.ConnectionGroupId = aws.String(v.(string))
		}
		if v, ok := d.GetOk("customizations"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.Customizations = expandCustomizations(v.([]any)[0].(map[string]any))
		}
		if v, ok := d.GetOk("managed_certificate_request"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
			input.ManagedCertificateRequest = expandManagedCertificateRequest(v.([]any)[0].(map[string]any))
		}
		if v, ok := d.GetOk(names.AttrParameters); ok {
			input.Parameters = expandParameters(v.([]any))
		}

		// ACM and IAM certificate eventual consistency.
		// InvalidViewerCertificate: The specified SSL certificate doesn't exist, isn't in us-east-1 region, isn't valid, or doesn't include a valid certificate chain.
		const (
			timeout = 1 * time.Minute
		)
		_, err := tfresource.RetryWhenIsA[any, *awstypes.InvalidViewerCertificate](ctx, timeout, func(ctx context.Context) (any, error) {
			return conn.UpdateDistributionTenant(ctx, &input)
		})

		// Refresh our ETag if it is out of date and attempt update again.
		if errs.IsA[*awstypes.PreconditionFailed](err) {
			var etag string
			etag, err = distributionTenantETag(ctx, conn, d.Id())

			if err != nil {
				return sdkdiag.AppendFromErr(diags, err)
			}

			input.IfMatch = aws.String(etag)

			_, err = conn.UpdateDistributionTenant(ctx, &input)
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating CloudFront Distribution Tenant (%s): %s", d.Id(), err)
		}
		if d.Get("wait_for_deployment").(bool) {
			if _, err := waitDistributionTenantDeployed(ctx, conn, d.Id()); err != nil {
				return sdkdiag.AppendErrorf(diags, "waiting for CloudFront Distribution Tenant (%s) deploy: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceDistributionTenantRead(ctx, d, meta)...)
}

func resourceDistributionTenantDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).CloudFrontClient(ctx)

	if d.Get(names.AttrARN).(string) == "" {
		diags = append(diags, resourceDistributionTenantRead(ctx, d, meta)...)
	}

	if err := disableDistributionTenant(ctx, conn, d.Id()); err != nil {
		if ret.NotFound(err) {
			return diags
		}
		return sdkdiag.AppendFromErr(diags, err)
	}

	err := deleteDistributionTenant(ctx, conn, d.Id())

	if err == nil || ret.NotFound(err) || errs.IsA[*awstypes.EntityNotFound](err) {
		return diags
	}

	// Disable distribution tenant if it is not yet disabled and attempt deletion again.
	// Here we update via the deployed configuration to ensure we are not submitting an out of date
	// configuration from the Terraform configuration, should other changes have occurred manually.
	if errs.IsA[*awstypes.ResourceNotDisabled](err) {
		if err := disableDistributionTenant(ctx, conn, d.Id()); err != nil {
			if ret.NotFound(err) {
				return diags
			}

			return sdkdiag.AppendFromErr(diags, err)
		}

		const (
			timeout = 3 * time.Minute
		)
		_, err = tfresource.RetryWhenIsA[any, *awstypes.ResourceNotDisabled](ctx, timeout, func(ctx context.Context) (any, error) {
			return nil, deleteDistributionTenant(ctx, conn, d.Id())
		})
	}

	if errs.IsA[*awstypes.PreconditionFailed](err) || errs.IsA[*awstypes.InvalidIfMatchVersion](err) {
		const (
			timeout = 1 * time.Minute
		)
		_, err = tfresource.RetryWhenIsOneOf2[any, *awstypes.PreconditionFailed, *awstypes.InvalidIfMatchVersion](ctx, timeout, func(ctx context.Context) (any, error) {
			return nil, deleteDistributionTenant(ctx, conn, d.Id())
		})
	}

	if errs.IsA[*awstypes.ResourceNotDisabled](err) {
		if err := disableDistributionTenant(ctx, conn, d.Id()); err != nil {
			if ret.NotFound(err) {
				return diags
			}

			return sdkdiag.AppendFromErr(diags, err)
		}

		err = deleteDistributionTenant(ctx, conn, d.Id())
	}

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	return diags
}

func waitDistributionTenantDeployed(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionTenantOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{distributionTenantStatusInProgress},
		Target:     []string{distributionTenantStatusDeployed},
		Refresh:    statusDistributionTenant(ctx, conn, id),
		Timeout:    90 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetDistributionTenantOutput); ok {
		return output, err
	}

	return nil, err
}

func waitDistributionTenantDeleted(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionTenantOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{distributionTenantStatusInProgress, distributionTenantStatusDeployed},
		Target:     []string{},
		Refresh:    statusDistributionTenant(ctx, conn, id),
		Timeout:    90 * time.Minute,
		MinTimeout: 15 * time.Second,
		Delay:      15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudfront.GetDistributionTenantOutput); ok {
		return output, err
	}

	return nil, err
}

func statusDistributionTenant(ctx context.Context, conn *cloudfront.Client, id string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDistributionTenantByID(ctx, conn, id)

		if ret.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, aws.ToString(output.DistributionTenant.Status), nil
	}
}

func findDistributionTenantByID(ctx context.Context, conn *cloudfront.Client, id string) (*cloudfront.GetDistributionTenantOutput, error) {
	input := cloudfront.GetDistributionTenantInput{
		Identifier: aws.String(id),
	}

	output, err := conn.GetDistributionTenant(ctx, &input)

	if errs.IsA[*awstypes.EntityNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DistributionTenant == nil || output.DistributionTenant.Domains == nil || output.DistributionTenant.DistributionId == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func disableDistributionTenant(ctx context.Context, conn *cloudfront.Client, id string) error {
	output, err := findDistributionTenantByID(ctx, conn, id)

	if err != nil {
		return fmt.Errorf("reading CloudFront Distribution Tenant (%s): %w", id, err)
	}

	if aws.ToString(output.DistributionTenant.Status) == distributionTenantStatusInProgress {
		output, err = waitDistributionTenantDeployed(ctx, conn, id)

		if err != nil {
			return fmt.Errorf("waiting for CloudFront Distribution (%s) deploy: %w", id, err)
		}
	}

	if !aws.ToBool(output.DistributionTenant.Enabled) {
		return nil
	}

	input := cloudfront.UpdateDistributionTenantInput{
		Id:                aws.String(id),
		IfMatch:           output.ETag,
		ConnectionGroupId: output.DistributionTenant.ConnectionGroupId,
		Customizations:    output.DistributionTenant.Customizations,
		DistributionId:    output.DistributionTenant.DistributionId,
		Domains:           convertDomainResultsToDomainItems(output.DistributionTenant.Domains),
		Parameters:        output.DistributionTenant.Parameters,
	}

	input.Enabled = aws.Bool(false)

	_, err = conn.UpdateDistributionTenant(ctx, &input)

	if err != nil {
		return fmt.Errorf("updating CloudFront Distribution Tenant (%s): %w", id, err)
	}

	if _, err := waitDistributionTenantDeployed(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Distribution Tenant (%s) deploy: %w", id, err)
	}

	return nil
}

func convertDomainResultsToDomainItems(results []awstypes.DomainResult) []awstypes.DomainItem {
	items := make([]awstypes.DomainItem, len(results))
	for i, result := range results {
		items[i] = awstypes.DomainItem{
			Domain: result.Domain,
		}
	}
	return items
}

func distributionTenantETag(ctx context.Context, conn *cloudfront.Client, id string) (string, error) {
	output, err := findDistributionTenantByID(ctx, conn, id)

	if err != nil {
		return "", fmt.Errorf("reading CloudFront Distribution Tenant (%s): %w", id, err)
	}

	return aws.ToString(output.ETag), nil
}

func deleteDistributionTenant(ctx context.Context, conn *cloudfront.Client, id string) error {
	etag, err := distributionTenantETag(ctx, conn, id)

	if err != nil {
		return err
	}

	input := cloudfront.DeleteDistributionTenantInput{
		Id:      aws.String(id),
		IfMatch: aws.String(etag),
	}

	_, err = conn.DeleteDistributionTenant(ctx, &input)

	if err != nil {
		return fmt.Errorf("deleting CloudFront Distribution Tenant (%s): %w", id, err)
	}

	if _, err := waitDistributionTenantDeleted(ctx, conn, id); err != nil {
		return fmt.Errorf("waiting for CloudFront Distribution Tenant (%s) delete: %w", id, err)
	}

	return nil
}

func expandDomains(tfSet *schema.Set) []awstypes.DomainItem {
	if tfSet.Len() == 0 {
		return nil
	}

	domains := make([]awstypes.DomainItem, tfSet.Len())
	for i, v := range tfSet.List() {
		domains[i] = awstypes.DomainItem{
			Domain: aws.String(v.(string)),
		}
	}
	return domains
}

func flattenDomains(apiObjects []awstypes.DomainResult) *schema.Set {
	if len(apiObjects) == 0 {
		return nil
	}

	tfSet := schema.NewSet(schema.HashString, []any{})
	for _, apiObject := range apiObjects {
		if apiObject.Domain != nil {
			tfSet.Add(aws.ToString(apiObject.Domain))
		}
	}
	return tfSet
}

func expandCustomizations(tfMap map[string]any) *awstypes.Customizations {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Customizations{}

	if v, ok := tfMap["geo_restriction"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.GeoRestrictions = expandTenantGeoRestriction(v[0].(map[string]any))
	}

	if v, ok := tfMap["certificate"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.Certificate = expandCertificate(v[0].(map[string]any))
	}

	if v, ok := tfMap["web_acl"].([]any); ok && len(v) > 0 && v[0] != nil {
		apiObject.WebAcl = expandWebAcl(v[0].(map[string]any))
	}

	return apiObject
}

func expandTenantGeoRestriction(tfMap map[string]any) *awstypes.GeoRestrictionCustomization {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.GeoRestrictionCustomization{}

	if v, ok := tfMap["locations"].(*schema.Set); ok && v.Len() > 0 {
		apiObject.Locations = flex.ExpandStringValueSet(v)
	}

	if v, ok := tfMap["restriction_type"].(string); ok && v != "" {
		apiObject.RestrictionType = awstypes.GeoRestrictionType(v)
	}

	return apiObject
}

func expandCertificate(tfMap map[string]any) *awstypes.Certificate {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.Certificate{}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	return apiObject
}

func expandWebAcl(tfMap map[string]any) *awstypes.WebAclCustomization {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.WebAclCustomization{}

	if v, ok := tfMap[names.AttrAction].(string); ok && v != "" {
		apiObject.Action = awstypes.CustomizationActionType(v)
	}

	if v, ok := tfMap[names.AttrARN].(string); ok && v != "" {
		apiObject.Arn = aws.String(v)
	}

	return apiObject
}

func flattenCustomizations(apiObject *awstypes.Customizations) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.GeoRestrictions != nil {
		tfMap["geo_restriction"] = []any{flattenTenantGeoRestriction(apiObject.GeoRestrictions)}
	}

	if apiObject.Certificate != nil {
		tfMap["certificate"] = []any{flattenCertificate(apiObject.Certificate)}
	}

	if apiObject.WebAcl != nil {
		tfMap["web_acl"] = []any{flattenWebAcl(apiObject.WebAcl)}
	}

	return tfMap
}

func flattenTenantGeoRestriction(apiObject *awstypes.GeoRestrictionCustomization) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Locations != nil {
		tfMap["locations"] = flex.FlattenStringValueSet(apiObject.Locations)
	}

	tfMap["restriction_type"] = string(apiObject.RestrictionType)

	return tfMap
}

func flattenCertificate(apiObject *awstypes.Certificate) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if apiObject.Arn != nil {
		tfMap[names.AttrARN] = aws.ToString(apiObject.Arn)
	}

	return tfMap
}

func flattenWebAcl(apiObject *awstypes.WebAclCustomization) map[string]any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	tfMap[names.AttrAction] = string(apiObject.Action)

	if apiObject.Arn != nil {
		tfMap[names.AttrARN] = aws.ToString(apiObject.Arn)
	}

	return tfMap
}

func expandManagedCertificateRequest(tfMap map[string]any) *awstypes.ManagedCertificateRequest {
	if tfMap == nil {
		return nil
	}

	apiObject := &awstypes.ManagedCertificateRequest{}

	if v, ok := tfMap["certificate_transparency_logging_preference"].(string); ok && v != "" {
		apiObject.CertificateTransparencyLoggingPreference = awstypes.CertificateTransparencyLoggingPreference(v)
	}

	if v, ok := tfMap["primary_domain_name"].(string); ok && v != "" {
		apiObject.PrimaryDomainName = aws.String(v)
	}

	if v, ok := tfMap["validation_token_host"].(string); ok && v != "" {
		apiObject.ValidationTokenHost = awstypes.ValidationTokenHost(v)
	}

	return apiObject
}

func expandParameters(tfList []any) []awstypes.Parameter {
	if len(tfList) == 0 {
		return nil
	}

	parameters := make([]awstypes.Parameter, len(tfList))
	for i, v := range tfList {
		tfMap := v.(map[string]any)
		parameters[i] = awstypes.Parameter{
			Name:  aws.String(tfMap[names.AttrName].(string)),
			Value: aws.String(tfMap[names.AttrValue].(string)),
		}
	}
	return parameters
}

func flattenParameters(apiObjects []awstypes.Parameter) []any {
	if len(apiObjects) == 0 {
		return nil
	}

	tfList := make([]any, len(apiObjects))
	for i, apiObject := range apiObjects {
		tfList[i] = map[string]any{
			names.AttrName:  aws.ToString(apiObject.Name),
			names.AttrValue: aws.ToString(apiObject.Value),
		}
	}
	return tfList
}
