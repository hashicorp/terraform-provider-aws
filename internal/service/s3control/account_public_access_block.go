package s3control

import (
	"context"
	"log"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func init() {
	_sp.registerSDKResourceFactory("aws_s3_account_public_access_block", resourceAccountPublicAccessBlock)
}

func resourceAccountPublicAccessBlock() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceAccountPublicAccessBlockCreate,
		ReadWithoutTimeout:   resourceAccountPublicAccessBlockRead,
		UpdateWithoutTimeout: resourceAccountPublicAccessBlockUpdate,
		DeleteWithoutTimeout: resourceAccountPublicAccessBlockDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"account_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidAccountID,
			},
			"block_public_acls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"block_public_policy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ignore_public_acls": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"restrict_public_buckets": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceAccountPublicAccessBlockCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	accountID := meta.(*conns.AWSClient).AccountID
	if v, ok := d.GetOk("account_id"); ok {
		accountID = v.(string)
	}

	input := &s3control.PutPublicAccessBlockInput{
		AccountId: aws.String(accountID),
		PublicAccessBlockConfiguration: &s3control.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
			BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
			IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
			RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
		},
	}

	_, err := conn.PutPublicAccessBlockWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("creating S3 Account Public Access Block (%s): %s", accountID, err)
	}

	d.SetId(accountID)

	_, err = tfresource.RetryWhenNotFound(ctx, propagationTimeout, func() (interface{}, error) {
		return FindPublicAccessBlockByAccountID(ctx, conn, d.Id())
	})

	if err != nil {
		return diag.Errorf("waiting for S3 Account Public Access Block (%s) create: %s", d.Id(), err)
	}

	return resourceAccountPublicAccessBlockRead(ctx, d, meta)
}

func resourceAccountPublicAccessBlockRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	output, err := FindPublicAccessBlockByAccountID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Account Public Access Block (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return diag.Errorf("reading S3 Account Public Access Block (%s): %s", d.Id(), err)
	}

	d.Set("account_id", d.Id())
	d.Set("block_public_acls", output.BlockPublicAcls)
	d.Set("block_public_policy", output.BlockPublicPolicy)
	d.Set("ignore_public_acls", output.IgnorePublicAcls)
	d.Set("restrict_public_buckets", output.RestrictPublicBuckets)

	return nil
}

func resourceAccountPublicAccessBlockUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	publicAccessBlockConfiguration := &s3control.PublicAccessBlockConfiguration{
		BlockPublicAcls:       aws.Bool(d.Get("block_public_acls").(bool)),
		BlockPublicPolicy:     aws.Bool(d.Get("block_public_policy").(bool)),
		IgnorePublicAcls:      aws.Bool(d.Get("ignore_public_acls").(bool)),
		RestrictPublicBuckets: aws.Bool(d.Get("restrict_public_buckets").(bool)),
	}
	input := &s3control.PutPublicAccessBlockInput{
		AccountId:                      aws.String(d.Id()),
		PublicAccessBlockConfiguration: publicAccessBlockConfiguration,
	}

	_, err := conn.PutPublicAccessBlockWithContext(ctx, input)

	if err != nil {
		return diag.Errorf("updating S3 Account Public Access Block (%s): %s", d.Id(), err)
	}

	if _, err := waitPublicAccessBlockEqual(ctx, conn, d.Id(), publicAccessBlockConfiguration); err != nil {
		return diag.Errorf("waiting for S3 Account Public Access Block (%s) update: %s", d.Id(), err)
	}

	return resourceAccountPublicAccessBlockRead(ctx, d, meta)
}

func resourceAccountPublicAccessBlockDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).S3ControlConn()

	log.Printf("[DEBUG] Deleting S3 Account Public Access Block: %s", d.Id())
	_, err := conn.DeletePublicAccessBlockWithContext(ctx, &s3control.DeletePublicAccessBlockInput{
		AccountId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration) {
		return nil
	}

	if err != nil {
		return diag.Errorf("deleting S3 Account Public Access Block (%s): %s", d.Id(), err)
	}

	return nil
}

func FindPublicAccessBlockByAccountID(ctx context.Context, conn *s3control.S3Control, accountID string) (*s3control.PublicAccessBlockConfiguration, error) {
	input := &s3control.GetPublicAccessBlockInput{
		AccountId: aws.String(accountID),
	}

	output, err := conn.GetPublicAccessBlockWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, s3control.ErrCodeNoSuchPublicAccessBlockConfiguration) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.PublicAccessBlockConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.PublicAccessBlockConfiguration, nil
}

func statusPublicAccessBlockEqual(ctx context.Context, conn *s3control.S3Control, accountID string, target *s3control.PublicAccessBlockConfiguration) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindPublicAccessBlockByAccountID(ctx, conn, accountID)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, strconv.FormatBool(reflect.DeepEqual(output, target)), nil
	}
}

func waitPublicAccessBlockEqual(ctx context.Context, conn *s3control.S3Control, accountID string, target *s3control.PublicAccessBlockConfiguration) (*s3control.PublicAccessBlockConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{strconv.FormatBool(false)},
		Target:                    []string{strconv.FormatBool(true)},
		Refresh:                   statusPublicAccessBlockEqual(ctx, conn, accountID, target),
		Timeout:                   propagationTimeout,
		MinTimeout:                propagationMinTimeout,
		ContinuousTargetOccurence: propagationContinuousTargetOccurence,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*s3control.PublicAccessBlockConfiguration); ok {
		return output, err
	}

	return nil, err
}
