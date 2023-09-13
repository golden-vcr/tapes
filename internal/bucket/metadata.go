package bucket

import (
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	awsS3 "github.com/aws/aws-sdk-go/service/s3"
)

var hexColorRegexp = regexp.MustCompile(`^#[a-zA-Z0-9]{3}(?:[a-zA-Z0-9]{3})?$`)

type metadataCache map[string]*imageMetadata

type imageMetadata struct {
	width   int
	height  int
	color   string
	rotated bool
}

func (mc metadataCache) fetch(ctx context.Context, s3 *awsS3.S3, bucketName string, key string) (*imageMetadata, error) {
	if existing, found := mc[key]; found {
		return existing, nil
	}
	fetched, err := fetchImageMetadata(ctx, s3, bucketName, key)
	if err != nil {
		return nil, err
	}
	mc[key] = fetched
	return fetched, nil
}

func fetchImageMetadata(ctx context.Context, s3 *awsS3.S3, bucketName string, key string) (*imageMetadata, error) {
	r, err := s3.HeadObjectWithContext(ctx, &awsS3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	fmt.Printf("Fetched S3 object metadata for: %s\n", key)
	return parseImageMetadata(r.Metadata), nil
}

func parseImageMetadata(md map[string]*string) *imageMetadata {
	// Parse the image dimensions: if either value is unspecified, the image does not
	// have valid metadata
	width := 0
	if widthValue, ok := md["Width"]; ok && widthValue != nil {
		if value, err := strconv.Atoi(*widthValue); err == nil && value > 0 {
			width = value
		}
	}
	height := 0
	if heightValue, ok := md["Height"]; ok && heightValue != nil {
		if value, err := strconv.Atoi(*heightValue); err == nil && value > 0 {
			height = value
		}
	}
	if width == 0 || height == 0 {
		return nil
	}

	color := "#cccccc"
	if colorValue, ok := md["Color"]; ok && colorValue != nil {
		if hexColorRegexp.MatchString(*colorValue) {
			color = *colorValue
		}
	}

	rotated := false
	if rotatedValue, ok := md["Rotated"]; ok && rotatedValue != nil {
		if *rotatedValue == "true" {
			rotated = true
		}
	}

	return &imageMetadata{
		width:   width,
		height:  height,
		color:   color,
		rotated: rotated,
	}
}
