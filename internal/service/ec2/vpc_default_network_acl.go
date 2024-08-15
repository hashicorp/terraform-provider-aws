// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// ACL Network ACLs all contain explicit deny-all rules that cannot be
// destroyed or changed by users. This rules are numbered very high to be a
// catch-all.
// See http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_ACLs.html#default-network-acl
const (
	defaultACLRuleNumberIPv4 = 32767
	defaultACLRuleNumberIPv6 = 32768
)

// @SDKResource("aws_default_network_acl", name="Network ACL")
// @Tags(identifierAttribute="id")
// @Testing(tagsTest=false)
func resourceDefaultNetworkACL() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultNetworkACLCreate,
		ReadWithoutTimeout:   resourceNetworkACLRead,
		UpdateWithoutTimeout: resourceDefaultNetworkACLUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("default_network_acl_id", d.Id())

				return []*schema.ResourceData{d}, nil
			},
		},

		// Keep in sync with aws_network_acl's schema with the following changes:
		//    - egress and ingress are not Computed and don't have "Attributes as Blocks" processing mode set
		//    - subnet_ids is not Computed
		// and additions:
		//   - default_network_acl_id Required/ForceNew
		SchemaFunc: func() map[string]*schema.Schema {
			networkACLRuleSetNestedBlock := func() *schema.Schema {
				return &schema.Schema{
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     networkACLRuleNestedBlock(),
					Set:      networkACLRuleHash,
				}
			}

			return map[string]*schema.Schema{
				names.AttrARN: {
					Type:     schema.TypeString,
					Computed: true,
				},
				"default_network_acl_id": {
					Type:     schema.TypeString,
					Required: true,
					ForceNew: true,
				},
				// We want explicit management of Rules here, so we do not allow them to be
				// computed. Instead, an empty config will enforce just that; removal of the
				// rules
				"egress":  networkACLRuleSetNestedBlock(),
				"ingress": networkACLRuleSetNestedBlock(),
				names.AttrOwnerID: {
					Type:     schema.TypeString,
					Computed: true,
				},
				// We want explicit management of Subnets here, so we do not allow them to be
				// computed. Instead, an empty config will enforce just that; removal of the
				// any Subnets that have been assigned to the Default Network ACL. Because we
				// can't actually remove them, this will be a continual plan until the
				// Subnets are themselves destroyed or reassigned to a different Network
				// ACL
				names.AttrSubnetIDs: {
					Type:     schema.TypeSet,
					Optional: true,
					Elem:     &schema.Schema{Type: schema.TypeString},
				},
				names.AttrTags:    tftags.TagsSchema(),
				names.AttrTagsAll: tftags.TagsSchemaComputed(),
				names.AttrVPCID: {
					Type:     schema.TypeString,
					Computed: true,
				},
			}
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDefaultNetworkACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics { // nosemgrep:ci.semgrep.tags.calling-UpdateTags-in-resource-create
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	naclID := d.Get("default_network_acl_id").(string)
	nacl, err := findNetworkACLByID(ctx, conn, naclID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network ACL (%s): %s", naclID, err)
	}

	if !aws.ToBool(nacl.IsDefault) {
		return sdkdiag.AppendErrorf(diags, "use the `aws_network_acl` resource instead")
	}

	d.SetId(naclID)

	// Revoke all default and pre-existing rules on the default network ACL.
	if err := deleteNetworkACLEntries(ctx, conn, d.Id(), nacl.Entries); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	if err := modifyNetworkACLAttributesOnCreate(ctx, conn, d); err != nil {
		return sdkdiag.AppendFromErr(diags, err)
	}

	// Configure tags.
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	newTags := keyValueTags(ctx, getTagsIn(ctx))
	oldTags := keyValueTags(ctx, nacl.Tags).IgnoreSystem(names.EC2).IgnoreConfig(ignoreTagsConfig)

	if !oldTags.Equal(newTags) {
		if err := updateTags(ctx, conn, d.Id(), oldTags, newTags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Default Network ACL (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceNetworkACLRead(ctx, d, meta)...)
}

func resourceDefaultNetworkACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Client(ctx)

	// Subnets *must* belong to a Network ACL. Subnets are not "removed" from
	// Network ACLs, instead their association is replaced. In a normal
	// Network ACL, any removal of a Subnet is done by replacing the
	// Subnet/ACL association with an association between the Subnet and the
	// Default Network ACL. Because we're managing the default here, we cannot
	// do that, so we simply log a NO-OP. In order to remove the Subnet here,
	// it must be destroyed, or assigned to different Network ACL. Those
	// operations are not handled here.
	if err := modifyNetworkACLAttributesOnUpdate(ctx, conn, d, false); err != nil {
		return sdkdiag.AppendErrorf(diags, "updating EC2 Default Network ACL (%s): %s", d.Id(), err)
	}

	return append(diags, resourceNetworkACLRead(ctx, d, meta)...)
}
