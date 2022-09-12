package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceQueuePolicy() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		CreateContext: generateQueueAttributeUpsertFunc(sqs.QueueAttributeNamePolicy),
		ReadContext:   generateQueueAttributeReadFunc(sqs.QueueAttributeNamePolicy),
		UpdateContext: generateQueueAttributeUpsertFunc(sqs.QueueAttributeNamePolicy),
		DeleteContext: generateQueueAttributeDeleteFunc(sqs.QueueAttributeNamePolicy),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		MigrateState:  QueuePolicyMigrateState,
		SchemaVersion: 1,
		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
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
