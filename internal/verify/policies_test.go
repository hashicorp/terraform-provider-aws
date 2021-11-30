package verify

import "testing"

func TestPolicyToSet(t *testing.T) {
	testCases := []struct {
		name      string
		oldPolicy string
		newPolicy string
		want      string
	}{
		{
			name: "new in random order",
			oldPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/felixjaehn",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/kidnap",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/tinlicker"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			newPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/kidnap",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			want: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/felixjaehn",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/kidnap",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/tinlicker"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
		},
		{
			name: "actual change",
			oldPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/felixjaehn",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/kidnap",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/tinlicker"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			newPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			want: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
		},
		{
			name:      "empty old",
			oldPolicy: "",
			newPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			want: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
		},
		{
			name: "empty new",
			oldPolicy: `{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "arn:aws:iam::012345678901:role/tinlicker",
          "arn:aws:iam::012345678901:role/paulvandyk",
          "arn:aws:iam::012345678901:role/garethemery",
          "arn:aws:iam::012345678901:role/felixjaehn"
        ]
      },
      "Action": [
        "kms:CreateKey",
        "kms:DescribeKey",
        "kms:ScheduleKeyDeletion",
        "kms:Describe*",
        "kms:Get*",
        "kms:List*",
        "kms:TagResource",
        "kms:UntagResource"
      ],
      "Resource": "*"
    }
  ]
}`,
			newPolicy: "",
			want:      "",
		},
	}

	for _, v := range testCases {
		got, err := PolicyToSet(v.oldPolicy, v.newPolicy)

		if err != nil {
			t.Fatalf("unexpected error with test case %s: %s", v.name, err)
		}

		if got != v.want {
			t.Fatalf("for test case %s, got %s, wanted %s", v.name, got, v.want)
		}
	}
}
