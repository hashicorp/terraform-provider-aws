package appsync

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func ResourceAPICache() *schema.Resource {

	return &schema.Resource{
		Create: resourceAPICacheCreate,
		Read:   resourceAPICacheRead,
		Update: resourceAPICacheUpdate,
		Delete: resourceAPICacheDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"api_caching_behavior": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appsync.ApiCachingBehavior_Values(), false),
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(appsync.ApiCacheType_Values(), false),
			},
			"ttl": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"at_rest_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"transit_encryption_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAPICacheCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID := d.Get("api_id").(string)

	params := &appsync.CreateApiCacheInput{
		ApiId:              aws.String(apiID),
		Type:               aws.String(d.Get("type").(string)),
		ApiCachingBehavior: aws.String(d.Get("api_caching_behavior").(string)),
		Ttl:                aws.Int64(int64(d.Get("ttl").(int))),
	}

	if v, ok := d.GetOk("at_rest_encryption_enabled"); ok {
		params.AtRestEncryptionEnabled = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("transit_encryption_enabled"); ok {
		params.TransitEncryptionEnabled = aws.Bool(v.(bool))
	}

	_, err := conn.CreateApiCache(params)
	if err != nil {
		return fmt.Errorf("error creating Appsync API Cache: %w", err)
	}

	d.SetId(apiID)

	if err := waitAPICacheAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Appsync API Cache (%s) availability: %w", d.Id(), err)
	}

	return resourceAPICacheRead(d, meta)
}

func resourceAPICacheRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	cache, err := FindAPICacheByID(conn, d.Id())
	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AppSync API Cache (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Appsync API Cache %q: %s", d.Id(), err)
	}

	d.Set("api_id", d.Id())
	d.Set("type", cache.Type)
	d.Set("api_caching_behavior", cache.ApiCachingBehavior)
	d.Set("ttl", cache.Ttl)
	d.Set("at_rest_encryption_enabled", cache.AtRestEncryptionEnabled)
	d.Set("transit_encryption_enabled", cache.TransitEncryptionEnabled)

	return nil
}

func resourceAPICacheUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	params := &appsync.UpdateApiCacheInput{
		ApiId: aws.String(d.Id()),
	}

	if d.HasChange("type") {
		params.Type = aws.String(d.Get("type").(string))
	}

	if d.HasChange("api_caching_behavior") {
		params.ApiCachingBehavior = aws.String(d.Get("api_caching_behavior").(string))
	}

	if d.HasChange("ttl") {
		params.Ttl = aws.Int64(int64(d.Get("ttl").(int)))
	}

	_, err := conn.UpdateApiCache(params)
	if err != nil {
		return fmt.Errorf("error updating Appsync API Cache %q: %w", d.Id(), err)
	}

	if err := waitAPICacheAvailable(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Appsync API Cache (%s) availability: %w", d.Id(), err)
	}

	return resourceAPICacheRead(d, meta)

}

func resourceAPICacheDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	input := &appsync.DeleteApiCacheInput{
		ApiId: aws.String(d.Id()),
	}
	_, err := conn.DeleteApiCache(input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, appsync.ErrCodeNotFoundException) {
			return nil
		}
		return fmt.Errorf("error deleting Appsync API Cache: %w", err)
	}

	if err := waitAPICacheDeleted(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for Appsync API Cache (%s) to be deleted: %w", d.Id(), err)
	}

	return nil
}
