export USER="nobody"
export GROUP="nobody"
id -u $USER &>/dev/null || useradd $USER
id -u $GROUP &>/dev/null || useradd $GROUP

systemctl daemon-reload
