// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/redshift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/redshift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_partner", name="Partner")
func resourcePartner() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePartnerCreate,
		ReadWithoutTimeout:   resourcePartnerRead,
		DeleteWithoutTimeout: resourcePartnerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrAccountID: {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrDatabaseName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"partner_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrStatus: {
				Type:     schema.TypeString,
				Computed: true,
			},
			names.AttrStatusMessage: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourcePartnerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	accountID, clusterID, databaseName, partnerName := d.Get(names.AttrAccountID).(string), d.Get(names.AttrClusterIdentifier).(string), d.Get(names.AttrDatabaseName).(string), d.Get("partner_name").(string)
	id := fmt.Sprintf("%s:%s:%s:%s", accountID, clusterID, databaseName, partnerName)
	input := redshift.AddPartnerInput{
		AccountId:         aws.String(accountID),
		ClusterIdentifier: aws.String(clusterID),
		DatabaseName:      aws.String(databaseName),
		PartnerName:       aws.String(partnerName),
	}
	_, err := conn.AddPartner(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Partner (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourcePartnerRead(ctx, d, meta)...)
}

func resourcePartnerRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	accountID, clusterID, databaseName, partnerName, err := decodePartnerID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	out, err := findPartnerByFourPartKey(ctx, conn, accountID, clusterID, databaseName, partnerName)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift Partner (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Partner (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrAccountID, accountID)
	d.Set(names.AttrClusterIdentifier, clusterID)
	d.Set(names.AttrDatabaseName, out.DatabaseName)
	d.Set("partner_name", out.PartnerName)
	d.Set(names.AttrStatus, out.Status)
	d.Set(names.AttrStatusMessage, out.StatusMessage)

	return diags
}

func resourcePartnerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	accountID, clusterID, databaseName, partnerName, err := decodePartnerID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	log.Printf("[DEBUG] Deleting Redshift Partner: %s", d.Id())
	input := redshift.DeletePartnerInput{
		AccountId:         aws.String(accountID),
		ClusterIdentifier: aws.String(clusterID),
		DatabaseName:      aws.String(databaseName),
		PartnerName:       aws.String(partnerName),
	}
	_, err = conn.DeletePartner(ctx, &input)

	if errs.IsA[*awstypes.PartnerNotFoundFault](err) || errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Partner (%s): %s", d.Id(), err)
	}

	return diags
}

func decodePartnerID(id string) (string, string, string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		return "", "", "", "", fmt.Errorf("expected ID to be the form account:cluster_identifier:database_name:partner_name, given: %s", id)
	}

	return idParts[0], idParts[1], idParts[2], idParts[3], nil
}
