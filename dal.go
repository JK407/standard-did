package main

import (
	"crypto/sha256"
	"did/standard"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"chainmaker.org/chainmaker/contract-sdk-go/v2/sdk"
)

const (
	keyDid             = "d" // 此为存入数据库的世界状态key，故越短越好
	keyIndexPubKey     = "p"
	keyIndexAddress    = "a"
	keyTrustIssuer     = "ti"
	keyTrustRoot       = "tr"
	keyRevokeVc        = "r"
	keyBlackList       = "b"
	keyDelegate        = "g"
	keyVcTemplate      = "vt"
	keyAdmin           = "Admin"
	keyVcIssueLog      = "l"
	keyVcIndexIssueLog = "vl"
)

var (
	errDidNotFound      = errors.New("did not found")
	errTemplateNotFound = errors.New("template not found")
	errDataNotFound     = errors.New("data not found")
)

// Dal 数据库访问层
type Dal struct {
}

// Db 获取数据库实例
func (dal *Dal) Db() sdk.SDKInterface {
	return sdk.Instance
}
func processDid4Key(did string) string {
	if len(did) > 9 {
		did = did[9:] //去掉did:cnbn:
	}
	return strings.ReplaceAll(did, ":", "_")
}
func processPubKey4Key(pubKey string) string {
	hash := sha256.Sum256([]byte(pubKey))
	return hex.EncodeToString(hash[:])
}

func (dal *Dal) putDidDocument(did string, didDocument []byte) error {
	//将DID Document存入数据库
	err := dal.Db().PutStateByte(keyDid, processDid4Key(did), didDocument)
	if err != nil {
		return err
	}
	return nil
}
func (dal *Dal) getDidDocument(did string) ([]byte, error) {
	//从数据库中获取DID Document
	didDocument, err := dal.Db().GetStateByte(keyDid, processDid4Key(did))
	if err != nil {
		return nil, err
	}
	if len(didDocument) == 0 {
		return nil, errDidNotFound
	}
	return didDocument, nil
}

func (dal *Dal) putIndexPubKey(pubKey string, did string) error {
	//将索引存入数据库
	err := dal.Db().PutStateByte(keyIndexPubKey, processPubKey4Key(pubKey), []byte(did))
	if err != nil {
		return err
	}
	return nil
}
func (dal *Dal) deleteIndexPubKey(pubKey string) error {
	//从数据库中删除索引
	err := dal.Db().DelState(keyIndexPubKey, processPubKey4Key(pubKey))
	if err != nil {
		return err
	}
	return nil
}

func (dal *Dal) getDidByPubKey(pubKey string) (string, error) {
	//从数据库中获取索引
	did, err := dal.Db().GetStateByte(keyIndexPubKey, processPubKey4Key(pubKey))
	if err != nil {
		return "", err
	}
	if len(did) == 0 {
		return "", errDidNotFound
	}
	return string(did), nil
}

func (dal *Dal) putIndexAddress(address string, did string) error {
	//将索引存入数据库
	err := dal.Db().PutStateByte(keyIndexAddress, address, []byte(did))
	if err != nil {
		return err
	}
	return nil
}

func (dal *Dal) deleteIndexAddress(address string) error {
	//从数据库中删除索引
	err := dal.Db().DelState(keyIndexAddress, address)
	if err != nil {
		return err
	}
	return nil
}

func (dal *Dal) getDidByAddress(address string) (string, error) {
	//从数据库中获取索引
	did, err := dal.Db().GetStateByte(keyIndexAddress, address)
	if err != nil {
		return "", err
	}
	if len(did) == 0 {
		return "", errDidNotFound
	}
	return string(did), nil
}

func (dal *Dal) putTrustRootList(dids []string) error {
	//将TrustRootList一次性存入数据库
	values, _ := json.Marshal(dids)
	err := dal.Db().PutStateFromKeyByte(keyTrustRoot, values)
	if err != nil {
		return err
	}
	return nil
}

func (dal *Dal) getTrustRootList() ([]string, error) {
	//从数据库中获取TrustRootList
	dids, err := dal.Db().GetStateFromKeyByte(keyTrustRoot)
	if err != nil {
		return nil, err
	}
	var didsSlice []string
	_ = json.Unmarshal([]byte(dids), &didsSlice)
	return didsSlice, nil
}

