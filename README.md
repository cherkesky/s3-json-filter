<img src="https://raw.githubusercontent.com/cherkesky/inmemcache/master/logo.png" height="250" width="250">


### by Guy Cherkesky | [LinkedIn](http://linkedin.com/in/cherkesky) | [Website](http://cherkesky.com)

### Dockerized NDJSON S3 Filtering
#### Go implementation

## Docker commands:

Build:
```bash
docker build -t ndjsonfilter:v1 .
```

Run examples:
```bash
docker run --rm -e AWS_REGION -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY ndjsonfilter:v1 -input s3://ndjson-bucket/1M.ndjson.gz -with-word titans

docker run --rm -e AWS_REGION -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY ndjsonfilter:v1 -input s3://ndjson-bucket/1M.ndjson.gz -to-time 1990-02-02T11:32:32.102118268-07:00 -with-word titans

docker run --rm -e AWS_REGION -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY ndjsonfilter:v1 -input s3://ndjson-bucket/1M.ndjson.gz -from-time 1970-02-02T11:32:32.102118268-07:00 -to-time 1980-02-02T11:32:32.102118268-07:00 -with-word titans

docker run --rm -e AWS_REGION -e AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY ndjsonfilter:v1 -input s3://ndjson-bucket/1M.ndjson.gz -from-time 1970-02-02T11:2:32.102118268-07:00 -to-time 2022-02-02T11:32:32.102118268-07:00 -with-word titans -with-id 4151711053299985798
```
