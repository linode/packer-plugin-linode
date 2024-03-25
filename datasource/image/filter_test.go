package image

import (
	"testing"

	"github.com/linode/linodego"
)

func TestImageDatasourceFilter_IDExactFilter(t *testing.T) {
	targetID := "test_image"

	images := []linodego.Image{
		{ID: "some_other_image"},
		{ID: targetID},
	}

	config := Config{ID: targetID}

	image, err := filterImageResults(images, config)
	if err != nil {
		t.Fatalf("error filtering by exact image ID: %v", err)
	}
	if image.ID != targetID {
		t.Fatalf(
			"incorrect image with ID '%q' got selected, image "+
				"with ID '%q' should be selected instead",
			image.ID, targetID,
		)
	}
}

func TestImageDatasourceFilter_IDRegexFilter(t *testing.T) {
	targetPartialID := "test_image"
	targetIDRegex := targetPartialID + "*"
	targetID := targetPartialID + "1.0"

	images := []linodego.Image{
		{ID: "some_other_image1.0"},
		{ID: targetID},
	}

	config := Config{IDRegex: targetIDRegex}

	image, err := filterImageResults(images, config)
	if err != nil {
		t.Fatalf("error filtering by regex image ID: %v", err)
	}
	if image.ID != targetID {
		t.Fatalf(
			"incorrect image with ID '%q' got selected, image "+
				"with ID '%q' should be selected instead",
			image.ID, targetID,
		)
	}
}

func TestImageDatasourceFilter_LabelRegexFilter(t *testing.T) {
	targetPartialLabel := "test_image"
	targetLabelRegex := targetPartialLabel + "*"
	targetLabel := targetPartialLabel + "1.0"

	images := []linodego.Image{
		{Label: "some_other_image1.0"},
		{Label: targetLabel},
	}

	config := Config{LabelRegex: targetLabelRegex}

	image, err := filterImageResults(images, config)
	if err != nil {
		t.Fatalf("error filtering by regex image label: %v", err)
	}
	if image.Label != targetLabel {
		t.Fatalf(
			"incorrect image with label '%q' got selected, image "+
				"with label '%q' should be selected instead",
			image.Label, targetLabel,
		)
	}
}
