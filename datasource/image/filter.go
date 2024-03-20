package image

import (
	"errors"
	"log"
	"regexp"
	"sort"

	"github.com/linode/linodego"
)

type ImageFilter func(linodego.Image) bool

func filterImages(images []linodego.Image, filter ImageFilter) []linodego.Image {
	result := make([]linodego.Image, 0)

	for _, image := range images {
		if filter(image) {
			result = append(result, image)
		}
	}

	return result
}

func filterImagesByID(images []linodego.Image, id string) []linodego.Image {
	idFilter := func(image linodego.Image) bool {
		return image.ID == id
	}
	return filterImages(images, idFilter)
}

func filterImagesByIDRegex(images []linodego.Image, idRegex string) []linodego.Image {
	r := regexp.MustCompile(idRegex)
	idRegexFilter := func(image linodego.Image) bool {
		return r.MatchString(image.ID)
	}
	return filterImages(images, idRegexFilter)
}

func filterImagesByLabelRegex(images []linodego.Image, labelRegex string) []linodego.Image {
	r := regexp.MustCompile(labelRegex)
	labelRegexFilter := func(image linodego.Image) bool {
		return r.MatchString(image.Label)
	}
	return filterImages(images, labelRegexFilter)
}

func filterImageResults(images []linodego.Image, config Config) (linodego.Image, error) {
	if config.LabelRegex != "" {
		images = filterImagesByLabelRegex(images, config.LabelRegex)
	}
	if config.ID != "" {
		images = filterImagesByID(images, config.ID)
	}
	if config.IDRegex != "" {
		images = filterImagesByIDRegex(images, config.IDRegex)
	}
	if len(images) > 1 {

		if config.Latest {
			log.Default().Print(images)
			sort.Slice(images, func(i, j int) bool {
				return images[i].Created.After(*images[j].Created)
			})
			log.Default().Print(images)
			return images[0], nil
		}

		return linodego.Image{}, errors.New(
			"Multiple images found. Please try a more specific search, " +
				"or set latest to true in the data source config block.",
		)
	}
	if len(images) == 0 {
		return linodego.Image{}, errors.New("No image found.")
	}

	return images[0], nil
}
