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

package object

import (
	"bytes"
	"context"
	"sync"

	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/internal/ledger/store"
)

// TypeID encodes a record object type.
type TypeID uint32

// TypeIDSize is a size of TypeID type.
const TypeIDSize = 4

//go:generate minimock -i github.com/insolar/insolar/ledger/object.RecordStorage -o ./ -s _mock.go

// RecordStorage is an union of RecordAccessor and RecordModifier
type RecordStorage interface {
	RecordAccessor
	RecordModifier
}

//go:generate minimock -i github.com/insolar/insolar/ledger/object.RecordAccessor -o ./ -s _mock.go

// RecordAccessor provides info about record-values from storage.
type RecordAccessor interface {
	// ForID returns record for provided id.
	ForID(ctx context.Context, id insolar.ID) (record.Material, error)
}

//go:generate minimock -i github.com/insolar/insolar/ledger/object.RecordCollectionAccessor -o ./ -s _mock.go

// RecordCollectionAccessor provides methods for querying records with specific search conditions.
type RecordCollectionAccessor interface {
	// ForPulse returns []MaterialRecord for a provided jetID and a pulse number.
	ForPulse(ctx context.Context, jetID insolar.JetID, pn insolar.PulseNumber) []record.Material
}

//go:generate minimock -i github.com/insolar/insolar/ledger/object.RecordModifier -o ./ -s _mock.go

// RecordModifier provides methods for setting record-values to storage.
type RecordModifier interface {
	// Set saves new record-value in storage.
	Set(ctx context.Context, id insolar.ID, rec record.Material) error
}

//go:generate minimock -i github.com/insolar/insolar/ledger/object.RecordCleaner -o ./ -s _mock.go

// RecordCleaner provides an interface for removing records from a storage.
type RecordCleaner interface {
	// DeleteForPN method removes records from a storage for a pulse
	DeleteForPN(ctx context.Context, pulse insolar.PulseNumber)
}

// RecordMemory is an in-indexStorage struct for record-storage.
type RecordMemory struct {
	jetIndex         store.JetIndexModifier
	jetIndexAccessor store.JetIndexAccessor

	lock     sync.RWMutex
	recsStor map[insolar.ID]record.Material
}

// NewRecordMemory creates a new instance of RecordMemory storage.
func NewRecordMemory() *RecordMemory {
	ji := store.NewJetIndex()
	return &RecordMemory{
		recsStor:         map[insolar.ID]record.Material{},
		jetIndex:         ji,
		jetIndexAccessor: ji,
	}
}

// Set saves new record-value in storage.
func (m *RecordMemory) Set(ctx context.Context, id insolar.ID, rec record.Material) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	_, ok := m.recsStor[id]
	if ok {
		return ErrOverride
	}

	m.recsStor[id] = rec
	m.jetIndex.Add(id, rec.JetID)

	stats.Record(ctx,
		statRecordInMemoryAddedCount.M(1),
	)

	return nil
}

// ForID returns record for provided id.
func (m *RecordMemory) ForID(ctx context.Context, id insolar.ID) (rec record.Material, err error) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	rec, ok := m.recsStor[id]
	if !ok {
		err = ErrNotFound
		return
	}

	return
}

// ForPulse returns []MaterialRecord for a provided jetID and a pulse number.
func (m *RecordMemory) ForPulse(
	ctx context.Context, jetID insolar.JetID, pn insolar.PulseNumber,
) []record.Material {
	m.lock.RLock()
	defer m.lock.RUnlock()

	ids := m.jetIndexAccessor.For(jetID)
	var res []record.Material
	for id := range ids {
		if id.Pulse() == pn {
			rec := m.recsStor[id]
			res = append(res, rec)
		}
	}

	return res
}

// DeleteForPN method removes records from a storage for all pulses until pulse (pulse included)
func (m *RecordMemory) DeleteForPN(ctx context.Context, pulse insolar.PulseNumber) {
	m.lock.Lock()
	defer m.lock.Unlock()

	for id, rec := range m.recsStor {
		if id.Pulse() != pulse {
			continue
		}

		m.jetIndex.Delete(id, rec.JetID)
		delete(m.recsStor, id)

		stats.Record(ctx,
			statRecordInMemoryRemovedCount.M(1),
		)
	}
}

// RecordDB is a DB storage implementation. It saves records to disk and does not allow removal.
type RecordDB struct {
	lock sync.RWMutex
	db   store.DB
}

type recordKey insolar.ID

func (k recordKey) Scope() store.Scope {
	return store.ScopeRecord
}

func (k recordKey) ID() []byte {
	id := insolar.ID(k)
	return bytes.Join([][]byte{id.Pulse().Bytes(), id.Hash()}, nil)
}

func newRecordKey(raw []byte) recordKey {
	pulse := insolar.NewPulseNumber(raw)
	hash := raw[pulse.Size():]

	return recordKey(*insolar.NewID(pulse, hash))
}

// NewRecordDB creates new DB storage instance.
func NewRecordDB(db store.DB) *RecordDB {
	return &RecordDB{db: db}
}

// Set saves new record-value in storage.
func (r *RecordDB) Set(ctx context.Context, id insolar.ID, rec record.Material) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.set(id, rec)
}

// TruncateHead remove all records after lastPulse
func (r *RecordDB) TruncateHead(ctx context.Context, from insolar.PulseNumber) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	it := r.db.NewIterator(recordKey(*insolar.NewID(from, nil)), false)
	defer it.Close()

	var hasKeys bool
	for it.Next() {
		hasKeys = true
		key := newRecordKey(it.Key())
		keyID := insolar.ID(key)
		err := r.db.Delete(&key)
		if err != nil {
			return errors.Wrapf(err, "can't delete key: %+v", key)
		}

		inslogger.FromContext(ctx).Debugf("Erased key with pulse number: %s. ID: %s", keyID.Pulse().String(), keyID.String())
	}

	if !hasKeys {
		inslogger.FromContext(ctx).Infof("No records. Nothing done. Pulse number: %s", from.String())
	}

	return nil
}

// ForID returns record for provided id.
func (r *RecordDB) ForID(ctx context.Context, id insolar.ID) (record.Material, error) {
	r.lock.RLock()
	defer r.lock.RUnlock()

	return r.get(id)
}

func (r *RecordDB) set(id insolar.ID, rec record.Material) error {
	key := recordKey(id)

	_, err := r.db.Get(key)
	if err == nil {
		return ErrOverride
	}

	data, err := rec.Marshal()
	if err != nil {
		return err
	}

	return r.db.Set(key, data)
}

func (r *RecordDB) get(id insolar.ID) (record.Material, error) {
	buff, err := r.db.Get(recordKey(id))
	if err == store.ErrNotFound {
		err = ErrNotFound
		return record.Material{}, err
	}
	if err != nil {
		return record.Material{}, err
	}

	rec := record.Material{}
	err = rec.Unmarshal(buff)

	return rec, err
}
