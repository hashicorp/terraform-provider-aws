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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

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
			names.AttrName: {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceIPSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &wafregional.CreateIPSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateIPSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional IPSet (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*wafregional.CreateIPSetOutput).IPSet.IPSetId))

	return append(diags, resourceIPSetUpdate(ctx, d, meta)...)
}

func resourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)

	ipSet, err := findIPSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF Regional IPSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF Regional (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf-regional",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "ipset/" + d.Id(),
	}
	d.Set(names.AttrARN, arn.String())
	if err := d.Set("ip_set_descriptor", flattenIPSetDescriptors(ipSet.IPSetDescriptors)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_set_descriptor: %s", err)
	}
	d.Set(names.AttrName, ipSet.Name)

	return diags
}

func resourceIPSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalClient(ctx)
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("ip_set_descriptor") {
		o, n := d.GetChange("ip_set_descriptor")
		oldD, newD := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateIPSet(ctx, conn, region, d.Id(), oldD, newD); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
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
		if err := updateIPSet(ctx, conn, region, d.Id(), oldD, newD); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF Regional IPSet: %s", d.Id())
	_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
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

func findIPSetByID(ctx context.Context, conn *wafregional.Client, id string) (*awstypes.IPSet, error) {
	input := &wafregional.GetIPSetInput{
		IPSetId: aws.String(id),
	}

	output, err := conn.GetIPSet(ctx, input)

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.IPSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.IPSet, nil
}

func updateIPSet(ctx context.Context, conn *wafregional.Client, region, ipSetID string, oldD, newD []interface{}) error {
	for _, ipSetUpdates := range diffIPSetDescriptors(oldD, newD) {
		_, err := newRetryer(conn, region).RetryWithToken(ctx, func(token *string) (interface{}, error) {
			input := &wafregional.UpdateIPSetInput{
				ChangeToken: token,
				IPSetId:     aws.String(ipSetID),
				Updates:     ipSetUpdates,
			}

			return conn.UpdateIPSet(ctx, input)
		})

		if err != nil {
			return fmt.Errorf("updating WAF Regional IPSet (%s): %w", ipSetID, err)
		}
	}

	return nil
}

func flattenIPSetDescriptors(in []awstypes.IPSetDescriptor) []interface{} {
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

func diffIPSetDescriptors(oldD, newD []interface{}) [][]awstypes.IPSetUpdate {
	// WAF requires UpdateIPSet operations be split into batches of 1000 Updates
	const (
		ipSetUpdatesLimit = 1000
	)
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
