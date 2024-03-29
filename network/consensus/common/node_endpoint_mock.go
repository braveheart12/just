package common

/*
DO NOT EDIT!
This code was generated automatically using github.com/gojuno/minimock v1.9
The original interface "NodeEndpoint" can be found in github.com/insolar/insolar/network/consensus/common
*/
import (
	"sync/atomic"
	"time"

	"github.com/gojuno/minimock"
	packets "github.com/insolar/insolar/network/consensusv1/packets"
)

//NodeEndpointMock implements github.com/insolar/insolar/network/consensus/common.NodeEndpoint
type NodeEndpointMock struct {
	t minimock.Tester

	GetEndpointTypeFunc       func() (r NodeEndpointType)
	GetEndpointTypeCounter    uint64
	GetEndpointTypePreCounter uint64
	GetEndpointTypeMock       mNodeEndpointMockGetEndpointType

	GetIPAddressFunc       func() (r packets.NodeAddress)
	GetIPAddressCounter    uint64
	GetIPAddressPreCounter uint64
	GetIPAddressMock       mNodeEndpointMockGetIPAddress

	GetNameAddressFunc       func() (r HostAddress)
	GetNameAddressCounter    uint64
	GetNameAddressPreCounter uint64
	GetNameAddressMock       mNodeEndpointMockGetNameAddress

	GetRelayIDFunc       func() (r ShortNodeID)
	GetRelayIDCounter    uint64
	GetRelayIDPreCounter uint64
	GetRelayIDMock       mNodeEndpointMockGetRelayID
}

//NewNodeEndpointMock returns a mock for github.com/insolar/insolar/network/consensus/common.NodeEndpoint
func NewNodeEndpointMock(t minimock.Tester) *NodeEndpointMock {
	m := &NodeEndpointMock{t: t}

	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.GetEndpointTypeMock = mNodeEndpointMockGetEndpointType{mock: m}
	m.GetIPAddressMock = mNodeEndpointMockGetIPAddress{mock: m}
	m.GetNameAddressMock = mNodeEndpointMockGetNameAddress{mock: m}
	m.GetRelayIDMock = mNodeEndpointMockGetRelayID{mock: m}

	return m
}

type mNodeEndpointMockGetEndpointType struct {
	mock              *NodeEndpointMock
	mainExpectation   *NodeEndpointMockGetEndpointTypeExpectation
	expectationSeries []*NodeEndpointMockGetEndpointTypeExpectation
}

type NodeEndpointMockGetEndpointTypeExpectation struct {
	result *NodeEndpointMockGetEndpointTypeResult
}

type NodeEndpointMockGetEndpointTypeResult struct {
	r NodeEndpointType
}

//Expect specifies that invocation of NodeEndpoint.GetEndpointType is expected from 1 to Infinity times
func (m *mNodeEndpointMockGetEndpointType) Expect() *mNodeEndpointMockGetEndpointType {
	m.mock.GetEndpointTypeFunc = nil
	m.expectationSeries = nil

	if m.mainExpectation == nil {
		m.mainExpectation = &NodeEndpointMockGetEndpointTypeExpectation{}
	}

	return m
}

//Return specifies results of invocation of NodeEndpoint.GetEndpointType
func (m *mNodeEndpointMockGetEndpointType) Return(r NodeEndpointType) *NodeEndpointMock {
	m.mock.GetEndpointTypeFunc = nil
	m.expectationSeries = nil

	if m.mainExpectation == nil {
		m.mainExpectation = &NodeEndpointMockGetEndpointTypeExpectation{}
	}
	m.mainExpectation.result = &NodeEndpointMockGetEndpointTypeResult{r}
	return m.mock
}

//ExpectOnce specifies that invocation of NodeEndpoint.GetEndpointType is expected once
func (m *mNodeEndpointMockGetEndpointType) ExpectOnce() *NodeEndpointMockGetEndpointTypeExpectation {
	m.mock.GetEndpointTypeFunc = nil
	m.mainExpectation = nil

	expectation := &NodeEndpointMockGetEndpointTypeExpectation{}

	m.expectationSeries = append(m.expectationSeries, expectation)
	return expectation
}

func (e *NodeEndpointMockGetEndpointTypeExpectation) Return(r NodeEndpointType) {
	e.result = &NodeEndpointMockGetEndpointTypeResult{r}
}

