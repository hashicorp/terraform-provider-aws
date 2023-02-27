package securityhub

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Associated is the member status naming for regions that do not support Organizations
	memberStatusAssociated = "Associated"
	memberStatusInvited    = "Invited"
	memberStatusEnabled    = "Enabled"
	memberStatusResigned   = "Resigned"
)

func ResourceMember() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceMemberCreate,
		ReadWithoutTimeout:   resourceMemberRead,
		DeleteWithoutTimeout: resourceMemberDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"email": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"invite": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"master_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"member_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()
	log.Printf("[DEBUG] Creating Security Hub Member %s", d.Get("account_id").(string))

	resp, err := conn.CreateMembersWithContext(ctx, &securityhub.CreateMembersInput{
		AccountDetails: []*securityhub.AccountDetails{
			{
				AccountId: aws.String(d.Get("account_id").(string)),
				Email:     aws.String(d.Get("email").(string)),
			},
		},
	})

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Member %s: %s", d.Get("account_id").(string), err)
	}

	if len(resp.UnprocessedAccounts) > 0 {
		return sdkdiag.AppendErrorf(diags, "creating Security Hub Member %s: UnprocessedAccounts is not empty", d.Get("account_id").(string))
	}

	d.SetId(d.Get("account_id").(string))

	if d.Get("invite").(bool) {
		log.Printf("[INFO] Inviting Security Hub Member %s", d.Id())
		iresp, err := conn.InviteMembersWithContext(ctx, &securityhub.InviteMembersInput{
			AccountIds: []*string{aws.String(d.Get("account_id").(string))},
		})

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "inviting Security Hub Member %s: %s", d.Id(), err)
		}

		if len(iresp.UnprocessedAccounts) > 0 {
			return sdkdiag.AppendErrorf(diags, "inviting Security Hub Member %s: UnprocessedAccounts is not empty", d.Id())
		}
	}

	return append(diags, resourceMemberRead(ctx, d, meta)...)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()

	log.Printf("[DEBUG] Reading Security Hub Member %s", d.Id())
	resp, err := conn.GetMembersWithContext(ctx, &securityhub.GetMembersInput{
		AccountIds: []*string{aws.String(d.Id())},
	})

	if err != nil {
		if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
			log.Printf("[WARN] Security Hub Member (%s) not found, removing from state", d.Id())
			d.SetId("")
			return diags
		}
		return sdkdiag.AppendErrorf(diags, "reading Security Hub Member (%s): %s", d.Id(), err)
	}

	if len(resp.Members) == 0 {
		log.Printf("[WARN] Security Hub Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	member := resp.Members[0]

	d.Set("account_id", member.AccountId)
	d.Set("email", member.Email)
	d.Set("master_id", member.MasterId)

	status := aws.StringValue(member.MemberStatus)
	d.Set("member_status", status)

	invited := status == memberStatusInvited || status == memberStatusEnabled || status == memberStatusAssociated || status == memberStatusResigned
	d.Set("invite", invited)

	return diags
}

func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).SecurityHubConn()

	_, err := conn.DisassociateMembersWithContext(ctx, &securityhub.DisassociateMembersInput{
		AccountIds: []*string{aws.String(d.Id())},
	})
	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "disassociating Security Hub Member %s: %s", d.Id(), err)
	}

	resp, err := conn.DeleteMembersWithContext(ctx, &securityhub.DeleteMembersInput{
		AccountIds: []*string{aws.String(d.Id())},
	})

	if tfawserr.ErrCodeEquals(err, securityhub.ErrCodeResourceNotFoundException) {
		return diags
	}
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Member %s: %s", d.Id(), err)
	}

	if len(resp.UnprocessedAccounts) > 0 {
		return sdkdiag.AppendErrorf(diags, "deleting Security Hub Member %s: UnprocessedAccounts is not empty", d.Get("account_id").(string))
	}

	return diags
}
