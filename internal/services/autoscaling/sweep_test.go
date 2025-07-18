package autoscaling_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	autoscalingtestfuncs "github.com/scaleway/terraform-provider-scaleway/v2/internal/services/autoscaling/testfuncs"
)

func init() {
	autoscalingtestfuncs.AddTestSweepers()
}

func TestMain(m *testing.M) {
	resource.TestMain(m)
}
