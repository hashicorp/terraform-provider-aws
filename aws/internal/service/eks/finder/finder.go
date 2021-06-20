package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func ClusterByName(conn *eks.EKS, name string) (*eks.Cluster, error) {
	input := &eks.DescribeClusterInput{
		Name: aws.String(name),
	}

	output, err := conn.DescribeCluster(input)

	// Sometimes the EKS API returns the ResourceNotFound error in this form:
	// ClientException: No cluster found for name: tf-acc-test-0o1f8
	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) || tfawserr.ErrMessageContains(err, eks.ErrCodeClientException, "No cluster found for name:") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Cluster == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Cluster, nil
}

func FargateProfileByClusterNameAndFargateProfileName(conn *eks.EKS, clusterName, fargateProfileName string) (*eks.FargateProfile, error) {
	input := &eks.DescribeFargateProfileInput{
		ClusterName:        aws.String(clusterName),
		FargateProfileName: aws.String(fargateProfileName),
	}

	output, err := conn.DescribeFargateProfile(input)

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.FargateProfile == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.FargateProfile, nil
}

func UpdateByNameAndID(conn *eks.EKS, name, id string) (*eks.Update, error) {
	input := &eks.DescribeUpdateInput{
		Name:     aws.String(name),
		UpdateId: aws.String(id),
	}

	output, err := conn.DescribeUpdate(input)

	if tfawserr.ErrCodeEquals(err, eks.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Update == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Update, nil
}
