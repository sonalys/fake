package hashCheck

type FileHashData struct {
	Hash  string `json:"hash,omitempty"`
	GoSum string `json:"gosum,omitempty"`
}

type Hashes map[string]FileHashData
