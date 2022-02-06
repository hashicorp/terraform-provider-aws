package iam

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceOpenIDConnectProviderClientID() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpenIDConnectProviderClientIDAdd,
		Delete: resourceOpenIDConnectProviderClientIDRemove,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"client_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},
		},
	}
}

func resourceOpenIDConnectProviderClientIDAdd(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	clientID := d.Get("client_id").(string)
	arn := d.Get("arn").(string)

	input := &iam.AddClientIDToOpenIDConnectProviderInput{
		ClientID:                 aws.String(clientID),
		OpenIDConnectProviderArn: aws.String(arn),
	}

	_, err := conn.AddClientIDToOpenIDConnectProvider(input)
	if err != nil {
		return fmt.Errorf("error adding ClientID to IAM OIDC Provider (%s): %w", arn, err)
	}

	return nil
}

func resourceOpenIDConnectProviderClientIDRemove(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).IAMConn

	clientID := d.Get("client_id").(string)
	arn := d.Get("arn").(string)

	input := &iam.RemoveClientIDFromOpenIDConnectProviderInput{
		ClientID:                 aws.String(clientID),
		OpenIDConnectProviderArn: aws.String(arn),
	}
	_, err := conn.RemoveClientIDFromOpenIDConnectProvider(input)
	if err != nil {
		return fmt.Errorf("error removing ClientID from IAM OIDC Provider (%s): %w", d.Id(), err)
	}

	return nil
}