//Set uses given function f as a mock of NodeEndpoint.GetEndpointType method
func (m *mNodeEndpointMockGetEndpointType) Set(f func() (r NodeEndpointType)) *NodeEndpointMock {
	m.mainExpectation = nil
	m.expectationSeries = nil

	m.mock.GetEndpointTypeFunc = f
	return m.mock
}

//GetEndpointType implements github.com/insolar/insolar/network/consensus/common.NodeEndpoint interface
func (m *NodeEndpointMock) GetEndpointType() (r NodeEndpointType) {
	counter := atomic.AddUint64(&m.GetEndpointTypePreCounter, 1)
	defer atomic.AddUint64(&m.GetEndpointTypeCounter, 1)

	if len(m.GetEndpointTypeMock.expectationSeries) > 0 {
		if counter > uint64(len(m.GetEndpointTypeMock.expectationSeries)) {
			m.t.Fatalf("Unexpected call to NodeEndpointMock.GetEndpointType.")
			return
		}

		result := m.GetEndpointTypeMock.expectationSeries[counter-1].result
		if result == nil {
			m.t.Fatal("No results are set for the NodeEndpointMock.GetEndpointType")
			return
		}

		r = result.r

		return
	}

	if m.GetEndpointTypeMock.mainExpectation != nil {

		result := m.GetEndpointTypeMock.mainExpectation.result
		if result == nil {
			m.t.Fatal("No results are set for the NodeEndpointMock.GetEndpointType")
		}

		r = result.r

		return
	}

	if m.GetEndpointTypeFunc == nil {
		m.t.Fatalf("Unexpected call to NodeEndpointMock.GetEndpointType.")
		return
	}

	return m.GetEndpointTypeFunc()
}

//GetEndpointTypeMinimockCounter returns a count of NodeEndpointMock.GetEndpointTypeFunc invocations
func (m *NodeEndpointMock) GetEndpointTypeMinimockCounter() uint64 {
	return atomic.LoadUint64(&m.GetEndpointTypeCounter)
}

//GetEndpointTypeMinimockPreCounter returns the value of NodeEndpointMock.GetEndpointType invocations
func (m *NodeEndpointMock) GetEndpointTypeMinimockPreCounter() uint64 {
	return atomic.LoadUint64(&m.GetEndpointTypePreCounter)
}

//GetEndpointTypeFinished returns true if mock invocations count is ok
func (m *NodeEndpointMock) GetEndpointTypeFinished() bool {
	// if expectation series were set then invocations count should be equal to expectations count
	if len(m.GetEndpointTypeMock.expectationSeries) > 0 {
		return atomic.LoadUint64(&m.GetEndpointTypeCounter) == uint64(len(m.GetEndpointTypeMock.expectationSeries))
	}

	// if main expectation was set then invocations count should be greater than zero
	if m.GetEndpointTypeMock.mainExpectation != nil {
		return atomic.LoadUint64(&m.GetEndpointTypeCounter) > 0
	}

	// if func was set then invocations count should be greater than zero
	if m.GetEndpointTypeFunc != nil {
		return atomic.LoadUint64(&m.GetEndpointTypeCounter) > 0
	}

	return true
}

type mNodeEndpointMockGetIPAddress struct {
	mock              *NodeEndpointMock
	mainExpectation   *NodeEndpointMockGetIPAddressExpectation
	expectationSeries []*NodeEndpointMockGetIPAddressExpectation
}

type NodeEndpointMockGetIPAddressExpectation struct {
	result *NodeEndpointMockGetIPAddressResult
}

type NodeEndpointMockGetIPAddressResult struct {
	r packets.NodeAddress
}

//Expect specifies that invocation of NodeEndpoint.GetIPAddress is expected from 1 to Infinity times
func (m *mNodeEndpointMockGetIPAddress) Expect() *mNodeEndpointMockGetIPAddress {
	m.mock.GetIPAddressFunc = nil
	m.expectationSeries = nil

	if m.mainExpectation == nil {
		m.mainExpectation = &NodeEndpointMockGetIPAddressExpectation{}
	}

	return m
}

//Return specifies results of invocation of NodeEndpoint.GetIPAddress
func (m *mNodeEndpointMockGetIPAddress) Return(r packets.NodeAddress) *NodeEndpointMock {
	m.mock.GetIPAddressFunc = nil
	m.expectationSeries = nil

	if m.mainExpectation == nil {
		m.mainExpectation = &NodeEndpointMockGetIPAddressExpectation{}
	}
	m.mainExpectation.result = &NodeEndpointMockGetIPAddressResult{r}
	return m.mock
}

