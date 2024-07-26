// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package grafana

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/grafana"
	awstypes "github.com/aws/aws-sdk-go-v2/service/grafana/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// @SDKResource("aws_grafana_license_association", name="License Association")
func resourceLicenseAssociation() *schema.Resource {
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
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: enum.Validate[awstypes.LicenseType](),
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
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	workspaceID := d.Get("workspace_id").(string)
	input := &grafana.AssociateLicenseInput{
		LicenseType: awstypes.LicenseType(d.Get("license_type").(string)),
		WorkspaceId: aws.String(workspaceID),
	}

	output, err := conn.AssociateLicense(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Grafana License Association (%s): %s", workspaceID, err)
	}

	d.SetId(aws.ToString(output.Workspace.Id))

	if _, err := waitLicenseAssociationCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for Grafana License Association (%s) create: %s", d.Id(), err)
	}

	return append(diags, resourceLicenseAssociationRead(ctx, d, meta)...)
}

func resourceLicenseAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	workspace, err := findLicensedWorkspaceByID(ctx, conn, d.Id())

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
	d.Set("workspace_id", d.Id())

	return diags
}

func resourceLicenseAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).GrafanaClient(ctx)

	log.Printf("[DEBUG] Deleting Grafana License Association: %s", d.Id())
	_, err := conn.DisassociateLicense(ctx, &grafana.DisassociateLicenseInput{
		LicenseType: awstypes.LicenseType(d.Get("license_type").(string)),
		WorkspaceId: aws.String(d.Id()),
	})

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
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

func findLicensedWorkspaceByID(ctx context.Context, conn *grafana.Client, id string) (*awstypes.WorkspaceDescription, error) {
	output, err := findWorkspaceByID(ctx, conn, id)

	if err != nil {
		return nil, err
	}

	if output.LicenseType == "" {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}

func waitLicenseAssociationCreated(ctx context.Context, conn *grafana.Client, id string, timeout time.Duration) (*awstypes.WorkspaceDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.WorkspaceStatusUpgrading),
		Target:  enum.Slice(awstypes.WorkspaceStatusActive),
		Refresh: statusWorkspace(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.WorkspaceDescription); ok {
		return output, err
	}

	return nil, err
}
