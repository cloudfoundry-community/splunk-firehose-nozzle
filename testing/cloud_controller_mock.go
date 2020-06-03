package testing

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type CloudControllerMock struct {
	port   int
	server *http.Server
}

func NewCloudControllerMock(port int) *CloudControllerMock {
	return &CloudControllerMock{
		port:   port,
		server: &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: nil},
	}
}

func (c *CloudControllerMock) Start() error {
	// /v2/info
	info := []byte(`{"name":"","build":"","support":"https://support.pivotal.io","version":0,"description":"","authorization_endpoint":"","token_endpoint":"","min_cli_version":"6.23.0","min_recommended_cli_version":"6.23.0","api_version":"2.82.0","app_ssh_endpoint":"","app_ssh_host_key_fingerprint":"f3:f1:53:6d:dd:a3:94:37:0a:f8:ab:2b:3e:f7:56:27","app_ssh_oauth_client":"ssh-proxy","routing_endpoint":"","doppler_logging_endpoint":"","user":"5404b6b1-d8da-4f94-bcdf-1b78d8fed7eb"}`)
	var v map[string]interface{}
	err := json.Unmarshal(info, &v)
	if err != nil {
		return err
	}

	v["authorization_endpoint"] = fmt.Sprintf("http://localhost:%d", c.port)
	v["token_endpoint"] = fmt.Sprintf("http://localhost:%d", c.port)
	v["doppler_logging_endpoint"] = fmt.Sprintf("ws://localhost:%d", c.port)

	info, err = json.Marshal(v)
	if err != nil {
		return err
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v2/info", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, string(info))
	})

	// Authorization code endpoint
	mux.HandleFunc("/oauth/auth", func(w http.ResponseWriter, r *http.Request) {
	})

	// Access token endpoint
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		// Should return acccess token back to the user
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"access_token":"mocktoken","scope":"user","token_type":"bearer"}`))
	})

	mux.HandleFunc("/v2/read", func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		messageChan := make(chan []byte)
		go serverSentEvents(messageChan)
		for {
			// Write to the ResponseWriter
			// Server Sent Events compatible
			w.Write([]byte(<-messageChan))

			// Flush the data immediatly instead of buffering it for later.
			flusher.Flush()
		}
	})

	c.server.Handler = mux
	return c.server.ListenAndServe()
}

func (c *CloudControllerMock) Stop() error {
	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
	return c.server.Shutdown(ctx)
}

func serverSentEvents(messageChan chan []byte) {
	for {
		<-time.After(100 * time.Millisecond)
		messageChan <- []byte(`"data":"hello"`)
	}
}
