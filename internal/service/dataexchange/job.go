package dataexchange

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"regexp"

	"github.com/aws/aws-sdk-go-v2/service/dataexchange"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dataexchange/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/sdkdiag"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Attributes
const (
	attrDataSetId                             = "data_set_id"
	attrStartOnCreation                       = "start_on_creation"
	attrRevisionId                            = "revision_id"
	attrAssetId                               = "asset_id"
	attrS3AccessAssetSourceBucket             = "s3_access_asset_source_bucket"
	attrS3AccessAssetSourceKeyPrefixes        = "s3_access_asset_source_key_prefixes"
	attrS3AccessAssetSourceKeys               = "s3_access_asset_source_keys"
	attrS3AccessAssetSourceKmsKeysToGrant     = "s3_access_asset_source_kms_keys_to_grant"
	attrS3AssetDestinations                   = "s3_asset_destinations"
	attrS3AssetDestinationEncryptionKMSKeyArn = "s3_asset_destination_encryption_kms_key_arn"
	attrS3AssetDestinationEncryptionType      = "s3_asset_destination_encryption_type"
	attrSignedUrl                             = "signed_url"
	attrSignedUrlExpiresAt                    = "signed_url_expires_at"
	attrS3RevisionDestinations                = "s3_revision_destinations"
	attrS3EventActionArn                      = "s3_event_action_arn"
	attrApiGtwDescription                     = "api_gtw_description"
	attrApiGtwId                              = "api_gtw_id"
	attrApiGtwKey                             = "api_gtw_key"
	attrApiGtwName                            = "api_gtw_name"
	attrApiGtwSpecMd5Hash                     = "api_gtw_spec_md5_hash"
	attrApiGtwProtoType                       = "api_gtw_proto_type"
	attrApiGtwStage                           = "api_gtw_stage"
	attrApiSpecUploadUrl                      = "api_specification_upload_url"
	attrApiSPecUploadUrlExpAt                 = "api_specification_upload_url_expires_at"
	attrUrlAssetName                          = "url_asset_name"
	attrUrlMd5Hash                            = "url_md5_hash"
	attrRedshiftAssetSources                  = "redshift_asset_sources"
	attrS3AssetSources                        = "s3_asset_sources"
	attrLakeFormationCatalogId                = "lake_formation_catalog_id"
	attrLakeFormationDatabaseExpression       = "lake_formation_database_expression"
	attrLakeFormationDatabasePermissions      = "lake_formation_database_permissions"
	attrLakeFormationRoleArn                  = "lake_formation_role_arn"
	attrLakeFormationTableExpression          = "lake_formation_table_expression"
	attrLakeFormationTablePermissions         = "lake_formation_table_permissions"
	attrBucket                                = "bucket"
	attrKey                                   = "key"
	attrKeyPattern                            = "key_pattern"
	attrTagKey                                = "tag_key"
	attrTagValue                              = "tag_value"
)

// @SDKResource("aws_dataexchange_job", name="Job")
func ResourceJob() *schema.Resource {
	return &schema.Resource{
		CreateWithoutTimeout: resourceJobCreate,
		ReadWithoutTimeout:   resourceJobRead,
		UpdateWithoutTimeout: resourceJobUpdate,
		DeleteWithoutTimeout: resourceJobDelete,

		SchemaFunc: resourceJobSchema,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// CRUD
func resourceJobCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	input := buildCreateJobInput(d)
	output, err := conn.CreateJob(ctx, input)
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "creating Dataexchange Job: %s", err)
	}

	if output.Errors != nil {
		for _, jobError := range output.Errors {
			diags = sdkdiag.AppendErrorf(diags, "creating Dataexchange Job: %s", *jobError.Message)
		}
	}

	if startOnCreation := d.Get(attrStartOnCreation).(bool); startOnCreation {
		_, err := conn.StartJob(ctx, &dataexchange.StartJobInput{
			JobId: output.Id,
		})

		if err != nil {
			// to confirm
			resourceJobDelete(ctx, d, meta)
			return sdkdiag.AppendErrorf(diags, "creating Dataexchange Job: %s", err)
		}
	}

	d.SetId(*output.Id)
	return append(diags, resourceJobRead(ctx, d, meta)...)
}

func resourceJobRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)

	jobId := d.Get(names.AttrID).(string)
	output, err := conn.GetJob(ctx, &dataexchange.GetJobInput{
		JobId: aws.String(jobId),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "reading Dataexchange Job: %s", err)
	}
	if output.Errors != nil {
		for _, jobError := range output.Errors {
			diags = sdkdiag.AppendErrorf(diags, "reading Dataexchange Job: %s", *jobError.Message)
		}
	}

	buildResourceData(output, d)
	return diags
}

func resourceJobUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange(attrStartOnCreation) && !d.HasChangeExcept(attrStartOnCreation) {
		if startOnCreation := d.Get(attrStartOnCreation).(bool); startOnCreation {
			var diags diag.Diagnostics
			conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)
			_, err := conn.StartJob(ctx, &dataexchange.StartJobInput{
				JobId: aws.String(d.Get(names.AttrID).(string)),
			})

			if err != nil {
				return sdkdiag.AppendErrorf(diags, "starting Dataexchange Job: %s", err)
			}

			return append(diags, resourceJobRead(ctx, d, meta)...)
		}
	}
	resourceJobDelete(ctx, d, meta)
	return resourceJobCreate(ctx, d, meta)
}

func resourceJobDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if d.Get(names.AttrState).(string) != string(awstypes.StateWaiting) {
		return diags
	}
	conn := meta.(*conns.AWSClient).DataExchangeClient(ctx)
	jobId := d.Get(names.AttrID).(string)
	_, err := conn.CancelJob(ctx, &dataexchange.CancelJobInput{
		JobId: aws.String(jobId),
	})
	if err != nil {
		return sdkdiag.AppendErrorf(diags, "canceling Dataexchange Job: %s", err)
	}

	return diags
}

