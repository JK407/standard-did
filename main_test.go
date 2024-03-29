package main

import (
	"reflect"
	"strings"
	"testing"

	"chainmaker.org/chainmaker/contract-sdk-go/v2/sdk"
	"chainmaker.org/chainmaker/contract-utils/standard"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestInvokeContract(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockInstance := sdk.NewMockSDKInterface(ctrl)
	mockSdkInstance(mockInstance, t)
	adminPubKeyPem := getPubKeyPem("admin")
	mockInstance.EXPECT().GetSenderPk().AnyTimes().Return(string(adminPubKeyPem), nil)
	mockInstance.EXPECT().Sender().AnyTimes().Return(getAddressByName("admin"), nil)
	mockInstance.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes().Do(func(format string, args ...interface{}) {
		t.Logf(format, args...)
		t.Fatal(format)
	})
	sdk.Instance = mockInstance
	//didJson := generateDidDocument("admin", "admin")
	contract := &mockContractAll{}
	mainContract := &MainContract{c: contract}
	//result := mainContract.InvokeContract("DidMethod")
	//assert.Equal(t, didMethod, string(result.Payload))
	//err := contract.SetAdmin(didJson)
	//assert.NoError(t, err)
	//userDidJson := generateDidDocument("client1", "admin")
	//userDid, userPk, userAdd, _ := parsePubKeyAddress(NewDIDDocument(userDidJson))
	//userPkStr := strings.Join(userPk, "")
	//userAddStr := strings.Join(userAdd, "")
	//// GenerateVc
	//issuerDidJson := generateDidDocument("issuer", "admin")
	//err = contract.AddDidDocument(issuerDidJson)
	//assert.NoError(t, err)
	//initVcTemplate(contract, t)
	//_, err = contract.GetVcTemplate("1")
	//assert.NoError(t, err)
	//// GetVcTemplateList 获取VC模板列表
	//vctList, getVctList := contract.GetVcTemplateList("", 0, 10)
	//assert.NoError(t, getVctList)
	////t.Logf("vctList:%v", vctList)
	//assert.Equal(t, 1, len(vctList))
	//vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	//t.Logf("vcJson:%s", vcJson)
	//// GenerateVP
	//vpJson := generateVP("client1", vcJson, "实名登录", "challenge")
	//t.Logf("vpJson:%s", vpJson)
	////revoke vc
	//vc := NewVerifiableCredential(vcJson)
	//vcId := vc.ID
	//// 模拟GetArgs方法返回一个包含did键的map
	//vcIDSearch := "1"
	mockInstance.EXPECT().GetArgs().AnyTimes().Return(map[string][]byte{
		"didDocument":  []byte("userDidJson"),
		"did":          []byte("userDid"),
		"pubKey":       []byte("userPkStr"),
		"address":      []byte("userAddStr"),
		"vcJson":       []byte("vcJson"),
		"vpJson":       []byte("vpJson"),
		"vcID":         []byte("vcId"),
		"vcIDSearch":   []byte("vcIDSearch"),
		"delegateeDid": []byte("userDid"),
		"delegatorDid": []byte("userDid"),
		"resource":     []byte("vcID1"),
		"issuer":       []byte("userDid"),
		"didSearch":    []byte("did"),
		"action":       []byte(defaultDelegateAction),
		"id":           []byte("1"),
		"name":         []byte("name1"),
		"version":      []byte("1"),
		"template":     []byte("XXX"),
		"nameSearch":   []byte("XXX"),
		"standardName": []byte("CMDID"),
		"vcType":       []byte("ID"),
	})
	//sdk.Instance = mockInstance
	var f = func(method string) {
		defer func() {
			if r := recover(); r != nil {
				require.Equal(t, "implement me", r.(string))
			}
		}()

		mainContract.InvokeContract(method)
	}
	methods := GetInterfaceMethods((*DidContractAll)(nil))
	for _, method := range methods {
		t.Logf("method:%s", method)
		if !strings.HasPrefix(method, "Emit") && method != "InitAdmin" {
			if !EnableVcIssueLog && (method == "VcIssueLog" || method == "GetVcIssueLogs" || method == "GetVcIssuers") {
				continue
			}
			f(method)
		}
	}

	// IsValidDid
	//validResult := contract.InvokeContract("IsValidDid")
	//t.Logf("validResult:%v", string(validResult.Payload))
	//assert.Equal(t, "true", string(validResult.Payload))
	//// AddDidDocument
	//addResult := contract.InvokeContract("AddDidDocument")
	//t.Logf("addResult:%v", string(addResult.Payload))
	//assert.Equal(t, "ok", string(addResult.Payload))
	//// GetDidDocument
	//getResult := contract.InvokeContract("GetDidDocument")
	//t.Logf("getResult:%v", string(getResult.Payload))
	//assert.Equal(t, userDidJson, string(getResult.Payload))
	//// UpdateDidDocument
	//newUserDidJson := generateDidDocument("client1", "admin1")
	//t.Logf("NewDidJson:%s", newUserDidJson)
	//mockInstance.EXPECT().GetArgs().AnyTimes().Return(map[string][]byte{
	//	"didDocument": []byte(newUserDidJson),
	//})
	//sdk.Instance = mockInstance
	//updateResult := contract.InvokeContract("UpdateDidDocument")
	//t.Logf("updateResult:%v", string(updateResult.Payload))
	//assert.Equal(t, "ok", string(updateResult.Payload))
	//// GetDidByPubkey
	//getByPkResult := contract.InvokeContract("GetDidByPubkey")
	//t.Logf("getByPkResult:%v", string(getResult.Payload))
	//assert.Equal(t, userDid, string(getByPkResult.Payload))
	//// GetDidByAddress
	//getByAddResult := contract.InvokeContract("GetDidByAddress")
	//t.Logf("getByAddResult:%v", string(getByAddResult.Payload))
	//assert.Equal(t, userDid, string(getByAddResult.Payload))
	//// VerifyVc
	//verifyVcResult := contract.InvokeContract("VerifyVc")
	//t.Logf("verifyVcResult:%v", string(verifyVcResult.Payload))
	//assert.Equal(t, "true", string(verifyVcResult.Payload))
	//// VerifyVp
	//verifyVpResult := contract.InvokeContract("VerifyVp")
	//t.Logf("verifyVpResult:%v", string(verifyVpResult.Payload))
	//assert.Equal(t, "true", string(verifyVpResult.Payload))
	//// RevokeVc
	//revokeVcResult := contract.InvokeContract("RevokeVc")
	//t.Logf("revokeVcResult:%v", string(revokeVcResult.Payload))
	//assert.Equal(t, "true", string(verifyVpResult.Payload))
	//// GetRevokedVcList
	//getRevokedVcListResult := contract.InvokeContract("GetRevokedVcList")
	//t.Logf("getRevokedVcListResult:%v", string(getRevokedVcListResult.Payload))
	//assert.Equal(t, `["`+vcId+`"]`, string(getRevokedVcListResult.Payload))
	//// AddBlackList
	//addBlackListResult := contract.InvokeContract("AddBlackList")
	//t.Logf("addBlackListResult:%v", string(addBlackListResult.Payload))
	//assert.Equal(t, "ok", string(addBlackListResult.Payload))
	////// GetBlackList
	////getBlackListResult := c
}
func GetInterfaceMethods(i interface{}) []string {
	var methodNames []string

	// 获取接口的反射类型
	interfaceType := reflect.TypeOf(i).Elem()

	// 遍历接口的所有方法
	for i := 0; i < interfaceType.NumMethod(); i++ {
		// 获取方法的名称
		methodName := interfaceType.Method(i).Name
		methodNames = append(methodNames, methodName)
	}

	return methodNames
}

type mockContractAll struct {
}

func (m mockContractAll) DidMethod() string {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) IsValidDid(did string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) AddDidDocument(didDocument string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetDidDocument(did string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetDidByPubkey(pk string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetDidByAddress(address string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) VerifyVc(vcJson string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) VerifyVp(vpJson string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitSetDidDocumentEvent(did string, didDocument string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) RevokeVc(vcID string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetRevokedVcList(vcIDSearch string, start int, count int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitRevokeVcEvent(vcID string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) UpdateDidDocument(didDocument string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) AddBlackList(dids []string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) DeleteBlackList(dids []string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetBlackList(didSearch string, start int, count int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitAddBlackListEvent(dids []string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitDeleteBlackListEvent(dids []string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) SetTrustRootList(dids []string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetTrustRootList() (dids []string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitSetTrustRootListEvent(dids []string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) AddTrustIssuer(dids []string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) DeleteTrustIssuer(dids []string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetTrustIssuer(didSearch string, start int, count int) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitAddTrustIssuerEvent(dids []string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitDeleteTrustIssuerEvent(dids []string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) Delegate(delegateeDid string, resource string, action string, expiration int64) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitDelegateEvent(delegatorDid string, delegateeDid string, resource string, action string, start int64, expiration int64) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) RevokeDelegate(delegateeDid string, resource string, action string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitRevokeDelegateEvent(delegatorDid string, delegateeDid string, resource string, action string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetDelegateList(delegatorDid, delegateeDid string, resource string, action string, start int, count int) ([]*standard.DelegateInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) SetVcTemplate(id string, name string, vcType string, version string, template string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetVcTemplate(id, version string) (*standard.VcTemplate, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetVcTemplateList(nameSearch string, start int, count int) ([]*standard.VcTemplate, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitSetVcTemplateEvent(templateID string, templateName string, vcType string, version string, vcTemplate string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) VcIssueLog(issuer string, did string, templateID string, vcID string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetVcIssueLogs(issuer string, did string, templateID string, start int, count int) ([]*standard.VcIssueLog, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetVcIssuers(did string) (issuerDid []string, err error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) EmitVcIssueLogEvent(issuer string, did string, templateID string, vcID string) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) InitAdmin(didJson string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) SetAdmin(did string) error {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) GetAdmin() (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) Standards() []string {
	//TODO implement me
	panic("implement me")
}

func (m mockContractAll) SupportStandard(standardName string) bool {
	//TODO implement me
	panic("implement me")
}

var _ DidContractAll = (*mockContractAll)(nil)
