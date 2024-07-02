package template

import (
	"certification/config"
	"certification/database"
	"fmt"
)

func TemplateOTP(initializer *database.Initializer, username string, otp string) string {

	return fmt.Sprintf(
		`
	<mjml>
		<mj-head>
			<mj-attributes>
				<mj-text line-height="20px" align="center"/>
			</mj-attributes>
		</mj-head>
  
		<mj-body background-color="#1454E2">
      <mj-section>
				<mj-column>
					<mj-image width="200px" src="%[1]s" />
				</mj-column>
      </mj-section>
      
      <mj-section background-color="#FFFFFF" padding="24px" border-radius="10px">
				<mj-column>
          <mj-text font-weight="bold" font-size="24px" padding="30px">
            You have certificate claim request
          </mj-text>
          <mj-text>Hello %[2]s,</mj-text>
          <mj-text>
            We received your request for claim under email: <strong>theodore@gmail.com</strong> 
            Just a reminder, we'll create your account using this email. 
            Once you log in, this email will serve as your credential. 
            Please make sure to check this email for further instructions.
          </mj-text>
          <mj-text>
            To start this process, you will need to provide this one-time password (OTP) on 
            <strong>FirstCert</strong>
          </mj-text>
					<mj-text>
						Please enter this number when requested: <strong>%[3]s</strong>
					</mj-text>
					<mj-divider border-width="1px" padding-top="24px" />
					<mj-text line-height="24px" padding="24px 10%%">
						If you believe you have received this email in error, please delete it and notify 
            <a href="mailto:support@first-cert.com">support@first-cert.com</a>
					</mj-text>
				</mj-column>
      </mj-section>
      
      <mj-section>
        <mj-column>
					<mj-text color="#FFFFFF">
						First Cert Copyrights © 2024. All right reserved
					</mj-text>
					<mj-text color="#FFFFFF" padding="0 0 10%%">
						To unsubscribe and stop receiving these emails please <a href="https://www.first-cert.com/unsubscribe" style="color: #00CBD2; text-decoration: none;">Click Here</a>
					</mj-text>
				</mj-column>
			</mj-section>
		</mj-body>
	</mjml>
	`, config.EMAIL_LOGO_URL, username, otp)
}