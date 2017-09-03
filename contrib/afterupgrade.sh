if [ "`systemctl is-active ovpmd`" != "active" ]
then
    systemctl restart ovpmd
fi
