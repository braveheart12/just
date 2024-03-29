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

// +build functest

package functest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsgorundReload(t *testing.T) {
	_ = getBalanceNoErr(t, &root, root.ref)
	err := stopAllInsgorunds()
	// No need to stop test if this fails. All tests may stack
	assert.NoError(t, err)

	err = startAllInsgorunds()
	require.NoError(t, err)

	_ = getBalanceNoErr(t, &root, root.ref)
}
