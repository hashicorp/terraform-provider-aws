package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsApiGatewayDomainName() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsApiGatewayDomainNameCreate,
		Read:   resourceAwsApiGatewayDomainNameRead,
		Update: resourceAwsApiGatewayDomainNameUpdate,
		Delete: resourceAwsApiGatewayDomainNameDelete,

		Schema: map[string]*schema.Schema{

			//According to AWS Documentation, ACM will be the only way to add certificates
			//to ApiGateway DomainNames. When this happens, we will be deprecating all certificate methods
			//except certificate_arn. We are not quite sure when this will happen.
			"certificate_body": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"certificate_arn", "regional_certificate_arn", "regional_certificate_name"},
			},

			"certificate_chain": {
				Type:          schema.TypeString,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"certificate_arn", "regional_certificate_arn", "regional_certificate_name"},
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
				ConflictsWith: []string{"certificate_arn", "regional_certificate_arn", "regional_certificate_name"},
			},

			"domain_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
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

			"regional_certificate_arn": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"certificate_arn", "certificate_body", "certificate_chain", "certificate_name", "certificate_private_key"},
			},

			"regional_certificate_name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"certificate_arn", "certificate_body", "certificate_chain", "certificate_name", "certificate_private_key"},
			},

			"regional_domain_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"regional_zone_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsApiGatewayDomainNameCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	endpointType := apigateway.EndpointTypeEdge
	log.Printf("[DEBUG] Creating API Gateway Domain Name")

	params := &apigateway.CreateDomainNameInput{
		DomainName: aws.String(d.Get("domain_name").(string)),
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

	if v, ok := d.GetOk("regional_certificate_arn"); ok {
		endpointType = apigateway.EndpointTypeRegional
		params.RegionalCertificateArn = aws.String(v.(string))
	}

	if v, ok := d.GetOk("regional_certificate_name"); ok {
		endpointType = apigateway.EndpointTypeRegional
		params.RegionalCertificateName = aws.String(v.(string))
	}

	params.EndpointConfiguration = &apigateway.EndpointConfiguration{
		Types: []*string{aws.String(endpointType)},
	}

	domainName, err := conn.CreateDomainName(params)
	if err != nil {
		return fmt.Errorf("Error creating API Gateway Domain Name: %s", err)
	}

	d.SetId(*domainName.DomainName)
	d.Set("cloudfront_domain_name", domainName.DistributionDomainName)
	d.Set("cloudfront_zone_id", cloudFrontRoute53ZoneID)
	d.Set("regional_domain_name", domainName.RegionalDomainName)
	d.Set("regional_zone_id", domainName.RegionalHostedZoneId)

	return resourceAwsApiGatewayDomainNameRead(d, meta)
}

func resourceAwsApiGatewayDomainNameRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Reading API Gateway Domain Name %s", d.Id())

	domainName, err := conn.GetDomainName(&apigateway.GetDomainNameInput{
		DomainName: aws.String(d.Id()),
	})
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "NotFoundException" {
			log.Printf("[WARN] API Gateway Domain Name (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("certificate_name", domainName.CertificateName)
	if err := d.Set("certificate_upload_date", domainName.CertificateUploadDate.Format(time.RFC3339)); err != nil {
		log.Printf("[DEBUG] Error setting certificate_upload_date: %s", err)
	}
	d.Set("cloudfront_domain_name", domainName.DistributionDomainName)
	d.Set("domain_name", domainName.DomainName)
	d.Set("certificate_arn", domainName.CertificateArn)
	d.Set("regional_certificate_arn", domainName.RegionalCertificateArn)
	d.Set("regional_certificate_name", domainName.RegionalCertificateName)

	return nil
}

func resourceAwsApiGatewayDomainNameUpdateOperations(d *schema.ResourceData) []*apigateway.PatchOperation {
	endpointType := ""
	operations := make([]*apigateway.PatchOperation, 0)

	if d.HasChange("certificate_name") {
		endpointType = apigateway.EndpointTypeEdge
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/certificateName"),
			Value: aws.String(d.Get("certificate_name").(string)),
		})
	}

	if d.HasChange("certificate_arn") {
		endpointType = apigateway.EndpointTypeEdge
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/certificateArn"),
			Value: aws.String(d.Get("certificate_arn").(string)),
		})
	}

	if d.HasChange("regional_certificate_name") {
		endpointType = apigateway.EndpointTypeRegional
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/regionalCertificateName"),
			Value: aws.String(d.Get("regional_certificate_name").(string)),
		})
	}

	if d.HasChange("regional_certificate_arn") {
		endpointType = apigateway.EndpointTypeRegional
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/regionalCertificateArn"),
			Value: aws.String(d.Get("regional_certificate_arn").(string)),
		})
	}

	// If the certificate name or ARN is changed it's possible that we've changed from one endpoint type to another, so
	// we'll always update the endpoint type in that case.

	if endpointType != "" {
		operations = append(operations, &apigateway.PatchOperation{
			Op:    aws.String("replace"),
			Path:  aws.String("/endpointConfiguration/types/0"),
			Value: aws.String(endpointType),
		})
	}

	return operations
}

func resourceAwsApiGatewayDomainNameUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Updating API Gateway Domain Name %s", d.Id())

	_, err := conn.UpdateDomainName(&apigateway.UpdateDomainNameInput{
		DomainName:      aws.String(d.Id()),
		PatchOperations: resourceAwsApiGatewayDomainNameUpdateOperations(d),
	})

	if err != nil {
		return err
	}

	return resourceAwsApiGatewayDomainNameRead(d, meta)
}

func resourceAwsApiGatewayDomainNameDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).apigateway
	log.Printf("[DEBUG] Deleting API Gateway Domain Name: %s", d.Id())

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteDomainName(&apigateway.DeleteDomainNameInput{
			DomainName: aws.String(d.Id()),
		})

		if err == nil {
			return nil
		}

		if apigatewayErr, ok := err.(awserr.Error); ok && apigatewayErr.Code() == "NotFoundException" {
			return nil
		}

		return resource.NonRetryableError(err)
	})
}
