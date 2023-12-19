# Copyright (c) 2023-2024 Red Hat, Inc.
# This program and the accompanying materials are made
# available under the terms of the Eclipse Public License 2.0
# which is available at https://www.eclipse.org/legal/epl-2.0/
#
# SPDX-License-Identifier: EPL-2.0
#
# Contributors:
#   Red Hat, Inc. - initial API and implementation
#

#@follow_tag(registry.redhat.io/rhel9/go-toolset:latest)
FROM registry.access.redhat.com/ubi9/go-toolset:1.20.10-6 as builder
# hadolint ignore=DL3002
USER 0
ENV GOPATH=/go/
# update RPMs
RUN dnf -q -y update

# Upstream sources
# Downstream comment
# ENV EXTERNAL_SOURCE=.
# ENV CONTAINER_SOURCE=/opt/app-root/src
# WORKDIR /workspace
#/ Downstream comment

# Downstream sources
# Downstream uncomment
ENV EXTERNAL_SOURCE=$REMOTE_SOURCES/upstream1/app/distgit/containers/rhdh-operator
ENV CONTAINER_SOURCE=$REMOTE_SOURCES_DIR
WORKDIR $CONTAINER_SOURCE/
#/ Downstream uncomment

COPY $EXTERNAL_SOURCE ./

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
# Downstream comment
# RUN go mod download
#/ Downstream comment

# Downstream uncomment
COPY $REMOTE_SOURCES/upstream1/cachito.env ./
RUN source ./cachito.env && rm -f ./cachito.env && mkdir -p /workspace
#/ Downstream uncomment

# Build
# hadolint ignore=SC3010
RUN export ARCH="$(uname -m)" && if [[ ${ARCH} == "x86_64" ]]; then export ARCH="amd64"; elif [[ ${ARCH} == "aarch64" ]]; then export ARCH="arm64"; fi && \
    CGO_ENABLED=1 GOOS=linux GOARCH=${ARCH} go build -a -o manager main.go

# NOTE: ubi-micro will not be FIPS compliant, if openssl is not installed
#@follow_tag(registry.redhat.io/ubi9/ubi-micro:latest)
FROM registry.access.redhat.com/ubi9/ubi-micro:9.3-9
# Upstream sources
# Downstream comment
# ENV CONTAINER_SOURCE=/workspace
#/ Downstream comment

# Downstream sources
# Downstream uncomment
ENV CONTAINER_SOURCE=$REMOTE_SOURCES_DIR
#/ Downstream uncomment

ENV HOME=/ \
    USER_NAME=backstage \
    USER_UID=1001

RUN echo "${USER_NAME}:x:${USER_UID}:0:${USER_NAME} user:${HOME}:/sbin/nologin" >> /etc/passwd

# Copy manager binary
COPY --from=builder $CONTAINER_SOURCE/manager .

USER ${USER_UID}

WORKDIR ${HOME}

ENTRYPOINT ["/manager"]

# append Brew metadata here