func (dal *Dal) putTrustIssuer(did string) error {
	//将TrustIssuer存入数据库
	err := dal.Db().PutStateByte(keyTrustIssuer, processDid4Key(did), []byte(did))
	if err != nil {
		return err
	}
	return nil
}
func (dal *Dal) getTrustIssuer(did string) (string, error) {
	//从数据库中获取TrustIssuer
	didUrl, err := dal.Db().GetStateByte(keyTrustIssuer, processDid4Key(did))
	if err != nil || len(didUrl) == 0 {
		return "", errDataNotFound
	}
	return string(didUrl), nil
}
func (dal *Dal) deleteTrustIssuer(did string) error {
	//从数据库中删除TrustIssuer
	err := dal.Db().DelState(keyTrustIssuer, processDid4Key(did))
	if err != nil {
		return err
	}
	return nil
}
func (dal *Dal) searchTrustIssuer(didSearch string, start int, count int) ([]string, error) {
	//从数据库中查询RevokeVc迭代器
	iter, err := dal.Db().NewIteratorPrefixWithKeyField(keyTrustIssuer, processDid4Key(didSearch))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var dids []string
	i := 0
	if count == 0 {
		count = defaultSearchCount
	}
	for iter.HasNext() {
		_, _, value, err1 := iter.Next()
		if err1 != nil {
			return nil, err1
		}
		if i >= start+count {
			break
		}
		i++
		if i < start {
			continue
		}
		dids = append(dids, string(value))

	}
	return dids, nil
}

func processVcId(vcID string) string {
	//vcid 是一个http url，为了存入数据库，需要将其转换为一个只有字母大小写、数字、下划线的字符串
	vcID = strings.ReplaceAll(vcID, ":", "_")
	vcID = strings.ReplaceAll(vcID, "/", "_")
	vcID = strings.ReplaceAll(vcID, ".", "_")
	vcID = strings.ReplaceAll(vcID, "-", "_")
	return vcID
}

func (dal *Dal) putRevokeVc(vcID string) error {
	//将RevokeVc存入数据库
	err := dal.Db().PutStateByte(keyRevokeVc, processVcId(vcID), []byte(vcID))
	if err != nil {
		return err
	}
	return nil
}
func (dal *Dal) getRevokeVc(vcID string) (string, error) {
	//从数据库中获取RevokeVc
	vcIDUrl, err := dal.Db().GetStateByte(keyRevokeVc, processVcId(vcID))
	if err != nil || len(vcIDUrl) == 0 {
		return "", errDataNotFound
	}
	return string(vcIDUrl), nil
}

// searchRevokeVc 根据vcID前缀查询RevokeVc,start为起始位置从0开始，count为查询数量
func (dal *Dal) searchRevokeVc(vcIDSearch string, start int, count int) ([]string, error) {
	//从数据库中查询RevokeVc迭代器
	iter, err := dal.Db().NewIteratorPrefixWithKeyField(keyRevokeVc, processVcId(vcIDSearch))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var vcIDSlice []string
	i := 0
	if count == 0 {
		count = defaultSearchCount
	}
	for iter.HasNext() {
		_, _, value, err1 := iter.Next()
		if err1 != nil {
			return nil, err1
		}
		if i >= start+count {
			break
		}
		i++
		if i < start {
			continue
		}
		vcIDSlice = append(vcIDSlice, string(value))

	}
	return vcIDSlice, nil
}

func (dal *Dal) putBlackList(did string) error {
	//将BlackList存入数据库
	err := dal.Db().PutStateByte(keyBlackList, processDid4Key(did), []byte(did))
	if err != nil {
		return err
	}
	return nil
}
func (dal *Dal) isInBlackList(did string) bool {
	//从数据库中获取BlackList
	dbId, err := dal.Db().GetStateByte(keyBlackList, processDid4Key(did))
	if err != nil || len(dbId) == 0 {
		return false
	}
	return true
}
func (dal *Dal) deleteBlackList(did string) error {
	//从数据库中删除BlackList
	err := dal.Db().DelState(keyBlackList, processDid4Key(did))
	if err != nil {
		return err
	}
	return nil
}

// fixme 当start=1，count为2时返回的是三条数据
func (dal *Dal) searchBlackList(didSearch string, start int, count int) ([]string, error) {
	//从数据库中查询BlackList迭代器
	iter, err := dal.Db().NewIteratorPrefixWithKeyField(keyBlackList, processDid4Key(didSearch))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var didSlice []string
	//i := 0
	//if count == 0 {
	//	count = defaultSearchCount
	//}
	//for iter.HasNext() {
	//	_, _, value, err1 := iter.Next()
	//	if err1 != nil {
	//		return nil, err1
	//	}
	//	if i >= start+count {
	//		break
	//	}
	//	i++
	//	if i < start {
	//		continue
	//	}
	//	didSlice = append(didSlice, string(value))
	//
	//}

	// fixed
	i := 0         // 用于追踪当前迭代到的项
	collected := 0 // 用于追踪已收集的项的数量

	if count == 0 {
		count = defaultSearchCount
	}

	for iter.HasNext() {
		if collected >= count {
			break
		}
		_, _, value, err1 := iter.Next()
		if err1 != nil {
			return nil, err1
		}
		i++
		if i < start {
			continue
		}
		didSlice = append(didSlice, string(value))
		collected++
	}
	return didSlice, nil
}

