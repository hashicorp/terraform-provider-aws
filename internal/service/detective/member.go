package detective

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceMember() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMemberCreate,
		ReadContext:   resourceMemberRead,
		DeleteContext: resourceMemberDelete,
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
			"administrator_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"disable_email_notification": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"disabled_reason": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"email_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"graph_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"invited_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"message": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_time": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"volume_usage_in_bytes": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceMemberCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	accountId := d.Get("account_id").(string)
	graphArn := d.Get("graph_arn").(string)

	accountInput := &detective.Account{
		AccountId:    aws.String(accountId),
		EmailAddress: aws.String(d.Get("email_address").(string)),
	}

	input := &detective.CreateMembersInput{
		Accounts: []*detective.Account{accountInput},
		GraphArn: aws.String(graphArn),
	}

	if v := d.Get("disable_email_notification").(bool); v {
		input.DisableEmailNotification = aws.Bool(v)
	}

	if v, ok := d.GetOk("message"); ok {
		input.Message = aws.String(v.(string))
	}

	var err error
	err = resource.RetryContext(ctx, d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		_, err := conn.CreateMembersWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, detective.ErrCodeInternalServerException) {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.CreateMembersWithContext(ctx, input)
	}

	if err != nil {
		return diag.Errorf("error creating Detective Member: %s", err)
	}

	if _, err = MemberStatusUpdated(ctx, conn, graphArn, accountId, detective.MemberStatusInvited); err != nil {
		return diag.Errorf("error waiting for Detective Member (%s) to be invited: %s", d.Id(), err)
	}

	d.SetId(EncodeMemberID(graphArn, accountId))

	return resourceMemberRead(ctx, d, meta)
}

func resourceMemberRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	graphArn, accountId, err := DecodeMemberID(d.Id())
	if err != nil {
		return diag.Errorf("error decoding ID Detective Member (%s): %s", d.Id(), err)
	}

	resp, err := FindMemberByGraphARNAndAccountID(ctx, conn, graphArn, accountId)

	if err != nil {
		if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) ||
			tfresource.NotFound(err) {
			log.Printf("[WARN] Detective Member (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("error reading Detective Member (%s): %s", d.Id(), err)
	}

	if !d.IsNewResource() && resp == nil {
		log.Printf("[WARN] Detective Member (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("account_id", resp.AccountId)
	d.Set("administrator_id", resp.AdministratorId)
	d.Set("disabled_reason", resp.DisabledReason)
	d.Set("email_address", resp.EmailAddress)
	d.Set("graph_arn", resp.GraphArn)
	d.Set("invited_time", aws.TimeValue(resp.InvitedTime).Format(time.RFC3339))
	d.Set("status", resp.Status)
	d.Set("updated_time", aws.TimeValue(resp.UpdatedTime).Format(time.RFC3339))
	d.Set("volume_usage_in_bytes", resp.VolumeUsageInBytes)

	return nil
}
func resourceMemberDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).DetectiveConn

	graphArn, accountId, err := DecodeMemberID(d.Id())
	if err != nil {
		return diag.Errorf("error decoding ID Detective Member (%s): %s", d.Id(), err)
	}

	input := &detective.DeleteMembersInput{
		AccountIds: []*string{aws.String(accountId)},
		GraphArn:   aws.String(graphArn),
	}

	_, err = conn.DeleteMembersWithContext(ctx, input)
	if err != nil {
		if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.Errorf("error deleting Detective Member (%s): %s", d.Id(), err)
	}
	return nil
}

func EncodeMemberID(graphArn, accountId string) string {
	return fmt.Sprintf("%s/%s", graphArn, accountId)
}

func DecodeMemberID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return "", "", fmt.Errorf("expected ID in the form of graph_arn/account_id, given: %q", id)
	}
	return idParts[0], idParts[1], nil
}
