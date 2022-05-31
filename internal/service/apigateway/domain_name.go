package apigateway

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceDomainName() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainNameCreate,
		Read:   resourceDomainNameRead,
		Update: resourceDomainNameUpdate,
		Delete: resourceDomainNameDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{

			//According to AWS Documentation, ACM will be the only way to add certificates
			//to ApiGateway DomainNames. When this happens, we will be deprecating all certificate methods
			//except certificate_arn. We are not quite sure when this will happen.
			"certificate_body": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"certificate_arn", "regional_certificate_arn"},
			},

			"certificate_chain": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"certificate_arn", "regional_certificate_arn"},
			},

			"certificate_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"certificate_arn", "regional_certificate_arn", "regional_certificate_name"},
			},

			"certificate_private_key": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"certificate_arn", "regional_certificate_arn"},
			},

			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"security_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					apigateway.SecurityPolicyTls10,
					apigateway.SecurityPolicyTls12,
				}, true),
			},

			"certificate_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"certificate_body", "certificate_chain", "certificate_name", "certificate_private_key", "regional_certificate_arn", "regional_certificate_name"},
			},

			"cloudfront_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"certificate_upload_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cloudfront_zone_id": {
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
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									apigateway.EndpointTypeEdge,
									apigateway.EndpointTypeRegional,
								}, false),
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
							ForceNew: true,
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

			"regional_certificate_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"certificate_arn", "certificate_body", "certificate_chain", "certificate_name", "certificate_private_key", "regional_certificate_name"},
			},

			"regional_certificate_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"certificate_arn", "certificate_name", "regional_certificate_arn"},
			},

			"regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"regional_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDomainNameCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	log.Printf("[DEBUG] Creating API Gateway Domain Name")

	params := &apigateway.CreateDomainNameInput{
		DomainName:              aws.String(d.Get("domain_name").(string)),
		MutualTlsAuthentication: expandMutualTLSAuthentication(d.Get("mutual_tls_authentication").([]interface{})),
	}

	if v, ok := d.GetOk("certificate_arn"); ok {
		params.CertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_name"); ok {
		params.CertificateName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_body"); ok {
		params.CertificateBody = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_chain"); ok {
		params.CertificateChain = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_private_key"); ok {
		params.CertificatePrivateKey = aws.String(v.(string))
	}

	if v, ok := d.GetOk("endpoint_configuration"); ok {
		params.EndpointConfiguration = expandEndpointConfiguration(v.([]interface{}))
	}

	if v, ok := d.GetOk("regional_certificate_arn"); ok {
		params.RegionalCertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("regional_certificate_name"); ok {
		params.RegionalCertificateName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("security_policy"); ok {
		params.SecurityPolicy = aws.String(v.(string))
	}

	if v, ok := d.GetOk("ownership_verification_certificate_arn"); ok {
		params.OwnershipVerificationCertificateArn = aws.String(v.(string))
	}

	if len(tags) > 0 {
		params.Tags = Tags(tags.IgnoreAWS())
	}

	domainName, err := conn.CreateDomainName(params)
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Domain Name: %s", err)
	}

	d.SetId(aws.StringValue(domainName.DomainName))

	return resourceDomainNameRead(d, meta)
}

func resourceDomainNameRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading API Gateway Domain Name %s", d.Id())

	domainName, err := conn.GetDomainName(&apigateway.GetDomainNameInput{
		DomainName: aws.String(d.Id()),
	})
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
			log.Printf("[WARN] API Gateway Domain Name (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading API Gateway Domain Name (%s): %w", d.Id(), err)
	}

	tags := KeyValueTags(domainName.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "apigateway",
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("/domainnames/%s", d.Id()),
	}.String()
	d.Set("arn", arn)
	d.Set("certificate_arn", domainName.CertificateArn)
	d.Set("certificate_name", domainName.CertificateName)
	if err := d.Set("certificate_upload_date", domainName.CertificateUploadDate.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting certificate_upload_date: %s", err)
	}
	d.Set("cloudfront_domain_name", domainName.DistributionDomainName)
	d.Set("cloudfront_zone_id", cloudFrontRoute53ZoneID)
	d.Set("domain_name", domainName.DomainName)
	d.Set("security_policy", domainName.SecurityPolicy)
	d.Set("ownership_verification_certificate_arn", domainName.OwnershipVerificationCertificateArn)

	if err := d.Set("endpoint_configuration", flattenEndpointConfiguration(domainName.EndpointConfiguration)); err != nil {
		return fmt.Errorf("error setting endpoint_configuration: %s", err)
	}

	if err = d.Set("mutual_tls_authentication", flattenMutualTLSAuthentication(domainName.MutualTlsAuthentication)); err != nil {
		return fmt.Errorf("error setting mutual_tls_authentication: %w", err)
	}

	d.Set("regional_certificate_arn", domainName.RegionalCertificateArn)
	d.Set("regional_certificate_name", domainName.RegionalCertificateName)
	d.Set("regional_domain_name", domainName.RegionalDomainName)
	d.Set("regional_zone_id", domainName.RegionalHostedZoneId)

	return nil
}

