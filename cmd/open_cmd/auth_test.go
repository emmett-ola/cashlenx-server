package open_cmd

import "testing"

func TestOpenAuthValidationErrors(t *testing.T) {
	tests := []struct {
		name  string
		reset func()
		run   func() error
	}{
		{
			name: "login requires credentials",
			reset: func() {
				loginUsername = ""
				loginPassword = ""
				loginRefreshToken = ""
			},
			run: func() error { return loginCmd.RunE(loginCmd, nil) },
		},
		{
			name: "register requires all fields",
			reset: func() {
				registerUsername = ""
				registerPassword = ""
				registerEmail = ""
				registerVerificationToken = ""
			},
			run: func() error { return registerCmd.RunE(registerCmd, nil) },
		},
		{
			name: "reset password requires identity",
			reset: func() {
				resetEmailOrUsername = ""
			},
			run: func() error { return resetPasswordCmd.RunE(resetPasswordCmd, nil) },
		},
		{
			name: "reset password confirm requires token and password",
			reset: func() {
				resetToken = ""
				resetPassword = ""
			},
			run: func() error { return resetPasswordConfirmCmd.RunE(resetPasswordConfirmCmd, nil) },
		},
		{
			name: "send verification code requires purpose and email",
			reset: func() {
				verificationPurpose = ""
				verificationEmail = ""
			},
			run: func() error { return sendVerificationCodeCmd.RunE(sendVerificationCodeCmd, nil) },
		},
		{
			name: "verify code requires purpose email and code",
			reset: func() {
				verificationPurpose = ""
				verificationEmail = ""
				verificationCode = ""
			},
			run: func() error { return verifyVerificationCodeCmd.RunE(verifyVerificationCodeCmd, nil) },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.reset()
			if err := tt.run(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
