/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/
/*
参照CMEVI合约标准实现：
https://git.chainmaker.org.cn/contracts/standard/-/blob/master/draft/CM-CS-221221-Evidence.md
*/

// Package main is the entry of the contract
package main

import (
	"did/standard"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"chainmaker.org/chainmaker/contract-sdk-go/v2/sdk"
	"github.com/buger/jsonparser"
	"github.com/xeipuuv/gojsonschema"
)

const (
	didMethod             = "cnbn"
	defaultDelegateAction = "sign"
	defaultSearchCount    = 1000
)

var (
	// MaxDateTime 最大时间，表示永不过期
	MaxDateTime   = int64(math.MaxInt64)
	errInvalidDid = errors.New("invalid did")
)

// 标记 DidContract 结构体实现 CMDID 接口
var _ standard.CMDID = (*DidContract)(nil)
var _ standard.CMDIDOption = (*DidContract)(nil)
var _ standard.CMBC = (*DidContract)(nil)

// DidContract 存证合约实现
type DidContract struct {
	dal *Dal
}

// NewDidContract 创建存证合约实例
func NewDidContract() *DidContract {
	return &DidContract{
		dal: &Dal{},
	}
}

// Standards  获取当前合约支持的标准协议列表
func (e *DidContract) Standards() []string {
	return []string{standard.ContractStandardNameCMDID, standard.ContractStandardNameCMBC}
}

// SupportStandard  获取当前合约是否支持某合约标准协议
func (e *DidContract) SupportStandard(standardName string) bool {
	return standardName == standard.ContractStandardNameCMDID || standardName == standard.ContractStandardNameCMBC
}

// InitAdmin 设置合约管理员
func (e *DidContract) InitAdmin(didJson string) error {
	didDoc := NewDIDDocument(didJson)
	if didDoc == nil {
		return errors.New("invalid did document")
	}
	err := e.addDidDocument(didDoc, false)
	if err != nil {
		return err
	}
	adminDid := didDoc.ID
	if err != nil {
		return err
	}
	err = e.dal.putAdmin(adminDid)
	if err != nil {
		return err
	}
	return nil
}

// SetAdmin 修改合约管理员
func (e *DidContract) SetAdmin(did string) error {
	//检查sender是否是admin
	if !e.isAdmin() {
		return errors.New("only admin can set admin")
	}
	//检查did是否有效
	valid, err := e.IsValidDid(did)
	if err != nil {
		return err
	}
	if !valid {
		return errInvalidDid
	}
	//保存admin
	err = e.dal.putAdmin(did)
	if err != nil {
		return err
	}
	return nil
}

// GetAdmin 获取合约管理员DID
func (e *DidContract) GetAdmin() (string, error) {
	return e.dal.getAdmin()
}

// DidMethod 获取DID Method
func (e *DidContract) DidMethod() string {
	return didMethod
}

// IsValidDid 判断DID URL是否合法
func (e *DidContract) IsValidDid(did string) (bool, error) {
	if len(did) < 9 {
		return false, errInvalidDid
	}
	//check did method
	if did[4:4+len(didMethod)] != didMethod {
		return false, errors.New("invalid did method")
	}
	//is did in black list
	if e.dal.isInBlackList(did) {
		return false, errors.New("did is in black list")
	}
	//检查DID Document是否存在
	didDocumentJson, err := e.dal.getDidDocument(did)
	if err != nil || len(didDocumentJson) == 0 {
		return false, errDidNotFound
	}
	return true, nil
}

func (e *DidContract) getDidDocument(did string) (*DIDDocument, error) {
	didDocumentJson, err := e.dal.getDidDocument(did)
	if err != nil || len(didDocumentJson) == 0 {
		return nil, errors.New("did document not found, did=" + did)
	}
	didDoc := NewDIDDocument(string(didDocumentJson))
	if didDoc == nil {
		return nil, errors.New("invalid did document")
	}
	return didDoc, nil
}

