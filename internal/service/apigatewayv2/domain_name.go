// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package apigatewayv2

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_apigatewayv2_domain_name", name="Domain Name")
// @Tags(identifierAttribute="arn")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/apigatewayv2;apigatewayv2.GetDomainNameOutput")
// @Testing(generator="github.com/hashicorp/terraform-provider-aws/internal/acctest;acctest.RandomSubdomain()")
// @Testing(tlsKey=true)
// @Testing(tlsKeyDomain=rName)
func resourceDomainName() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDomainNameCreate,
		ReadWithoutTimeout:   resourceDomainNameRead,
		UpdateWithoutTimeout: resourceDomainNameUpdate,
		DeleteWithoutTimeout: resourceDomainNameDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(60 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"api_mapping_selection_expression": {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrDomainName: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 512),
			},
			"domain_name_configuration": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrCertificateARN: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						names.AttrEndpointType: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(enum.Slice(awstypes.EndpointTypeRegional), true),
						},
						names.AttrHostedZoneID: {
							Type:     schema.TypeString,
							Computed: true,
						},
						names.AttrIPAddressType: {
							Type:             schema.TypeString,
							Optional:         true,
							Computed:         true,
							ValidateDiagFunc: enum.Validate[awstypes.IpAddressType](),
						},
						"ownership_verification_certificate_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidARN,
						},
						"security_policy": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(enum.Slice(awstypes.SecurityPolicyTls12), true),
						},
						"target_domain_name": {
							Type:     schema.TypeString,
							Computed: true,
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
			"routing_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:         true,
				ValidateDiagFunc: enum.ValidateIgnoreCase[awstypes.RoutingMode](),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
	}
}

func resourceDomainNameCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	domainName := d.Get(names.AttrDomainName).(string)
	input := apigatewayv2.CreateDomainNameInput{
		DomainName:               aws.String(domainName),
		DomainNameConfigurations: expandDomainNameConfigurations(d.Get("domain_name_configuration").([]any)),
		MutualTlsAuthentication:  expandMutualTLSAuthenticationInput(d.Get("mutual_tls_authentication").([]any)),
		Tags:                     getTagsIn(ctx),
	}
	if v, ok := d.GetOk("routing_mode"); ok {
		input.RoutingMode = awstypes.RoutingMode(v.(string))
	}

	output, err := conn.CreateDomainName(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 Domain Name (%s): %s", domainName, err)
	}

	d.SetId(aws.ToString(output.DomainName))

	if _, err := waitDomainNameAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 Domain Name (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceDomainNameRead(ctx, d, meta)...)
}

func resourceDomainNameRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	c := meta.(*conns.AWSClient)
	conn := c.APIGatewayV2Client(ctx)

	output, err := findDomainName(ctx, conn, d.Id())

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 Domain Name (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 Domain Name (%s): %s", d.Id(), err)
	}

	d.Set("api_mapping_selection_expression", output.ApiMappingSelectionExpression)
	d.Set(names.AttrARN, domainNameARN(ctx, c, d.Id()))
	d.Set(names.AttrDomainName, output.DomainName)
	if err := d.Set("domain_name_configuration", flattenDomainNameConfiguration(output.DomainNameConfigurations[0])); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting domain_name_configuration: %s", err)
	}
	if err := d.Set("mutual_tls_authentication", flattenMutualTLSAuthentication(output.MutualTlsAuthentication)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mutual_tls_authentication: %s", err)
	}
	d.Set("routing_mode", output.RoutingMode)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceDomainNameUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	if d.HasChanges("domain_name_configuration", "mutual_tls_authentication", "routing_mode") {
		input := apigatewayv2.UpdateDomainNameInput{
			DomainName:               aws.String(d.Id()),
			DomainNameConfigurations: expandDomainNameConfigurations(d.Get("domain_name_configuration").([]any)),
		}

		if d.HasChange("mutual_tls_authentication") {
			if v, ok := d.GetOk("mutual_tls_authentication"); ok && len(v.([]any)) > 0 && v.([]any)[0] != nil {
				tfMap := v.([]any)[0].(map[string]any)

				input.MutualTlsAuthentication = &awstypes.MutualTlsAuthenticationInput{}

				if d.HasChange("mutual_tls_authentication.0.truststore_uri") {
					input.MutualTlsAuthentication.TruststoreUri = aws.String(tfMap["truststore_uri"].(string))
				}

				if d.HasChange("mutual_tls_authentication.0.truststore_version") {
					input.MutualTlsAuthentication.TruststoreVersion = aws.String(tfMap["truststore_version"].(string))
				}
			} else {
				// To disable mutual TLS for a custom domain name, remove the truststore from your custom domain name.
				input.MutualTlsAuthentication = &awstypes.MutualTlsAuthenticationInput{
					TruststoreUri: aws.String(""),
				}
			}
		}

		if d.HasChange("routing_mode") {
			input.RoutingMode = awstypes.RoutingMode(d.Get("routing_mode").(string))
		}

		_, err := conn.UpdateDomainName(ctx, &input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 Domain Name (%s): %s", d.Id(), err)
		}

		if _, err := waitDomainNameAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 Domain Name (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainNameRead(ctx, d, meta)...)
}

func resourceDomainNameDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Client(ctx)

	log.Printf("[DEBUG] Deleting API Gateway v2 Domain Name: %s", d.Id())
	input := apigatewayv2.DeleteDomainNameInput{
		DomainName: aws.String(d.Id()),
	}
	_, err := conn.DeleteDomainName(ctx, &input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 Domain Name (%s): %s", d.Id(), err)
	}

	return diags
}

func findDomainName(ctx context.Context, conn *apigatewayv2.Client, name string) (*apigatewayv2.GetDomainNameOutput, error) {
	input := &apigatewayv2.GetDomainNameInput{
		DomainName: aws.String(name),
	}

	output, err := conn.GetDomainName(ctx, input)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError: err,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.DomainNameConfigurations) == 0 {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}

func statusDomainName(conn *apigatewayv2.Client, name string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findDomainName(ctx, conn, name)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.DomainNameConfigurations[0].DomainNameStatus), nil
	}
}

func waitDomainNameAvailable(ctx context.Context, conn *apigatewayv2.Client, name string, timeout time.Duration) (*apigatewayv2.GetDomainNameOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.DomainNameStatusUpdating),
		Target:  enum.Slice(awstypes.DomainNameStatusAvailable),
		Refresh: statusDomainName(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*apigatewayv2.GetDomainNameOutput); ok {
		retry.SetLastError(err, errors.New(aws.ToString(output.DomainNameConfigurations[0].DomainNameStatusMessage)))

		return output, err
	}

	return nil, err
}

func expandDomainNameConfiguration(tfMap map[string]any) awstypes.DomainNameConfiguration {
	apiObject := awstypes.DomainNameConfiguration{}

	if v, ok := tfMap[names.AttrCertificateARN].(string); ok && v != "" {
		apiObject.CertificateArn = aws.String(v)
	}

	if v, ok := tfMap[names.AttrEndpointType].(string); ok && v != "" {
		apiObject.EndpointType = awstypes.EndpointType(v)
	}

	if v, ok := tfMap[names.AttrIPAddressType].(string); ok && v != "" {
		apiObject.IpAddressType = awstypes.IpAddressType(v)
	}

	if v, ok := tfMap["ownership_verification_certificate_arn"].(string); ok && v != "" {
		apiObject.OwnershipVerificationCertificateArn = aws.String(v)
	}

	if v, ok := tfMap["security_policy"].(string); ok && v != "" {
		apiObject.SecurityPolicy = awstypes.SecurityPolicy(v)
	}

	return apiObject
}

func expandDomainNameConfigurations(tfList []any) []awstypes.DomainNameConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []awstypes.DomainNameConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]any)

		if !ok {
			continue
		}

		apiObjects = append(apiObjects, expandDomainNameConfiguration(tfMap))
	}

	return apiObjects
}

func flattenDomainNameConfiguration(apiObject awstypes.DomainNameConfiguration) []any {
	tfMap := map[string]any{
		names.AttrEndpointType:  apiObject.EndpointType,
		names.AttrIPAddressType: apiObject.IpAddressType,
		"security_policy":       apiObject.SecurityPolicy,
	}

	if v := apiObject.CertificateArn; v != nil {
		tfMap[names.AttrCertificateARN] = aws.ToString(v)
	}

	if v := apiObject.HostedZoneId; v != nil {
		tfMap[names.AttrHostedZoneID] = aws.ToString(v)
	}

	if v := apiObject.OwnershipVerificationCertificateArn; v != nil {
		tfMap["ownership_verification_certificate_arn"] = aws.ToString(v)
	}

	if v := apiObject.ApiGatewayDomainName; v != nil {
		tfMap["target_domain_name"] = aws.ToString(v)
	}

	return []any{tfMap}
}

func expandMutualTLSAuthenticationInput(tfList []any) *awstypes.MutualTlsAuthenticationInput {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]any)

	apiObject := &awstypes.MutualTlsAuthenticationInput{}

	if v, ok := tfMap["truststore_uri"].(string); ok && v != "" {
		apiObject.TruststoreUri = aws.String(v)
	}

	if v, ok := tfMap["truststore_version"].(string); ok && v != "" {
		apiObject.TruststoreVersion = aws.String(v)
	}

	return apiObject
}

func flattenMutualTLSAuthentication(apiObject *awstypes.MutualTlsAuthentication) []any {
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

func domainNameARN(ctx context.Context, c *conns.AWSClient, name string) string {
	return c.RegionalARNNoAccount(ctx, "apigateway", "/domainnames/"+name)
}
