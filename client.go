package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
)

type headerInterceptor struct {
	envdPort  int
	sandboxID string
	user      string
}

func SetSandboxHeader(header http.Header, envdPort int, sandboxID string) {
	host := fmt.Sprintf("%d-%s-00000000.e2b.app", envdPort, sandboxID)
	header.Set("Host", host)
}

func SetUserHeader(header http.Header, user string) {
	userString := fmt.Sprintf("%s:", user)
	userBase64 := base64.StdEncoding.EncodeToString([]byte(userString))
	basic := fmt.Sprintf("Basic %s", userBase64)
	header.Set("Authorization", basic)
}

func (i *headerInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Set the sandbox header
		SetSandboxHeader(req.Header(), i.envdPort, i.sandboxID)

		// Set the user header for authentication
		SetUserHeader(req.Header(), i.user)

		return next(ctx, req)
	}
}

func (i *headerInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		conn := next(ctx, spec)

		// Set the sandbox header
		SetSandboxHeader(conn.RequestHeader(), i.envdPort, i.sandboxID)

		// Set the user header for authentication
		SetUserHeader(conn.RequestHeader(), i.user)

		return conn
	}
}

func (i *headerInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
