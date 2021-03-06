// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package grpc

import (
	"context"
	"crypto/tls"
	"github.com/elastic/harp-plugins/cmd/harp-server/internal/config"
	"github.com/elastic/harp-plugins/cmd/harp-server/internal/dispatchers/grpc/server"
	"github.com/elastic/harp-plugins/cmd/harp-server/pkg/server/manager"
	"github.com/elastic/harp-plugins/cmd/harp-server/pkg/server/storage/backends/container"
	"github.com/elastic/harp/api/gen/go/harp/bundle/v1"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/tlsconfig"
	"github.com/elastic/harp/pkg/vault/path"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// Injectors from wire.go:

func setup(ctx context.Context, cfg *config.Configuration) (*grpc.Server, error) {
	backend, err := backendManager(ctx, cfg)
	if err != nil {
		return nil, err
	}
	server, err := grpcServer(ctx, cfg, backend)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// wire.go:

func backendManager(ctx context.Context, cfg *config.Configuration) (manager.Backend, error) {

	bm := manager.Default()

	for _, b := range cfg.Backends {

		if err := bm.Register(ctx, path.SanitizePath(b.NS), b.URL); err != nil {
			return nil, err
		}
	}

	return bm, nil
}

func grpcServer(ctx context.Context, cfg *config.Configuration, bm manager.Backend) (*grpc.Server, error) {
	container.SetKeyring(cfg.Keyring)

	sopts := []grpc.ServerOption{}

	if cfg.GRPC.UseTLS {

		clientAuth := tls.VerifyClientCertIfGiven
		if cfg.GRPC.TLS.ClientAuthenticationRequired {
			clientAuth = tls.RequireAndVerifyClientCert
		}

		tlsConfig, err := tlsconfig.Server(&tlsconfig.Options{
			KeyFile:    cfg.GRPC.TLS.PrivateKeyPath,
			CertFile:   cfg.GRPC.TLS.CertificatePath,
			CAFile:     cfg.GRPC.TLS.CACertificatePath,
			ClientAuth: clientAuth,
		})
		if err != nil {
			log.For(ctx).Error("Unable to build TLS configuration from settings", zap.Error(err))
			return nil, err
		}

		sopts = append(sopts, grpc.Creds(credentials.NewTLS(tlsConfig)))
	} else {
		log.For(ctx).Info("No transport encryption enabled for gRPC server")
	}
	grpcServer2 := grpc.NewServer(sopts...)
	bundlev1.RegisterBundleAPIServer(grpcServer2, server.Bundle(bm))
	reflection.Register(grpcServer2)

	return grpcServer2, nil
}
