#! /bin/bash

# このシェルスクリプトはイベントの確定結果を保存するためのものです
# 確定結果発表（イベント終了日の翌日12:00頃）の直後に起動します。

cd /home/chouette/MyProject/Showroom/SaveConfirmedData

source my_script.env

./SaveConfirmedData Sdat
