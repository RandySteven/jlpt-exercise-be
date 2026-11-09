package whatsapp_client

type WhatsappClient interface {
	SendMessage(message string) error
	ReceiveMessage() (string, error)
}
