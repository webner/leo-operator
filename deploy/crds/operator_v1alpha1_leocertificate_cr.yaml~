#!/bin/bash

set -eo pipefail

export GO111MODULE=on
go build -o operator github.com/webner/leo-operator/cmd/manager
export WATCH_NAMESPACE=
#export OPERATOR_NAME=leo-operator

./operator


