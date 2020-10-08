package service

type Token struct {
	Id         int32  `json:"id,omitempty"`
	ObjectType string `json:"object_type,omitempty"`
	Payload    string `json:"payload,omitempty"`
}
