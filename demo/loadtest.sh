#!/bin/bash


while true;
do
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/
    # curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    # curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    # curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    echo "rest"
    sleep 300
done


echo "Load test complete"