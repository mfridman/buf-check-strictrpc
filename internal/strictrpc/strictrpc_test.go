package strictrpc

import (
	"testing"

	"github.com/bufbuild/bufplugin-go/check/checktest"
)

func TestRule(t *testing.T) {
	t.Parallel()

	t.Run("invalid", func(t *testing.T) {
		checktest.CheckTest{
			Spec: Spec,
			Request: &checktest.RequestSpec{
				Files: &checktest.ProtoFileSpec{
					DirPaths:  []string{"testdata/multiple"},
					FilePaths: []string{"many_services.proto"},
				},
			},
			ExpectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  "STRICT_RPC",
					Message: "only one service definition allowed per file, but 2 were found",
					Location: &checktest.ExpectedLocation{
						FileName: "many_services.proto",
					},
				},
			},
		}.Run(t)
	})

	t.Run("streaming no allowed", func(t *testing.T) {
		checktest.CheckTest{
			Spec: Spec,
			Request: &checktest.RequestSpec{
				Files: &checktest.ProtoFileSpec{
					DirPaths:  []string{"testdata/streaming"},
					FilePaths: []string{"a_service.proto"},
				},
			},
			ExpectedAnnotations: []checktest.ExpectedAnnotation{
				{
					RuleID:  "STRICT_RPC",
					Message: `method "GetMovie" is streaming, but streaming is not allowed`,
					Location: &checktest.ExpectedLocation{
						FileName:    "a_service.proto",
						StartLine:   2,
						EndLine:     4,
						StartColumn: 0,
						EndColumn:   1,
					},
				},
			},
		}.Run(t)
	})

	t.Run("valid", func(t *testing.T) {
		checktest.CheckTest{
			Spec: Spec,
			Request: &checktest.RequestSpec{
				Files: &checktest.ProtoFileSpec{
					DirPaths: []string{"testdata/correct"},
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
