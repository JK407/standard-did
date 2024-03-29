package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"chainmaker.org/chainmaker/common/v2/serialize"
	"chainmaker.org/chainmaker/contract-sdk-go/v2/sdk"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type ResultSetKV struct {
	kv      []common.KeyValuePair
	current int
}

func (r *ResultSetKV) NextRow() (*serialize.EasyCodec, error) {
	panic("implement me")
}

func (r *ResultSetKV) HasNext() bool {
	return r.current < len(r.kv)
}

func (r *ResultSetKV) Close() (bool, error) {
	return true, nil
}

func (r *ResultSetKV) Next() (string, string, []byte, error) {
	if r.current >= len(r.kv) {
		return "", "", nil, nil
	}
	kv := r.kv[r.current]
	r.current++
	keys := strings.Split(kv.Key, "#")
	return keys[0], keys[1], kv.Value, nil
}

var _ sdk.ResultSetKV = (*ResultSetKV)(nil)

func mockSdkInstance(mockInstance *sdk.MockSDKInterface, t *testing.T) {
	var kv = &mockKv{
		kv: make(map[string][]byte),
	}
	mockInstance.EXPECT().GetStateByte(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(key, field string) ([]byte, error) {
		return kv.getStateByte(key, field)
	})
	mockInstance.EXPECT().PutStateByte(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(key, field string, value []byte) error {
		return kv.putStateByte(key, field, value)
	})
	mockInstance.EXPECT().PutStateFromKey(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(func(key string, value string) error {
		return kv.putStateByte(key, "", []byte(value))
	})
	mockInstance.EXPECT().GetStateFromKey(gomock.Any()).AnyTimes().DoAndReturn(func(key string) (string, error) {
		v, e := kv.getStateByte(key, "")
		return string(v), e
	})
	mockInstance.EXPECT().GetStateFromKeyByte(gomock.Any()).AnyTimes().DoAndReturn(func(key string) ([]byte, error) {
		return kv.getStateByte(key, "")
	})

	mockInstance.EXPECT().GetTxTimeStamp().AnyTimes().Return(fmt.Sprintf("%d", time.Now().Unix()), nil)
	mockInstance.EXPECT().EmitEvent(gomock.Any(), gomock.Any()).AnyTimes().Do(func(topic string, data []string) {
		t.Logf("emit event: Topic[%s], Data: %v", topic, data)
	})
	mockInstance.EXPECT().NewIteratorPrefixWithKeyField(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(key, field string) (sdk.ResultSetKV, error) {
			result := &ResultSetKV{kv: make([]common.KeyValuePair, 0)}
			for k, v := range kv.kv {
				if strings.HasPrefix(k, key+"#") {
					kv := common.KeyValuePair{
						Key:   k,
						Value: v,
					}
					result.kv = append(result.kv, kv)
				}
			}
			return result, nil
		})

	mockInstance.EXPECT().PutStateFromKeyByte(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(key string, value []byte) error {
			return kv.putStateByte(key, "", value)
		})

	mockInstance.EXPECT().DelState(gomock.Any(), gomock.Any()).AnyTimes().DoAndReturn(
		func(key, field string) error {
			return kv.delState(key, field)
		})
}

func TestGenerateDidDocument(t *testing.T) {
	names := []string{"admin1", "admin2", "client1", "issuer"}
	for _, name := range names {
		didJson := generateDidDocument(name, "admin1")
		t.Logf("%s did document:\n%s\n", name, didJson)
	}
}

func TestDidContract_AddDidDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)

	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")
	t.Logf("admin did document:\n%s\n", didJson)
	//didDoc := NewDIDDocument(didJson)
	//did := didDoc.ID
	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	require.NoError(t, err)
	userDidJson := generateDidDocument("client1", "client1")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	t.Logf("userDidJson:%s", userDidJson)
	//查询数据
	userDid, userPk, userAddr, _ := parsePubKeyAddress(NewDIDDocument(userDidJson))

	didDoc, err := contract.GetDidDocument(userDid)
	assert.NoError(t, err)
	t.Logf("DidDoc:%s", didDoc)
	var chainDid string
	testpk := "-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoEcz1UBgi0DQgAEHkAxpMxw5GtT+rD53MzSxtqxZZ54\nM1vQ2xoFKnwfftjsoMMZlqtHieu1QQEqJBVI0voMsajDrH8dW9mR/jZ7IQ==\n-----END PUBLIC KEY-----"
	chainDid, err = contract.GetDidByPubkey(testpk)
	assert.NoError(t, err)
	fmt.Println("chainDid:", chainDid)
	chainDid, err = contract.GetDidByPubkey(userPk[0])
	assert.NoError(t, err)
	assert.Equal(t, userDid, chainDid)
	chainDid, err = contract.GetDidByAddress(userAddr[0])
	assert.NoError(t, err)
	assert.Equal(t, userDid, chainDid)
}

