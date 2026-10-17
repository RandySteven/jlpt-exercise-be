package notif

type (
	Notif interface {
		SendEmail(email string, subject string, body string) error
		SendNotif()
	}

	notifClient struct {
		
	}
)
