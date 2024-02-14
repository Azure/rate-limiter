#!/bin/bash


for i in $(seq 1 100);
do
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
    curl -i -X POST -d "{\"billingAccount\":\"newuser$i\"}" localhost:8080/billingAccount/ >/dev/null 2>&1 &
done


echo "Load test complete"