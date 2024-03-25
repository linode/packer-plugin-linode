package image

import (
	"testing"

	"github.com/linode/packer-plugin-linode/helper"
)

func TestImageDatasourceConfigure_MissingToken(t *testing.T) {
	t.Setenv(helper.TokenEnvVar, "")

	datasource := Datasource{
		config: Config{},
	}
	if err := datasource.Configure(nil); err == nil {
		t.Fatalf(
			"Should error if both environment variable %q "+
				"and linode_token config are unset",
			helper.TokenEnvVar,
		)
	}
}

func TestImageDatasourceConfigure_EnvToken(t *testing.T) {
	t.Setenv(helper.TokenEnvVar, "IAMATOKEN")

	datasource := Datasource{
		config: Config{},
	}
	if err := datasource.Configure(nil); err != nil {
		t.Fatalf(
			"Should not error if environment variable %q is set.",
			helper.TokenEnvVar,
		)
	}
}

func TestImageDatasourceConfigure_ConfigToken(t *testing.T) {
	t.Setenv(helper.TokenEnvVar, "")

	datasource := Datasource{
		config: Config{
			LinodeCommon: helper.LinodeCommon{
				PersonalAccessToken: "IAMATOKEN",
			},
		},
	}
	if err := datasource.Configure(nil); err != nil {
		t.Fatalf("Should not error if linode_token is configured.")
	}
}
