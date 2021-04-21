package aws

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"log"
	"time"
)

const (
	errorMacie2CustomDataIdentifierCreate  = "error creating Macie2 CustomDataIdentifier: %s"
	errorMacie2CustomDataIdentifierRead    = "error reading Macie2 CustomDataIdentifier (%s): %w"
	errorMacie2CustomDataIdentifierDelete  = "error deleting Macie2 CustomDataIdentifier (%s): %w"
	errorMacie2CustomDataIdentifierSetting = "error setting `%s` for Macie2 CustomDataIdentifier (%s): %s"
)

func resourceAwsMacie2CustomDataIdentifier() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMacie2CustomDataIdentifierCreate,
		ReadContext:   resourceMacie2CustomDataIdentifierRead,
		UpdateContext: resourceMacie2CustomDataIdentifierUpdate,
		DeleteContext: resourceMacie2CustomDataIdentifierDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"regex": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"keywords": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"client_token": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ignore_words": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"maximum_match_distance": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"tags": tagsSchemaComputed(),
			"deleted": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMacie2CustomDataIdentifierCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.CreateCustomDataIdentifierInput{}

	if v, ok := d.GetOk("regex"); ok {
		input.SetRegex(v.(string))
	}
	if v, ok := d.GetOk("keywords"); ok {
		input.SetKeywords(expandStringList(v.([]interface{})))
	}
	if v, ok := d.GetOk("client_token"); ok {
		input.SetClientToken(v.(string))
	}
	if v, ok := d.GetOk("ignore_words"); ok {
		input.SetIgnoreWords(expandStringList(v.([]interface{})))
	}
	if v, ok := d.GetOk("name"); ok {
		input.SetName(v.(string))
	}
	if v, ok := d.GetOk("description"); ok {
		input.SetDescription(v.(string))
	}
	if v, ok := d.GetOk("maximum_match_distance"); ok {
		input.SetMaximumMatchDistance(int64(v.(int)))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.SetTags(keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().AppsyncTags())
	}

	log.Printf("[DEBUG] Creating Macie2 CustomDataIdentifier: %v", input)

	var err error
	var output macie2.CreateCustomDataIdentifierOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		resp, err := conn.CreateCustomDataIdentifierWithContext(ctx, input)
		if err != nil {
			if isAWSErr(err, macie2.ErrorCodeClientError, "") {
				log.Printf(errorMacie2CustomDataIdentifierCreate, err)
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		output = *resp

		return nil
	})

	if isResourceTimeoutError(err) {
		_, _ = conn.CreateCustomDataIdentifierWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierCreate, err))
	}

	d.SetId(aws.StringValue(output.CustomDataIdentifierId))

	return resourceMacie2CustomDataIdentifierRead(ctx, d, meta)
}

func resourceMacie2CustomDataIdentifierRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig
	input := &macie2.GetCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading Macie2 CustomDataIdentifier: %s", input)
	resp, err := conn.GetCustomDataIdentifierWithContext(ctx, input)
	if err != nil {
		if isAWSErr(err, macie2.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Macie2 CustomDataIdentifier does not exist, removing from state: %s", d.Id())
			d.SetId("")
			return nil
		}
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierRead, d.Id(), err))
	}

	if err = d.Set("regex", aws.StringValue(resp.Regex)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "regex", d.Id(), err))
	}
	if err = d.Set("keywords", flattenStringList(resp.Keywords)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "keywords", d.Id(), err))
	}
	if err = d.Set("ignore_words", flattenStringList(resp.IgnoreWords)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "ignore_words", d.Id(), err))
	}
	if err = d.Set("name", aws.StringValue(resp.Name)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "name", d.Id(), err))
	}
	if err = d.Set("description", aws.StringValue(resp.Description)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "description", d.Id(), err))
	}
	if err = d.Set("maximum_match_distance", aws.Int64Value(resp.MaximumMatchDistance)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "maximum_match_distance", d.Id(), err))
	}
	if err = d.Set("tags", keyvaluetags.AppsyncKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "tags", d.Id(), err))
	}
	if err = d.Set("deleted", aws.BoolValue(resp.Deleted)); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "deleted", d.Id(), err))
	}
	if err = d.Set("created_at", resp.CreatedAt.String()); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "created_at", d.Id(), err))
	}
	if err = d.Set("arn", resp.Arn); err != nil {
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierSetting, "arn", d.Id(), err))
	}

	return nil
}

func resourceMacie2CustomDataIdentifierUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceMacie2CustomDataIdentifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting Macie2 CustomDataIdentifier: %s", input)
	_, err := conn.DeleteCustomDataIdentifierWithContext(ctx, input)
	if err != nil {
		if isAWSErr(err, macie2.ErrorCodeInternalError, "") {
			return nil
		}
		return diag.FromErr(fmt.Errorf(errorMacie2CustomDataIdentifierDelete, d.Id(), err))
	}
	return nil
}