func resourceDomainNameUpdateOperations(d *schema.ResourceData) []*apigateway.PatchOperation {
	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("certificate_name") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/certificateName"),
			Value: aws.String(d.Get("certificate_name").(string)),
		})
	}

	if d.HasChange("certificate_arn") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/certificateArn"),
			Value: aws.String(d.Get("certificate_arn").(string)),
		})
	}

	if d.HasChange("regional_certificate_name") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/regionalCertificateName"),
			Value: aws.String(d.Get("regional_certificate_name").(string)),
		})
	}

	if d.HasChange("regional_certificate_arn") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/regionalCertificateArn"),
			Value: aws.String(d.Get("regional_certificate_arn").(string)),
		})
	}

	if d.HasChange("security_policy") {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String(apigateway.OpReplace),
			Path:  aws.String("/securityPolicy"),
			Value: aws.String(d.Get("security_policy").(string)),
		})
	}

	if d.HasChange("endpoint_configuration.0.types") {
		// The domain name must have an endpoint type.
		// If attempting to remove the configuration, do nothing.
		if v, ok := d.GetOk("endpoint_configuration"); ok && len(v.([]interface{})) > 0 {
			m := v.([]interface{})[0].(map[string]interface{})

			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/endpointConfiguration/types/0"),
				Value: aws.String(m["types"].([]interface{})[0].(string)),
			})
		}
	}

	if d.HasChange("mutual_tls_authentication") {
		mutTLSAuth := d.Get("mutual_tls_authentication").([]interface{})

		if len(mutTLSAuth) == 0 || mutTLSAuth[0] == nil {
			// To disable mutual TLS for a custom domain name, remove the truststore from your custom domain name.
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/mutualTlsAuthentication/truststoreUri"),
				Value: aws.String(""),
			})
		} else {
			operations = append(operations, &apigateway.PatchOperation{
				Op:    aws.String(apigateway.OpReplace),
				Path:  aws.String("/mutualTlsAuthentication/truststoreVersion"),
				Value: aws.String(mutTLSAuth[0].(map[string]interface{})["truststore_version"].(string)),
			})
		}
	}

	return operations
}

func resourceDomainNameUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Updating API Gateway Domain Name %s", d.Id())

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	_, err := conn.UpdateDomainName(&apigateway.UpdateDomainNameInput{
		DomainName:      aws.String(d.Id()),
		PatchOperations: resourceDomainNameUpdateOperations(d),
	})

	if err != nil {
		return err
	}

	return resourceDomainNameRead(d, meta)
}

func resourceDomainNameDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Deleting API Gateway Domain Name: %s", d.Id())

	_, err := conn.DeleteDomainName(&apigateway.DeleteDomainNameInput{
		DomainName: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, apigateway.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting API Gateway domain name: %s", err)
	}

	return nil
}

func expandMutualTLSAuthentication(tfList []interface{}) *apigateway.MutualTlsAuthenticationInput {
	if len(tfList) == 0 || tfList[0] == nil {
		return nil
	}

	tfMap := tfList[0].(map[string]interface{})

	apiObject := &apigateway.MutualTlsAuthenticationInput{}

	if v, ok := tfMap["truststore_uri"].(string); ok && v != "" {
		apiObject.TruststoreUri = aws.String(v)
	}

	if v, ok := tfMap["truststore_version"].(string); ok && v != "" {
		apiObject.TruststoreVersion = aws.String(v)
	}

	return apiObject
}

func flattenMutualTLSAuthentication(apiObject *apigateway.MutualTlsAuthentication) []interface{} {
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
