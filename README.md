<img src="https://raw.githubusercontent.com/cherkesky/ndjson-filter/master/assets/logo.png" height="250" width="300">


## by Guy Cherkesky | [LinkedIn](http://linkedin.com/in/cherkesky) | [Website](http://cherkesky.com)

## Dockerized NDJSON S3 Filtering
#### Go implementation


This is a Go implementation of a filtering command line tool (CLI) that extract metadata that is represented as a sequence of [newline-delimited JSON](http://ndjson.org) objects. This simply means that the JSON objects are encoded without newlines (`'\n'`), and are separated by either a newline (`'\n'`) or a carriage-return (`'\r'`) followed by a newline (`'\n'`).

## Data

There are a number of objects in [AWS S3](https://aws.amazon.com/s3) that contain newline-delimited JSON data. Each JSON object contains several fields:

| Name | JSON Type | Description |
| ---- | ---- | ------------|
| `id` | `number` | A randomly generated 64-bit integer identifier. |
| `time` | `string` | A string containing an [RFC3339](https://tools.ietf.org/html/rfc3339) timestamp. |
| `words` | `array` | An array of strings containing randomly selected English words. |

Some examples of data:

```json
{"id": 6831346064777208556, "time": "2000-04-05T16:50:45", "words": ["vialing"]}
{"id": 5359261249102867223, "time": "2000-10-28T12:33:56", "words": ["calamined", "lepidopterist"]}
{"id": 5253074597765577362, "time": "2000-06-29T03:02:52", "words": ["realigns", "botanizer"]}
{"id": 4834718291991736770, "time": "2000-08-28T20:34:34", "words": ["nonplanar", "formee", "wavier", "haunches"]}
```

The objects in S3 are GZIP compressed.

## S3 Inputs

| S3 URI | Number of JSON |
| ------ | -------------- |
| s3://s3://ndjson-bucket/1.ndjson.gz | 1 |
| s3://s3://ndjson-bucket/100.ndjson.gz | 100 |
| s3://s3://ndjson-bucket/1000.ndjson.gz | 1000 |
| s3://s3://ndjson-bucket/100000.ndjson.gz | 100000 |

## Flags

The utility supports the following flags:

| Name | Required | Description |
| ---- | -------- | ----------- |
| `-input` | Yes | An S3 URI (`s3://{bucket}/{key}`) that refers to the source object to be filtered. |
| `-with-id` | No | An integer that contains the `id` of a JSON object to be selected. |
| `-from-time` | No | An RFC3339 timestamp that represents the earliest `time` of a JSON object to be selected. |
| `-to-time` | No | An RFC3339 timestamp that represents the latest `time` of JSON object to be selected. |
| `-with-word` | No | A string containing a word that must be contained in `words` of a JSON objec to be selected. |

If no `-input` flag is present, the utility will print a usage message and exit with a non-zero status. If only the `-input` flag is present, the output of the S3 object will be written to `stdout`. The `-with-id` flag will select only JSON objects that have the specified `id` field. The `-from-time` flag will select only objects with `time` fields greater than or equal to the specified value. The `-to-time` flag will select only objects with `time` fields less than the specified value. Finally, the `-with-word` flag will select only object where the `words` array contains the specified word. If multiple filtering flags are provided, the conjunction (AND) of the conditions will be applied.

## Examples
Docker command:
```bash
docker run --rm -e AWS_REGION=<will be provided> -e AWS_ACCESS_KEY_ID=<will be provided> -e AWS_SECRET_ACCESS_KEY=<will be provided> -input s3://ndjson-bucket/1000.ndjson.gz -from-time=2000-02-02T13:20:40 -to-time=2004-01-01T00:00:00
```
Output:
```bash
{1958328460630293988 2000-10-16T04:23:53 [roam glycerinating status ruggedize]}
{7373173639641088967 2000-05-24T07:24:13 [cosmologists adducts endemisms]}
{8852888256814150800 2000-01-31T03:50:29 [outgains prefreezing pervades drear]}
{5972579255125398127 2000-10-27T16:31:57 [polysyndeton citrulline bamboozles posttraumatic]}
{8430259396552752009 2000-10-23T17:09:09 [databased accelerando caudillismo]}
{7832874182759211418 2000-07-28T00:46:55 [spanner scrootch unscrew organisms]}
{6013077318414246325 2000-05-10T06:47:15 [postages scratchboard rondos downbeats debunks]}
{4518711722961714278 2000-05-07T20:33:32 [casualties boracic touted ascidium asocial]}
{3570732578126078709 2000-10-15T00:06:17 [rubricator babool]}
{8135075639657369229 2000-07-17T21:33:31 [antecedence bilharzias nodular ergots cholestyramines]}

```
## Environment

| Name | Value |
| ---- | ----- |
| `AWS_REGION` | `will be provided` |
| `AWS_ACCESS_KEY_ID` | `will be provided` |
| `AWS_SECRET_ACCESS_KEY` | `will be provided` |


## Docker commands:
Build:
```bash
docker build -t ndjsonfilter:v1 .
```

Run examples:
```bash
docker run --rm -e AWS_REGION=<will be provided> -e AWS_ACCESS_KEY_ID=<will be provided> -e AWS_SECRET_ACCESS_KEY=<will be provided> ndjsonfilter:v1 -input s3://ndjson-bucket/100000.ndjson.gz -with-word titans

docker run --rm -e AWS_REGION=<will be provided> -e AWS_ACCESS_KEY_ID=<will be provided> -e AWS_SECRET_ACCESS_KEY=<will be provided> ndjsonfilter:v1 -input s3://ndjson-bucket/100000.ndjson.gz -to-time 2002-02-02T11:32:32.102118268-07:00 -with-word titans

docker run --rm -e AWS_REGION=<will be provided> -e AWS_ACCESS_KEY_ID=<will be provided> -e AWS_SECRET_ACCESS_KEY=<will be provided> ndjsonfilter:v1 -input s3://ndjson-bucket/100000.ndjson.gz -from-time 1970-02-02T11:32:32.102118268-07:00 -to-time 2002-02-02T11:32:32.102118268-07:00 -with-word titans

docker run --rm -e AWS_REGION=<will be provided> -e AWS_ACCESS_KEY_ID=<will be provided> -e AWS_SECRET_ACCESS_KEY=<will be provided> ndjsonfilter:v1 -input s3://ndjson-bucket/100000.ndjson.gz -from-time 1970-02-02T11:2:32.102118268-07:00 -to-time 2002-02-02T11:32:32.102118268-07:00 -with-word titans -with-id 3516453660759435053
```

## User Permissions

```json
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Sid": "VisualEditor0",
			"Effect": "Allow",
			"Action": 
				"s3:GetObject",
			"Resource": 
				"arn:aws:s3:::ndjson-bucket/*"
		}
	]
}
```
