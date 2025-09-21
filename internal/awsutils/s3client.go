package awsutils

import (
    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

// NewS3Client returns a configured *s3.Client.
// - cfg: AWS SDK config
// - endpoint: optional custom S3 endpoint URL (empty = AWS default)
// - usePathStyle: enable path-style addressing (for local S3-compatible endpoints)
func NewS3Client(cfg aws.Config, endpoint string, usePathStyle bool) *s3.Client {
    opts := []func(*s3.Options){}

    opts = append(opts, func(o *s3.Options) {
        o.UsePathStyle = usePathStyle
        if endpoint != "" {
            o.EndpointResolver = s3.EndpointResolverFunc(func(service string, _ s3.EndpointResolverOptions) (aws.Endpoint, error) {
                return aws.Endpoint{
                    URL: endpoint,
                    SigningRegion: cfg.Region,
                }, nil
            })
        }
    })

    return s3.NewFromConfig(cfg, opts...)
}