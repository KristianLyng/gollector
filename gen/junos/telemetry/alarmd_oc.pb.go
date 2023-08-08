// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: alarmd_oc.proto

package telemetry

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type SystemAlarm struct {
	Alarms               *SystemAlarmAlarmsType `protobuf:"bytes,151,opt,name=alarms" json:"alarms,omitempty"`
	XXX_NoUnkeyedLiteral struct{}               `json:"-"`
	XXX_unrecognized     []byte                 `json:"-"`
	XXX_sizecache        int32                  `json:"-"`
}

func (m *SystemAlarm) Reset()         { *m = SystemAlarm{} }
func (m *SystemAlarm) String() string { return proto.CompactTextString(m) }
func (*SystemAlarm) ProtoMessage()    {}
func (*SystemAlarm) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb073604d2df3a66, []int{0}
}
func (m *SystemAlarm) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SystemAlarm.Unmarshal(m, b)
}
func (m *SystemAlarm) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SystemAlarm.Marshal(b, m, deterministic)
}
func (m *SystemAlarm) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SystemAlarm.Merge(m, src)
}
func (m *SystemAlarm) XXX_Size() int {
	return xxx_messageInfo_SystemAlarm.Size(m)
}
func (m *SystemAlarm) XXX_DiscardUnknown() {
	xxx_messageInfo_SystemAlarm.DiscardUnknown(m)
}

var xxx_messageInfo_SystemAlarm proto.InternalMessageInfo

func (m *SystemAlarm) GetAlarms() *SystemAlarmAlarmsType {
	if m != nil {
		return m.Alarms
	}
	return nil
}

type SystemAlarmAlarmsType struct {
	Alarm                []*SystemAlarmAlarmsTypeAlarmList `protobuf:"bytes,151,rep,name=alarm" json:"alarm,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                          `json:"-"`
	XXX_unrecognized     []byte                            `json:"-"`
	XXX_sizecache        int32                             `json:"-"`
}

func (m *SystemAlarmAlarmsType) Reset()         { *m = SystemAlarmAlarmsType{} }
func (m *SystemAlarmAlarmsType) String() string { return proto.CompactTextString(m) }
func (*SystemAlarmAlarmsType) ProtoMessage()    {}
func (*SystemAlarmAlarmsType) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb073604d2df3a66, []int{0, 0}
}
func (m *SystemAlarmAlarmsType) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SystemAlarmAlarmsType.Unmarshal(m, b)
}
func (m *SystemAlarmAlarmsType) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SystemAlarmAlarmsType.Marshal(b, m, deterministic)
}
func (m *SystemAlarmAlarmsType) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SystemAlarmAlarmsType.Merge(m, src)
}
func (m *SystemAlarmAlarmsType) XXX_Size() int {
	return xxx_messageInfo_SystemAlarmAlarmsType.Size(m)
}
func (m *SystemAlarmAlarmsType) XXX_DiscardUnknown() {
	xxx_messageInfo_SystemAlarmAlarmsType.DiscardUnknown(m)
}

var xxx_messageInfo_SystemAlarmAlarmsType proto.InternalMessageInfo

func (m *SystemAlarmAlarmsType) GetAlarm() []*SystemAlarmAlarmsTypeAlarmList {
	if m != nil {
		return m.Alarm
	}
	return nil
}

type SystemAlarmAlarmsTypeAlarmList struct {
	Id                   *string                                  `protobuf:"bytes,151,opt,name=id" json:"id,omitempty"`
	State                *SystemAlarmAlarmsTypeAlarmListStateType `protobuf:"bytes,152,opt,name=state" json:"state,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                 `json:"-"`
	XXX_unrecognized     []byte                                   `json:"-"`
	XXX_sizecache        int32                                    `json:"-"`
}

func (m *SystemAlarmAlarmsTypeAlarmList) Reset()         { *m = SystemAlarmAlarmsTypeAlarmList{} }
func (m *SystemAlarmAlarmsTypeAlarmList) String() string { return proto.CompactTextString(m) }
func (*SystemAlarmAlarmsTypeAlarmList) ProtoMessage()    {}
func (*SystemAlarmAlarmsTypeAlarmList) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb073604d2df3a66, []int{0, 0, 0}
}
func (m *SystemAlarmAlarmsTypeAlarmList) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SystemAlarmAlarmsTypeAlarmList.Unmarshal(m, b)
}
func (m *SystemAlarmAlarmsTypeAlarmList) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SystemAlarmAlarmsTypeAlarmList.Marshal(b, m, deterministic)
}
func (m *SystemAlarmAlarmsTypeAlarmList) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SystemAlarmAlarmsTypeAlarmList.Merge(m, src)
}
func (m *SystemAlarmAlarmsTypeAlarmList) XXX_Size() int {
	return xxx_messageInfo_SystemAlarmAlarmsTypeAlarmList.Size(m)
}
func (m *SystemAlarmAlarmsTypeAlarmList) XXX_DiscardUnknown() {
	xxx_messageInfo_SystemAlarmAlarmsTypeAlarmList.DiscardUnknown(m)
}

