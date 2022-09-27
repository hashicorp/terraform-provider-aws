package kms

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func ResourceCustomKeyStore() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceCustomKeyStoreCreate,
		ReadWithoutTimeout:   resourceCustomKeyStoreRead,
		UpdateWithoutTimeout: resourceCustomKeyStoreUpdate,
		DeleteWithoutTimeout: resourceCustomKeyStoreDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"cloud_hsm_cluster_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"custom_key_store_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_store_password": {
				Type:     schema.TypeString,
				Required: true,
			},
			"trust_anchor_certificate": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

const (
	ResNameCustomKeyStore = "Custom Key Store"
)

func resourceCustomKeyStoreCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KMSConn

	in := &kms.CreateCustomKeyStoreInput{
		CloudHsmClusterId:      aws.String(d.Get("cloud_hsm_cluster_id").(string)),
		CustomKeyStoreName:     aws.String(d.Get("custom_key_store_name").(string)),
		KeyStorePassword:       aws.String(d.Get("key_store_password").(string)),
		TrustAnchorCertificate: aws.String(d.Get("trust_anchor_certificate").(string)),
	}

	out, err := conn.CreateCustomKeyStoreWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.KMS, create.ErrActionCreating, ResNameCustomKeyStore, d.Get("custom_key_store_name").(string), err)
	}

	if out == nil {
		return create.DiagError(names.KMS, create.ErrActionCreating, ResNameCustomKeyStore, d.Get("custom_key_store_name").(string), errors.New("empty output"))
	}

	d.SetId(aws.StringValue(out.CustomKeyStoreId))

	//if _, err := waitCustomKeyStoreCreated(ctx, conn, d.Id(), d.Timeout(schema.TimeoutCreate)); err != nil {
	//	return create.DiagError(names.KMS, create.ErrActionWaitingForCreation, ResNameCustomKeyStore, d.Id(), err)
	//}

	return resourceCustomKeyStoreRead(ctx, d, meta)
}

func resourceCustomKeyStoreRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KMSConn

	out, err := FindCustomKeyStoreByID(ctx, conn, d.Id())

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] KMS CustomKeyStore (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return create.DiagError(names.KMS, create.ErrActionReading, ResNameCustomKeyStore, d.Id(), err)
	}

	d.Set("cloud_hms_cluster_id", out.CloudHsmClusterId)
	d.Set("custom_key_store_name", out.CustomKeyStoreName)
	d.Set("trust_anchor_certificate", out.TrustAnchorCertificate)

	return nil
}

func resourceCustomKeyStoreUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KMSConn

	update := false

	in := &kms.UpdateCustomKeyStoreInput{
		CustomKeyStoreId:  aws.String(d.Id()),
		CloudHsmClusterId: aws.String(d.Get("cloud_hsm_cluster_id").(string)),
	}

	if d.HasChange("key_store_password") {
		in.KeyStorePassword = aws.String(d.Get("key_store_password").(string))
		update = true
	}

	if d.HasChange("custom_key_store_name") {
		in.NewCustomKeyStoreName = aws.String(d.Get("custom_key_store_name").(string))
		update = true
	}

	if !update {
		return nil
	}

	log.Printf("[DEBUG] Updating KMS CustomKeyStore (%s): %#v", d.Id(), in)
	_, err := conn.UpdateCustomKeyStoreWithContext(ctx, in)
	if err != nil {
		return create.DiagError(names.KMS, create.ErrActionUpdating, ResNameCustomKeyStore, d.Id(), err)
	}

	//if _, err := waitCustomKeyStoreUpdated(ctx, conn, aws.ToString(out.OperationId), d.Timeout(schema.TimeoutUpdate)); err != nil {
	//	return create.DiagError(names.KMS, create.ErrActionWaitingForUpdate, ResNameCustomKeyStore, d.Id(), err)
	//}

	return resourceCustomKeyStoreRead(ctx, d, meta)
}

func resourceCustomKeyStoreDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).KMSConn

	log.Printf("[INFO] Deleting KMS CustomKeyStore %s", d.Id())

	_, err := conn.DeleteCustomKeyStoreWithContext(ctx, &kms.DeleteCustomKeyStoreInput{
		CustomKeyStoreId: aws.String(d.Id()),
	})

	if tfawserr.ErrCodeEquals(err, kms.ErrCodeNotFoundException) {
		return nil
	}

	//if _, err := waitCustomKeyStoreDeleted(ctx, conn, d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
	//	return create.DiagError(names.KMS, create.ErrActionWaitingForDeletion, ResNameCustomKeyStore, d.Id(), err)
	//}

	return nil
}

//const (
//	statusChangePending = "Pending"
//	statusDeleting      = "Deleting"
//	statusNormal        = "Normal"
//	statusUpdated       = "Updated"
//)
//
//func waitCustomKeyStoreCreated(ctx context.Context, conn *kms.Client, id string, timeout time.Duration) (*kms.CustomKeyStore, error) {
//	stateConf := &resource.StateChangeConf{
//		Pending:                   []string{},
//		Target:                    []string{statusNormal},
//		Refresh:                   statusCustomKeyStore(ctx, conn, id),
//		Timeout:                   timeout,
//		NotFoundChecks:            20,
//		ContinuousTargetOccurence: 2,
//	}
//
//	outputRaw, err := stateConf.WaitForStateContext(ctx)
//	if out, ok := outputRaw.(*kms.CustomKeyStore); ok {
//		return out, err
//	}
//
//	return nil, err
//}
//
//func waitCustomKeyStoreUpdated(ctx context.Context, conn *kms.Client, id string, timeout time.Duration) (*kms.CustomKeyStore, error) {
//	stateConf := &resource.StateChangeConf{
//		Pending:                   []string{statusChangePending},
//		Target:                    []string{statusUpdated},
//		Refresh:                   statusCustomKeyStore(ctx, conn, id),
//		Timeout:                   timeout,
//		NotFoundChecks:            20,
//		ContinuousTargetOccurence: 2,
//	}
//
//	outputRaw, err := stateConf.WaitForStateContext(ctx)
//	if out, ok := outputRaw.(*kms.CustomKeyStore); ok {
//		return out, err
//	}
//
//	return nil, err
//}
//
//func waitCustomKeyStoreDeleted(ctx context.Context, conn *kms.Client, id string, timeout time.Duration) (*kms.CustomKeyStore, error) {
//	stateConf := &resource.StateChangeConf{
//		Pending: []string{statusDeleting, statusNormal},
//		Target:  []string{},
//		Refresh: statusCustomKeyStore(ctx, conn, id),
//		Timeout: timeout,
//	}
//
//	outputRaw, err := stateConf.WaitForStateContext(ctx)
//	if out, ok := outputRaw.(*kms.CustomKeyStore); ok {
//		return out, err
//	}
//
//	return nil, err
//}
//
//func statusCustomKeyStore(ctx context.Context, conn *kms.Client, id string) resource.StateRefreshFunc {
//	return func() (interface{}, string, error) {
//		out, err := findCustomKeyStoreByID(ctx, conn, id)
//		if tfresource.NotFound(err) {
//			return nil, "", nil
//		}
//
//		if err != nil {
//			return nil, "", err
//		}
//
//		return out, aws.ToString(out.Status), nil
//	}
//}
