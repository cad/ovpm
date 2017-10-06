#!/bin/bash
if [ "`systemctl is-active ovpmd`" != "active" ]
then
    systemctl daemon-reload
    systemctl restart ovpmd
fi