var xxx_messageInfo_SystemAlarmAlarmsTypeAlarmList proto.InternalMessageInfo

func (m *SystemAlarmAlarmsTypeAlarmList) GetId() string {
	if m != nil && m.Id != nil {
		return *m.Id
	}
	return ""
}

func (m *SystemAlarmAlarmsTypeAlarmList) GetState() *SystemAlarmAlarmsTypeAlarmListStateType {
	if m != nil {
		return m.State
	}
	return nil
}

type SystemAlarmAlarmsTypeAlarmListStateType struct {
	Id                   *string  `protobuf:"bytes,51,opt,name=id" json:"id,omitempty"`
	Resource             *string  `protobuf:"bytes,53,opt,name=resource" json:"resource,omitempty"`
	Text                 *string  `protobuf:"bytes,52,opt,name=text" json:"text,omitempty"`
	TimeCreated          *uint64  `protobuf:"varint,54,opt,name=time_created,json=timeCreated" json:"time_created,omitempty"`
	Severity             *string  `protobuf:"bytes,55,opt,name=severity" json:"severity,omitempty"`
	TypeId               *string  `protobuf:"bytes,56,opt,name=type_id,json=typeId" json:"type_id,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SystemAlarmAlarmsTypeAlarmListStateType) Reset() {
	*m = SystemAlarmAlarmsTypeAlarmListStateType{}
}
func (m *SystemAlarmAlarmsTypeAlarmListStateType) String() string { return proto.CompactTextString(m) }
func (*SystemAlarmAlarmsTypeAlarmListStateType) ProtoMessage()    {}
func (*SystemAlarmAlarmsTypeAlarmListStateType) Descriptor() ([]byte, []int) {
	return fileDescriptor_bb073604d2df3a66, []int{0, 0, 0, 0}
}
func (m *SystemAlarmAlarmsTypeAlarmListStateType) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SystemAlarmAlarmsTypeAlarmListStateType.Unmarshal(m, b)
}
func (m *SystemAlarmAlarmsTypeAlarmListStateType) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SystemAlarmAlarmsTypeAlarmListStateType.Marshal(b, m, deterministic)
}
func (m *SystemAlarmAlarmsTypeAlarmListStateType) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SystemAlarmAlarmsTypeAlarmListStateType.Merge(m, src)
}
func (m *SystemAlarmAlarmsTypeAlarmListStateType) XXX_Size() int {
	return xxx_messageInfo_SystemAlarmAlarmsTypeAlarmListStateType.Size(m)
}
func (m *SystemAlarmAlarmsTypeAlarmListStateType) XXX_DiscardUnknown() {
	xxx_messageInfo_SystemAlarmAlarmsTypeAlarmListStateType.DiscardUnknown(m)
}

var xxx_messageInfo_SystemAlarmAlarmsTypeAlarmListStateType proto.InternalMessageInfo

func (m *SystemAlarmAlarmsTypeAlarmListStateType) GetId() string {
	if m != nil && m.Id != nil {
		return *m.Id
	}
	return ""
}

func (m *SystemAlarmAlarmsTypeAlarmListStateType) GetResource() string {
	if m != nil && m.Resource != nil {
		return *m.Resource
	}
	return ""
}

func (m *SystemAlarmAlarmsTypeAlarmListStateType) GetText() string {
	if m != nil && m.Text != nil {
		return *m.Text
	}
	return ""
}

func (m *SystemAlarmAlarmsTypeAlarmListStateType) GetTimeCreated() uint64 {
	if m != nil && m.TimeCreated != nil {
		return *m.TimeCreated
	}
	return 0
}

func (m *SystemAlarmAlarmsTypeAlarmListStateType) GetSeverity() string {
	if m != nil && m.Severity != nil {
		return *m.Severity
	}
	return ""
}

func (m *SystemAlarmAlarmsTypeAlarmListStateType) GetTypeId() string {
	if m != nil && m.TypeId != nil {
		return *m.TypeId
	}
	return ""
}

var E_JnprSystemAlarmExt = &proto.ExtensionDesc{
	ExtendedType:  (*JuniperNetworksSensors)(nil),
	ExtensionType: (*SystemAlarm)(nil),
	Field:         111,
	Name:          "jnpr_system_alarm_ext",
	Tag:           "bytes,111,opt,name=jnpr_system_alarm_ext",
	Filename:      "alarmd_oc.proto",
}

func init() {
	proto.RegisterType((*SystemAlarm)(nil), "system_alarm")
	proto.RegisterType((*SystemAlarmAlarmsType)(nil), "system_alarm.alarms_type")
	proto.RegisterType((*SystemAlarmAlarmsTypeAlarmList)(nil), "system_alarm.alarms_type.alarm_list")
	proto.RegisterType((*SystemAlarmAlarmsTypeAlarmListStateType)(nil), "system_alarm.alarms_type.alarm_list.state_type")
	proto.RegisterExtension(E_JnprSystemAlarmExt)
}

func init() { proto.RegisterFile("alarmd_oc.proto", fileDescriptor_bb073604d2df3a66) }

var fileDescriptor_bb073604d2df3a66 = []byte{
	// 337 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0xbb, 0x4e, 0x33, 0x31,
	0x10, 0x85, 0xb5, 0x9b, 0xcb, 0x9f, 0x7f, 0x36, 0x80, 0x64, 0x04, 0x31, 0x5b, 0x05, 0x44, 0x91,
	0x6a, 0x91, 0xc2, 0x55, 0xd0, 0x70, 0x11, 0x48, 0x50, 0x50, 0x6c, 0x2a, 0x2a, 0x6b, 0x95, 0x9d,
	0xc2, 0x90, 0x8d, 0x57, 0xf6, 0x04, 0xb2, 0x2d, 0xaf, 0x40, 0x01, 0x0f, 0x00, 0xef, 0xc5, 0xa3,
	0x20, 0xdb, 0x40, 0x42, 0x81, 0x44, 0x37, 0xf3, 0x8d, 0xe7, 0x9c, 0xe3, 0x81, 0xa5, 0x6c, 0x94,
	0xe9, 0x22, 0x17, 0x6a, 0x98, 0x94, 0x5a, 0x91, 0x8a, 0x97, 0x09, 0x47, 0x58, 0x20, 0xe9, 0x4a,
	0x90, 0x2a, 0x3d, 0xdc, 0x78, 0xab, 0x41, 0xdb, 0x54, 0x86, 0xb0, 0x10, 0xee, 0x3d, 0xeb, 0x43,
	0xd3, 0x15, 0x86, 0x3f, 0x07, 0xdd, 0xa0, 0x17, 0xf5, 0xd7, 0x92, 0xf9, 0x79, 0xe2, 0x87, 0x82,
	0xaa, 0x12, 0xd3, 0xcf, 0x97, 0xf1, 0x7b, 0x08, 0xd1, 0x1c, 0x67, 0x47, 0xd0, 0x70, 0xad, 0x95,
	0xa8, 0xf5, 0xa2, 0xfe, 0xe6, 0xaf, 0x12, 0xbe, 0x16, 0x23, 0x69, 0x28, 0xf5, 0x3b, 0xf1, 0x53,
	0x08, 0x30, 0xa3, 0x6c, 0x15, 0x42, 0x99, 0xfb, 0x2c, 0xff, 0x4f, 0x1b, 0x8f, 0xc7, 0x61, 0x2b,
	0x48, 0x43, 0x99, 0xb3, 0x0b, 0x68, 0x18, 0xca, 0x08, 0xf9, 0x8b, 0x8f, 0xb9, 0xf5, 0x17, 0x8f,
	0xc4, 0xad, 0xf8, 0xf0, 0x7e, 0x3d, 0x7e, 0x0d, 0x00, 0x66, 0x94, 0x2d, 0x3a, 0xbb, 0x6d, 0xeb,
	0xe6, 0x6c, 0x62, 0x68, 0x69, 0x34, 0x6a, 0xa2, 0x87, 0xc8, 0x77, 0x1d, 0xfd, 0xee, 0x19, 0x83,
	0x3a, 0xe1, 0x94, 0xf8, 0x8e, 0xe3, 0xae, 0x66, 0xeb, 0xd0, 0x26, 0x59, 0xa0, 0x18, 0x6a, 0xcc,
	0x08, 0x73, 0xbe, 0xd7, 0x0d, 0x7a, 0xf5, 0x34, 0xb2, 0xec, 0xcc, 0x23, 0x2b, 0x69, 0xf0, 0x1e,
	0xb5, 0xa4, 0x8a, 0xef, 0x7b, 0xc9, 0xaf, 0x9e, 0x75, 0xe0, 0x9f, 0x8d, 0x21, 0x64, 0xce, 0x0f,
	0xdc, 0xa8, 0x69, 0xdb, 0xcb, 0xfc, 0xf0, 0x06, 0x56, 0x6e, 0xc7, 0xa5, 0x16, 0xf3, 0x9f, 0x14,
	0xd6, 0xb0, 0x93, 0x5c, 0x4d, 0xc6, 0xb2, 0x44, 0x7d, 0x8d, 0xf4, 0xa0, 0xf4, 0x9d, 0x19, 0xe0,
	0xd8, 0x28, 0x6d, 0xb8, 0x72, 0x67, 0x59, 0xf8, 0x71, 0x96, 0x94, 0x59, 0x91, 0x81, 0x23, 0x27,
	0x16, 0x9c, 0x4f, 0xe9, 0x23, 0x00, 0x00, 0xff, 0xff, 0x54, 0x7d, 0x6a, 0xc4, 0x29, 0x02, 0x00,
	0x00,
}
