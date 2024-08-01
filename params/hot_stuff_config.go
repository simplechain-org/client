package params

// HotStuff

// String implements the stringer interface, returning the impl engine details.
func (h *HotStuffConfig) String() string {
	return "hotStuff"
}

// HotStuffConfig 这个配置是创世文件里需要使用的
type HotStuffConfig struct {
	RoundTimeout         uint64 `json:"roundTimeout"`
	RoundTimeoutInterval uint64 `json:"roundTimeoutInterval"`
	MaxTimeout           uint64 `json:"maxTimeout"`
	ViewNumsPerEpoch     uint64 `json:"viewNumsPerEpoch"`
	Crypto               string `json:"crypto"`
	// The name of the impl implementation to use.
	Consensus string `json:"impl"`
	// The name of the leader rotation algorithm to use.
	LeaderRotation string `json:"leaderRotation"`
	// Epoch length to reset votes and checkpoint
	Epoch uint64 `json:"epoch"`
}

// /HotStuff
