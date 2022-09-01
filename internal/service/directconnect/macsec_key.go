package directconnect

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceMacSecKey() *schema.Resource {
	return &schema.Resource{
		// MacSecKey resource only supports create (Associate), read (Describe) and delete (Disassociate)
		Create: resourceMacSecKeyCreate,
		Read:   resourceMacSecKeyRead,
		// You cannot modify a MACsec secret key after you associate it with a connection.
		// To modify the key, disassociate the key from the connection, and then associate
		// a new key with the connection
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
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-fA-F0-9]{64}$`), "Must be 64-character hex code string"),
			},
			"ckn": {
				Type:         schema.TypeString,
				Computed:     true,
				Optional:     true,
				AtLeastOneOf: []string{"ckn", "secret_arn"},
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`[a-fA-F0-9]{64}$`), "Must be 64-character hex code string"),
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
		ConnectionId: aws.String(d.Get("connection_id").(string)),
	}

	if d.Get("ckn").(string) != "" {
		input.Cak = aws.String(d.Get("cak").(string))
		input.Ckn = aws.String(d.Get("ckn").(string))
	}

	if d.Get("secret_arn").(string) != "" {
		input.SecretARN = aws.String(d.Get("secret_arn").(string))
	}

	log.Printf("[DEBUG] Creating MACSec secret key on Direct Connect Connection: %s", *input.ConnectionId)
	output, err := conn.AssociateMacSecKey(input)

	if err != nil {
		return fmt.Errorf("error creating MACSec secret key on Direct Connect Connection (%s): %w", *input.ConnectionId, err)
	}

	secret_arn := MacSecKeyParseSecretARN(output)

	// Create a composite ID based on connection ID and secret ARN
	d.SetId(fmt.Sprintf("%s/%s", secret_arn, aws.StringValue(output.ConnectionId)))

	d.Set("secret_arn", secret_arn)

	return nil
}

func resourceMacSecKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	secretArn, connId, err := MacSecKeyParseID(d.Id())

	connection, err := FindConnectionByID(conn, connId)

	if err != nil {
		return fmt.Errorf("error reading Direct Connect Connection (%s): %w", d.Id(), err)
	}

	if connection.MacSecKeys != nil {
		for _, key := range connection.MacSecKeys {
			if key.SecretARN == &secretArn {
				d.Set("ckn", key.Ckn)
				d.Set("connection_id", connId)
				d.Set("secret_arn", key.SecretARN)
				d.Set("start_on", key.StartOn)
				d.Set("state", key.State)
			}
		}
	} else {
		return fmt.Errorf("No MACSec keys found on Direct Connect Connection (%s)", d.Id())
	}

	return nil
}

func resourceMacSecKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DirectConnectConn

	input := &directconnect.DisassociateMacSecKeyInput{
		ConnectionId: aws.String(d.Get("connection_id").(string)),
		SecretARN:    aws.String(d.Get("secret_arn").(string)),
	}

	log.Printf("[DEBUG] Disassociating MACSec secret key on Direct Connect Connection: %s", *input.ConnectionId)
	_, err := conn.DisassociateMacSecKey(input)

	if err != nil {
		return fmt.Errorf("Unable to disassociate MACSec secret key on Direct Connect Connection (%s): %w", *input.ConnectionId, err)
	}

	// Disassociating the key does not delete it from Secrets Manager, do that here
	err = resourceMacSecKeySecretDelete(*input.SecretARN, meta)
	if err != nil {
		return fmt.Errorf("Unable to delete MACSec secret key %s: %w", *input.SecretARN, err)
	}

	return nil
}

func resourceMacSecKeySecretDelete(secretId string, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SecretsManagerConn
	input := &secretsmanager.DeleteSecretInput{
		SecretId: aws.String(secretId),
	}

	log.Printf("[DEBUG] Deleting MACSec secret key: %s", *input.SecretId)
	_, err := conn.DeleteSecret(input)

	if err != nil {
		return err
	}

	return nil
}

// MacSecKeyParseSecretARN parses the secret ARN returned from a CMK or secret_arn
func MacSecKeyParseSecretARN(output *directconnect.AssociateMacSecKeyOutput) string {
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

// MacSecKeyParseID parses the resource ID and returns the secret ARN and connection ID
func MacSecKeyParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected secretArn:connectionId", id)
	}

	return parts[0], parts[1], nil
}
