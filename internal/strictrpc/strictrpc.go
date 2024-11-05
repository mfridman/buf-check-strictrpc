package strictrpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"buf.build/go/bufplugin/check"
	"buf.build/go/bufplugin/info"
	"buf.build/go/bufplugin/option"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var Spec = &check.Spec{
	Rules: []*check.RuleSpec{Rule},
	Info: &info.Spec{
		SPDXLicenseID: "MIT",
		LicenseURL:    "https://github.com/mfridman/buf-check-strictrpc/blob/main/LICENSE",
		Documentation: `Enforces an opinionated structure for RPC definitions, including strict file naming, single-service per file, and consistent request/response message naming patterns.`,
	},
}

const RuleID = "STRICT_RPC"

var Rule = &check.RuleSpec{
	ID:      RuleID,
	Purpose: "Ensures that RPCs are defined in a strict manner.",
	Type:    check.RuleTypeLint,
	Handler: check.RuleHandlerFunc(ruleFunc),

	// TODO(mf): As an end-user, I'm not sure how to correctly use these fields. It feels like
	// Buf-specific concepts are leaking. Would like some guidance on this. For example, if I'm
	// writing a specific plugin, why wouldn't default always be true?
	Default:        true,
	CategoryIDs:    nil,
	Deprecated:     false,
	ReplacementIDs: nil,
}

type result struct {
	msg  string
	desc protoreflect.Descriptor
}

func newResultf(desc protoreflect.Descriptor, msg string, args ...any) *result {
	if len(args) == 0 {
		return &result{
			msg:  msg,
			desc: desc,
		}
	}
	return &result{
		msg:  fmt.Sprintf(msg, args...),
		desc: desc,
	}
}

type config struct {
	disableStreaming   bool
	allowProtobufEmpty bool
}

func newConfigFromOptions(opt option.Options) (*config, error) {
	disableStreaming, err := option.GetBoolValue(opt, "disable_streaming")
	if err != nil {
		return nil, err
	}
	allowProtobufEmpty, err := option.GetBoolValue(opt, "allow_protobuf_empty")
	if err != nil {
		return nil, err
	}
	return &config{
		disableStreaming:   disableStreaming,
		allowProtobufEmpty: allowProtobufEmpty,
	}, nil
}

func ruleFunc(ctx context.Context, w check.ResponseWriter, r check.Request) error {
	log.SetFlags(0)
	log.SetPrefix("strictrpc: ")

	conf, err := newConfigFromOptions(r.Options())
	if err != nil {
		return err
	}

	for _, f := range r.FileDescriptors() {
		fd := f.ProtoreflectFileDescriptor()

		result := checkFile(conf, fd)
		if result != nil {
			var annotations []check.AddAnnotationOption
			if result.msg != "" {
				result.msg = period(result.msg)
				if DEBUG {
					log.Println("DEBUG:", result.msg)
				}
				annotations = append(annotations, check.WithMessage(result.msg))
			}
			if result.desc != nil {
				annotations = append(annotations, check.WithDescriptor(result.desc))
			} else {
				annotations = append(annotations, check.WithDescriptor(fd))
			}
			// TODO(mf): stop after accumulating N annotations? Need to see how this is displayed to
			// the user in buf.
			w.AddAnnotation(annotations...)
		}
	}
	return nil
}

// period adds a period to the end of a string if it doesn't already have one.
func period(s string) string {
	return strings.TrimSuffix(s, ".") + "."
}

func checkFile(conf *config, fileDesc protoreflect.FileDescriptor) *result {
	filename := strings.TrimSuffix(filepath.Base(fileDesc.Path()), ".proto")
	services := fileDesc.Services()
	switch n := services.Len(); {
	case n == 0:
		// No services. No problem, except if a file ends with _service.proto but does not have a
		// service. No good.
		if strings.HasSuffix(filename, "_service") {
			return newResultf(fileDesc, "file %q does not have a service, but ends with _service.proto", filename)
		}
		return nil
	case n == 1:
		// Okay. Exactly one service.
		if !strings.HasSuffix(filename, "_service") {
			return newResultf(fileDesc, "service %q must end with _service.proto", filename)
		}
	default:
		return newResultf(fileDesc, "only one service definition allowed per file, but %d were found", n)
	}
	serviceDesc := services.Get(0)

	if res := checkService(serviceDesc, conf.disableStreaming); res != nil {
		return res
	}

	return nil
}

func checkService(
	serviceDesc protoreflect.ServiceDescriptor,
	disableStreaming bool,
) *result {
	methods := serviceDesc.Methods()
	for i := range methods.Len() {
		m := methods.Get(i)
		if disableStreaming && (m.IsStreamingClient() || m.IsStreamingServer()) {
			return newResultf(serviceDesc, "method %q uses streaming, which is disabled by the disable_streaming option.", m.Name())
		}
		for _, in := range []struct {
			suffix string
			msg    protoreflect.MessageDescriptor
		}{
			{suffix: "Request", msg: m.Input()},
			{suffix: "Response", msg: m.Output()},
		} {
			if res := checkMessageSuffix(m, in.msg, in.suffix); res != nil {
				return res
			}
		}
	}
	return nil
}

func checkMessageSuffix(
	method protoreflect.MethodDescriptor,
	message protoreflect.MessageDescriptor,
	suffix string,
) *result {
	methodName := string(method.Name())
	messageName := string(message.Name())
	_, remain, ok := strings.Cut(messageName, methodName)
	if !ok || remain != suffix {
		return newResultf(message, "invalid %s message name %q, expecting %q", strings.ToLower(suffix), messageName, methodName+suffix)
	}
	return nil
}

var DEBUG = os.Getenv("DEBUG") != ""
