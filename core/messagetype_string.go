// Code generated by "stringer -type MessageType"; DO NOT EDIT.

package core

import "strconv"

const _MessageType_name = "TypeCallMethodTypeCallConstructorTypeRequestCallTypeGetCodeTypeGetClassTypeGetObjectTypeGetDelegateTypeDeclareTypeTypeDeployCodeTypeActivateClassTypeDeactivateClassTypeUpdateClassTypeActivateObjectTypeActivateObjectDelegateTypeDeactivateObjectTypeUpdateObjectTypeRegisterChild"

var _MessageType_index = [...]uint16{0, 14, 33, 48, 59, 71, 84, 99, 114, 128, 145, 164, 179, 197, 223, 243, 259, 276}

func (i MessageType) String() string {
	if i >= MessageType(len(_MessageType_index)-1) {
		return "MessageType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _MessageType_name[_MessageType_index[i]:_MessageType_index[i+1]]
}
