package awspolicy

// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

import (
	"testing"
)

func TestPolicyEquivalence(t *testing.T) {
	cases := []struct {
		name       string
		policy1    string
		policy2    string
		equivalent bool
		err        bool
	}{
		{
			name:       "Invalid policy JSON",
			policy1:    policyTest0,
			policy2:    policyTest0,
			equivalent: false,
			err:        true,
		},

		{
			name:       "Idential policy text",
			policy1:    policyTest1,
			policy2:    policyTest1,
			equivalent: true,
		},

		{
			name:       "Action block as single item array versus string",
			policy1:    policyTest2a,
			policy2:    policyTest2b,
			equivalent: true,
		},

		{
			name:       "Action block as single item array versus string, different action",
			policy1:    policyTest3a,
			policy2:    policyTest3b,
			equivalent: false,
		},

		{
			name:       "NotAction block and ActionBlock, mixed string versus array",
			policy1:    policyTest4a,
			policy2:    policyTest4b,
			equivalent: true,
		},

		{
			name:       "NotAction block on one side",
			policy1:    policyTest5a,
			policy2:    policyTest5b,
			equivalent: false,
		},

		{
			name:       "Principal in single item array versus string",
			policy1:    policyTest6a,
			policy2:    policyTest6b,
			equivalent: true,
		},

		{
			name:       "Different principal in single item array versus string",
			policy1:    policyTest7a,
			policy2:    policyTest7b,
			equivalent: false,
		},

		{
			name:       "String principal",
			policy1:    policyTest8a,
			policy2:    policyTest8b,
			equivalent: true,
		},

		{
			name:       "String NotPrincipal",
			policy1:    policyTest9a,
			policy2:    policyTest9b,
			equivalent: true,
		},

		{
			name:       "Different NotPrincipal in single item array versus string",
			policy1:    policyTest10a,
			policy2:    policyTest10b,
			equivalent: false,
		},

		{
			name:       "Different Effect",
			policy1:    policyTest11a,
			policy2:    policyTest11b,
			equivalent: false,
		},

		{
			name:       "Different Version",
			policy1:    policyTest12a,
			policy2:    policyTest12b,
			equivalent: false,
		},

		{
			name:       "Same Condition",
			policy1:    policyTest13a,
			policy2:    policyTest13b,
			equivalent: true,
		},

		{
			name:       "Different Condition",
			policy1:    policyTest14a,
			policy2:    policyTest14b,
			equivalent: false,
		},

		{
			name:       "Condition in single string instead of array",
			policy1:    policyTest15a,
			policy2:    policyTest15b,
			equivalent: true,
		},

		{
			name:       "Multiple Condition Blocks in one policy",
			policy1:    policyTest16a,
			policy2:    policyTest16b,
			equivalent: false,
		},

		{
			name:       "Multiple Condition Blocks, same in both policies",
			policy1:    policyTest17a,
			policy2:    policyTest17b,
			equivalent: true,
		},

		{
			name:       "Multiple Statements, Equivalent",
			policy1:    policyTest18a,
			policy2:    policyTest18b,
			equivalent: true,
		},

		{
			name:       "Multiple Statements, missing one from policy 2",
			policy1:    policyTest19a,
			policy2:    policyTest19b,
			equivalent: false,
		},

		{
			name:       "Casing of Effect",
			policy1:    policyTest20a,
			policy2:    policyTest20b,
			equivalent: true,
		},
		{
			name:       "Single Statement vs []Statement",
			policy1:    policyTest21a,
			policy2:    policyTest21b,
			equivalent: true,
		},
	}

	for _, tc := range cases {
		equal, err := PoliciesAreEquivalent(tc.policy1, tc.policy2)
		if !tc.err && err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		if tc.err && err == nil {
			t.Fatal("Expected error, none produced")
		}

		if equal != tc.equivalent {
			t.Fatalf("Bad: %s\n  Expected: %t\n       Got: %t\n", tc.name, tc.equivalent, equal)
		}
	}
}

const policyTest0 = `{
  "Version": "2012-10-17",
  "Statement": [
    {
  ]
}`

const policyTest1 = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow", "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest2a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest2b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": ["sts:AssumeRole"],
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      }
    }
  ]
}`

const policyTest3a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest3b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": ["sts:GetSessionToken"],
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      }
    }
  ]
}`

const policyTest4a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "NotAction": ["sts:GetSessionToken"]
    }
  ]
}`

const policyTest4b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": "sts:AssumeRole",
      "NotAction": "sts:GetSessionToken",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      }
    }
  ]
}`

const policyTest5a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest5b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": "sts:AssumeRole",
      "NotAction": "sts:GetSessionToken",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      }
    }
  ]
}`

const policyTest6a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest6b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": ["spotfleet.amazonaws.com"]
      }
    }
  ]
}`

const policyTest7a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest7b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": ["spotfleet.amazonaws.com"]
      }
    }
  ]
}`

const policyTest8a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest8b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": "*"
    }
  ]
}`

