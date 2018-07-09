package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func resourceAwsAppsyncApiKey() *schema.Resource {

	return &schema.Resource{
		Create: resourceAwsAppsyncApiKeyCreate,
		Read:   resourceAwsAppsyncApiKeyRead,
		Update: resourceAwsAppsyncApiKeyUpdate,
		Delete: resourceAwsAppsyncApiKeyDelete,

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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.ValidateRFC3339TimeString,
			},
			"key": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAwsAppsyncApiKeyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn
	params := &appsync.CreateApiKeyInput{
		ApiId:       aws.String(d.Get("api_id").(string)),
		Description: aws.String(d.Get("description").(string)),
	}
	if v, ok := d.GetOk("expires"); ok {
		t, _ := time.Parse(time.RFC3339, v.(string))
		params.Expires = aws.Int64(t.Unix())
	}
	resp, err := conn.CreateApiKey(params)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s:%s", d.Get("api_id").(string), *resp.ApiKey.Id))
	return resourceAwsAppsyncApiKeyRead(d, meta)
}

func resourceAwsAppsyncApiKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn
	var listKeys func(*appsync.ListApiKeysInput) (*appsync.ApiKey, error)
	ApiId, Id, er := decodeAppSyncApiKeyId(d.Id())
	if er != nil {
		return er
	}
	listKeys = func(input *appsync.ListApiKeysInput) (*appsync.ApiKey, error) {
		resp, err := conn.ListApiKeys(input)
		if err != nil {
			return nil, err
		}
		for _, v := range resp.ApiKeys {
			if *v.Id == Id {
				return v, nil
			}
		}
		if resp.NextToken != nil {
			listKeys(&appsync.ListApiKeysInput{
				ApiId:     aws.String(ApiId),
				NextToken: resp.NextToken,
			})
		}
		return nil, nil
	}
	key, err := listKeys(
		&appsync.ListApiKeysInput{
			ApiId: aws.String(ApiId),
		})
	if err != nil {
		return err
	}
	if key == nil {
		log.Printf("[WARN] AppSync API Key %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("key", key.Id)
	d.Set("description", key.Description)
	d.Set("expires", time.Unix(*key.Expires, 0).Format(time.RFC3339))
	return nil
}

func resourceAwsAppsyncApiKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn
	ApiId, Id, er := decodeAppSyncApiKeyId(d.Id())
	if er != nil {
		return er
	}
	params := &appsync.UpdateApiKeyInput{
		ApiId: aws.String(ApiId),
		Id:    aws.String(Id),
	}
	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}
	if d.HasChange("expires") {
		t, _ := time.Parse(time.RFC3339, d.Get("expires").(string))
		params.Expires = aws.Int64(t.Unix())
	}

	_, err := conn.UpdateApiKey(params)
	if err != nil {
		return err
	}

	return resourceAwsAppsyncApiKeyRead(d, meta)

}

func resourceAwsAppsyncApiKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn
	ApiId, Id, er := decodeAppSyncApiKeyId(d.Id())
	if er != nil {
		return er
	}
	input := &appsync.DeleteApiKeyInput{
		ApiId: aws.String(ApiId),
		Id:    aws.String(Id),
	}
	_, err := conn.DeleteApiKey(input)
	if err != nil {
		if isAWSErr(err, appsync.ErrCodeNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

func decodeAppSyncApiKeyId(id string) (string, string, error) {
	parts := strings.Split(id, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("Unexpected format of ID (%q), expected API-ID:API-KEY-ID", id)
	}
	return parts[0], parts[1], nil
}
