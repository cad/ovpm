#!/bin/bash
export USER="nobody"
export GROUP="nogroup"
getent passwd $USER || useradd $USER
getent group $GROUP || groupadd $GROUP
