package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"e2b-dev/envd-connect-example/client/generated/process"
	"e2b-dev/envd-connect-example/client/generated/process/processconnect"

	"connectrpc.com/connect"
)

const (
	envdPort = 49983
)

func main() {
	sandbox, err := CreateSandbox("base", 10)
	defer KillSandbox(sandbox.SandboxID)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	fmt.Println("Created sandbox", sandbox)

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second, // Adjust timeout as needed
	}

	// Create process client with headers
	client := processconnect.NewProcessClient(
		httpClient,
		fmt.Sprintf("https://%d-%s-%s.e2b.app", envdPort, sandbox.SandboxID, sandbox.ClientID),
		connect.WithInterceptors(&headerInterceptor{envdPort: envdPort, sandboxID: sandbox.SandboxID, user: "user"}),
	)

	// Create context
	ctx := context.Background()

	// Create request
	req := connect.NewRequest(&process.StartRequest{
		Process: &process.ProcessConfig{
			Cmd: "/bin/bash",
			Args: []string{
				"-l", "-c", "echo 'Hello, World!'",
			},
		},
	})

	// Execute command and get stream
	stream, err := client.Start(ctx, req)
	if err != nil {
		log.Fatalf("Failed to start process: %v", err)
	}

	if err := HandleProcessStream(stream); err != nil {
		log.Fatalf("Stream error: %v", err)
	}
}
