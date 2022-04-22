package cloudfront

const (
	StreamTypeKinesis = "Kinesis"
)

func StreamType_Values() []string {
	return []string{
		StreamTypeKinesis,
	}
}
