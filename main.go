package main

import (
	"context"
	"errors"
	"flag"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"

	gw "binwang.me/rss/grpc-gateway/gen/go" // Update
)

var (
	// command-line options:
	// gRPC server endpoint
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "grpc.rssbrain.com:443", "gRPC server endpoint")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	// opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	creds := credentials.NewClientTLSFromCert(nil, "") // 'nil' means use system roots
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	err := errors.Join(
		gw.RegisterArticleAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts),
		gw.RegisterSourceAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts),
		gw.RegisterFolderAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts),
		gw.RegisterUserAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts),
		gw.RegisterMoreLikeThisAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts),
		gw.RegisterSystemAPIHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts),
	)
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(":8881", mux)
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		grpclog.Fatal(err)
	}
}
