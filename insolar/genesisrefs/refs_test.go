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

package genesisrefs

import (
	"testing"

	"github.com/insolar/insolar/insolar"
	"github.com/insolar/insolar/insolar/rootdomain"
	"github.com/stretchr/testify/require"
)

func TestReferences(t *testing.T) {
	pairs := map[string]struct {
		got    insolar.Reference
		expect string
	}{
		insolar.GenesisNameRootDomain: {
			got:    ContractRootDomain,
			expect: "1tJDfbCkSyKhquYqHjthqYBdkbPhu12Pqqsz2Q2ypT.11111111111111111111111111111111",
		},
		insolar.GenesisNameNodeDomain: {
			got:    ContractNodeDomain,
			expect: "1tJCmjNyfpFW9be4tXkdmzWYVksCmPyRnYowXA3NFL.11111111111111111111111111111111",
		},
		insolar.GenesisNameNodeRecord: {
			got:    ContractNodeRecord,
			expect: "1tJBibKKGMKYGYoq4ZGuEALzRJZhLgH6nbwjeMghux.11111111111111111111111111111111",
		},
		insolar.GenesisNameRootMember: {
			got:    ContractRootMember,
			expect: "1tJD88Em52WfcZD58UyAfHqKQG1hGuRbsBR3DAZeuQ.11111111111111111111111111111111",
		},
		insolar.GenesisNameRootWallet: {
			got:    ContractWallet,
			expect: "1tJC2w4YZCzUZUKg1TbFdsFVAwpnuq2RiiMH6Nwnq5.11111111111111111111111111111111",
		},
		insolar.GenesisNameAllowance: {
			got:    ContractAllowance,
			expect: "1tJDi9hBDac9ULJFs4AUUTTKzgwjhikA4po7zfg6rN.11111111111111111111111111111111",
		},
	}

	for n, p := range pairs {
		t.Run(n, func(t *testing.T) {
			require.Equal(t, p.expect, p.got.String(), "reference is stable")
		})
	}
}

func TestRootDomain(t *testing.T) {
	ref1 := rootdomain.RootDomain.Ref()
	ref2 := rootdomain.GenesisRef(insolar.GenesisNameRootDomain)
	require.Equal(t, ref1.String(), ref2.String(), "reference is the same")
}
