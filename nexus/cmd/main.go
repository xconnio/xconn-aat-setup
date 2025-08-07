package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/gammazero/nexus/v3/client"
	"github.com/gammazero/nexus/v3/router"
	"github.com/gammazero/nexus/v3/router/auth"
	"github.com/gammazero/nexus/v3/wamp"
	"golang.org/x/crypto/pbkdf2"
)

const (
	address       = "0.0.0.0:8082"
	realm         = "realm1"
	procedureName = "io.xconn.backend.add2"
	topicName     = "io.xconn.backend.topic1"
)

type StaticKeyStore struct {
	secretByAuthID map[string][]byte
}

func NewStaticKeyStore(users map[string][]byte) *StaticKeyStore {
	return &StaticKeyStore{
		secretByAuthID: users,
	}
}

func (ks *StaticKeyStore) AuthKey(authid, authmethod string) ([]byte, error) {
	key, ok := ks.secretByAuthID[authid]
	if !ok {
		return nil, fmt.Errorf("authid %q not found", authid)
	}

	if authid == "wamp-cra-salt-user" {
		salt, keylen, iterations := ks.PasswordInfo(authid)
		dk := pbkdf2.Key(key, []byte(salt), iterations, keylen, sha256.New)
		return []byte(base64.StdEncoding.EncodeToString(dk)), nil
	}

	return key, nil
}

func (ks *StaticKeyStore) PasswordInfo(authid string) (string, int, int) {
	if authid == "wamp-cra-salt-user" {
		return "cra-salt", 32, 1000
	}
	return "", 0, 0
}

func (ks *StaticKeyStore) AuthRole(authid string) (string, error) {
	return "anonymous", nil
}

func (ks *StaticKeyStore) Provider() string {
	return "static"
}

func main() {
	publicKey, err := hex.DecodeString("ddc2838ede4304c1082c503f0af4f0c5ea7dea9fe436127643364c0670b69b08")
	if err != nil {
		log.Fatal(err)
	}

	keyStore := NewStaticKeyStore(map[string][]byte{
		"cryptosign-user":    publicKey,
		"wamp-cra-user":      []byte("cra-secret"),
		"wamp-cra-salt-user": []byte("cra-salt-secret"),
		"ticket-user":        []byte("ticket-pass"),
	})
	routerCfg := &router.Config{
		RealmConfigs: []*router.RealmConfig{
			{
				URI: realm,
				Authenticators: []auth.Authenticator{
					auth.NewCryptoSignAuthenticator(keyStore, 0),
					auth.NewTicketAuthenticator(keyStore, 0),
					auth.NewCRAuthenticator(keyStore, 0),
				},
				AnonymousAuth: true,
			},
		},
	}
	r, err := router.NewRouter(routerCfg, nil)
	if err != nil {
		log.Fatalln(err)
	}

	session, err := client.ConnectLocal(r, client.Config{Realm: realm})
	if err != nil {
		log.Fatalln(err)
	}

	err = session.Register(procedureName, func(ctx context.Context, invocation *wamp.Invocation) client.InvokeResult {
		if len(invocation.Arguments) != 2 {
			return client.InvokeResult{Err: wamp.ErrInvalidArgument, Args: []any{"must be called with exactly 2 arguments"}}
		}

		firstNumber, ok := wamp.AsInt64(invocation.Arguments[0])
		if !ok {
			return client.InvokeResult{Err: wamp.ErrInvalidArgument, Args: []any{"arguments must be int"}}
		}

		secondNumber, ok := wamp.AsInt64(invocation.Arguments[1])
		if !ok {
			return client.InvokeResult{Err: wamp.ErrInvalidArgument, Args: []any{"arguments must be int"}}
		}

		return client.InvokeResult{Args: []any{firstNumber + secondNumber}}
	}, nil)
	if err != nil {
		log.Fatalln(err)
	}

	err = session.Subscribe(topicName, func(event *wamp.Event) {
		fmt.Printf("event received: args: %v, kwargs: %v\n", event.Arguments, event.ArgumentsKw)
	}, nil)
	if err != nil {
		log.Fatalln(err)
	}

	server := router.NewWebsocketServer(r)
	closer, err := server.ListenAndServe(address)
	if err != nil {
		log.Fatalln(err)
	}

	defer closer.Close()

	// Close server if SIGINT (CTRL-c) received.
	closeChan := make(chan os.Signal, 1)
	signal.Notify(closeChan, os.Interrupt)
	<-closeChan
}
