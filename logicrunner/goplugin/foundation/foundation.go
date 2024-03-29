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

// Package foundation server implementation of smartcontract functions
package foundation

import (
	"github.com/tylerb/gls"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/logicrunner/common"
)

// BaseContract is a base class for all contracts.
type BaseContract struct {
}

// ProxyInterface interface any proxy of a contract implements
type ProxyInterface interface {
	GetReference() insolar.Reference
	GetPrototype() (insolar.Reference, error)
	GetCode() (insolar.Reference, error)
}

// BaseContractInterface is an interface to deal with any contract same way
type BaseContractInterface interface {
	GetReference() insolar.Reference
	GetPrototype() insolar.Reference
	GetCode() insolar.Reference
}

// GetReference - Returns public reference of contract
func (bc *BaseContract) GetReference() insolar.Reference {
	ctx := bc.GetContext()
	if ctx.Callee == nil {
		panic("context has no callee set")
	}
	return *ctx.Callee
}

// GetPrototype - Returns prototype of contract
func (bc *BaseContract) GetPrototype() insolar.Reference {
	return *bc.GetContext().Prototype
}

// GetCode - Returns prototype of contract
func (bc *BaseContract) GetCode() insolar.Reference {
	return *bc.GetContext().Code
}

// GetContext returns current calling context OBSOLETED.
func (bc *BaseContract) GetContext() *insolar.LogicCallContext {
	return GetContext()
}

// GetContext returns current calling context.
func GetContext() *insolar.LogicCallContext {
	ctx := gls.Get("callCtx")
	if ctx == nil {
		panic("object has no context")
	} else if ctx, ok := ctx.(*insolar.LogicCallContext); ok {
		return ctx
	} else {
		panic("wrong type of context")
	}
}

// GetImplementationFor finds delegate typed r in object and returns it
func GetImplementationFor(object, ofType insolar.Reference) (insolar.Reference, error) {
	return common.CurrentProxyCtx.GetDelegate(object, ofType)
}

// NewChildrenTypedIterator returns children with corresponding type iterator
func (bc *BaseContract) NewChildrenTypedIterator(childPrototype insolar.Reference) (*common.ChildrenTypedIterator, error) {
	return common.CurrentProxyCtx.GetObjChildrenIterator(bc.GetReference(), childPrototype, "")
}

// GetObject create proxy by address
// unimplemented
func GetObject(ref insolar.Reference) ProxyInterface {
	panic("not implemented")
}

// SelfDestruct contract will be marked as deleted
func (bc *BaseContract) SelfDestruct() error {
	return common.CurrentProxyCtx.DeactivateObject(bc.GetReference())
}

// Error elementary string based error struct satisfying builtin error interface
//    foundation.Error{"some err"}
type Error struct {
	S string
}

// Error returns error in string format
func (e *Error) Error() string {
	return e.S
}
