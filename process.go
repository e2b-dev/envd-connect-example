package main

import (
	"log"

	"e2b-dev/envd-connect-example/client/generated/process"

	"connectrpc.com/connect"
)

func HandleProcessStream(stream *connect.ServerStreamForClient[process.StartResponse]) error {
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

	return stream.Err()
}
