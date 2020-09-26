package aws

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
)

func resourceAwsSsoPermissionSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsoPermissionSetCreate,
		Read:   resourceAwsSsoPermissionSetRead,
		Update: resourceAwsSsoPermissionSetUpdate,
		Delete: resourceAwsSsoPermissionSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"instance_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 1224),
					validation.StringMatch(regexp.MustCompile(`^arn:aws:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}$`), "must match arn:aws:sso:::instance/(sso)?ins-[a-zA-Z0-9-.]{16}"),
				),
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 32),
					validation.StringMatch(regexp.MustCompile(`^[\w+=,.@-]+$`), "must match [\\w+=,.@-]"),
				),
			},

			"description": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 700),
					validation.StringMatch(regexp.MustCompile(`^[\p{L}\p{M}\p{Z}\p{S}\p{N}\p{P}]*$`), "must match [\\p{L}\\p{M}\\p{Z}\\p{S}\\p{N}\\p{P}]"),
				),
			},

			"session_duration": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringLenBetween(1, 100),
			},

			"relay_state": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 240),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9&$@#\\\/%?=~\-_'"|!:,.;*+\[\]\(\)\{\} ]+$`), "must match [a-zA-Z0-9&$@#\\\\\\/%?=~\\-_'\"|!:,.;*+\\[\\]\\(\\)\\{\\} ]"),
				),
			},

			"inline_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validateIAMPolicyJson,
				DiffSuppressFunc: suppressEquivalentAwsPolicyDiffs,
			},

			"managed_policies": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validateArn,
				},
			},

			"tags": tagsSchema(),
		},
	}
}

func resourceAwsSsoPermissionSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	// d.SetId(*resp.PermissionSetArn)
	return resourceAwsSsoPermissionSetRead(d, meta)
}

func resourceAwsSsoPermissionSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return nil
}

func resourceAwsSsoPermissionSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return resourceAwsSsoPermissionSetRead(d, meta)
}

func resourceAwsSsoPermissionSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return nil
}

// func waitForPermissionSetProvisioning(conn *identitystore.IdentityStore, arn string) error {
// }
