/*
 Copyright (C) BABEC. All rights reserved.
 Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package standard

const (
	Topic_Register    = "Register"
	Topic_SetResolver = "SetResolver"
	Topic_Bind        = "Bind"
	Topic_UnBind      = "UnBind"
	Topic_ReverseBind = "ReverseBind"
)

// CMBNS 长安链区块链名称服务go接口，因为CMBNS合约也会支持CMNFA标准，所以不需要再单独定义Owner查询、Transfer转移域名等接口
type CMBNS interface {
	//CMNFA 嵌入的非同质化通证接口
	CMNFA
	// Domain 获取根域名
	Domain() string
	// Register 注册域名
	Register(domain string, owner string, metadata string, expirationTime int) error
	// Renew 续期域名
	Renew(domain string, expirationTime int) error
	// GetDomainInfo 获取域名信息
	GetDomainInfo(domain string) (*DomainInfo, error)
	// GetDomainList 获取域名列表
	GetDomainList(domainSearch string, owner string, start int, count int) ([]*DomainInfo, error)
	// EmitRegisterEvent 发送注册域名事件
	EmitRegisterEvent(domain string, owner string, metadata string, expirationTime int)
	// SetResolver 设置域名解析器
	SetResolver(domain string, resolver string) error
	// ResetResolver 重置域名解析器
	ResetResolver(domain string, resourceType string) error
	// EmitSetResolverEvent 发送设置域名解析器事件
	EmitSetResolverEvent(domain string, resolver string)
	// EmitReverseBindEvent 发送反向绑定事件
	EmitReverseBindEvent(value string, domain string, resourceType string)
	//CMBNSResolver 嵌入的接口
	CMBNSResolver
}

// CMBNSResolver 解析器合约应该满足的标准，只有满足这些接口的合约才能被设置为解析器
type CMBNSResolver interface {
	// Bind 绑定域名
	Bind(domain string, resolveValue string, resourceType string) error
	// Unbind 解绑域名
	Unbind(domain string, resourceType string) error
	// Resolve 解析域名
	Resolve(domain string, resourceType string) (string, error)
	// ReverseResolve 反向解析地址
	ReverseResolve(address string, resourceType string) (string, error)
	// GetBindList 按前缀匹配的方式搜索绑定列表
	// @param domainSearch 域名搜索关键字
	GetBindList(domainSearch string, start int, count int) ([]*DomainInfo, error)
	// EmitBindEvent 发送域名绑定事件
	EmitBindEvent(domain string, resolveValue string, resourceType string)
	// EmitUnBindEvent 发送域名解绑
	EmitUnBindEvent(domain string, resourceType string)
}

// DomainInfo 域名信息
type DomainInfo struct {
	// Domain 域名
	Domain string `json:"domain"`
	// ResolveValue 解析值
	ResolveValue string `json:"resolveValue"`
	// 绑定的类型
	ResourceType string `json:"resourceType"`
	// Owner 所有者
	Owner string `json:"owner"`
	// ExpirationTime 过期时间
	ExpirationTime int `json:"expirationTime"`
	// Resolver 解析器合约地址
	Resolver string `json:"resolver"`
	// Status 域名状态，正常，过期，禁用
	Status string `json:"status"`
	// Metadata 元数据
	Metadata string `json:"metadata"`
}

// CMBNSOption 可选的CMBNS接口
type CMBNSOption interface {
	// AddBlackList 添加黑名单
	AddBlackList(domains []string) error
	// DeleteBlackList 删除黑名单
	DeleteBlackList(domains []string) error
	// GetBlackList 获取黑名单
	GetBlackList(domainSearch string, start int, count int) ([]string, error)
	// EmitAddBlackListEvent 发送添加黑名单事件
	EmitAddBlackListEvent(domains []string)
	// EmitDeleteBlackListEvent 发送删除黑名单事件
	EmitDeleteBlackListEvent(domains []string)
}
