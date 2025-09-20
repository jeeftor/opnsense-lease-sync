#!/bin/bash

echo "This is a sync plugin for development purposes"

PROJECT=Dhcpsync
IP=192.168.1.103

# Copy the views
rsync -aza --delete --partial \
./mvc/app/views/OPNsense/${PROJECT}/ \
root@$IP:/usr/local/opnsense/mvc/app/views/OPNsense/${PROJECT}/

rsync -aza --delete --partial \
./mvc/app/models/OPNsense/${PROJECT}/ \
root@$IP:/usr/local/opnsense/mvc/app/models/OPNsense/${PROJECT}/

rsync -aza --delete --partial \
./mvc/app/controllers/OPNsense/${PROJECT}/ \
root@$IP:/usr/local/opnsense/mvc/app/controllers/OPNsense/${PROJECT}/

rsync -aza --delete --partial \
./service/templates/OPNsense/${PROJECT}/ \
root@$IP:/usr/local/opnsense/service/templates/OPNsense/${PROJECT}/
