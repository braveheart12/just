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

package record

import (
	"hash"

	"github.com/pkg/errors"
)

// Hash hashes record with provided hash implementation.
func Hash(h hash.Hash, rec Record) []byte {
	buf, err := rec.Marshal()
	if err != nil {
		panic(errors.Wrap(err, "failed to marshal virtual record"))
	}
	_, err = h.Write(buf)
	if err != nil {
		panic(errors.Wrap(err, "failed to write virtual record hash to buffer"))
	}
	return h.Sum(nil)
}
