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

package contracts

import (
	"github.com/insolar/insolar/application/contract/member"
	"github.com/insolar/insolar/application/contract/nodedomain"
	"github.com/insolar/insolar/application/contract/rootdomain"
	"github.com/insolar/insolar/application/contract/wallet"
	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/genesisrefs"
	insrootdomain "github.com/insolar/insolar/insolar/rootdomain"
)

// GenesisContractsStates returns list contract configs for genesis.
//
// Hint: order matters, because of dependency contracts on each other.
func GenesisContractsStates(cfg insolar.GenesisContractsConfig) []insolar.GenesisContractState {
	for name := range cfg.OraclePublicKeys {
		genesisrefs.ContractOracleMembers[name] = insrootdomain.GenesisRef(name + insolar.GenesisNameMember)
	}
	result := []insolar.GenesisContractState{
		rootDomain(),
		nodeDomain(),
		getMemberGenesisContractState(cfg.RootPublicKey, insolar.GetGenesisNameRootMember(), insolar.GetGenesisNameRootDomain()),
		getWalletGenesisContractState(cfg.RootBalance, insolar.GetGenesisNameRootWallet(), insolar.GetGenesisNameRootMember()),
		getMemberGenesisContractState(cfg.MDAdminPublicKey, insolar.GetGenesisNameMDAdminMember(), insolar.GetGenesisNameRootDomain()),
		getWalletGenesisContractState(cfg.MDBalance, insolar.GetGenesisNameMDWallet(), insolar.GetGenesisNameMDAdminMember()),
	}
	for name, key := range cfg.OraclePublicKeys {
		result = append(result, getMemberGenesisContractState(key, insolar.GetGenesisNameOracleMembers(name), insolar.GetGenesisNameRootDomain()))
	}
	return result
}

func rootDomain() insolar.GenesisContractState {
	return insolar.GenesisContractState{
		Name:       insolar.GenesisNameRootDomain,
		ParentName: "",

		Memory: mustGenMemory(&rootdomain.RootDomain{
			RootMember:        genesisrefs.ContractRootMember,
			OracleMembers:     genesisrefs.ContractOracleMembers,
			MDAdminMember:     genesisrefs.ContractMDAdminMember,
			MDWallet:          genesisrefs.ContractMDWallet,
			BurnAddressMap:    map[string]insolar.Reference{},
			PublicKeyMap:      map[string]insolar.Reference{},
			FreeBurnAddresses: []string{},
			NodeDomain:        genesisrefs.ContractNodeDomain,
		}),
	}
}

func nodeDomain() insolar.GenesisContractState {
	nd, _ := nodedomain.NewNodeDomain()
	return insolar.GenesisContractState{
		Name:       insolar.GenesisNameNodeDomain,
		ParentName: insolar.GenesisNameRootDomain,
		Memory:     mustGenMemory(nd),
	}
}

func getMemberGenesisContractState(publicKey string, name string, parrent string) insolar.GenesisContractState {
	m, err := member.NewBasicMember(name, publicKey)
	if err != nil {
		panic("`" + name + "` member constructor failed")
	}

	return insolar.GenesisContractState{
		Name:       name,
		ParentName: parrent,
		Memory:     mustGenMemory(m),
	}
}

func getWalletGenesisContractState(balance string, name string, parrent string) insolar.GenesisContractState {
	w, err := wallet.New(balance)
	if err != nil {
		panic("failed to create ` " + name + "` wallet instance")
	}

	return insolar.GenesisContractState{
		Name:       name,
		ParentName: parrent,
		Delegate:   true,
		Memory:     mustGenMemory(w),
	}
}

func mustGenMemory(data interface{}) []byte {
	b, err := insolar.Serialize(data)
	if err != nil {
		panic("failed to serialize contract data")
	}
	return b
}
