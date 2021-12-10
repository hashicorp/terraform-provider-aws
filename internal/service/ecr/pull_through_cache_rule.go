package ecr

import (
	"context"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourcePullThroughCacheRule() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePullThroughCacheRuleCreate,
		ReadContext:   resourcePullThroughCacheRuleRead,
		DeleteContext: resourcePullThroughCacheRuleDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"ecr_repository_prefix": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(2, 20),
					validation.StringMatch(
						regexp.MustCompile(`^[a-z0-9]+(?:[._-][a-z0-9]+)*$`),
						"must only include alphanumeric, underscore, period, or hyphen characters"),
				),
			},
			"registry_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"upstream_registry_url": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourcePullThroughCacheRuleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ECRConn

	input := &ecr.CreatePullThroughCacheRuleInput{
		EcrRepositoryPrefix: aws.String(d.Get("ecr_repository_prefix").(string)),
		UpstreamRegistryUrl: aws.String(d.Get("upstream_registry_url").(string)),
	}

	_, err := conn.CreatePullThroughCacheRuleWithContext(ctx, input)
	if err != nil {
		return diag.Errorf("failed to create pull through cache rule: %s", err)
	}

	d.SetId(d.Get("ecr_repository_prefix").(string))

	return resourcePullThroughCacheRuleRead(ctx, d, meta)
}

func resourcePullThroughCacheRuleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ECRConn

	rule, err := FindPullThroughCacheRuleByRepositoryPrefix(ctx, conn, d.Id())
	if err != nil {
		return diag.Errorf("failed to find ECR Pull Through Cache Rule: %s", err)
	}

	if rule == nil {
		log.Printf("[WARN] ECR Pull Through Cache Rule %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("ecr_repository_prefix", rule.EcrRepositoryPrefix)
	d.Set("registry_id", rule.RegistryId)
	d.Set("upstream_registry_url", rule.UpstreamRegistryUrl)

	return nil
}

func resourcePullThroughCacheRuleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ECRConn

	input := &ecr.DeletePullThroughCacheRuleInput{
		EcrRepositoryPrefix: aws.String(d.Get("ecr_repository_prefix").(string)),
		RegistryId:          aws.String(d.Get("registry_id").(string)),
	}

	_, err := conn.DeletePullThroughCacheRuleWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, ecr.ErrCodePullThroughCacheRuleNotFoundException) {
			return nil
		}

		return diag.Errorf("failed to delete pull through cache rule: %s", err)
	}

	return nil
}
