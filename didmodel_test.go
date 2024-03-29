package main

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/common/v2/crypto/asym"
	"chainmaker.org/chainmaker/common/v2/evmutils"
)

func TestDIDDocument_VerifySignature(t *testing.T) {
	didDocumentJson := generateDidDocument("admin1", "admin1")
	didDocumentJson = strings.ReplaceAll(didDocumentJson, ",", ",\n")
	didDoc := NewDIDDocument(didDocumentJson)
	t.Logf("newDidDocument:\n%s\n", didDocumentJson)
	pass, err := didDoc.VerifySignature(func(did string) (*DIDDocument, error) {
		return didDoc, nil
	})
	if err != nil || !pass {
		t.Error("verify signature failed")
	}
}

func getPubKeyPem(name string) []byte {
	pem, _ := os.ReadFile("testdata/" + name + ".pem")
	return pem
}
func getPubKey(name string) crypto.PublicKey {
	pubKey, err := asym.PublicKeyFromPEM(getPubKeyPem(name))
	if err != nil {
		panic(err)
	}
	return pubKey
}
func getPrivateKey(name string) crypto.PrivateKey {
	pem, _ := os.ReadFile("testdata/" + name + ".key")
	privKey, err := asym.PrivateKeyFromPEM(pem, nil)
	if err != nil {
		panic(err)
	}
	return privKey
}
func getAddress(pk crypto.PublicKey) string {
	pkBytes, err := evmutils.MarshalPublicKey(pk)
	if err != nil {
		panic(err)
	}
	data := pkBytes[1:]
	bytesAddr := evmutils.Keccak256(data)
	addr := hex.EncodeToString(bytesAddr)[24:]
	return addr
}
func signDidDocument(didDocument *DIDDocument, privKey crypto.PrivateKey) string {
	didDocumentWithoutProof := *didDocument
	didDocumentWithoutProof.Proof = nil
	didDocBytes, err := json.Marshal(didDocumentWithoutProof)
	if err != nil {
		panic(err)
	}
	sig, err := privKey.Sign(didDocBytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(sig)
}
func signVC(vc *VerifiableCredential, privKey crypto.PrivateKey) string {
	didDocumentWithoutProof := *vc
	didDocumentWithoutProof.Proof = nil
	didDocBytes, err := json.Marshal(didDocumentWithoutProof)
	if err != nil {
		panic(err)
	}
	sig, err := privKey.Sign(didDocBytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(sig)
}
func signVP(vp *VerifiablePresentation, privKey crypto.PrivateKey) string {
	didDocumentWithoutProof := *vp
	didDocumentWithoutProof.Proof = nil
	didDocBytes, err := json.Marshal(didDocumentWithoutProof)
	if err != nil {
		panic(err)
	}
	sig, err := privKey.Sign(didDocBytes)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(sig)
}
func generateDidDocument(user, signer string) string {
	didDocTemp := `{
  "@context": "https://www.w3.org/ns/did/v1",
  "id": "%s",
  "verificationMethod": [
    {
      "id": "%s#keys-1",
      "type": "SM2VerificationKey2020",
      "controller": "%s",
      "publicKeyPem": "%s",
      "address": "%s"
    }
  ],
  "authentication": [
    "%s#keys-1"
  ],
  "controller":["%s"]
}`
	pubKeyPem := string(getPubKeyPem(user))
	userPubKey := getPubKey(user)
	userAddr := getAddress(userPubKey)
	userDid := "did:cnbn:" + userAddr

	pubKeyPem = strings.ReplaceAll(strings.ReplaceAll(pubKeyPem, "\n", "\\n"), "\r", "")
	didDocJson := fmt.Sprintf(didDocTemp, userDid, userDid, userDid, pubKeyPem, userAddr, userDid, userDid)
	didDoc := NewDIDDocument(didDocJson)
	if didDoc == nil {
		panic("generate did document failed")
	}
	signerPk := getPubKey(signer)
	signerAddr := getAddress(signerPk)
	signerPrvKey := getPrivateKey(signer)
	singerDid := "did:cnbn:" + signerAddr
	signature := signDidDocument(didDoc, signerPrvKey)
	proof := &Proof{
		Type:               "SM2Signature",
		Created:            "2023-01-01T00:00:00Z",
		ProofPurpose:       "verificationMethod",
		VerificationMethod: singerDid + "#keys-1",
		ProofValue:         signature,
	}
	proofj, _ := json.Marshal(proof)
	didDoc.Proof = proofj
	newDidDocument, _ := json.Marshal(didDoc)
	return string(newDidDocument)
}
func getDid(name string) string {
	addr := getAddressByName(name)
	return "did:cnbn:" + addr
}
func getAddressByName(name string) string {
	pubKey := getPubKey(name)
	addr := getAddress(pubKey)
	return addr
}

func generateVC(user, userName, id, phone, issuer string) string {
	vcTemp := `{
  "@context": [
    "https://www.w3.org/2018/credentials/v1",
    "https://www.w3.org/2018/credentials/examples/v1"
  ],
  "id": "https://example.com/credentials/123",
  "type": ["VerifiableCredential", "IdentityCredential"],
  "issuer": "%s",
  "issuanceDate": "2023-01-01T00:00:00Z",
  "expirationDate": "2042-01-01T00:00:00Z",
  "credentialSubject": {
    "id": "%s",
    "name": "%s",
    "idNumber": "%s",
    "phoneNumber": "%s"
  },
  "template": {
    "id": "1",
    "name": "个人实名认证",
	"version":"v1",
	"vcType":"ID"
  }
}`
	issuerDid := getDid(issuer)
	userDid := getDid(user)
	vcJson := fmt.Sprintf(vcTemp, issuerDid, userDid, userName, id, phone)
	vc := NewVerifiableCredential(vcJson)
	if vc == nil {
		panic("generate vc failed")
	}
	//签名
	signerPrvKey := getPrivateKey(issuer)
	signature := signVC(vc, signerPrvKey)
	vc.Proof = &Proof{
		Type:               "SM2Signature",
		Created:            "2023-01-01T00:00:00Z",
		ProofPurpose:       "assertionMethod",
		VerificationMethod: issuerDid + "#keys-1",
		ProofValue:         signature,
	}
	signedVC, _ := json.Marshal(vc)
	return string(signedVC)
}
func TestVerifiableCredential_VerifySignature(t *testing.T) {
	vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	vc := NewVerifiableCredential(vcJson)
	t.Logf("VC:%s", vcJson)
	didDocumentJson := generateDidDocument("issuer", "admin")
	didDoc := NewDIDDocument(didDocumentJson)
	pass, err := vc.VerifySignature(func(did string) (*DIDDocument, error) {
		return didDoc, nil
	})
	if err != nil || !pass {
		t.Error("verify signature failed")
	}
}
func generateVP(user string, vcJson string, usage string, challenge string) string {
	vpTemp := `{
  "@context": [
    "https://www.w3.org/2018/credentials/v1",
    "https://www.w3.org/2018/credentials/examples/v1"
  ],
  "type": "VerifiablePresentation",
  "id": "https://example.com/presentations/123",
  "verifiableCredential": [
    %s
  ],
  "presentationUsage": "%s",
  "expirationDate": "2024-01-01T00:00:00Z",
  "verifier": "%s"
}`
	userDid := getDid(user)
	vpJson := fmt.Sprintf(vpTemp, vcJson, usage, userDid)
	vp := NewVerifiablePresentation(vpJson)
	if vp == nil {
		panic("generate vp failed")
	}
	//签名
	signerPrvKey := getPrivateKey(user)
	signature := signVP(vp, signerPrvKey)
	vp.Proof = &Proof{
		Type:               "SM2Signature",
		Created:            "2023-01-01T00:00:00Z",
		ProofPurpose:       "authentication",
		VerificationMethod: userDid + "#keys-1",
		ProofValue:         signature,
		Challenge:          challenge,
	}
	signedVP, _ := json.Marshal(vp)
	return string(signedVP)
}
func TestVerifiablePresentation_VerifySignature(t *testing.T) {
	vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	vpJson := generateVP("client1", vcJson, "实名登录", "challenge")
	vp := NewVerifiablePresentation(vpJson)
	t.Logf("VP:%s", vpJson)
	didDocumentJson := generateDidDocument("client1", "admin")
	didDoc := NewDIDDocument(didDocumentJson)
	pass, err := vp.VerifySignature(func(did string) (*DIDDocument, error) {
		return didDoc, nil
	})
	if err != nil || !pass {
		t.Error("verify signature failed")
	}
}

func TestVcTemplateVerify(t *testing.T) {
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
	vcJson := generateVC("client1", "张三", "511112198811110011", "13800000000", "issuer")
	vc := NewVerifiableCredential(vcJson)
	pass, err := vc.VerifyVcTemplateContent(vcTemplate)
	if err != nil || !pass {
		t.Error("verify vc template failed")
	}
}

func TestDIDDocument_GetProofs(t *testing.T) {
	t.Run("TestNilProof", func(t *testing.T) {
		didDocumentJson := generateDidDocument("admin1", "admin1")
		didDocumentJson = strings.ReplaceAll(didDocumentJson, ",", ",\n")
		didDoc := NewDIDDocument(didDocumentJson)

		// Set proof to nil
		didDoc.Proof = nil

		didProof, err := parseProof(didDoc.Proof)
		assert.Error(t, err)
		didActualProof := didProof.Array

		didDocProof := didDoc.GetProofs()
		assert.Equal(t, didActualProof, didDocProof)
	})
	t.Run("TestSingleProof", func(t *testing.T) {
		didDocumentJson := generateDidDocument("admin1", "admin1")
		didDocumentJson = strings.ReplaceAll(didDocumentJson, ",", ",\n")
		didDoc := NewDIDDocument(didDocumentJson)
		didProof, err := parseProof(didDoc.Proof)
		assert.NoError(t, err)
		didActualProof := didProof.Single
		didDocProofList := didDoc.GetProofs()
		assert.Equal(t, didActualProof, didDocProofList[0])
	})
	t.Run("TestArrayProof", func(t *testing.T) {
		didDocumentJson := generateDidDocument("admin1", "admin1")
		didDocumentJson = strings.ReplaceAll(didDocumentJson, ",", ",\n")
		didDoc := NewDIDDocument(didDocumentJson)
		didDoc.Proof = []byte(`[{"type":"SM2Signature","created":"2023-01-01T00:00:00Z","proofPurpose":"verificationMethod","verificationMethod":"did:cnbn:783862c31c3a9d276657002b6bb1fd2139564eae#keys-1","signatureValue":"MEYCIQCEJ57bjl2xos/55f2Y3jEWIeWLvJ7sRBrCy1M1lvI09AIhAKeA1bIpactfikPkQOjBdmrMze8uDg/T7RtOat6Ebno9"}]`)
		didProof, err := parseProof(didDoc.Proof)
		assert.NoError(t, err)
		didActualProof := didProof.Array
		didDocProofList := didDoc.GetProofs()
		assert.Equal(t, didActualProof, didDocProofList)
	})
}
