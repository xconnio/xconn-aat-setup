package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/xconnio/nxt/util"
	"github.com/xconnio/wampproto-go"
	wampprotoutil "github.com/xconnio/wampproto-go/util"
	"github.com/xconnio/xconn-go"
)

const (
	serverURL     = "ws://0.0.0.0:8080/ws"
	realm         = "realm1"
	procedureName = "io.xconn.backend.add2"
	topicName     = "io.xconn.backend.topic1"
)

func main() {
	configFile := flag.String("config", "./cmd/config.yaml", "Path to the configuration file")
	flag.Parse()

	closers, err := util.StartServerFromConfigFile(*configFile)
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	session, err := xconn.ConnectAnonymous(context.Background(), serverURL, realm)
	if err != nil {
		log.Fatalf("failed to connect session: %v", err)
	}

	registerResponse := session.Register(procedureName,
		func(ctx context.Context, invocation *xconn.Invocation) *xconn.Result {
			if len(invocation.Arguments) != 2 {
				return &xconn.Result{Err: wampproto.ErrInvalidArgument, Arguments: []any{"must be called with exactly 2 arguments"}}
			}

			firstNumber, ok := wampprotoutil.AsUInt64(invocation.Arguments[0])
			if !ok {
				return &xconn.Result{Err: wampproto.ErrInvalidArgument, Arguments: []any{"arguments must be int"}}
			}

			secondNumber, ok := wampprotoutil.AsUInt64(invocation.Arguments[1])
			if !ok {
				return &xconn.Result{Err: wampproto.ErrInvalidArgument, Arguments: []any{"arguments must be int"}}
			}

			return xconn.NewInvocationResult(firstNumber + secondNumber)
		}).Do()
	if registerResponse.Err != nil {
		log.Fatalf("failed to register procedure: %v", err)
	}

	subscribeResponse := session.Subscribe(topicName, func(event *xconn.Event) {
		fmt.Printf("event received: args: %v, kwargs: %v\n", event.Arguments, event.KwArguments)
	}).Do()
	if subscribeResponse.Err != nil {
		log.Fatalf("failed to subscribe to topic: %v", err)
	}

	// Close server if SIGINT (CTRL-c) received.
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, os.Interrupt)
	<-closeChan

	for _, closer := range closers {
		_ = closer.Close()
	}
}
