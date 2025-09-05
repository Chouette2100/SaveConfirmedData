// Copyright © 2025 chouette2100@gmail.com
// Released under the MIT license
// https://opensource.org/licenses/mit-license.php

// 指定した条件にあうイベントをイベントテーブルから抽出する。
// この関数は、イベントテーブルからイベントを抽出するための関数である。
// 引数として、抽出条件を指定するための構造体を受け取り、その条件にあうイベントを抽出する。
// 抽出したイベントは、イベントテーブルから抽出されたイベントの構造体の配列として返す。
package main

import (
	"fmt"
	"time"

	"github.com/Chouette2100/srdblib/v2"
)

type EventCondition struct {
	LowerLimitOfEndtime time.Time
	UpperLimitOfEndtime time.Time
	Sdat                bool
	Uinf                bool
}

// イベントの開始日時の範囲
// ExtractEvents 指定した条件にあうイベントをイベントテーブルから抽出する。
func ExtractEvents(
	condition *EventCondition,
) (
	// events []*srdblib.Wevent,
	intf []interface{},
	err error,
) {

	sqlst := ""
	if cl, ok := clmlist["event"]; !ok {
		err = fmt.Errorf("clmlist['event'] is empty")
		return
	} else {
		// カラムが存在する場合の処理
		sqlst = "SELECT " + cl + " FROM event WHERE "
	}
	sqlst = "SELECT " + clmlist["event"] + " FROM event WHERE "
	sqlst += " rstatus != 'Confirmed' "
	if !condition.LowerLimitOfEndtime.IsZero() {
		sqlst += " AND endtime >= ? "
	}
	if !condition.UpperLimitOfEndtime.IsZero() {
		sqlst += " AND endtime <= ? "
	}
	// var intf interface{}
	// uloe := condition.UpperLimitOfEndtime
	// if condition.OnlyUinf {
	// 	// ユーザ情報の更新のみを行う場合は、プログラム実行時時点で終了したイベントを対象とする
	// 	uloe = time.Now()
	// }
	intf, err = srdblib.Dbmap.Select(&srdblib.Event{}, sqlst, condition.LowerLimitOfEndtime, condition.UpperLimitOfEndtime)
	// intf, err = srdblib.Dbmap.Select(&srdblib.Event{}, sqlst, condition.LowerLimitOfEndtime, uloe)
	if err != nil {
		err = fmt.Errorf("srdblib.Dbmap.Select failed: %w", err)
		return
	}
	// var pintf = &intf
	// tmp := intf.([]*srdblib.Wevent)
	// events = tmp

	return
}
