// Copyright © 2025 chouette2100@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	//"errors"
	"fmt"
	//	"log"
	//"strconv"
	//"strings"

	"net/http"

	"github.com/Chouette2100/srapi"

)

/*
type Block_ranking struct {
	//	boolとintのinterface{}型はデータ取得時にbool型として値を設定しておくこと
	//	該当する変数を使うときは一律に if br.Is_official.(bool) { ... } のようにすればよい
	Is_official      interface{} //	block_id != 0のとき false/true, bock_id == 0 のとき 0/1
	Room_url_key     string
	Room_description string
	Image_s          string
	Is_online        interface{} //	block_id != 0のとき false/true, bock_id == 0 のとき 0/1
	Is_fav           bool
	Genre_id         int
	Point            int
	Room_id          string
	Rank             int
	Room_name        string
}

type EventBlockRanking struct {
	Total_entries      int
	Entries_per_pages  int
	Current_page       int
	Block_ranking_list []Block_ranking
}
	*/

// ブロックランキングイベント参加中のルーム情報の一覧を取得する（srapiにある関数に問題があるため作成した、srapiの関数を修正すべき）
func GetEventBlockRanking(
	client *http.Client,
	eventid int,
	blockid int,
	ib int,
	ie int,
) (
	ebr *srapi.EventBlockRanking,
	err error,
) {

	ebr = new(srapi.EventBlockRanking)
	ebr.Block_ranking_list = make([]srapi.Block_ranking, 0)

	noroom := 0

	for page := 1; page > 0; {
		tebr := new(srapi.EventBlockRanking)
		//	tebr.Block_ranking_list = make([]Block_ranking, 0)

		//	イベント参加ルーム一覧のデータ（htmmの一部をぬきだした形になっている）を取得する。
		//	データを分割して取得するタイプのAPIを使うときはこのような処理を入れておいた方が安全です。
		tebr, err = srapi.ApiBlockEventRnaking(client, eventid, blockid, page)
		if err != nil {
			err = fmt.Errorf("ApiEventRoomList(): %w", err)
			return nil, err
		}
		if len(tebr.Block_ranking_list) == 0 {
			break
		}

		if page == 1 {
			ebr = tebr
		} else {
			ebr.Block_ranking_list = append(
				ebr.Block_ranking_list,
				tebr.Block_ranking_list...,
			)
		}

		noroom = len(ebr.Block_ranking_list)

		if noroom == ebr.Total_entries || noroom >= ie {
			break
		}

		//	次のページへ
		page++
	}

	if ib <= noroom {
		if ie > noroom {
			ie = noroom
		}
		ebr.Block_ranking_list = ebr.Block_ranking_list[ib-1 : ie]
	} else {
		ebr.Block_ranking_list = nil
		return
	}

	return
}