func (e *DidContract) verifyDidDocument(didDoc *DIDDocument) error {
	//检查DID Document有效性
	if didDoc == nil {
		return errors.New("invalid did document")
	}
	did, pubKeys, address, err := parsePubKeyAddress(didDoc)

	for _, pk := range pubKeys {
		//检查公钥是否存在
		dbDid, _ := e.dal.getDidByPubKey(pk)
		if len(dbDid) > 0 && dbDid != did {
			return errors.New("public key already exists")
		}
	}
	for _, addr := range address {
		//检查地址是否存在
		dbDid, _ := e.dal.getDidByAddress(addr)
		if len(dbDid) > 0 && dbDid != did {
			return errors.New("address already exists")
		}
	}

	if err != nil {
		return err
	}
	//check did method
	if did[4:4+len(didMethod)] != didMethod {
		return errors.New("invalid did method")
	}
	//check did document signature
	if didDoc.Proof == nil {
		return errors.New("invalid did document, need proof")
	}
	pass, err := didDoc.VerifySignature(func(_did string) (*DIDDocument, error) {
		//如果是DID用户自己签名，那么DID Document还没有上链，直接返回didDoc
		if _did == did {
			return didDoc, nil
		}
		return e.getDidDocument(_did)
	})

	if err != nil {
		return err
	}

	if !pass {
		return errors.New("invalid did document signature")
	}
	return nil
}

// AddDidDocument 添加DID Document
func (e *DidContract) AddDidDocument(didDocument string) error {
	didDoc := NewDIDDocument(didDocument)
	if didDoc == nil {
		return errors.New("invalid did document")
	}
	err := e.verifyDidDocument(didDoc)
	if err != nil {
		return err
	}
	//存储DID Document
	return e.addDidDocument(didDoc, true)
}
func (e *DidContract) addDidDocument(didDoc *DIDDocument, checkExist bool) error {
	if checkExist {
		//检查DID Document是否存在
		dbDidDoc, _ := e.dal.getDidDocument(didDoc.ID)
		if len(dbDidDoc) != 0 {
			return errors.New("did document already exists")
		}
	}
	did, pubKeys, addresses, err := parsePubKeyAddress(didDoc)
	if err != nil {
		return err
	}
	//在存储DID文档到状态数据库时，不需要Proof信息
	withoutProof := jsonparser.Delete(didDoc.rawData, proof)
	//压缩DID Document，去掉空格和换行符
	compactDidDoc, err := compactJson(withoutProof)
	if err != nil {
		return err
	}
	//Save did document
	err = e.dal.putDidDocument(did, compactDidDoc)
	if err != nil {
		return err
	}
	//Save pubkey index
	for _, pk := range pubKeys {
		err = e.dal.putIndexPubKey(pk, did)
		if err != nil {
			return err
		}
	}
	//save address index
	for _, addr := range addresses {
		err = e.dal.putIndexAddress(addr, did)
		if err != nil {
			return err
		}
	}
	// 发送事件
	e.EmitSetDidDocumentEvent(did, string(compactDidDoc))
	return nil
}

func parsePubKeyAddress(didDoc *DIDDocument) (didUrl string, pubKeys []string, addresses []string, err error) {
	pubKeys = make([]string, 0)
	addresses = make([]string, 0)
	for _, pk := range didDoc.VerificationMethod {
		pubKeys = append(pubKeys, pk.PublicKeyPem)
		addresses = append(addresses, pk.Address)
	}
	return didDoc.ID, pubKeys, addresses, nil

}

// GetDidDocument 获取DID Document
func (e *DidContract) GetDidDocument(did string) (string, error) {
	// check did valid
	valid, err := e.IsValidDid(did)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", errors.New("invalid did")
	}
	didDoc, err := e.dal.getDidDocument(did)
	if err != nil {
		return "", err
	}
	return string(didDoc), nil
}

