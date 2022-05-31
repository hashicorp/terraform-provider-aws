package detective

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceInvitationAccepter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInvitationAccepterCreate,
		ReadContext:   resourceInvitationAccepterRead,
		DeleteContext: resourceInvitationAccepterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"graph_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceInvitationAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	graphArn := d.Get("graph_arn").(string)

	acceptInvitationInput := &detective.AcceptInvitationInput{
		GraphArn: aws.String(graphArn),
	}

	_, err := conn.AcceptInvitationWithContext(ctx, acceptInvitationInput)

	if err != nil {
		return diag.Errorf("error accepting Detective InvitationAccepter (%s): %s", d.Id(), err)
	}

	d.SetId(graphArn)

	return resourceInvitationAccepterRead(ctx, d, meta)
}

func resourceInvitationAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	graphArn, err := FindInvitationByGraphARN(ctx, conn, d.Id())

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Detective InvitationAccepter (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("error listing Detective InvitationAccepter (%s): %s", d.Id(), err)
	}

	if err != nil {
		return diag.Errorf("error reading Detective InvitationAccepter (%s): %s", d.Id(), err)
	}

	d.Set("graph_arn", graphArn)
	return nil
}

func resourceInvitationAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	input := &detective.DisassociateMembershipInput{
		GraphArn: aws.String(d.Id()),
	}

	_, err := conn.DisassociateMembershipWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("error disassociating Detective InvitationAccepter (%s): %s", d.Id(), err)
	}
	return nil
}
