package ec2

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// ACL Network ACLs all contain explicit deny-all rules that cannot be
// destroyed or changed by users. This rules are numbered very high to be a
// catch-all.
// See http://docs.aws.amazon.com/AmazonVPC/latest/UserGuide/VPC_ACLs.html#default-network-acl
const (
	defaultACLRuleNumberIPv4 = 32767
	defaultACLRuleNumberIPv6 = 32768
)

func ResourceDefaultNetworkACL() *schema.Resource {
	networkACLRuleSetNestedBlock := &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem:     networkACLRuleNestedBlock,
		Set:      networkACLRuleHash,
	}

	return &schema.Resource{
		CreateWithoutTimeout: resourceDefaultNetworkACLCreate,
		ReadWithoutTimeout:   resourceNetworkACLRead,
		UpdateWithoutTimeout: resourceDefaultNetworkACLUpdate,
		DeleteContext:        resourceDefaultNetworkACLDelete, // nosemgrep:ci.avoid-context-CRUD-handlers

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
		Schema: map[string]*schema.Schema{
			"arn": {
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
			"egress":  networkACLRuleSetNestedBlock,
			"ingress": networkACLRuleSetNestedBlock,
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			// We want explicit management of Subnets here, so we do not allow them to be
			// computed. Instead, an empty config will enforce just that; removal of the
			// any Subnets that have been assigned to the Default Network ACL. Because we
			// can't actually remove them, this will be a continual plan until the
			// Subnets are themselves destroyed or reassigned to a different Network
			// ACL
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"vpc_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceDefaultNetworkACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

	naclID := d.Get("default_network_acl_id").(string)
	nacl, err := FindNetworkACLByID(ctx, conn, naclID)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading EC2 Network ACL (%s): %s", naclID, err)
	}

	if !aws.BoolValue(nacl.IsDefault) {
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
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig
	newTags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{}))).IgnoreConfig(ignoreTagsConfig)
	oldTags := KeyValueTags(nacl.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	if !oldTags.Equal(newTags) {
		if err := UpdateTags(ctx, conn, d.Id(), oldTags, newTags); err != nil {
			return sdkdiag.AppendErrorf(diags, "updating EC2 Default Network ACL (%s) tags: %s", d.Id(), err)
		}
	}

	return append(diags, resourceNetworkACLRead(ctx, d, meta)...)
}

func resourceDefaultNetworkACLUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).EC2Conn()

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

func resourceDefaultNetworkACLDelete(_ context.Context, d *schema.ResourceData, meta interface{}) (diags diag.Diagnostics) {
	return sdkdiag.AppendWarningf(diags, "EC2 Default Network ACL (%s) not deleted, removing from state", d.Id())
}
