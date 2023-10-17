# token_bucket_cache

## throttle algorithm
 ![image](https://github.com/Xinyue-Wang/token_bucket_cache/assets/37516611/27cf75b1-2198-466b-9f57-a26a82f40c0e)

## self-maintained cache
1. Start each bucket full:
![image](https://github.com/Xinyue-Wang/token_bucket_cache/assets/37516611/661b1819-d24e-4a06-b6d8-f1f508b43be2)
2. Each bucket auto expired after reach max tokens:
![image](https://github.com/Xinyue-Wang/token_bucket_cache/assets/37516611/faa8a8ee-4f6d-4a8f-bdbb-e32460e47901)
3. Only need to save token number and timestamp when bucket reached saved token number
![image](https://github.com/Xinyue-Wang/token_bucket_cache/assets/37516611/5839180f-da50-4eef-8aae-0bd7a0150050)


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
  
