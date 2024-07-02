package template

import (
	"certification/config"
	"certification/database"
	"fmt"
)

func TemplateWelcomeResetPassword(initializer database.Initializer, token string, username string) string {
	resetPassword := config.API_URL + "/reset-password?token=" + token

	return fmt.Sprintf(
		`<mjml>
			<mj-head>
				<mj-style>
					a:hover {
						background-color: #53C8EC !important;
					}
				</mj-style>
			</mj-head>
			<mj-body background-color="#d8e5eb">
				<mj-section>
					<mj-column>
						<mj-image padding-top="40px" padding-bottom="0px" width="137px" src="%s"></mj-image>
					</mj-column>
				</mj-section>
				<mj-section background-color="#ffffff" padding="10px" border-radius="5px">
					<mj-column>
						<mj-text padding-top="30px" line-height="1.5" align="center" padding-left="40px" padding-right="40px" font-size="25px" font-family="verdana" font-weight="bold">
							Welcome to CBS
						</mj-text>
						<mj-text padding-top="20px" line-height="2" align="center" padding-left="40px" padding-right="40px" font-size="15px" font-family="verdana">
						To ensure the security of your account, we kindly ask you to reset your password before you can start using our platform using this username: %s
						</mj-text>
						<mj-button href="%s" background-color="#00ADEE" border-radius="5px" font-size="16px" font-family="verdana" padding-top="20px" padding-bottom="30px">
							Reset Password
						</mj-button>
					</mj-column>
				</mj-section>
			</mj-body>
		</mjml>`, config.EMAIL_LOGO_URL, username, resetPassword,
	)
}
