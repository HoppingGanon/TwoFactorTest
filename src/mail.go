package main

import (
	"fmt"
	"net/smtp"
	"strings"
)

func SendMail(
	// SMTPサーバーのホスト名
	hostname string,
	// SMTPサーバーのポート番号
	port string,
	// ユーザー名(送信元Gmailアドレス)
	username string,
	// API キー
	password string,
	// 送信元アドレス
	from string,
	// 宛先アドレス
	to []string,
	// 件名
	subject string,
	// 本文
	body string,
) error {
	auth := smtp.PlainAuth("", username, password, hostname)
	msg := []byte(strings.ReplaceAll(fmt.Sprintf("To: %s\nSubject: %s\n\n%s", strings.Join(to, ","), subject, body), "\n", "\r\n"))
	return smtp.SendMail(fmt.Sprintf("%s:%s", hostname, port), auth, from, to, msg)
}
