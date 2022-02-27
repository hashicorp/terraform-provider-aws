package workmail

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationCreate,
		Read:   resourceOrganizationRead,
		Delete: resourceOrganizationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"alias": {
				Type:     schema.TypeString,
				Required: true,
			},
			"client_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_interoperability": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"directory_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(12, 12),
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidARN,
			},
			"domains": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:         schema.TypeString,
							Computed:     true,
							ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
						},
						"hosted_zone_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func resourceOrganizationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkMailConn
	alias := d.Get("alias").(string)

	opts := &workmail.CreateOrganizationInput{
		Alias:       aws.String(alias),
		ClientToken: aws.String(resource.UniqueId()),
		KmsKeyArn:   aws.String(d.Get("kms_key_arn").(string)),
	}

	log.Printf("[DEBUG] Create WorkMail Organization: %s", opts)

	resp, err := conn.CreateOrganization(opts)
	if err != nil {
		return fmt.Errorf("error creating WorkMail Organization: %w", err)
	}

	d.SetId(aws.StringValue(resp.OrganizationId))

	return resourceOrganizationRead(d, meta)
}

func resourceOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkMailConn

	organizationID := d.Get("organization_id").(string)

	readOpts := &workmail.DescribeOrganizationInput{
		OrganizationId: aws.String(organizationID),
	}

	response, err := conn.DescribeOrganization(readOpts)

	if err != nil {
		return fmt.Errorf("error reading WorkMail Organization (%s): %w", d.Id(), err)
	}

	if err != nil {
		return err
	}

	d.Set("arn", response.ARN)
	d.Set("alias", response.Alias)
	d.Set("state", response.State)
	d.Set("completed_data", response.CompletedDate)
	d.Set("default_mail_domain", response.DefaultMailDomain)
	d.Set("directory_id", response.DirectoryId)
	d.Set("directory_type", response.DirectoryType)

	return nil
}

func resourceOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkMailConn

	organizationId := d.Get("organization_id").(string)
	clientToken := d.Get("client_token").(string)
	deleteDirectory := d.Get("delete_directory").(bool)

	deleteOpts := &workmail.DeleteOrganizationInput{
		OrganizationId:  aws.String(organizationId),
		ClientToken:     aws.String(clientToken),
		DeleteDirectory: aws.Bool(deleteDirectory),
	}

	_, err := conn.DeleteOrganization(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting WorkMail organization: %s", err)
	}

	return nil
}

// func flattenDomains(domains *workmail.Domains) []map[string]interface{} {
// 	result := make([]map[string]interface{}, 0, 1)

// 	if domains != nil {
// 		result = append(result, map[string]interface{}{
// 			"domain_name":    aws.StringValue(domains.DomainName),
// 			"hosted_zone_id": aws.StringValue(domains.HostedZoneId),
// 		})
// 	}
// 	return result
// }
