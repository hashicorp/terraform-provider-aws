package cloudfront

const (
	StreamTypeKinesis = "Kinesis"

	ResDistribution         = "Distribution"
	ResPublicKey            = "Public Key"
	ResOriginAccessIdentity = "Origin Access Identity"
)

func StreamType_Values() []string {
	return []string{
		StreamTypeKinesis,
	}
}
