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

package replication

import (
	"context"
	"fmt"
	"sync"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/jet"
	"github.com/insolar/insolar/insolar/message"
	"github.com/insolar/insolar/insolar/pulse"
	"github.com/insolar/insolar/insolar/record"
	"github.com/insolar/insolar/insolar/reply"
	"github.com/insolar/insolar/insolar/utils"
	"github.com/insolar/insolar/instrumentation/inslogger"
	"github.com/insolar/insolar/ledger/drop"
	"github.com/insolar/insolar/ledger/light/executor"
	"github.com/insolar/insolar/ledger/object"
	"github.com/pkg/errors"
	"go.opencensus.io/stats"
)

// LightReplicator is a base interface for a sync component
type LightReplicator interface {
	// NotifyAboutPulse is method for notifying a sync component about new pulse
	NotifyAboutPulse(ctx context.Context, pn insolar.PulseNumber)
}

// LightReplicatorDefault is a base impl of LightReplicator
type LightReplicatorDefault struct {
	once sync.Once

	jetCalculator   executor.JetCalculator
	cleaner         Cleaner
	msgBus          insolar.MessageBus
	pulseCalculator pulse.Calculator

	dropAccessor drop.Accessor
	recsAccessor object.RecordCollectionAccessor
	idxAccessor  object.IndexAccessor
	jetAccessor  jet.Accessor

	syncWaitingPulses chan insolar.PulseNumber
}

// NewReplicatorDefault creates new instance of LightReplicator
func NewReplicatorDefault(
	jetCalculator executor.JetCalculator,
	cleaner Cleaner,
	msgBus insolar.MessageBus,
	calculator pulse.Calculator,
	dropAccessor drop.Accessor,
	recsAccessor object.RecordCollectionAccessor,
	idxAccessor object.IndexAccessor,
	jetAccessor jet.Accessor,
) *LightReplicatorDefault {
	return &LightReplicatorDefault{
		jetCalculator:   jetCalculator,
		cleaner:         cleaner,
		msgBus:          msgBus,
		pulseCalculator: calculator,

		dropAccessor: dropAccessor,
		recsAccessor: recsAccessor,
		idxAccessor:  idxAccessor,
		jetAccessor:  jetAccessor,

		syncWaitingPulses: make(chan insolar.PulseNumber),
	}
}

// NotifyAboutPulse is method for notifying a sync component about new pulse
// When it's called, a provided pulse is added to a channel.
// There is a special gorutine that is reading that channel. When a new pulse is being received,
// the routine starts to gather data (with using of LightDataGatherer). After gathering all the data,
// it attempts to send it to the heavy. After sending a heavy payload to a heavy, data is deleted
// with help of Cleaner
func (lr *LightReplicatorDefault) NotifyAboutPulse(ctx context.Context, pn insolar.PulseNumber) {
	lr.once.Do(func() {
		go lr.sync(context.Background())
	})

	logger := inslogger.FromContext(ctx)
	logger.Debugf("[Replicator][NotifyAboutPulse] received pulse - %v", pn)

	prevPN, err := lr.pulseCalculator.Backwards(ctx, pn, 1)
	if err != nil {
		logger.Error("[Replicator][NotifyAboutPulse]", err)
		return
	}

	logger.Debugf("[Replicator][NotifyAboutPulse] start replication, pulse - %v", prevPN.PulseNumber)
	lr.syncWaitingPulses <- prevPN.PulseNumber
}

func (lr *LightReplicatorDefault) sync(ctx context.Context) {
	for pn := range lr.syncWaitingPulses {
		ctx, logger := inslogger.WithTraceField(ctx, utils.RandTraceID())
		logger.Debugf("[Replicator][sync] pn received - %v", pn)

		allIndexes := lr.filterAndGroupIndexes(ctx, pn)
		jets := lr.jetCalculator.MineForPulse(ctx, pn)
		logger.Debugf("[Replicator][sync] founds %v jets", len(jets))

		for _, jetID := range jets {
			msg, err := lr.heavyPayload(ctx, pn, jetID, allIndexes[jetID])
			if err != nil {
				panic(
					fmt.Sprintf(
						"[Replicator][sync] Problems with gather data for a pulse - %v and jet - %v. err - %v",
						pn,
						jetID.DebugString(),
						err,
					),
				)
			}
			err = lr.sendToHeavy(ctx, msg)
			if err != nil {
				logger.Errorf("[Replicator][sync]  Problems with sending msg to a heavy node", err)
			} else {
				logger.Debugf("[Replicator][sync]  Data has been sent to a heavy. pn - %v, jetID - %v", msg.PulseNum, msg.JetID.DebugString())
			}
		}

		lr.cleaner.NotifyAboutPulse(ctx, pn)
	}
}

func (lr *LightReplicatorDefault) sendToHeavy(ctx context.Context, data insolar.Message) error {
	rep, err := lr.msgBus.Send(ctx, data, nil)
	if err != nil {
		stats.Record(ctx,
			statErrHeavyPayloadCount.M(1),
		)
		return err
	}
	if rep != nil {
		err, ok := rep.(*reply.HeavyError)
		if ok {
			stats.Record(ctx,
				statErrHeavyPayloadCount.M(1),
			)
			return err
		}
	}
	stats.Record(ctx,
		statHeavyPayloadCount.M(1),
	)
	return nil
}

func (lr *LightReplicatorDefault) filterAndGroupIndexes(
	ctx context.Context, pn insolar.PulseNumber,
) map[insolar.JetID][]object.FilamentIndex {
	byJet := map[insolar.JetID][]object.FilamentIndex{}
	indexes := lr.idxAccessor.ForPulse(ctx, pn)
	for _, idx := range indexes {
		jetID, _ := lr.jetAccessor.ForID(ctx, pn, idx.ObjID)
		byJet[jetID] = append(byJet[jetID], idx)
	}
	return byJet
}

// ForPulseAndJet returns HeavyPayload message for a provided pulse and a jetID
func (lr *LightReplicatorDefault) heavyPayload(
	ctx context.Context,
	pn insolar.PulseNumber,
	jetID insolar.JetID,
	indexes []object.FilamentIndex,
) (*message.HeavyPayload, error) {
	dr, err := lr.dropAccessor.ForPulse(ctx, jetID, pn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch drop")
	}

	records := lr.recsAccessor.ForPulse(ctx, jetID, pn)

	return &message.HeavyPayload{
		JetID:        jetID,
		PulseNum:     pn,
		IndexBuckets: convertIndexBuckets(ctx, indexes),
		Drop:         drop.MustEncode(&dr),
		Records:      convertRecords(ctx, records),
	}, nil
}

func convertIndexBuckets(ctx context.Context, buckets []object.FilamentIndex) [][]byte {
	convertedBucks := make([][]byte, len(buckets))
	for i, buck := range buckets {
		buff, err := buck.Marshal()
		if err != nil {
			inslogger.FromContext(ctx).Errorf("problems with marshaling bucket - %v", err)
			continue
		}
		convertedBucks[i] = buff
	}

	return convertedBucks
}

func convertRecords(ctx context.Context, records []record.Material) [][]byte {
	res := make([][]byte, len(records))
	for i, r := range records {
		data, err := r.Marshal()
		if err != nil {
			inslogger.FromContext(ctx).Error("Can't serialize record", r)
		}
		res[i] = data
	}
	return res
}
