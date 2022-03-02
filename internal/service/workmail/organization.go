package workmail

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"alias": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"enable_interoperability": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"directory_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(12, 12),
			},
			"kms_key_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"domains": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"domain_name": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
						},
						"hosted_zone_id": {
							Type:     schema.TypeString,
							Optional: true,
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
	}

	if v, ok := d.GetOk("kms_key_arn"); ok {
		opts.KmsKeyArn = aws.String(v.(string))
	}
	if v, ok := d.GetOk("enable_interoperability"); ok {
		opts.EnableInteroperability = aws.Bool(v.(bool))
	}
	if v, ok := d.GetOk("directory_id"); ok {
		opts.DirectoryId = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Create WorkMail Organization: %s", opts)

	resp, err := conn.CreateOrganization(opts)
	if err != nil {
		return fmt.Errorf("error creating WorkMail Organization: %w", err)
	}

	d.SetId(aws.StringValue(resp.OrganizationId))

	_, err = waitOrganizationActive(conn, d.Id())

	if err != nil {
		return fmt.Errorf("error waiting for WorkMail Organization (%s) to create: %w", d.Id(), err)
	}

	return resourceOrganizationRead(d, meta)
}

func resourceOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkMailConn

	readOpts := &workmail.DescribeOrganizationInput{
		OrganizationId: aws.String(d.Id()),
	}

	response, err := conn.DescribeOrganization(readOpts)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WorkMail Organization (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading WorkMail Organization (%s): %w", d.Id(), err)
	}

	if err != nil {
		return err
	}

	d.Set("arn", response.ARN)
	d.Set("alias", response.Alias)
	d.Set("state", response.State)
	if response.CompletedDate != nil {
		d.Set("completed_date", aws.TimeValue(response.CompletedDate).Format(time.RFC3339))
	} else {
		d.Set("completed_date", nil)
	}
	d.Set("default_mail_domain", response.DefaultMailDomain)
	d.Set("directory_id", response.DirectoryId)
	d.Set("directory_type", response.DirectoryType)

	return nil
}

func resourceOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkMailConn
	deleteDirectory := true

	deleteOpts := &workmail.DeleteOrganizationInput{
		OrganizationId:  aws.String(d.Id()),
		DeleteDirectory: &deleteDirectory,
	}

	_, err := conn.DeleteOrganization(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting WorkMail organization: %s", err)
	}

	_, err = waitOrganizationDeleted(conn, d.Id())

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
