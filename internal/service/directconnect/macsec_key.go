package directconnect

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceMacSecKey() *schema.Resource {
	return &schema.Resource{
		// MacSecKey resource only supports create (Associate) and delete (Disassociate)
		Create: resourceMacSecKeyCreate,
		// Read is performed via DescribeConnection
		Read: schema.Noop,
		// No update functionality
		Update: schema.Noop,
		Delete: resourceMacSecKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"cak": {
				Type:     schema.TypeString,
				Optional: true,
				// CAK requires CKN
				RequiredWith: []string{"ckn"},
			},
			"ckn": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"ckn", "secret_arn"},
			},
			"connection_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"secret_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				AtLeastOneOf: []string{"ckn", "secret_arn"},
			},
			"start_on": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacSecKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	input := &directconnect.AssociateMacSecKeyInput{
		Cak:          aws.String(d.Get("cak").(string)),
		Ckn:          aws.String(d.Get("ckn").(string)),
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		SecretARN:    aws.String(d.Get("secret_arn").(string)),
	}

	log.Printf("[DEBUG] Creating MACSec secret key on Direct Connect Connection: %s", *input.ConnectionId)
	output, err := conn.AssociateMacSecKey(input)

	if err != nil {
		return fmt.Errorf("error creating MACSec secret key on Direct Connect Connection (%s): %w", *input.ConnectionId, err)
	}

	secret_arn := MacSecKeyParseSecretArn(output)

	// Create a composite ID based on connection ID and secret ARN
	d.SetId(fmt.Sprintf("%s:%s", secret_arn, aws.StringValue(output.ConnectionId)))

	return nil
}

func resourceMacSecKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	input := &directconnect.DisassociateMacSecKeyInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		SecretARN:    aws.String(d.Get("secret_arn").(string)),
	}

	log.Printf("[DEBUG] Deleting MACSec secret key on Direct Connect Connection: %s", *input.ConnectionId)
	_, err := conn.DisassociateMacSecKey(input)

	if err != nil {
		return fmt.Errorf("error delecting MACSec secret key on Direct Connect Connection (%s): %w", *input.ConnectionId, err)
	}

	return nil
}

// MacSecKeyParseSecretArn parses the secret ARN returned from a CMK or secret_arn
func MacSecKeyParseSecretArn(output *directconnect.AssociateMacSecKeyOutput) string {
	var result string

	for _, key := range output.MacSecKeys {
		if key == nil {
			continue
		}
		if key != nil {
			result = *key.SecretARN
		}
	}
	return result
}

// Define an exported function to parse the resource ID
func MacSecKeyParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, ":", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected :connectionId", id)
	}

	return parts[0], parts[1], nil
}