// GetDidByPubkey 根据公钥获取DID
func (e *DidContract) GetDidByPubkey(pk string) (string, error) {
	//get did by pubkey
	did, err := e.dal.getDidByPubKey(pk)
	if err != nil {
		return "", err
	}

	return did, nil
}

// GetDidDocumentByPubkey 根据公钥获取DID Document
func (e *DidContract) GetDidDocumentByPubkey(pk string) (string, error) {
	//get did by pubkey
	did, err := e.dal.getDidByPubKey(pk)
	if err != nil {
		return "", err
	}
	//get did document
	return e.GetDidDocument(did)
}

// GetDidByAddress 根据地址获取DID
func (e *DidContract) GetDidByAddress(address string) (string, error) {
	//get did by address
	return e.dal.getDidByAddress(address)
}

// GetDidDocumentByAddress 根据地址获取DID Document
func (e *DidContract) GetDidDocumentByAddress(address string) (string, error) {
	//get did by address
	did, err := e.dal.getDidByAddress(address)
	if err != nil {
		return "", err
	}
	//get did document
	return e.GetDidDocument(did)
}

// VerifyVc 验证VC的有效性
func (e *DidContract) VerifyVc(vcJson string) (bool, error) {

	vc := NewVerifiableCredential(vcJson)
	if vc == nil {
		return false, errors.New("invalid vc")
	}
	if EnableVcIssueLog {
		//检查vcId是否在VcIssueLog表中
		vcIssueLogs, err := e.dal.searchVcIssueLogByVcID(vc.ID, 0, 1)
		if err != nil {
			return false, err
		}
		if len(vcIssueLogs) == 0 {
			return false, errors.New("vc is not issued")
		}
	}
	//检查vc拥有者是否在黑名单中
	if e.dal.isInBlackList(vc.GetCredentialSubjectID()) {
		return false, errors.New("vc owner is in black list")
	}
	// Check if the issuance date is before the expiration date
	issuanceDate, err := time.Parse(time.RFC3339, vc.IssuanceDate)
	if err != nil {
		return false, err
	}

	expirationDate, err := time.Parse(time.RFC3339, vc.ExpirationDate)
	if err != nil {
		return false, err
	}

	if issuanceDate.After(expirationDate) {
		return false, errors.New("issuance date is after the expiration date")
	}
	//检查当前时间是否在有效期内
	myTime, err := getTxTime()
	if err != nil {
		return false, err
	}
	if myTime < issuanceDate.Unix() || myTime > expirationDate.Unix() {
		return false, errors.New("vc is expired")
	}
	// Check if the VC type is correct
	if len(vc.Type) == 0 || vc.Type[0] != "VerifiableCredential" {
		return false, errors.New("invalid VC type")
	}
	//Check Issuer Validity
	if EnableTrustIssuer {
		err = e.checkIssuer(vc.Issuer)
		if err != nil {
			return false, err
		}
	}
	// Check  Signature
	pass, err := vc.VerifySignature(e.getDidDocument)
	if err != nil {
		return false, err
	}
	if !pass {
		return false, errors.New("invalid VC signature")
	}
	//检查vc template
	if vc.Template != nil {
		vcTemplate, err := e.dal.getVcTemplate(vc.Template.ID, vc.Template.Version)
		if err != nil {
			return false, err
		}
		if vcTemplate == nil {
			return false, errors.New("invalid VC template")
		}
		//检查vc template name
		if vcTemplate.Name != vc.Template.Name {
			return false, errors.New("invalid VC template name")
		}
		if vcTemplate.VcType != vc.Template.VcType {
			return false, errors.New("invalid VC type")
		}
		//检查vc template
		result, err := vc.VerifyVcTemplateContent(vcTemplate.Template)
		if err != nil {
			return false, err
		}
		if !result {
			return false, errors.New("credentialSubject of VC not match template")
		}
	}
	//检查是否被撤销
	if e.isInRevokeVcList(vc.ID) {
		return false, errors.New("vc is revoked")
	}
	return true, nil
}

