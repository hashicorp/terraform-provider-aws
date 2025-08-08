// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func Test_resourceRecordSetEqual(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s1   awstypes.ResourceRecordSet
		s2   awstypes.ResourceRecordSet
		want bool
	}{
		{
			name: "equal",
			s1: awstypes.ResourceRecordSet{
				Name: aws.String("example.com"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			s2: awstypes.ResourceRecordSet{
				Name: aws.String("example.com"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			want: true,
		},
		{
			name: "equal, normalized name",
			s1: awstypes.ResourceRecordSet{
				Name: aws.String("EXAMPLE.COM"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			s2: awstypes.ResourceRecordSet{
				Name: aws.String("example.com."),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			want: true,
		},
		{
			name: "not equal, name",
			s1: awstypes.ResourceRecordSet{
				Name: aws.String("example.com"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			s2: awstypes.ResourceRecordSet{
				Name: aws.String("otherexample.com"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			want: false,
		},
		{
			name: "not equal, type",
			s1: awstypes.ResourceRecordSet{
				Name: aws.String("example.com"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			s2: awstypes.ResourceRecordSet{
				Name: aws.String("example.com"),
				Type: awstypes.RRTypeAaaa,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			want: false,
		},
		{
			name: "not equal, set identifier",
			s1: awstypes.ResourceRecordSet{
				Name:          aws.String("example.com"),
				Type:          awstypes.RRTypeA,
				SetIdentifier: aws.String("foo"),
				TTL:           aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			s2: awstypes.ResourceRecordSet{
				Name:          aws.String("example.com"),
				Type:          awstypes.RRTypeA,
				TTL:           aws.Int64(300),
				SetIdentifier: aws.String("bar"),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			want: false,
		},
		{
			name: "not equal, region",
			s1: awstypes.ResourceRecordSet{
				Name:   aws.String("example.com"),
				Type:   awstypes.RRTypeA,
				TTL:    aws.Int64(300),
				Region: awstypes.ResourceRecordSetRegionUsEast1,
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			s2: awstypes.ResourceRecordSet{
				Name:   aws.String("example.com"),
				Type:   awstypes.RRTypeA,
				TTL:    aws.Int64(300),
				Region: awstypes.ResourceRecordSetRegionUsWest2,
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			want: false,
		},
		{
			name: "not equal, resource records",
			s1: awstypes.ResourceRecordSet{
				Name: aws.String("example.com"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			s2: awstypes.ResourceRecordSet{
				Name: aws.String("example.com"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.2"),
					},
				},
			},
			want: false,
		},
		{
			name: "not equal, ttl",
			s1: awstypes.ResourceRecordSet{
				Name: aws.String("example.com"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(300),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			s2: awstypes.ResourceRecordSet{
				Name: aws.String("example.com"),
				Type: awstypes.RRTypeA,
				TTL:  aws.Int64(100),
				ResourceRecords: []awstypes.ResourceRecord{
					{
						Value: aws.String("127.0.0.1"),
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := resourceRecordSetEqual(tt.s1, tt.s2); got != tt.want {
				t.Errorf("resourceRecordSetEqual() = %v, want %v", got, tt.want)
			}
		})
	}
}
