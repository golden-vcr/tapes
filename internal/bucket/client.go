package bucket

import (
	"context"
	"fmt"
	"sort"
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

	metadata       metadataCache
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

		metadata:       make(metadataCache),
		imageCounts:    nil,
		expirationTime: time.UnixMilli(0),
	}
	if err := c.updateImageCounts(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) GetImageData(ctx context.Context) map[int][]ImageData {
	tapeIds := make([]int, 0, len(c.imageCounts))
	for tapeId := range c.imageCounts {
		tapeIds = append(tapeIds, tapeId)
	}
	sort.Ints(tapeIds)

	result := make(map[int][]ImageData)
	for tapeId := range tapeIds {
		numImages, ok := c.imageCounts[tapeId]
		if ok && numImages > 0 {
			images := make([]ImageData, 0, numImages)
			for imageIndex := 0; imageIndex < numImages; imageIndex++ {
				filename := GetImageKey(tapeId, imageIndex)
				metadata, err := c.metadata.fetch(ctx, c.s3, c.bucketName, filename)
				if err != nil || metadata == nil {
					images = nil
					break
				}
				images = append(images, ImageData{
					Filename: filename,
					Width:    metadata.width,
					Height:   metadata.height,
					Color:    metadata.color,
					Rotated:  metadata.rotated,
				})
			}
			if len(images) > 0 {
				result[tapeId] = images
			}
		}
	}
	return result
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
			parsed := ParseKey(*obj.Key)
			if parsed != nil {
				details, ok := detailsByTapeId[parsed.TapeId]
				if !ok {
					newDetails := &imageDetails{maxImageIndex: -1}
					detailsByTapeId[parsed.TapeId] = newDetails
					details = newDetails
				}
				if parsed.IsThumbnail {
					details.hasThumbnail = true
				} else if parsed.ImageIndex > details.maxImageIndex {
					details.maxImageIndex = parsed.ImageIndex
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
	fmt.Printf("Got S3 image counts: %d tapes have images\n", len(counts))
	return counts, nil
}

func (c *Client) updateImageCounts(ctx context.Context) error {
	counts, err := c.fetchImageCounts(ctx)
	if err != nil {
		return err
	}
	c.imageCounts = counts
	c.expirationTime = time.Now().Add(c.ttl)
	return c.primeMetadataCache(ctx)
}

func (c *Client) primeMetadataCache(ctx context.Context) error {
	tapeIds := make([]int, 0, len(c.imageCounts))
	for tapeId := range c.imageCounts {
		tapeIds = append(tapeIds, tapeId)
	}
	sort.Ints(tapeIds)

	for _, tapeId := range tapeIds {
		numImages := c.imageCounts[tapeId]
		for imageIndex := 0; imageIndex < numImages; imageIndex++ {
			key := GetImageKey(tapeId, imageIndex)
			_, err := c.metadata.fetch(ctx, c.s3, c.bucketName, key)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Client) getImageCounts(ctx context.Context) (map[int]int, error) {
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

type imageDetails struct {
	hasThumbnail  bool
	maxImageIndex int
}
