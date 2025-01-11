package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/xconnio/wampproto-go"
	wampprotoutil "github.com/xconnio/wampproto-go/util"
	"github.com/xconnio/xconn-go"
	"github.com/xconnio/xconn-go/util"
)

const (
	serverURL     = "ws://0.0.0.0:8080/ws"
	realm         = "realm1"
	procedureName = "io.xconn.backend.add2"
	topicName     = "io.xconn.backend.topic1"
)

func main() {
	closers, err := util.StartServerFromConfigFile("./cmd/config.yaml")
	if err != nil {
		log.Fatalf("failed to start server: %v", err)
	}

	session, err := xconn.Connect(context.Background(), serverURL, realm)
	if err != nil {
		log.Fatalf("failed to connect session: %v", err)
	}

	_, err = session.Register(procedureName, func(ctx context.Context, invocation *xconn.Invocation) *xconn.Result {
		if len(invocation.Arguments) != 2 {
			return &xconn.Result{Err: wampproto.ErrInvalidArgument, Arguments: []any{"must be called with exactly 2 arguments"}}
		}

		firstNumber, ok := wampprotoutil.AsInt64(invocation.Arguments[0])
		if !ok {
			return &xconn.Result{Err: wampproto.ErrInvalidArgument, Arguments: []any{"arguments must be int"}}
		}

		secondNumber, ok := wampprotoutil.AsInt64(invocation.Arguments[1])
		if !ok {
			return &xconn.Result{Err: wampproto.ErrInvalidArgument, Arguments: []any{"arguments must be int"}}
		}

		return &xconn.Result{Arguments: []any{firstNumber + secondNumber}}
	}, nil)
	if err != nil {
		log.Fatalf("failed to register procedure: %v", err)
	}

	_, err = session.Subscribe(topicName, func(event *xconn.Event) {
		fmt.Printf("event received: args: %v, kwargs: %v\n", event.Arguments, event.KwArguments)
	}, nil)
	if err != nil {
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