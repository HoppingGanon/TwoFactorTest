package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"

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

// ======================================
// ユーザーから受け取るデータの構造体

// ユーザーから送付される認証情報
type AuthData struct {
	UserId   string `json:"userId"`
	Password string `json:"password"`
}

// ユーザーから送付される二段階認証情報
type TowFactorData struct {
	TempToken string `json:"tempToken"`
	OneTime   string `json:"oneTime"`
}

// ======================================

func main() {
	e := echo.New()

	// 環境変数の設定
	loadEnv()

	// ミドルウェアからCORSの使用を設定する
	// これを設定しないと、別サイト等からのアクセスが拒否される
	e.Use(middleware.CORS())

	// 二段階認証テストのAPI
	e.POST("/temp-token", postTempToken)
	e.POST("/token", postToken)
	e.GET("/test", getTest)

	e.Logger.Fatal(e.Start(":" + envAuthServerPort))
}

// 最初のログイン画面から送信された認証情報を処理する
// 認証に成功したら、一段階目の認証トークンを送付する
func postTempToken(c echo.Context) error {
	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		println("認証情報の送信方式が間違っています")
		println(err.Error())
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	// 認証情報のパース
	var authData AuthData
	err = json.Unmarshal(b, &authData)
	if err != nil {
		println("認証情報の送信方式が間違っています")
		println(err.Error())
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	fmt.Printf("ユーザーID\n%s\nおよびパスワード\n%s\nを受け取りました\n", authData.UserId, authData.Password)

	// 受け取ったユーザーIDとハッシュ化したパスワードが一致しない場合はエラー
	if authData.UserId != userId || getSHA256(authData.Password) != passwordHash {
		println("ユーザー名またはパスワードが違います")
		return c.String(403, "ユーザー名またはパスワードが違います")
	}

	// 一段階目の認証トークンをランダムな文字列(Token68形式と互換のあるBase64形式)で生成
	tempToken, err := getRandomBase64()
	if err != nil {
		println("一段階目の認証トークンの生成に失敗しました")
		println(err.Error())
		return c.String(400, "一段階目の認証トークンの生成に失敗しました")
	}

	fmt.Printf("一段階目の認証トークン'%s'を生成しました\n", tempToken)

	// ワンタイムパスワードの数字列を生成
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		println("ワンタイムパスワードの生成に失敗しました")
		println(err.Error())
		return c.String(400, "ワンタイムパスワードの生成に失敗しました")
	}

	// 一段階目の認証トークンとワンタイムパスワードを保管
	savedTempToken = tempToken
	savedOnetime = fmt.Sprintf("%06d", n)

	fmt.Printf("ワンタイムパスワード'%s'を生成しました\n", savedOnetime)

	// メールでワンタイムパスワードを送信
	err = SendMail(
		envSmtpHost,
		envSmtpPort,
		envSmtpUser,
		envSmtpPassword,
		envSmtpFromAddress,
		[]string{envSmtpToAddress},
		"二段階認証テスト ワンタイムパスワード通知",
		fmt.Sprintf("二段階認証テストのワンタイムパスワードを通知します。\n\n\tワンタイムパスワード\n\t%s", savedOnetime),
	)
	if err != nil {
		println("ワンタイムパスワードのメール送信ができませんでした")
		println(err.Error())
		return c.String(400, "ワンタイムパスワードのメール送信ができませんでした")
	}

	println("メールを送信しました")

	return c.String(201, tempToken)
}

func postToken(c echo.Context) error {
	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		println("認証情報の送信方式が間違っています")
		println(err.Error())
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	// 二段階認証情報のパース
	var tfd TowFactorData
	err = json.Unmarshal(b, &tfd)
	if err != nil {
		println("認証情報の送信方式が間違っています")
		println(err.Error())
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	fmt.Printf("一段階目の認証済みトークン\n%s\nおよびワンタイムパスワード\n%s\nを受け取りました\n", tfd.TempToken, tfd.OneTime)

	// トークンとワンタイムパスワードが一致しなければエラー
	if savedTempToken != tfd.TempToken || savedOnetime != tfd.OneTime {
		println("ワンタイムパスワードが一致しません")
		return c.String(403, "ワンタイムパスワードが一致しません")
	}

	// 二段階認証済みトークンをランダムな文字列(Token68形式と互換のあるBase64形式)で生成
	token, err := getRandomBase64()
	if err != nil {
		println("トークンの生成に失敗しました")
		println(err.Error())
		return c.String(400, "トークンの生成に失敗しました")
	}

	// 二段階認証済みトークンを保管
	savedToken = token

	fmt.Printf("二段階認証済みトークン'%s'を生成しました\n", savedToken)

	// 以前の認証データを破棄
	savedOnetime = ""
	savedTempToken = ""

	return c.String(201, token)
}

func getTest(c echo.Context) error {
	token, f := getBearer(c)

	fmt.Printf("トークン'%s'を受け取りました\n", token)

	if !f {
		println("認証情報の送信方式が間違っています")
		return c.String(403, "認証情報の送信方式が間違っています")
	} else if token != savedToken {
		println("ベアラートークンが一致しません")
		return c.String(403, "ベアラートークンが一致しません")
	}

	println("トークンは正しく認証されたものです")
	return c.String(200, "トークンは正しく認証されたものです")
}
