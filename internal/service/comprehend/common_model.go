package comprehend

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

type safeMutex struct {
	locked bool
	mutex  sync.Mutex
}

func (m *safeMutex) Lock() {
	m.mutex.Lock()
	m.locked = true
}

func (m *safeMutex) Unlock() {
	if m.locked {
		m.locked = false
		m.mutex.Unlock()
	}
}

var modelVPCENILock safeMutex

func findNetworkInterfaces(ctx context.Context, conn *ec2.EC2, securityGroups []string, subnets []string) ([]*ec2.NetworkInterface, error) {
	networkInterfaces, err := tfec2.FindNetworkInterfacesWithContext(ctx, conn, &ec2.DescribeNetworkInterfacesInput{
		Filters: []*ec2.Filter{
			tfec2.NewFilter("group-id", securityGroups),
			tfec2.NewFilter("subnet-id", subnets),
		},
	})
	if err != nil {
		return []*ec2.NetworkInterface{}, err
	}

	comprehendENIs := make([]*ec2.NetworkInterface, 0, len(networkInterfaces))
	for _, v := range networkInterfaces {
		if strings.HasSuffix(aws.ToString(v.RequesterId), ":Comprehend") {
			comprehendENIs = append(comprehendENIs, v)
		}
	}

	return comprehendENIs, nil
}

func waitNetworkInterfaceCreated(ctx context.Context, conn *ec2.EC2, initialENIIds map[string]bool, securityGroups []string, subnets []string, timeout time.Duration) (*ec2.NetworkInterface, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{},
		Target:     []string{ec2.NetworkInterfaceStatusInUse},
		Refresh:    statusNetworkInterfaces(ctx, conn, initialENIIds, securityGroups, subnets),
		Delay:      4 * time.Minute,
		MinTimeout: 10 * time.Second,
		Timeout:    timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*ec2.NetworkInterface); ok {
		return output, err
	}

	return nil, err
}

func statusNetworkInterfaces(ctx context.Context, conn *ec2.EC2, initialENIs map[string]bool, securityGroups []string, subnets []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findNetworkInterfaces(ctx, conn, securityGroups, subnets)
		if err != nil {
			return nil, "", err
		}

		var added *ec2.NetworkInterface
		for _, v := range out {
			if _, ok := initialENIs[aws.ToString(v.NetworkInterfaceId)]; !ok {
				added = v
				break
			}
		}

		if added == nil {
			return nil, "", nil
		}

		return added, aws.ToString(added.Status), nil
	}
}
