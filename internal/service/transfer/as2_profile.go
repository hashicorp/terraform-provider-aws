package transfer

import ( // nosemgrep:ci.aws-sdk-go-multiple-service-imports
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/flex"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// @SDKResource("aws_transfer_as2_profile", name="Profile")
// @Tags(identifierAttribute="profile_id")
func ResourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceProfileCreate,
		ReadWithoutTimeout:   resourceProfileRead,
		UpdateWithoutTimeout: resourceProfileUpdate,
		DeleteWithoutTimeout: resourceProfileDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"as2_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"certificate_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"profile_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"profile_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(transfer.ProfileType_Values(), false),
			},
			names.AttrTags:    tftags.TagsSchema(),
			names.AttrTagsAll: tftags.TagsSchemaComputed(),
		},
		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn()

	input := &transfer.CreateProfileInput{
		Tags: GetTagsIn(ctx),
	}

	if v, ok := d.GetOk("as2_id"); ok {
		input.As2Id = aws.String(v.(string))
	}

	if v, ok := d.GetOk("certificate_ids"); ok && v.(*schema.Set).Len() > 0 {
		input.CertificateIds = flex.ExpandStringSet(v.(*schema.Set))
	}

	if v, ok := d.GetOk("profile_type"); ok {
		input.ProfileType = aws.String(v.(string))
	}

	output, err := conn.CreateProfileWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Transfer AS2 Profile: %s", err)
	}

	d.SetId(aws.StringValue(output.ProfileId))

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn()

	output, err := FindProfileByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] AS2 Profile (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading AS2 Profile (%s): %s", d.Id(), err)
	}

	d.Set("as2_id", output.As2Id)
	if output.CertificateIds != nil {
		d.Set("certificate_ids", output.CertificateIds)
	}

	d.Set("profile_id", output.ProfileId)
	d.Set("profile_type", output.ProfileType)
	SetTagsOut(ctx, output.Tags)

	return diags
}

func resourceProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn()

	if d.HasChangesExcept("tags", "tags_all") {

		input := &transfer.UpdateProfileInput{
			ProfileId: aws.String(d.Id()),
		}

		if d.HasChange("certificate_ids") {
			input.CertificateIds = flex.ExpandStringSet(d.Get("certificate_ids").(*schema.Set))
		}

		if _, err := conn.UpdateProfileWithContext(ctx, input); err != nil {
			return sdkdiag.AppendErrorf(diags, "removing AS2 Profile IDs: %s", err)
		}

		if _, err := conn.UpdateProfileWithContext(ctx, input); err != nil {
			return sdkdiag.AppendFromErr(diags, err)
		}
	}

	return append(diags, resourceProfileRead(ctx, d, meta)...)
}

func resourceProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).TransferConn()

	log.Printf("[DEBUG] Deleting AS2 Profile: (%s)", d.Id())
	_, err := conn.DeleteProfileWithContext(ctx, &transfer.DeleteProfileInput{
		ProfileId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, transfer.ErrCodeResourceNotFoundException) {
		return diags
	}

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting AS2 Profile (%s): %s", d.Id(), err)
	}

	return diags
}
