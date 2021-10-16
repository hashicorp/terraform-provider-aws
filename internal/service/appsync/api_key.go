package appsync

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func ResourceAPIKey() *schema.Resource {

	return &schema.Resource{
		Create: resourceAPIKeyCreate,
		Read:   resourceAPIKeyRead,
		Update: resourceAPIKeyUpdate,
		Delete: resourceAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "Managed by Terraform",
			},
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"expires": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					// Ignore unsetting value
					if old != "" && new == "" {
						return true
					}
					return false
				},
				ValidateFunc: validation.IsRFC3339Time,
			},
			"key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAPIKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID := d.Get("api_id").(string)

	params := &appsync.CreateApiKeyInput{
		ApiId:       aws.String(apiID),
		Description: aws.String(d.Get("description").(string)),
	}
	if v, ok := d.GetOk("expires"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		params.Expires = aws.Int64(t.Unix())
	}
	resp, err := conn.CreateApiKey(params)
	if err != nil {
		return fmt.Errorf("error creating Appsync API Key: %s", err)
	}

	d.SetId(fmt.Sprintf("%s:%s", apiID, aws.StringValue(resp.ApiKey.Id)))
	return resourceAPIKeyRead(d, meta)
}

func resourceAPIKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, keyID, err := DecodeAPIKeyID(d.Id())
	if err != nil {
		return err
	}

	key, err := GetAPIKey(apiID, keyID, conn)
	if err != nil {
		return fmt.Errorf("error getting Appsync API Key %q: %s", d.Id(), err)
	}
	if key == nil {
		log.Printf("[WARN] AppSync API Key %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("api_id", apiID)
	d.Set("key", key.Id)
	d.Set("description", key.Description)
	d.Set("expires", time.Unix(aws.Int64Value(key.Expires), 0).UTC().Format(time.RFC3339))
	return nil
}

func resourceAPIKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, keyID, err := DecodeAPIKeyID(d.Id())
	if err != nil {
		return err
	}

	params := &appsync.UpdateApiKeyInput{
		ApiId: aws.String(apiID),
		Id:    aws.String(keyID),
	}
	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("expires") {
		t, _ := time.Parse(time.RFC3339, d.Get("expires").(string))
		params.Expires = aws.Int64(t.Unix())
	}

	_, err = conn.UpdateApiKey(params)
	if err != nil {
		return err
	}

	return resourceAPIKeyRead(d, meta)

}

func resourceAPIKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, keyID, err := DecodeAPIKeyID(d.Id())
	if err != nil {
		return err
	}

	input := &appsync.DeleteApiKeyInput{
		ApiId: aws.String(apiID),
		Id:    aws.String(keyID),
	}
	_, err = conn.DeleteApiKey(input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, appsync.ErrCodeNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

func DecodeAPIKeyID(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected API-ID:API-KEY-ID", id)
	}
	return parts[0], parts[1], nil
}

func GetAPIKey(apiID, keyID string, conn *appsync.AppSync) (*appsync.ApiKey, error) {
	input := &appsync.ListApiKeysInput{
		ApiId: aws.String(apiID),
	}
	for {
		resp, err := conn.ListApiKeys(input)
		if err != nil {
			return nil, err
		}
		for _, apiKey := range resp.ApiKeys {
			if aws.StringValue(apiKey.Id) == keyID {
				return apiKey, nil
			}
		}
		if resp.NextToken == nil {
			break
		}
		input.NextToken = resp.NextToken
	}
	return nil, nil
}
