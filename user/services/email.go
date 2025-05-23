package services

// Email options.
type Email struct {
	// From is the source email.
	From string

	Cc []string

	// To is a set of destination emails.
	To []string

	// Subject is the email subject text.
	Subject string

	// Text is the plain text representation of the body.
	Text string

	// HTML is the HTML representation of the body.
	HTML string
}

var RegisterCodeEmailTemplate = Email{
	From:    "noreply@deepeye.cc",
	Subject: "Your DeepEye Verification Code",
	Text:    "To complete your sign in, please enter the following code: %v. This code will expire in 10 minutes.",
	HTML: `<p>To complete your sign in, please enter the following code:</p>
	<p><strong>%v</strong></p>
	<p>This code will expire in 10 minutes.</p>
	<p><br/></p>
	<p>Thanks for visiting DeepEye!</p>`,
}
