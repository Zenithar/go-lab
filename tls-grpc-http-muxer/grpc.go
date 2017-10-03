package main

import (
	"context"
	"net"

	"github.com/soheilhy/cmux"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	grpchello "google.golang.org/grpc/examples/helloworld/helloworld"
)

type grpcServer struct{}

func (s *grpcServer) SayHello(ctx context.Context, in *grpchello.HelloRequest) (
	*grpchello.HelloReply, error) {

	return &grpchello.HelloReply{Message: "Hello " + in.Name + " from cmux"}, nil
}

func serveGRPC(l net.Listener) {
	grpcs := grpc.NewServer()
	grpchello.RegisterGreeterServer(grpcs, &grpcServer{})
	if err := grpcs.Serve(l); err != cmux.ErrListenerClosed {
		logger.Panic("Unable to start gRPC server", zap.Error(err))
	}
}
