#! /bin/bash

# このシェルスクリプトはイベント参加ユーザーのユーザー情報を更新するためのものです。
# イベント確定結果を保存するまえに余裕をもって(例えばイベント終了日翌日の未明あるいは早朝に)起動します。

cd /home/chouette/MyProject/Showroom/SaveConfirmedData

source ./my_script.env

./SaveConfirmedData Uinf
