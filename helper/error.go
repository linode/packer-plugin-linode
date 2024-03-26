package helper

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// errorHelper is a helper function to reduce the amount of bloat and complexity
// caused by redundant error handling logic.
func ErrorHelper(state multistep.StateBag, ui packersdk.Ui, prefix string, err error) multistep.StepAction {
	wrappedError := fmt.Errorf("%s: %w", prefix, err)
	state.Put("error", wrappedError)
	ui.Error(wrappedError.Error())
	return multistep.ActionHalt
}
