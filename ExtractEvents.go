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

	"github.com/Chouette2100/srdblib"
)

type EventCondition struct {
	LowerLimitOfEndtime time.Time
	UpperLimitOfEndtime time.Time
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

	sqlst := "SELECT * FROM event WHERE "
	sqlst += " rstatus != 'Confirmed' "
	if !condition.LowerLimitOfEndtime.IsZero() {
		sqlst += " AND endtime >= ? "
	}
	if !condition.UpperLimitOfEndtime.IsZero() {
		sqlst += " AND endtime <= ? "
	}
	// var intf interface{}
	intf, err = srdblib.Dbmap.Select(&srdblib.Event{}, sqlst, condition.LowerLimitOfEndtime, condition.UpperLimitOfEndtime)
	if err != nil {
		err = fmt.Errorf("srdblib.Dbmap.Select failed: %w", err)
		return
	}
	// var pintf = &intf
	// tmp := intf.([]*srdblib.Wevent)
	// events = tmp

	return
}
