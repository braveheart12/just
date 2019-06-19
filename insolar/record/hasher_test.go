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

package record_test

import (
	"crypto/sha256"
	"testing"

	"github.com/insolar/insolar/insolar/gen"
	"github.com/insolar/insolar/insolar/record"
	"github.com/stretchr/testify/assert"
)

func TestHashVirtual(t *testing.T) {
	t.Parallel()

	t.Run("check consistent hash for virtual record", func(t *testing.T) {
		t.Parallel()

		rec := genRecord()
		h := sha256.New()
		hash1 := record.Hash(h, rec)

		h = sha256.New()
		hash2 := record.Hash(h, rec)
		assert.Equal(t, hash1, hash2)
	})

	t.Run("different hash for changed virtual record", func(t *testing.T) {
		t.Parallel()

		rec1 := genRecord()
		h := sha256.New()
		hash1 := record.Hash(h, rec1)

		rec2 := &record.Request{}
		h = sha256.New()
		hash2 := record.Hash(h, rec2)
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("different hashes for different virtual records", func(t *testing.T) {
		t.Parallel()

		recFoo := genRecord()
		h := sha256.New()
		hashFoo := record.Hash(h, recFoo)

		recBar := genRecord()
		h = sha256.New()
		hashBar := record.Hash(h, recBar)

		assert.NotEqual(t, hashFoo, hashBar)
	})
}

// genRecord generates random record.
func genRecord() record.Record {
	obj := gen.Reference()
	return &record.Request{
		Object: &obj,
	}
}
