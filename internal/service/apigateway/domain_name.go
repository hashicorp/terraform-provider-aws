// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_api_gateway_domain_name", name="Domain Name")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigateway;apigateway.GetDomainNameOutput")
// @Testing(generator="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.RandomSubdomain()")
// @Testing(tlsKey=true, tlsKeyDomain="rName")
func resourceDomainName() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainNameCreate,
		ReadWithoutTimeout:   resourceDomainNameRead,
		UpdateWithoutTimeout: resourceDomainNameUpdate,
		DeleteWithoutTimeout: resourceDomainNameDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			//According to AWS Documentation, ACM will be the only way to add certificates
			//to ApiGateway DomainNames. When this happens, we will be deprecating all certificate methods
			//except certificate_arn. We are not quite sure when this will happen.
			names.AttrCertificateARN: {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"certificate_body", names.AttrCertificateChain, "certificate_name", "certificate_private_key", "regional_certificate_arn", "regional_certificate_name"},
			},
			"certificate_body": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{names.AttrCertificateARN, "regional_certificate_arn"},
			},
			names.AttrCertificateChain: {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{names.AttrCertificateARN, "regional_certificate_arn"},
			},
			"certificate_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrCertificateARN, "regional_certificate_arn", "regional_certificate_name"},
			},
			"certificate_private_key": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{names.AttrCertificateARN, "regional_certificate_arn"},
			},
			"certificate_upload_date": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"cloudfront_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"domain_name_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"endpoint_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"types": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							// BadRequestException: Cannot create an api with multiple Endpoint Types
							MaxItems: 1,
							Elem: &schema.Schema{
								Type:             schema.TypeString,
								ValidateDiagFunc: enum.Validate[types.EndpointType](),
							},
						},
					},
				},
			},
			"mutual_tls_authentication": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"truststore_uri": {
							Type:     schema.TypeString,
							Required: true,
						},
						"truststore_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"ownership_verification_certificate_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: verify.ValidARN,
			},
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Optional:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"regional_certificate_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrCertificateARN, "certificate_body", names.AttrCertificateChain, "certificate_name", "certificate_private_key", "regional_certificate_name"},
			},
			"regional_certificate_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{names.AttrCertificateARN, "certificate_name", "regional_certificate_arn"},
			},
			"regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"regional_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"security_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.Validate[types.SecurityPolicy](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDomainNameCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	input := apigateway.CreateDomainNameInput{
		DomainName:              aws.String(domainName),
		MutualTlsAuthentication: expandMutualTLSAuthentication(d.Get("mutual_tls_authentication").([]any)),
		Tags:                    getTagsIn(ctx),
	}

	if v, ok := d.GetOk(names.AttrCertificateARN); ok {
		input.CertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_body"); ok {
		input.CertificateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrCertificateChain); ok {
		input.CertificateChain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_name"); ok {
		input.CertificateName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_private_key"); ok {
		input.CertificatePrivateKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		input.EndpointConfiguration = expandEndpointConfiguration(v.([]any))
	}

	if v, ok := d.GetOk("ownership_verification_certificate_arn"); ok {
		input.OwnershipVerificationCertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk(names.AttrPolicy); ok {
		input.Policy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("regional_certificate_arn"); ok {
		input.RegionalCertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("regional_certificate_name"); ok {
		input.RegionalCertificateName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_policy"); ok {
		input.SecurityPolicy = types.SecurityPolicy(v.(string))
	}

	output, err := conn.CreateDomainName(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway Domain Name (%s): %s", domainName, err)
	}

	d.SetId(domainNameCreateResourceID(aws.ToString(output.DomainName), aws.ToString(output.DomainNameId)))

	return append(diags, resourceDomainNameRead(ctx, d, meta)...)
}

func resourceDomainNameRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, domainNameID, err := domainNameParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	output, err := findDomainNameByTwoPartKey(ctx, conn, domainName, domainNameID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway Domain Name (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway Domain Name (%s): %s", d.Id(), err)
	}

	if output.DomainNameArn != nil { // nosemgrep: ci.helper-schema-ResourceData-Set-extraneous-nil-check
		d.Set(names.AttrARN, output.DomainNameArn)
	} else {
		d.Set(names.AttrARN, domainNameARN(ctx, meta.(*conns.AWSClient), d.Id()))
	}
	d.Set(names.AttrCertificateARN, output.CertificateArn)
	d.Set("certificate_name", output.CertificateName)
	if output.CertificateUploadDate != nil {
		d.Set("certificate_upload_date", output.CertificateUploadDate.Format(time.RFC3339))
	} else {
		d.Set("certificate_upload_date", nil)
	}
	d.Set("cloudfront_domain_name", output.DistributionDomainName)
	d.Set("cloudfront_zone_id", meta.(*conns.AWSClient).CloudFrontDistributionHostedZoneID(ctx))
	d.Set(names.AttrDomainName, output.DomainName)
	d.Set("domain_name_id", output.DomainNameId)
	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(output.EndpointConfiguration)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting endpoint_configuration: %s", err)
	}
	if err = d.Set("mutual_tls_authentication", flattenMutualTLSAuthentication(output.MutualTlsAuthentication)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mutual_tls_authentication: %s", err)
	}
	d.Set("ownership_verification_certificate_arn", output.OwnershipVerificationCertificateArn)
	d.Set(names.AttrPolicy, output.Policy)
	d.Set("regional_certificate_arn", output.RegionalCertificateArn)
	d.Set("regional_certificate_name", output.RegionalCertificateName)
	d.Set("regional_domain_name", output.RegionalDomainName)
	d.Set("regional_zone_id", output.RegionalHostedZoneId)
	d.Set("security_policy", output.SecurityPolicy)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceDomainNameUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, domainNameID, err := domainNameParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		var operations []types.PatchOperation

		if d.HasChange(names.AttrCertificateARN) {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/certificateArn"),
				Value: aws.String(d.Get(names.AttrCertificateARN).(string)),
			})
		}

		if d.HasChange("certificate_name") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/certificateName"),
				Value: aws.String(d.Get("certificate_name").(string)),
			})
		}

		if d.HasChange("endpoint_configuration.0.types") {
			// The domain name must have an endpoint type.
			// If attempting to remove the configuration, do nothing.
			if v, ok := d.GetOk("endpoint_configuration"); ok && len(v.([]any)) > 0 {
				tfMap := v.([]any)[0].(map[string]any)

				operations = append(operations, types.PatchOperation{
					Op:    types.OpReplace,
					Path:  aws.String("/endpointConfiguration/types/0"),
					Value: aws.String(tfMap["types"].([]any)[0].(string)),
				})
			}
		}

		if d.HasChange("mutual_tls_authentication") {
			if v, ok := d.GetOk("mutual_tls_authentication"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				tfMap := v.([]any)[0].(map[string]any)

				if d.HasChange("mutual_tls_authentication.0.truststore_uri") {
					operations = append(operations, types.PatchOperation{
						Op:    types.OpReplace,
						Path:  aws.String("/mutualTlsAuthentication/truststoreUri"),
						Value: aws.String(tfMap["truststore_uri"].(string)),
					})
				}

				if d.HasChange("mutual_tls_authentication.0.truststore_version") {
					operations = append(operations, types.PatchOperation{
						Op:    types.OpReplace,
						Path:  aws.String("/mutualTlsAuthentication/truststoreVersion"),
						Value: aws.String(tfMap["truststore_version"].(string)),
					})
				}
			} else {
				// To disable mutual TLS for a custom domain name, remove the truststore from your custom domain name.
				operations = append(operations, types.PatchOperation{
					Op:    types.OpReplace,
					Path:  aws.String("/mutualTlsAuthentication/truststoreUri"),
					Value: aws.String(""),
				})
			}
		}

		if d.HasChange("ownership_verification_certificate_arn") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/ownershipVerificationCertificateArn"),
				Value: aws.String(d.Get("ownership_verification_certificate_arn").(string)),
			})
		}

		if d.HasChange(names.AttrPolicy) {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/policy"),
				Value: aws.String(d.Get(names.AttrPolicy).(string)),
			})
		}

		if d.HasChange("regional_certificate_arn") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/regionalCertificateArn"),
				Value: aws.String(d.Get("regional_certificate_arn").(string)),
			})
		}

		if d.HasChange("regional_certificate_name") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/regionalCertificateName"),
				Value: aws.String(d.Get("regional_certificate_name").(string)),
			})
		}

		if d.HasChange("security_policy") {
			operations = append(operations, types.PatchOperation{
				Op:    types.OpReplace,
				Path:  aws.String("/securityPolicy"),
				Value: aws.String(d.Get("security_policy").(string)),
			})
		}

		input := apigateway.UpdateDomainNameInput{
			DomainName:      aws.String(domainName),
			PatchOperations: operations,
		}
		if domainNameID != "" {
			input.DomainNameId = aws.String(domainNameID)
		}

		_, err := conn.UpdateDomainName(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway Domain Name (%s): %s", d.Id(), err)
		}

		if _, err := waitDomainNameUpdated(ctx, conn, domainName, domainNameID); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for API Gateway Domain Name (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainNameRead(ctx, d, meta)...)
}

func resourceDomainNameDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayClient(ctx)

	domainName, domainNameID, err := domainNameParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	input := apigateway.DeleteDomainNameInput{
		DomainName: aws.String(domainName),
	}
	if domainNameID != "" {
		input.DomainNameId = aws.String(domainNameID)
	}

	log.Printf("[DEBUG] Deleting API Gateway Domain Name: %s", d.Id())
	_, err = conn.DeleteDomainName(ctx, &input)

	if errs.IsA[*types.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway Domain Name (%s): %s", d.Id(), err)
	}

	return diags
}

const domainNameResourceIDSeparator = "/"

func domainNameCreateResourceID(domainName, domainNameID string) string {
	var id string

	if domainNameID == "" {
		id = domainName
	} else {
		parts := []string{domainName, domainNameID}
		id = strings.Join(parts, domainNameResourceIDSeparator)
	}

	return id
}

func domainNameParseResourceID(id string) (string, string, error) {
	switch parts := strings.SplitN(id, domainNameResourceIDSeparator, 2); len(parts) {
	case 1:
		if domainName := parts[0]; domainName != "" {
			return domainName, "", nil
		}
	case 2:
		if domainName, domainNameID := parts[0], parts[1]; domainName != "" && domainNameID != "" {
			return domainName, domainNameID, nil
		}
	}

	return "", "", fmt.Errorf("unexpected format for ID (%[1]s), expected DOMAIN-NAME or DOMAIN-NAME%[2]sDOMAIN-NAME-ID", id, domainNameResourceIDSeparator)
}

