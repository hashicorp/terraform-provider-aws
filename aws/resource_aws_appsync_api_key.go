package aws

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform/helper/schema"
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
			"appsync_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"valid_till_date": {
				Type:          schema.TypeString,
				ConflictsWith: []string{"validity_period_days"},
				Optional:      true,
				ValidateFunc: func(v interface{}, k string) (ws []string, es []error) {
					value := v.(string)
					// reference - http://www.regexlib.com/REDetails.aspx?regexp_id=409
					if !regexp.MustCompile(`^(((0[1-9]|[12]\d|3[01])\/(0[13578]|1[02])\/((1[6-9]|[2-9]\d)\d{2}))|((0[1-9]|[12]\d|30)\/(0[13456789]|1[012])\/((1[6-9]|[2-9]\d)\d{2}))|((0[1-9]|1\d|2[0-8])\/02\/((1[6-9]|[2-9]\d)\d{2}))|(29\/02\/((1[6-9]|[2-9]\d)(0[48]|[2468][048]|[13579][26])|((16|[2468][048]|[3579][26])00))))$`).MatchString(value) {
						es = append(es, fmt.Errorf(
							"only dd/mm/yyyy in %q", k))
					}
					return
				},
			},
			"validity_period_days": {
				Type:          schema.TypeInt,
				ConflictsWith: []string{"valid_till_date"},
				Optional:      true,
			},
			"expiry_date": {
				Type:     schema.TypeString,
				Computed: true,
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
		ApiId:       aws.String(d.Get("appsync_api_id").(string)),
		Description: aws.String(d.Get("description").(string)),
	}
	layout := "02/01/2006 15:04:05 -0700 MST"
	if v, ok := d.GetOk("validity_period_days"); ok {
		params.Expires = aws.Int64(time.Now().Add(time.Hour * 24 * time.Duration(v.(int))).Unix())
	}
	if v, ok := d.GetOk("valid_till_date"); ok {
		tx := strings.Split(time.Now().Format(layout), " ")
		tx[0] = v.(string)
		t, _ := time.Parse(layout, strings.Join(tx, " "))
		params.Expires = aws.Int64(t.Unix())
	}

	resp, err := conn.CreateApiKey(params)
	if err != nil {
		return err
	}

	d.SetId(*resp.ApiKey.Id)
	return resourceAwsAppsyncApiKeyRead(d, meta)
}

func resourceAwsAppsyncApiKeyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.ListApiKeysInput{
		ApiId: aws.String(d.Get("appsync_api_id").(string)),
	}

	resp, err := conn.ListApiKeys(input)
	if err != nil {
		return err
	}
	var key appsync.ApiKey
	for _, v := range resp.ApiKeys {
		if *v.Id == d.Id() {
			key = *v
		}
	}

	d.Set("key", key.Id)
	d.Set("description", key.Description)
	d.Set("expiry_date", time.Unix(*key.Expires, 0).String())
	return nil
}

func resourceAwsAppsyncApiKeyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	params := &appsync.UpdateApiKeyInput{
		ApiId: aws.String(d.Get("appsync_api_id").(string)),
		Id:    aws.String(d.Id()),
	}
	if d.HasChange("description") {
		params.Description = aws.String(d.Get("description").(string))
	}
	if v, ok := d.GetOk("validity_period_days"); ok {

		if d.HasChange("validity_period_days") {
			params.Expires = aws.Int64(time.Now().Add(time.Hour * 24 * time.Duration(v.(int))).Unix())
		}
	}
	if v, ok := d.GetOk("valid_till_date"); ok {
		layout := "02/01/2006 15:04:05 -0700 MST"
		if d.HasChange("valid_till_date") {
			tx := strings.Split(time.Now().Format(layout), " ")
			tx[0] = v.(string)
			t, _ := time.Parse(layout, strings.Join(tx, " "))
			params.Expires = aws.Int64(t.Unix())

		}
	}

	_, err := conn.UpdateApiKey(params)
	if err != nil {
		return err
	}

	return resourceAwsAppsyncApiKeyRead(d, meta)

}

func resourceAwsAppsyncApiKeyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appsyncconn

	input := &appsync.DeleteApiKeyInput{
		ApiId: aws.String(d.Get("appsync_api_id").(string)),
		Id:    aws.String(d.Id()),
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
