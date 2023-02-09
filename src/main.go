package main

import (
	"fmt"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// ======================================
// 環境変数から読み込む変数(環境変数が設定されてなければそのまま使用する)

// 二段階認証サーバーのポート 環境変数'AUTH_SERVER_PORT'に対応する
var envAuthServerPort = "8080"

// 二段階認証サーバーのポート 環境変数'AUTH_BASE_URI'に対応する
var envAuthBaseUri = "http://localhost:8080"

// SMTPサーバーのホスト 環境変数'SMTP_HOST'に対応する
var envSmtpHost = "smtp.gmail.com"

// SMTPの宛先ポート 環境変数'SMTP_PORT'に対応する
var envSmtpPort = "587"

// SMTPに含めるユーザー名 環境変数'SMTP_USER'に対応する
var envSmtpUser = "<from>@gmail.com"

// SMTPに含めるパスワード 環境変数'SMTP_PASSWORD'に対応する
var envSmtpPassword = "<gmail api key>"

// 送信元のメールアドレス 環境変数'SMTP_FROM_ADDRESS'に対応する
var envSmtpFromAddress = "<from>@gmail.com"

// 送信先のメールアドレス 環境変数'SMTP_TO_ADDRESS'に対応する
var envSmtpToAddress = "<to>@gmail.com"

// ======================================
// テスト用の変数、定数
// 実際の運用ではデータベース等で管理する

// ユーザーID
const userId = "user"

// パスワード「password」をSHA256でハッシュ化したもの
const passwordHash = "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"

// 保存する一段階目の認証トークン
var savedTempToken = ""

// 保存するワンタイムパスワード
var savedOnetime = ""

// 保存する二段階認証済みトークン
var savedToken = ""

func main() {
	e := echo.New()

	// 環境変数の設定
	loadEnv()

	// ミドルウェアからCORSの使用を設定する
	// これを設定しないと、別サイト等からのアクセスが拒否される
	e.Use(middleware.CORS())

	// 二段階認証テストのAPI
	e.POST("/api/temp-token", postTempToken)
	e.POST("/api/token", postToken)
	e.GET("/api/test", getTest)

	// テスト用のページ
	e.GET("/view/login", createLogin)
	e.GET("/view/onetime", createOnetime)
	e.GET("/view/test", createTest)

	e.Logger.Fatal(e.Start(":" + envAuthServerPort))
	fmt.Printf("\n'%s/view/login.html'にアクセスすることで、動作を確認できます\n", envAuthBaseUri)
	fmt.Printf("[注意] 一段階目の認証プロセスに成功すると、'%s'から'%s'へメールが送信されるため、アドレス等が間違ってないか十分に確認してから実行してください\n",
		envSmtpFromAddress,
		envSmtpToAddress,
	)
}
