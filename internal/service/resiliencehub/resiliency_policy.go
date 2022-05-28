package resiliencehub

import (
	"context"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"regexp"

	"github.com/aws/aws-sdk-go/service/resiliencehub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func ResourceResiliencyPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: ResourceResiliencyPolicyCreate,
		ReadContext:   ResourceResiliencyPolicyRead,
		UpdateContext: ResourceResiliencyPolicyUpdate,
		DeleteContext: ResourceResiliencyPolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[A-Za-z\d][A-Za-z\d_\\-]{1,59}$`), "Up to 60 alphanumeric characters, or hyphens, without spaces. The first character must be a letter or a number."),
			},
			"policy_description": {
				Type:         schema.TypeString,
				Required:     false,
				ForceNew:     false,
				ValidateFunc: validation.StringLenBetween(0, 500),
			},
			"policy": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"AZ":       FailurePolicy(true),
						"Hardware": FailurePolicy(true),
						"Software": FailurePolicy(true),
						"Region":   FailurePolicy(false),
					},
				},
			},
			"data_location_constraint": {
				Type:         schema.TypeString,
				Required:     false,
				ForceNew:     false,
				ValidateFunc: validation.StringInSlice(resiliencehub.DataLocationConstraint_Values(), false),
			},
			"tier": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     false,
				ValidateFunc: validation.StringInSlice(resiliencehub.ResiliencyPolicyTier_Values(), false),
			},
			"tags": tftags.TagsSchema(),
		},
	}
}

func ResourceResiliencyPolicyCreate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResilienceHubConn

	input := &resiliencehub.CreateResiliencyPolicyInput{
		Policy: data.Get("policy"),
	}
}

func ResourceResiliencyPolicyRead(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResilienceHubConn
	return nil
}

func ResourceResiliencyPolicyUpdate(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResilienceHubConn
	return nil
}

func ResourceResiliencyPolicyDelete(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ResilienceHubConn
	return nil
}

func FailurePolicy(required bool) *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeMap,
		Required: required,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"rpoInSecs": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntAtLeast(0),
				},
				"rtoInSecs": {
					Type:         schema.TypeInt,
					Required:     true,
					ValidateFunc: validation.IntAtLeast(0),
				},
			},
		},
	}
}
