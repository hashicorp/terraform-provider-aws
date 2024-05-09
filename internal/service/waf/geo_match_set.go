// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package waf

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/waf"
	awstypes "github.com/aws/aws-sdk-go-v2/service/waf/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_geo_match_set")
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			names.AttrARN: {
				Type:     schema.TypeString,
				Computed: true,
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
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	log.Printf("[INFO] Creating GeoMatchSet: %s", d.Get(names.AttrName).(string))

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateGeoMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}

		return conn.CreateGeoMatchSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating GeoMatchSet: %s", err)
	}
	resp := out.(*waf.CreateGeoMatchSetOutput)

	d.SetId(aws.ToString(resp.GeoMatchSet.GeoMatchSetId))

	return append(diags, resourceGeoMatchSetUpdate(ctx, d, meta)...)
}

func resourceGeoMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)
	log.Printf("[INFO] Reading GeoMatchSet: %s", d.Get(names.AttrName).(string))
	params := &waf.GetGeoMatchSetInput{
		GeoMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetGeoMatchSet(ctx, params)
	if err != nil {
		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] WAF GeoMatchSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF GeoMatchSet: %s", err)
	}

	d.Set(names.AttrName, resp.GeoMatchSet.Name)
	d.Set("geo_match_constraint", FlattenGeoMatchConstraint(resp.GeoMatchSet.GeoMatchConstraints))

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("geomatchset/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	return diags
}

func resourceGeoMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("geo_match_constraint") {
		o, n := d.GetChange("geo_match_constraint")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateGeoMatchSetResource(ctx, d.Id(), oldT, newT, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating GeoMatchSet: %s", err)
		}
	}

	return append(diags, resourceGeoMatchSetRead(ctx, d, meta)...)
}

func resourceGeoMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	oldConstraints := d.Get("geo_match_constraint").(*schema.Set).List()
	if len(oldConstraints) > 0 {
		noConstraints := []interface{}{}
		err := updateGeoMatchSetResource(ctx, d.Id(), oldConstraints, noConstraints, conn)
		if err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendErrorf(diags, "updating GeoMatchConstraint: %s", err)
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteGeoMatchSet(ctx, req)
	})
	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting GeoMatchSet: %s", err)
	}

	return diags
}

func updateGeoMatchSetResource(ctx context.Context, id string, oldT, newT []interface{}, conn *waf.Client) error {
	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.UpdateGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(id),
			Updates:       DiffGeoMatchSetConstraints(oldT, newT),
		}

		log.Printf("[INFO] Updating GeoMatchSet constraints: %s", id)
		return conn.UpdateGeoMatchSet(ctx, req)
	})
	return err
}
