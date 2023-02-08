package ram

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceResourceShareAccepter() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceResourceShareAccepterCreate,
		ReadWithoutTimeout:   resourceResourceShareAccepterRead,
		DeleteWithoutTimeout: resourceResourceShareAccepterDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"share_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"invitation_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"receiver_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"sender_account_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_name": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"resources": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceResourceShareAccepterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn()

	shareARN := d.Get("share_arn").(string)

	invitation, err := FindResourceShareInvitationByResourceShareARNAndStatus(ctx, conn, shareARN, ram.ResourceShareInvitationStatusPending)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating RAM Resource Share Accepter: %s", err)
	}

	if invitation == nil || aws.StringValue(invitation.ResourceShareInvitationArn) == "" {
		return sdkdiag.AppendErrorf(diags, "No RAM Resource Share (%s) invitation found\n\n"+
			"NOTE: If both AWS accounts are in the same AWS Organization and RAM Sharing with AWS Organizations is enabled, this resource is not necessary",
			shareARN)
	}

	input := &ram.AcceptResourceShareInvitationInput{
		ClientToken:                aws.String(resource.UniqueId()),
		ResourceShareInvitationArn: invitation.ResourceShareInvitationArn,
	}

	log.Printf("[DEBUG] Accept RAM resource share invitation request: %s", input)
	output, err := conn.AcceptResourceShareInvitationWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "accepting RAM resource share invitation: %s", err)
	}

	d.SetId(shareARN)

	_, err = WaitResourceShareInvitationAccepted(ctx, conn,
		aws.StringValue(output.ResourceShareInvitation.ResourceShareInvitationArn),
		d.Timeout(schema.TimeoutCreate),
	)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM resource share (%s) state: %s", d.Id(), err)
	}

	return append(diags, resourceResourceShareAccepterRead(ctx, d, meta)...)
}

func resourceResourceShareAccepterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	accountID := meta.(*conns.AWSClient).AccountID
	conn := meta.(*conns.AWSClient).RAMConn()

	invitation, err := FindResourceShareInvitationByResourceShareARNAndStatus(ctx, conn, d.Id(), ram.ResourceShareInvitationStatusAccepted)

	if err != nil && !tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceShareInvitationArnNotFoundException) {
		return sdkdiag.AppendErrorf(diags, "retrieving invitation for resource share %s: %s", d.Id(), err)
	}

	if invitation != nil {
		d.Set("invitation_arn", invitation.ResourceShareInvitationArn)
		d.Set("receiver_account_id", invitation.ReceiverAccountId)
	} else {
		d.Set("receiver_account_id", accountID)
	}

	resourceShare, err := FindResourceShareOwnerOtherAccountsByARN(ctx, conn, d.Id())

	if !d.IsNewResource() && (tfawserr.ErrCodeEquals(err, ram.ErrCodeResourceArnNotFoundException) || tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException)) {
		log.Printf("[WARN] No RAM resource share with ARN (%s) found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "retrieving resource share: %s", err)
	}

	if resourceShare == nil {
		return sdkdiag.AppendErrorf(diags, "getting resource share (%s): empty result", d.Id())
	}

	d.Set("status", resourceShare.Status)
	d.Set("sender_account_id", resourceShare.OwningAccountId)
	d.Set("share_arn", resourceShare.ResourceShareArn)
	d.Set("share_id", resourceResourceShareGetIDFromARN(d.Id()))
	d.Set("share_name", resourceShare.Name)

	listInput := &ram.ListResourcesInput{
		MaxResults:        aws.Int64(500),
		ResourceOwner:     aws.String(ram.ResourceOwnerOtherAccounts),
		ResourceShareArns: aws.StringSlice([]string{d.Id()}),
	}

	var resourceARNs []*string
	err = conn.ListResourcesPagesWithContext(ctx, listInput, func(page *ram.ListResourcesOutput, lastPage bool) bool {
		for _, resource := range page.Resources {
			resourceARNs = append(resourceARNs, resource.Arn)
		}

		return !lastPage
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading RAM resource share resources %s: %s", d.Id(), err)
	}

	if err := d.Set("resources", flex.FlattenStringList(resourceARNs)); err != nil {
		return sdkdiag.AppendErrorf(diags, "unable to set resources: %s", err)
	}

	return diags
}

func resourceResourceShareAccepterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).RAMConn()

	receiverAccountID := d.Get("receiver_account_id").(string)

	if receiverAccountID == "" {
		return sdkdiag.AppendErrorf(diags, "The receiver account ID is required to leave a resource share")
	}

	input := &ram.DisassociateResourceShareInput{
		ClientToken:      aws.String(resource.UniqueId()),
		ResourceShareArn: aws.String(d.Id()),
		Principals:       []*string{aws.String(receiverAccountID)},
	}
	log.Printf("[DEBUG] Leave RAM resource share request: %s", input)

	_, err := conn.DisassociateResourceShareWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, ram.ErrCodeOperationNotPermittedException) {
		log.Printf("[WARN] Resource share could not be disassociated, but continuing: %s", err)
	}

	if err != nil && !tfawserr.ErrCodeEquals(err, ram.ErrCodeOperationNotPermittedException) {
		return sdkdiag.AppendErrorf(diags, "leaving RAM resource share: %s", err)
	}

	_, err = WaitResourceShareOwnedBySelfDisassociated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete))

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "waiting for RAM resource share (%s) state: %s", d.Id(), err)
	}

	return diags
}

func resourceResourceShareGetIDFromARN(arn string) string {
	return strings.Replace(arn[strings.LastIndex(arn, ":")+1:], "resource-share/", "rs-", -1)
}
