package sesv2

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceEmailIdentity() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceEmailIdentityCreate,
		ReadContext:   resourceEmailIdentityRead,
		UpdateContext: resourceEmailIdentityUpdate,
		DeleteContext: resourceEmailIdentityDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				StateFunc: func(v interface{}) string {
					return strings.TrimSuffix(v.(string), ".")
				},
			},
			"default_configuration_set": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceEmailIdentityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	email := d.Get("email").(string)
	email = strings.TrimSuffix(email, ".")

	createOpts := &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(email),
		Tags:          Tags(tags.IgnoreAWS()),
	}

	if v, ok := d.GetOk("default_configuration_set"); ok {
		createOpts.ConfigurationSetName = aws.String(v.(string))
	}

	_, err := conn.CreateEmailIdentity(ctx, createOpts)
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionCreating, ResEmailIdentity, d.Id(), err)
	}

	d.SetId(email)

	return resourceEmailIdentityRead(ctx, d, meta)
}

func resourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	email := d.Id()
	d.Set("email", email)
	getOpts := &sesv2.GetEmailIdentityInput{EmailIdentity: aws.String(email)}
	response, err := conn.GetEmailIdentity(ctx, getOpts)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Email Identity (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, ResEmailIdentity, d.Id(), err)
	}

	arn := arn.ARN{
		AccountID: meta.(*conns.AWSClient).AccountID,
		Partition: meta.(*conns.AWSClient).Partition,
		Region:    meta.(*conns.AWSClient).Region,
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
		Service:   "ses",
	}.String()

	if err := d.Set("arn", arn); err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, ResEmailIdentity, d.Id(), err)
	}

	if err := d.Set("default_configuration_set", response.ConfigurationSetName); err != nil {
		return create.DiagError(names.SESV2, create.ErrActionReading, ResEmailIdentity, d.Id(), err)
	}

	tags := KeyValueTags(response.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return create.DiagError(names.SESV2, create.ErrActionSetting, ResEmailIdentity, d.Id(), err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return create.DiagError(names.SESV2, create.ErrActionSetting, ResEmailIdentity, d.Id(), err)
	}

	return nil
}

func resourceEmailIdentityUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	email := d.Get("email").(string)

	if d.HasChange("ConfigurationSetName") {
		input := &sesv2.PutEmailIdentityConfigurationSetAttributesInput{EmailIdentity: aws.String(email)}

		_, err := conn.PutEmailIdentityConfigurationSetAttributes(ctx, input)
		if err != nil {
			return create.DiagError(names.SESV2, create.ErrActionUpdating, ResEmailIdentity, d.Id(), err)
		}
	}

	return resourceEmailIdentityRead(ctx, d, meta)
}

func resourceEmailIdentityDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	email := d.Get("email").(string)

	deleteOps := &sesv2.DeleteEmailIdentityInput{EmailIdentity: aws.String(email)}
	_, err := conn.DeleteEmailIdentity(ctx, deleteOps)

	if tfawserr.ErrCodeEquals(err, (&types.NotFoundException{}).ErrorMessage()) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionDeleting, ResEmailIdentity, d.Id(), err)
	}

	return nil
}
