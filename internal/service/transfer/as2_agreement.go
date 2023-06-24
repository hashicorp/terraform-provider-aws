package transfer

import ( // nosemgrep:ci.aws-sdk-go-multiple-service-imports
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_as2_agreement", name="Agreement")
// @Tags(identifierAttribute="agreement_id")
func ResourceAgreement() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAgreementCreate,
		ReadWithoutTimeout:   resourceAgreementRead,
		UpdateWithoutTimeout: resourceAgreementUpdate,
		DeleteWithoutTimeout: resourceAgreementDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"access_role": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: verify.ValidARN,
			},
			"agreement_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"base_directory": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"local_profile_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"partner_profile_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"serverid": {
				Type:     schema.TypeString,
				Required: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				//ValidateFunc: validation.StringInSlice(transfer.AgreementStatusType_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceAgreementCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	input := &transfer.CreateAgreementInput{
		Tags: getTagsIn(ctx),
	}

	if v, ok := d.GetOk("access_role"); ok {
		input.AccessRole = aws.String(v.(string))
	}

	if v, ok := d.GetOk("base_directory"); ok {
		input.BaseDirectory = aws.String(v.(string))
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("local_profile_id"); ok {
		input.LocalProfileId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("partner_profile_id"); ok {
		input.PartnerProfileId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("serverid"); ok {
		input.ServerId = aws.String(v.(string))
	}

	output, err := conn.CreateAgreementWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer Agreement: %s", err)
	}

	agreementID := aws.StringValue(output.AgreementId)
	serverID := d.Get("serverid").(string)
	id := AccessCreateResourceID(agreementID, serverID)
	d.SetId(id)
	//d.SetId(aws.StringValue(output.AgreementId))

	return append(diags, resourceAgreementRead(ctx, d, meta)...)
}

func resourceAgreementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)
	agreementID, serverID, err := AccessParseResourceID(d.Id())

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Transfer Agreement ID: %s", err)
	}

	output, err := FindAgreementByID(ctx, conn, agreementID, serverID)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AS2 Agreement (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AS2 Agreement (%s): %s", d.Id(), err)
	}

	d.Set("access_role", output.AccessRole)
	d.Set("agreement_id", output.AgreementId)
	d.Set("base_directory", output.BaseDirectory)
	d.Set("description", output.Description)
	d.Set("local_profile_id", output.LocalProfileId)
	d.Set("partner_profile_id", output.PartnerProfileId)
	d.Set("serverid", output.ServerId)
	d.Set("status", output.Status)

	setTagsOut(ctx, output.Tags)

	return diags
}

func resourceAgreementUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)

	agreementID, serverID, err := AccessParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Transfer Agreement ID: %s", err)
	}

	if d.HasChangesExcept("tags", "tags_all") {

		input := &transfer.UpdateAgreementInput{
			AgreementId: aws.String(agreementID),
			ServerId:    aws.String(serverID),
		}

		if d.HasChange("access_role") {
			input.AccessRole = aws.String(d.Get("access_role").(string))
		}

		if d.HasChange("base_directory") {
			input.BaseDirectory = aws.String(d.Get("base_directory").(string))
		}

		if d.HasChange("description") {
			input.Description = aws.String(d.Get("description").(string))
		}

		if d.HasChange("local_profile_id") {
			input.LocalProfileId = aws.String(d.Get("local_profile_id").(string))
		}

		if d.HasChange("partner_profile_id") {
			input.PartnerProfileId = aws.String(d.Get("partner_profile_id").(string))
		}

		if _, err := conn.UpdateAgreementWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing AS2 Agreement IDs: %s", err)
		}

		if _, err := conn.UpdateAgreementWithContext(ctx, input); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceAgreementRead(ctx, d, meta)...)
}

func resourceAgreementDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn(ctx)
	agreementID, serverID, err := AccessParseResourceID(d.Id())
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "parsing Transfer Agreement ID: %s", err)
	}

	log.Printf("[DEBUG] Deleting AS2 Agreement: (%s)", d.Id())
	_, err = conn.DeleteAgreementWithContext(ctx, &transfer.DeleteAgreementInput{
		AgreementId: aws.String(agreementID),
		ServerId:    aws.String(serverID),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AS2 Agreement (%s): %s", d.Id(), err)
	}

	return diags
}
