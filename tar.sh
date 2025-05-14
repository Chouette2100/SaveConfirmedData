#! /bin/bash
filename=`date +%Y%m%d-%H%M`
tar zcvf SaveConfirmedData_$filename.tar.gz \
DBConfig.yml \
Env.yml \
my_script.env \
sdat.sh \
uinf.sh \
SaveConfirmedData
