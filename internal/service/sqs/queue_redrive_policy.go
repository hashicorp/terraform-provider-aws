package sqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceQueueRedrivePolicy() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"redrive_policy": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
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
		SchemaVersion: 0,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CreateContext: generateQueueAttributeUpsertFunc(sqs.QueueAttributeNameRedrivePolicy),
		ReadContext:   generateQueueAttributeReadFunc(sqs.QueueAttributeNameRedrivePolicy),
		UpdateContext: generateQueueAttributeUpsertFunc(sqs.QueueAttributeNameRedrivePolicy),
		DeleteContext: generateQueueAttributeDeleteFunc(sqs.QueueAttributeNameRedrivePolicy),
	}
}