func findDomainNameByTwoPartKey(ctx context.Context, conn *apigateway.Client, domainName, domainNameID string) (*apigateway.GetDomainNameOutput, error) {
	input := apigateway.GetDomainNameInput{
		DomainName: aws.String(domainName),
	}
	if domainNameID != "" {
		input.DomainNameId = aws.String(domainNameID)
	}

	output, err := conn.GetDomainName(ctx, &input)

	if errs.IsA[*types.NotFoundException](err) {
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

func statusDomainName(ctx context.Context, conn *apigateway.Client, domainName, domainNameID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		output, err := findDomainNameByTwoPartKey(ctx, conn, domainName, domainNameID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.DomainNameStatus), nil
	}
}

func waitDomainNameUpdated(ctx context.Context, conn *apigateway.Client, domainName, domainNameID string) (*types.DomainName, error) {
	const (
		timeout = 15 * time.Minute
	)
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(types.DomainNameStatusUpdating),
		Target:     enum.Slice(types.DomainNameStatusAvailable),
		Refresh:    statusDomainName(ctx, conn, domainName, domainNameID),
		Timeout:    timeout,
		Delay:      1 * time.Minute,
		MinTimeout: 10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.DomainName); ok {
		tfresource.SetLastError(err, errors.New(aws.ToString(output.DomainNameStatusMessage)))

		return output, err
	}

	return nil, err
}

func expandMutualTLSAuthentication(tfList []any) *types.MutualTlsAuthenticationInput {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &types.MutualTlsAuthenticationInput{} // nosemgrep:ci.semgrep.aws.input-on-heap

	if v, ok := tfMap["truststore_uri"].(string); ok && v != "" {
		apiObject.TruststoreUri = aws.String(v)
	}

	if v, ok := tfMap["truststore_version"].(string); ok && v != "" {
		apiObject.TruststoreVersion = aws.String(v)
	}

	return apiObject
}

func flattenMutualTLSAuthentication(apiObject *types.MutualTlsAuthentication) []any {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]any{}

	if v := apiObject.TruststoreUri; v != nil {
		tfMap["truststore_uri"] = aws.ToString(v)
	}

	if v := apiObject.TruststoreVersion; v != nil {
		tfMap["truststore_version"] = aws.ToString(v)
	}

	return []any{tfMap}
}

func domainNameARN(ctx context.Context, c *conns.AWSClient, domainName string) string {
	return c.RegionalARNNoAccount(ctx, "apigateway", fmt.Sprintf("/domainnames/%s", domainName))
}
