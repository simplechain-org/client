package types

// HotStuff
import (
	"errors"
	"github.com/simplechain-org/client/common"
	"github.com/simplechain-org/client/crypto"
)

var (
	// HotStuffDigest represents a hash of "The scalable HotStuff"
	// to identify whether the block is from HotStuff impl engine
	HotStuffDigest = common.HexToHash("0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff")

	// HotStuffExtraVanity (Genesis): 32B+initialSigners+65B
	// (Non-genesis): 32B+Proposer+Validators+View+Signature+65B
	HotStuffExtraVanity = crypto.DigestLength // Fixed number of extra-data bytes reserved for validator vanity

	// ErrInvalidHotStuffHeaderExtra is returned if the length of extra-data is less than 32 bytes
	ErrInvalidHotStuffHeaderExtra = errors.New("invalid hotstuff header extra-data")

	// HotStuffExtraSeal Fixed number of extra-data bytes reserved for validator seal
	HotStuffExtraSeal = crypto.SignatureLength
)

// /HotStuff
