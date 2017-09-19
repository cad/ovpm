#!/bin/bash
export USER="nobody"
export GROUP="nobody"
getent passwd $USER || useradd $USER
getent group $GROUP || groupadd $GROUP
