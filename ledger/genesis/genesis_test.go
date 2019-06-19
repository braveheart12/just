//
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
//

package genesis

import (
	"encoding/hex"
	"testing"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/record"
	"github.com/stretchr/testify/require"
)

func TestGenesisRecordMarshalUnmarshal(t *testing.T) {
	genIn := &record.Genesis{
		Hash: insolar.GenesisRecord,
	}

	dataIn, err := genIn.Marshal()
	require.NoError(t, err)

	require.Equal(t,
		"a20101ac", hex.EncodeToString(dataIn),
		"genesis binary representation always the same")

	genOut := &record.Genesis{}
	err = genOut.Unmarshal(dataIn)
	require.NoError(t, err, "genesis record unmarshal w/o error")

	dataOut, err := genOut.Marshal()
	require.NoError(t, err)
	require.Equal(t, dataIn, dataOut, "marshal-unmarshal-marshal gives the same binary result")
}
