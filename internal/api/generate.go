// Copyright (c) Hatchet Technologies Inc.
// SPDX-License-Identifier: MIT

//go:build generate

package managementclient

//go:generate npx --yes swagger-cli bundle ../../../hatchet-control-plane/api-contracts/openapi/openapi.yaml --outfile openapi.yaml --type yaml
//go:generate go run -modfile ../../tools/go.mod github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config oapi-codegen.yaml --exclude-operation-ids sso:list openapi.yaml
