// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_sqs_queue_policy")
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

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		MigrateState:  QueuePolicyMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			names.AttrPolicy: {
				Type:                  schema.TypeString,
				Required:              true,
				ValidateFunc:          validation.StringIsJSON,
				DiffSuppressFunc:      verify.SuppressEquivalentPolicyDiffs,
				DiffSuppressOnRefresh: true,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
			},
			"queue_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}
