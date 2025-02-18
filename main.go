// Copyright © 2025 chouette2100@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/go-gorp/gorp"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srdblib"
)

/*
00AA00 最初のバージョン
*/

const version = "00AA00"

// イベントの最終結果（獲得ポイント）を取得して、ポイントテーブルとイベントユーザーテーブルに格納する。
// イベント終了の翌日12時〜翌々日12時にクローンなどで実行する。
// 複数回実行した場合は、最初の実行で取得したデータを上書きしない。
// 再実行でデータを上書きしたいときは event.rstatus を "Confirmed" 以外（例えば'Provisional'）に変更する。
func main() {

	// ログファイルを開く
	logfilename := version + "_" + time.Now().Format("20060102") + ".txt"
	logfile, err := os.OpenFile(logfilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Println("ログファイルを開けません:", err)
		os.Exit(1)
	}
	defer logfile.Close()

	// フォアグラウンド（端末に接続されているか）を判定
	isForeground := terminal.IsTerminal(int(os.Stdout.Fd()))

	var logOutput io.Writer
	if isForeground {
		// フォアグラウンドならログファイル + コンソール
		logOutput = io.MultiWriter(os.Stdout, logfile)
	} else {
		// バックグラウンドならログファイルのみ
		logOutput = logfile
	}

	// ロガーの設定
	log.SetOutput(logOutput)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// ログ出力テスト
	log.Println("アプリケーションを起動しました")

	// データベース接続
	dbconfig, err := srdblib.OpenDb("DBConfig.yml")
	if err != nil {
		log.Printf("Database error. err = %v\n", err)
		return
	}
	if dbconfig.UseSSH {
		defer srdblib.Dialer.Close()
	}
	defer srdblib.Db.Close()

	dial := gorp.MySQLDialect{Engine: "InnoDB", Encoding: "utf8mb4"}
	srdblib.Dbmap = &gorp.DbMap{Db: srdblib.Db,
		Dialect:         dial,
		ExpandSliceArgs: true, //スライス引数展開オプションを有効化する
	}
	srdblib.Dbmap.AddTableWithName(srdblib.User{}, "user").SetKeys(false, "Userno")
	srdblib.Dbmap.AddTableWithName(srdblib.Userhistory{}, "userhistory").SetKeys(false, "Userno", "Ts")
	// srdblib.Dbmap.AddTableWithName(srdblib.Wuser{}, "wuser").SetKeys(false, "Userno")
	srdblib.Dbmap.AddTableWithName(TWuser{}, "wuser").SetKeys(false, "Userno")
	srdblib.Dbmap.AddTableWithName(srdblib.Wuserhistory{}, "wuserhistory").SetKeys(false, "Userno", "Ts")
	srdblib.Dbmap.AddTableWithName(srdblib.Event{}, "event").SetKeys(false, "Eventid")
	srdblib.Dbmap.AddTableWithName(srdblib.Eventuser{}, "eventuser").SetKeys(false, "Eventid", "Userno")
	srdblib.Dbmap.AddTableWithName(srdblib.Wevent{}, "wevent").SetKeys(false, "Eventid")
	srdblib.Dbmap.AddTableWithName(srdblib.Weventuser{}, "weventuser").SetKeys(false, "Eventid", "Userno")

	/*
		srdblib.Dbmap.AddTableWithName(srdblib.GiftScore{}, "giftscore").SetKeys(false, "Giftid", "Ts", "Userno")
		srdblib.Dbmap.AddTableWithName(srdblib.ViewerGiftScore{}, "viewergiftscore").SetKeys(false, "Giftid", "Ts", "Viewerid")
		srdblib.Dbmap.AddTableWithName(srdblib.Viewer{}, "viewer").SetKeys(false, "Viewerid")
		srdblib.Dbmap.AddTableWithName(srdblib.ViewerHistory{}, "viewerhistory").SetKeys(false, "Viewerid", "Ts")

		srdblib.Dbmap.AddTableWithName(srdblib.Campaign{}, "campaign").SetKeys(false, "Campaignid")
		srdblib.Dbmap.AddTableWithName(srdblib.GiftRanking{}, "giftranking").SetKeys(false, "Campaignid", "Grid")
		srdblib.Dbmap.AddTableWithName(srdblib.Accesslog{}, "accesslog").SetKeys(false, "Ts", "Eventid")
	*/

	client, cookiejar, err := exsrapi.CreateNewClient("")
	if err != nil {
		err = fmt.Errorf("exsrapi.CeateNewClient(): %s", err.Error())
		log.Printf("err=%s\n", err.Error())
		return //       エラーがあれば、ここで終了
	}
	defer cookiejar.Save()

	SetConfirmedToEvent(client)

}
