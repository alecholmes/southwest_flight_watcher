package app

import (
	"net/smtp"
)

var (
	mime = "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
)

type EmailFlightsNotifier struct {
	SmtpAddress string
	Auth        smtp.Auth
	From        string
	To          string
}

var _ SearchUpdateNotifier = &EmailFlightsNotifier{}

func (e *EmailFlightsNotifier) Notify(searchStates FlightSearchStates) error {
	available, improved := searchStates.OnlyAvailable()

	// Only send notification if there are flights and at least one flight was updated
	if !improved || len(available) == 0 {
		return nil
	}

	body, err := BodyToHTML(NewBody(available))
	if err != nil {
		return err
	}

	// Send email
	msg := "From: " + e.From + "\n" +
		"To: " + e.To + "\n" +
		"Subject: Southwest Flight Price Update\n" +
		mime + "\n\n" +
		body + "\n"

	return smtp.SendMail(e.SmtpAddress, e.Auth, e.From, []string{e.To}, []byte(msg))
}
