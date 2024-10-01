package strictrpc

import (
	"testing"

	"buf.build/go/bufplugin/check/checktest"
)

func TestRule(t *testing.T) {
	t.Parallel()

	t.Run("invalid", func(t *testing.T) {
		request := newRequest(
			"testdata/multiple",
			[]string{"many_services.proto"},
			nil,
		)
		want := checktest.ExpectedAnnotation{
			RuleID:  RuleID,
			Message: "only one service definition allowed per file, but 2 were found.",
			FileLocation: &checktest.ExpectedFileLocation{
				FileName: "many_services.proto",
			},
		}
		runCheckTest(t, request, want)
	})

	t.Run("streaming", func(t *testing.T) {
		t.Run("allowed", func(t *testing.T) {
			request := newRequest(
				"testdata/streaming",
				[]string{"a_service.proto"},
				nil,
			)
			runCheckTest(t, request)
		})
		t.Run("not allowed", func(t *testing.T) {
			request := newRequest(
				"testdata/streaming",
				[]string{"a_service.proto"},
				map[string]any{"disable_streaming": true},
			)
			want := checktest.ExpectedAnnotation{
				RuleID:  RuleID,
				Message: `method "GetMovie" uses streaming, which is disabled by the disable_streaming option.`,
				FileLocation: &checktest.ExpectedFileLocation{
					FileName:    "a_service.proto",
					StartLine:   2,
					EndLine:     4,
					StartColumn: 0,
					EndColumn:   1,
				},
			}
			runCheckTest(t, request, want)
		})
	})

	t.Run("valid", func(t *testing.T) {
		request := newRequest(
			"testdata/correct",
			[]string{"user/v1/user.proto", "user/v1/user_service.proto"},
			nil,
		)
		runCheckTest(t, request)
	})
}

func runCheckTest(t *testing.T, request *checktest.RequestSpec, want ...checktest.ExpectedAnnotation) {
	checktest.CheckTest{
		Spec:                Spec,
		Request:             request,
		ExpectedAnnotations: want,
	}.Run(t)
}

func newRequest(dir string, files []string, options map[string]any) *checktest.RequestSpec {
	return &checktest.RequestSpec{
		Files: &checktest.ProtoFileSpec{
			DirPaths:  []string{dir},
			FilePaths: files,
		},
		Options: options,
		// RuleIDs: []string{RuleID}, // The plugin is set to default=true
	}
}
