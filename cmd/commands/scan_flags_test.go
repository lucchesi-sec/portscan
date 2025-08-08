package commands

import (
	"testing"

	"github.com/spf13/viper"
)

func TestFlagBindings_StdinJson(t *testing.T) {
	// Ensure defaults are false and keys are readable
	if viper.GetBool("stdin") || viper.GetBool("json") {
		t.Fatalf("expected default stdin/json to be false")
	}
}
