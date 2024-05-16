// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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

func resourceEndpointAuthorizationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	account := d.Get("account").(string)
	input := redshift.AuthorizeEndpointAccessInput{
		Account:           aws.String(account),
		ClusterIdentifier: aws.String(d.Get(names.AttrClusterIdentifier).(string)),
	}

	if v, ok := d.GetOk("vpc_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.VpcIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	output, err := conn.AuthorizeEndpointAccessWithContext(ctx, &input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Redshift Endpoint Authorization: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", account, aws.StringValue(output.ClusterIdentifier)))
	log.Printf("[INFO] Redshift Endpoint Authorization ID: %s", d.Id())

	return append(diags, resourceEndpointAuthorizationRead(ctx, d, meta)...)
}

func resourceEndpointAuthorizationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	endpoint, err := findEndpointAuthorizationByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
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
	d.Set("vpc_ids", flex.FlattenStringSet(endpoint.AllowedVPCs))
	d.Set("allowed_all_vpcs", endpoint.AllowedAllVPCs)
	d.Set("endpoint_count", endpoint.EndpointCount)

	return diags
}

func resourceEndpointAuthorizationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	if d.HasChanges("vpc_ids") {
		account, clusterId, err := DecodeEndpointAuthorizationID(d.Id())
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Redshift Endpoint Authorization (%s): %s", d.Id(), err)
		}

		o, n := d.GetChange("vpc_ids")
		ns := n.(*schema.Set)
		os := o.(*schema.Set)
		added := ns.Difference(os)
		removed := os.Difference(ns)

		if added.Len() > 0 {
			_, err := conn.AuthorizeEndpointAccessWithContext(ctx, &redshift.AuthorizeEndpointAccessInput{
				Account:           aws.String(account),
				ClusterIdentifier: aws.String(clusterId),
				VpcIds:            flex.ExpandStringSet(added),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Redshift Endpoint Authorization (%s): authorizing VPCs: %s", d.Id(), err)
			}
		}

		if removed.Len() > 0 {
			_, err := conn.RevokeEndpointAccessWithContext(ctx, &redshift.RevokeEndpointAccessInput{
				Account:           aws.String(account),
				ClusterIdentifier: aws.String(clusterId),
				VpcIds:            flex.ExpandStringSet(removed),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "updating Redshift Endpoint Authorization (%s): revoking VPCs: %s", d.Id(), err)
			}
		}
	}

	return append(diags, resourceEndpointAuthorizationRead(ctx, d, meta)...)
}

func resourceEndpointAuthorizationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RedshiftConn(ctx)

	account, clusterId, err := DecodeEndpointAuthorizationID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Endpoint Authorization (%s): %s", d.Id(), err)
	}

	input := &redshift.RevokeEndpointAccessInput{
		Account:           aws.String(account),
		ClusterIdentifier: aws.String(clusterId),
		Force:             aws.Bool(d.Get(names.AttrForceDelete).(bool)),
	}

	_, err = conn.RevokeEndpointAccessWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, redshift.ErrCodeEndpointAuthorizationNotFoundFault) {
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "deleting Redshift Endpoint Authorization (%s): %s", d.Id(), err)
	}

	return diags
}

func DecodeEndpointAuthorizationID(id string) (string, string, error) {
	idParts := strings.Split(id, ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID to be the form account:cluster_identifier, given: %s", id)
	}

	return idParts[0], idParts[1], nil
}
