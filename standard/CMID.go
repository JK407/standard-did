/*
 Copyright (C) BABEC. All rights reserved.
 Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package standard

// CMID 长安链身份认证go接口
// https://git.chainmaker.org.cn/contracts/standard/-/blob/master/draft/CM-CS-221221-Identity.md
type CMID interface {
	// Identities 获取该合约支持的所有认证类型
	// @return metas, 所有的认证类型编号和认证类型描述
	Identities() (metas []IdentityMeta)

	// SetIdentity 为地址设置认证类型，管理员可调用
	// @param address 必填，公钥/证书的地址。一个地址仅能绑定一个公钥和认证类型编号，重复输入则覆盖。
	// @param pkPem 选填,pem格式公钥，可用于验签
	// @param level 必填,认证类型编号
	// @param metadata 选填,其他信息，json格式字符串，比如：地址类型，上链人身份、组织信息，上链可信时间，上链批次等等
	// @return error 返回错误信息
	// @event topic: setIdentity(address, level, pkPem)
	SetIdentity(address, pkPem string, level int, metadata string) error

	// IdentityOf 获取认证信息
	// @param address 地址
	// @return int 返回当前认证类型编号
	// @return identity 认证信息
	// @return err 返回错误信息
	IdentityOf(address string) (identity Identity, err error)

	// LevelOf 获取认证编号
	// @param address 地址
	// @return level 返回当前认证类型编号
	// @return err 返回错误信息
	LevelOf(address string) (level int, err error)

	// EmitSetIdentityEvent 发送设置认证类型事件
	// @param address 地址
	// @param pkPem pem格式公钥
	// @param level 认证类型编号
	EmitSetIdentityEvent(address, pkPem string, level int)
}

type CMIDOption interface {
	// PkPemOf 获取公钥
	// @param address 地址
	// @return string 返回当前地址绑定的公钥
	// @return error 返回错误信息
	PkPemOf(address string) (string, error)

	// SetIdentityBatch 设置多个认证类型，管理员可调用
	// @param identities, 入参json格式字符串
	// @event topic: setIdentity(address, level, pkPem)
	SetIdentityBatch(identities []Identity) error

	// AlterAdminAddress 修改管理员，管理员可调用
	// @param adminAddresses 管理员地址，可为空，默认为创建人地址。入参为以逗号分隔的地址字符串"addr1,addr2"
	// @return error 返回错误信息
	// @event topic: alterAdminAddress（adminAddresses）
	AlterAdminAddress(adminAddresses string) error
}

// IdentityMeta 认证类型基础信息
type IdentityMeta struct {
	// Level 认证类型编号
	Level int `json:"level"`
	// Description 认证类型描述
	Description string `json:"description"`
}

// Identity 认证入参
type Identity struct {
	// Address 公钥地址
	Address string `json:"address"`
	// PkPem 公钥详情
	PkPem string `json:"pkPem"`
	// Level 认证类型编号
	Level int `json:"level"`
	// Metadata 其他，json格式字符串，可包含地址类型，上链人身份、组织信息，上链可信时间，上链批次等等
	Metadata string `json:"metadata"`
}

// IdentityMetadata 可选信息建议字段，若包含以下相关信息，建议采用以下字段
type IdentityMetadata struct {
	// AddressType 地址类型：0-chainmaker, 1-zxl, 2-ethereum，长安链默认是2
	AddressType string `json:"addressType"`
	// OrgId 组织ID
	OrgId string `json:"orgId"`
	// Role 上链人身份角色
	Role string `json:"role"`
	// Timestamp 可信存证时间
	Timestamp string `json:"timestamp"`
	// ProveTimestamp 可信存证时间证明
	ProveTimestamp string `json:"proveTimestamp"`
	// BatchId 批次ID
	BatchId string `json:"batchId"`
	// 其他自定义扩展字段
	// ...
}
