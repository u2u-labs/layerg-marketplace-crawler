package helpers

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// AddressSet holds unique Ethereum addresses using a map
type AddressSet struct {
	addresses map[common.Address]struct{}
}

// NewAddressSet initializes a new AddressSet
func NewAddressSet() *AddressSet {
	return &AddressSet{
		addresses: make(map[common.Address]struct{}),
	}
}

// AddAddress adds a new address to the set if it doesn't already exist
func (a *AddressSet) AddAddress(addr common.Address) {
	if addr == (common.Address{}) {
		return
	}

	if _, exists := a.addresses[addr]; exists {
		return
	}
	a.addresses[addr] = struct{}{} // Use an empty struct for memory efficiency
}

// GetAddresses returns the list of unique addresses
func (a *AddressSet) GetAddresses() []common.Address {
	addressList := make([]common.Address, 0, len(a.addresses))
	for addr := range a.addresses {
		addressList = append(addressList, addr)
	}
	return addressList
}

// Reset clears all addresses in the AddressSet
func (a *AddressSet) Reset() {
	a.addresses = make(map[common.Address]struct{}) // Reinitialize the map
}

// FOR ERC 721

// TokenIdSet keeps track of unique token IDs and their associated transaction data
type TokenIdSet struct {
	tokenIds map[string]struct{}     // Set of unique token IDs
	txEvents map[string][]TokenEvent // Map of token ID to its events
}

// TokenEvent stores data about a single token transfer event
type TokenEvent struct {
	TxHash      string
	From        string
	To          string
	LogIndex    uint
	BlockNumber string
}

// NewTokenIdSet creates a new TokenIdSet
func NewTokenIdSet() *TokenIdSet {
	return &TokenIdSet{
		tokenIds: make(map[string]struct{}),
		txEvents: make(map[string][]TokenEvent),
	}
}

// AddTokenId adds a new token ID to the set if it doesn't already exist
// and appends the transaction data to its events
func (t *TokenIdSet) AddTokenId(tokenId *big.Int, txHash, from, to string, logIndex uint, blockNumber uint64) {
	if tokenId == nil {
		return
	}

	// Use string representation for the token ID
	tokenStr := tokenId.String()

	// Add to set of unique tokens (this doesn't change)
	t.tokenIds[tokenStr] = struct{}{}
	// uint64 to string
	blockNumberStr := strconv.FormatUint(blockNumber, 10)

	// Add new event to the token's event list
	event := TokenEvent{
		TxHash:      txHash,
		From:        from,
		To:          to,
		LogIndex:    logIndex,
		BlockNumber: blockNumberStr,
	}

	t.txEvents[tokenStr] = append(t.txEvents[tokenStr], event)
}

// GetTxHash returns the transaction hash of the last event for the specified token ID
func (t *TokenIdSet) GetTxHash(tokenId string) string {
	events := t.txEvents[tokenId]
	if len(events) == 0 {
		return ""
	}
	return events[len(events)-1].TxHash
}

// GetFrom returns the 'from' address of the last event for the specified token ID
func (t *TokenIdSet) GetFrom(tokenId string) string {
	events := t.txEvents[tokenId]
	if len(events) == 0 {
		return ""
	}
	return events[len(events)-1].From
}

// GetTo returns the 'to' address of the last event for the specified token ID
func (t *TokenIdSet) GetTo(tokenId string) string {
	events := t.txEvents[tokenId]
	if len(events) == 0 {
		return ""
	}
	return events[len(events)-1].To
}

// GetLogIndex returns the log index of the last event for the specified token ID
func (t *TokenIdSet) GetLogIndex(tokenId string) uint {
	events := t.txEvents[tokenId]
	if len(events) == 0 {
		return 0
	}
	return events[len(events)-1].LogIndex
}

// GetAllEvents returns all events for the specified token ID
func (t *TokenIdSet) GetAllEvents(tokenId string) []TokenEvent {
	return t.txEvents[tokenId]
}

// GetLatestEvent returns the most recent event for the specified token ID
func (t *TokenIdSet) GetLatestEvent(tokenId string) (TokenEvent, bool) {
	events := t.txEvents[tokenId]
	if len(events) == 0 {
		return TokenEvent{}, false
	}
	return events[len(events)-1], true
}

// GetTokenIds returns the list of unique token IDs
func (t *TokenIdSet) GetTokenIds() []*big.Int {
	tokenIdList := make([]*big.Int, 0, len(t.tokenIds))
	for tokenStr := range t.tokenIds {
		id := new(big.Int)
		id.SetString(tokenStr, 10) // Convert string back to big.Int
		tokenIdList = append(tokenIdList, id)
	}
	return tokenIdList
}

// Reset clears all token IDs and events in the TokenIdSet
func (t *TokenIdSet) Reset() {
	t.tokenIds = make(map[string]struct{})
	t.txEvents = make(map[string][]TokenEvent)
}

// ///
// FOR ERC 1155
// ///

// TokenIdContractAddressSet holds unique token ID and contract address pairs using a map of values
type TokenIdContractAddressSet struct {
	tokenIds map[string]struct{}
}

// NewTokenIdContractAddressSet initializes a new TokenIdContractAddressSet
func NewTokenIdContractAddressSet() *TokenIdContractAddressSet {
	return &TokenIdContractAddressSet{
		tokenIds: make(map[string]struct{}),
	}
}

// AddTokenId adds a new token ID and contract address pair to the set if it doesn't already exist
func (t *TokenIdContractAddressSet) AddTokenIdContractAddress(tokenId *big.Int, contractAddress string) {
	if tokenId == nil || tokenId.Cmp(big.NewInt(0)) == 0 || contractAddress == "" {
		return
	}

	// Create a unique key combining token ID and contract address
	tokenStr := tokenId.String() + ":" + contractAddress

	if _, exists := t.tokenIds[tokenStr]; exists {
		return
	}
	t.tokenIds[tokenStr] = struct{}{} // Use an empty struct for memory efficiency
}

// GetTokenIds returns the list of unique token ID and contract address pairs
func (t *TokenIdContractAddressSet) GetTokenIdContractAddressses() []struct {
	TokenId         *big.Int
	ContractAddress string
} {
	tokenIdList := make([]struct {
		TokenId         *big.Int
		ContractAddress string
	}, 0, len(t.tokenIds))

	for tokenStr := range t.tokenIds {
		// Split the unique key back into token ID and contract address
		parts := strings.Split(tokenStr, ":")
		if len(parts) != 2 {
			continue // Skip malformed entries
		}

		id := new(big.Int)
		id.SetString(parts[0], 10) // Convert string back to big.Int
		tokenIdList = append(tokenIdList, struct {
			TokenId         *big.Int
			ContractAddress string
		}{
			TokenId:         id,
			ContractAddress: parts[1],
		})
	}
	return tokenIdList
}

// Reset clears all token ID and contract address pairs in the TokenIdContractAddressSet
func (t *TokenIdContractAddressSet) Reset() {
	t.tokenIds = make(map[string]struct{}) // Reinitialize the map
}
