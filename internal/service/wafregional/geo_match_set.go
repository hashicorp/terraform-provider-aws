// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGeoMatchSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	outputRaw, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
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

	geoMatchSet, err := findGeoMatchSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regional Geo Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Regional Geo Match Set (%s): %s", d.Id(), err)
	}

	if err := d.Set("geo_match_constraint", flattenGeoMatchConstraint(geoMatchSet.GeoMatchConstraints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting geo_match_constraint: %s", err)
	}
	d.Set(names.AttrName, geoMatchSet.Name)

	return diags
}

func resourceGeoMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("geo_match_constraint") {
		o, n := d.GetChange("geo_match_constraint")
		oldConstraints, newConstraints := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateGeoMatchSet(ctx, conn, region, d.Id(), oldConstraints, newConstraints); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
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
		if err := updateGeoMatchSet(ctx, conn, region, d.Id(), oldConstraints, newConstraints); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regional Geo Match Set: %s", d.Id())
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
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

func findGeoMatchSetByID(ctx context.Context, conn *wafregional.Client, id string) (*awstypes.GeoMatchSet, error) {
	input := &wafregional.GetGeoMatchSetInput{
		GeoMatchSetId: aws.String(id),
	}

	output, err := conn.GetGeoMatchSet(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.GeoMatchSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.GeoMatchSet, nil
}

func updateGeoMatchSet(ctx context.Context, conn *wafregional.Client, region string, geoMatchSetID string, oldConstraints, newConstraints []interface{}) error {
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.UpdateGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(geoMatchSetID),
			Updates:       diffGeoMatchSetConstraints(oldConstraints, newConstraints),
		}

		return conn.UpdateGeoMatchSet(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF Regional Geo Match Set (%s): %w", geoMatchSetID, err)
	}

	return nil
}

func flattenGeoMatchConstraint(ts []awstypes.GeoMatchConstraint) []interface{} {
	out := make([]interface{}, len(ts))
	for i, t := range ts {
		m := make(map[string]interface{})
		m[names.AttrType] = string(t.Type)
		m[names.AttrValue] = string(t.Value)
		out[i] = m
	}
	return out
}

func diffGeoMatchSetConstraints(oldT, newT []interface{}) []awstypes.GeoMatchSetUpdate {
	updates := make([]awstypes.GeoMatchSetUpdate, 0)

	for _, od := range oldT {
		constraint := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newT, constraint); contains {
			newT = append(newT[:idx], newT[idx+1:]...)
			continue
		}

		updates = append(updates, awstypes.GeoMatchSetUpdate{
			Action: awstypes.ChangeActionDelete,
			GeoMatchConstraint: &awstypes.GeoMatchConstraint{
				Type:  awstypes.GeoMatchConstraintType(constraint[names.AttrType].(string)),
				Value: awstypes.GeoMatchConstraintValue(constraint[names.AttrValue].(string)),
			},
		})
	}

	for _, nd := range newT {
		constraint := nd.(map[string]interface{})

		updates = append(updates, awstypes.GeoMatchSetUpdate{
			Action: awstypes.ChangeActionInsert,
			GeoMatchConstraint: &awstypes.GeoMatchConstraint{
				Type:  awstypes.GeoMatchConstraintType(constraint[names.AttrType].(string)),
				Value: awstypes.GeoMatchConstraintValue(constraint[names.AttrValue].(string)),
			},
		})
	}
	return updates
}
