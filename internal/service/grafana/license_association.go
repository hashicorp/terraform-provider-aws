package grafana

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceLicenseAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceLicenseAssociationCreate,
		ReadWithoutTimeout:   resourceLicenseAssociationRead,
		DeleteWithoutTimeout: resourceLicenseAssociationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func resourceLicenseAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn()

	input := &managedgrafana.AssociateLicenseInput{
		LicenseType: aws.String(d.Get("license_type").(string)),
		WorkspaceId: aws.String(d.Get("workspace_id").(string)),
	}

	log.Printf("[DEBUG] Creating Grafana License Association: %s", input)
	output, err := conn.AssociateLicenseWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Grafana License Association: %s", err)
	}

	d.SetId(aws.StringValue(output.Workspace.Id))

	if _, err := waitLicenseAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Grafana License Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLicenseAssociationRead(ctx, d, meta)...)
}

func resourceLicenseAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn()

	workspace, err := FindLicensedWorkspaceByID(ctx, conn, d.Id())

	if tfresource.NotFound(err) && !d.IsNewResource() {
		log.Printf("[WARN] Grafana License Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Grafana License Association (%s): %s", d.Id(), err)
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

	return diags
}

func resourceLicenseAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaConn()

	log.Printf("[DEBUG] Deleting Grafana License Association: %s", d.Id())
	_, err := conn.DisassociateLicenseWithContext(ctx, &managedgrafana.DisassociateLicenseInput{
		LicenseType: aws.String(d.Get("license_type").(string)),
		WorkspaceId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, managedgrafana.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Grafana License Association (%s): %s", d.Id(), err)
	}

	if _, err := waitWorkspaceUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Grafana License Association (%s) delete: %s", d.Id(), err)
	}

	return diags
}
