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

package pulse

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/internal/ledger/store"
	"github.com/insolar/insolar/pulsar"
	"github.com/insolar/insolar/pulsar/entropygenerator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPulseKey(t *testing.T) {
	t.Parallel()

	expectedKey := pulseKey(insolar.GenesisPulse.PulseNumber)

	rawID := expectedKey.ID()

	actualKey := newPulseKey(rawID)
	require.Equal(t, expectedKey, actualKey)
}

func TestDropStorageDB_TruncateHead_NoSuchPulse(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	defer os.RemoveAll(tmpdir)
	assert.NoError(t, err)

	dbMock, err := store.NewBadgerDB(tmpdir)
	defer dbMock.Stop(ctx)
	require.NoError(t, err)

	pulseStore := NewDB(dbMock)

	err = pulseStore.TruncateHead(ctx, 77)
	require.Contains(t, err.Error(), "No required pulse")
}

func TestDBStore_TruncateHead(t *testing.T) {
	t.Parallel()

	ctx := inslogger.TestContext(t)
	tmpdir, err := ioutil.TempDir("", "bdb-test-")
	defer os.RemoveAll(tmpdir)
	assert.NoError(t, err)

	dbMock, err := store.NewBadgerDB(tmpdir)
	defer dbMock.Stop(ctx)
	require.NoError(t, err)

	dbStore := NewDB(dbMock)

	numElements := 100

	startPulseNumber := insolar.GenesisPulse.PulseNumber
	for i := 0; i < numElements; i++ {
		pn := startPulseNumber + insolar.PulseNumber(i)
		pulse := *pulsar.NewPulse(0, pn, &entropygenerator.StandardEntropyGenerator{})
		err := dbStore.Append(ctx, pulse)
		require.NoError(t, err)
	}

	for i := 0; i < numElements; i++ {
		_, err := dbStore.ForPulseNumber(ctx, startPulseNumber+insolar.PulseNumber(i))
		require.NoError(t, err)
	}

	numLeftElements := numElements / 2
	err = dbStore.TruncateHead(ctx, startPulseNumber+insolar.PulseNumber(numLeftElements))
	require.NoError(t, err)

	for i := 0; i < numLeftElements; i++ {
		_, err := dbStore.ForPulseNumber(ctx, startPulseNumber+insolar.PulseNumber(i))
		require.NoError(t, err)
	}

	for i := numElements - 1; i >= numLeftElements; i-- {
		_, err := dbStore.ForPulseNumber(ctx, startPulseNumber+insolar.PulseNumber(i))
		require.EqualError(t, err, ErrNotFound.Error())
	}
}