//ExpectOnce specifies that invocation of NodeEndpoint.GetIPAddress is expected once
func (m *mNodeEndpointMockGetIPAddress) ExpectOnce() *NodeEndpointMockGetIPAddressExpectation {
	m.mock.GetIPAddressFunc = nil
	m.mainExpectation = nil

	expectation := &NodeEndpointMockGetIPAddressExpectation{}

	m.expectationSeries = append(m.expectationSeries, expectation)
	return expectation
}

func (e *NodeEndpointMockGetIPAddressExpectation) Return(r packets.NodeAddress) {
	e.result = &NodeEndpointMockGetIPAddressResult{r}
}

//Set uses given function f as a mock of NodeEndpoint.GetIPAddress method
func (m *mNodeEndpointMockGetIPAddress) Set(f func() (r packets.NodeAddress)) *NodeEndpointMock {
	m.mainExpectation = nil
	m.expectationSeries = nil

	m.mock.GetIPAddressFunc = f
	return m.mock
}

//GetIPAddress implements github.com/insolar/insolar/network/consensus/common.NodeEndpoint interface
func (m *NodeEndpointMock) GetIPAddress() (r packets.NodeAddress) {
	counter := atomic.AddUint64(&m.GetIPAddressPreCounter, 1)
	defer atomic.AddUint64(&m.GetIPAddressCounter, 1)

	if len(m.GetIPAddressMock.expectationSeries) > 0 {
		if counter > uint64(len(m.GetIPAddressMock.expectationSeries)) {
			m.t.Fatalf("Unexpected call to NodeEndpointMock.GetIPAddress.")
			return
		}

		result := m.GetIPAddressMock.expectationSeries[counter-1].result
		if result == nil {
			m.t.Fatal("No results are set for the NodeEndpointMock.GetIPAddress")
			return
		}

		r = result.r

		return
	}

	if m.GetIPAddressMock.mainExpectation != nil {

		result := m.GetIPAddressMock.mainExpectation.result
		if result == nil {
			m.t.Fatal("No results are set for the NodeEndpointMock.GetIPAddress")
		}

		r = result.r

		return
	}

	if m.GetIPAddressFunc == nil {
		m.t.Fatalf("Unexpected call to NodeEndpointMock.GetIPAddress.")
		return
	}

	return m.GetIPAddressFunc()
}

//GetIPAddressMinimockCounter returns a count of NodeEndpointMock.GetIPAddressFunc invocations
func (m *NodeEndpointMock) GetIPAddressMinimockCounter() uint64 {
	return atomic.LoadUint64(&m.GetIPAddressCounter)
}

//GetIPAddressMinimockPreCounter returns the value of NodeEndpointMock.GetIPAddress invocations
func (m *NodeEndpointMock) GetIPAddressMinimockPreCounter() uint64 {
	return atomic.LoadUint64(&m.GetIPAddressPreCounter)
}

//GetIPAddressFinished returns true if mock invocations count is ok
func (m *NodeEndpointMock) GetIPAddressFinished() bool {
	// if expectation series were set then invocations count should be equal to expectations count
	if len(m.GetIPAddressMock.expectationSeries) > 0 {
		return atomic.LoadUint64(&m.GetIPAddressCounter) == uint64(len(m.GetIPAddressMock.expectationSeries))
	}

	// if main expectation was set then invocations count should be greater than zero
	if m.GetIPAddressMock.mainExpectation != nil {
		return atomic.LoadUint64(&m.GetIPAddressCounter) > 0
	}

	// if func was set then invocations count should be greater than zero
	if m.GetIPAddressFunc != nil {
		return atomic.LoadUint64(&m.GetIPAddressCounter) > 0
	}

	return true
}

type mNodeEndpointMockGetNameAddress struct {
	mock              *NodeEndpointMock
	mainExpectation   *NodeEndpointMockGetNameAddressExpectation
	expectationSeries []*NodeEndpointMockGetNameAddressExpectation
}

type NodeEndpointMockGetNameAddressExpectation struct {
	result *NodeEndpointMockGetNameAddressResult
}

type NodeEndpointMockGetNameAddressResult struct {
	r HostAddress
}

//Expect specifies that invocation of NodeEndpoint.GetNameAddress is expected from 1 to Infinity times
func (m *mNodeEndpointMockGetNameAddress) Expect() *mNodeEndpointMockGetNameAddress {
	m.mock.GetNameAddressFunc = nil
	m.expectationSeries = nil

	if m.mainExpectation == nil {
		m.mainExpectation = &NodeEndpointMockGetNameAddressExpectation{}
	}

	return m
}

