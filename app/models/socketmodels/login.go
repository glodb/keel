package socketmodels

type Login struct {
	Token string `json:"token"`
}

func (a *Login) GetLength() int {
	return len(a.Token)
}

func (d *Login) EncodeData() []byte {
	// var buf []byte
	// buf = PackString(d.Token, buf)
	return nil
}

func (d *Login) DecodeData(buf []byte) {
	// d.Token, buf = UnPackString(buf)
}

func (d *Login) String() string {
	return "{Token:" + d.Token + "}"

}
