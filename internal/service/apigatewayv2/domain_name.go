package apigatewayv2

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainName() *schema.Resource {
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain_name": {
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
						"certificate_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: verify.ValidARN,
						},
						"endpoint_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								apigatewayv2.EndpointTypeRegional,
							}, true),
						},
						"hosted_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"security_policy": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								apigatewayv2.SecurityPolicyTls12,
							}, true),
						},
						"target_domain_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ownership_verification_certificate_arn": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: verify.ValidARN,
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
							ForceNew: true,
						},
						"truststore_version": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDomainNameCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	domainName := d.Get("domain_name").(string)

	input := &apigatewayv2.CreateDomainNameInput{
		DomainName:               aws.String(domainName),
		DomainNameConfigurations: expandDomainNameConfigurations(d.Get("domain_name_configuration").([]interface{})),
		MutualTlsAuthentication:  expandMutualTLSAuthentication(d.Get("mutual_tls_authentication").([]interface{})),
		Tags:                     Tags(tags.IgnoreAWS()),
	}

	log.Printf("[DEBUG] Creating API Gateway v2 domain name: %s", input)
	output, err := conn.CreateDomainNameWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating API Gateway v2 domain name (%s): %s", domainName, err)
	}

	d.SetId(aws.StringValue(output.DomainName))

	if _, err := WaitDomainNameAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 domain name (%s) to become available: %s", d.Id(), err)
	}

	return append(diags, resourceDomainNameRead(ctx, d, meta)...)
}

func resourceDomainNameRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	output, err := FindDomainNameByName(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] API Gateway v2 domain name (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading API Gateway v2 domain name (%s): %s", d.Id(), err)
	}

	d.Set("api_mapping_selection_expression", output.ApiMappingSelectionExpression)
	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/domainnames/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("domain_name", output.DomainName)

	err = d.Set("domain_name_configuration", flattenDomainNameConfiguration(output.DomainNameConfigurations[0]))
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "setting domain_name_configuration: %s", err)
	}

	if err = d.Set("mutual_tls_authentication", flattenMutualTLSAuthentication(output.MutualTlsAuthentication)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting mutual_tls_authentication: %s", err)
	}

	tags := KeyValueTags(output.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting tags_all: %s", err)
	}

	return diags
}

func resourceDomainNameUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

	if d.HasChanges("domain_name_configuration", "mutual_tls_authentication") {
		input := &apigatewayv2.UpdateDomainNameInput{
			DomainName:               aws.String(d.Id()),
			DomainNameConfigurations: expandDomainNameConfigurations(d.Get("domain_name_configuration").([]interface{})),
		}

		if d.HasChange("mutual_tls_authentication") {
			mutTLSAuth := d.Get("mutual_tls_authentication").([]interface{})

			if len(mutTLSAuth) == 0 || mutTLSAuth[0] == nil {
				// To disable mutual TLS for a custom domain name, remove the truststore from your custom domain name.
				input.MutualTlsAuthentication = &apigatewayv2.MutualTlsAuthenticationInput{
					TruststoreUri: aws.String(""),
				}
			} else {
				input.MutualTlsAuthentication = &apigatewayv2.MutualTlsAuthenticationInput{
					TruststoreVersion: aws.String(mutTLSAuth[0].(map[string]interface{})["truststore_version"].(string)),
				}
			}
		}

		log.Printf("[DEBUG] Updating API Gateway v2 domain name: %s", input)
		_, err := conn.UpdateDomainNameWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 domain name (%s): %s", d.Id(), err)
		}

		if _, err := WaitDomainNameAvailable(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for API Gateway v2 domain name (%s) to become available: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating API Gateway v2 domain name (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceDomainNameRead(ctx, d, meta)...)
}

func resourceDomainNameDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).APIGatewayV2Conn()

	log.Printf("[DEBUG] Deleting API Gateway v2 Domain Name: %s", d.Id())
	_, err := conn.DeleteDomainNameWithContext(ctx, &apigatewayv2.DeleteDomainNameInput{
		DomainName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, apigatewayv2.ErrCodeNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting API Gateway v2 domain name (%s): %s", d.Id(), err)
	}

	return diags
}

func expandDomainNameConfiguration(tfMap map[string]interface{}) *apigatewayv2.DomainNameConfiguration {
	if tfMap == nil {
		return nil
	}

	apiObject := &apigatewayv2.DomainNameConfiguration{}

	if v, ok := tfMap["certificate_arn"].(string); ok && v != "" {
		apiObject.CertificateArn = aws.String(v)
	}

	if v, ok := tfMap["endpoint_type"].(string); ok && v != "" {
		apiObject.EndpointType = aws.String(v)
	}

	if v, ok := tfMap["security_policy"].(string); ok && v != "" {
		apiObject.SecurityPolicy = aws.String(v)
	}

	if v, ok := tfMap["ownership_verification_certificate_arn"].(string); ok && v != "" {
		apiObject.OwnershipVerificationCertificateArn = aws.String(v)
	}

	return apiObject
}

func expandDomainNameConfigurations(tfList []interface{}) []*apigatewayv2.DomainNameConfiguration {
	if len(tfList) == 0 {
		return nil
	}

	var apiObjects []*apigatewayv2.DomainNameConfiguration

	for _, tfMapRaw := range tfList {
		tfMap, ok := tfMapRaw.(map[string]interface{})

		if !ok {
			continue
		}

		apiObject := expandDomainNameConfiguration(tfMap)

		if apiObject == nil {
			continue
		}

		apiObjects = append(apiObjects, apiObject)
	}

	return apiObjects
}

func flattenDomainNameConfiguration(apiObject *apigatewayv2.DomainNameConfiguration) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.CertificateArn; v != nil {
		tfMap["certificate_arn"] = aws.StringValue(v)
	}

	if v := apiObject.EndpointType; v != nil {
		tfMap["endpoint_type"] = aws.StringValue(v)
	}

	if v := apiObject.HostedZoneId; v != nil {
		tfMap["hosted_zone_id"] = aws.StringValue(v)
	}

	if v := apiObject.SecurityPolicy; v != nil {
		tfMap["security_policy"] = aws.StringValue(v)
	}

	if v := apiObject.ApiGatewayDomainName; v != nil {
		tfMap["target_domain_name"] = aws.StringValue(v)
	}

	if v := apiObject.OwnershipVerificationCertificateArn; v != nil {
		tfMap["ownership_verification_certificate_arn"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}

func expandMutualTLSAuthentication(tfList []interface{}) *apigatewayv2.MutualTlsAuthenticationInput {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &apigatewayv2.MutualTlsAuthenticationInput{}

	if v, ok := tfMap["truststore_uri"].(string); ok && v != "" {
		apiObject.TruststoreUri = aws.String(v)
	}

	if v, ok := tfMap["truststore_version"].(string); ok && v != "" {
		apiObject.TruststoreVersion = aws.String(v)
	}

	return apiObject
}

func flattenMutualTLSAuthentication(apiObject *apigatewayv2.MutualTlsAuthentication) []interface{} {
	if apiObject == nil {
		return nil
	}

	tfMap := map[string]interface{}{}

	if v := apiObject.TruststoreUri; v != nil {
		tfMap["truststore_uri"] = aws.StringValue(v)
	}

	if v := apiObject.TruststoreVersion; v != nil {
		tfMap["truststore_version"] = aws.StringValue(v)
	}

	return []interface{}{tfMap}
}
