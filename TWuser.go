/*
!
Copyright © 2024 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php
*/
package main

import (
	"fmt"
	//	"io"
	"log"
	//	"os"
	"strconv"
	"strings"
	"time"

	"net/http"

	//	"github.com/go-gorp/gorp"
	//      "gopkg.in/gorp.v2"

	"github.com/dustin/go-humanize"
	"github.com/jinzhu/copier"

	//	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srapi"
	"github.com/Chouette2100/srdblib"
)

/*

0.0.1 UserでGenreが空白の行のGenreをSHOWROOMのAPIで更新する。
0.0.2 Userでirankが-1の行のランクが空白の行のランク情報をSHOWROOMのAPIで更新する。
0.1.0 DBのアクセスにgorpを導入する。
0.1.1 database/sqlを使った部分（コメント）を削除する
0.2.0 Event.goを追加し、User.goにEvent, Eventuser, Wuserを追加する。

*/

//	const Version = "0.1.1"

type TWuser struct {
	Userno    int
	Userid    string
	User_name string
	Longname  string
	Shortname string
	Genre     string
	// GenreID      int
	Rank  string
	Nrank string
	Prank string
	// Irank        int
	// Inrank       int
	// Iprank       int
	// Itrank       int
	Level        int
	Followers    int
	Fans         int
	FanPower     int
	Fans_lst     int
	FanPower_lst int
	Ts           time.Time
	Getp         string
	Graph        string
	Color        string
	Currentevent string
}

// ルーム番号 user.Userno が テーブル user に存在しないときは新しいデータを挿入し、存在するときは 既存のデータを更新する。
// 本来は srdblib.UpinsWuserSetPropertyを使うべきであるが、wuserのテーブルスキーマがuserと違うため使えないので新規に作成した。
func UpinsTWuserSetProperty(client *http.Client, tnow time.Time, wuser *TWuser, lmin int, wait int) (
	err error,
) {

	// row, err := srdblib.Dbmap.Get(TWuser{}, wuser.Userno)
	var vdata *srdblib.User
	var estatus int

	twuser := new(srdblib.Wuser)
	copier.Copy(twuser, wuser)
	tuser := new(srdblib.User)
	copier.Copy(tuser, wuser)

	estatus, vdata, err = srdblib.GetLastUserdata(twuser, lmin)
	if err != nil {
		err = fmt.Errorf("GetNewUserdata(userno=%d) returned error. %w", wuser.Userno, err)
		return err
	}
	switch estatus {
	case 0:
		err = InsertIntoTWuser(client, tnow, wuser.Userno)
		if err != nil {
			err = fmt.Errorf("InsertIntoUser(userno=%d) returned error. %w", wuser.Userno, err)
			return
		}
		time.Sleep(time.Duration(wait) * time.Millisecond)
	case 1:
		err = UpdateTWuserSetProperty(client, tnow, wuser)
		if err != nil {
			err = fmt.Errorf("Dbmap.Insert(userno=%d) returned error. %w", wuser.Userno, err)
			return
		}
		time.Sleep(time.Duration(wait) * time.Millisecond)
	case 2:	// 最新のデータがuserテーブルにある場合
		copier.Copy(wuser, vdata)
		_, err = srdblib.Dbmap.Update(wuser)
		if err != nil {
			err = fmt.Errorf("Dbmap.Update(userno=%d) returned error. %w", wuser.Userno, err)
			return
		}
	default:
	}

	return
}

/*
テーブル user を SHOWROOMのAPI api/roomprofile を使って得られる情報で更新する。
*/
func UpdateTWuserSetProperty(client *http.Client, tnow time.Time, wuser *TWuser) (
	err error,
) {

	lastrank := wuser.Rank

	//	ユーザーのランク情報を取得する
	ria, err := srapi.ApiRoomProfile(client, fmt.Sprintf("%d", wuser.Userno))
	if err != nil {
		err = fmt.Errorf("ApiRoomProfile_All(%d) returned error. %w", wuser.Userno, err)
		return err
	}
	if ria.Errors != nil {
		//	err = fmt.Errorf("ApiRoomProfile_All(%d) returned error. %v", userno, ria.Errors)
		//	return err
		ria.ShowRankSubdivided = "unknown"
		ria.NextScore = 0
		ria.PrevScore = 0
	}

	if ria.ShowRankSubdivided == "" {
		err = fmt.Errorf("ApiRoomProfile_All(%d) returned empty.ShowRankSubdivided", wuser.Userno)
		return err
	}

	//	user.Userno =	キー
	wuser.Userid = ria.RoomURLKey
	wuser.User_name = ria.RoomName
	//	user.Longname =		新規作成時に設定され、変更はSRCGIで行う
	//	user.Shortname =　　〃
	wuser.Genre = ria.GenreName
	//	wuser.GenreID = ria.GenreID
	wuser.Rank = ria.ShowRankSubdivided
	wuser.Nrank = humanize.Comma(int64(ria.NextScore))
	wuser.Prank = humanize.Comma(int64(ria.PrevScore))
	//	wuser.Irank = MakeSortKeyOfRank(ria.ShowRankSubdivided, ria.NextScore)
	//	wuser.Inrank = ria.NextScore
	//	wuser.Iprank = ria.PrevScore
	//	if wuser.Itrank > user.Irank {
	//		user.Itrank = user.Irank
	//	}
	wuser.Level = ria.RoomLevel
	wuser.Followers = ria.FollowerNum

	pafr, err := srapi.ApiActivefanRoom(client, strconv.Itoa(wuser.Userno), tnow.Format("200601"))
	if err != nil {
		err = fmt.Errorf("ApiActivefanRoom(%d) returned error. %w", wuser.Userno, err)
		return err
	}
	wuser.Fans = pafr.TotalUserCount
	//	wuser.FanPower = pafr.FanPower
	yy := tnow.Year()
	mm := tnow.Month() - 1
	if mm < 0 {
		yy -= 1
		mm = 12
	}
	pafr, err = srapi.ApiActivefanRoom(client, strconv.Itoa(wuser.Userno), fmt.Sprintf("%04d%02d", yy, mm))
	if err != nil {
		err = fmt.Errorf("ApiActivefanRoom(%d) returned error. %w", wuser.Userno, err)
		return err
	}
	wuser.Fans_lst = pafr.TotalUserCount
	//	wuser.FanPower_lst = pafr.FanPower

	wuser.Ts = tnow
	//	user.Getp =		ここから三つのフィールドは現在使われていない。
	//	user.Graph =	default: ''
	//	user.Color =
	eurl := ria.Event.URL
	eurla := strings.Split(eurl, "/")
	wuser.Currentevent = eurla[len(eurla)-1]

	//	cnt, err := Dbmap.Update(user)
	_, err = srdblib.Dbmap.Update(wuser)
	if err != nil {
		log.Printf("error! %v", err)
		return
	}
	//	log.Printf("cnt = %d\n", cnt)

	log.Printf("UPDATE userno=%d rank=%s -> %s nscore=%d pscore=%d longname=%s\n", wuser.Userno, lastrank, ria.ShowRankSubdivided, ria.NextScore, ria.PrevScore, ria.RoomName)
	return
}