//Return specifies results of invocation of NodeEndpoint.GetNameAddress
func (m *mNodeEndpointMockGetNameAddress) Return(r HostAddress) *NodeEndpointMock {
	m.mock.GetNameAddressFunc = nil
	m.expectationSeries = nil

	if m.mainExpectation == nil {
		m.mainExpectation = &NodeEndpointMockGetNameAddressExpectation{}
	}
	m.mainExpectation.result = &NodeEndpointMockGetNameAddressResult{r}
	return m.mock
}

//ExpectOnce specifies that invocation of NodeEndpoint.GetNameAddress is expected once
func (m *mNodeEndpointMockGetNameAddress) ExpectOnce() *NodeEndpointMockGetNameAddressExpectation {
	m.mock.GetNameAddressFunc = nil
	m.mainExpectation = nil

	expectation := &NodeEndpointMockGetNameAddressExpectation{}

	m.expectationSeries = append(m.expectationSeries, expectation)
	return expectation
}

func (e *NodeEndpointMockGetNameAddressExpectation) Return(r HostAddress) {
	e.result = &NodeEndpointMockGetNameAddressResult{r}
}

//Set uses given function f as a mock of NodeEndpoint.GetNameAddress method
func (m *mNodeEndpointMockGetNameAddress) Set(f func() (r HostAddress)) *NodeEndpointMock {
	m.mainExpectation = nil
	m.expectationSeries = nil

	m.mock.GetNameAddressFunc = f
	return m.mock
}

//GetNameAddress implements github.com/insolar/insolar/network/consensus/common.NodeEndpoint interface
func (m *NodeEndpointMock) GetNameAddress() (r HostAddress) {
	counter := atomic.AddUint64(&m.GetNameAddressPreCounter, 1)
	defer atomic.AddUint64(&m.GetNameAddressCounter, 1)

	if len(m.GetNameAddressMock.expectationSeries) > 0 {
		if counter > uint64(len(m.GetNameAddressMock.expectationSeries)) {
			m.t.Fatalf("Unexpected call to NodeEndpointMock.GetNameAddress.")
			return
		}

		result := m.GetNameAddressMock.expectationSeries[counter-1].result
		if result == nil {
			m.t.Fatal("No results are set for the NodeEndpointMock.GetNameAddress")
			return
		}

		r = result.r

		return
	}

	if m.GetNameAddressMock.mainExpectation != nil {

		result := m.GetNameAddressMock.mainExpectation.result
		if result == nil {
			m.t.Fatal("No results are set for the NodeEndpointMock.GetNameAddress")
		}

		r = result.r

		return
	}

	if m.GetNameAddressFunc == nil {
		m.t.Fatalf("Unexpected call to NodeEndpointMock.GetNameAddress.")
		return
	}

	return m.GetNameAddressFunc()
}

//GetNameAddressMinimockCounter returns a count of NodeEndpointMock.GetNameAddressFunc invocations
func (m *NodeEndpointMock) GetNameAddressMinimockCounter() uint64 {
	return atomic.LoadUint64(&m.GetNameAddressCounter)
}

//GetNameAddressMinimockPreCounter returns the value of NodeEndpointMock.GetNameAddress invocations
func (m *NodeEndpointMock) GetNameAddressMinimockPreCounter() uint64 {
	return atomic.LoadUint64(&m.GetNameAddressPreCounter)
}

//GetNameAddressFinished returns true if mock invocations count is ok
func (m *NodeEndpointMock) GetNameAddressFinished() bool {
	// if expectation series were set then invocations count should be equal to expectations count
	if len(m.GetNameAddressMock.expectationSeries) > 0 {
		return atomic.LoadUint64(&m.GetNameAddressCounter) == uint64(len(m.GetNameAddressMock.expectationSeries))
	}

	// if main expectation was set then invocations count should be greater than zero
	if m.GetNameAddressMock.mainExpectation != nil {
		return atomic.LoadUint64(&m.GetNameAddressCounter) > 0
	}

	// if func was set then invocations count should be greater than zero
	if m.GetNameAddressFunc != nil {
		return atomic.LoadUint64(&m.GetNameAddressCounter) > 0
	}

	return true
}

