package servicediscovery

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourcePrivateDNSNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourcePrivateDNSNamespaceCreate,
		Read:   resourcePrivateDNSNamespaceRead,
		Update: resourcePrivateDNSNamespaceUpdate,
		Delete: resourcePrivateDNSNamespaceDelete,

		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected NAMESPACE_ID:VPC_ID", d.Id())
				}
				d.SetId(idParts[0])
				d.Set("vpc", idParts[1])
				return []*schema.ResourceData{d}, nil
			},
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
			"vpc": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourcePrivateDNSNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &servicediscovery.CreatePrivateDnsNamespaceInput{
		CreatorRequestId: aws.String(resource.UniqueId()),
		Name:             aws.String(name),
		Vpc:              aws.String(d.Get("vpc").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Service Discovery Private DNS Namespace: %s", input)
	output, err := conn.CreatePrivateDnsNamespace(input)

	if err != nil {
		return fmt.Errorf("creating Service Discovery Private DNS Namespace (%s): %w", name, err)
	}

	operation, err := WaitOperationSuccess(conn, aws.StringValue(output.OperationId))

	if err != nil {
		return fmt.Errorf("waiting for Service Discovery Private DNS Namespace (%s) create: %w", name, err)
	}

	namespaceID, ok := operation.Targets[servicediscovery.OperationTargetTypeNamespace]

	if !ok {
		return fmt.Errorf("creating Service Discovery Private DNS Namespace (%s): operation response missing Namespace ID", name)
	}

	d.SetId(aws.StringValue(namespaceID))

	return resourcePrivateDNSNamespaceRead(d, meta)
}

func resourcePrivateDNSNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ns, err := FindNamespaceByID(conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Service Discovery Private DNS Namespace %s not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("reading Service Discovery Private DNS Namespace (%s): %w", d.Id(), err)
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
		return fmt.Errorf("listing tags for Service Discovery Private DNS Namespace (%s): %w", arn, err)
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

func resourcePrivateDNSNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("updating Service Discovery Private DNS Namespace (%s) tags: %w", d.Id(), err)
		}
	}

	return resourcePrivateDNSNamespaceRead(d, meta)
}

func resourcePrivateDNSNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn

	log.Printf("[INFO] Deleting Service Discovery Private DNS Namespace: %s", d.Id())
	output, err := conn.DeleteNamespace(&servicediscovery.DeleteNamespaceInput{
		Id: aws.String(d.Id()),
	})

	if err != nil {
		return fmt.Errorf("deleting Service Discovery Private DNS Namespace (%s): %w", d.Id(), err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := WaitOperationSuccess(conn, aws.StringValue(output.OperationId)); err != nil {
			return fmt.Errorf("waiting for Service Discovery Private DNS Namespace (%s) delete: %w", d.Id(), err)
		}
	}

	return nil
}
