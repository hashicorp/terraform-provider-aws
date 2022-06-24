package servicediscovery

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePublicDNSNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourcePublicDNSNamespaceCreate,
		Read:   resourcePublicDNSNamespaceRead,
		Update: resourcePublicDNSNamespaceUpdate,
		Delete: resourcePublicDNSNamespaceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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

func resourcePublicDNSNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
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
	output, err := conn.CreatePublicDnsNamespace(input)

	if err != nil {
		return fmt.Errorf("creating Service Discovery Public DNS Namespace (%s): %w", name, err)
	}

	operation, err := WaitOperationSuccess(conn, aws.StringValue(output.OperationId))

	if err != nil {
		return fmt.Errorf("waiting for Service Discovery Public DNS Namespace (%s) create: %w", name, err)
	}

	namespaceID, ok := operation.Targets[servicediscovery.OperationTargetTypeNamespace]

	if !ok {
		return fmt.Errorf("creating Service Discovery Public DNS Namespace (%s): operation response missing Namespace ID", name)
	}

	d.SetId(aws.StringValue(namespaceID))

	return resourcePublicDNSNamespaceRead(d, meta)
}

func resourcePublicDNSNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ns, err := FindNamespaceByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Public DNS Namespace %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Service Discovery Public DNS Namespace (%s): %w", d.Id(), err)
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

	tags, err := ListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("listing tags for Service Discovery Public DNS Namespace (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("setting tags_all: %w", err)
	}

	return nil
}

func resourcePublicDNSNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("updating Service Discovery Public DNS Namespace (%s) tags: %s", d.Id(), err)
		}
	}

	return resourcePublicDNSNamespaceRead(d, meta)
}

func resourcePublicDNSNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn

	log.Printf("[INFO] Deleting Service Discovery Public DNS Namespace: %s", d.Id())
	output, err := conn.DeleteNamespace(&servicediscovery.DeleteNamespaceInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("deleting Service Discovery Public DNS Namespace (%s): %w", d.Id(), err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(conn, aws.StringValue(output.OperationId)); err != nil {
			return fmt.Errorf("waiting for Service Discovery Public DNS Namespace (%s) delete: %w", d.Id(), err)
		}
	}

	return nil
}
