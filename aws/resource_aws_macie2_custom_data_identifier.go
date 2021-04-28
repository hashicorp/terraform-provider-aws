package aws

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"
)

func resourceAwsMacie2CustomDataIdentifier() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMacie2CustomDataIdentifierCreate,
		ReadWithoutTimeout:   resourceMacie2CustomDataIdentifierRead,
		UpdateWithoutTimeout: resourceMacie2CustomDataIdentifierUpdate,
		DeleteWithoutTimeout: resourceMacie2CustomDataIdentifierDelete,
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
			"ignore_words": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name_prefix"},
			},
			"name_prefix": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
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

	input := &macie2.CreateCustomDataIdentifierInput{
		ClientToken: aws.String(resource.UniqueId()),
	}

	if v, ok := d.GetOk("regex"); ok {
		input.SetRegex(v.(string))
	}
	if v, ok := d.GetOk("keywords"); ok {
		input.SetKeywords(expandStringList(v.([]interface{})))
	}
	if v, ok := d.GetOk("ignore_words"); ok {
		input.SetIgnoreWords(expandStringList(v.([]interface{})))
	}
	input.SetName(naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string)))
	if v, ok := d.GetOk("description"); ok {
		input.SetDescription(v.(string))
	}
	if v, ok := d.GetOk("maximum_match_distance"); ok {
		input.SetMaximumMatchDistance(int64(v.(int)))
	}
	if v, ok := d.GetOk("tags"); ok {
		input.SetTags(keyvaluetags.New(v.(map[string]interface{})).IgnoreAws().AppsyncTags())
	}

	var err error
	var output macie2.CreateCustomDataIdentifierOutput
	err = resource.RetryContext(ctx, 4*time.Minute, func() *resource.RetryError {
		resp, err := conn.CreateCustomDataIdentifierWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeClientError) {
				return resource.RetryableError(err)
			}

			return resource.NonRetryableError(err)
		}
		output = *resp

		return nil
	})

	if isResourceTimeoutError(err) {
		_, err = conn.CreateCustomDataIdentifierWithContext(ctx, input)
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Macie CustomDataIdentifier: %w", err))
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

	resp, err := conn.GetCustomDataIdentifierWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, macie2.ErrCodeAccessDeniedException) {
		log.Printf("[WARN] Macie CustomDataIdentifier does not exist, removing from state: %s", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading Macie CustomDataIdentifier (%s): %w", d.Id(), err))
	}

	d.Set("regex", aws.StringValue(resp.Regex))
	if err = d.Set("keywords", flattenStringList(resp.Keywords)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie CustomDataIdentifier (%s): %w", "keywords", d.Id(), err))
	}
	if err = d.Set("ignore_words", flattenStringList(resp.IgnoreWords)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie CustomDataIdentifier (%s): %w", "ignore_words", d.Id(), err))
	}
	d.Set("name", aws.StringValue(resp.Name))
	d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(resp.Name)))
	d.Set("description", aws.StringValue(resp.Description))
	d.Set("maximum_match_distance", aws.Int64Value(resp.MaximumMatchDistance))
	if err = d.Set("tags", keyvaluetags.AppsyncKeyValueTags(resp.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting `%s` for Macie CustomDataIdentifier (%s): %w", "tags", d.Id(), err))
	}
	d.Set("deleted", aws.BoolValue(resp.Deleted))
	d.Set("created_at", aws.TimeValue(resp.CreatedAt).Format(time.RFC3339))
	d.Set("arn", aws.StringValue(resp.Arn))

	return nil
}

func resourceMacie2CustomDataIdentifierUpdate(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}

func resourceMacie2CustomDataIdentifierDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*AWSClient).macie2conn

	input := &macie2.DeleteCustomDataIdentifierInput{
		Id: aws.String(d.Id()),
	}

	_, err := conn.DeleteCustomDataIdentifierWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, macie2.ErrorCodeInternalError) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting Macie CustomDataIdentifier (%s): %w", d.Id(), err))
	}
	return nil
}