/*
テーブル user に新しいデータを追加する
*/
func InsertIntoTWuser(client *http.Client, tnow time.Time, userno int) (
	err error,
) {

	//	ユーザーのランク情報を取得する
	ria, err := srapi.ApiRoomProfile(client, fmt.Sprintf("%d", userno))
	if err != nil {
		err = fmt.Errorf("ApiRoomProfile_All(%d) returned error. %w", userno, err)
		return err
	}
	if ria.Errors != nil {
		//	err = fmt.Errorf("ApiRoomProfile_All(%d) returned error. %v", userno, ria.Errors)
		//	return err
		ria.ShowRankSubdivided = "unknown"
		ria.NextScore = 0
		ria.PrevScore = 0
	}

	if ria.ShowRankSubdivided == "" {
		err = fmt.Errorf("ApiRoomProfile_All(%d) returned empty.ShowRankSubdivided", userno)
		return err
	}

	wuser := new(TWuser)

	wuser.Userno = userno
	wuser.Userid = ria.RoomURLKey
	wuser.User_name = ria.RoomName
	wuser.Longname = ria.RoomName
	wuser.Shortname = strconv.Itoa(userno % 100)
	wuser.Genre = ria.GenreName
	//	wuser.GenreID = ria.GenreID
	wuser.Rank = ria.ShowRankSubdivided
	wuser.Nrank = humanize.Comma(int64(ria.NextScore))
	wuser.Prank = humanize.Comma(int64(ria.PrevScore))
	//	wuser.Irank = MakeSortKeyOfRank(ria.ShowRankSubdivided, ria.NextScore)
	//	wuser.Inrank = ria.NextScore
	//	wuser.Iprank = ria.PrevScore
	//	wuser.Itrank = user.Irank
	wuser.Level = ria.RoomLevel
	wuser.Followers = ria.FollowerNum

	pafr, err := srapi.ApiActivefanRoom(client, strconv.Itoa(wuser.Userno), tnow.Format("200601"))
	if err != nil {
		err = fmt.Errorf("ApiActivefanRoom(%d) returned error. %w", wuser.Userno, err)
		return err
	}
	wuser.Fans = pafr.TotalUserCount
	//	wuser.FanPower = pafr.FanPower
	yy := tnow.Year()
	mm := tnow.Month() - 1
	if mm < 0 {
		yy -= 1
		mm = 12
	}
	pafr, err = srapi.ApiActivefanRoom(client, strconv.Itoa(wuser.Userno), fmt.Sprintf("%04d%02d", yy, mm))
	if err != nil {
		err = fmt.Errorf("ApiActivefanRoom(%d) returned error. %w", wuser.Userno, err)
		return err
	}
	wuser.Fans_lst = pafr.TotalUserCount
	//	wuser.FanPower_lst = pafr.FanPower

	wuser.Ts = tnow
	wuser.Getp = ""
	wuser.Graph = ""
	wuser.Color = ""
	eurl := ria.Event.URL
	eurla := strings.Split(eurl, "/")
	wuser.Currentevent = eurla[len(eurla)-1]

	//	cnt, err := Dbmap.Update(user)
	err = srdblib.Dbmap.Insert(wuser)
	if err != nil {
		log.Printf("error! %v", err)
		return
	}
	//	log.Printf("cnt = %d\n", cnt)

	log.Printf("INSERT userno=%d rank=%s nscore=%d pscore=%d longname=%s\n", wuser.Userno, ria.ShowRankSubdivided, ria.NextScore, ria.PrevScore, ria.RoomName)
	return
}
