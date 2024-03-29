/*
 Copyright (C) BABEC. All rights reserved.
 Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package standard

// CMBC 长安链基础合约go接口
// https://git.chainmaker.org.cn/contracts/standard/-/blob/master/draft/CM-CS-221221-BC.md
type CMBC interface {
	// Standards  获取当前合约支持的标准协议列表
	// @return []string json格式字符串数组
	Standards() []string

	// SupportStandard  获取当前合约是否支持某合约标准协议
	// @return bool 存在：true，"true"；不存在：false，"false"
	SupportStandard(standardName string) bool
}
