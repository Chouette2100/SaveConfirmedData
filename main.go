// Copyright © 2025 chouette2100@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sys/unix"
	"syscall"

	"github.com/go-gorp/gorp"
	"golang.org/x/crypto/ssh/terminal"

	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srapi"
	"github.com/Chouette2100/srdblib"
)

/*
00AA00 最初のバージョン
00AA01 ユーザーテーブルの更新条件を1日以上から10日以上に変更する。
00AA02 参加者がいないイベントは処理の対象としない。
00AA03 バージョンにsrdblibとsrapiのバージョンを含める。
       UpinsTWuserSetProperty()で更新が最近でUpdateが不要な場合は更新しないようにする。
00AA04 srdblibを最新のバージョン(01AV01)にする。
00AB00 SIGNALでの実行中断に対応する。
00AB01 二重起動を禁止する。
00AB02 全部の処理が終了したらmain()も終了するようにする。
*/

const version = "00AB02"

// イベントの最終結果（獲得ポイント）を取得して、ポイントテーブルとイベントユーザーテーブルに格納する。
// イベント終了の翌日12時〜翌々日12時にクローンなどで実行する。
// 複数回実行した場合は、最初の実行で取得したデータを上書きしない。
// 再実行でデータを上書きしたいときは event.rstatus を "Confirmed" 以外（例えば'Provisional'）に変更する。
func main() {

	// ロックファイルのパス
	lockFilePath := "/tmp/SaveConfirmedData.lock"

	// 既存のロックファイルをチェック
	if checkExistingLock(lockFilePath) {
		fmt.Println("既に別のインスタンスが実行中です。終了します。")
		os.Exit(1)
	}

	// ロックファイルを作成・オープン
	lockFile, err := os.OpenFile(lockFilePath, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("ロックファイル作成エラー: %v\n", err)
		os.Exit(1)
	}
	// 現在のプロセスIDを書き込む
	fmt.Fprintf(lockFile, "%d", os.Getpid())
	lockFile.Close()

	// 終了時のクリーンアップ
	defer func() {
		os.Remove(lockFilePath)
		fmt.Println("ロックファイルを削除しました")
	}()

	// ログファイルを開く
	logfilename := version + "_" + srdblib.Version + "_" + srapi.Version + "_" + time.Now().Format("20060102") + ".txt"
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
	srdblib.Dbmap.AddTableWithName(srdblib.TWuser{}, "wuser").SetKeys(false, "Userno")
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

	// シグナル受信用のチャネルを作成
	sigChan := make(chan os.Signal, 1)
	// SIGINT(Ctrl+C)とSIGTERM(kill)を受け取れるように設定
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// signal.Notify(sigChan, unix.SIGINT, unix.SIGTERM)
	// signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGABRT)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGABRT)

	// 終了通知用のチャネル
	done := make(chan struct{})
	var wg sync.WaitGroup

	// 処理を実行するゴルーチン
	wg.Add(1)
	go func() {
		defer wg.Done()
		SetConfirmedToEvent(client, done)
	}()

	// シグナル受信とプロセス完了の両方を待つ
	go func() {
		sig := <-sigChan
		log.Printf("\nReceived signal: %v\n", sig)
		close(done)
	}()

	// 処理の完了を待つ
	wg.Wait()
	log.Println("すべての処理が完了しました。プログラムを終了します。")
	// ここでプログラムは自然に終了します
}

// 既存のロックファイルをチェックし、有効なプロセスが実行中かを確認
func checkExistingLock(lockFilePath string) bool {
	content, err := ioutil.ReadFile(lockFilePath)
	if err != nil {
		// ファイルが存在しない場合はロックなし
		return false
	}

	// ロックファイルからPIDを読み取る
	pid, err := strconv.Atoi(string(content))
	if err != nil {
		// PIDが読み取れない場合は古いロックファイルと判断して削除
		os.Remove(lockFilePath)
		return false
	}

	// プロセスが実行中かチェック
	// Linuxでは /proc/{pid} にアクセスする方法もあります
	process, err := os.FindProcess(pid)
	if err != nil {
		// プロセスが見つからない場合はロックファイルを削除
		os.Remove(lockFilePath)
		return false
	}

	// プロセスにシグナル0を送信してチェック（実際にはシグナルは送信されない）
	// err = process.Signal(syscall.Signal(0))
	err = process.Signal(unix.Signal(0))
	if err != nil {
		// プロセスが存在しない場合はロックファイルを削除
		os.Remove(lockFilePath)
		return false
	}

	// プロセスは実行中
	return true
}
