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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_geo_match_set", name="GeoMatchSet")
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
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateGeoMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateGeoMatchSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF GeoMatchSet (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*waf.CreateGeoMatchSetOutput).GeoMatchSet.GeoMatchSetId))

	return append(diags, resourceGeoMatchSetUpdate(ctx, d, meta)...)
}

func resourceGeoMatchSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	geoMatchSet, err := findGeoMatchSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF GeoMatchSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF GeoMatchSet (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "geomatchset/" + d.Id(),
	}
	d.Set(names.AttrARN, arn.String())
	if err := d.Set("geo_match_constraint", flattenGeoMatchConstraint(geoMatchSet.GeoMatchConstraints)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting geo_match_constraint: %s", err)
	}
	d.Set(names.AttrName, geoMatchSet.Name)

	return diags
}

func resourceGeoMatchSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("geo_match_constraint") {
		o, n := d.GetChange("geo_match_constraint")
		oldT, newT := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateGeoMatchSet(ctx, conn, d.Id(), oldT, newT); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceGeoMatchSetRead(ctx, d, meta)...)
}

func resourceGeoMatchSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if oldConstraints := d.Get("geo_match_constraint").(*schema.Set).List(); len(oldConstraints) > 0 {
		noConstraints := []interface{}{}
		if err := updateGeoMatchSet(ctx, conn, d.Id(), oldConstraints, noConstraints); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF GeoMatchSet: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteGeoMatchSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF GeoMatchSet (%s): %s", d.Id(), err)
	}

	return diags
}

func findGeoMatchSetByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.GeoMatchSet, error) {
	input := &waf.GetGeoMatchSetInput{
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

func updateGeoMatchSet(ctx context.Context, conn *waf.Client, id string, oldT, newT []interface{}) error {
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.UpdateGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(id),
			Updates:       diffGeoMatchSetConstraints(oldT, newT),
		}

		return conn.UpdateGeoMatchSet(ctx, input)
	})

	if err != nil {
		return fmt.Errorf("updating WAF GeoMatchSet (%s): %w", id, err)
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