func (dal *Dal) putDelegate(d *standard.DelegateInfo) error {
	//将Delegate存入数据库
	value, _ := json.Marshal(d)

	field := d.DelegatorDid + "_" + d.DelegateeDid + "_" + d.Resource + "_" + d.Action
	err := dal.Db().PutStateByte(keyDelegate, processVcId(field), value)
	if err != nil {
		return err
	}
	return nil
}
func (dal *Dal) searchDelegate(delegatorDid, delegateeDid, resource, action string, start int, count int) (
	[]*standard.DelegateInfo, error) {
	fieldPrefx := delegatorDid + "_"
	if len(delegateeDid) != 0 {
		fieldPrefx += delegateeDid + "_"
		if len(resource) != 0 {
			fieldPrefx += resource + "_"
			if len(action) != 0 {
				fieldPrefx += action
			}
		}
	}
	//从数据库中查询Delegate迭代器
	iter, err := dal.Db().NewIteratorPrefixWithKeyField(keyDelegate, processVcId(fieldPrefx))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var delegateSlice []*standard.DelegateInfo
	i := 0
	if count == 0 {
		count = defaultSearchCount
	}
	for iter.HasNext() {
		_, _, value, err1 := iter.Next()
		if err1 != nil {
			return nil, err1
		}
		if i >= start+count {
			break
		}
		i++
		if i < start {
			continue
		}
		var delegate standard.DelegateInfo
		_ = json.Unmarshal([]byte(value), &delegate)
		delegateSlice = append(delegateSlice, &delegate)
	}
	return delegateSlice, nil
}

func (dal *Dal) revokeDelegate(delegatorDid, delegateeDid string, resource string, action string) error {
	//从数据库中删除Delegate
	field := delegatorDid + "_" + delegateeDid + "_" + resource + "_" + action

	err := dal.Db().DelState(keyDelegate, processVcId(field))
	if err != nil {
		return err
	}
	return nil
}

func (dal *Dal) putVcTemplate(templateId string, templateName string, vcType, version string, vcTemplate string) error {
	//将VcTemplate存入数据库
	vcTemplateObj := standard.VcTemplate{
		Id:       templateId,
		Name:     templateName,
		VcType:   vcType,
		Template: vcTemplate,
		Version:  version,
	}
	value, _ := json.Marshal(vcTemplateObj)
	err := dal.Db().PutStateByte(keyVcTemplate, templateId+"_"+version, value)
	if err != nil {
		return err
	}
	return nil
}

func (dal *Dal) getVcTemplate(templateId, version string) (*standard.VcTemplate, error) {
	//从数据库中获取VcTemplate
	value, err := dal.Db().GetStateByte(keyVcTemplate, templateId+"_"+version)
	if err != nil || len(value) == 0 {
		return nil, errTemplateNotFound
	}
	var vcTemplateObj standard.VcTemplate
	_ = json.Unmarshal(value, &vcTemplateObj)
	return &vcTemplateObj, nil
}
func (dal *Dal) getVcTemplateById(templateId string) ([]*standard.VcTemplate, error) {
	//通过迭代器查询
	iter, err := dal.Db().NewIteratorPrefixWithKeyField(keyVcTemplate, templateId+"_")
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var vcTemplateSlice []*standard.VcTemplate
	for iter.HasNext() {
		_, _, value, err1 := iter.Next()
		if err1 != nil {
			return nil, err1
		}
		var vcTemplateObj standard.VcTemplate
		_ = json.Unmarshal(value, &vcTemplateObj)
		vcTemplateSlice = append(vcTemplateSlice, &vcTemplateObj)
	}
	return vcTemplateSlice, nil
}

