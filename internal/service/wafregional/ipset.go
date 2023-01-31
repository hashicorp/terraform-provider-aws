package wafregional

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
)

// WAF requires UpdateIPSet operations be split into batches of 1000 Updates
const ipSetUpdatesLimit = 1000

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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_set_descriptor": {
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

func resourceIPSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn()
	region := meta.(*conns.AWSClient).Region

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		params := &waf.CreateIPSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}
		return conn.CreateIPSetWithContext(ctx, params)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAF Regional IPSet: %s", err)
	}
	resp := out.(*waf.CreateIPSetOutput)
	d.SetId(aws.StringValue(resp.IPSet.IPSetId))
	return append(diags, resourceIPSetUpdate(ctx, d, meta)...)
}

func resourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn()

	params := &waf.GetIPSetInput{
		IPSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetIPSetWithContext(ctx, params)
	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, wafregional.ErrCodeWAFNonexistentItemException) {
			log.Printf("[WARN] WAF Regional IPSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "reading WAF Regional IPSet: %s", err)
	}

	d.Set("ip_set_descriptor", flattenIPSetDescriptorWR(resp.IPSet.IPSetDescriptors))
	d.Set("name", resp.IPSet.Name)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf-regional",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("ipset/%s", d.Id()),
	}
	d.Set("arn", arn.String())

	return diags
}

func flattenIPSetDescriptorWR(in []*waf.IPSetDescriptor) []interface{} {
	descriptors := make([]interface{}, len(in))

	for i, descriptor := range in {
		d := map[string]interface{}{
			"type":  *descriptor.Type,
			"value": *descriptor.Value,
		}
		descriptors[i] = d
	}

	return descriptors
}

func resourceIPSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn()
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("ip_set_descriptor") {
		o, n := d.GetChange("ip_set_descriptor")
		oldD, newD := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateIPSetResourceWR(ctx, d.Id(), oldD, newD, conn, region)
		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAF Regional IPSet: %s", err)
		}
	}
	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFRegionalConn()
	region := meta.(*conns.AWSClient).Region

	oldD := d.Get("ip_set_descriptor").(*schema.Set).List()

	if len(oldD) > 0 {
		noD := []interface{}{}
		err := updateIPSetResourceWR(ctx, d.Id(), oldD, noD, conn, region)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "deleting IPSetDescriptors: %s", err)
		}
	}

	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
		req := &waf.DeleteIPSetInput{
			ChangeToken: token,
			IPSetId:     aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF Regional IPSet")
		return conn.DeleteIPSetWithContext(ctx, req)
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAF Regional IPSet: %s", err)
	}

	return diags
}

func updateIPSetResourceWR(ctx context.Context, id string, oldD, newD []interface{}, conn *wafregional.WAFRegional, region string) error {
	for _, ipSetUpdates := range DiffIPSetDescriptors(oldD, newD) {
		wr := NewRetryer(conn, region)
		_, err := wr.RetryWithToken(ctx, func(token *string) (interface{}, error) {
			req := &waf.UpdateIPSetInput{
				ChangeToken: token,
				IPSetId:     aws.String(id),
				Updates:     ipSetUpdates,
			}

			return conn.UpdateIPSetWithContext(ctx, req)
		})
		if err != nil {
			return fmt.Errorf("updating WAF Regional IPSet: %s", err)
		}
	}

	return nil
}

func DiffIPSetDescriptors(oldD, newD []interface{}) [][]*waf.IPSetUpdate {
	updates := make([]*waf.IPSetUpdate, 0, ipSetUpdatesLimit)
	updatesBatches := make([][]*waf.IPSetUpdate, 0)

	for _, od := range oldD {
		descriptor := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newD, descriptor); contains {
			newD = append(newD[:idx], newD[idx+1:]...)
			continue
		}

		if len(updates) == ipSetUpdatesLimit {
			updatesBatches = append(updatesBatches, updates)
			updates = make([]*waf.IPSetUpdate, 0, ipSetUpdatesLimit)
		}

		updates = append(updates, &waf.IPSetUpdate{
			Action: aws.String(waf.ChangeActionDelete),
			IPSetDescriptor: &waf.IPSetDescriptor{
				Type:  aws.String(descriptor["type"].(string)),
				Value: aws.String(descriptor["value"].(string)),
			},
		})
	}

	for _, nd := range newD {
		descriptor := nd.(map[string]interface{})

		if len(updates) == ipSetUpdatesLimit {
			updatesBatches = append(updatesBatches, updates)
			updates = make([]*waf.IPSetUpdate, 0, ipSetUpdatesLimit)
		}

		updates = append(updates, &waf.IPSetUpdate{
			Action: aws.String(waf.ChangeActionInsert),
			IPSetDescriptor: &waf.IPSetDescriptor{
				Type:  aws.String(descriptor["type"].(string)),
				Value: aws.String(descriptor["value"].(string)),
			},
		})
	}
	updatesBatches = append(updatesBatches, updates)
	return updatesBatches
}
