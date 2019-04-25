package object

import (
	"context"
	"sync"

	"github.com/insolar/insolar/insolar"
)

type IndexAccessor interface {
	LifelineForPulse(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID) (Lifeline, error)
}

type IndexModifier interface {
	SetLifeline(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID, lifeline Lifeline)
	SetRequest(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID, reqID insolar.ID)
	SetResultRecord(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID, resID insolar.ID)
}

type indexBucket struct {
	lifelineLock sync.RWMutex
	lifeline     *Lifeline

	requestLock sync.RWMutex
	requests    []insolar.ID

	resultLock sync.RWMutex
	results    []insolar.ID
}

func (i *indexBucket) Lifeline() (*Lifeline, error) {
	i.lifelineLock.RLock()
	defer i.lifelineLock.RUnlock()
	if i.lifeline == nil {
		return nil, ErrLifelineNotFound
	}

	return i.lifeline
}

func (i *indexBucket) setLifeline(lifeline Lifeline) {
	i.lifelineLock.Lock()
	defer i.lifelineLock.Unlock()

	i.lifeline = &lifeline
}

func (i *indexBucket) setRequest(reqID insolar.ID) {
	i.requestLock.Lock()
	defer i.requestLock.Unlock()

	i.requests = append(i.requests, reqID)
}

func (i *indexBucket) setResult(resID insolar.ID) {
	i.resultLock.Lock()
	defer i.resultLock.Unlock()

	i.results = append(i.results, resID)
}

type InMemoryIndex struct {
	bucketsLock sync.RWMutex
	buckets     map[insolar.PulseNumber]map[insolar.ID]*indexBucket
}

func (i *InMemoryIndex) getBucket(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID) *indexBucket {
	i.bucketsLock.Lock()
	defer i.bucketsLock.Unlock()

	var objsByPn map[insolar.ID]*indexBucket
	objsByPn, ok := i.buckets[pn]
	if !ok {
		objsByPn = map[insolar.ID]*indexBucket{}
		i.buckets[pn] = objsByPn
	}

	bucket := objsByPn[objID]
	if bucket == nil {
		bucket = &indexBucket{
			requests: []insolar.ID{},
			results:  []insolar.ID{},
		}
		objsByPn[objID] = bucket
	}

	return bucket
}

func (i *InMemoryIndex) SetLifeline(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID, lifeline Lifeline) {
	b := i.getBucket(ctx, pn, objID)
	b.setLifeline(lifeline)
}

func (i *InMemoryIndex) SetRequest(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID, reqID insolar.ID) {
	b := i.getBucket(ctx, pn, objID)
	b.setRequest(reqID)
}

func (i *InMemoryIndex) SetResultRecord(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID, resID insolar.ID) {
	b := i.getBucket(ctx, pn, objID)
	b.setResult(resID)
}

func (i *InMemoryIndex) LifelineForPulse(ctx context.Context, pn insolar.PulseNumber, objID insolar.ID) (Lifeline, error) {
	b :=
}