func (dal *Dal) searchVcTemplate(templateNameSearch string, start int, count int) ([]*standard.VcTemplate, error) {
	//TODO:如果Name是中文，DockerGo不支持
	//从数据库中查询VcTemplate迭代器,在内存中进行Name模糊搜索过滤
	iter, err := dal.Db().NewIteratorPrefixWithKeyField(keyVcTemplate, "")
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var vcTemplateSlice []*standard.VcTemplate
	i := 0
	if count == 0 {
		count = defaultSearchCount
	}
	for iter.HasNext() {
		_, _, value, err1 := iter.Next()
		if err1 != nil {
			return nil, err1
		}
		if i >= start+count {
			break
		}
		var vcTemplateObj standard.VcTemplate
		_ = json.Unmarshal(value, &vcTemplateObj)
		if strings.Contains(vcTemplateObj.Name, templateNameSearch) {
			vcTemplateSlice = append(vcTemplateSlice, &vcTemplateObj)
			i++
		}
		if i < start {
			continue
		}
	}
	return vcTemplateSlice, nil
}

func (dal *Dal) putAdmin(admin string) error {
	//将Admin存入数据库
	err := dal.Db().PutStateFromKey(keyAdmin, admin)
	if err != nil {
		return err
	}
	return nil
}
func (dal *Dal) getAdmin() (string, error) {
	//从数据库中获取Admin
	admin, err := dal.Db().GetStateFromKey(keyAdmin)
	if err != nil || len(admin) == 0 {
		return "", errDataNotFound
	}
	return admin, nil
}
func (dal *Dal) putVcIssueLog(issuer string, holder string, templateId string, vcID string) error {
	myTime, err := getTxTime()
	if err != nil {
		return err
	}
	//将VcIssueLog存入数据库
	vcIssueLog := standard.VcIssueLog{
		Issuer:     issuer,
		Did:        holder,
		TemplateId: templateId,
		VcID:       vcID,
		IssueTime:  myTime,
	}
	//将VcIssueLog存入数据库,用VC持有人DID作为key，但是为了防止重复而覆盖历史数据，需要加上时间戳和hash
	value, _ := json.Marshal(vcIssueLog)
	hash := sha256.Sum256(value)
	err = dal.Db().PutStateByte(keyVcIssueLog, processDid4Key(holder)+fmt.Sprintf("-%d-%x", myTime, hash[:2]), value)
	if err != nil {
		return err
	}
	//将VcIssueLog存入数据库,用VC ID作为key，方便后续搜索
	err = dal.Db().PutStateByte(keyVcIndexIssueLog, processVcId(vcID), value)
	if err != nil {
		return err
	}
	return nil
}
func (dal *Dal) searchVcIssueLogByVcID(vcID string, start int, count int) ([]*standard.VcIssueLog, error) {
	//从数据库中查询VcIssueLog迭代器
	iter, err := dal.Db().NewIteratorPrefixWithKeyField(keyVcIndexIssueLog, processVcId(vcID))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var vcIssueLogSlice []*standard.VcIssueLog
	i := 0
	if count == 0 {
		count = defaultSearchCount
	}
	for iter.HasNext() {
		_, _, value, err1 := iter.Next()
		if err1 != nil {
			return nil, err1
		}
		if i >= start+count {
			break
		}
		i++
		if i < start {
			continue
		}
		var vcIssueLog standard.VcIssueLog
		_ = json.Unmarshal([]byte(value), &vcIssueLog)
		vcIssueLogSlice = append(vcIssueLogSlice, &vcIssueLog)
	}
	return vcIssueLogSlice, nil
}

func (dal *Dal) searchVcIssueLog(issuer string, did string, templateId string, start int, count int) (
	[]*standard.VcIssueLog, error) {
	//从数据库中查询VcIssueLog迭代器
	iter, err := dal.Db().NewIteratorPrefixWithKeyField(keyVcIssueLog, processDid4Key(did))
	if err != nil {
		return nil, err
	}
	defer iter.Close()
	var vcIssueLogSlice []*standard.VcIssueLog
	i := 0
	if count == 0 {
		count = defaultSearchCount
	}
	for iter.HasNext() {
		_, _, value, err1 := iter.Next()
		if err1 != nil {
			return nil, err1
		}
		if i >= start+count {
			break
		}
		i++
		if i < start {
			continue
		}
		var vcIssueLog standard.VcIssueLog
		_ = json.Unmarshal(value, &vcIssueLog)
		if len(issuer) != 0 && vcIssueLog.Issuer != issuer {
			i--
			continue
		}
		if len(templateId) != 0 && vcIssueLog.TemplateId != templateId {
			i--
			continue
		}
		vcIssueLogSlice = append(vcIssueLogSlice, &vcIssueLog)
	}
	return vcIssueLogSlice, nil
}
