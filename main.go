package main

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"

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

func serveSwagger(mux *http.ServeMux) {
	swaggerFile, err := os.ReadFile("gen/go/grpc-api.swagger.json")
	if err != nil {
		grpclog.Fatal(err)
	}

	mux.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(swaggerFile)
	})
}

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Create a new ServeMux for HTTP handlers
	httpMux := http.NewServeMux()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	grpcGatewayMux := runtime.NewServeMux()
	creds := credentials.NewClientTLSFromCert(nil, "") // 'nil' means use system roots
	opts := []grpc.DialOption{grpc.WithTransportCredentials(creds)}
	err := errors.Join(
		gw.RegisterArticleAPIHandlerFromEndpoint(ctx, grpcGatewayMux, *grpcServerEndpoint, opts),
		gw.RegisterSourceAPIHandlerFromEndpoint(ctx, grpcGatewayMux, *grpcServerEndpoint, opts),
		gw.RegisterFolderAPIHandlerFromEndpoint(ctx, grpcGatewayMux, *grpcServerEndpoint, opts),
		gw.RegisterUserAPIHandlerFromEndpoint(ctx, grpcGatewayMux, *grpcServerEndpoint, opts),
		gw.RegisterMoreLikeThisAPIHandlerFromEndpoint(ctx, grpcGatewayMux, *grpcServerEndpoint, opts),
		gw.RegisterSystemAPIHandlerFromEndpoint(ctx, grpcGatewayMux, *grpcServerEndpoint, opts),
	)
	if err != nil {
		return err
	}

	// Serve swagger JSON file
	serveSwagger(httpMux)

	// Handle gRPC-Gateway endpoints
	httpMux.Handle("/", grpcGatewayMux)

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(":8881", httpMux)
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		grpclog.Fatal(err)
	}
}
