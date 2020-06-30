package decoder

type Decoder interface {
	Decode(interface{})	error
	Unmarshal([]byte, interface{}) error
}

type Encoder interface {
	Encode(interface{}) error
	Marshal(interface{}) ([]byte, error)
}