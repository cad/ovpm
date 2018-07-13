#!/bin/bash
export USER="nobody"
export GROUP="nogroup"
id -u $USER &>/dev/null || useradd $USER
id -g $GROUP &>/dev/null || groupadd $GROUP

systemctl daemon-reload
