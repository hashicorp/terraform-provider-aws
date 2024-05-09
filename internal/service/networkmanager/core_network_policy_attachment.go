// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

import (
	"context"
	"log"
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go/private/protocol"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_networkmanager_core_network_policy_attachment")
func ResourceCoreNetworkPolicyAttachment() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCoreNetworkPolicyAttachmentCreate,
		ReadWithoutTimeout:   resourceCoreNetworkPolicyAttachmentRead,
		UpdateWithoutTimeout: resourceCoreNetworkPolicyAttachmentUpdate,
		DeleteWithoutTimeout: schema.NoopContext,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Update: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"core_network_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 50),
					validation.StringMatch(regexache.MustCompile(`^core-network-([0-9a-f]{8,17})$`), "must be a valid Core Network ID"),
				),
			},
			"policy_document": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(0, 10000000),
					validation.StringIsJSON,
				),
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			names.AttrState: {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCoreNetworkPolicyAttachmentCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	d.SetId(d.Get("core_network_id").(string))

	return resourceCoreNetworkPolicyAttachmentUpdate(ctx, d, meta)
}

func resourceCoreNetworkPolicyAttachmentRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	coreNetwork, err := FindCoreNetworkByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Network Manager Core Network %s not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Core Network (%s): %s", d.Id(), err)
	}

	d.Set("core_network_id", coreNetwork.CoreNetworkId)
	d.Set(names.AttrState, coreNetwork.State)

	// getting the policy document uses a different API call
	// pass in latestPolicyVersionId to get the latest version id by default
	coreNetworkPolicy, err := FindCoreNetworkPolicyByTwoPartKey(ctx, conn, d.Id(), latestPolicyVersionID)

	if tfresource.NotFound(err) {
		d.Set("policy_document", nil)
	} else if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Network Manager Core Network (%s) policy: %s", d.Id(), err)
	} else {
		encodedPolicyDocument, err := protocol.EncodeJSONValue(coreNetworkPolicy.PolicyDocument, protocol.NoEscape)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "encoding Network Manager Core Network (%s) policy document: %s", d.Id(), err)
		}

		d.Set("policy_document", encodedPolicyDocument)
	}
	return diags
}

func resourceCoreNetworkPolicyAttachmentUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*conns.AWSClient).NetworkManagerConn(ctx)

	if d.HasChange("policy_document") {
		err := PutAndExecuteCoreNetworkPolicy(ctx, conn, d.Id(), d.Get("policy_document").(string))

		if err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}

		if _, err := waitCoreNetworkUpdated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutUpdate)); err != nil {
			return sdkdiag.AppendErrorf(diags, "waiting for Network Manager Core Network (%s) update: %s", d.Id(), err)
		}
	}

	return append(diags, resourceCoreNetworkPolicyAttachmentRead(ctx, d, meta)...)
}
