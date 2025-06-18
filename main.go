package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"e2b.dev/groq/client/generated/process"
	"e2b.dev/groq/client/generated/process/processconnect"
)

const (
	envdPort = 49983
)

func SetSandboxHeader(header http.Header, sandboxID string) {
	host := fmt.Sprintf("%d-%s-00000000.e2b.app", envdPort, sandboxID)
	header.Set("Host", host)
}

func SetUserHeader(header http.Header, user string) {
	userString := fmt.Sprintf("%s:", user)
	userBase64 := base64.StdEncoding.EncodeToString([]byte(userString))
	basic := fmt.Sprintf("Basic %s", userBase64)
	header.Set("Authorization", basic)
}

func main() {
	sandbox, err := CreateSandbox("base", 10)
	if err != nil {
		log.Fatalf("Failed to create sandbox: %v", err)
	}

	fmt.Println(sandbox)

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second, // Adjust timeout as needed
	}

	// Create process client with headers
	client := processconnect.NewProcessClient(
		httpClient,
		fmt.Sprintf("https://%d-%s-%s.e2b.app", envdPort, sandbox.SandboxID, sandbox.ClientID),
		connect.WithInterceptors(
			connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
				return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
					// Set the sandbox header
					SetSandboxHeader(req.Header(), sandbox.SandboxID)

					// Set the user header for authentication
					SetUserHeader(req.Header(), "user")

					return next(ctx, req)
				}
			}),
		),
	)

	// Create context
	ctx := context.Background()

	// Create request
	req := connect.NewRequest(&process.StartRequest{
		Process: &process.ProcessConfig{
			Cmd: "/bin/bash",
			Args: []string{
				"-l", "-c", "npx create-next-app@latest nextapp --yes",
			},
		},
	})

	// Execute command and get stream
	stream, err := client.Start(ctx, req)
	if err != nil {
		log.Fatalf("Failed to start process: %v", err)
	}
	defer stream.Close()

	// Read stream responses
	for stream.Receive() {
		msg := stream.Msg()
		if msg.Event != nil {
			switch event := msg.Event.Event.(type) {
			case *process.ProcessEvent_Start:
				log.Printf("Process started with PID: %d", event.Start.Pid)
			case *process.ProcessEvent_Data:
				if data := event.Data; data != nil {
					switch output := data.Output.(type) {
					case *process.ProcessEvent_DataEvent_Stdout:
						log.Printf("stdout: %s", string(output.Stdout))
					case *process.ProcessEvent_DataEvent_Stderr:
						log.Printf("stderr: %s", string(output.Stderr))
					}
				}
			case *process.ProcessEvent_End:
				log.Printf("Process ended with exit code: %d, status: %s", event.End.ExitCode, event.End.Status)
				if event.End.Error != nil {
					log.Printf("Error: %s", *event.End.Error)
				}
			}
		}
	}

	if err := stream.Err(); err != nil {
		log.Fatalf("Stream error: %v", err)
	}
}
