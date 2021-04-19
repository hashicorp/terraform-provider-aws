package aws

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/chime"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsChimeVoiceConnector() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsChimeVoiceConnectorCreate,
		ReadContext: resourceAwsChimeVoiceConnectorRead,
		UpdateContext: resourceAwsChimeVoiceConnectorRead,
		DeleteContext: resourceAwsChimeVoiceConnectorRead,
	}
}

func resourceAwsChimeVoiceConnectorCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostic {
	conn := meta.(*AWSClient).chimeconn
	return resourceAwsChimeVoiceConnectorRead(ctx, d, meta)
}

func resourceAwsChimeVoiceConnectorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostic {

}