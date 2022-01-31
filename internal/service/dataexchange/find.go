package dataexchange

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dataexchange"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
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
