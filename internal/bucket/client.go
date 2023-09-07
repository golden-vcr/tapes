package bucket

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	awsSession "github.com/aws/aws-sdk-go/aws/session"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
)

type Client struct {
	s3         *awsS3.S3
	bucketName string
	ttl        time.Duration

	bucketUrl string

	imageCounts    map[int]int
	expirationTime time.Time
}

func NewClient(ctx context.Context, accessKeyId string, secretKey string, endpoint string, region string, bucketName string, ttl time.Duration) (*Client, error) {
	config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKeyId, secretKey, ""),
		Endpoint:         aws.String(fmt.Sprintf("https://%s", endpoint)),
		Region:           aws.String(region),
		S3ForcePathStyle: aws.Bool(false),
	}
	session, err := awsSession.NewSession(config)
	if err != nil {
		return nil, err
	}
	s3 := awsS3.New(session)

	c := &Client{
		s3:         s3,
		bucketName: bucketName,
		ttl:        ttl,

		bucketUrl: fmt.Sprintf("https://%s.%s", bucketName, endpoint),

		imageCounts:    nil,
		expirationTime: time.UnixMilli(0),
	}
	if err := c.updateImageCounts(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) GetImageHostURL() string {
	return c.bucketUrl
}

func (c *Client) GetImageCounts(ctx context.Context) (map[int]int, error) {
	var err error
	if time.Now().After(c.expirationTime) {
		err = c.updateImageCounts(ctx)
		if err != nil && c.imageCounts != nil {
			fmt.Printf("WARNING: Failed to fetch image counts; reusing stale data: %v", err)
			err = nil
		}
	}
	return c.imageCounts, err
}

func (c *Client) fetchImageCounts(ctx context.Context) (map[int]int, error) {
	detailsByTapeId := make(map[int]*imageDetails)
	input := &awsS3.ListObjectsV2Input{Bucket: aws.String(c.bucketName)}
	for {
		res, err := c.s3.ListObjectsV2WithContext(ctx, input)
		if err != nil {
			return nil, err
		}

		for i := range res.Contents {
			obj := res.Contents[i]
			parsed := parseKey(*obj.Key)
			if parsed != nil {
				details, ok := detailsByTapeId[parsed.tapeId]
				if !ok {
					newDetails := &imageDetails{maxImageIndex: -1}
					detailsByTapeId[parsed.tapeId] = newDetails
					details = newDetails
				}
				if parsed.isThumbnail {
					details.hasThumbnail = true
				} else if parsed.imageIndex > details.maxImageIndex {
					details.maxImageIndex = parsed.imageIndex
				}
			}
		}

		if res.IsTruncated != nil && *res.IsTruncated {
			input.SetContinuationToken(*res.NextContinuationToken)
		} else {
			break
		}
	}

	counts := make(map[int]int)
	for tapeId, details := range detailsByTapeId {
		if details.hasThumbnail && details.maxImageIndex >= 0 {
			counts[tapeId] = details.maxImageIndex + 1
		}
	}
	fmt.Printf("Got S3 image metadata: %d tapes have images\n", len(counts))
	return counts, nil
}

func (c *Client) updateImageCounts(ctx context.Context) error {
	counts, err := c.fetchImageCounts(ctx)
	if err != nil {
		return err
	}
	c.imageCounts = counts
	c.expirationTime = time.Now().Add(c.ttl)
	return nil
}

type imageDetails struct {
	hasThumbnail  bool
	maxImageIndex int
}
