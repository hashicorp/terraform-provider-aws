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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// WAF requires UpdateIPSet operations be split into batches of 1000 Updates
const ipSetUpdatesLimit = 1000

// @SDKResource("aws_waf_ipset")
func ResourceIPSet() *schema.Resource {
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
			"ip_set_descriptors": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						names.AttrType: {
							Type:             schema.TypeString,
							Required:         true,
							ValidateDiagFunc: enum.Validate[awstypes.IPSetDescriptorType](),
						},
						names.AttrValue: {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
					},
				},
			},
		},
	}
}

func resourceIPSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateIPSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get(names.AttrName).(string)),
		}
		return conn.CreateIPSet(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF IPSet (%s): %s", d.Get(names.AttrName).(string), err)
	}
	resp := out.(*waf.CreateIPSetOutput)
	d.SetId(aws.ToString(resp.IPSet.IPSetId))

	if v, ok := d.GetOk("ip_set_descriptors"); ok && v.(*schema.Set).Len() > 0 {
		err := updateIPSetDescriptors(ctx, d.Id(), nil, v.(*schema.Set).List(), conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "Setting IP Descriptors: %s", err)
		}
	}

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	params := &waf.GetIPSetInput{
		IPSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetIPSet(ctx, params)
	if err != nil {
		if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
			log.Printf("[WARN] WAF IPSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF IPSet (%s): %s", d.Get(names.AttrName).(string), err)
	}

	var descriptors []map[string]interface{}

	for _, descriptor := range resp.IPSet.IPSetDescriptors {
		d := map[string]interface{}{
			names.AttrType:  string(descriptor.Type),
			names.AttrValue: aws.ToString(descriptor.Value),
		}
		descriptors = append(descriptors, d)
	}

	d.Set("ip_set_descriptors", descriptors)

	d.Set(names.AttrName, resp.IPSet.Name)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("ipset/%s", d.Id()),
	}
	d.Set(names.AttrARN, arn.String())

	return diags
}

func resourceIPSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("ip_set_descriptors") {
		o, n := d.GetChange("ip_set_descriptors")
		oldD, newD := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateIPSetDescriptors(ctx, d.Id(), oldD, newD, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF IPSet: %s", err)
		}
	}

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	oldDescriptors := d.Get("ip_set_descriptors").(*schema.Set).List()

	if len(oldDescriptors) > 0 {
		err := updateIPSetDescriptors(ctx, d.Id(), oldDescriptors, nil, conn)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting IPSetDescriptors: %s", err)
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteIPSetInput{
			ChangeToken: token,
			IPSetId:     aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF IPSet")
		return conn.DeleteIPSet(ctx, req)
	})
	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF IPSet: %s", err)
	}

	return diags
}

func updateIPSetDescriptors(ctx context.Context, id string, oldD, newD []interface{}, conn *waf.Client) error {
	for _, ipSetUpdates := range DiffIPSetDescriptors(oldD, newD) {
		wr := NewRetryer(conn)
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			req := &waf.UpdateIPSetInput{
				ChangeToken: token,
				IPSetId:     aws.String(id),
				Updates:     ipSetUpdates,
			}
			log.Printf("[INFO] Updating IPSet descriptors: %s", id)
			return conn.UpdateIPSet(ctx, req)
		})
		if err != nil {
			if !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
				return fmt.Errorf("updating WAF IPSet: %s", err)
			}
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
