package macie

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/macie"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// @SDKResource("aws_macie_s3_bucket_association")
func ResourceS3BucketAssociation() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceS3BucketAssociationCreate,
		ReadWithoutTimeout:   resourceS3BucketAssociationRead,
		UpdateWithoutTimeout: resourceS3BucketAssociationUpdate,
		DeleteWithoutTimeout: resourceS3BucketAssociationDelete,

		Schema: map[string]*schema.Schema{
			"bucket_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"classification_type": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"continuous": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      macie.S3ContinuousClassificationTypeFull,
							ValidateFunc: validation.StringInSlice([]string{macie.S3ContinuousClassificationTypeFull}, false),
						},
						"one_time": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      macie.S3OneTimeClassificationTypeNone,
							ValidateFunc: validation.StringInSlice([]string{macie.S3OneTimeClassificationTypeFull, macie.S3OneTimeClassificationTypeNone}, false),
						},
					},
				},
			},
			"member_account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"prefix": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceS3BucketAssociationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	input := &macie.AssociateS3ResourcesInput{
		S3Resources: []*macie.S3ResourceClassification{
			{
				BucketName:         aws.String(d.Get("bucket_name").(string)),
				ClassificationType: expandClassificationType(d),
			},
		},
	}

	if v, ok := d.GetOk("member_account_id"); ok {
		input.MemberAccountId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("prefix"); ok {
		input.S3Resources[0].Prefix = aws.String(v.(string))
	}

	output, err := conn.AssociateS3ResourcesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Macie Classic S3 Bucket Association: %s", err)
	}

	if len(output.FailedS3Resources) > 0 {
		return sdkdiag.AppendErrorf(diags, "creating Macie Classic S3 Bucket Association: %s", output.FailedS3Resources[0])
	}

	d.SetId(fmt.Sprintf("%s/%s", d.Get("bucket_name"), d.Get("prefix")))

	return append(diags, resourceS3BucketAssociationRead(ctx, d, meta)...)
}

func resourceS3BucketAssociationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	output, err := FindS3ResourceClassificationByThreePartKey(ctx, conn, d.Get("member_account_id").(string), d.Get("bucket_name").(string), d.Get("prefix").(string))

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] Macie Classic S3 Bucket Association (%s) not found, removing from state", d.Id())
		d.SetId("")
		return diags
	}

	if err != nil {
		return diag.Errorf("reading Macie Classic S3 Bucket Association (%s): %s", d.Id(), err)
	}

	d.Set("bucket_name", output.BucketName)
	if err := d.Set("classification_type", flattenClassificationType(output.ClassificationType)); err != nil {
		return sdkdiag.AppendErrorf(diags, "setting classification_type: %s", err)
	}
	d.Set("prefix", output.Prefix)

	return diags
}

func resourceS3BucketAssociationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	if d.HasChange("classification_type") {
		input := &macie.UpdateS3ResourcesInput{
			S3ResourcesUpdate: []*macie.S3ResourceClassificationUpdate{
				{
					BucketName:               aws.String(d.Get("bucket_name").(string)),
					ClassificationTypeUpdate: expandClassificationTypeUpdate(d),
				},
			},
		}

		if v, ok := d.GetOk("member_account_id"); ok {
			input.MemberAccountId = aws.String(v.(string))
		}

		if v, ok := d.GetOk("prefix"); ok {
			input.S3ResourcesUpdate[0].Prefix = aws.String(v.(string))
		}

		output, err := conn.UpdateS3ResourcesWithContext(ctx, input)

		if err != nil {
			return sdkdiag.AppendErrorf(diags, "updating Macie Classic S3 Bucket Association (%s): %s", d.Id(), err)
		}

		if len(output.FailedS3Resources) > 0 {
			return sdkdiag.AppendErrorf(diags, "creating Macie Classic S3 Bucket Association (%s): %s", d.Id(), output.FailedS3Resources[0])
		}
	}

	return append(diags, resourceS3BucketAssociationRead(ctx, d, meta)...)
}

func resourceS3BucketAssociationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).MacieConn()

	input := &macie.DisassociateS3ResourcesInput{
		AssociatedS3Resources: []*macie.S3Resource{
			{
				BucketName: aws.String(d.Get("bucket_name").(string)),
			},
		},
	}

	if v, ok := d.GetOk("member_account_id"); ok {
		input.MemberAccountId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("prefix"); ok {
		input.AssociatedS3Resources[0].Prefix = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Deleting Macie Classic S3 Bucket Association: %s", d.Id())
	output, err := conn.DisassociateS3ResourcesWithContext(ctx, input)

	if err != nil {
		return sdkdiag.AppendErrorf(diags, "deleting Macie Classic S3 Bucket Association (%s): %s", d.Id(), err)
	}

	if len(output.FailedS3Resources) > 0 {
		failed := output.FailedS3Resources[0]
		// {
		// 	ErrorCode: "InvalidInputException",
		// 	ErrorMessage: "The request was rejected. The specified S3 resource (bucket or prefix) is not associated with Macie.",
		// 	FailedItem: {
		// 	 BucketName: "tf-macie-example-002"
		// 	}
		// }
		if aws.StringValue(failed.ErrorCode) == macie.ErrCodeInvalidInputException &&
			strings.Contains(aws.StringValue(failed.ErrorMessage), "is not associated with Macie") {
			return diags
		}

		return sdkdiag.AppendErrorf(diags, "deleting Macie Classic S3 Bucket Association (%s): %s", d.Id(), failed)
	}

	return diags
}

func FindS3ResourceClassificationByThreePartKey(ctx context.Context, conn *macie.Macie, memberAccountID, bucketName, prefix string) (*macie.S3ResourceClassification, error) {
	input := &macie.ListS3ResourcesInput{}
	if memberAccountID != "" {
		input.MemberAccountId = aws.String(memberAccountID)
	}
	var output *macie.S3ResourceClassification

	err := conn.ListS3ResourcesPagesWithContext(ctx, input, func(page *macie.ListS3ResourcesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.S3Resources {
			if v != nil && aws.StringValue(v.BucketName) == bucketName && aws.StringValue(v.Prefix) == prefix {
				output = v
				return false
			}
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, &retry.NotFoundError{}
	}

	return output, nil
}
