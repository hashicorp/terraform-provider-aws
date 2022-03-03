package grafana

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceLicenseAssociation() *schema.Resource {
	return &schema.Resource{
		Create: resourceLicenseAssociationCreate,
		Read:   resourceLicenseAssociationRead,
		Delete: resourceLicenseAssociationDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"free_trial_expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_expiration": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"license_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(managedgrafana.LicenseType_Values(), false),
			},
			"workspace_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceLicenseAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	input := &managedgrafana.AssociateLicenseInput{
		LicenseType: aws.String(d.Get("license_type").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] Creating Grafana License Association: %s", input)
	output, err := conn.AssociateLicense(input)

	if err != nil {
		return fmt.Errorf("error creating Grafana License Association: %w", err)
	}

	d.SetId(aws.StringValue(output.Workspace.Id))

	if _, err := waitLicenseAssociationCreated(conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return fmt.Errorf("error waiting for Grafana License Association (%s) create: %w", d.Id(), err)
	}

	return resourceLicenseAssociationRead(d, meta)
}

func resourceLicenseAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	workspace, err := FindLicensedWorkspaceByID(conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Grafana License Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Grafana License Association (%s): %w", d.Id(), err)
	}

	if workspace.FreeTrialExpiration != nil {
		d.Set("free_trial_expiration", workspace.FreeTrialExpiration.Format(time.RFC3339))
	} else {
		d.Set("free_trial_expiration", nil)
	}
	if workspace.LicenseExpiration != nil {
		d.Set("license_expiration", workspace.LicenseExpiration.Format(time.RFC3339))
	} else {
		d.Set("license_expiration", nil)
	}
	d.Set("license_type", workspace.LicenseType)

	return nil
}

func resourceLicenseAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).GrafanaConn

	log.Printf("[DEBUG] Deleting Grafana License Association: %s", d.Id())
	_, err := conn.DisassociateLicense(&managedgrafana.DisassociateLicenseInput{
		LicenseType: aws.String(d.Get("license_type").(string)),
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Grafana License Association (%s): %w", d.Id(), err)
	}

	if _, err := waitWorkspaceUpdated(conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error waiting for Grafana License Association (%s) delete: %w", d.Id(), err)
	}

	return nil
}