type mNodeEndpointMockGetRelayID struct {
	mock              *NodeEndpointMock
	mainExpectation   *NodeEndpointMockGetRelayIDExpectation
	expectationSeries []*NodeEndpointMockGetRelayIDExpectation
}

type NodeEndpointMockGetRelayIDExpectation struct {
	result *NodeEndpointMockGetRelayIDResult
}

type NodeEndpointMockGetRelayIDResult struct {
	r ShortNodeID
}

//Expect specifies that invocation of NodeEndpoint.GetRelayID is expected from 1 to Infinity times
func (m *mNodeEndpointMockGetRelayID) Expect() *mNodeEndpointMockGetRelayID {
	m.mock.GetRelayIDFunc = nil
	m.expectationSeries = nil

	if m.mainExpectation == nil {
		m.mainExpectation = &NodeEndpointMockGetRelayIDExpectation{}
	}

	return m
}

//Return specifies results of invocation of NodeEndpoint.GetRelayID
func (m *mNodeEndpointMockGetRelayID) Return(r ShortNodeID) *NodeEndpointMock {
	m.mock.GetRelayIDFunc = nil
	m.expectationSeries = nil

	if m.mainExpectation == nil {
		m.mainExpectation = &NodeEndpointMockGetRelayIDExpectation{}
	}
	m.mainExpectation.result = &NodeEndpointMockGetRelayIDResult{r}
	return m.mock
}

//ExpectOnce specifies that invocation of NodeEndpoint.GetRelayID is expected once
func (m *mNodeEndpointMockGetRelayID) ExpectOnce() *NodeEndpointMockGetRelayIDExpectation {
	m.mock.GetRelayIDFunc = nil
	m.mainExpectation = nil

	expectation := &NodeEndpointMockGetRelayIDExpectation{}

	m.expectationSeries = append(m.expectationSeries, expectation)
	return expectation
}

func (e *NodeEndpointMockGetRelayIDExpectation) Return(r ShortNodeID) {
	e.result = &NodeEndpointMockGetRelayIDResult{r}
}

//Set uses given function f as a mock of NodeEndpoint.GetRelayID method
func (m *mNodeEndpointMockGetRelayID) Set(f func() (r ShortNodeID)) *NodeEndpointMock {
	m.mainExpectation = nil
	m.expectationSeries = nil

	m.mock.GetRelayIDFunc = f
	return m.mock
}

//GetRelayID implements github.com/insolar/insolar/network/consensus/common.NodeEndpoint interface
func (m *NodeEndpointMock) GetRelayID() (r ShortNodeID) {
	counter := atomic.AddUint64(&m.GetRelayIDPreCounter, 1)
	defer atomic.AddUint64(&m.GetRelayIDCounter, 1)

	if len(m.GetRelayIDMock.expectationSeries) > 0 {
		if counter > uint64(len(m.GetRelayIDMock.expectationSeries)) {
			m.t.Fatalf("Unexpected call to NodeEndpointMock.GetRelayID.")
			return
		}

		result := m.GetRelayIDMock.expectationSeries[counter-1].result
		if result == nil {
			m.t.Fatal("No results are set for the NodeEndpointMock.GetRelayID")
			return
		}

		r = result.r

		return
	}

	if m.GetRelayIDMock.mainExpectation != nil {

		result := m.GetRelayIDMock.mainExpectation.result
		if result == nil {
			m.t.Fatal("No results are set for the NodeEndpointMock.GetRelayID")
		}

		r = result.r

		return
	}

	if m.GetRelayIDFunc == nil {
		m.t.Fatalf("Unexpected call to NodeEndpointMock.GetRelayID.")
		return
	}

	return m.GetRelayIDFunc()
}

//GetRelayIDMinimockCounter returns a count of NodeEndpointMock.GetRelayIDFunc invocations
func (m *NodeEndpointMock) GetRelayIDMinimockCounter() uint64 {
	return atomic.LoadUint64(&m.GetRelayIDCounter)
}

//GetRelayIDMinimockPreCounter returns the value of NodeEndpointMock.GetRelayID invocations
func (m *NodeEndpointMock) GetRelayIDMinimockPreCounter() uint64 {
	return atomic.LoadUint64(&m.GetRelayIDPreCounter)
}