type mockKv struct {
	kv map[string][]byte
}

func (kv *mockKv) getStateByte(key, field string) ([]byte, error) {
	return kv.kv[key+"#"+field], nil
}
func (kv *mockKv) putStateByte(key, field string, value []byte) error {
	kv.kv[key+"#"+field] = value
	return nil
}

func (kv *mockKv) delState(key string, field string) error {
	delete(kv.kv, key+"#"+field)
	return nil
}

func TestDidContract_VerifyVc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")

	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	assert.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	userDid := getDid("client1")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	issuerDidJson := generateDidDocument("issuer", "admin")
	err = contract.AddDidDocument(issuerDidJson)
	assert.NoError(t, err)
	issuerDid := getDid("issuer")
	// AddTrustIssuer 添加信任发行者
	addTrustIssuerErr := contract.AddTrustIssuer([]string{issuerDid})
	assert.NoError(t, addTrustIssuerErr)
	initVcTemplate(contract, t)
	_, err = contract.GetVcTemplate("1", "v1")
	assert.NoError(t, err)
	// GetVcTemplateList 获取VC模板列表
	vctList, getVctList := contract.GetVcTemplateList("", 0, 10)
	assert.NoError(t, getVctList)
	//t.Logf("vctList:%v", vctList)
	assert.Equal(t, 1, len(vctList))
	vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	t.Logf("vcJson:%s", vcJson)
	// VcIssueLog 记录VC签发日志

	err = contract.VcIssueLog(issuerDid, userDid, "1", "511112198811110011")
	assert.NoError(t, err)
	err = contract.VcIssueLog(issuerDid, userDid, "1", "511112198811110012")
	assert.NoError(t, err)
	// GetVcIssueLogs 获取VC签发日志
	vcIssueLogs, getVcIssueLogsErr := contract.GetVcIssueLogs(issuerDid, userDid, "1", 0, 10)
	assert.NoError(t, getVcIssueLogsErr)
	assert.Equal(t, 2, len(vcIssueLogs))
	t.Log("vcIssueLogs:", vcIssueLogs[0])
	pass, err := contract.VerifyVc(vcJson)
	assert.NoError(t, err)
	assert.True(t, pass)

	// GetVcIssuers 获取VC签发者列表
	vcIssuers, getVcIssuersErr := contract.GetVcIssuers(issuerDid)
	assert.NoError(t, getVcIssuersErr)
	t.Logf("vcIssuers:%v", vcIssuers)
	assert.Equal(t, []string{issuerDid}, vcIssuers)
}

func TestDidContract_VerifyVp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")

	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	assert.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	issuerDidJson := generateDidDocument("issuer", "admin")
	err = contract.AddDidDocument(issuerDidJson)
	assert.NoError(t, err)
	issuerDid := getDid("issuer")
	// AddTrustIssuer 添加信任发行者
	addTrustIssuerErr := contract.AddTrustIssuer([]string{issuerDid})
	assert.NoError(t, addTrustIssuerErr)
	initVcTemplate(contract, t)
	vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	pass, err := contract.VerifyVc(vcJson)
	assert.NoError(t, err)
	assert.True(t, pass)
	vpJson := generateVP("client1", vcJson, "实名登录", "challenge")
	t.Logf("vpJson:%s", vpJson)
	pass, err = contract.VerifyVp(vpJson)
	assert.NoError(t, err)
	assert.True(t, pass)
}

