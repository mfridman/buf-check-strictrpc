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

var DEBUG = os.Getenv("DEBUG") != ""

var Rule = &check.RuleSpec{
	ID:      "STRICT_RPC",
	Purpose: "Opinionated way to structure RPCs.",
	Type:    check.RuleTypeLint,
	Handler: check.RuleHandlerFunc(ruleFunc),
	// ?? no sure what these are for
	IsDefault:      true,
	Categories:     nil,
	Deprecated:     false,
	ReplacementIDs: nil,
}

func ruleFunc(
	ctx context.Context,
	w check.ResponseWriter,
	r check.Request,
) error {
	log.SetFlags(0)
	log.SetPrefix("strictrpc: ")

	for _, f := range r.Files() {
		fd := f.FileDescriptor()
		// TODO(mf): should this return a []string or string?
		msg := checkFile(fd)
		if msg == "" {
			continue
		} else if DEBUG {
			log.Println(msg)
		}
		// TODO(mf): stop after accumulating N annotations? Need to see how this is displayed to
		// the user in buf.
		w.AddAnnotation(check.WithDescriptor(fd), check.WithMessage(msg))
	}
	return nil
}

func checkFile(fd protoreflect.FileDescriptor) string {
	services := fd.Services()
	switch n := services.Len(); {
	case n == 0:
		// No services. No problem.
		return ""
	case n == 1:
		// Okay. Exactly one service.
		filename := strings.TrimSuffix(filepath.Base(fd.Path()), ".proto")
		if !strings.HasSuffix(filename, "_service") {
			return fmt.Sprintf("service %q must end with _service.proto", filename)
		}
	default:
		return fmt.Sprintf("only one service definition allowed per file, but %d were found", n)
	}
	// TODO:
	//  - iterate over methods, make sure they have Request/Response suffixes (duplicate)
	//  - iterate over all the messages, making sure there are exactly 2 or 3 messages per method:
	//    - request
	//    - response
	//    - optional, allow 1 ErrorDetails per method
	//    - as iterating, ensure they are in the correct order (request, response, error details), and
	//      are in the same order as the method definitions within the service

	// TODO(mf): make allowStreaming a plugin option
	if msg := checkService(services.Get(0), true); msg != "" {
		return msg
	}
	// TODO(mf): this is inefficient, because we're iterating over the methods twice
	_ = fd.Messages()

	return ""
}

func checkService(
	sd protoreflect.ServiceDescriptor,
	allowStreaming bool,
) string {
	methods := sd.Methods()
	for i := range methods.Len() {
		m := methods.Get(i)
		if !allowStreaming && (m.IsStreamingClient() || m.IsStreamingServer()) {
			return fmt.Sprintf("method %q is streaming, but streaming is not allowed", m.Name())
		}
		for _, in := range []struct {
			suffix string
			msg    protoreflect.MessageDescriptor
		}{
			{suffix: "Request", msg: m.Input()},
			{suffix: "Response", msg: m.Output()},
		} {
			if msg := checkMessageSuffix(m, in.msg, in.suffix); msg != "" {
				return msg
			}
		}
	}
	return ""
}

func checkMessageSuffix(
	method protoreflect.MethodDescriptor,
	message protoreflect.MessageDescriptor,
	suffix string,
) string {
	methodName := string(method.Name())
	messageName := string(message.Name())
	_, remain, ok := strings.Cut(messageName, methodName)
	if !ok || remain != suffix {
		return fmt.Sprintf("invalid %s message name %q, expecting %q",
			strings.ToLower(suffix), messageName, methodName+suffix,
		)
	}
	return ""
}
