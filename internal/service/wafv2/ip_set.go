// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/wafv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	ipSetDeleteTimeout = 5 * time.Minute
)

// @SDKResource("aws_wafv2_ip_set", name="IP Set")
// @Tags(identifierAttribute="arn")
func ResourceIPSet() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceIPSetCreate,
		ReadWithoutTimeout:   resourceIPSetRead,
		UpdateWithoutTimeout: resourceIPSetUpdate,
		DeleteWithoutTimeout: resourceIPSetDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected ID/NAME/SCOPE", d.Id())
				}
				id := idParts[0]
				name := idParts[1]
				scope := idParts[2]
				d.SetId(id)
				d.Set("name", name)
				d.Set("scope", scope)
				return []*schema.ResourceData{d}, nil
			},
		},

		SchemaFunc: func() map[string]*schema.Schema {
			return map[string]*schema.Schema{
				"addresses": {
					Type:     schema.TypeSet,
					Optional: true,
					MaxItems: 10000,
					Elem:     &schema.Schema{Type: schema.TypeString},
					DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
						if d.GetRawPlan().GetAttr("addresses").IsWhollyKnown() {
							o, n := d.GetChange("addresses")
							oldAddresses := o.(*schema.Set).List()
							newAddresses := n.(*schema.Set).List()
							if len(oldAddresses) == len(newAddresses) {
								for _, ov := range oldAddresses {
									hasAddress := false
									for _, nv := range newAddresses {
										if itypes.CIDRBlocksEqual(ov.(string), nv.(string)) {
											hasAddress = true
											break
										}
									}
									if !hasAddress {
										return false
									}
								}
								return true
							}
						}
						return false
					},
				},
				"arn": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"description": {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 256),
				},
				"ip_address_version": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice(wafv2.IPAddressVersion_Values(), false),
				},
				"lock_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
				"name": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				"scope": {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringInSlice(wafv2.Scope_Values(), false),
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceIPSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	name := d.Get("name").(string)
	input := &wafv2.CreateIPSetInput{
		Addresses:        aws.StringSlice([]string{}),
		IPAddressVersion: aws.String(d.Get("ip_address_version").(string)),
		Name:             aws.String(name),
		Scope:            aws.String(d.Get("scope").(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("addresses"); ok && v.(*schema.Set).Len() > 0 {
		input.Addresses = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateIPSetWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating WAFv2 IPSet (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.Summary.Id))

	return resourceIPSetRead(ctx, d, meta)
}

func resourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	output, err := FindIPSetByThreePartKey(ctx, conn, d.Id(), d.Get("name").(string), d.Get("scope").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 IPSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading WAFv2 IPSet (%s): %s", d.Id(), err)
	}

	ipSet := output.IPSet
	d.Set("addresses", aws.StringValueSlice(ipSet.Addresses))
	arn := aws.StringValue(ipSet.ARN)
	d.Set("arn", arn)
	d.Set("description", ipSet.Description)
	d.Set("ip_address_version", ipSet.IPAddressVersion)
	d.Set("lock_token", output.LockToken)
	d.Set("name", ipSet.Name)

	return nil
}

func resourceIPSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	if d.HasChangesExcept("tags", "tags_all") {
		input := &wafv2.UpdateIPSetInput{
			Addresses: aws.StringSlice([]string{}),
			Id:        aws.String(d.Id()),
			LockToken: aws.String(d.Get("lock_token").(string)),
			Name:      aws.String(d.Get("name").(string)),
			Scope:     aws.String(d.Get("scope").(string)),
		}

		if v, ok := d.GetOk("addresses"); ok && v.(*schema.Set).Len() > 0 {
			input.Addresses = flex.ExpandStringSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}

		log.Printf("[INFO] Updating WAFv2 IPSet: %s", input)
		_, err := conn.UpdateIPSetWithContext(ctx, input)

		if err != nil {
			return diag.Errorf("updating WAFv2 IPSet (%s): %s", d.Id(), err)
		}
	}

	return resourceIPSetRead(ctx, d, meta)
}

func resourceIPSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).WAFV2Conn(ctx)

	input := &wafv2.DeleteIPSetInput{
		Id:        aws.String(d.Id()),
		LockToken: aws.String(d.Get("lock_token").(string)),
		Name:      aws.String(d.Get("name").(string)),
		Scope:     aws.String(d.Get("scope").(string)),
	}

	log.Printf("[INFO] Deleting WAFv2 IPSet: %s", d.Id())
	_, err := tfresource.RetryWhenAWSErrCodeEquals(ctx, ipSetDeleteTimeout, func() (interface{}, error) {
		return conn.DeleteIPSetWithContext(ctx, input)
	}, wafv2.ErrCodeWAFAssociatedItemException)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting WAFv2 IPSet (%s): %s", d.Id(), err)
	}

	return nil
}

func FindIPSetByThreePartKey(ctx context.Context, conn *wafv2.WAFV2, id, name, scope string) (*wafv2.GetIPSetOutput, error) {
	input := &wafv2.GetIPSetInput{
		Id:    aws.String(id),
		Name:  aws.String(name),
		Scope: aws.String(scope),
	}

	output, err := conn.GetIPSetWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, wafv2.ErrCodeWAFNonexistentItemException) {
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

	return output, nil
}
