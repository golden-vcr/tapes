package storage

import (
	"context"
	"fmt"
	"sort"
)

type Warning struct {
	Filename string
	Message  string
}

func ListImages(ctx context.Context, c Client) ([]Image, []Warning, error) {
	// List the files in the S3-compatible bucket where we store scanned images of tapes
	rawFilenames, err := c.ListFilenames(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list filenames from storage bucket: %w", err)
	}

	// Collect a list of warnings for files that aren't named according to the expected
	// conventions, are missing related images, that don't have the required metadata,
	// etc., and keep track of which tape IDs have problematic images so we can omit all
	// images for that tape
	warnings := make([]Warning, 0)
	tapeIdsWithWarnings := make(map[int]struct{})

	// Parse each image filename, and sort each valid image file into one of two
	// categories - thumbnail images and gallery images - indexed by tape ID
	thumbnailImagesByTapeId := make(map[int]*Image)
	galleryImagesByTapeId := make(map[int][]*Image)
	for _, filename := range rawFilenames {
		imageId, err := parseImageFilename(filename)
		if err != nil {
			// If any file in the bucket is not a valid tape image, log a warning
			warnings = append(warnings, Warning{
				Filename: filename,
				Message:  err.Error(),
			})
			continue
		}
		if imageId.imageType == ImageTypeThumbnail {
			// Cache this image as the thumbnail for its tape
			if existing, found := thumbnailImagesByTapeId[imageId.tapeId]; found {
				warnings = append(warnings, Warning{
					Filename: filename,
					Message:  fmt.Sprintf("duplicate thumbnail image for tape %d (already have %s)", imageId.tapeId, existing.Filename),
				})
				tapeIdsWithWarnings[imageId.tapeId] = struct{}{}
				continue
			}
			thumbnailImagesByTapeId[imageId.tapeId] = &Image{
				Filename: filename,
				TapeId:   imageId.tapeId,
				Type:     ImageTypeThumbnail,
			}
		} else {
			// For a gallery image, fetch metadata from S3: if unable, fail hard
			md, err := c.GetFileMetadata(ctx, filename)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to get metadata for image file %s: %w", filename, err)
			}

			// Parse the raw key/value pairs to an ImageMetadata struct, which
			// represents the required fields that all images must have: if unable, log
			// a warning
			metadata, err := md.toImageMetadata()
			if err != nil {
				warnings = append(warnings, Warning{
					Filename: filename,
					Message:  err.Error(),
				})
				tapeIdsWithWarnings[imageId.tapeId] = struct{}{}
				continue
			}

			// Add this image to the list of gallery images cached for its tape
			galleryImagesByTapeId[imageId.tapeId] = append(galleryImagesByTapeId[imageId.tapeId], &Image{
				Filename: filename,
				TapeId:   imageId.tapeId,
				Type:     ImageTypeGallery,
				GalleryData: &GalleryImageData{
					Index:    imageId.galleryIndex,
					Metadata: metadata,
				},
			})
		}
	}

	// If any image file for a particular tape was invalid, forget about all other
	// images for that tape, so that we refuse to admit a tape until warnings are
	// addressed and we can guarantee that all its images are present and well-formed
	for tapeId := range tapeIdsWithWarnings {
		delete(thumbnailImagesByTapeId, tapeId)
		delete(galleryImagesByTapeId, tapeId)
	}

	// We require every tape to have a thumbnail image and at least one gallery image:
	// if there are any tapes which have images of one type but not another, log
	// warnings and exclude those tapes as well
	tapeIdsWithWarnings = make(map[int]struct{})
	for thumbnailTapeId, thumbnailImage := range thumbnailImagesByTapeId {
		if galleryImages, ok := galleryImagesByTapeId[thumbnailTapeId]; !ok || len(galleryImages) == 0 {
			warnings = append(warnings, Warning{
				Filename: thumbnailImage.Filename,
				Message:  fmt.Sprintf("tape %d has thumbnail image but no accompanying gallery image(s)", thumbnailTapeId),
			})
			tapeIdsWithWarnings[thumbnailTapeId] = struct{}{}
		}
	}
	for galleryTapeId, galleryImages := range galleryImagesByTapeId {
		if _, ok := thumbnailImagesByTapeId[galleryTapeId]; !ok {
			warnings = append(warnings, Warning{
				Filename: galleryImages[0].Filename,
				Message:  fmt.Sprintf("tape %d has gallery image(s) but no accompanying thumbnail image", galleryTapeId),
			})
			tapeIdsWithWarnings[galleryTapeId] = struct{}{}
		}
	}
	for tapeId := range tapeIdsWithWarnings {
		delete(thumbnailImagesByTapeId, tapeId)
		delete(galleryImagesByTapeId, tapeId)
	}

	// Finally, push all remaining images into a flat array, sorted by tape ID and with
	// gallery images sorted in display order for the sake of determinism
	images := make([]Image, 0, len(thumbnailImagesByTapeId)*4)
	tapeIds := make([]int, 0, len(thumbnailImagesByTapeId))
	for tapeId := range thumbnailImagesByTapeId {
		tapeIds = append(tapeIds, tapeId)
	}
	sort.Ints(tapeIds)
	for _, tapeId := range tapeIds {
		thumbnailImage := thumbnailImagesByTapeId[tapeId]
		images = append(images, *thumbnailImage)

		galleryImages := galleryImagesByTapeId[tapeId]
		sort.Slice(galleryImages, func(i, j int) bool {
			return galleryImages[i].GalleryData.Index < galleryImages[j].GalleryData.Index
		})
		for _, galleryImage := range galleryImages {
			images = append(images, *galleryImage)
		}
	}

	return images, warnings, nil
}
