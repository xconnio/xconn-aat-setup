package nxt_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	wampprotobuf "github.com/xconnio/wampproto-protobuf/go"
	wampprotocapnp "github.com/xconnio/wampproto-serializer-capnproto/go"

	"github.com/xconnio/wampproto-go/auth"
	"github.com/xconnio/wampproto-go/util"
	"github.com/xconnio/xconn-go"
)

const (
	xconnURL     = "ws://localhost:8080/ws"
	realm        = "realm1"
	procedureAdd = "io.xconn.backend.add2"

	ticketUserAuthID = "ticket-user"
	ticket           = "ticket-pass"

	craUserAuthID = "wamp-cra-user"
	secret        = "cra-secret"

	cryptosignUserAuthID = "cryptosign-user"
	privateKey           = "150085398329d255ad69e82bf47ced397bcec5b8fbeecd28a80edbbd85b49081"
)

func connectSession(t *testing.T, authenticator auth.ClientAuthenticator, serializer xconn.SerializerSpec,
	url string) *xconn.Session {
	client := xconn.Client{
		Authenticator:  authenticator,
		SerializerSpec: serializer,
	}

	session, err := client.Connect(context.Background(), url, realm)
	require.NoError(t, err)

	return session
}

func testCall(t *testing.T, authenticator auth.ClientAuthenticator, serializer xconn.SerializerSpec, url string) {
	session := connectSession(t, authenticator, serializer, url)

	callResponse := session.Call(procedureAdd).Args(2, 2).Do()
	require.NoError(t, callResponse.Err)

	sumResult, ok := util.AsUInt64(callResponse.Args.Raw()[0])
	require.True(t, ok)
	require.Equal(t, 4, int(sumResult))
}

func testRPC(t *testing.T, authenticator auth.ClientAuthenticator, serializer xconn.SerializerSpec, url string) {
	session := connectSession(t, authenticator, serializer, url)

	registerResponse := session.Register("io.xconn.test",
		func(ctx context.Context, invocation *xconn.Invocation) *xconn.InvocationResult {
			return &xconn.InvocationResult{Args: invocation.Args(), Kwargs: invocation.Kwargs()}
		}).Do()
	require.NoError(t, registerResponse.Err)

	args := []any{"Hello", "wamp"}
	callResponse := session.Call("io.xconn.test").Args(args...).Do()
	require.NoError(t, callResponse.Err)
	require.Equal(t, args, callResponse.Args.Raw())

	err := registerResponse.Unregister()
	require.NoError(t, err)
}

func testPubSub(t *testing.T, authenticator auth.ClientAuthenticator, serializer xconn.SerializerSpec, url string) {
	session := connectSession(t, authenticator, serializer, url)

	args := []any{"Hello", "wamp"}
	subscribeResponse := session.Subscribe("io.xconn.test", func(event *xconn.Event) {
		require.Equal(t, args, event.Args())
	}).Do()
	require.NoError(t, subscribeResponse.Err)

	publishResponse := session.Publish("io.xconn.test").Args(args...).Option("acknowledge", true).Do()
	require.NoError(t, publishResponse.Err)

	err := subscribeResponse.Unsubscribe()
	require.NoError(t, err)
}

func TestStaticSerializers(t *testing.T) {
	cryptosignAuthenticator, err := auth.NewCryptoSignAuthenticator(cryptosignUserAuthID, privateKey, map[string]any{})
	require.NoError(t, err)

	authenticators := map[string]auth.ClientAuthenticator{
		"AnonymousAuth":     auth.NewAnonymousAuthenticator("", map[string]any{}),
		"TicketAuth":        auth.NewTicketAuthenticator(ticketUserAuthID, ticket, map[string]any{}),
		"WAMPCRAAuth":       auth.NewWAMPCRAAuthenticator(craUserAuthID, secret, map[string]any{}),
		"WAMPCRAAuthSalted": auth.NewWAMPCRAAuthenticator("wamp-cra-salt-user", "cra-salt-secret", map[string]any{}),
		"CryptosignAuth":    cryptosignAuthenticator,
	}

	ProtobufSerializerSpec := xconn.NewSerializerSpec(wampprotobuf.ProtobufSplitSubProtocol,
		&wampprotobuf.ProtobufSerializer{}, xconn.SerializerID(wampprotobuf.ProtobufSerializerID))
	CapnprotoSerializerSpec := xconn.NewSerializerSpec(wampprotocapnp.CapnprotoSplitSubProtocol,
		&wampprotocapnp.CapnprotoSerializer{}, xconn.SerializerID(wampprotocapnp.CapnprotoSplitSerializerID))
	serializers := map[string]xconn.SerializerSpec{
		"Protobuf":  ProtobufSerializerSpec,
		"Capnproto": CapnprotoSerializerSpec,
	}

	for authName, authenticator := range authenticators {
		for serializerName, serializer := range serializers {
			t.Run(authName+"And"+serializerName, func(t *testing.T) {
				testCall(t, authenticator, serializer, xconnURL)
				testRPC(t, authenticator, serializer, xconnURL)
				testPubSub(t, authenticator, serializer, xconnURL)
			})
		}
	}
}
