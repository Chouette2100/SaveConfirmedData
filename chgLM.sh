#!/bin/bash

nfn=`ls -tr SaveConfirmedData.*|tail -1`; echo $nfn
nbe=$(basename "$nfn");echo $nbe

# systemctl --user stop SaveConfirmedData
rm SaveConfirmedData
ln -s ./$nbe SaveConfirmedData
ls -l
# systemctl --user start SaveConfirmedData
# systemctl --user status SaveConfirmedData
