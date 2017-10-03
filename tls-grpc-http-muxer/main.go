package main

import (
	"crypto/tls"
	"net"
	"strings"

	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
)

func mustLoadX509KeyPair(certFile, keyFile string) tls.Certificate {
	if cert, err := tls.LoadX509KeyPair(certFile, keyFile); err != nil {
		panic(err)
	} else {
		return cert
	}
}

func main() {
	logger, _ = zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any

	// Initialize core listener
	port, err := net.Listen("tcp", ":5555")
	if err != nil {
		logger.Panic("Unable to initialize main listener", zap.Error(err), zap.String("address", ":5555"))
	}

	// Prepare TLS Config
	tlsConfig := &tls.Config{
		ServerName: "localhost:5555",
		NextProtos: []string{},
		MinVersion: tls.VersionTLS12,
		Certificates: []tls.Certificate{
			mustLoadX509KeyPair("./certs/ecdsacert.pem", "./certs/ecdsakey.pem"),
			mustLoadX509KeyPair("./certs/rsacert.pem", "./certs/rsakey.pem"),
		},
		PreferServerCipherSuites: true,
		CurvePreferences: []tls.CurveID{
			tls.CurveP256,
			tls.X25519, // Go 1.8 only
		},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, // Go 1.8 only
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,   // Go 1.8 only
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	// Wrap in a TLS listener
	tlsL := tls.NewListener(port, tlsConfig)

	// Initialize a muxer
	tcpMux := cmux.New(tlsL)

	// Protocol matchers
	grpcL := tcpMux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	httpL := tcpMux.Match(cmux.HTTP2(), cmux.HTTP1Fast())

	// Start all servers
	go serveGRPC(grpcL)
	go serveHTTP(httpL)

	// Start muxer
	logger.Info("Listening muxed services", zap.String("address", ":5555"))
	if err := tcpMux.Serve(); !strings.Contains(err.Error(), "use of closed network connection") {
		logger.Panic("Muxer ended with error", zap.Error(err))
	}
}
