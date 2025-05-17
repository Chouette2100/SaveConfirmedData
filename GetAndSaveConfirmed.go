// Copyright © 2022-2024 chouette.21.00@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	//	"crypto/aes"
	//	"fmt"
	//	"log"
	//	"os"
	"strconv"
	//	"strings"
	//	"sync"
	"fmt"
	"time"

	"log"
	//	"bufio"
	//	"io"

	//	"runtime"

	"net/http"

	//	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	//	"github.com/go-gorp/gorp"

	//	"encoding/json"
	//	"github.com/360EntSecGroup-Skylar/excelize"

	//	. "MyModule/ShowroomCGIlib"

	//	"github.com/dustin/go-humanize"

	"github.com/Chouette2100/exsrapi/v2"
	"github.com/Chouette2100/srapi/v2"
	"github.com/Chouette2100/srdblib/v2"
)

const ConfirmedAt = 59

type Roomdata struct {
	RoomID int
	Name   string
	Rank   int
	Point  int
}

// 獲得ポイントの最終結果を取得して、ポイントテーブルとイベントユーザーテーブルに格納する。
func GetAndSaveConfirmed(client *http.Client, event *srdblib.Event, is_block bool, blockid int, sdat bool, uinf bool) (err error) {

	fncname := exsrapi.FuncNameOfThisFunction(1) + "()"
	cmt0 := event.Eventid
	log.Println(cmt0, ">>>>>>>>>>>>>>>>>>", fncname, ">>>>>>>>>>>>>>>>>>>")
	defer exsrapi.PrintExf(cmt0, fncname)()

	//	svtime := gschedule.Endtime.Add(1 * time.Second)
	svtime := event.Endtime.Add(time.Duration(ConfirmedAt) * time.Second)
	// svtime := event.Endtime // wevent.Endtime は格納時に +59秒されている模様
	eventid := event.Eventid
	ieventid := event.Ieventid

	isconfirm := false
	isquest := false
	tnow := time.Now().Truncate(time.Second)

	var roomdatalist []Roomdata
	var ebr *srapi.EventBlockRanking
	if is_block {
		// ebr, err = srapi.GetEventBlockRanking(client, ieventid, blockid, 1, 2000)
		ebr, err = GetEventBlockRanking(client, ieventid, blockid, 1, 2000)
		if err != nil {
			err = fmt.Errorf("%s GetEventBlockRanking() err=[%w]", eventid, err)
			return
		}
		roomdatalist = make([]Roomdata, len(ebr.Block_ranking_list))
		for i, blockranking := range ebr.Block_ranking_list {
			roomdatalist[i] = Roomdata{
				RoomID: func(rid string) (rno int) {
					rno, _ = strconv.Atoi(rid)
					return
				}(blockranking.Room_id),
				Name:  blockranking.Room_name,
				Rank:  blockranking.Rank,
				Point: blockranking.Point,
			}
		}
	} else {
		//	イベントに参加しているルームの一覧を取得します。
		//	ルーム名、ID、URLを取得しますが、イベント終了直後の場合の最終獲得ポイントが表示されている場合はそれも取得します。
		/*
			breg := 1
			//	確定値（最終獲得ポイント）が発表されるのは30位まで。確定値が発表されないイベントもあるので要注意。
			ereg := 100
			var Roomlistinf *srapi.RoomListInf
			Roomlistinf, err = srapi.GetRoominfFromEventByApi(http.DefaultClient, ieventid, breg, ereg)
			if err != nil {
				err = fmt.Errorf("%s GetRoominfFromEventByApi() err=[%s]", eventid, err.Error())
				return
			}

			// if Roomlistinf.RoomList[0].Rank == 0 {
			// 	log.Printf("     ****** %s is level ivent!\n", eventid)
			// 	return
			// }

			if len(Roomlistinf.RoomList) == 0 {
				log.Printf("	 ****** %s is no room!\n", eventid)
				return
			}

			pranking, err = srapi.ApiEventsRanking(http.DefaultClient, ieventid, Roomlistinf.RoomList[0].Room_id, 0)
			if err != nil {
				err = fmt.Errorf("%s ApiEventsRanking() err=[%w]", eventid, err)
				return
			}
		*/
		var eranking *srapi.EventRanking
		eranking, err = srapi.GetEventRankingByApi(client, event.Eventid, 1, 100)
		roomdatalist = make([]Roomdata, len(eranking.Ranking))
		for i, ranking := range eranking.Ranking {
			roomdatalist[i] = Roomdata{
				RoomID: ranking.RoomID,
				Name:   ranking.RoomName,
				Rank:   ranking.Rank,
				Point:  ranking.Point,
			}
		}
	}

	for i, roomdata := range roomdatalist {

		log.Printf(" i+1=%d, userno=%d, point=%d(%s)\n", i+1, roomdata.RoomID, roomdata.Point, svtime.Format("2006-01-02 15:04:05"))
		//	InsertIntoOrUpdatePoints(svtime, roominf.Userno, roominf.Point, i+1, 0, eventid, "Conf.", "", "", "")
		// InsertIntoOrUpdatePoints(svtime, roominf, i+1, 0, eventid, "Conf.", "", "", "")
		if sdat {
			InsertIntoOrUpdatePoints(svtime, roomdata, roomdata.Rank, 0, eventid, "Conf.", "", "", "")
			eu := srdblib.Eventuser{}
			err = UpinsEventuser(client, tnow, &eu, eventid, roomdata)
			weu := srdblib.Weventuser{}
			err = UpinsEventuser(client, tnow, &weu, eventid, roomdata)
			isconfirm = true
		}

		if uinf {
			// 獲得ポイントが確定する前に OnlyUinf で起動したときにユーザ情報を取得する
			// 確定結果を取得するときはユーザー情報は更新済みのはず
			eu := srdblib.Eventuser{}
			err = UpinsUser(client, tnow, &eu, eventid, roomdata)
			if err != nil {
				err = fmt.Errorf("%s UpinsUser() err=[%w]", eventid, err)
				return
			}
			weu := srdblib.Weventuser{}
			err = UpinsUser(client, tnow, &weu, eventid, roomdata)
			if err != nil {
				err = fmt.Errorf("%s UpinsUser() err=[%w]", eventid, err)
				return
			}
		}

	}

	log.Printf("%s isconfirm =%t, isquest=%t sdat=%t uinf=%t\n", eventid, isconfirm, isquest, sdat, uinf)
	if (isconfirm || isquest) && sdat {
		// 確定データの取得が成功した場合、イベントテーブルのrstatusを"Confirmed"に更新します。
		/*
			sqlstmt := "update event set rstatus = ? where eventid = ?"
			_, srdblib.Dberr = srdblib.Db.Exec(sqlstmt, "Confirmed", eventid)
			if srdblib.Dberr != nil {
				log.Printf("%s GetConfirmed() update event err=[%s]\n", eventid, srdblib.Dberr.Error())
				err = fmt.Errorf("%s GetConfirmed() update event err=[%s]", eventid, srdblib.Dberr.Error())
				return
			}
		*/
		var nu int64
		event.Rstatus = "Confirmed"
		nu, err = srdblib.Dbmap.Update(event)
		if err != nil || nu > 1 {
			err = fmt.Errorf("%s GetConfirmed() update event err=[%w] or num of Update != 1 (%d)", eventid, err, nu)
			return
		}

		var intf interface{}
		intf, err = srdblib.Dbmap.Get(srdblib.Wevent{}, eventid)
		if err != nil {
			err = fmt.Errorf("%s GetConfirmed() Dbmap.Get() err=[%w]", eventid, err)
			return
		}
		wevent := intf.(*srdblib.Wevent)
		wevent.Aclr = 1
		nu, err = srdblib.Dbmap.Update(wevent)
		if err != nil || nu > 1 {
			err = fmt.Errorf("%s GetConfirmed() update wevent err=[%w] or num of Update != 1 (%d)", eventid, err, nu)
			return
		}

		log.Printf("%s GetConfirmed() update event success\n", eventid)

		// if isconfirm {
		// 	srdblib.MakePointPerSlot(event.Event_name, eventid)
		// }
	}

	return

}