func initVcTemplate(contract *DidContract, t *testing.T) {
	vcTemplate := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "properties": {
    "name": {
      "type": "string"
    },
    "idNumber": {
      "type": "string"
    },
    "phoneNumber": {
      "type": "string"
    }
  },
  "required": ["name", "idNumber", "phoneNumber"],
  "additionalProperties": true
}`
	err := contract.SetVcTemplate("1", "个人实名认证", "ID", "v1", vcTemplate)
	assert.NoError(t, err)
}
func TestDidContract_RevokeVc(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")

	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	assert.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	issuerDidJson := generateDidDocument("issuer", "admin")
	err = contract.AddDidDocument(issuerDidJson)
	assert.NoError(t, err)
	issuerDid := getDid("issuer")
	// AddTrustIssuer 添加信任发行者
	addTrustIssuerErr := contract.AddTrustIssuer([]string{issuerDid})
	assert.NoError(t, addTrustIssuerErr)
	initVcTemplate(contract, t)
	vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	pass, err := contract.VerifyVc(vcJson)
	assert.NoError(t, err)
	assert.True(t, pass)
	vpJson := generateVP("client1", vcJson, "实名登录", "challenge")
	pass, err = contract.VerifyVp(vpJson)
	assert.NoError(t, err)
	assert.True(t, pass)
	//revoke vc
	vc := NewVerifiableCredential(vcJson)
	err = contract.RevokeVc(vc.ID)
	assert.NoError(t, err)
	err = contract.RevokeVc("fake vc id")
	assert.NoError(t, err)
	pass, err = contract.VerifyVp(vpJson)
	assert.False(t, pass)
	revokeVcList, err := contract.GetRevokedVcList("", 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(revokeVcList))
}
func TestDidContract_BlackList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")

	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	assert.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	issuerDidJson := generateDidDocument("issuer", "admin")
	err = contract.AddDidDocument(issuerDidJson)
	assert.NoError(t, err)
	issuerDid := getDid("issuer")
	// AddTrustIssuer 添加信任发行者
	addTrustIssuerErr := contract.AddTrustIssuer([]string{issuerDid})
	assert.NoError(t, addTrustIssuerErr)
	initVcTemplate(contract, t)
	vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	pass, err := contract.VerifyVc(vcJson)
	assert.NoError(t, err)
	assert.True(t, pass)
	vpJson := generateVP("client1", vcJson, "实名登录", "challenge")
	pass, err = contract.VerifyVp(vpJson)
	assert.NoError(t, err)
	assert.True(t, pass)
	//add black list
	userDid, _, _, _ := parsePubKeyAddress(NewDIDDocument(userDidJson))
	err = contract.AddBlackList([]string{userDid})
	assert.NoError(t, err)
	blackList, err := contract.GetBlackList("", 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(blackList))
	_, err = contract.GetDidDocument(userDid)
	assert.Error(t, err)
	pass, err = contract.VerifyVp(vpJson)
	assert.False(t, pass)
}
func TestDidContract_Delegate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")

	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	assert.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	issuerDidJson := generateDidDocument("issuer", "admin")
	err = contract.AddDidDocument(issuerDidJson)
	assert.NoError(t, err)
	issuerDid := getDid("issuer")
	// AddTrustIssuer 添加信任发行者
	addTrustIssuerErr := contract.AddTrustIssuer([]string{issuerDid})
	assert.NoError(t, addTrustIssuerErr)
	initVcTemplate(contract, t)
	vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	pass, err := contract.VerifyVc(vcJson)
	assert.NoError(t, err)
	assert.True(t, pass)
	//让Issuer代理client1出示VP，那么会失败，因为没有设置Delegate
	vpJson := generateVP("issuer", vcJson, "实名登录", "challenge")
	pass, err = contract.VerifyVp(vpJson)
	assert.Error(t, err)
	t.Log(err)
	assert.False(t, pass)
	//设置Delegate,让Issuer代理client1
	client1Pem := getPubKeyPem("client1")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(client1Pem), nil)
	vcID := "https://example.com/credentials/123"
	err = contract.Delegate(issuerDid, vcID, defaultDelegateAction, 0)
	pass, err = contract.VerifyVp(vpJson)
	assert.NoError(t, err)
	assert.True(t, pass)
	// GetDelegateList 获取委托列表
	clientDid := getDid("client1")
	delegateListPre, getDelegatePreErr := contract.GetDelegateList(clientDid, issuerDid, vcID, defaultDelegateAction, 0, 10)
	assert.NoError(t, getDelegatePreErr)
	t.Logf("delegateListPre:%v", delegateListPre)
	assert.Equal(t, 1, len(delegateListPre))
	// RevokeDelegate 撤销委托
	err = contract.RevokeDelegate(issuerDid, vcID, defaultDelegateAction)
	assert.NoError(t, err)
	pass, err = contract.VerifyVp(vpJson)
	assert.False(t, pass)
	// GetDelegateList 获取委托列表
	delegateListLast, getDelegateLastErr := contract.GetDelegateList(clientDid, issuerDid, vcID, defaultDelegateAction, 0, 10)
	assert.NoError(t, getDelegateLastErr)
	t.Logf("delegateListLast:%v", delegateListLast)
	assert.Equal(t, 0, len(delegateListLast))
}

// TestDidContract_GetDidDocument
// @Description 根据公钥或者地址获取did文档
// @Author Oberl-Fitzgerald 2024-01-04 14:11:17
// @Param  t *testing.T
func TestDidContract_GetDidDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")
	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	require.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	t.Logf("userDidJson:%s", userDidJson)
	_, userPk, userAddr, _ := parsePubKeyAddress(NewDIDDocument(userDidJson))
	//GetDidDocumentByPubkey 根据公钥获取did文档
	didDocByPk, GetByPkErr := contract.GetDidDocumentByPubkey(userPk[0])
	assert.NoError(t, GetByPkErr)
	//GetDidDocumentByAddress 根据地址获取did文档
	didDocByAd, GetByAdErr := contract.GetDidDocumentByAddress(userAddr[0])
	assert.NoError(t, GetByAdErr)
	assert.Equal(t, didDocByPk, didDocByAd)
}

// TestDidContract_TrustRootList
// @Description 信任根列表
// @Author Oberl-Fitzgerald 2024-01-04 14:31:02
// @Param  t *testing.T
func TestDidContract_TrustRootList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")
	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	require.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	t.Logf("userDidJson:%s", userDidJson)
	//查询数据
	userDid, _, _, _ := parsePubKeyAddress(NewDIDDocument(userDidJson))
	adminDid, _, _, _ := parsePubKeyAddress(NewDIDDocument(didJson))
	dids := []string{adminDid, userDid}
	t.Logf("dids:%v", dids)
	// SetTrustRootList 设置信任根列表
	setTrustRootListErr := contract.SetTrustRootList(dids)
	assert.NoError(t, setTrustRootListErr)
	// GetTrustRootList 获取信任根列表
	trustRootList, getTrustRootListErr := contract.GetTrustRootList()
	assert.NoError(t, getTrustRootListErr)
	assert.Equal(t, dids, trustRootList)
}

// TestDIDDocument_UpdateDidDocument
// @Description	更新DID Document
// @Author Oberl-Fitzgerald 2024-01-04 14:54:23
// @Param  t *testing.T
func TestDIDDocument_UpdateDidDocument(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")
	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	require.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	t.Logf("userDidJson:%s", userDidJson)
	//更新数据
	newUserDidJson := generateDidDocument("client1", "admin")
	t.Logf("NewDidJson:%s", newUserDidJson)
	err = contract.UpdateDidDocument(newUserDidJson)
	assert.NoError(t, err)
	//查询数据
	userDid, _, _, _ := parsePubKeyAddress(NewDIDDocument(newUserDidJson))
	didDoc, err := contract.GetDidDocument(userDid)
	assert.NoError(t, err)
	t.Logf("DidDoc:%s", didDoc)
	assert.Equal(t, newUserDidJson, didDoc)
}

// TestDidContract_DeleteBlackList
// @Description 删除黑名单
// @Author Oberl-Fitzgerald 2024-01-04 15:25:09
// @Param  t *testing.T
func TestDidContract_DeleteBlackList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")

	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	assert.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	issuerDidJson := generateDidDocument("issuer", "admin")
	err = contract.AddDidDocument(issuerDidJson)
	assert.NoError(t, err)
	initVcTemplate(contract, t)
	issuerDid := getDid("issuer")
	// AddTrustIssuer 添加信任发行者
	addTrustIssuerErr := contract.AddTrustIssuer([]string{issuerDid})
	assert.NoError(t, addTrustIssuerErr)
	vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	pass, err := contract.VerifyVc(vcJson)
	assert.NoError(t, err)
	assert.True(t, pass)
	vpJson := generateVP("client1", vcJson, "实名登录", "challenge")
	pass, err = contract.VerifyVp(vpJson)
	assert.NoError(t, err)
	assert.True(t, pass)
	//add black list
	userDid, _, _, _ := parsePubKeyAddress(NewDIDDocument(userDidJson))
	err = contract.AddBlackList([]string{userDid})
	assert.NoError(t, err)
	blackList, err := contract.GetBlackList("", 0, 10)
	t.Logf("blackList:%v", blackList)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(blackList))
	//delete black list
	err = contract.DeleteBlackList([]string{userDid})
	assert.NoError(t, err)
	blackList, err = contract.GetBlackList("", 0, 10)
	t.Logf("blackList:%v", blackList)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(blackList))
}

// TestDidContract_Issuer
// @Description 信任发行者
// @Author Oberl-Fitzgerald 2024-01-04 15:36:03
// @Param  t *testing.T
func TestDidContract_Issuer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")
	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	assert.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	t.Logf("userDidJson:%s", userDidJson)
	//查询数据
	userDid, _, _, _ := parsePubKeyAddress(NewDIDDocument(userDidJson))
	adminDid, _, _, _ := parsePubKeyAddress(NewDIDDocument(didJson))
	dids := []string{adminDid, userDid}
	t.Logf("dids:%v", dids)
	// AddTrustIssuer 添加信任发行者
	addTrustIssuerErr := contract.AddTrustIssuer(dids)
	assert.NoError(t, addTrustIssuerErr)
	// GetTrustIssuer 获取信任发行者
	oldTrustIssuer, getOldTrustIssuerErr := contract.GetTrustIssuer("", 0, 10)
	assert.NoError(t, getOldTrustIssuerErr)
	t.Logf("oldTrustIssuer:%v", oldTrustIssuer)
	//assert.Equal(t, dids, oldTrustIssuer) // fix me 无序集合
	assert.Equal(t, 2, len(oldTrustIssuer))
	// DeleteTrustIssuer 删除信任发行者
	deleteTrustIssuerErr := contract.DeleteTrustIssuer(dids)
	assert.NoError(t, deleteTrustIssuerErr)
	newTrustIssuer, getNewTrustIssuerErr := contract.GetTrustIssuer("", 0, 10)
	assert.NoError(t, getNewTrustIssuerErr)
	t.Logf("newTrustIssuer:%v", newTrustIssuer)
	assert.Equal(t, 0, len(newTrustIssuer))
}

func TestDidContract_DidMethod(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Origin().AnyTimes().Return(getAddressByName("admin"), nil)
	sdk.Instance = mockInstance
	didJson := generateDidDocument("admin", "admin")
	contract := &DidContract{dal: &Dal{}}
	err := contract.InitAdmin(didJson)
	require.NoError(t, err)
	userDidJson := generateDidDocument("client1", "admin")
	err = contract.AddDidDocument(userDidJson)
	assert.NoError(t, err)
	// DidMethod 获取DID Method
	method := contract.DidMethod()
	t.Logf("didMethod:%s", method)
	assert.Equal(t, method, didMethod)
}
func TestCheckTemplateValid(t *testing.T) {
	validTemplate := `{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type": "object",
		"properties": {
			"name": {
				"type": "string"
			},
			"age": {
				"type": "integer"
			}
		},
		"required": ["name", "age"]
	}`

	invalidTemplate := `{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
			},
			"age": {
				"type": "integer"
			}
		},
		"required": ["name", "age"]
	}`

	t.Run("valid template", func(t *testing.T) {
		err := checkTemplateValid(validTemplate)
		assert.NoError(t, err, "no error expected")
	})

	t.Run("invalid template", func(t *testing.T) {
		err := checkTemplateValid(invalidTemplate)
		assert.Error(t, err, "error expected")
	})

	t.Run("empty template", func(t *testing.T) {
		err := checkTemplateValid("")
		assert.Error(t, err, "error expected")
	})
}
