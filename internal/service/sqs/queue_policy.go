// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package sqs

import (
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/sdkv2"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sqs_queue_policy", name="Queue Policy")
// @IdentityVersion(1)
// @CustomInherentRegionIdentity("queue_url", "parseQueueURL")
// @Testing(preIdentityVersion="v6.9.0")
// @Testing(existsType="github.com/aws/aws-sdk-go-v2/service/sqs/types;awstypes;map[awstypes.QueueAttributeName]string")
// @Testing(identityVersion="0;v6.10.0")
// @Testing(identityVersion="1;v6.19.0")
func resourceQueuePolicy() *schema.Resource {
	h := &queueAttributeHandler{
		AttributeName: types.QueueAttributeNamePolicy,
		SchemaKey:     names.AttrPolicy,
		ToSet:         verify.PolicyToSet,
	}

	//lintignore:R011
	return &schema.Resource{
		CreateWithoutTimeout: h.Upsert,
		ReadWithoutTimeout:   h.Read,
		UpdateWithoutTimeout: h.Upsert,
		DeleteWithoutTimeout: h.Delete,

		MigrateState:  queuePolicyMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: sdkv2.IAMPolicyDocumentSchemaRequired(),
			"queue_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}
