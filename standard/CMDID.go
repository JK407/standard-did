/*
 Copyright (C) BABEC. All rights reserved.
 Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package standard

const (
	Topic_SetDidDocument    = "SetDidDocument"
	Topic_SetTrustRootList  = "SetTrustRootList"
	Topic_RevokeVc          = "RevokeVc"
	Topic_AddBlackList      = "AddBlackList"
	Topic_DeleteBlackList   = "DeleteBlackList"
	Topic_AddTrustIssuer    = "AddTrustIssuer"
	Topic_DeleteTrustIssuer = "DeleteTrustIssuer"
	Topic_Delegate          = "Delegate"
	Topic_RevokeDelegate    = "RevokeDelegate"
	Topic_SetVcTemplate     = "SetVcTemplate"
	Topic_VcIssueLog        = "VcIssueLog"
)

// CMDID 长安链DID
// https://git.chainmaker.org.cn/contracts/standard/-/blob/master/draft/CM-CS-221221-Identity.md
type CMDID interface {
	// DidMethod 获取DID方法
	DidMethod() string
	// IsValidDid 判断DID URL是否合法
	IsValidDid(did string) (bool, error)
	// AddDidDocument 添加DID文档
	AddDidDocument(didDocument string) error
	// GetDidDocument 根据DID URL获取DID文档
	GetDidDocument(did string) (string, error)
	// GetDidDocumentCreator 根据DID URL获取DID文档的创建者（一个DID URL）
	//GetDidDocumentCreator(did string) (creatorDid string, err error)

	// GetDidByPubkey 根据公钥获取DID URL
	GetDidByPubkey(pk string) (string, error)
	// GetDidByAddress 根据地址获取DID URL
	GetDidByAddress(address string) (string, error)
	// VerifyVc 验证vc
	VerifyVc(vcJson string) (bool, error)
	// VerifyVp 验证vp
	VerifyVp(vpJson string) (bool, error)
	// EmitSetDidDocumentEvent 发送添加DID文档事件
	EmitSetDidDocumentEvent(did string, didDocument string)

	// RevokeVc 撤销vc,撤销后的vc vp不能再被验证
	RevokeVc(vcID string) error
	// GetRevokedVcList 获取撤销vc列表
	GetRevokedVcList(vcIDSearch string, start int, count int) ([]string, error)
	// EmitRevokeVcEvent 发送撤销vc事件
	EmitRevokeVcEvent(vcID string)
}

type CMDIDOption interface {
	// UpdateDidDocument 更新DID文档
	UpdateDidDocument(didDocument string) error

	// AddBlackList 添加黑名单
	AddBlackList(dids []string) error
	// DeleteBlackList 删除黑名单
	DeleteBlackList(dids []string) error
	// GetBlackList 获取黑名单
	GetBlackList(didSearch string, start int, count int) ([]string, error)
	// EmitAddBlackListEvent 发送添加黑名单事件
	EmitAddBlackListEvent(dids []string)
	// EmitDeleteBlackListEvent 发送删除黑名单事件
	EmitDeleteBlackListEvent(dids []string)

	// SetTrustRootList 设置信任根列表
	SetTrustRootList(dids []string) error
	// GetTrustRootList 获取信任根列表
	GetTrustRootList() (dids []string, err error)
	// EmitSetTrustRootListEvent 发送设置信任根列表事件
	EmitSetTrustRootListEvent(dids []string)

	// AddTrustIssuer 添加信任的发行者
	AddTrustIssuer(dids []string) error
	// DeleteTrustIssuer 删除信任的发行者
	DeleteTrustIssuer(dids []string) error
	// GetTrustIssuer 获取信任的发行者
	GetTrustIssuer(didSearch string, start int, count int) ([]string, error)
	// EmitAddTrustIssuerEvent 发送添加信任的发行者事件
	EmitAddTrustIssuerEvent(dids []string)
	// EmitDeleteTrustIssuerEvent 发送删除信任的发行者事件
	EmitDeleteTrustIssuerEvent(dids []string)

	// Delegate 给delegateeDid授权delegatorDid的资源代理权限，在有效期内，delegateeDid可以代理delegatorDid对resource的action操作
	// @param delegateeDid 被授权者DID
	// @param resource 资源,一般是VcID
	// @param action 操作，一般是"issue"或"verify"
	// @param expiration 有效期，unix时间戳，0表示永久
	Delegate(delegateeDid string, resource string, action string, expiration int64) error
	// EmitDelegateEvent 发送授权事件
	EmitDelegateEvent(delegatorDid string, delegateeDid string, resource string, action string, start int64, expiration int64)
	// RevokeDelegate 撤销授权
	RevokeDelegate(delegateeDid string, resource string, action string) error
	// EmitRevokeDelegateEvent 发送撤销授权事件
	EmitRevokeDelegateEvent(delegatorDid string, delegateeDid string, resource string, action string)
	// GetDelegateList 查询授权列表
	GetDelegateList(delegatorDid, delegateeDid string, resource string, action string, start int, count int) ([]*DelegateInfo, error)

	// SetVcTemplate 设置vc模板
	SetVcTemplate(id string, name string, vcType string, version string, template string) error
	// GetVcTemplate 获取vc模板
	GetVcTemplate(id, version string) (*VcTemplate, error)
	// GetVcTemplateList 获取vc模板列表
	GetVcTemplateList(nameSearch string, start int, count int) ([]*VcTemplate, error)
	// EmitSetVcTemplateEvent 发送设置vc模板事件
	EmitSetVcTemplateEvent(templateID string, templateName string, vcType string, version string, vcTemplate string)

	// VcIssueLog 记录vc发行日志
	// @param issuer 必填，发行者DID
	// @param did 必填，vc持有者DID
	// @param templateID 选填，vc模板ID
	// @param vcID 必填，vcID或者vc hash
	VcIssueLog(issuer string, did string, templateID string, vcID string) error
	// GetVcIssueLogs 获取vc发行日志
	GetVcIssueLogs(issuer string, did string, templateID string, start int, count int) ([]*VcIssueLog, error)
	// GetVcIssuers 根据持证人DID获取vc发行者DID列表
	GetVcIssuers(did string) (issuerDid []string, err error)
	// EmitVcIssueLogEvent 发送vc发行日志事件
	EmitVcIssueLogEvent(issuer string, did string, templateID string, vcID string)
}

// VcIssueLog 记录vc发行日志
type VcIssueLog struct {
	// Issuer 发行者DID
	Issuer string `json:"issuer"`
	// Did vc持有者DID
	Did string `json:"did"`
	// TemplateId vc模板ID
	TemplateId string `json:"templateID"`
	// VcID vcID或者vc hash
	VcID string `json:"vcID"`
	// IssueTime 发行上链时间
	IssueTime int64 `json:"issueTime"`
}

// VcTemplate vc模板
type VcTemplate struct {
	// Id 模板ID
	Id string `json:"id"`
	// Name 模板名称
	Name string `json:"name"`
	// Version 模板版本
	Version string `json:"version"`
	// VcType vc类型
	VcType string `json:"vcType"`
	// Template 模板内容
	Template string `json:"template"`
}

// DelegateInfo 授权信息
type DelegateInfo struct {
	// DelegatorDid 授权者DID
	DelegatorDid string `json:"delegatorDid"`
	// DelegateeDid 被授权者DID
	DelegateeDid string `json:"delegateeDid"`
	// Resource 资源,一般是VcID
	Resource string `json:"resource"`
	// Action 操作，一般是"issue"或"verify"
	Action string `json:"action"`
	// StartTime 授权开始时间
	StartTime int64 `json:"startTime"`
	// Expiration 授权结束时间
	Expiration int64 `json:"expiration"`
}
