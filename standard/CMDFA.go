/*
 Copyright (C) BABEC. All rights reserved.
 Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package standard

import "chainmaker.org/chainmaker/contract-utils/safemath"

// Event top list
const (
	TopicTransfer = "transfer"
	TopicApprove  = "approve"
	TopicMint     = "mint"
	TopicBurn     = "burn"
)

// CMDFA 长安链同质化资产合约go接口
// https://git.chainmaker.org.cn/contracts/standard/-/blob/master/draft/CM-CS-221221-DFA.md
type CMDFA interface {

	// Name 查询Token的完整名称
	// @return name
	// @return err
	Name() (name string, err error)

	// Symbol 查询Token的简写符号
	// @return symbol
	// @return err
	Symbol() (symbol string, err error)

	// Decimals 查询Token支持的小数位数
	// @return decimals 返回支持的小数位数
	Decimals() (decimals uint8, err error)

	// TotalSupply 查询Token的发行总量
	// @return totalSupply 返回发行的Token总量
	TotalSupply() (totalSupply *safemath.SafeUint256, err error)

	// BalanceOf 查询账户的Token余额
	// @param account 指定要查询余额的账户
	// @return amount 返回指定账户的余额
	BalanceOf(account string) (amount *safemath.SafeUint256, err error)

	// Transfer 转账
	// @param to 收款账户
	// @param amount 转账金额
	// @return success 转账成功或失败
	// @return err 转账失败则返回Status：ERROR，Message具体错误
	Transfer(to string, amount *safemath.SafeUint256) error

	// TransferFrom 转账from账户下的指定amount金额给to账户
	// @param from 转出账户
	// @param to 转入账户
	// @param amount 转账金额
	// @return success 转账成功或失败
	// @return err 转账失败则返回Status：ERROR，Message具体错误
	TransferFrom(from string, to string, amount *safemath.SafeUint256) error

	// Approve 当前调用者授权指定的spender账户可以动用自己名下的amount金额给使用
	// @param spender 被授权账户
	// @param amount 授权使用的金额
	// @return success 授权成功或失败
	// @return error 授权失败则返回Status：ERROR，Message具体错误
	Approve(spender string, amount *safemath.SafeUint256) error

	// Allowance 查询owner授权多少额度给spender
	// @param owner 授权人账户
	// @param spender 被授权使用的账户
	// @return amount 返回授权金额
	Allowance(owner string, spender string) (amount *safemath.SafeUint256, err error)

	// EmitTransferEvent 发送转账事件
	// @param spender 转出账户
	// @param to 转入账户
	// @param amount 转账金额
	EmitTransferEvent(spender, to string, amount *safemath.SafeUint256)

	// EmitApproveEvent 发送授权事件
	// @param owner 授权人账户
	// @param spender 被授权使用的账户
	// @param amount 转账金额
	EmitApproveEvent(owner, spender string, amount *safemath.SafeUint256)
}

// CMDFAOption 可选的CMDFA接口
type CMDFAOption interface {
	// Mint 铸造发行新Token
	// @param account 发行到指定账户
	// @param amount 铸造发行新Token的数量
	// @return success 发行成功或失败
	// @return err 发行失败则返回Status：ERROR，Message具体错误
	Mint(account string, amount *safemath.SafeUint256) error

	// Burn 销毁当前用户名下指定数量的Token
	// @param amount 要销毁的Token数量
	// @return success 销毁成功或失败
	// @return err 销毁失败则返回Status：ERROR，Message具体错误
	Burn(amount *safemath.SafeUint256) error

	// BurnFrom 从spender账户名下销毁指定amount数量的Token
	// @param spender 被销毁Token的所属账户
	// @param amount 要销毁的数量
	// @return success 销毁成功或失败
	// @return err 销毁失败则返回Status：ERROR，Message具体错误
	BurnFrom(spender string, amount *safemath.SafeUint256) error

	// EmitMintEvent 触发铸造事件
	// @param account
	// @param amount
	EmitMintEvent(account string, amount *safemath.SafeUint256)

	// EmitBurnEvent 触发销毁事件
	// @param spender
	// @param amount
	EmitBurnEvent(spender string, amount *safemath.SafeUint256)

	// BatchTransfer 批量转账
	// @param to 收款账户
	// @param amount 转账金额,可以是1个表示每个to的金额相同，或多个（与to的数量相同）
	BatchTransfer(to []string, amount []*safemath.SafeUint256) error

	// Metadata 查询Token的元数据，其中包含Token的名称、符号、小数位数、LogoUrl、描述等信息
	Metadata() (metadata []byte, err error)
}
