// Code generated by "stringer -type=ClaimType"; DO NOT EDIT.

package packets

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TypeNodeJoinClaim-1]
	_ = x[TypeNodeAnnounceClaim-2]
	_ = x[TypeNodeLeaveClaim-3]
	_ = x[TypeChangeNetworkClaim-4]
}

const _ClaimType_name = "TypeNodeJoinClaimTypeNodeAnnounceClaimTypeNodeLeaveClaimTypeChangeNetworkClaim"

var _ClaimType_index = [...]uint8{0, 17, 38, 56, 78}

func (i ClaimType) String() string {
	i -= 1
	if i >= ClaimType(len(_ClaimType_index)-1) {
		return "ClaimType(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _ClaimType_name[_ClaimType_index[i]:_ClaimType_index[i+1]]
}