const policyTest9a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "NotPrincipal": "*",
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest9b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "NotPrincipal": "*"
    }
  ]
}`

const policyTest10a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "NotPrincipal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest10b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "NotPrincipal": {
        "Service": ["spotfleet.amazonaws.com"]
      }
    }
  ]
}`

const policyTest11a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest11b = `{
  "Version": "2012-06-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest12a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Deny",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest12b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest13a = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition": {
       "StringEquals": {
         "s3:x-amz-acl": [
           "public-read"
         ]
       }
     }
   }
 ]
}`

const policyTest13b = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition": {
       "StringEquals": {
         "s3:x-amz-acl": [
           "public-read"
         ]
       }
     }
   }
 ]
}`

const policyTest14a = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition": {
       "StringNotEquals": {
         "s3:x-amz-acl": [
           "public-read"
         ]
       }
     }
   }
 ]
}`

const policyTest14b = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition": {
       "StringEquals": {
         "s3:x-amz-acl": [
           "public-read"
         ]
       }
     }
   }
 ]
}`

const policyTest15a = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition": {
       "StringEquals": {
         "s3:x-amz-acl": "public-read"
       }
     }
   }
 ]
}`

const policyTest15b = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition": {
       "StringEquals": {
         "s3:x-amz-acl": [
           "public-read"
         ]
       }
     }
   }
 ]
}`

const policyTest16a = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition": {
       "StringEquals": {
         "s3:x-amz-acl": "public-read"
       }
     }
   }
 ]
}`

const policyTest16b = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition" :  {
       "DateGreaterThan" : {
         "aws:CurrentTime" : "2013-08-16T12:00:00Z"
       },
       "DateLessThan": {
         "aws:CurrentTime" : "2013-08-16T15:00:00Z"
       },
       "IpAddress" : {
         "aws:SourceIp" : ["192.0.2.0/24", "203.0.113.0/24"]
       }
     }
   }
 ]
}`

const policyTest17a = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition" :  {
       "DateGreaterThan" : {
         "aws:CurrentTime" : "2013-08-16T12:00:00Z"
       },
       "DateLessThan": {
         "aws:CurrentTime" : "2013-08-16T15:00:00Z"
       },
       "IpAddress" : {
         "aws:SourceIp" : ["192.0.2.0/24", "203.0.113.0/24"]
       }
     }
   }
 ]
}`

const policyTest17b = `{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Sid": "statement1",
     "Effect": "Allow",
     "Action": [
       "s3:PutObject"
     ],
     "Resource": [
       "arn:aws:s3:::examplebucket/*"
     ],
     "Condition" :  {
       "DateGreaterThan" : {
         "aws:CurrentTime" : "2013-08-16T12:00:00Z"
       },
       "DateLessThan": {
         "aws:CurrentTime" : "2013-08-16T15:00:00Z"
       },
       "IpAddress" : {
         "aws:SourceIp" : ["192.0.2.0/24", "203.0.113.0/24"]
       }
     }
   }
 ]
}`

const policyTest18a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:aws:s3:::*"
    },
    {
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:aws:s3:::BUCKET-NAME",
      "Condition": {"StringLike": {"s3:prefix": [
        "",
        "home/",
        "home/${aws:username}/"
      ]}}
    },
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::BUCKET-NAME/home/${aws:username}",
        "arn:aws:s3:::BUCKET-NAME/home/${aws:username}/*"
      ]
    }
  ]
}`

const policyTest18b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:aws:s3:::*"
    },
    {
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:aws:s3:::BUCKET-NAME",
      "Condition": {"StringLike": {"s3:prefix": [
        "",
        "home/",
        "home/${aws:username}/"
      ]}}
    },
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::BUCKET-NAME/home/${aws:username}",
        "arn:aws:s3:::BUCKET-NAME/home/${aws:username}/*"
      ]
    }
  ]
}`

const policyTest19a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:aws:s3:::*"
    },
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::BUCKET-NAME/home/${aws:username}",
        "arn:aws:s3:::BUCKET-NAME/home/${aws:username}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": "arn:aws:s3:::BUCKET-NAME",
      "Condition": {"StringLike": {"s3:prefix": [
        "",
        "home/",
        "home/${aws:username}/"
      ]}}
    }
  ]
}`

const policyTest19b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:ListAllMyBuckets",
        "s3:GetBucketLocation"
      ],
      "Resource": "arn:aws:s3:::*"
    },
    {
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": [
        "arn:aws:s3:::BUCKET-NAME/home/${aws:username}",
        "arn:aws:s3:::BUCKET-NAME/home/${aws:username}/*"
      ]
    }
  ]
}`

const policyTest20a = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}`

const policyTest20b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": ["sts:AssumeRole"],
      "Effect": "allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      }
    }
  ]
}`

const policyTest21a = `{
  "Version": "2012-10-17",
  "Statement":
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
}`

const policyTest21b = `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Action": ["sts:AssumeRole"],
      "Effect": "allow",
      "Principal": {
        "Service": "spotfleet.amazonaws.com"
      }
    }
  ]
}`
