package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/linode/linodego"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

const (
	packerTemplate = "template/test_image_template.json"
)

func TestBuildPackerImage(t *testing.T) {
	linodeToken := os.Getenv("LINODE_TOKEN")

	if linodeToken == "" {
		t.Fatal("Linode token is not set. Please set LINODE_TOKEN as environment variable.")
	}

	// Run the Packer build command from terminal
	cmd := exec.Command("packer", "build", packerTemplate)
	output, err := cmd.CombinedOutput()
	// Check if the Packer build was successful
	if err != nil {
		t.Fatalf("Error building Packer image: %v\nOutput:\n%s", err, output)
	}

	// Assert the output contains expected strings
	expectedSubstring := "Builds finished. The artifacts of successful builds"
	assert.True(t, strings.Contains(string(output), expectedSubstring), "Expected successful build output to contain: %s", expectedSubstring)

	// Assert other fields
	err = assertLinodeImage("test-packer-image-", t)

	if err != nil {
		t.Fatalf("Error asserting Linode builder image: %v", err)
	}

	defer func() {
		if err := teardown(); err != nil {
			fmt.Printf("Error during deleting image after test execution: %v\n", err)
		}
	}()
}

func assertLinodeImage(imageLabelPrefix string, t *testing.T) error {
	client := getLinodegoClient()

	images, err := client.ListImages(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error listing Linode images: %v", err)
	}

	// Find the desired image by label prefix
	var targetImage *linodego.Image
	for _, image := range images {
		if image.Label != "" && strings.HasPrefix(image.Label, imageLabelPrefix) {
			targetImage = &image
			break
		}
	}

	if targetImage == nil {
		return fmt.Errorf("image with label prefix '%s' not found", imageLabelPrefix)
	}

	assert.Equal(t, "manual", targetImage.Type, "unexpected instance type")
	expectedInstanceIDFormat := "private/"
	assert.True(t, strings.HasPrefix(targetImage.ID, expectedInstanceIDFormat), "unexpected instance ID prefix")
	expectedInstanceLabel := "test-packer-image-"
	assert.True(t, strings.HasPrefix(targetImage.Label, expectedInstanceLabel), "unexpected instance label prefix")
	expectedImageDescription := "My Test Image Description"
	assert.Equal(t, expectedImageDescription, targetImage.Description, "unexpected image description")

	return nil
}

func teardown() error {
	client := getLinodegoClient()
	images, err := client.ListImages(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error listing Linode images: %v", err)
	}

	for _, image := range images {
		if image.Label != "" && strings.HasPrefix(image.Label, "test-packer-image-") {
			err = client.DeleteImage(context.Background(), image.ID)
			if err != nil {
				return fmt.Errorf("error during Linode image deletion: %v", err)
			}
		}
	}

	return nil
}

func getLinodegoClient() linodego.Client {
	linodeToken := os.Getenv("LINODE_TOKEN")

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: linodeToken})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}
	linodeClient := linodego.NewClient(oauth2Client)

	return linodeClient
}
