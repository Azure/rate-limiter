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

   
  
