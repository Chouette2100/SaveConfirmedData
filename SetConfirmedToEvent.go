// Copyright © 2025 chouette2100@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php
package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Chouette2100/srdblib/v2"
)

// イベントが終了直後（終了日翌日12時〜翌々日12時）にイベントの確定結果を取得して、ポイントテーブル、イベントテーブルに格納する。
func SetConfirmedToEvent(client *http.Client, sdat bool, uinf bool, done chan struct{}) (
	err error,
) {
	//      sdat    uinf   00-12
	// A.   true    true   true     前々日のイベントの結果を取得する ＋ ユーザ情報の更新(事故対策：前日何もできなかった場合)
	// B.   true    true   false    前日のイベントの結果を取得する + ユーザ情報の更新(事故対策：午前中ユーザー情報が更新できなかった場合)
	// C.   true    false  true     前々日のイベントの結果を取得する(事故対策：前日確定結果が取得できなかった場合)
	// D.   true    false  false	前日のイベントの結果を取得する(通常)
	// E.   false   true   true 	前日のイベントのユーザー情報の更新(通常)
	// F.   false   true   false	前日のイベントのユーザー情報の更新(事故対策：午前中ユーザー情報が更新できず、しかもなんらかの理由で結果取得と同時にできない場合)
	//      false   false  n/a      何もしない（起動時パラメータのチェックでエラーになる）

	/* テスト用
	tt := time.Now().Add(9 * time.Hour).Truncate(24 * time.Hour).Add(-9 * time.Hour)
	ll := tt.Add(-72 * time.Hour)
	ul := tt.Add(-48 * time.Hour)
	--- */
	/* 本番用 */
	// 以下の件をそのまま使うのはA.とC.のときだけ、つまり sdat && time.Since(tt) <= 12*time.Hour のときだけ
	tt := time.Now().Add(9 * time.Hour).Truncate(24 * time.Hour).Add(-9 * time.Hour)
	ll := tt.Add(-48 * time.Hour)
	ul := tt.Add(-24 * time.Hour)

	if !sdat || time.Since(tt) > 12*time.Hour { // B., D., E., F.のとき、つまり5行上の条件と逆のとき
		ll = ll.Add(24 * time.Hour)
		ul = ul.Add(24 * time.Hour)
		if time.Since(tt) > 12*time.Hour {
		}
	}
	/* --- */

	condition := &EventCondition{
		LowerLimitOfEndtime: ll,
		UpperLimitOfEndtime: ul,
		Sdat:                sdat,
		Uinf:                uinf,
	}
	intf, err := ExtractEvents(condition) // 戻り値は []*srdblib.Wevent
	if err != nil {
		log.Printf("ExtractEvents failed: %v\n", err)
		return
	}
	var blockid int
	for i, v := range intf {
		select {
		case <-done:
			// 強制終了させられた場合
			fmt.Printf("Interrupted at iteration %d\n", i)
			return
		default:
			fmt.Printf("Processing iteration %d\n", i)

			event := v.(*srdblib.Event)
			log.Printf("  *****************************************************\n  event: %d(%s) [%s] %s\n",
				event.Ieventid, event.Eventid, event.Rstatus, event.Event_name)
			is_block := false
			eida := strings.Split(event.Eventid, "?block_id=")
			if len(eida) > 1 {
				is_block = true
				blockid, _ = strconv.Atoi(eida[1])
			}
			err = GetAndSaveConfirmed(client, event, is_block, blockid, sdat, uinf)
			if err != nil {
				err = fmt.Errorf("GetAndSaveConfirmed failed: %w", err)
				return
			}
		}
	}

	return
}