// Builders
func buildResourceData(out *dataexchange.GetJobOutput, d *schema.ResourceData) {
	d.SetId(*out.Id)
	d.Set(names.AttrARN, out.Arn)
	d.Set(names.AttrType, out.Type)
	d.Set(names.AttrCreatedAt, out.CreatedAt.String())
	d.Set(names.AttrLastUpdatedTime, out.UpdatedAt.String())
	d.Set(names.AttrState, out.State)

	if out.Details != nil {
		switch out.Type {
		case awstypes.TypeExportAssetsToS3:
			if out.Details.ExportAssetsToS3 != nil {
				d.Set(attrDataSetId, out.Details.ExportAssetsToS3.DataSetId)
				if out.Details.ExportAssetsToS3.Encryption != nil {
					d.Set(attrS3AssetDestinationEncryptionType, out.Details.ExportAssetsToS3.Encryption.Type)
					d.Set(attrS3AssetDestinationEncryptionKMSKeyArn, out.Details.ExportAssetsToS3.Encryption.KmsKeyArn)
				}
				d.Set(attrS3AssetDestinations, buildS3AssetsDestinationsAttr(out.Details.ExportAssetsToS3.AssetDestinations))
			}
		case awstypes.TypeExportAssetToSignedUrl:
			if out.Details.ExportAssetToSignedUrl != nil {
				d.Set(attrAssetId, out.Details.ExportAssetToSignedUrl.AssetId)
				d.Set(attrSignedUrl, out.Details.ExportAssetToSignedUrl.SignedUrl)
				d.Set(attrSignedUrlExpiresAt, out.Details.ExportAssetToSignedUrl.SignedUrlExpiresAt)
				d.Set(attrDataSetId, out.Details.ExportAssetToSignedUrl.DataSetId)
				d.Set(attrRevisionId, out.Details.ExportAssetToSignedUrl.RevisionId)
			}
			break
		case awstypes.TypeExportRevisionsToS3:
			if out.Details.ExportRevisionsToS3 != nil {
				if out.Details.ExportAssetsToS3.Encryption != nil {
					d.Set(attrS3AssetDestinationEncryptionType, out.Details.ExportRevisionsToS3.Encryption.Type)
					d.Set(attrS3AssetDestinationEncryptionKMSKeyArn, out.Details.ExportRevisionsToS3.Encryption.KmsKeyArn)
				}
				d.Set(attrDataSetId, out.Details.ExportRevisionsToS3.DataSetId)
				d.Set(attrS3RevisionDestinations, buildS3RevisionDestinationsAttr(out.Details.ExportRevisionsToS3.RevisionDestinations))
				d.Set(attrS3EventActionArn, out.Details.ExportRevisionsToS3.EventActionArn)
			}
			break
		case awstypes.TypeCreateS3DataAccessFromS3Bucket:
			if out.Details.CreateS3DataAccessFromS3Bucket != nil {
				d.Set(attrRevisionId, out.Details.CreateS3DataAccessFromS3Bucket.DataSetId)
				d.Set(attrDataSetId, out.Details.CreateS3DataAccessFromS3Bucket.DataSetId)
				if out.Details.CreateS3DataAccessFromS3Bucket.AssetSource != nil {
					d.Set(attrS3AccessAssetSourceBucket, out.Details.CreateS3DataAccessFromS3Bucket.AssetSource.Bucket)
					d.Set(attrS3AccessAssetSourceKeyPrefixes, out.Details.CreateS3DataAccessFromS3Bucket.AssetSource.KeyPrefixes)
					d.Set(attrS3AccessAssetSourceKeys, out.Details.CreateS3DataAccessFromS3Bucket.AssetSource.Keys)
					d.Set(attrS3AccessAssetSourceKmsKeysToGrant, buildKmsKeysToGrantAttr(out.Details.CreateS3DataAccessFromS3Bucket.AssetSource.KmsKeysToGrant))
				}
			}
			break
		case awstypes.TypeImportAssetsFromLakeFormationTagPolicy:
			if out.Details.ImportAssetsFromLakeFormationTagPolicy != nil {
				d.Set(attrLakeFormationCatalogId, out.Details.ImportAssetsFromLakeFormationTagPolicy.CatalogId)
				d.Set(attrDataSetId, out.Details.ImportAssetsFromLakeFormationTagPolicy.DataSetId)
				d.Set(attrRevisionId, out.Details.ImportAssetsFromLakeFormationTagPolicy.RevisionId)
				d.Set(attrLakeFormationRoleArn, out.Details.ImportAssetsFromLakeFormationTagPolicy.RoleArn)
				if out.Details.ImportAssetsFromLakeFormationTagPolicy.Table != nil {
					d.Set(attrLakeFormationTablePermissions, out.Details.ImportAssetsFromLakeFormationTagPolicy.Table.Permissions)
					d.Set(attrLakeFormationTableExpression, buildLFTagAttr(out.Details.ImportAssetsFromLakeFormationTagPolicy.Table.Expression))
				}
				if out.Details.ImportAssetsFromLakeFormationTagPolicy.Database != nil {
					d.Set(attrLakeFormationDatabasePermissions, out.Details.ImportAssetsFromLakeFormationTagPolicy.Database.Expression)
					d.Set(attrLakeFormationDatabaseExpression, buildLFTagAttr(out.Details.ImportAssetsFromLakeFormationTagPolicy.Database.Expression))
				}
			}
			break
		case awstypes.TypeImportAssetFromSignedUrl:
			if out.Details.ImportAssetFromSignedUrl != nil {
				d.Set(attrUrlAssetName, out.Details.ImportAssetFromSignedUrl.AssetName)
				d.Set(attrSignedUrl, out.Details.ImportAssetFromSignedUrl.SignedUrl)
				d.Set(attrSignedUrlExpiresAt, out.Details.ImportAssetFromSignedUrl.SignedUrlExpiresAt.String())
				d.Set(attrUrlMd5Hash, out.Details.ImportAssetFromSignedUrl.Md5Hash)
				d.Set(attrDataSetId, out.Details.ImportAssetFromSignedUrl.DataSetId)
				d.Set(attrRevisionId, out.Details.ImportAssetFromSignedUrl.RevisionId)
			}
			break
		case awstypes.TypeImportAssetFromApiGatewayApi:
			if out.Details.ImportAssetFromApiGatewayApi != nil {
				d.Set(attrDataSetId, out.Details.ImportAssetFromApiGatewayApi.DataSetId)
				d.Set(attrRevisionId, out.Details.ImportAssetFromApiGatewayApi.RevisionId)
				d.Set(attrApiGtwId, out.Details.ImportAssetFromApiGatewayApi.ApiId)
				d.Set(attrApiGtwName, out.Details.ImportAssetFromApiGatewayApi.ApiName)
				d.Set(attrApiGtwProtoType, out.Details.ImportAssetFromApiGatewayApi.ProtocolType)
				d.Set(attrApiGtwKey, out.Details.ImportAssetFromApiGatewayApi.ApiKey)
				d.Set(attrApiGtwDescription, out.Details.ImportAssetFromApiGatewayApi.ApiDescription)
				d.Set(attrApiGtwStage, out.Details.ImportAssetFromApiGatewayApi.Stage)
				d.Set(attrApiGtwSpecMd5Hash, out.Details.ImportAssetFromApiGatewayApi.ApiSpecificationMd5Hash)
				d.Set(attrApiSpecUploadUrl, out.Details.ImportAssetFromApiGatewayApi.ApiSpecificationUploadUrl)
				d.Set(attrApiSPecUploadUrlExpAt, out.Details.ImportAssetFromApiGatewayApi.ApiSpecificationUploadUrlExpiresAt.String())
			}
			break
		case awstypes.TypeImportAssetsFromRedshiftDataShares:
			if out.Details.ImportAssetsFromRedshiftDataShares != nil {
				d.Set(attrDataSetId, out.Details.ImportAssetsFromRedshiftDataShares.DataSetId)
				d.Set(attrRevisionId, out.Details.ImportAssetsFromRedshiftDataShares.RevisionId)
				d.Set(attrRedshiftAssetSources, buildRedshiftDataSharesAssetsAttr(out.Details.ImportAssetsFromRedshiftDataShares.AssetSources))
			}
			break
		case awstypes.TypeImportAssetsFromS3:
			if out.Details.ImportAssetsFromS3 != nil {
				d.Set(attrDataSetId, out.Details.ImportAssetsFromS3.DataSetId)
				d.Set(attrRevisionId, out.Details.ImportAssetsFromS3.RevisionId)
				d.Set(attrS3AssetSources, buildS3AssetsSourceAttr(out.Details.ImportAssetsFromS3.AssetSources))
			}
			break
		}
	}
}

