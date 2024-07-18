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
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_waf_ipset", name="IPSet")
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
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	name := d.Get(names.AttrName).(string)
	output, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.CreateIPSetInput{
			ChangeToken: token,
			Name:        aws.String(name),
		}

		return conn.CreateIPSet(ctx, input)
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF IPSet (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.(*waf.CreateIPSetOutput).IPSet.IPSetId))

	if v, ok := d.GetOk("ip_set_descriptors"); ok && v.(*schema.Set).Len() > 0 {
		if err := updateIPSet(ctx, conn, d.Id(), nil, v.(*schema.Set).List()); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	ipSet, err := findIPSetByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAF IPSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAF IPSet (%s): %s", d.Id(), err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  "ipset/" + d.Id(),
	}
	d.Set(names.AttrARN, arn.String())

	var descriptors []map[string]interface{}

	for _, descriptor := range ipSet.IPSetDescriptors {
		d := map[string]interface{}{
			names.AttrType:  string(descriptor.Type),
			names.AttrValue: aws.ToString(descriptor.Value),
		}
		descriptors = append(descriptors, d)
	}

	if err := d.Set("ip_set_descriptors", descriptors); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting ip_set_descriptors: %s", err)
	}
	d.Set(names.AttrName, ipSet.Name)

	return diags
}

func resourceIPSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if d.HasChange("ip_set_descriptors") {
		o, n := d.GetChange("ip_set_descriptors")
		oldD, newD := o.(*schema.Set).List(), n.(*schema.Set).List()
		if err := updateIPSet(ctx, conn, d.Id(), oldD, newD); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFClient(ctx)

	if oldDescriptors := d.Get("ip_set_descriptors").(*schema.Set).List(); len(oldDescriptors) > 0 {
		if err := updateIPSet(ctx, conn, d.Id(), oldDescriptors, nil); err != nil && !errs.IsA[*awstypes.WAFNonexistentItemException](err) && !errs.IsA[*awstypes.WAFNonexistentContainerException](err) {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	log.Printf("[INFO] Deleting WAF IPSet: %s", d.Id())
	_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
		input := &waf.DeleteIPSetInput{
			ChangeToken: token,
			IPSetId:     aws.String(d.Id()),
		}

		return conn.DeleteIPSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF IPSet (%s): %s", d.Id(), err)
	}

	return diags
}

func findIPSetByID(ctx context.Context, conn *waf.Client, id string) (*awstypes.IPSet, error) {
	input := &waf.GetIPSetInput{
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

func updateIPSet(ctx context.Context, conn *waf.Client, id string, oldD, newD []interface{}) error {
	for _, ipSetUpdates := range diffIPSetDescriptors(oldD, newD) {
		_, err := newRetryer(conn).RetryWithToken(ctx, func(token *string) (interface{}, error) {
			input := &waf.UpdateIPSetInput{
				ChangeToken: token,
				IPSetId:     aws.String(id),
				Updates:     ipSetUpdates,
			}

			return conn.UpdateIPSet(ctx, input)
		})

		if err != nil {
			return fmt.Errorf("updating WAF IPSet (%s): %w", id, err)
		}
	}

	return nil
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
