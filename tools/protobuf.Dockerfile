# Copyright (c) 2026 Ant Group Corporation.
#
# SPDX-License-Identifier: Apache-2.0

ARG GO_BUILD_IMAGE=golang:1.25.5-bookworm
FROM ${GO_BUILD_IMAGE}

ARG PROTOC_VERSION=3.21.12
ARG PROTOC_GEN_GO_VERSION=v1.36.11
ARG PROTOC_GEN_GO_GRPC_VERSION=v1.6.2
ARG GO_FIX_ACRONYM_VERSION=v0.3.0

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        make \
        protobuf-compiler && \
    rm -rf /var/lib/apt/lists/* && \
    test "$(protoc --version)" = "libprotoc ${PROTOC_VERSION}"

RUN GOBIN=/usr/local/bin go install \
        "google.golang.org/protobuf/cmd/protoc-gen-go@${PROTOC_GEN_GO_VERSION}" && \
    GOBIN=/usr/local/bin go install \
        "google.golang.org/grpc/cmd/protoc-gen-go-grpc@${PROTOC_GEN_GO_GRPC_VERSION}" && \
    GOBIN=/usr/local/bin go install \
        "github.com/containerd/protobuild/cmd/go-fix-acronym@${GO_FIX_ACRONYM_VERSION}"

WORKDIR /workspace

ENTRYPOINT ["make", "protos-local"]
