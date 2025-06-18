package main

import (
	"context"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"e2b.dev/groq/client/generated/process"
	"e2b.dev/groq/client/generated/process/processconnect"
)

func main() {
	// Create HTTP client
	httpClient := &http.Client{}

	// Create process client
	client := processconnect.NewProcessClient(
		httpClient,
		"https://49983-4ibr3mm98t7io3xw86wwiw-1df5cfa7.e2b.dev",
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