/*
// EventuserテーブルとUserテーブルを更新する。
func UpinsUserAndEventuser[T srdblib.EventuserR](
	client *http.Client,
	tnow time.Time,
	xeu T,
	eventid string,
	roomdata Roomdata,
) (
	err error,
) {
	switch any(xeu).(type) {
	case *srdblib.Eventuser:
		xuser := srdblib.User{
			Userno: roomdata.RoomID,
		}
		srdblib.UpinsUser(client, tnow, &xuser)
	case *srdblib.Weventuser:
		xuser := srdblib.Wuser{
			Userno: roomdata.RoomID,
		}
		srdblib.UpinsUser(client, tnow, &xuser)
	default:
		err = fmt.Errorf("UpinstUserAndEventuser() invalid type of xeu")
		return
	}

	eu := srdblib.Eventuser{
		EventuserBR: srdblib.EventuserBR{
			Eventid: eventid,
			Userno:  roomdata.RoomID, // roomdata.Userno,
			Vld:     roomdata.Rank,
			Point:   roomdata.Point,
		},
	}
	xeu.Set(&eu)
	err = srdblib.UpinsEventuserG(xeu, tnow)
	if err != nil {
		err = fmt.Errorf("UpinsEventuserG() err=[%w]", err)
		return
	}

	return

}
*/

// Userテーブルを更新する。存在しなかったデータを追加し、古くなったデータを削除する。
func UpinsUser[T srdblib.EventuserR](
	client *http.Client,
	tnow time.Time,
	xeu T,
	eventid string,
	roomdata Roomdata,
) (
	err error,
) {
	switch any(xeu).(type) {
	case *srdblib.Eventuser:
		xuser := srdblib.User{
			Userno: roomdata.RoomID,
		}
		srdblib.UpinsUser(client, tnow, &xuser)
	case *srdblib.Weventuser:
		xuser := srdblib.Wuser{
			Userno: roomdata.RoomID,
		}
		srdblib.UpinsUser(client, tnow, &xuser)
	default:
		err = fmt.Errorf("UpinstUserAndEventuser() invalid type of xeu")
	}

	return

}

// Eventuserテーブルを更新する。獲得ポイントが0のときは更新しない（事故防止）
func UpinsEventuser[T srdblib.EventuserR](
	client *http.Client,
	tnow time.Time,
	xeu T,
	eventid string,
	roomdata Roomdata,
) (
	err error,
) {

	if roomdata.Point == 0 {
		log.Printf(" UpinsEventuser() roomdata.Point == 0, eventid=%s, roomid=%d\n", eventid, roomdata.RoomID)
		return
	}

	eu := srdblib.Eventuser{
		EventuserBR: srdblib.EventuserBR{
			Eventid: eventid,
			Userno:  roomdata.RoomID, // roomdata.Userno,
			Vld:     roomdata.Rank,
			Point:   roomdata.Point,
		},
	}
	xeu.Set(&eu)
	err = srdblib.UpinsEventuserG(xeu, tnow)
	if err != nil {
		err = fmt.Errorf("UpinsEventuserG() err=[%w]", err)
		return
	}

	return

}
