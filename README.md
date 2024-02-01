# rate-limiter-backed-by-redis-cache
## rate limiter workflow

![ratelimiter design (2)](https://github.com/Xinyue-Wang/rate-limiter-backed-by-redis-cache/assets/37516611/529faeec-7269-4701-9ac9-f97914551020)

## counter
aggregated counting + local counting
![Counter design with cache](https://github.com/Xinyue-Wang/rate-limiter-backed-by-redis-cache/assets/37516611/0bf4e623-17eb-4466-9a49-40fdddba71f8)

## remote store
### Option1: Azure Redis Cache


### Option2: Redis cluster


## rate limiting algorithm
### Option1: token bucket
 ![image](https://github.com/Xinyue-Wang/token_bucket_cache/assets/37516611/27cf75b1-2198-466b-9f57-a26a82f40c0e)

#### Implementation of a self-maintained token bucket cache:
Prerequist: have a redis cache to store key-value pair

Goal:  
1. Provide TakeToken and GetBucketStats api
2. No need to a separte process to keep adding token to bucket or remove key-value pair from cache to prevent overuse memory

Implementation:
1. Start each bucket full:
![image](https://github.com/Xinyue-Wang/token_bucket_cache/assets/37516611/661b1819-d24e-4a06-b6d8-f1f508b43be2)
2. Each bucket auto expired after reach max tokens:
![image](https://github.com/Xinyue-Wang/token_bucket_cache/assets/37516611/faa8a8ee-4f6d-4a8f-bdbb-e32460e47901)
3. Only need to save token number and timestamp when bucket reached saved token number
![image](https://github.com/Xinyue-Wang/rate-limiting-with-distributed-cache/assets/37516611/87df442d-2048-45f7-94be-f0fcbc01486c)


## run the prototype
### With azure redis example

1. Create an Azure Redis Cache
```shell
export AZURE_SUBSCRIPTION_ID=8ecadfc9-d1a3-4ea4-b844-0d9f87e4d7c8
export AZURE_RESOURCE_GROUP=xinywaTestRG
export location=eastus

# create resouregroup
az group create --name $AZURE_RESOURCE_GROUP --location $location --subscription $AZURE_SUBSCRIPTION_ID

# create azure redis cache
export AZURE_REDIS_NAME=xinywaRedisCache
az redis create --location $location --name $AZURE_REDIS_NAME --resource-group $AZURE_RESOURCE_GROUP --subscription $AZURE_SUBSCRIPTION_ID --sku Basic --vm-size c0
```

2. Set them to the respective environment variables

```shell
set AZURE_SUBSCRIPTION_ID=8ecadfc9-d1a3-4ea4-b844-0d9f87e4d7c8
set AZURE_RESOURCE_GROUP=xinywaTestRG
set AZURE_REDIS_NAME=xinywaRedisCache
```
3. Clone the repo. Then in the terminal, run the following command to start the application.

```shell
cd cmd/azureredis
go run main.go
```
The HTTP server will start on port `8080`.

4. Send request to test cache and throttle:

   Create request with billingAccount id1:
   ```bash
   curl -i -X POST -d '{"billingAccount":"id1"}' localhost:8080/billingAccount/
   ``` 

   Get bucket stats for billingAccount id1:
   ```bash
   curl -i localhost:8080/billingAccount/id1
   ```

### With redis cluster example
1. Deployer redis cluster (minimum 6 nodes)
```
kubectl apply -f rediscluster/redis-cluster.yaml
kubectl apply -f rediscluster/redis-configmap.yaml
kubectl apply -f rediscluster/redis-service.yaml
rediscluster/roles.sh
```

2. Deploy test server and expose service
```
kubectl apply -f cmd/rediscluster/template/deployment.yaml
kubectl expose deployment rate-limit-server --name=rate-limit-svc --port=8080 --target-port=8080 --type=NodePort
kubectl port-forward svc/rate-limit-svc 8080:80
```
3. Send request to test cache and throttle:

   Create request with billingAccount id1:
   ```bash
   curl -i -X POST -d '{"billingAccount":"id1"}' localhost:8080/billingAccount/
   ``` 

   Get bucket stats for billingAccount id1:
   ```bash
   curl -i localhost:8080/billingAccount/id1
   ```

  