func (e *DidContract) isInRevokeVcList(id string) bool {
	dbId, err := e.dal.getRevokeVc(id)
	if err != nil || len(dbId) == 0 {
		return false
	}
	return true
}

func (e *DidContract) checkIssuer(issuer string) error {
	//check if issuer is in trustIssuer list
	_, err := e.dal.getTrustIssuer(issuer)
	if err != nil {
		return err
	}
	return nil
}

func getTxTime() (int64, error) {
	timestamp, err := sdk.Instance.GetTxTimeStamp()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(timestamp, 10, 64)
}

// VerifyVp 验证VP的有效性
func (e *DidContract) VerifyVp(vpJson string) (bool, error) {
	myTime, err := getTxTime()
	if err != nil {
		return false, err
	}
	vp := NewVerifiablePresentation(vpJson)
	if vp == nil {
		return false, errors.New("invalid vp")
	}
	// Check if the VP type is correct
	if vp.Type != "VerifiablePresentation" {
		return false, errors.New("invalid VP type")
	}
	//验证亮证人是否在黑名单中
	userDid := vp.Proof.VerificationMethod[0:strings.Index(vp.Proof.VerificationMethod, "#")]
	if e.dal.isInBlackList(userDid) {
		return false, errors.New("vp owner is in black list")
	}
	// Validate all VCs in the VP
	for _, vc := range vp.VerifiableCredential {
		vcString, _ := json.Marshal(vc)
		//验证vc的有效性
		_, err = e.VerifyVc(string(vcString))
		if err != nil {
			return false, fmt.Errorf("invalid VC: %w", err)
		}
		//如果userDid和vc中的id不一致，则验证是否存在delegate，如果没有对应的delegate，则验证失败
		if userDid != vc.GetCredentialSubjectID() {
			//验证是否存在delegate
			delegates, err1 := e.dal.searchDelegate(userDid, vc.GetCredentialSubjectID(), vc.ID,
				defaultDelegateAction, 0, 0)
			if err1 != nil {
				return false, err1
			}
			if len(delegates) == 0 {
				return false, errors.New("no delegate")
			}
			//验证delegate是否过期
			hasDelegate := false
			for _, delegate := range delegates {
				if delegate.StartTime <= myTime && delegate.Expiration > myTime {
					hasDelegate = true
					break
				}
			}
			if !hasDelegate {
				return false, errors.New("delegate is expired")
			}
		}
	}

	// Validate the proof in the VP
	// In this example, we will only check the proof purpose
	if vp.Proof.ProofPurpose != "authentication" {
		return false, errors.New("invalid proof purpose")
	}

	// Validate the VP signature using the CheckJws function
	pass, err := vp.VerifySignature(e.getDidDocument)
	if err != nil {
		return false, err
	}
	if !pass {
		return false, errors.New("invalid vp signature")
	}
	return true, nil
}

// EmitSetDidDocumentEvent 发送设置DID Document事件
func (e *DidContract) EmitSetDidDocumentEvent(did string, didDocument string) {
	sdk.Instance.EmitEvent(standard.Topic_SetDidDocument, []string{did, didDocument})
}

// SetTrustRootList 设置信任根列表
func (e *DidContract) SetTrustRootList(dids []string) error {
	// check did valid
	for _, did := range dids {
		valid, err := e.IsValidDid(did)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("did not found")
		}
	}
	err := e.dal.putTrustRootList(dids)
	if err != nil {
		return err
	}
	e.EmitSetTrustRootListEvent(dids)
	return nil
}

// GetTrustRootList 获取信任根列表
func (e *DidContract) GetTrustRootList() (dids []string, err error) {
	return e.dal.getTrustRootList()
}