func buildS3AssetsDestinationsAttr(out []awstypes.AssetDestinationEntry) []map[string]any {
	res := make([]map[string]any, len(out))
	for i, entry := range out {
		res[i] = map[string]any{
			attrBucket:  entry.Bucket,
			attrAssetId: entry.AssetId,
			attrKey:     entry.Key,
		}
	}

	return res
}

func buildS3AssetsSourceAttr(out []awstypes.AssetSourceEntry) []map[string]any {
	res := make([]map[string]any, len(out))
	for i, entry := range out {
		res[i] = map[string]any{
			attrBucket: entry.Bucket,
			attrKey:    entry.Key,
		}
	}

	return res
}

func buildS3RevisionDestinationsAttr(out []awstypes.RevisionDestinationEntry) []map[string]any {
	res := make([]map[string]any, len(out))
	for i, entry := range out {
		res[i] = map[string]any{
			attrBucket:     entry.Bucket,
			attrRevisionId: entry.RevisionId,
			attrKeyPattern: entry.KeyPattern,
		}
	}

	return res
}

func buildKmsKeysToGrantAttr(out []awstypes.KmsKeyToGrant) []string {
	res := make([]string, len(out))
	for i, entry := range out {
		if entry.KmsKeyArn != nil {
			res[i] = *entry.KmsKeyArn
		}
	}

	return res
}

func buildLFTagAttr(out []awstypes.LFTag) []map[string]any {
	res := make([]map[string]any, len(out))
	for i, entry := range out {
		res[i] = map[string]any{
			attrTagKey:   entry.TagKey,
			attrTagValue: entry.TagValues,
		}
	}

	return res
}

