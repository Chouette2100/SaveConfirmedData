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

	"github.com/Chouette2100/srdblib"
)

// イベントが終了直後（終了日翌日12時〜翌々日12時）にイベントの確定結果を取得して、ポイントテーブル、イベントテーブルに格納する。
func SetConfirmedToEvent(client *http.Client, done chan struct{}) (
	err error,
) {

	tt := time.Now().Add(9 * time.Hour).Truncate(24 * time.Hour).Add(-9 * time.Hour)
	ll := tt.Add(-48 * time.Hour)
	ul := tt.Add(-24 * time.Hour)

	if time.Since(tt) > 12*time.Hour {
		ll = ll.Add(24 * time.Hour)
		ul = ul.Add(24 * time.Hour)
	}

	condition := &EventCondition{
		LowerLimitOfEndtime: ll,
		UpperLimitOfEndtime: ul,
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
			err = GetAndSaveConfirmed(client, event, is_block, blockid)
			if err != nil {
				err = fmt.Errorf("GetAndSaveConfirmed failed: %w", err)
				return
			}
		}
	}

	return
}