//func (e *DidContract) isInTrustRootList(did string) bool {
//	dids, err := e.dal.getTrustRootList()
//	if err != nil {
//		return false
//	}
//	for _, d := range dids {
//		if d == did {
//			return true
//		}
//	}
//	return false
//}

// EmitSetTrustRootListEvent 发送设置信任根列表事件
func (e *DidContract) EmitSetTrustRootListEvent(dids []string) {
	j, _ := json.Marshal(dids)
	sdk.Instance.EmitEvent(standard.Topic_SetTrustRootList, []string{string(j)})
}

// RevokeVc 撤销VC
func (e *DidContract) RevokeVc(vcID string) error {
	if !e.isAdmin() {
		return errors.New("only admin can revoke vc")
	}
	err := e.dal.putRevokeVc(vcID)
	if err != nil {
		return err
	}
	e.EmitRevokeVcEvent(vcID)
	return nil
}

// GetRevokedVcList 获取撤销VC列表
func (e *DidContract) GetRevokedVcList(vcIDSearch string, start int, count int) ([]string, error) {
	return e.dal.searchRevokeVc(vcIDSearch, start, count)
}

// EmitRevokeVcEvent 发送撤销VC事件
func (e *DidContract) EmitRevokeVcEvent(vcID string) {
	sdk.Instance.EmitEvent(standard.Topic_RevokeVc, []string{vcID})
}

// UpdateDidDocument 更新DID Document
func (e *DidContract) UpdateDidDocument(didDocument string) error {
	didDoc := NewDIDDocument(didDocument)
	if didDoc == nil {
		return errors.New("invalid did document")
	}
	//判断SenderDID是不是DID Document的创建者
	senderDid, err := e.getSenderDid()
	if err != nil {
		return err
	}
	if senderDid != didDoc.ID {
		if !e.isAdmin() {
			return errors.New("only admin or did owner can update did document")
		}
	}
	//检查新DID Document有效性
	err = e.verifyDidDocument(didDoc)
	if err != nil {
		return err
	}
	//检查DID Document是否存在
	did, pubKeys, addresses, err := parsePubKeyAddress(didDoc)
	if err != nil {
		return err
	}
	//根据DID查询已有的DID Document，并删除Index
	oldDidDocument, err := e.dal.getDidDocument(did)
	if err != nil {
		return err
	}
	oldDidDoc := NewDIDDocument(string(oldDidDocument))
	_, oldPubKeys, oldAddresses, _ := parsePubKeyAddress(oldDidDoc)
	//如果oldPubKeys在新的pubKeys中不存在，则删除
	for _, oldPk := range oldPubKeys {
		if !isInList(oldPk, pubKeys) {
			err = e.dal.deleteIndexPubKey(oldPk)
			if err != nil {
				return err
			}
		}
	}
	//如果oldAddresses在新的addresses中不存在，则删除
	for _, oldAddr := range oldAddresses {
		if !isInList(oldAddr, addresses) {
			err = e.dal.deleteIndexAddress(oldAddr)
			if err != nil {
				return err
			}
		}
	}
	//压缩DID Document，去掉空格和换行符
	compactDidDoc, err := compactJson([]byte(didDocument))
	if err != nil {
		return err
	}
	//保存新的DID Document
	err = e.dal.putDidDocument(did, compactDidDoc)
	if err != nil {
		return err
	}
	//保存新的pubKeys
	for _, pk := range pubKeys {
		if !isInList(pk, oldPubKeys) {
			err = e.dal.putIndexPubKey(pk, did)
			if err != nil {
				return err
			}
		}
	}
	//保存新的addresses
	for _, addr := range addresses {
		if !isInList(addr, oldAddresses) {
			err = e.dal.putIndexAddress(addr, did)
			if err != nil {
				return err
			}
		}
	}
	e.EmitSetDidDocumentEvent(did, didDocument)
	return nil
}

