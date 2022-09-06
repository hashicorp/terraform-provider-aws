package ses

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sesv2"
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
		},
	}
}

func resourceEmailIdentityCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	email := d.Get("email").(string)
	email = strings.TrimSuffix(email, ".")

	createOpts := &sesv2.CreateEmailIdentityInput{
		EmailIdentity: aws.String(email),
	}

	if v, ok := d.GetOk("default_configuration_set"); ok {
		createOpts.ConfigurationSetName = aws.String(v.(string))
	}

	_, err := conn.CreateEmailIdentityWithContext(ctx, createOpts)
	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionCreating, ResEmailIdentity, d.Id(), err)
	}

	d.SetId(email)

	return resourceEmailIdentityRead(ctx, d, meta)
}

func resourceEmailIdentityRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	email := d.Id()
	d.Set("email", email)
	getOpts := &sesv2.GetEmailIdentityInput{EmailIdentity: aws.String(email)}
	response, err := conn.GetEmailIdentityWithContext(ctx, getOpts)

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

	return nil
}

func resourceEmailIdentityUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).SESV2Conn

	email := d.Get("email").(string)

	if d.HasChange("ConfigurationSetName") {
		input := &sesv2.PutEmailIdentityConfigurationSetAttributesInput{EmailIdentity: aws.String(email)}

		_, err := conn.PutEmailIdentityConfigurationSetAttributesWithContext(ctx, input)
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
	_, err := conn.DeleteEmailIdentityWithContext(ctx, deleteOps)

	if tfawserr.ErrCodeEquals(err, sesv2.ErrCodeNotFoundException) {
		return nil
	}

	if err != nil {
		return create.DiagError(names.SESV2, create.ErrActionDeleting, ResEmailIdentity, d.Id(), err)
	}

	return nil
}
