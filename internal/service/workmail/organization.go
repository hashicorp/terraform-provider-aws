package workmail

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/workmail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationCreate,
		Read:   resourceOrganizationRead,
		Update: resourceOrganizationUpdate,
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
	}

	log.Printf("[DEBUG] Create WorkMail Organization: %s", opts)

	resp, err := conn.CreateOrganization(opts)
	if err != nil {
		return fmt.Errorf("error creating WorkMail Organization: %w", err)
	}

	d.SetId(aws.StringValue(resp.Organization.OrganizationArn))

	return resourceOrganizationRead(d, meta)
}

func resourceOrganizationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkMailConn

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

	d.Set("alias", alias)
	d.Set("client_affinity", listener.ClientAffinity)
	d.Set("protocol", listener.Protocol)

	return nil
}

func resourceOrganizationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WorkMailConn

	organizationId := d.Get("organization_id").(string)

	deleteOpts := &workmail.DeleteIdentityInput{
		OrganizationId: aws.String(organizationId),
	}

	_, err := conn.DeleteIdentity(deleteOpts)
	if err != nil {
		return fmt.Errorf("Error deleting WorkMail organization: %s", err)
	}

	return nil
}
