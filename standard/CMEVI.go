/*
 Copyright (C) BABEC. All rights reserved.
 Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package standard

const (
	Topic_Evidence = "Evidence"
)

// CMEVI 长安链存证合约go接口
// https://git.chainmaker.org.cn/contracts/standard/-/blob/master/draft/CM-CS-221221-Evidence.md
type CMEVI interface {
	// Evidence 存证
	// @param id 必填，流水号
	// @param hash 必填，上链哈希值
	// @param metadata 可选，其他信息；比如：哈希的类型（文字，文件）、文字描述的json格式字符串，具体参考下方 Metadata 对象。
	// @return error 返回错误信息
	Evidence(id string, hash string, metadata string) error

	// ExistsOfHash 哈希是否存在
	// @param hash 必填，上链的哈希值
	// @return exist 存在：true，"true"；不存在：false，"false"
	// @return err 错误信息
	ExistsOfHash(hash string) (exist bool, err error)

	// ExistsOfId ID是否存在
	// @param id 必填，上链的ID值
	// @return exist 存在：true，"true"；不存在：false，"false"
	// @return err 错误信息
	ExistsOfId(id string) (exist bool, err error)

	// FindByHash 根据哈希查找
	// @param hash 必填，上链哈希值
	// @return evidence 上链时传入的evidence信息
	// @return err 返回错误信息
	FindByHash(hash string) (evidence *Evidence, err error)

	// FindById 根据id查找
	// @param id 必填，流水号
	// @return evidence 上链时传入的evidence信息
	// @return err 返回错误信息
	FindById(id string) (evidence *Evidence, err error)

	// EmitEvidenceEvent 发送存证事件
	EmitEvidenceEvent(id string, hash string, metadata string) error
}

type CMEVIOption interface {

	// EvidenceBatch 批量存证
	// @param evidences 必填，存证信息
	// @return error 返回错误信息
	EvidenceBatch(evidences []Evidence) error

	// UpdateEvidence 根据ID更新存证哈希和metadata
	// @param id 必填，已经在链上存证的流水号。 如果是新流水号返回错误信息不存在
	// @param hash 必填，上链哈希值。必须与链上已经存储的hash不同
	// @param metadata 可选，其他信息；具体参考下方 Metadata 对象。
	// @return error 返回错误信息
	// @desc 该方法由长安链社区志愿者@sunhuiyuan提供建议，感谢参与
	UpdateEvidence(id string, hash string, metadata string) error

	// FindHisById 根据ID流水号查找存证历史(可以使用go合约接口：sdk.Instance.NewHistoryKvIterForKey或NewIterator实现)
	// @param id 必填，流水号
	// @return string 上链时传入的evidence信息的各个版本JSON数组对象。如果之前上链没有调用过updateEvidence、效果等同于findById，数组大小为1；
	//                如果之前上链调用过updateEvidence，则结果数组长度>1。
	// @return error 返回错误信息
	// @desc 该方法由长安链社区志愿者@sunhuiyuan提供建议，感谢参与
	FindHisById(id string) (evidence Evidence, err error)
}

// Evidence 存证结构体
type Evidence struct {
	// Id 业务流水号
	Id string `json:"id"`
	// Hash 哈希值
	Hash string `json:"hash"`
	// TxId 存证时交易ID
	TxId string `json:"txId"`
	// BlockHeight 存证时区块高度
	BlockHeight int `json:"blockHeight"`
	// Timestamp 存证时区块时间
	Timestamp string `json:"timestamp"`
	// Metadata 可选，其他信息；具体参考下方 Metadata 对象。
	Metadata string `json:"metadata"`
}

// Metadata 可选信息建议字段，若包含以下相关信息存证，请采用以下字段
type Metadata struct {
	// HashType 哈希的类型，文字、文件、视频、音频等
	HashType string `json:"hashType"`
	// HashAlgorithm 哈希算法，sha256、sm3等
	HashAlgorithm string `json:"hashAlgorithm"`
	// Username 存证人，用于标注存证的身份
	Username string `json:"username"`
	// Timestamp 可信存证时间
	Timestamp string `json:"timestamp"`
	// ProveTimestamp 可信存证时间证明
	ProveTimestamp string `json:"proveTimestamp"`
	// 存证内容
	Content string `json:"content"`
	// 其他自定义扩展字段
	// ...
}
