# did合约

分布式身份, Decentralized Identifiers (DIDs) v1.0

## 主要合约接口

参考： 长安链CMDID（CM-CS-231201-DID）存证合约标准实现：
https://git.chainmaker.org.cn/contracts/standard/-/blob/master/living/CM-CS-231201-DID.md

```go
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
SetVcTemplate(id string, name string, version string, template string) error
// GetVcTemplate 获取vc模板
GetVcTemplate(id string) (*VcTemplate, error)
// GetVcTemplateList 获取vc模板列表
GetVcTemplateList(nameSearch string, start int, count int) ([]*VcTemplate, error)
// EmitSetVcTemplateEvent 发送设置vc模板事件
EmitSetVcTemplateEvent(templateID string, templateName string, version string, vcTemplate string)

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

```

## cmc使用示例

命令行工具使用示例

```sh
echo ""
echo "create DOCKER_GO DID contract"
./cmc client contract user create --contract-name=DID --runtime-type=DOCKER_GO --byte-code-path=./did.7z --version=1.0 --sdk-conf-path=../config/sdk_config.yml \
--admin-key-file-paths=../config/node1/admin/admin1/admin1.key,../config/node2/admin/admin2/admin2.key,../config/node3/admin/admin3/admin3.key --gas-limit=999999999 --sync-result=true \
--params='{"didDocument":"{\"@context\":\"https://www.w3.org/ns/did/v1\",\"id\":\"did:cnbn:7d5e485e5fb34bc1846848c50c9eeb38e8ba62fe\",\"controller\":[\"did:cnbn:7d5e485e5fb34bc1846848c50c9eeb38e8ba62fe\"],\"verificationMethod\":[{\"id\":\"did:cnbn:7d5e485e5fb34bc1846848c50c9eeb38e8ba62fe#keys-1\",\"publicKeyPem\":\"-----BEGIN PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEBtUSf7SDTxemXSHKgIrblrzQM2xx\\n3mqoAA4vDTYm3txZ5lfnAB7DBGyAX5Qbap9QLcCrcCN56WGO5iGYN7Splg==\\n-----END PUBLIC KEY-----\\n\",\"controller\":\"did:cnbn:7d5e485e5fb34bc1846848c50c9eeb38e8ba62fe\",\"address\":\"7d5e485e5fb34bc1846848c50c9eeb38e8ba62fe\"}],\"authentication\":[\"did:cnbn:7d5e485e5fb34bc1846848c50c9eeb38e8ba62fe#keys-1\"],\"proof\":{\"type\":\"SM2Signature\",\"created\":\"2023-01-01T00:00:00Z\",\"proofPurpose\":\"verificationMethod\",\"verificationMethod\":\"did:cnbn:7d5e485e5fb34bc1846848c50c9eeb38e8ba62fe#keys-1\",\"proofValue\":\"MEUCIQDnzPad6d/PaEKJCW5OAZNuuY036+9OvcouQgSA7vlENQIgdoxpu3ZI/VKeBBGkPuiT+O6C3794sQCYD433b9qLDp0=\"}}"}'

echo ""
echo "upgrade DOCKER_GO DID contract"
./cmc client contract user upgrade --contract-name=DID --runtime-type=DOCKER_GO --byte-code-path=./did.7z --version=2.0 --sdk-conf-path=../config/sdk_config.yml --admin-key-file-paths=../config/node1/admin/admin1/admin1.key,../config/node2/admin/admin2/admin2.key,../config/node3/admin/admin3/admin3.key --gas-limit=999999999 --sync-result=true 

#echo ""
#echo "invoke SetAdmin"
#./cmc client contract user invoke --contract-name=DID --method=SetAdmin --sdk-conf-path=../config/sdk_config.yml --sync-result=true --gas-limit=99999999 --result-to-string=true --params='{"didDocument":"{\"id\":\"did:cnbn:ab108fc6c3850e01cee01e419d07f097186c3982\",\"controller\":[\"did:cnbn:ab108fc6c3850e01cee01e419d07f097186c3982\"],\"publicKey\":[{\"id\":\"did:cnbn:ab108fc6c3850e01cee01e419d07f097186c3982#keys-1\",\"publicKeyPem\":\"-----BEGIN PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEQHaJwSxa+1YwUsq5Cf4Wh+U8eDrM\\n4zJUrU3lRcH0hPIQyrgPp1m4rbdkNiJca3jRFQ2Av7EbDRDFZbnR/RHikw==\\n-----END PUBLIC KEY-----\\n\",\"address\":\"ab108fc6c3850e01cee01e419d07f097186c3982\"}],\"authentication\":[\"did:cnbn:ab108fc6c3850e01cee01e419d07f097186c3982#keys-1\"],\"proof\":{\"type\":\"SM2Signature\",\"created\":\"2023-01-01T00:00:00Z\",\"proofPurpose\":\"verificationMethod\",\"verificationMethod\":\"did:cnbn:ab108fc6c3850e01cee01e419d07f097186c3982#keys-1\",\"signatureValue\":\"MEQCIQCDJAQbdrQnZ6ar7XnyEH5yqQ7KqG6CTwfnAhTj4bD7/wIfO6FeEgVskX55Cu+DF75U+oZmyCUH0Olk4SsWU9Axdw==\"}}"}'

echo ""
echo "根据地址、公钥查询某个用户的DID"
./cmc client contract user get --contract-name=DID --method=GetDidByAddress --sdk-conf-path=../config/sdk_config.yml --result-to-string=true --params='{"address":"7d5e485e5fb34bc1846848c50c9eeb38e8ba62fe"}'
./cmc client contract user get --contract-name=DID --method=GetDidByPubkey --sdk-conf-path=../config/sdk_config.yml --result-to-string=true --params='{"pubKey":"-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEBtUSf7SDTxemXSHKgIrblrzQM2xx\n3mqoAA4vDTYm3txZ5lfnAB7DBGyAX5Qbap9QLcCrcCN56WGO5iGYN7Splg==\n-----END PUBLIC KEY-----\n"}'

echo "Admin2"
echo "添加新的DID文档到链上"
./cmc client contract user invoke --contract-name=DID --method=AddDidDocument --sdk-conf-path=../config/sdk_config.yml --sync-result=true --gas-limit=99999999 --result-to-string=true \
--params='{"didDocument":"{\"@context\":\"https://www.w3.org/ns/did/v1\",\"id\":\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7\",\"controller\":[\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7\"],\"verificationMethod\":[{\"id\":\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7#keys-1\",\"publicKeyPem\":\"-----BEGIN PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEhTNgKCYa5QCmchf3hCNX8e0Xz0gU\\nbVQ6r0bg47GA+zbHqPdDRx6ZZ0LuVK6Ojc2Td4NcO4udLsaTOV9R+4QZFw==\\n-----END PUBLIC KEY-----\\n\",\"controller\":\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7\",\"address\":\"5eb4e668952dcef3018a5bc03ca9517eff1cbfa7\"}],\"authentication\":[\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7#keys-1\"],\"proof\":{\"type\":\"SM2Signature\",\"created\":\"2023-01-01T00:00:00Z\",\"proofPurpose\":\"verificationMethod\",\"verificationMethod\":\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7#keys-1\",\"proofValue\":\"MEUCIQDqWhbQtdSCXF5tgal3cwbZOatcLMrtrHHiSqLF5k6zIQIgIAE684MAIbLbjr6MnzkH8kdhBo6jOgYkC8SjxE4KbGA=\"}}"}'

echo "Admin3"
echo "设置Issuer"
./cmc client contract user invoke --contract-name=DID --method=AddDidDocument --sdk-conf-path=../config/sdk_config.yml --sync-result=true --gas-limit=99999999 --result-to-string=true \
--params='{"didDocument":"{\"@context\":\"https://www.w3.org/ns/did/v1\",\"id\":\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a\",\"controller\":[\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a\"],\"verificationMethod\":[{\"id\":\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a#keys-1\",\"publicKeyPem\":\"-----BEGIN PUBLIC KEY-----\\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEH9bprLppfniFZHUcoPlco1PZg6iT\\nqTlk16kPVvXuNwhEWhCBnBAl0aDIHDZx2UTvZrH0Wn9QYJSPIJvUUepZsw==\\n-----END PUBLIC KEY-----\\n\",\"controller\":\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a\",\"address\":\"eadf82170c8d6f2ea9349f921be50967ba62b18a\"}],\"service\":[{\"id\":\"http://issuer.cnbn.org.cn\",\"type\":\"IssuerService\",\"serviceEndpoint\":\"http://issuer.cnbn.org.cn\"}],\"authentication\":[\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a#keys-1\"],\"proof\":{\"type\":\"SM2Signature\",\"created\":\"2023-01-01T00:00:00Z\",\"proofPurpose\":\"verificationMethod\",\"verificationMethod\":\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a#keys-1\",\"proofValue\":\"MEUCIQDT6ChI/e1M6uPjWJKO6MXBMUg1le5zawQCAflFLc8ykgIgbxZuameFanyI7OLdwOYj+3S4WnhN+rl1cDhIB01D3H8=\"}}"}'

./cmc client contract user invoke --contract-name=DID --method=AddTrustIssuer --sdk-conf-path=../config/sdk_config.yml --sync-result=true --gas-limit=99999999 --result-to-string=true --params='{"did":"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a"}'

echo "添加实名认证模板"
./cmc client contract user invoke --contract-name=DID --method=SetVcTemplate --sdk-conf-path=../config/sdk_config.yml --sync-result=true --gas-limit=99999999 --result-to-string=true \
--params='{"id":"1","name":"个人实名认证","version":"v1.0","template":"{\"$schema\":\"http://json-schema.org/draft-07/schema#\",\"type\":\"object\",\"properties\":{\"name\":{\"type\":\"string\"},\"idNumber\":{\"type\":\"string\"},\"phoneNumber\":{\"type\":\"string\"}},\"required\":[\"name\",\"idNumber\",\"phoneNumber\"],\"additionalProperties\":true}"}'


echo "验证DID有效性"

./cmc client contract user get --contract-name=DID --method=GetDidDocument --sdk-conf-path=../config/sdk_config.yml --result-to-string=true --params='{"did":"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a"}'

./cmc client contract user get --contract-name=DID --method=IsValidDid --sdk-conf-path=../config/sdk_config.yml --result-to-string=true --params='{"did":"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a"}'

echo "验证VC(admin3给admin2颁发的实名认证)"
./cmc client contract user get --contract-name=DID --method=VerifyVc --sdk-conf-path=../config/sdk_config.yml --result-to-string=true \
--params='{"vcJson":"{\"@context\":[\"https://www.w3.org/2018/credentials/v1\",\"https://www.w3.org/2018/credentials/examples/v1\"],\"id\":\"https://example.com/credentials/123\",\"type\":[\"VerifiableCredential\",\"IdentityCredential\"],\"issuer\":\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a\",\"issuanceDate\":\"2023-01-01T00:00:00Z\",\"expirationDate\":\"2042-01-01T00:00:00Z\",\"credentialSubject\":{\"id\":\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7\",\"idNumber\":\"511112188501010001\",\"name\":\"Devin\",\"phoneNumber\":\"13811888888\"},\"template\":{\"id\":\"1\",\"name\":\"个人实名认证\",\"version\":\"1.0\"},\"proof\":{\"type\":\"SM2Signature\",\"created\":\"2023-01-01T00:00:00Z\",\"proofPurpose\":\"assertionMethod\",\"verificationMethod\":\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a#keys-1\",\"proofValue\":\"MEUCIQCNZ7sSa4vcC03HYVMQdN/B3t1e25fnB3H6L77s3eGUZgIgHhFn84qtg/meCNNjDQKz+X/WUWKJSBmNK/b4ZIlytnM=\"}}"}'

echo "验证VP"
./cmc client contract user get --contract-name=DID --method=VerifyVp --sdk-conf-path=../config/sdk_config.yml --result-to-string=true \
--params='{"vpJson":"{\"@context\":[\"https://www.w3.org/2018/credentials/v1\",\"https://www.w3.org/2018/credentials/examples/v1\"],\"type\":\"VerifiablePresentation\",\"id\":\"https://example.com/presentations/123\",\"verifiableCredential\":[{\"@context\":[\"https://www.w3.org/2018/credentials/v1\",\"https://www.w3.org/2018/credentials/examples/v1\"],\"id\":\"https://example.com/credentials/123\",\"type\":[\"VerifiableCredential\",\"IdentityCredential\"],\"issuer\":\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a\",\"issuanceDate\":\"2023-01-01T00:00:00Z\",\"expirationDate\":\"2042-01-01T00:00:00Z\",\"credentialSubject\":{\"id\":\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7\",\"idNumber\":\"511112188501010001\",\"name\":\"Devin\",\"phoneNumber\":\"13811888888\"},\"template\":{\"id\":\"1\",\"name\":\"个人实名认证\",\"version\":\"1.0\"},\"proof\":{\"type\":\"SM2Signature\",\"created\":\"2023-01-01T00:00:00Z\",\"proofPurpose\":\"assertionMethod\",\"verificationMethod\":\"did:cnbn:eadf82170c8d6f2ea9349f921be50967ba62b18a#keys-1\",\"proofValue\":\"MEUCIQCNZ7sSa4vcC03HYVMQdN/B3t1e25fnB3H6L77s3eGUZgIgHhFn84qtg/meCNNjDQKz+X/WUWKJSBmNK/b4ZIlytnM=\"}}],\"presentationUsage\":\"租房\",\"expirationDate\":\"2024-01-01T00:00:00Z\",\"verifier\":\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1fa7\",\"proof\":{\"type\":\"SM2Signature\",\"created\":\"2023-01-01T00:00:00Z\",\"proofPurpose\":\"authentication\",\"challenge\":\"123\",\"verificationMethod\":\"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7#keys-1\",\"proofValue\":\"MEUCIFmfSg6HEOzECmh6svzRMddEiqY16C9GNCtMG72Yw1/lAiEA3SmYgypj3F9TodrVUN3t45xtv3jU7FfS56dYiwY5Sdk=\"}}"}'


echo "Delegate"
./cmc client contract user invoke --contract-name=DID --method=Delegate --sdk-conf-path=../config/sdk_config.yml --sync-result=true --gas-limit=99999999 --result-to-string=true \
--params='{"delegateeDid":"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7","resource":"https://www.w3.org/2018/credentials/examples/v1","action":"sign"}'

echo "查询代理"
./cmc client contract user get --contract-name=DID --method=GetDelegateList --sdk-conf-path=../config/sdk_config.yml --result-to-string=true \
--params='{"delegatorDid":"did:cnbn:7d5e485e5fb34bc1846848c50c9eeb38e8ba62fe","delegateeDid":"did:cnbn:5eb4e668952dcef3018a5bc03ca9517eff1cbfa7"}'

echo
echo "查询合约 DID.SupportStandard，是否支持合约标准CMBC"
./cmc client contract user get \
--contract-name=DID \
--method=SupportStandard \
--sdk-conf-path=../config/sdk_config.yml \
--params="{\"standardName\":\"CMBC\"}" \
--result-to-string=true

echo
echo "查询合约 DID.SupportStandard，是否支持合约标准CMDID"
./cmc client contract user get \
--contract-name=DID \
--method=SupportStandard \
--sdk-conf-path=../config/sdk_config.yml \
--params="{\"standardName\":\"CMDID\"}" \
--result-to-string=true

echo
echo "查询合约 DID.Standards，支持的合约标准列表"
./cmc client contract user get \
--contract-name=DID \
--method=Standards \
--sdk-conf-path=../config/sdk_config.yml \
--result-to-string=true
```

