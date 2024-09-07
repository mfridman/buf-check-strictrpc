package strictrpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/bufbuild/bufplugin-go/check"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// TODO(mf): make allowStreaming a plugin option
	streamingAllowed = false
)

var DEBUG = os.Getenv("DEBUG") != ""

var Spec = &check.Spec{
	Rules: []*check.RuleSpec{
		{
			ID:      "STRICT_RPC",
			Purpose: "An opinionated way to structure RPCs.",
			Type:    check.RuleTypeLint,
			Handler: check.RuleHandlerFunc(ruleFunc),
			Default: true,

			// TODO(mf): What should this be and when would I use it?
			CategoryIDs:    nil,
			ReplacementIDs: nil,
			Deprecated:     false,
		},
	},
}

// TODO(mf): is there an opportunity for the [check] library to make it easier to use
// check.ResponseWriter and accumulating results without having to pass down the writer to all the
// places that need it?

type result struct {
	msg string
	fd  protoreflect.Descriptor
}

func newResult(fd protoreflect.Descriptor, msg string, args ...any) *result {
	return &result{
		msg: fmt.Sprintf(msg, args...),
		fd:  fd,
	}
}

func ruleFunc(ctx context.Context, w check.ResponseWriter, r check.Request) error {
	log.SetFlags(0)
	log.SetPrefix("strictrpc: ")

	for _, f := range r.Files() {
		fd := f.FileDescriptor()

		res := checkFile(fd)
		if res != nil {
			var annotations []check.AddAnnotationOption
			if res.msg != "" {
				annotations = append(annotations, check.WithMessage(res.msg))
				if DEBUG {
					log.Println("DEBUG:", res.msg)
				}
			}
			if res.fd != nil {
				annotations = append(annotations, check.WithDescriptor(res.fd))
			} else {
				annotations = append(annotations, check.WithDescriptor(fd))
			}
			// TODO(mf): stop after accumulating N annotations? Need to see how this is displayed to the
			// user in buf.
			w.AddAnnotation(annotations...)
		}
	}
	return nil
}

func checkFile(fd protoreflect.FileDescriptor) *result {
	services := fd.Services()
	switch n := services.Len(); {
	case n == 0:
		// No services. No problem.
		return nil
	case n == 1:
		// Okay. Exactly one service.
		filename := strings.TrimSuffix(filepath.Base(fd.Path()), ".proto")
		if !strings.HasSuffix(filename, "_service") {
			return newResult(fd, "service %q must end with _service.proto", filename)
		}
	default:
		return newResult(fd, fmt.Sprintf("only one service definition allowed per file, but %d were found", n))
	}
	// TODO:
	//  - iterate over methods, make sure they have Request/Response suffixes (duplicate)
	//  - iterate over all the messages, making sure there are exactly 2 or 3 messages per method:
	//    - request
	//    - response
	//    - optional, allow 1 ErrorDetails per method
	//    - as iterating, ensure they are in the correct order (request, response, error details), and
	//      are in the same order as the method definitions within the service

	if res := checkService(services.Get(0), streamingAllowed); res != nil {
		return res
	}
	// TODO(mf): this is inefficient, because we're iterating over the methods twice
	_ = fd.Messages()

	return nil
}

func checkService(
	sd protoreflect.ServiceDescriptor,
	allowStreaming bool,
) *result {
	methods := sd.Methods()
	for i := range methods.Len() {
		m := methods.Get(i)
		if !allowStreaming && (m.IsStreamingClient() || m.IsStreamingServer()) {
			return newResult(sd, "method %q is streaming, but streaming is not allowed", m.Name())
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
		return newResult(message, "invalid %s message name %q, expecting %q",
			strings.ToLower(suffix), messageName, methodName+suffix,
		)
	}
	return nil
}
