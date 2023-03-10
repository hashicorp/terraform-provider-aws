package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/hashicorp/terraform/helper/schema"
)

type pointerStruct struct {
	aBool   *bool
	aFloat  *float64
	anInt   *int64
	aString *string
	aTime   *time.Time
}

func dereferenceAndWrap() pointerStruct {
	var (
		aBool   *bool
		aFloat  *float64
		anInt   *int64
		aString *string
		aTime   *time.Time
	)

	return pointerStruct{
		// ruleid: pointer-conversion-dereference-and-wrap
		aBool: aBool,
		// ruleid: pointer-conversion-dereference-and-wrap
		aFloat: aFloat,
		// ruleid: pointer-conversion-dereference-and-wrap
		anInt: anInt,
		// ruleid: pointer-conversion-dereference-and-wrap
		aString: aString,
		// ruleid: pointer-conversion-dereference-and-wrap
		aTime: aTime,
	}
}

type valueStruct struct {
	aBool   bool
	aFloat  float64
	anInt   int64
	aString string
	aTime   time.Time
}

func wrapAndDereference() valueStruct {
	var (
		aBool   bool
		aFloat  float64
		anInt   int64
		aString string
		aTime   time.Time
	)

	return valueStruct{
		// ruleid: pointer-conversion-wrap-and-dereference
		aBool: aBool,
		// ruleid: pointer-conversion-wrap-and-dereference
		aFloat: aFloat,
		// ruleid: pointer-conversion-wrap-and-dereference
		anInt: anInt,
		// ruleid: pointer-conversion-wrap-and-dereference
		aString: aString,
		// ruleid: pointer-conversion-wrap-and-dereference
		aTime: aTime,
	}
}

func setID(d *schema.ResourceData) {
	var value *string

	// ruleid: pointer-conversion-ResourceData-SetId
	d.SetId(aws.ToString(value))
}

func intConversion() int {
	var i *int64

	// ruleid: int64-pointer-dereference-and-convert
	return int(aws.Int64Value(i))
}

func assignment() string {
	var ptr *string

	// ruleid: pointer-conversion-on-assignment
	val := *ptr

	return val
}

func assignmentToRef(ref *string) {
	var ptr *string

	// ok: pointer-conversion-on-assignment
	*ref = *ptr
}
