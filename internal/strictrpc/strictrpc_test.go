package strictrpc

import (
	"testing"

	"github.com/bufbuild/bufplugin-go/check"
	"github.com/bufbuild/bufplugin-go/check/checktest"
)

// TODO(mf): debugging is a bit of a pain, what is the annotation message, how can I print it, or view it?

func TestRule(t *testing.T) {
	t.Parallel()

	t.Run("invalid", func(t *testing.T) {
		checktest.TestCase{
			Spec: &check.Spec{Rules: []*check.RuleSpec{Rule}},
			Request: &checktest.RequestSpec{
				Files: &checktest.ProtoFileSpec{
					DirPaths: []string{
						"testdata/multiple",
					},
					FilePaths: []string{
						"many_services.proto",
					},
				},
			},
			ExpectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID: "STRICT_RPC",
					Location: &checktest.ExpectedLocation{
						FileName: "many_services.proto",
					},
				},
			},
		}.Run(t)
	})

	t.Run("valid", func(t *testing.T) {
		checktest.TestCase{
			Spec: &check.Spec{Rules: []*check.RuleSpec{Rule}},
			Request: &checktest.RequestSpec{
				Files: &checktest.ProtoFileSpec{
					DirPaths: []string{
						"testdata/correct",
					},
					FilePaths: []string{
						"user/v1/user.proto",
						"user/v1/user_service.proto",
					},
				},
			},
			ExpectedAnnotations: nil,
		}.Run(t)
	})
}
