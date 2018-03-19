#!/bin/bash
export USER="nobody"
export GROUP="nogroup"
id -u $USER &>/dev/null || useradd $USER
id -g $GROUP &>/dev/null || groudadd $GROUP

systemctl daemon-reload
