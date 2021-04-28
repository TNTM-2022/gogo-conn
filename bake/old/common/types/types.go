package types

type UserReq struct {
	UID        int32
	Route      string
	ServerType string
	Payload    []byte
	PkgID      int64
	Sid        uint64
}
