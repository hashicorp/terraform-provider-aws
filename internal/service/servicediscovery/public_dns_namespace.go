package servicediscovery

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePublicDNSNamespace() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourcePublicDNSNamespaceCreate,
		ReadWithoutTimeout:   resourcePublicDNSNamespaceRead,
		UpdateWithoutTimeout: resourcePublicDNSNamespaceUpdate,
		DeleteWithoutTimeout: resourcePublicDNSNamespaceDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"hosted_zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validNamespaceName,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePublicDNSNamespaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &servicediscovery.CreatePublicDnsNamespaceInput{
		CreatorRequestId: aws.String(resource.UniqueId()),
		Name:             aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Service Discovery Public DNS Namespace: %s", input)
	output, err := conn.CreatePublicDnsNamespaceWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating Service Discovery Public DNS Namespace (%s): %s", name, err)
	}

	operation, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId))

	if err != nil {
		return diag.Errorf("waiting for Service Discovery Public DNS Namespace (%s) create: %s", name, err)
	}

	namespaceID, ok := operation.Targets[servicediscovery.OperationTargetTypeNamespace]

	if !ok {
		return diag.Errorf("creating Service Discovery Public DNS Namespace (%s): operation response missing Namespace ID", name)
	}

	d.SetId(aws.StringValue(namespaceID))

	return resourcePublicDNSNamespaceRead(ctx, d, meta)
}

func resourcePublicDNSNamespaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn()
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ns, err := FindNamespaceByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Public DNS Namespace %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading Service Discovery Public DNS Namespace (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(ns.Arn)
	d.Set("arn", arn)
	d.Set("description", ns.Description)
	if ns.Properties != nil && ns.Properties.DnsProperties != nil {
		d.Set("hosted_zone", ns.Properties.DnsProperties.HostedZoneId)
	} else {
		d.Set("hosted_zone", nil)
	}
	d.Set("name", ns.Name)

	tags, err := ListTags(ctx, conn, arn)

	if err != nil {
		return diag.Errorf("listing tags for Service Discovery Public DNS Namespace (%s): %s", arn, err)
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

func resourcePublicDNSNamespaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn()

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(ctx, conn, d.Get("arn").(string), o, n); err != nil {
			return diag.Errorf("updating Service Discovery Public DNS Namespace (%s) tags: %s", d.Id(), err)
		}
	}

	return resourcePublicDNSNamespaceRead(ctx, d, meta)
}

func resourcePublicDNSNamespaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn()

	log.Printf("[INFO] Deleting Service Discovery Public DNS Namespace: %s", d.Id())
	output, err := conn.DeleteNamespaceWithContext(ctx, &servicediscovery.DeleteNamespaceInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		return diag.Errorf("deleting Service Discovery Public DNS Namespace (%s): %s", d.Id(), err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(ctx, conn, aws.StringValue(output.OperationId)); err != nil {
			return diag.Errorf("waiting for Service Discovery Public DNS Namespace (%s) delete: %s", d.Id(), err)
		}
	}

	return nil
}
