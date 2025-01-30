package email

import (
	"fmt"
)

func (c *Client) SendPasswordResetEmail(to, resetLink string) error {
	subject := "Reset Your Password - Vinyl Collection"

	textBody := fmt.Sprintf(`Hello,

A password reset has been requested for your account. If you did not request this, please ignore this email.

To reset your password, click the following link:
%s

This link will expire in 1 hour.

Best regards,
Vinyl Collection Team`, resetLink)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <h2>Reset Your Password</h2>
    <p>Hello,</p>
    <p>A password reset has been requested for your account. If you did not request this, please ignore this email.</p>
    <p>To reset your password, click the following link:</p>
    <p><a href="%s" style="display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px;">Reset Password</a></p>
    <p style="color: #666; font-size: 0.9em;">This link will expire in 1 hour.</p>
    <br>
    <p>Best regards,<br>Vinyl Collection Team</p>
</body>
</html>`, resetLink)

	return c.SendEmail(to, subject, textBody, htmlBody)
}

func (c *Client) SendVerificationEmail(to, verificationLink string) error {
	subject := "Verify Your Email - Vinyl Collection"

	textBody := fmt.Sprintf(`Welcome to Vinyl Collection!

Please verify your email address by clicking the following link:
%s

This link will expire in 24 hours.

Best regards,
Vinyl Collection Team`, verificationLink)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <h2>Welcome to Vinyl Collection!</h2>
    <p>Please verify your email address by clicking the button below:</p>
    <p><a href="%s" style="display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px;">Verify Email</a></p>
    <p style="color: #666; font-size: 0.9em;">This link will expire in 24 hours.</p>
    <br>
    <p>Best regards,<br>Vinyl Collection Team</p>
</body>
</html>`, verificationLink)

	return c.SendEmail(to, subject, textBody, htmlBody)
}
