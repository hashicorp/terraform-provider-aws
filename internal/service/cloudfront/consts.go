package cloudfront

const (
	StreamTypeKinesis = "Kinesis"

	ResNameDistribution         = "Distribution"
	ResNamePublicKey            = "Public Key"
	ResNameOriginAccessIdentity = "Origin Access Identity"
)

func StreamType_Values() []string {
	return []string{
		StreamTypeKinesis,
	}
}
