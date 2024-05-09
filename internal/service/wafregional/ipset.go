// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafregional

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/wafregional"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafregional/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// WAF requires UpdateIPSet operations be split into batches of 1000 Updates
const ipSetUpdatesLimit = 1000

// @SDKResource("aws_wafregional_ipset", name="IPSet")
func resourceIPSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPSetCreate,
		ReadWithoutTimeout:   resourceIPSetRead,
		UpdateWithoutTimeout: resourceIPSetUpdate,
		DeleteWithoutTimeout: resourceIPSetDelete,

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
			"ip_set_descriptor": {
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

func resourceIPSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &wafregional.CreateIPSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}
		return conn.CreateIPSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional IPSet: %s", err)
	}
	resp := out.(*wafregional.CreateIPSetOutput)
	d.SetId(aws.ToString(resp.IPSet.IPSetId))
	return append(diags, resourceIPSetUpdate(ctx, d, meta)...)
}

func resourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	params := &wafregional.GetIPSetInput{
		IPSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetIPSet(ctx, params)
	if err != nil {
		if !d.IsNewResource() && errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] WAF Regional IPSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF Regional IPSet: %s", err)
	}

	d.Set("ip_set_descriptor", flattenIPSetDescriptorWR(resp.IPSet.IPSetDescriptors))
	d.Set(names.AttrName, resp.IPSet.Name)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf-regional",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("ipset/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	return diags
}

func flattenIPSetDescriptorWR(in []awstypes.IPSetDescriptor) []interface{} {
	descriptors := make([]interface{}, len(in))

	for i, descriptor := range in {
		d := map[string]interface{}{
			names.AttrType:  descriptor.Type,
			names.AttrValue: *descriptor.Value,
		}
		descriptors[i] = d
	}

	return descriptors
}

func resourceIPSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("ip_set_descriptor") {
		o, n := d.GetChange("ip_set_descriptor")
		oldD, newD := o.(*schema.Set).List(), n.(*schema.Set).List()

		if err := updateIPSetResourceWR(ctx, conn, region, d.Id(), oldD, newD); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional IPSet (%s): %s", d.Id(), err)
		}
	}
	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if oldD := d.Get("ip_set_descriptor").(*schema.Set).List(); len(oldD) > 0 {
		var newD []interface{}
		err := updateIPSetResourceWR(ctx, conn, region, d.Id(), oldD, newD)
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return sdkdiag.AppendErrorf(diags, "updating WAF Regional IPSet (%s): %s", d.Id(), err)
			}
		}
	}

	log.Printf("[INFO] Deleting WAF Regional IPSet: %s", d.Id())
	_, err := NewRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.DeleteIPSetInput{
			ChangeToken: token,
			IPSetId:     aws.String(d.Id()),
		}

		return conn.DeleteIPSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional IPSet (%s): %s", d.Id(), err)
	}

	return diags
}

func updateIPSetResourceWR(ctx context.Context, conn *wafregional.Client, region, ipSetID string, oldD, newD []interface{}) error {
	wr := NewRetryer(conn, region)

	for _, ipSetUpdates := range DiffIPSetDescriptors(oldD, newD) {
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			input := &wafregional.UpdateIPSetInput{
				ChangeToken: token,
				IPSetId:     aws.String(ipSetID),
				Updates:     ipSetUpdates,
			}

			return conn.UpdateIPSet(ctx, input)
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func DiffIPSetDescriptors(oldD, newD []interface{}) [][]awstypes.IPSetUpdate {
	updates := make([]awstypes.IPSetUpdate, 0, ipSetUpdatesLimit)
	updatesBatches := make([][]awstypes.IPSetUpdate, 0)

	for _, od := range oldD {
		descriptor := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newD, descriptor); contains {
			newD = append(newD[:idx], newD[idx+1:]...)
			continue
		}

		if len(updates) == ipSetUpdatesLimit {
			updatesBatches = append(updatesBatches, updates)
			updates = make([]awstypes.IPSetUpdate, 0, ipSetUpdatesLimit)
		}

		updates = append(updates, awstypes.IPSetUpdate{
			Action: awstypes.ChangeActionDelete,
			IPSetDescriptor: &awstypes.IPSetDescriptor{
				Type:  awstypes.IPSetDescriptorType(descriptor[names.AttrType].(string)),
				Value: aws.String(descriptor[names.AttrValue].(string)),
			},
		})
	}

	for _, nd := range newD {
		descriptor := nd.(map[string]interface{})

		if len(updates) == ipSetUpdatesLimit {
			updatesBatches = append(updatesBatches, updates)
			updates = make([]awstypes.IPSetUpdate, 0, ipSetUpdatesLimit)
		}

		updates = append(updates, awstypes.IPSetUpdate{
			Action: awstypes.ChangeActionInsert,
			IPSetDescriptor: &awstypes.IPSetDescriptor{
				Type:  awstypes.IPSetDescriptorType(descriptor[names.AttrType].(string)),
				Value: aws.String(descriptor[names.AttrValue].(string)),
			},
		})
	}
	updatesBatches = append(updatesBatches, updates)
	return updatesBatches
}