//GetRelayIDFinished returns true if mock invocations count is ok
func (m *NodeEndpointMock) GetRelayIDFinished() bool {
	// if expectation series were set then invocations count should be equal to expectations count
	if len(m.GetRelayIDMock.expectationSeries) > 0 {
		return atomic.LoadUint64(&m.GetRelayIDCounter) == uint64(len(m.GetRelayIDMock.expectationSeries))
	}

	// if main expectation was set then invocations count should be greater than zero
	if m.GetRelayIDMock.mainExpectation != nil {
		return atomic.LoadUint64(&m.GetRelayIDCounter) > 0
	}

	// if func was set then invocations count should be greater than zero
	if m.GetRelayIDFunc != nil {
		return atomic.LoadUint64(&m.GetRelayIDCounter) > 0
	}

	return true
}

//ValidateCallCounters checks that all mocked methods of the interface have been called at least once
//Deprecated: please use MinimockFinish method or use Finish method of minimock.Controller
func (m *NodeEndpointMock) ValidateCallCounters() {

	if !m.GetEndpointTypeFinished() {
		m.t.Fatal("Expected call to NodeEndpointMock.GetEndpointType")
	}

	if !m.GetIPAddressFinished() {
		m.t.Fatal("Expected call to NodeEndpointMock.GetIPAddress")
	}

	if !m.GetNameAddressFinished() {
		m.t.Fatal("Expected call to NodeEndpointMock.GetNameAddress")
	}

	if !m.GetRelayIDFinished() {
		m.t.Fatal("Expected call to NodeEndpointMock.GetRelayID")
	}

}

//CheckMocksCalled checks that all mocked methods of the interface have been called at least once
//Deprecated: please use MinimockFinish method or use Finish method of minimock.Controller
func (m *NodeEndpointMock) CheckMocksCalled() {
	m.Finish()
}

//Finish checks that all mocked methods of the interface have been called at least once
//Deprecated: please use MinimockFinish or use Finish method of minimock.Controller
func (m *NodeEndpointMock) Finish() {
	m.MinimockFinish()
}

//MinimockFinish checks that all mocked methods of the interface have been called at least once
func (m *NodeEndpointMock) MinimockFinish() {

	if !m.GetEndpointTypeFinished() {
		m.t.Fatal("Expected call to NodeEndpointMock.GetEndpointType")
	}

	if !m.GetIPAddressFinished() {
		m.t.Fatal("Expected call to NodeEndpointMock.GetIPAddress")
	}

	if !m.GetNameAddressFinished() {
		m.t.Fatal("Expected call to NodeEndpointMock.GetNameAddress")
	}

	if !m.GetRelayIDFinished() {
		m.t.Fatal("Expected call to NodeEndpointMock.GetRelayID")
	}

}

//Wait waits for all mocked methods to be called at least once
//Deprecated: please use MinimockWait or use Wait method of minimock.Controller
func (m *NodeEndpointMock) Wait(timeout time.Duration) {
	m.MinimockWait(timeout)
}

//MinimockWait waits for all mocked methods to be called at least once
//this method is called by minimock.Controller
func (m *NodeEndpointMock) MinimockWait(timeout time.Duration) {
	timeoutCh := time.After(timeout)
	for {
		ok := true
		ok = ok && m.GetEndpointTypeFinished()
		ok = ok && m.GetIPAddressFinished()
		ok = ok && m.GetNameAddressFinished()
		ok = ok && m.GetRelayIDFinished()

		if ok {
			return
		}

		select {
		case <-timeoutCh:

			if !m.GetEndpointTypeFinished() {
				m.t.Error("Expected call to NodeEndpointMock.GetEndpointType")
			}

			if !m.GetIPAddressFinished() {
				m.t.Error("Expected call to NodeEndpointMock.GetIPAddress")
			}

			if !m.GetNameAddressFinished() {
				m.t.Error("Expected call to NodeEndpointMock.GetNameAddress")
			}

			if !m.GetRelayIDFinished() {
				m.t.Error("Expected call to NodeEndpointMock.GetRelayID")
			}

			m.t.Fatalf("Some mocks were not called on time: %s", timeout)
			return
		default:
			time.Sleep(time.Millisecond)
		}
	}
}

//AllMocksCalled returns true if all mocked methods were called before the execution of AllMocksCalled,
//it can be used with assert/require, i.e. assert.True(mock.AllMocksCalled())
func (m *NodeEndpointMock) AllMocksCalled() bool {

	if !m.GetEndpointTypeFinished() {
		return false
	}

	if !m.GetIPAddressFinished() {
		return false
	}

	if !m.GetNameAddressFinished() {
		return false
	}

	if !m.GetRelayIDFinished() {
		return false
	}

	return true
}
