// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
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
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	outputRaw, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.CreateGeoMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateGeoMatchSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional Geo Match Set (%s): %s", name, err)
	}

	d.SetId(aws.ToString(outputRaw.(*wafregional.CreateGeoMatchSetOutput).GeoMatchSet.GeoMatchSetId))

	return append(diags, resourceGeoMatchSetUpdate(ctx, d, meta)...)
}

func resourceGeoMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	params := &wafregional.GetGeoMatchSetInput{
		GeoMatchSetId: aws.String(d.Id()),
	}
	resp, err := conn.GetGeoMatchSet(ctx, params)

	if !d.IsNewResource() && errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		log.Printf("[WARN] WAF WAF Regional Geo Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "getting WAF Regional Geo Match Set (%s): %s", d.Id(), err)
	}

	d.Set(names.AttrName, resp.GeoMatchSet.Name)
	d.Set("geo_match_constraint", FlattenGeoMatchConstraint(resp.GeoMatchSet.GeoMatchConstraints))

	return diags
}

func resourceGeoMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
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
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldConstraints := d.Get("geo_match_constraint").(*schema.Set).List(); len(oldConstraints) > 0 {
		var newConstraints []interface{}
		err := updateGeoMatchSetResourceWR(ctx, conn, region, d.Id(), oldConstraints, newConstraints)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF Regional Geo Match Set (%s): %s", d.Id(), err)
			}
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Geo Match Set: %s", d.Id())
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.DeleteGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteGeoMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional Geo Match Set (%s): %s", d.Id(), err)
	}

	return diags
}

func updateGeoMatchSetResourceWR(ctx context.Context, conn *wafregional.Client, region string, geoMatchSetID string, oldConstraints, newConstraints []interface{}) error {
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.UpdateGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(geoMatchSetID),
			Updates:       DiffGeoMatchSetConstraints(oldConstraints, newConstraints),
		}

		return conn.UpdateGeoMatchSet(ctx, input)
	})

	return err
}
