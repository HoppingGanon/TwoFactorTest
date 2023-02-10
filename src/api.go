package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/labstack/echo"
)

// 最初のログイン画面から送信された認証情報を処理する
// 認証に成功したら、一段階目の認証トークンを送付する
func postTempToken(c echo.Context) error {
	fmt.Println("\n-------------------------")

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Println("認証情報の送信方式が間違っています")
		fmt.Println(err.Error())
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	// 認証情報のパース
	var authData AuthData
	err = json.Unmarshal(b, &authData)
	if err != nil {
		fmt.Println("認証情報の送信方式が間違っています")
		fmt.Println(err.Error())
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	fmt.Printf("ユーザーID'%s'およびパスワード'%s'を受け取りました\n", authData.UserId, authData.Password)

	// 受け取ったユーザーIDとハッシュ化したパスワードが一致しない場合はエラー
	if authData.UserId != userId || getSHA256(authData.Password) != passwordHash {
		fmt.Println("ユーザー名またはパスワードが違います")
		return c.String(403, "ユーザー名またはパスワードが違います")
	}

	// 一段階目の認証トークンをランダムな文字列(Token68形式と互換のあるBase64形式)で生成
	tempToken, err := getRandomBase64()
	if err != nil {
		fmt.Println("一段階目の認証トークンの生成に失敗しました")
		fmt.Println(err.Error())
		return c.String(400, "一段階目の認証トークンの生成に失敗しました")
	}

	fmt.Printf("一段階目の認証トークン'%s'を生成しました\n", tempToken)

	// ワンタイムパスワードの数字列を生成
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		fmt.Println("ワンタイムパスワードの生成に失敗しました")
		fmt.Println(err.Error())
		return c.String(400, "ワンタイムパスワードの生成に失敗しました")
	}
	onetime := fmt.Sprintf("%06d", n)

	fmt.Printf("ワンタイムパスワード'%s'を生成しました\n", onetime)

	// メールでワンタイムパスワードを送信
	err = SendMail(
		envSmtpHost,
		envSmtpPort,
		envSmtpUser,
		envSmtpPassword,
		envSmtpFromAddress,
		[]string{envSmtpToAddress},
		"二段階認証テスト ワンタイムパスワード通知",
		fmt.Sprintf("二段階認証テストのワンタイムパスワードを通知します。\n\n\tワンタイムパスワード\n\t%s", onetime),
	)
	if err != nil {
		fmt.Println("ワンタイムパスワードのメール送信ができませんでした")
		fmt.Println(err.Error())
		return c.String(400, "ワンタイムパスワードのメール送信ができませんでした")
	}

	fmt.Println("メールを送信しました")

	// 一段階目の認証トークンとワンタイムパスワードを保管
	savedTempToken = tempToken
	savedOnetime = onetime

	return c.String(201, tempToken)
}

func postToken(c echo.Context) error {
	fmt.Println("\n-------------------------")

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		fmt.Println("認証情報の送信方式が間違っています")
		fmt.Println(err.Error())
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	// 二段階認証情報のパース
	var tfd TowFactorData
	err = json.Unmarshal(b, &tfd)
	if err != nil {
		fmt.Println("認証情報の送信方式が間違っています")
		fmt.Println(err.Error())
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	fmt.Printf("一段階目の認証済みトークン'%s'およびワンタイムパスワード'%s'を受け取りました\n", tfd.TempToken, tfd.OneTime)

	// トークンとワンタイムパスワードが一致しなければエラー
	if savedTempToken != tfd.TempToken || savedOnetime != tfd.OneTime {
		fmt.Println("ワンタイムパスワードが一致しません")
		return c.String(403, "ワンタイムパスワードが一致しません")
	}

	// 二段階認証済みトークンをランダムな文字列(Token68形式と互換のあるBase64形式)で生成
	token, err := getRandomBase64()
	if err != nil {
		fmt.Println("トークンの生成に失敗しました")
		fmt.Println(err.Error())
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
	fmt.Println("\n-------------------------")

	token, f := getBearer(c)

	fmt.Printf("二段階認証済みトークン'%s'を受け取りました\n", token)

	if !f {
		fmt.Println("認証情報の送信方式が間違っています")
		return c.String(403, "認証情報の送信方式が間違っています")
	} else if token != savedToken {
		fmt.Println("ベアラートークンが一致しません")
		return c.String(403, "ベアラートークンが一致しません")
	}

	fmt.Println("トークンは正しく認証されたものです")
	return c.String(200, "トークンは正しく認証されたものです")
}