func isInList(pk string, keys []string) bool {
	for _, k := range keys {
		if k == pk {
			return true
		}
	}
	return false
}

// AddBlackList 添加黑名单
func (e *DidContract) AddBlackList(dids []string) error {
	if !e.isAdmin() {
		return errors.New("only admin can add black list")
	}
	for _, did := range dids {
		// check did valid
		valid, err := e.IsValidDid(did)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("did not found")
		}
		err = e.dal.putBlackList(did)
		if err != nil {
			return err
		}
	}
	e.EmitAddBlackListEvent(dids)
	return nil
}

//// 自定义错误类型，用于累积错误
//type AccumulatedError struct {
//	Errors []string
//}
//
//func (ae *AccumulatedError) Error() string {
//	return strings.Join(ae.Errors, "; ")
//}

// AddBlackList 添加黑名单
//func (e *DidContract) AddBlackList(dids []string) error {
//	if !e.isAdmin() {
//		return errors.New("only admin can add black list")
//	}
//
//	var accumulatedError AccumulatedError
//
//	for _, did := range dids {
//		valid, err := e.IsValidDid(did)
//		if err != nil {
//			accumulatedError.Errors = append(accumulatedError.Errors, fmt.Sprintf("Error checking validity of %s: %v", did, err))
//			continue // 继续处理下一个DID
//		}
//		if !valid {
//			accumulatedError.Errors = append(accumulatedError.Errors, fmt.Sprintf("DID not valid: %s", did))
//			continue // 继续处理下一个DID
//		}
//		err = e.dal.putBlackList(did)
//		if err != nil {
//			accumulatedError.Errors = append(accumulatedError.Errors, fmt.Sprintf("Error adding %s to blacklist: %v", did, err))
//			continue // 继续处理下一个DID
//		}
//	}
//
//	if len(accumulatedError.Errors) > 0 {
//		return &accumulatedError // 返回累积的错误
//	}
//
//	e.EmitAddBlackListEvent(dids)
//	return nil
//}

// DeleteBlackList 删除黑名单
func (e *DidContract) DeleteBlackList(dids []string) error {
	if !e.isAdmin() {
		return errors.New("only admin can delete black list")
	}
	for _, did := range dids {
		err := e.dal.deleteBlackList(did)
		if err != nil {
			return err
		}
	}
	e.EmitDeleteBlackListEvent(dids)
	return nil
}

// GetBlackList 获取黑名单
func (e *DidContract) GetBlackList(didSearch string, start int, count int) ([]string, error) {
	return e.dal.searchBlackList(didSearch, start, count)
}

// EmitAddBlackListEvent 发送添加黑名单事件
func (e *DidContract) EmitAddBlackListEvent(dids []string) {
	value, _ := json.Marshal(dids)
	sdk.Instance.EmitEvent(standard.Topic_AddBlackList, []string{string(value)})
}

// EmitDeleteBlackListEvent 发送删除黑名单事件
func (e *DidContract) EmitDeleteBlackListEvent(dids []string) {
	value, _ := json.Marshal(dids)
	sdk.Instance.EmitEvent(standard.Topic_DeleteBlackList, []string{string(value)})
}

// AddTrustIssuer 添加信任发行者
func (e *DidContract) AddTrustIssuer(dids []string) error {
	if !e.isAdmin() {
		return errors.New("only admin can add trust issuer")
	}
	for _, did := range dids {
		// check did valid
		valid, err := e.IsValidDid(did)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("did not found")
		}
		err = e.dal.putTrustIssuer(did)
		if err != nil {
			return err
		}
	}
	e.EmitAddTrustIssuerEvent(dids)
	return nil
}

// DeleteTrustIssuer 删除信任发行者
func (e *DidContract) DeleteTrustIssuer(dids []string) error {
	if !e.isAdmin() {
		return errors.New("only admin can delete trust issuer")
	}
	for _, did := range dids {
		// check did valid
		valid, err := e.IsValidDid(did)
		if err != nil {
			return err
		}
		if !valid {
			return errors.New("did not found")
		}
		err = e.dal.deleteTrustIssuer(did)
		if err != nil {
			return err
		}
	}
	e.EmitDeleteTrustIssuerEvent(dids)
	return nil
}