func buildRedshiftDataSharesAssetsAttr(out []awstypes.RedshiftDataShareAssetSourceEntry) []string {
	res := make([]string, len(out))
	for i, entry := range out {
		if entry.DataShareArn != nil {
			res[i] = *entry.DataShareArn
		}
	}

	return res
}

func buildCreateJobInput(d *schema.ResourceData) *dataexchange.CreateJobInput {
	in := &dataexchange.CreateJobInput{
		Type:    awstypes.Type(d.Get(names.AttrType).(string)),
		Details: &awstypes.RequestDetails{},
	}

	switch in.Type {
	case awstypes.TypeExportAssetsToS3:
		in.Details.ExportAssetsToS3 = &awstypes.ExportAssetsToS3RequestDetails{
			DataSetId:  aws.String(d.Get(attrDataSetId).(string)),
			RevisionId: aws.String(d.Get(attrRevisionId).(string)),
		}

		if v, ok := d.GetOk(attrS3AssetDestinations); ok {
			dest := make([]awstypes.AssetDestinationEntry, len(v.([]any)))
			for i, values := range v.([]any) {
				valuesM := values.(map[string]any)
				dest[i] = awstypes.AssetDestinationEntry{
					AssetId: aws.String(valuesM[attrAssetId].(string)),
					Bucket:  aws.String(valuesM[attrBucket].(string)),
					Key:     aws.String(valuesM[attrKey].(string)),
				}
			}
			in.Details.ExportAssetsToS3.AssetDestinations = dest
		}

		if _, ok := d.GetOk(attrS3AssetDestinationEncryptionType); ok {
			in.Details.ExportAssetsToS3.Encryption = &awstypes.ExportServerSideEncryption{
				Type:      awstypes.ServerSideEncryptionTypes(d.Get(attrS3AssetDestinationEncryptionType).(string)),
				KmsKeyArn: aws.String(d.Get(attrS3AssetDestinationEncryptionKMSKeyArn).(string)),
			}
		}
		break
	case awstypes.TypeExportAssetToSignedUrl:
		in.Details.ExportAssetToSignedUrl = &awstypes.ExportAssetToSignedUrlRequestDetails{
			AssetId:    aws.String(d.Get(attrAssetId).(string)),
			DataSetId:  aws.String(d.Get(attrDataSetId).(string)),
			RevisionId: aws.String(d.Get(attrRevisionId).(string)),
		}
		break
	case awstypes.TypeExportRevisionsToS3:
		in.Details.ExportRevisionsToS3 = &awstypes.ExportRevisionsToS3RequestDetails{
			DataSetId: aws.String(d.Get(attrDataSetId).(string)),
		}

		if v, ok := d.GetOk(attrS3RevisionDestinations); ok {
			dest := make([]awstypes.RevisionDestinationEntry, len(v.([]any)))
			for i, values := range v.([]interface{}) {
				valuesM := values.(map[string]any)
				dest[i] = awstypes.RevisionDestinationEntry{
					RevisionId: aws.String(valuesM[attrRevisionId].(string)),
					Bucket:     aws.String(valuesM[attrBucket].(string)),
					KeyPattern: aws.String(valuesM[attrKeyPattern].(string)),
				}
			}
			in.Details.ExportRevisionsToS3.RevisionDestinations = dest
		}

		if _, ok := d.GetOk(attrS3AssetDestinationEncryptionType); ok {
			in.Details.ExportRevisionsToS3.Encryption = &awstypes.ExportServerSideEncryption{
				Type:      awstypes.ServerSideEncryptionTypes(d.Get(attrS3AssetDestinationEncryptionType).(string)),
				KmsKeyArn: aws.String(d.Get(attrS3AssetDestinationEncryptionKMSKeyArn).(string)),
			}
		}
		break
	case awstypes.TypeCreateS3DataAccessFromS3Bucket:
		in.Details.CreateS3DataAccessFromS3Bucket = &awstypes.CreateS3DataAccessFromS3BucketRequestDetails{
			DataSetId:  aws.String(d.Get(attrDataSetId).(string)),
			RevisionId: aws.String(d.Get(attrRevisionId).(string)),
			AssetSource: &awstypes.S3DataAccessAssetSourceEntry{
				Bucket:      aws.String(d.Get(attrS3AccessAssetSourceBucket).(string)),
				KeyPrefixes: d.Get(attrS3AccessAssetSourceKeyPrefixes).([]string),
				Keys:        d.Get(attrS3AccessAssetSourceKeys).([]string),
			},
		}

		if kmsKeys, ok := d.GetOk(attrS3AccessAssetSourceKmsKeysToGrant); ok {
			keyIn := make([]awstypes.KmsKeyToGrant, len(kmsKeys.([]string)))
			for i, key := range kmsKeys.([]string) {
				keyIn[i] = awstypes.KmsKeyToGrant{
					KmsKeyArn: &key,
				}
			}
			in.Details.CreateS3DataAccessFromS3Bucket.AssetSource.KmsKeysToGrant = keyIn
		}
		break
	case awstypes.TypeImportAssetsFromS3:
		in.Details.ImportAssetsFromS3 = &awstypes.ImportAssetsFromS3RequestDetails{
			DataSetId:  aws.String(d.Get(attrDataSetId).(string)),
			RevisionId: aws.String(d.Get(attrRevisionId).(string)),
		}

		if sources, ok := d.GetOk(attrS3AssetSources); ok {
			sourcesIn := make([]awstypes.AssetSourceEntry, len(sources.([]any)))
			for i, source := range sources.([]any) {
				sourceM := source.(map[string]any)
				sourcesIn[i] = awstypes.AssetSourceEntry{
					Bucket: aws.String(sourceM[attrBucket].(string)),
					Key:    aws.String(sourceM[attrKey].(string)),
				}
			}
			in.Details.ImportAssetsFromS3.AssetSources = sourcesIn
		}
		break
	case awstypes.TypeImportAssetFromSignedUrl:
		in.Details.ImportAssetFromSignedUrl = &awstypes.ImportAssetFromSignedUrlRequestDetails{
			AssetName:  aws.String(d.Get(attrUrlAssetName).(string)),
			DataSetId:  aws.String(d.Get(attrDataSetId).(string)),
			Md5Hash:    aws.String(d.Get(attrUrlMd5Hash).(string)),
			RevisionId: aws.String(d.Get(attrRevisionId).(string)),
		}
		break
	case awstypes.TypeImportAssetFromApiGatewayApi:
		in.Details.ImportAssetFromApiGatewayApi = &awstypes.ImportAssetFromApiGatewayApiRequestDetails{
			ApiId:                   aws.String(d.Get(attrApiGtwId).(string)),
			ApiName:                 aws.String(d.Get(attrApiGtwName).(string)),
			ApiSpecificationMd5Hash: aws.String(d.Get(attrApiGtwSpecMd5Hash).(string)),
			DataSetId:               aws.String(d.Get(attrDataSetId).(string)),
			ProtocolType:            awstypes.ProtocolType(d.Get(attrApiGtwProtoType).(string)),
			RevisionId:              aws.String(d.Get(attrRevisionId).(string)),
			Stage:                   aws.String(d.Get(attrApiGtwStage).(string)),
			ApiDescription:          aws.String(d.Get(attrApiGtwDescription).(string)),
			ApiKey:                  aws.String(d.Get(attrApiGtwKey).(string)),
		}
		break
	case awstypes.TypeImportAssetsFromRedshiftDataShares:
		in.Details.ImportAssetsFromRedshiftDataShares = &awstypes.ImportAssetsFromRedshiftDataSharesRequestDetails{
			DataSetId:  aws.String(d.Get(attrDataSetId).(string)),
			RevisionId: aws.String(d.Get(attrRevisionId).(string)),
		}

		if v, ok := d.GetOk(attrRedshiftAssetSources); ok {
			redshiftSrcIn := make([]awstypes.RedshiftDataShareAssetSourceEntry, len(v.([]any)))
			for i, redshiftSrcAttr := range v.([]string) {
				redshiftSrcIn[i] = awstypes.RedshiftDataShareAssetSourceEntry{
					DataShareArn: &redshiftSrcAttr,
				}
			}
			in.Details.ImportAssetsFromRedshiftDataShares.AssetSources = redshiftSrcIn
		}
		break
	case awstypes.TypeImportAssetsFromLakeFormationTagPolicy:
		in.Details.ImportAssetsFromLakeFormationTagPolicy = &awstypes.ImportAssetsFromLakeFormationTagPolicyRequestDetails{
			CatalogId:  aws.String(d.Get(attrLakeFormationCatalogId).(string)),
			DataSetId:  aws.String(d.Get(attrDataSetId).(string)),
			RevisionId: aws.String(d.Get(attrRevisionId).(string)),
			RoleArn:    aws.String(d.Get(attrLakeFormationRoleArn).(string)),
		}

		if _, ok := d.GetOk(attrLakeFormationDatabaseExpression); ok {
			in.Details.ImportAssetsFromLakeFormationTagPolicy.Database = &awstypes.DatabaseLFTagPolicyAndPermissions{
				Expression:  buildLFTagInput(d.Get(attrLakeFormationDatabaseExpression).([]any)),
				Permissions: d.Get(attrLakeFormationDatabasePermissions).([]awstypes.DatabaseLFTagPolicyPermission),
			}
		}

		if _, ok := d.GetOk(attrLakeFormationTableExpression); ok {
			in.Details.ImportAssetsFromLakeFormationTagPolicy.Table = &awstypes.TableLFTagPolicyAndPermissions{
				Expression:  buildLFTagInput(d.Get(attrLakeFormationTableExpression).([]any)),
				Permissions: d.Get(attrLakeFormationTablePermissions).([]awstypes.TableTagPolicyLFPermission),
			}
		}
	}

	return in
}

