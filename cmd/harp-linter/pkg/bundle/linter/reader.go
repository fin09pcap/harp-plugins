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

package linter

import (
	"fmt"
	"io"
	"io/ioutil"

	"google.golang.org/protobuf/encoding/protojson"

	"github.com/elastic/harp/pkg/sdk/convert"
	"github.com/elastic/harp/pkg/sdk/types"

	linterv1 "github.com/elastic/harp-plugins/cmd/harp-linter/api/gen/go/harp/linter/v1"
)

// YAML a given reader in order to extract a BundlePatch sepcification
func YAML(r io.Reader) (*linterv1.RuleSet, error) {
	// Check arguments
	if types.IsNil(r) {
		return nil, fmt.Errorf("reader is nil")
	}

	// Drain the reader
	jsonReader, err := convert.YAMLtoJSON(r)
	if err != nil {
		return nil, fmt.Errorf("unable to parse input as BundlePatch: %w", err)
	}

	// Drain reader
	jsonData, err := ioutil.ReadAll(jsonReader)
	if err != nil {
		return nil, fmt.Errorf("unbale to drain all json reader content: %w", err)
	}

	// Initialize empty definition object
	def := linterv1.RuleSet{}
	def.Reset()

	// Deserialize JSON with JSONPB wrapper
	if err := protojson.Unmarshal(jsonData, &def); err != nil {
		return nil, fmt.Errorf("unable to decode spec as json: %w", err)
	}

	// Validate spec
	if err := Validate(&def); err != nil {
		return nil, fmt.Errorf("unable to validate descriptor: %w", err)
	}

	// No error
	return &def, nil
}
