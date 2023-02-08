package route53resolver

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFirewallDomainList() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceFirewallDomainListCreate,
		ReadWithoutTimeout:   resourceFirewallDomainListRead,
		UpdateWithoutTimeout: resourceFirewallDomainListUpdate,
		DeleteWithoutTimeout: resourceFirewallDomainListDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domains": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validResolverName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFirewallDomainListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &route53resolver.CreateFirewallDomainListInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-firewall-domain-list-")),
		Name:             aws.String(name),
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	output, err := conn.CreateFirewallDomainListWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Route53 Resolver Firewall Domain List (%s): %s", name, err)
	}

	d.SetId(aws.StringValue(output.FirewallDomainList.Id))

	if v, ok := d.GetOk("domains"); ok && v.(*schema.Set).Len() > 0 {
		_, err := conn.UpdateFirewallDomainsWithContext(ctx, &route53resolver.UpdateFirewallDomainsInput{
			FirewallDomainListId: aws.String(d.Id()),
			Domains:              flex.ExpandStringSet(v.(*schema.Set)),
			Operation:            aws.String(route53resolver.FirewallDomainUpdateOperationAdd),
		})

		if err != nil {
			return diag.Errorf("updating Route53 Resolver Firewall Domain List (%s) domains: %s", d.Id(), err)
		}

		if _, err = waitFirewallDomainListUpdated(ctx, conn, d.Id()); err != nil {
			return diag.Errorf("waiting for Route53 Resolver Firewall Domain List (%s) update: %s", d.Id(), err)
		}
	}

	return resourceFirewallDomainListRead(ctx, d, meta)
}

func resourceFirewallDomainListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	firewallDomainList, err := FindFirewallDomainListByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Route53 Resolver Firewall Domain List (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Route53 Resolver Firewall Domain List (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(firewallDomainList.Arn)
	d.Set("arn", arn)
	d.Set("name", firewallDomainList.Name)

	input := &route53resolver.ListFirewallDomainsInput{
		FirewallDomainListId: aws.String(d.Id()),
	}
	var output []*string

	err = conn.ListFirewallDomainsPagesWithContext(ctx, input, func(page *route53resolver.ListFirewallDomainsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.Domains...)

		return !lastPage
	})

	if err != nil {
		return diag.Errorf("listing Route53 Resolver Firewall Domain List (%s) domains: %s", d.Id(), err)
	}

	d.Set("domains", aws.StringValueSlice(output))

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Route53 Resolver Firewall Domain List (%s): %s", arn, err)
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

func resourceFirewallDomainListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	if d.HasChange("domains") {
		o, n := d.GetChange("domains")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}
		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		domains := ns
		operation := route53resolver.FirewallDomainUpdateOperationReplace

		if domains.Len() == 0 {
			domains = os
			operation = route53resolver.FirewallDomainUpdateOperationRemove
		}

		_, err := conn.UpdateFirewallDomainsWithContext(ctx, &route53resolver.UpdateFirewallDomainsInput{
			FirewallDomainListId: aws.String(d.Id()),
			Domains:              flex.ExpandStringSet(domains),
			Operation:            aws.String(operation),
		})

		if err != nil {
			return diag.Errorf("updating Route53 Resolver Firewall Domain List (%s) domains: %s", d.Id(), err)
		}

		if _, err = waitFirewallDomainListUpdated(ctx, conn, d.Id()); err != nil {
			return diag.Errorf("waiting for Route53 Resolver Firewall Domain List (%s) update: %s", d.Id(), err)
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Route53 Resolver Firewall Domain List (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceFirewallDomainListRead(ctx, d, meta)
}

func resourceFirewallDomainListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).Route53ResolverConn()

	log.Printf("[DEBUG] Deleting Route53 Resolver Firewall Domain List: %s", d.Id())
	_, err := conn.DeleteFirewallDomainListWithContext(ctx, &route53resolver.DeleteFirewallDomainListInput{
		FirewallDomainListId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting Route53 Resolver Firewall Domain List (%s): %s", d.Id(), err)
	}

	if _, err = waitFirewallDomainListDeleted(ctx, conn, d.Id()); err != nil {
		return diag.Errorf("waiting for Route53 Resolver Firewall Domain List (%s) delete: %s", d.Id(), err)
	}

	return nil
}

func FindFirewallDomainListByID(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallDomainList, error) {
	input := &route53resolver.GetFirewallDomainListInput{
		FirewallDomainListId: aws.String(id),
	}

	output, err := conn.GetFirewallDomainListWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, route53resolver.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FirewallDomainList == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.FirewallDomainList, nil
}

func statusFirewallDomainList(ctx context.Context, conn *route53resolver.Route53Resolver, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindFirewallDomainListByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

const (
	firewallDomainListUpdatedTimeout = 5 * time.Minute
	firewallDomainListDeletedTimeout = 5 * time.Minute
)

func waitFirewallDomainListUpdated(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallDomainList, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallDomainListStatusUpdating, route53resolver.FirewallDomainListStatusImporting},
		Target: []string{route53resolver.FirewallDomainListStatusComplete,
			route53resolver.FirewallDomainListStatusCompleteImportFailed,
		},
		Refresh: statusFirewallDomainList(ctx, conn, id),
		Timeout: firewallDomainListUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.FirewallDomainList); ok {
		if status := aws.StringValue(output.Status); status == route53resolver.FirewallDomainListStatusCompleteImportFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusMessage)))
		}

		return output, err
	}

	return nil, err
}

func waitFirewallDomainListDeleted(ctx context.Context, conn *route53resolver.Route53Resolver, id string) (*route53resolver.FirewallDomainList, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{route53resolver.FirewallDomainListStatusDeleting},
		Target:  []string{},
		Refresh: statusFirewallDomainList(ctx, conn, id),
		Timeout: firewallDomainListDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*route53resolver.FirewallDomainList); ok {
		return output, err
	}

	return nil, err
}
