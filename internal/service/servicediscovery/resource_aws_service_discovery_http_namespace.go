package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	tftags "github.com/hashicorp/terraform-provider-aws/aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceHTTPNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceHTTPNamespaceCreate,
		Read:   resourceHTTPNamespaceRead,
		Update: resourceHTTPNamespaceUpdate,
		Delete: resourceHTTPNamespaceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validNamespaceName,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceHTTPNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	name := d.Get("name").(string)
	input := &servicediscovery.CreateHttpNamespaceInput{
		CreatorRequestId: aws.String(resource.UniqueId()),
		Name:             aws.String(name),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = tags.IgnoreAws().ServicediscoveryTags()
	}

	output, err := conn.CreateHttpNamespace(input)

	if err != nil {
		return fmt.Errorf("error creating Service Discovery HTTP Namespace (%s): %w", name, err)
	}

	if output == nil || output.OperationId == nil {
		return fmt.Errorf("error creating Service Discovery HTTP Namespace (%s): creation response missing Operation ID", name)
	}

	operation, err := waiter.OperationSuccess(conn, aws.StringValue(output.OperationId))

	if err != nil {
		return fmt.Errorf("error waiting for Service Discovery HTTP Namespace (%s) creation: %w", name, err)
	}

	namespaceID, ok := operation.Targets[servicediscovery.OperationTargetTypeNamespace]

	if !ok {
		return fmt.Errorf("error creating Service Discovery HTTP Namespace (%s): operation response missing Namespace ID", name)
	}

	d.SetId(aws.StringValue(namespaceID))

	return resourceHTTPNamespaceRead(d, meta)
}

func resourceHTTPNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &servicediscovery.GetNamespaceInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetNamespace(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Service Discovery HTTP Namespace (%s): %s", d.Id(), err)
	}

	arn := aws.StringValue(resp.Namespace.Arn)
	d.Set("name", resp.Namespace.Name)
	d.Set("description", resp.Namespace.Description)
	d.Set("arn", arn)

	tags, err := tftags.ServicediscoveryListTags(conn, arn)

	if err != nil {
		return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
	}

	tags = tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceHTTPNamespaceUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := tftags.ServicediscoveryUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Service Discovery HTTP Namespace (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceHTTPNamespaceRead(d, meta)
}

func resourceHTTPNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceDiscoveryConn

	input := &servicediscovery.DeleteNamespaceInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.DeleteNamespace(input)

	if tfawserr.ErrMessageContains(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Service Discovery HTTP Namespace (%s): %w", d.Id(), err)
	}

	if output != nil && output.OperationId != nil {
		if _, err := waiter.OperationSuccess(conn, aws.StringValue(output.OperationId)); err != nil {
			return fmt.Errorf("error waiting for Service Discovery HTTP Namespace (%s) deletion: %w", d.Id(), err)
		}
	}

	return nil
}
