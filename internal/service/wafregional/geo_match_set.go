// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tfwaf "github.com/hashicorp/terraform-provider-aws/internal/service/waf"
)

// @SDKResource("aws_wafregional_geo_match_set")
func ResourceGeoMatchSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGeoMatchSetCreate,
		ReadWithoutTimeout:   resourceGeoMatchSetRead,
		UpdateWithoutTimeout: resourceGeoMatchSetUpdate,
		DeleteWithoutTimeout: resourceGeoMatchSetDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"geo_match_constraint": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceGeoMatchSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateGeoMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}

		return conn.CreateGeoMatchSetWithContext(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Geo Match Set: %s", err)
	}
	resp := out.(*waf.CreateGeoMatchSetOutput)

	d.SetId(aws.StringValue(resp.GeoMatchSet.GeoMatchSetId))

	return append(diags, resourceGeoMatchSetUpdate(ctx, d, meta)...)
}

func resourceGeoMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)

	params := &waf.GetGeoMatchSetInput{
		GeoMatchSetId: aws.String(d.Id()),
	}
	resp, err := conn.GetGeoMatchSetWithContext(ctx, params)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		log.Printf("[WARN] WAF WAF Regional Geo Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting WAF Regional Geo Match Set (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.GeoMatchSet.Name)
	d.Set("geo_match_constraint", tfwaf.FlattenGeoMatchConstraint(resp.GeoMatchSet.GeoMatchConstraints))

	return diags
}

func resourceGeoMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("geo_match_constraint") {
		o, n := d.GetChange("geo_match_constraint")
		oldConstraints, newConstraints := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateGeoMatchSetResourceWR(ctx, d.Id(), oldConstraints, newConstraints, conn, region)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Geo Match Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGeoMatchSetRead(ctx, d, meta)...)
}

func resourceGeoMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	oldConstraints := d.Get("geo_match_constraint").(*schema.Set).List()
	if len(oldConstraints) > 0 {
		noConstraints := []interface{}{}
		err := updateGeoMatchSetResourceWR(ctx, d.Id(), oldConstraints, noConstraints, conn, region)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Geo Match Constraint (%s): %s", d.Id(), err)
		}
	}

	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteGeoMatchSetWithContext(ctx, req)
	})
	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Geo Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func updateGeoMatchSetResourceWR(ctx context.Context, id string, oldConstraints, newConstraints []interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(id),
			Updates:       tfwaf.DiffGeoMatchSetConstraints(oldConstraints, newConstraints),
		}

		return conn.UpdateGeoMatchSetWithContext(ctx, req)
	})
	if err != nil {
		return fmt.Errorf("failed updating WAF Regional Geo Match Set (%s): %w", id, err)
	}

	return nil
}
