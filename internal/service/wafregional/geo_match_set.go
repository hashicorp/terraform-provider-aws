// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
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
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafregional_geo_match_set", name="Geo Match Set")
func resourceGeoMatchSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceGeoMatchSetCreate,
		ReadWithoutTimeout:   resourceGeoMatchSetRead,
		UpdateWithoutTimeout: resourceGeoMatchSetUpdate,
		DeleteWithoutTimeout: resourceGeoMatchSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"geo_match_constraint": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:     schema.TypeString,
							Required: true,
						},
						names.AttrValue: {
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

	name := d.Get(names.AttrName).(string)
	outputRaw, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateGeoMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateGeoMatchSetWithContext(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Geo Match Set (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(outputRaw.(*waf.CreateGeoMatchSetOutput).GeoMatchSet.GeoMatchSetId))

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

	d.Set(names.AttrName, resp.GeoMatchSet.Name)
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

		if err := updateGeoMatchSetResourceWR(ctx, conn, region, d.Id(), oldConstraints, newConstraints); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Geo Match Set (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceGeoMatchSetRead(ctx, d, meta)...)
}

func resourceGeoMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldConstraints := d.Get("geo_match_constraint").(*schema.Set).List(); len(oldConstraints) > 0 {
		var newConstraints []interface{}

		err := updateGeoMatchSetResourceWR(ctx, conn, region, d.Id(), oldConstraints, newConstraints)

		if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentContainerException, wafregional.ErrCodeWAFNonexistentItemException) {
			return diags
		}

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional Geo Match Set (%s): %s", d.Id(), err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Geo Match Set: %s", d.Id())
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteGeoMatchSetWithContext(ctx, input)
	})

	if tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Geo Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func updateGeoMatchSetResourceWR(ctx context.Context, conn *wafregional.WAFRegional, region string, geoMatchSetID string, oldConstraints, newConstraints []interface{}) error {
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(geoMatchSetID),
			Updates:       tfwaf.DiffGeoMatchSetConstraints(oldConstraints, newConstraints),
		}

		return conn.UpdateGeoMatchSetWithContext(ctx, input)
	})

	return err
}
