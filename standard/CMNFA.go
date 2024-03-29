/*
 Copyright (C) BABEC. All rights reserved.
 Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

 SPDX-License-Identifier: Apache-2.0
*/

package standard

import "chainmaker.org/chainmaker/contract-utils/safemath"

// CMNFA Chainmaker NFA standard interface
// https://git.chainmaker.org.cn/contracts/standard/-/blob/master/draft/CM-CS-221221-NFA.md
type CMNFA interface {
	// Mint a token, Obligatory.
	// @param to, the owner address of the token. Obligatory.
	// @param tokenId, the id of the token. Obligatory.
	// @param categoryName, the name of the category. If categoryName is empty, the token's category
	// will be the default category. Optional.
	// @param metadata, the metadata of the token. Optional.
	// @return error, the error msg if some error occur.
	// @event, topic: 'mint'; data: to, tokenId, categoryName, metadata
	Mint(to, tokenId, categoryName string, metadata []byte) error

	// MintBatch mint nft tokens batch. Obligatory.
	// @param tokens, the tokens to mint. Obligatory.
	// @return error, the error msg if some error occur.
	// @event, topic: 'mintBatch'; data: tokens
	MintBatch(tokens []NFA) error

	// SetApproval approve or cancel approve token to 'to' account. Obligatory.
	// @param owner, the owner of token. Obligatory.
	// @param to, destination approve to. Obligatory.
	// @param tokenId, the token id. Obligatory.
	// @param isApproval, to approve or to cancel approve
	// @return error, the error msg if some error occur.
	// @event, topic: 'setApproval'; data: to, tokenId, isApproval
	SetApproval(owner, to, tokenId string, isApproval bool) error

	// SetApprovalForAll approve or cancel approve all token to 'to' account. Obligatory.
	// @param owner, the owner of token. Obligatory.
	// @param to, destination address approve to. Obligatory.
	// @isApprove, true means approve and false means cancel approve. Obligatory.
	// @return error, the error msg if some error occur.
	// @event, topic: 'setApprovalForAll'; data: to, isApproval
	SetApprovalForAll(owner, to string, isApproval bool) error

	// TransferFrom transfer single token after approve. Obligatory.
	// @param from, owner account of token. Obligatory.
	// @param to, destination account transferred to. Obligatory.
	// @param tokenId, the token being transferred. Obligatory.
	// @return error, the error msg if some error occur.
	// @event, topic: 'transferFrom'; data: from, to, tokenId
	TransferFrom(from, to, tokenId string) error

	// TransferFromBatch transfer tokens after approve. Obligatory.
	// @param from, owner account of token. Obligatory.
	// @param to, destination account transferred to. Obligatory.
	// @param tokenIds, the tokens being transferred. Obligatory.
	// @return error, the error msg if some error occur.
	// @event, topic: 'transferFromBatch'; data: from, to, tokenIds
	TransferFromBatch(from, to string, tokenIds []string) error

	// OwnerOf get the owner of a token. Obligatory.
	// @param tokenId, the token which will be queried. Obligatory.
	// @return account, the token's account.
	// @return err, the error msg if some error occur.
	OwnerOf(tokenId string) (account string, err error)

	// TokenURI get the URI of the token. a token's uri consists of CategoryURI and tokenId. Obligatory.
	// @param tokenId, tokenId be queried. Obligatory.
	// @return uri, the uri of the token.
	// @return err, the error msg if some error occur.
	TokenURI(tokenId string) (uri string, err error)

	// EmitMintEvent emit mint event
	// @param to  destination account transferred to
	// @param tokenId
	// @param categoryName category name
	// @param metadata other info
	EmitMintEvent(to, tokenId, categoryName, metadata string)

	// EmitSetApprovalEvent emit set approval event
	// @param owner  account of token.
	// @param to  destination account transferred to
	// @param tokenId
	// @param isApproval true means approve and false means cancel approve.
	EmitSetApprovalEvent(owner, to, tokenId string, isApproval bool)

	// EmitSetApprovalForAllEvent emit set approval for all event
	// @param owner  account of token.
	// @param to  destination account transferred to
	// @param isApproval true means approve and false means cancel approve.
	EmitSetApprovalForAllEvent(owner, to string, isApproval bool)

	// EmitTransferFromEvent emit transfer from event
	// @param from owner  account of token.
	// @param to  destination account transferred to
	// @param tokenId
	EmitTransferFromEvent(from, to, tokenId string)
}
type CMNFAOption interface {
	// SetApprovalByCategory approve or cancel approve tokens of category to 'to' account. Optional.
	// @param owner, the owner of token. Obligatory.
	// @param to, destination address approve to. Obligatory.
	// @categoryName, the category of tokens. Obligatory.
	// @isApproval, to approve or to cancel approve. Obligatory.
	// @return error, the error msg if some error occur.
	// @event, topic: 'setApprovalByCategory'; data: to, categoryName, isApproval
	SetApprovalByCategory(owner, to, categoryName string, isApproval bool) error

	// CreateOrSetCategory create a category of tokens. Optional.
	// @param categoryName, the category name. Obligatory.
	// @param categoryURI, the category uri. Obligatory.
	// @return error, the error msg if some error occur.
	// @event, topic: 'createOrSetCategory'; data: category
	CreateOrSetCategory(category *Category) error

	// Burn burn token
	// @param tokenId, tokenId
	Burn(tokenId string) error

	// GetCategoryByName get specific category by name. Optional.
	// @param categoryName, the name of the category. Obligatory.
	// @return category, the category returned.
	// @return err, the error msg if some error occur.
	GetCategoryByName(categoryName string) (category *Category, err error)

	// GetCategoryByTokenId get a specific category by tokenId. Optional.
	// @param tokenId, the names of category to be queried. Obligatory.
	// @return category, the result queried.
	// @return err, the error msg if some error occur.
	GetCategoryByTokenId(tokenId string) (category *Category, err error)

	// TotalSupply get total token supply of this contract. Optional.
	// @return totalSupply, the total token supply value returned.
	// @return err, the error msg if some error occur.
	TotalSupply() (totalSupply *safemath.SafeUint256, err error)

	// TotalSupplyOfCategory get total token supply of the category. Optional.
	// @param category, the category of tokens. Obligatory.
	// @return totalSupply, the total token supply value returned.
	// @return err, the error msg if some error occur.
	TotalSupplyOfCategory(category string) (totalSupply *safemath.SafeUint256, err error)

	// BalanceOf get total token number of the account. Optional
	// @param account, the account which will be queried. Obligatory.
	// @return balance, the token number of the account.
	// @return err, the error msg if some error occur.
	BalanceOf(account string) (balance *safemath.SafeUint256, err error)

	// AccountTokens get the token list of the account. Optional
	// @param account, the account which will be queried. Obligatory.
	// @return tokenId, the list of tokenId.
	// @return err, the error msg if some error occur.
	AccountTokens(account string) (tokenId []string, err error)

	// TokenMetadata get the metadata of a token. Optional.
	// @param tokenId, tokenId which will be queried.
	// @return metadata, the metadata of the token.
	// @return err, the error msg if some error occur.
	TokenMetadata(tokenId string) (metadata []byte, err error)

	// EmitBurnEvent emit burn event
	// @param tokenId
	EmitBurnEvent(tokenId string)

	// EmitCreateOrSetCategoryEvent emit CreateOrSetCategory event
	// @param categoryName
	// @param categoryURI
	EmitCreateOrSetCategoryEvent(categoryName, categoryURI string)

	// EmitSetApprovalByCategoryEvent emit set approval by category event
	// @param owner  account of token.
	// @param to  destination account transferred to
	// @param categoryName
	// @param isApproval true means approve and false means cancel approve.
	EmitSetApprovalByCategoryEvent(owner, to, categoryName string, isApproval bool)
}

// Category the tokens' category info
type Category struct {
	// CategoryName, the name of the category
	CategoryName string `json:"categoryName"`
	// CategoryURI, the uri of the category
	CategoryURI string `json:"categoryURI"`
}

// NFA a Digital Non-Fungible Assets
type NFA struct {
	// TokenId, the id of the token
	TokenId string `json:"tokenId"`
	// CategoryName, the name of the category
	CategoryName string `json:"categoryName"`
	// To, the address which the token minted to
	To string `json:"to"`
	// Metadata, the metadata of the token
	Metadata []byte `json:"metadata"`
}

type AccountTokens struct {
	Account string   `json:"account"`
	Tokens  []string `json:"tokens"`
}