func buildLFTagInput(attr []any) []awstypes.LFTag {
	res := make([]awstypes.LFTag, len(attr))
	for i, aEl := range attr {
		aM := aEl.(map[string]any)
		res[i] = awstypes.LFTag{
			TagKey:    aws.String(aM[attrTagKey].(string)),
			TagValues: aM[attrTagValue].([]string),
		}
	}
	return res
}

// Schema
func resourceJobSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		names.AttrARN: {
			Type:     schema.TypeString,
			Computed: true,
		},
		names.AttrID: {
			Type:     schema.TypeString,
			Computed: true,
		},
		names.AttrCreatedAt: {
			Type:     schema.TypeString,
			Computed: true,
		},
		names.AttrLastUpdatedTime: {
			Type:     schema.TypeString,
			Computed: true,
		},
		names.AttrState: {
			Type:     schema.TypeString,
			Computed: true,
		},
		names.AttrType: {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: enum.Validate[awstypes.Type](),
		},
		attrDataSetId: {
			Type:     schema.TypeString,
			Required: true,
		},
		attrStartOnCreation: {
			Type:     schema.TypeBool,
			Default:  false,
			Optional: true,
		},
		attrRevisionId: { // todo: validation
			Type:     schema.TypeString,
			Optional: true,
		},
		attrS3AccessAssetSourceBucket: {
			Type:     schema.TypeString,
			Optional: true,
		},
		attrS3AccessAssetSourceKeyPrefixes: {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},
		attrS3AccessAssetSourceKeys: {
			Type:     schema.TypeSet,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Optional: true,
		},
		attrS3AccessAssetSourceKmsKeysToGrant: {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringLenBetween(1, 2048),
			},
			Optional: true,
		},
		attrS3AssetDestinations: {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					attrAssetId: {
						Type:     schema.TypeString,
						Required: true,
					},
					attrBucket: {
						Type:     schema.TypeString,
						Required: true,
					},
					attrKey: {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			Optional: true,
		},
		attrS3AssetDestinationEncryptionKMSKeyArn: {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: verify.ValidARN,
			RequiredWith: []string{attrS3AssetDestinationEncryptionType},
		},
		attrS3AssetDestinationEncryptionType: {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: enum.Validate[awstypes.ServerSideEncryptionTypes](),
			RequiredWith:     []string{attrS3AssetDestinationEncryptionKMSKeyArn},
		},
		attrAssetId: {
			Type:     schema.TypeString,
			Optional: true,
		},
		attrSignedUrl: {
			Type:     schema.TypeString,
			Computed: true,
		},
		attrSignedUrlExpiresAt: {
			Type:     schema.TypeString,
			Computed: true,
		},
		attrS3RevisionDestinations: {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					attrBucket: {
						Type:     schema.TypeString,
						Required: true,
					},
					attrRevisionId: {
						Type:     schema.TypeString,
						Required: true,
					},
					attrKeyPattern: {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			Optional: true,
		},
		attrS3EventActionArn: {
			Type:     schema.TypeString,
			Computed: true,
		},
		attrApiGtwDescription: {
			Type:     schema.TypeString,
			Optional: true,
		},
		attrApiGtwId: {
			Type:     schema.TypeString,
			Optional: true,
		},
		attrApiGtwKey: {
			Type:     schema.TypeString,
			Optional: true,
		},
		attrApiGtwName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		attrApiGtwSpecMd5Hash: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringMatch(
				regexp.MustCompile("(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?"),
				"must be base64 encrypted md5 hash",
			),
		},
		attrApiGtwProtoType: {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateDiagFunc: enum.Validate[awstypes.ProtocolType](),
		},
		attrApiGtwStage: {
			Type:     schema.TypeString,
			Optional: true,
		},
		attrApiSpecUploadUrl: {
			Type:     schema.TypeString,
			Computed: true,
		},
		attrApiSPecUploadUrlExpAt: {
			Type:     schema.TypeBool,
			Computed: true,
		},
		attrUrlAssetName: {
			Type:     schema.TypeString,
			Optional: true,
		},
		attrUrlMd5Hash: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringMatch(
				regexp.MustCompile("(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?"),
				"must be base64 encrypted md5 hash",
			),
		},
		attrRedshiftAssetSources: {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: verify.ValidARN,
			},
			Optional: true,
		},
		attrS3AssetSources: {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					attrBucket: {
						Type:     schema.TypeString,
						Required: true,
					},
					attrKey: {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			Optional: true,
		},
		attrLakeFormationCatalogId: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringMatch(
				regexp.MustCompile(".*/^[\\d]{12}$/.*"),
				"must be valid catalog id",
			),
		},
		attrLakeFormationDatabaseExpression: {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					attrTagKey: {
						Type:     schema.TypeString,
						Required: true,
					},
					attrTagValue: {
						Type:     schema.TypeSet,
						Elem:     &schema.Schema{Type: schema.TypeString},
						Required: true,
					},
				},
			},
			Optional:     true,
			RequiredWith: []string{attrLakeFormationDatabasePermissions},
		},
		attrLakeFormationDatabasePermissions: {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: enum.Validate[awstypes.DatabaseLFTagPolicyPermission](),
			},
			Optional:     true,
			RequiredWith: []string{attrLakeFormationDatabaseExpression},
		},
		attrLakeFormationRoleArn: {
			Type:     schema.TypeString,
			Optional: true,
			ValidateFunc: validation.StringMatch(
				regexp.MustCompile("arn:aws:iam::(\\d{12}):role\\/.+"),
				"must be valid IAM Role ARN",
			),
		},
		attrLakeFormationTableExpression: {
			Type: schema.TypeList,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					attrTagKey: {
						Type:     schema.TypeString,
						Required: true,
					},
					attrTagValue: {
						Type:     schema.TypeString,
						Required: true,
					},
				},
			},
			Optional:     true,
			RequiredWith: []string{attrLakeFormationTablePermissions},
		},
		attrLakeFormationTablePermissions: {
			Type: schema.TypeSet,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: enum.Validate[awstypes.TableTagPolicyLFPermission](),
			},
			Optional:     true,
			RequiredWith: []string{attrLakeFormationTableExpression},
		},
	}
}
