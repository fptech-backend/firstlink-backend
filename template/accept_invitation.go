package template

import (
	"certification/config"
	"certification/database"
	"fmt"
)

func TemplateEmailInvitation(initializer *database.Initializer, username string, token string) string {
	acceptInvitationLink := fmt.Sprintf("%s/auth/activate?token=%s", config.API_URL, token)

	return fmt.Sprintf(
		`<mjml>
		<mj-body background-color="#f0f0f0">
		  <mj-section background-color="#ffffff" padding="20px">
			<mj-column>
			  <mj-image src="%[1]s" alt="Logo" width="200px"></mj-image>
			</mj-column>
		  </mj-section>
		  <mj-section background-color="#ffffff" padding="20px">
			<mj-column>
			  <mj-text color="#F45E43" font-size="24px" font-weight="bold">Hi, %[2]s</mj-text>
			  <mj-text color="#000000">We would love to have you aboard! Please validate your registration with us and reset your password.</mj-text>
			</mj-column>
		  </mj-section>
		  <mj-section background-color="#ffffff" padding="10px">
			<mj-column>
			  <mj-button background-color="#22BC66" color="#ffffff" font-size="20px" href="%[3]s">Activate My Account</mj-button>
			</mj-column>
		  </mj-section>
		  <mj-section background-color="#ffffff" padding-left="20px" padding-right="20px" padding-bottom="10px">
			<mj-column>
			  <mj-text color="#626262">If you're not able to click on the button above, copy and paste the following link to your browser:</mj-text>
			  <mj-text color="#5e5e5e" font-size="12px">%[3]s</mj-text>
			</mj-column>
		  </mj-section>
		  <mj-section background-color="#ffffff" padding="20px">
			<mj-column>
			  <mj-divider border-color="#F45E43"></mj-divider>
			  <mj-text color="#626262">Tokenize. Organize. Track. Validate. One tool, unlimited potential. Work gets better on the CertFirst.</mj-text>
			  <mj-text color="#626262" font-size="12px">Learn more at <a href="https://tokenfirst.com">https://tokenfirst.com</a></mj-text>
			</mj-column>
		  </mj-section>
		</mj-body>
	  </mjml>
	  `, config.EMAIL_LOGO_URL, username, acceptInvitationLink,
	)
}
