package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/servicediscovery/waiter"
)

func resourceAwsServiceDiscoveryHttpNamespace() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsServiceDiscoveryHttpNamespaceCreate,
		Read:   resourceAwsServiceDiscoveryHttpNamespaceRead,
		Delete: resourceAwsServiceDiscoveryHttpNamespaceDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateServiceDiscoveryHttpNamespaceName,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsServiceDiscoveryHttpNamespaceCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	name := d.Get("name").(string)

	input := &servicediscovery.CreateHttpNamespaceInput{
		Name:             aws.String(name),
		CreatorRequestId: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	output, err := conn.CreateHttpNamespace(input)

	if err != nil {
		return fmt.Errorf("error creating Service Discovery HTTP Namespace (%s): %w", name, err)
	}

	if output == nil || output.OperationId == nil {
		return fmt.Errorf("error creating Service Discovery HTTP Namespace (%s): creation response missing Operation ID", name)
	}

	operationOutput, err := waiter.OperationSuccess(conn, aws.StringValue(output.OperationId))

	if err != nil {
		return fmt.Errorf("error waiting for Service Discovery HTTP Namespace (%s) creation: %w", name, err)
	}

	if operationOutput == nil || operationOutput.Operation == nil {
		return fmt.Errorf("error creating Service Discovery HTTP Namespace (%s): operation response missing Operation information", name)
	}

	namespaceID, ok := operationOutput.Operation.Targets[servicediscovery.OperationTargetTypeNamespace]

	if !ok {
		return fmt.Errorf("error creating Service Discovery HTTP Namespace (%s): operation response missing Namespace ID", name)
	}

	d.SetId(aws.StringValue(namespaceID))

	return resourceAwsServiceDiscoveryHttpNamespaceRead(d, meta)
}

func resourceAwsServiceDiscoveryHttpNamespaceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.GetNamespaceInput{
		Id: aws.String(d.Id()),
	}

	resp, err := conn.GetNamespace(input)
	if err != nil {
		if isAWSErr(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading Service Discovery HTTP Namespace (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.Namespace.Name)
	d.Set("description", resp.Namespace.Description)
	d.Set("arn", resp.Namespace.Arn)

	return nil
}

func resourceAwsServiceDiscoveryHttpNamespaceDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).sdconn

	input := &servicediscovery.DeleteNamespaceInput{
		Id: aws.String(d.Id()),
	}

	output, err := conn.DeleteNamespace(input)

	if isAWSErr(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
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
