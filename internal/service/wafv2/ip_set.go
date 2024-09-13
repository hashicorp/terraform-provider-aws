// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package wafv2

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_wafv2_ip_set", name="IP Set")
// @Tags(identifierAttribute="arn")
func resourceIPSet() *schema.Resource {
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
				d.Set(names.AttrName, name)
				d.Set(names.AttrScope, scope)
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
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrDescription: {
					Type:         schema.TypeString,
					Optional:     true,
					ValidateFunc: validation.StringLenBetween(1, 256),
				},
				"ip_address_version": {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.IPAddressVersion](),
				},
				"lock_token": {
					Type:     schema.TypeString,
					Computed: true,
				},
				names.AttrName: {
					Type:         schema.TypeString,
					Required:     true,
					ForceNew:     true,
					ValidateFunc: validation.StringLenBetween(1, 128),
				},
				names.AttrScope: {
					Type:             schema.TypeString,
					Required:         true,
					ForceNew:         true,
					ValidateDiagFunc: enum.Validate[awstypes.Scope](),
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceIPSetCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	name := d.Get(names.AttrName).(string)
	input := &wafv2.CreateIPSetInput{
		Addresses:        []string{},
		IPAddressVersion: awstypes.IPAddressVersion(d.Get("ip_address_version").(string)),
		Name:             aws.String(name),
		Scope:            awstypes.Scope(d.Get(names.AttrScope).(string)),
		Tags:             getTagsIn(ctx),
	}

	if v, ok := d.GetOk("addresses"); ok && v.(*schema.Set).Len() > 0 {
		input.Addresses = flex.ExpandStringValueSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk(names.AttrDescription); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateIPSet(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating WAFv2 IPSet (%s): %s", name, err)
	}

	d.SetId(aws.ToString(output.Summary.Id))

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	output, err := findIPSetByThreePartKey(ctx, conn, d.Id(), d.Get(names.AttrName).(string), d.Get(names.AttrScope).(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] WAFv2 IPSet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading WAFv2 IPSet (%s): %s", d.Id(), err)
	}

	ipSet := output.IPSet
	d.Set("addresses", ipSet.Addresses)
	arn := aws.ToString(ipSet.ARN)
	d.Set(names.AttrARN, arn)
	d.Set(names.AttrDescription, ipSet.Description)
	d.Set("ip_address_version", ipSet.IPAddressVersion)
	d.Set("lock_token", output.LockToken)
	d.Set(names.AttrName, ipSet.Name)

	return diags
}

func resourceIPSetUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	if d.HasChangesExcept(names.AttrTags, names.AttrTagsAll) {
		input := &wafv2.UpdateIPSetInput{
			Addresses: []string{},
			Id:        aws.String(d.Id()),
			LockToken: aws.String(d.Get("lock_token").(string)),
			Name:      aws.String(d.Get(names.AttrName).(string)),
			Scope:     awstypes.Scope(d.Get(names.AttrScope).(string)),
		}

		if v, ok := d.GetOk("addresses"); ok && v.(*schema.Set).Len() > 0 {
			input.Addresses = flex.ExpandStringValueSet(v.(*schema.Set))
		}

		if v, ok := d.GetOk(names.AttrDescription); ok {
			input.Description = aws.String(v.(string))
		}

		log.Printf("[INFO] Updating WAFv2 IPSet: %s", d.Id())
		_, err := conn.UpdateIPSet(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating WAFv2 IPSet (%s): %s", d.Id(), err)
		}
	}

	return append(diags, resourceIPSetRead(ctx, d, meta)...)
}

func resourceIPSetDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).WAFV2Client(ctx)

	input := &wafv2.DeleteIPSetInput{
		Id:        aws.String(d.Id()),
		LockToken: aws.String(d.Get("lock_token").(string)),
		Name:      aws.String(d.Get(names.AttrName).(string)),
		Scope:     awstypes.Scope(d.Get(names.AttrScope).(string)),
	}

	log.Printf("[INFO] Deleting WAFv2 IPSet: %s", d.Id())
	const (
		timeout = 5 * time.Minute
	)
	_, err := tfresource.RetryWhenIsA[*awstypes.WAFAssociatedItemException](ctx, timeout, func() (interface{}, error) {
		return conn.DeleteIPSet(ctx, input)
	})

	if errs.IsA[*awstypes.WAFNonexistentItemException](err) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting WAFv2 IPSet (%s): %s", d.Id(), err)
	}

	return diags
}

func findIPSetByThreePartKey(ctx context.Context, conn *wafv2.Client, id, name, scope string) (*wafv2.GetIPSetOutput, error) {
	input := &wafv2.GetIPSetInput{
		Id:    aws.String(id),
		Name:  aws.String(name),
		Scope: awstypes.Scope(scope),
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

	return output, nil
}
