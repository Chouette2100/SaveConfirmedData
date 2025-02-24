// Copyright © 2025 chouette2100@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	"fmt"
	"log"
	"time"

	"net/http"

	// "github.com/Chouette2100/srapi"
	"github.com/Chouette2100/srdblib"
)

// イベント最終結果（確定結果）をeventuserテーブルに格納する（既存の場合は更新する）
func UpinsEventuserByEventranking(client *http.Client, tnow time.Time, eventid string,
	roomdata Roomdata,
) (
	err error,
) {

	// ユーザーを更新（あるいは追加）する
	user := srdblib.User{
		Userno: roomdata.RoomID,
	}
	srdblib.UpinsUserSetProperty(client, tnow, &user, 14400, 5000)

	// イベントユーザー情報を取得する
	var intf interface{}
	intf, err = srdblib.Dbmap.Get(srdblib.Eventuser{}, eventid, roomdata.RoomID)
	if err != nil {
		err = fmt.Errorf("Dbmap.Get failed: %w", err)
		return
	} else if intf == nil {
		// なければ新規作成
		eventuser := &srdblib.Eventuser{
			Eventid:       eventid,
			Userno:        roomdata.RoomID,
			Vld:           roomdata.Rank,
			Point:         roomdata.Point,
			Istarget:      "Y",
			Iscntrbpoints: "N",
			Graph: func(v int) (yes string) {
				if v < 21 {
					return "Y"
				} else {
					return "N"
				}
			}(roomdata.Rank),
			Color: srdblib.Colorlist2[(roomdata.Rank-1)%len(srdblib.Colorlist2)].Name,
			// Status: 0,
		}
		err = srdblib.Dbmap.Insert(eventuser)
		if err != nil {
			err = fmt.Errorf("Dbmap.Insert failed: %w", err)
			return
		} else {
			log.Printf("Dbmap.Insert() success: %s %d\n", eventuser.Eventid, eventuser.Userno)
		}
	} else {
		// あれば更新
		eventuser := intf.(*srdblib.Eventuser)
		eventuser.Vld = roomdata.Rank
		eventuser.Point = roomdata.Point
		eventuser.Color = srdblib.Colorlist2[(roomdata.Rank-1)%len(srdblib.Colorlist2)].Name
		_, err = srdblib.Dbmap.Update(intf.(*srdblib.Eventuser))
		if err != nil {
			err = fmt.Errorf("Dbmap.Insert failed: %w", err)
			return
		} else {
			log.Printf("Dbmap.Update() success: %s %d\n", eventuser.Eventid, eventuser.Userno)
		}
	}
	return
}

func UpinsWeventuserByEventranking(client *http.Client, tnow time.Time, eventid string,
	roomdata Roomdata,
) (
	err error,
) {

	// ユーザーを更新（あるいは追加）する
	wuser := TWuser{
		Userno: roomdata.RoomID,
	}
	UpinsTWuserSetProperty(client, tnow, &wuser, 14400, 5000)

	// イベントユーザー情報を取得する
	var intf interface{}
	intf, err = srdblib.Dbmap.Get(srdblib.Weventuser{}, eventid, roomdata.RoomID)
	if err != nil {
		err = fmt.Errorf("Dbmap.Get failed: %w", err)
		return
	} else if intf == nil {
		// なければ新規作成
		weventuser := &srdblib.Weventuser{
			Eventid:       eventid,
			Userno:        roomdata.RoomID,
			Vld:           roomdata.Rank,
			Point:         roomdata.Point,
			Istarget:      "Y",
			Iscntrbpoints: "N",
			Graph: func(v int) (yes string) {
				if v < 21 {
					return "Y"
				} else {
					return "N"
				}
			}(roomdata.Rank),
			Color: srdblib.Colorlist2[(roomdata.Rank-1)%len(srdblib.Colorlist2)].Name,
			// Status: 0,
		}
		err = srdblib.Dbmap.Insert(weventuser)
		if err != nil {
			err = fmt.Errorf("Dbmap.Insert failed: %w", err)
			return
		} else {
			log.Printf("Dbmap.Insert() success: %s %d\n", weventuser.Eventid, weventuser.Userno)
		}
	} else {
		// あれば更新
		weventuser := intf.(*srdblib.Weventuser)
		weventuser.Vld = roomdata.Rank
		weventuser.Point = roomdata.Point
		weventuser.Color = srdblib.Colorlist2[(roomdata.Rank-1)%len(srdblib.Colorlist2)].Name
		_, err = srdblib.Dbmap.Update()
		if err != nil {
			err = fmt.Errorf("Dbmap.Insert failed: %w", err)
			return
		} else {
			log.Printf("Dbmap.Update() success: %s %d\n", weventuser.Eventid, weventuser.Userno)
		}
	}
	return
}
