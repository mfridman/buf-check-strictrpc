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
					Message: "only one service definition allowed per file, but 2 were found.",
					Location: &checktest.ExpectedLocation{
						FileName: "many_services.proto",
					},
				},
			},
		}.Run(t)
	})

	t.Run("streaming", func(t *testing.T) {
		request := &checktest.RequestSpec{
			Files: &checktest.ProtoFileSpec{
				DirPaths:  []string{"testdata/streaming"},
				FilePaths: []string{"a_service.proto"},
			},
		}
		t.Run("allowed", func(t *testing.T) {
			checktest.CheckTest{
				Spec:    Spec,
				Request: request,
			}.Run(t)
		})
		t.Run("not allowed", func(t *testing.T) {
			request.Options = map[string]any{
				"disable_streaming": true,
			}
			checktest.CheckTest{
				Spec:    Spec,
				Request: request,
				ExpectedAnnotations: []checktest.ExpectedAnnotation{
					{
						RuleID:  "STRICT_RPC",
						Message: `method "GetMovie" uses streaming, which is disabled by the disable_streaming option.`,
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
	})

	t.Run("valid", func(t *testing.T) {
		checktest.CheckTest{
			Spec: Spec,
			Request: newRequest(
				"testdata/correct",
				[]string{"user/v1/user.proto", "user/v1/user_service.proto"},
				map[string]any{"disable_streaming": true},
			),
			ExpectedAnnotations: nil,
		}.Run(t)
	})
}

func newRequest(dir string, files []string, options map[string]any) *checktest.RequestSpec {
	return &checktest.RequestSpec{
		Files: &checktest.ProtoFileSpec{
			DirPaths:  []string{dir},
			FilePaths: files,
		},
		Options: options,
	}
}
