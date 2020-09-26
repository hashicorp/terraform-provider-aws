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

func resourceAwsSsoAssignment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsSsoAssignmentCreate,
		Read:   resourceAwsSsoAssignmentRead,
		Update: resourceAwsSsoAssignmentUpdate,
		Delete: resourceAwsSsoAssignmentDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"created_date": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"failure_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
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

			"permission_set_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(10, 1224),
					validation.StringMatch(regexp.MustCompile(`^arn:aws:sso:::permissionSet/(sso)?ins-[a-zA-Z0-9-.]{16}/ps-[a-zA-Z0-9-./]{16}$`), "must match arn:aws:sso:::permissionSet/(sso)?ins-[a-zA-Z0-9-.]{16}/ps-[a-zA-Z0-9-./]{16}"),
				),
			},

			"principal_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 47),
					validation.StringMatch(regexp.MustCompile(`^([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}$`), "must match ([0-9a-f]{10}-|)[A-Fa-f0-9]{8}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{4}-[A-Fa-f0-9]{12}"),
				),
			},

			"principal_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"USER", "GROUP"}, false),
			},

			"target_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validateAwsAccountId,
			},

			"target_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				Default:      "AWS_ACCOUNT",
				ValidateFunc: validation.StringInSlice([]string{"AWS_ACCOUNT"}, false),
			},
		},
	}
}

func resourceAwsSsoAssignmentCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	// d.SetId(*resp.PermissionSetArn)
	return resourceAwsSsoAssignmentRead(d, meta)
}

func resourceAwsSsoAssignmentRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return nil
}

func resourceAwsSsoAssignmentUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return resourceAwsSsoAssignmentRead(d, meta)
}

func resourceAwsSsoAssignmentDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ssoadminconn
	// TODO
	return nil
}

// func waitForAssignmentCreation(conn *identitystore.IdentityStore, instanceArn string, requestId string) error {
// }

// func waitForAssignmentDeletion(conn *identitystore.IdentityStore, instanceArn string, requestId string) error {
// }
