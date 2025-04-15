// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_sqs_queue_redrive_allow_policy", name="Queue Redrive Allow Policy")
func resourceQueueRedriveAllowPolicy() *schema.Resource {
	h := &queueAttributeHandler{
		AttributeName: types.QueueAttributeNameRedriveAllowPolicy,
		SchemaKey:     "redrive_allow_policy",
		ToSet: func(old, new string) (string, error) {
			if verify.JSONBytesEqual([]byte(old), []byte(new)) {
				return old, nil
			}
			return new, nil
		},
	}

	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"queue_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"redrive_allow_policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CreateWithoutTimeout: h.Upsert,
		ReadWithoutTimeout:   h.Read,
		UpdateWithoutTimeout: h.Upsert,
		DeleteWithoutTimeout: h.Delete,
	}
}
