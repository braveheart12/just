/*
 *    Copyright 2019 Insolar Technologies
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package drop

import (
	"context"

	"github.com/insolar/insolar/core"
	"github.com/insolar/insolar/ledger/storage/jet"
)

// Modifier provides an interface for modifying jetdrops.
//go:generate minimock -i github.com/insolar/insolar/ledger/storage/drop.Modifier -o ./ -s _mock.go
type Modifier interface {
	Set(ctx context.Context, jetID core.JetID, drop jet.Drop) error
	Delete(pulse core.PulseNumber)
}

// Accessor provides an interface for accessing jetdrops.
//go:generate minimock -i github.com/insolar/insolar/ledger/storage/drop.Accessor -o ./ -s _mock.go
type Accessor interface {
	ForPulse(ctx context.Context, jetID core.JetID, pulse core.PulseNumber) (jet.Drop, error)
}
