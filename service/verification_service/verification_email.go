package verification_service

import (
	"html/template"
	"strings"
)

var verificationEmailTemplate = template.Must(template.New("verification-email").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>CashLenX verification code</title>
</head>
<body style="margin:0; padding:0; background:#f3f7f7; color:#1f2937; font-family:Arial,Helvetica,sans-serif;">
  <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width:100%; background:#f3f7f7;">
    <tr>
      <td align="center" style="padding:24px 12px;">
        <table role="presentation" width="100%" cellspacing="0" cellpadding="0" border="0" style="width:100%; max-width:560px; background:#ffffff; border:1px solid #dce9e8; border-radius:16px;">
          <tr>
            <td style="padding:32px 24px; text-align:center;">
              <p style="margin:0 0 8px; color:#008080; font-size:14px; font-weight:700; letter-spacing:0.08em; text-transform:uppercase;">CashLenX</p>
              <h1 style="margin:0; color:#111827; font-size:26px; line-height:1.25;">Confirm your email</h1>
              <p style="margin:16px 0 0; color:#4b5563; font-size:16px; line-height:1.55;">Use this verification code to continue your CashLenX request.</p>
              <p style="margin:24px 0;">
                <code aria-label="Verification code" style="display:inline-block; padding:16px 22px; border:1px solid #99cccc; border-radius:12px; background:#eaf6f5; color:#006666; font-family:Arial,Helvetica,sans-serif; font-size:30px; font-weight:700; letter-spacing:0.18em; line-height:1; user-select:all;">{{.Code}}</code>
              </p>
              <p style="margin:0; color:#4b5563; font-size:14px; line-height:1.5;">This code expires in {{.Expiry}}.</p>
              <p style="margin:20px 0 0; color:#6b7280; font-size:13px; line-height:1.5;">If you did not request this code, you can safely ignore this email.</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`))

type verificationEmailData struct {
	Code   string
	Expiry string
}

func renderVerificationEmail(code string, expiry string) (string, error) {
	var body strings.Builder
	err := verificationEmailTemplate.Execute(&body, verificationEmailData{
		Code:   code,
		Expiry: expiry,
	})
	return body.String(), err
}
