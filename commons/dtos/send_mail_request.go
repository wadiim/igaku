package dtos

type SendMailRequest struct {
	To  []string `json:"to"`
	Msg []byte   `json:"message"`
}