// GetTrustIssuer 获取信任发行者
func (e *DidContract) GetTrustIssuer(didSearch string, start int, count int) ([]string, error) {
	return e.dal.searchTrustIssuer(didSearch, start, count)
}

// EmitAddTrustIssuerEvent 发送添加信任发行者事件
func (e *DidContract) EmitAddTrustIssuerEvent(dids []string) {
	for _, did := range dids {
		sdk.Instance.EmitEvent(standard.Topic_AddTrustIssuer, []string{did})
	}
}

// EmitDeleteTrustIssuerEvent 发送删除信任发行者事件
func (e *DidContract) EmitDeleteTrustIssuerEvent(dids []string) {
	for _, did := range dids {
		sdk.Instance.EmitEvent(standard.Topic_DeleteTrustIssuer, []string{did})
	}
}

func (e *DidContract) getSenderDid() (string, error) {
	sender, err := sdk.Instance.Origin()
	if err != nil {
		return "", err
	}
	return e.dal.getDidByAddress(sender)
}

// Delegate 委托设置
func (e *DidContract) Delegate(delegateeDid string, resource string, action string, expiration int64) error {
	exp := MaxDateTime
	if expiration != 0 {
		exp = expiration
	}
	senderDid, err := e.getSenderDid()
	if err != nil {
		return err
	}
	myTime, err := getTxTime()
	if err != nil {
		return err
	}
	var delegate = &standard.DelegateInfo{
		DelegatorDid: senderDid,
		DelegateeDid: delegateeDid,
		Resource:     resource,
		Action:       action,
		StartTime:    myTime,
		Expiration:   exp,
	}
	err = e.dal.putDelegate(delegate)
	if err != nil {
		return err
	}
	e.EmitDelegateEvent(delegate.DelegatorDid, delegate.DelegateeDid, delegate.Resource, delegate.Action,
		delegate.StartTime, delegate.Expiration)
	return nil
}

// EmitDelegateEvent 发送委托事件
func (e *DidContract) EmitDelegateEvent(delegatorDid string, delegateeDid string, resource string, action string,
	start, expiration int64) {
	sdk.Instance.EmitEvent(standard.Topic_Delegate, []string{delegatorDid, delegateeDid, resource, action,
		strconv.FormatInt(start, 10), strconv.FormatInt(expiration, 10)})
}

// RevokeDelegate 撤销委托
func (e *DidContract) RevokeDelegate(delegateeDid string, resource string, action string) error {
	senderDid, err := e.getSenderDid()
	if err != nil {
		return err
	}
	err = e.dal.revokeDelegate(senderDid, delegateeDid, resource, action)
	if err != nil {
		return err
	}
	e.EmitRevokeDelegateEvent(senderDid, delegateeDid, resource, action)
	return nil
}

// EmitRevokeDelegateEvent 发送撤销委托事件
func (e *DidContract) EmitRevokeDelegateEvent(delegatorDid string, delegateeDid string,
	resource string, action string) {
	sdk.Instance.EmitEvent(standard.Topic_RevokeDelegate, []string{delegatorDid, delegateeDid, resource, action})
}

// GetDelegateList 获取委托列表
func (e *DidContract) GetDelegateList(delegatorDid, delegateeDid string, resource string, action string,
	start int, count int) ([]*standard.DelegateInfo, error) {
	return e.dal.searchDelegate(delegatorDid, delegateeDid, resource, action, start, count)
}

func checkTemplateValid(template string) error {
	//检查模板是否有效
	if len(template) == 0 {
		return errors.New("vc template is empty")
	}
	// 将JSON Schema字符串加载到gojsonschema.Schema
	schemaLoader := gojsonschema.NewStringLoader(template)
	_, err := gojsonschema.NewSchema(schemaLoader)
	return err
}

