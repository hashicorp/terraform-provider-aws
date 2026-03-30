// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

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
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_redshift_endpoint_authorization", name="Endpoint Authorization")
func resourceEndpointAuthorization() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceEndpointAuthorizationCreate,
		ReadWithoutTimeout:   resourceEndpointAuthorizationRead,
		UpdateWithoutTimeout: resourceEndpointAuthorizationUpdate,
		DeleteWithoutTimeout: resourceEndpointAuthorizationDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account": {
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"allowed_all_vpcs": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			names.AttrClusterIdentifier: {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"endpoint_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			names.AttrForceDelete: {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"grantee": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"grantor": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceEndpointAuthorizationCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	accountID, clusterID := d.Get("account").(string), d.Get(names.AttrClusterIdentifier).(string)
	id := fmt.Sprintf("%s:%s", accountID, clusterID)
	input := redshift.AuthorizeEndpointAccessInput{
		Account:           aws.String(accountID),
		ClusterIdentifier: aws.String(clusterID),
	}

	if v, ok := d.GetOk("vpc_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.VpcIds = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	_, err := conn.AuthorizeEndpointAccess(ctx, &input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Endpoint Authorization (%s): %s", id, err)
	}

	d.SetId(id)

	return append(diags, resourceEndpointAuthorizationRead(ctx, d, meta)...)
}

func resourceEndpointAuthorizationRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	accountID, clusterID, err := decodeEndpointAuthorizationID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	endpoint, err := findEndpointAuthorizationByTwoPartKey(ctx, conn, accountID, clusterID)

	if !d.IsNewResource() && retry.NotFound(err) {
		log.Printf("[WARN] Redshift Endpoint Authorization (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Redshift Endpoint Authorization (%s): %s", d.Id(), err)
	}

	d.Set("account", endpoint.Grantee)
	d.Set("grantee", endpoint.Grantee)
	d.Set("grantor", endpoint.Grantor)
	d.Set(names.AttrClusterIdentifier, endpoint.ClusterIdentifier)
	d.Set("vpc_ids", endpoint.AllowedVPCs)
	d.Set("allowed_all_vpcs", endpoint.AllowedAllVPCs)
	d.Set("endpoint_count", endpoint.EndpointCount)

	return diags
}

func resourceEndpointAuthorizationUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	if d.HasChanges("vpc_ids") {
		accountID, clusterID, err := decodeEndpointAuthorizationID(d.Id())
		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		o, n := d.GetChange("vpc_ids")
		ns, os := n.(*schema.Set), o.(*schema.Set)
		add, del := ns.Difference(os), os.Difference(ns)

		if add.Len() > 0 {
			input := redshift.AuthorizeEndpointAccessInput{
				Account:           aws.String(accountID),
				ClusterIdentifier: aws.String(clusterID),
				VpcIds:            flex.ExpandStringValueSet(add),
			}
			_, err := conn.AuthorizeEndpointAccess(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Redshift Endpoint Authorization (%s): authorizing VPCs: %s", d.Id(), err)
			}
		}

		if del.Len() > 0 {
			input := redshift.RevokeEndpointAccessInput{
				Account:           aws.String(accountID),
				ClusterIdentifier: aws.String(clusterID),
				VpcIds:            flex.ExpandStringValueSet(del),
			}
			_, err := conn.RevokeEndpointAccess(ctx, &input)

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Redshift Endpoint Authorization (%s): revoking VPCs: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceEndpointAuthorizationRead(ctx, d, meta)...)
}

func resourceEndpointAuthorizationDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftClient(ctx)

	accountID, clusterID, err := decodeEndpointAuthorizationID(d.Id())
	if err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}
	input := redshift.RevokeEndpointAccessInput{
		Account:           aws.String(accountID),
		ClusterIdentifier: aws.String(clusterID),
		Force:             aws.Bool(d.Get(names.AttrForceDelete).(bool)),
	}
	_, err = conn.RevokeEndpointAccess(ctx, &input)

	if errs.IsA[*awstypes.EndpointAuthorizationNotFoundFault](err) || errs.IsA[*awstypes.ClusterNotFoundFault](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Endpoint Authorization (%s): %s", d.Id(), err)
	}

	return diags
}

func decodeEndpointAuthorizationID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID to be the form account:cluster_identifier, given: %s", id)
	}

	return idParts[0], idParts[1], nil
}
