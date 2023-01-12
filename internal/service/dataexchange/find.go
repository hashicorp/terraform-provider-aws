package dataexchange

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dataexchange"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindDataSetById(conn *dataexchange.DataExchange, id string) (*dataexchange.GetDataSetOutput, error) {
	input := &dataexchange.GetDataSetInput{
		DataSetId: aws.String(id),
	}
	output, err := conn.GetDataSet(input)

	if tfawserr.ErrCodeEquals(err, dataexchange.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindRevisionById(conn *dataexchange.DataExchange, dataSetId, revisionId string) (*dataexchange.GetRevisionOutput, error) {
	input := &dataexchange.GetRevisionInput{
		DataSetId:  aws.String(dataSetId),
		RevisionId: aws.String(revisionId),
	}
	output, err := conn.GetRevision(input)

	if tfawserr.ErrCodeEquals(err, dataexchange.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