// SetVcTemplate 设置VC模板
func (e *DidContract) SetVcTemplate(id string, name string, vcType, version string, template string) error {
	if !e.isAdmin() {
		return errors.New("only admin can set vc template")
	}
	err := checkTemplateValid(template)
	if err != nil {
		return errors.New("invalid vc template: " + err.Error())
	}
	err = e.dal.putVcTemplate(id, name, vcType, version, template)
	if err != nil {
		return err
	}
	e.EmitSetVcTemplateEvent(id, name, vcType, version, template)
	return nil
}
func (e *DidContract) isAdmin() bool {
	senderDid, err := e.getSenderDid()
	if err != nil {
		return false
	}
	adminDid, err := e.dal.getAdmin()
	if err != nil {
		return false
	}
	return senderDid == adminDid
}

// GetVcTemplate 获取VC模板
func (e *DidContract) GetVcTemplate(id, version string) (*standard.VcTemplate, error) {
	return e.dal.getVcTemplate(id, version)
}

// GetVcTemplateList 获取VC模板列表
func (e *DidContract) GetVcTemplateList(templateNameSearch string, start int, count int) (
	[]*standard.VcTemplate, error) {
	return e.dal.searchVcTemplate(templateNameSearch, start, count)
}

// EmitSetVcTemplateEvent 发送设置VC模板事件
func (e *DidContract) EmitSetVcTemplateEvent(templateId, templateName, vcType, version, vcTemplate string) {
	sdk.Instance.EmitEvent(standard.Topic_SetVcTemplate,
		[]string{templateId, templateName, vcType, version, vcTemplate})
}

// VcIssueLog 记录VC签发日志
func (e *DidContract) VcIssueLog(issuer string, did string, templateId string, vcID string) error {
	//检查Issuer，did，templateId的有效性
	valid, err := e.IsValidDid(issuer)
	if err != nil || !valid {
		return errInvalidDid
	}
	valid, err = e.IsValidDid(did)
	if err != nil || !valid {
		return errInvalidDid
	}
	templates, err := e.dal.getVcTemplateById(templateId)
	if err != nil || len(templates) == 0 {
		return err
	}
	//保存VC签发日志
	err = e.dal.putVcIssueLog(issuer, did, templateId, vcID)
	if err != nil {
		return err
	}
	e.EmitVcIssueLogEvent(issuer, did, templateId, vcID)
	return nil
}

// GetVcIssueLogs 获取VC签发日志
func (e *DidContract) GetVcIssueLogs(issuer string, did string, templateId string, start int, count int) (
	[]*standard.VcIssueLog, error) {
	return e.dal.searchVcIssueLog(issuer, did, templateId, start, count)
}

// EmitVcIssueLogEvent 发送VC签发日志事件
func (e *DidContract) EmitVcIssueLogEvent(issuer string, did string, templateId string, vcID string) {
	sdk.Instance.EmitEvent(standard.Topic_VcIssueLog, []string{issuer, did, templateId, vcID})
}

// GetVcIssuers 获取VC签发者列表
func (e *DidContract) GetVcIssuers(did string) ([]string, error) {
	issueLogs, err := e.GetVcIssueLogs("", did, "", 0, 0)
	if err != nil {
		return nil, err
	}
	issuerDidMap := make(map[string]bool)
	for _, issueLog := range issueLogs {
		if _, ok := issuerDidMap[issueLog.Issuer]; !ok {
			issuerDidMap[issueLog.Issuer] = true
		}
	}
	//按Key排序,并返回
	issuerDid := make([]string, 0)
	for k := range issuerDidMap {
		issuerDid = append(issuerDid, k)
	}
	sort.Strings(issuerDid)
	return issuerDid, nil
}
