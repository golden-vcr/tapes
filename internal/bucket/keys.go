package bucket

import (
	"fmt"
	"regexp"
	"strconv"
)

func GetThumbnailKey(tapeId int) string {
	return fmt.Sprintf("%04d_thumb.jpg", tapeId)
}

func GetImageKey(tapeId int, imageIndex int) string {
	if imageIndex < 0 {
		imageIndex = 0
	}
	if imageIndex > 25 {
		imageIndex = 25
	}
	ord := 'a' + imageIndex
	return fmt.Sprintf("%04d_%c.jpg", tapeId, ord)
}

type parseKeyResult struct {
	tapeId      int
	imageIndex  int
	isThumbnail bool
}

func parseKey(k string) *parseKeyResult {
	re := regexp.MustCompile(`^(\d{4})_(thumb|[a-z])\.jpg$`)
	match := re.FindStringSubmatch(k)
	if match != nil {
		tapeId, _ := strconv.Atoi(match[1])
		if match[2] == "thumb" {
			return &parseKeyResult{tapeId: tapeId, isThumbnail: true}
		} else {
			ord := match[2][0]
			imageIndex := ord - 'a'
			return &parseKeyResult{tapeId: tapeId, imageIndex: int(imageIndex)}
		}
	}
	return nil
}
