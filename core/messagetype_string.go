// Code generated by "stringer -type=MessageType"; DO NOT EDIT.

package core

import "strconv"

const _MessageType_name = "TypeCallMethodTypeCallConstructorTypeExecutorResultsTypeValidateCaseBindTypeValidationResultsTypeGetCodeTypeGetObjectTypeGetDelegateTypeGetChildrenTypeUpdateObjectTypeRegisterChildTypeJetDropTypeSetRecordTypeValidateRecordTypeSetBlobTypeGetObjectIndexTypeHeavyStartStopTypeHeavyPayloadTypeBootstrapRequest"

var _MessageType_index = [...]uint16{0, 14, 33, 52, 72, 93, 104, 117, 132, 147, 163, 180, 191, 204, 222, 233, 251, 269, 285, 305}

func (i MessageType) String() string {
	if i >= MessageType(len(_MessageType_index)-1) {
		return "MessageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MessageType_name[_MessageType_index[i]:_MessageType_index[i+1]]
}
