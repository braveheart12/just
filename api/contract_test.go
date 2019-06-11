///
// Copyright 2019 Insolar Technologies GmbH
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
///

package api

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestUnmarshalUpload(t *testing.T) {
	jsonResponse := `
{
    "jsonrpc": "2.0",
    "Result": {
        "Test": "Test",
        "PrototypeRef": "6R46iNSizv7pzHrLiR8m1qtEPC9FvLtsdKoFV9w2r6V.11111111111111111111111111111111"
    },
    "id": ""
}`
	res := struct {
		Version string      `json:"jsonrpc"`
		ID      string      `json:"id"`
		Result  UploadReply `json:"Result"`
	}{}

	expectedRes := struct {
		Version string      `json:"jsonrpc"`
		ID      string      `json:"id"`
		Result  UploadReply `json:"Result"`
	}{
		Version: "2.0",
		ID:      "",
	}

	err := json.Unmarshal([]byte(jsonResponse), &res)
	require.NoError(t, err)

	require.Equal(t, expectedRes.Version, res.Version)
	require.Equal(t, expectedRes.ID, res.ID)
	require.NotEqual(t, "", res.Result.PrototypeRef)
}
