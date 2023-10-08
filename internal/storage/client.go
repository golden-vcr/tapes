package storage

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
)

// Metadata is the raw, unvalidated metadata associated with an object in an
// S3-compatible bucket, as a series of string key/value pairs
type Metadata map[string]string

// Client handles listing files and metadata from an S3-compatible bucket
type Client interface {
	ListFilenames(ctx context.Context) ([]string, error)
	GetFileMetadata(ctx context.Context, filename string) (Metadata, error)
}

// NewClient creates a new storage.Client that uses the AWS S3 client to access a
// DigitalOcean Spaces bucket
func NewClient(spacesAccessKeyId, spacesSecretKey, spacesEndpointOrigin, spacesRegionName, spacesBucketName string) (Client, error) {
	config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(spacesAccessKeyId, spacesSecretKey, ""),
		Endpoint:         aws.String(fmt.Sprintf("https://%s", spacesEndpointOrigin)),
		Region:           aws.String(spacesRegionName),
		S3ForcePathStyle: aws.Bool(false),
	}
	session, err := awsSession.NewSession(config)
	if err != nil {
		return nil, err
	}
	s3 := awsS3.New(session)
	return &client{
		s3:         s3,
		bucketName: spacesBucketName,
	}, nil
}

// client is the implementation of storage.Client for use with DigitalOcean Spaces
type client struct {
	s3         *awsS3.S3
	bucketName string
}

func (c *client) ListFilenames(ctx context.Context) ([]string, error) {
	// Prepare a list of filenames as our result
	filenames := make([]string, 0)
	input := &awsS3.ListObjectsV2Input{Bucket: aws.String(c.bucketName)}
	for {
		// Get a list of objects in the bucket, and append their keys to our list
		res, err := c.s3.ListObjectsV2WithContext(ctx, input)
		if err != nil {
			return nil, err
		}
		for _, obj := range res.Contents {
			filenames = append(filenames, *obj.Key)
		}

		// Continue getting paginated results until we've seen all filenames
		if res.IsTruncated != nil && *res.IsTruncated {
			input.SetContinuationToken(*res.NextContinuationToken)
		} else {
			break
		}
	}
	return filenames, nil
}

func (c *client) GetFileMetadata(ctx context.Context, filename string) (Metadata, error) {
	// Perform a HEAD request to get the headers associated with the given file,
	// including 'x-amz-meta-*' headers that encode the custom metadata we associated
	// with each file upon upload
	r, err := c.s3.HeadObjectWithContext(ctx, &awsS3.HeadObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(filename),
	})
	if err != nil {
		return nil, err
	}

	// Every metadata key should have a string value associated with it: do away with
	// the AWS SDK's pointer indirection
	result := make(map[string]string)
	for k, v := range r.Metadata {
		result[k] = *v
	}
	return result, nil
}

var _ Client = (*client)(nil)
