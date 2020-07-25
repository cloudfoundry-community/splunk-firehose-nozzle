package loggregator_test

import (
	"crypto/tls"
	"net"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"code.cloudfoundry.org/go-loggregator/v8/rpc/loggregator_v2"
	"code.cloudfoundry.org/tlsconfig"
)

type testIngressServer struct {
	receivers    chan loggregator_v2.Ingress_BatchSenderServer
	sendReceiver chan *loggregator_v2.EnvelopeBatch
	addr         string
	tlsConfig    *tls.Config
	grpcServer   *grpc.Server
	grpc.Stream
}

func newTestIngressServer(serverCert, serverKey, caCert string) (*testIngressServer, error) {
	tlsConfig, err := tlsconfig.Build(
		tlsconfig.WithInternalServiceDefaults(),
		tlsconfig.WithIdentityFromFile(serverCert, serverKey),
	).Server(
		tlsconfig.WithClientAuthenticationFromFile(caCert),
	)

	if err != nil {
		return nil, err
	}

	return &testIngressServer{
		tlsConfig:    tlsConfig,
		receivers:    make(chan loggregator_v2.Ingress_BatchSenderServer),
		sendReceiver: make(chan *loggregator_v2.EnvelopeBatch, 100),
		addr:         "localhost:0",
	}, nil
}

func (*testIngressServer) Sender(srv loggregator_v2.Ingress_SenderServer) error {
	return nil
}

func (t *testIngressServer) BatchSender(srv loggregator_v2.Ingress_BatchSenderServer) error {
	t.receivers <- srv

	<-srv.Context().Done()

	return nil
}

func (t *testIngressServer) Send(_ context.Context, b *loggregator_v2.EnvelopeBatch) (*loggregator_v2.SendResponse, error) {
	t.sendReceiver <- b
	return &loggregator_v2.SendResponse{}, nil
}

func (t *testIngressServer) start() error {
	listener, err := net.Listen("tcp4", t.addr)
	if err != nil {
		return err
	}
	t.addr = listener.Addr().String()

	var opts []grpc.ServerOption
	if t.tlsConfig != nil {
		opts = append(opts, grpc.Creds(credentials.NewTLS(t.tlsConfig)))
	}
	t.grpcServer = grpc.NewServer(opts...)

	loggregator_v2.RegisterIngressServer(t.grpcServer, t)

	go t.grpcServer.Serve(listener)

	return nil
}

func (t *testIngressServer) stop() {
	t.grpcServer.Stop()
}
