package model

type Email struct {
	From string
	To   []string
	Data []byte
}
