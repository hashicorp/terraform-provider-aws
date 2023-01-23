package route53resolver

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFirewallRuleGroup() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallRuleGroupCreate,
		ReadWithoutTimeout:   resourceFirewallRuleGroupRead,
		UpdateWithoutTimeout: resourceFirewallRuleGroupUpdate,
		DeleteWithoutTimeout: resourceFirewallRuleGroupDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validResolverName,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFirewallRuleGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &route53resolver.CreateFirewallRuleGroupInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-firewall-rule-group-")),
		Name:             aws.String(name),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateFirewallRuleGroupWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Route53 Resolver Firewall Rule Group (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.FirewallRuleGroup.Id))

	return resourceFirewallRuleGroupRead(ctx, d, meta)
}

func resourceFirewallRuleGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ruleGroup, err := FindFirewallRuleGroupByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Rule Group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Firewall Rule Group (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(ruleGroup.Arn)
	d.Set("arn", arn)
	d.Set("name", ruleGroup.Name)
	d.Set("owner_id", ruleGroup.OwnerId)
	d.Set("share_status", ruleGroup.ShareStatus)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Route53 Resolver Firewall Rule Group (%s): %s", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.Errorf("setting tags: %s", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.Errorf("setting tags_all: %s", err)
	}

	return nil
}

func resourceFirewallRuleGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Route53 Resolver Firewall Rule Group (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceFirewallRuleGroupRead(ctx, d, meta)
}

func resourceFirewallRuleGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Rule Group: %s", d.Id())
	_, err := conn.DeleteFirewallRuleGroupWithContext(ctx, &route53resolver.DeleteFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Route53 Resolver Firewall Rule Group (%s): %s", d.Id(), err)
	}

	return nil
}

func FindFirewallRuleGroupByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallRuleGroup, error) {
	input := &route53resolver.GetFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(id),
	}

	output, err := conn.GetFirewallRuleGroupWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FirewallRuleGroup == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.FirewallRuleGroup, nil
}
