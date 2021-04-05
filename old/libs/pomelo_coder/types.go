package coder

// MType msg type
type MType struct {
	Type byte
	Body []byte
}

// DecodedMsg struct
type DecodedMsg struct {
	ID            int64    `json:"id"`
	Type          byte   `json:"type"`
	CompressRoute bool   `json:"compressRoute"`
	Route         string `json:"route"`
	Body          []byte `json:"body"`
	CompressGzip  bool   `json:"compressGzip"`
}
