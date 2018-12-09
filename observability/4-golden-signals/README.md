# bcrypt-worker

An HTTP server to hash passwords using [bcrypt](https://en.wikipedia.org/wiki/Bcrypt). 

## Getting Started

- Start all dependencies and the worker service using docker-compose
  ```
  $ make stack WORKER_NUM_DECRYPTERS=2
  ```

### Check a password
```
$ make ping-server
curl \
        -X POST \
        -H "Content-Type: application/json" \
        -d @tests/fixtures/password_no_match.json \
        http://localhost:8080/decrypt -v
Note: Unnecessary use of -X or --request, POST is already inferred.
*   Trying 127.0.0.1...
* Connected to localhost (127.0.0.1) port 8080 (#0)
> POST /decrypt HTTP/1.1
> Host: localhost:8080
> User-Agent: curl/7.47.0
> Accept: */*
> Content-Type: application/json
> Content-Length: 102
>
* upload completely sent off: 102 out of 102 bytes
< HTTP/1.1 200 OK
< Content-Type: application/json
< Date: Fri, 02 Nov 2018 23:23:36 GMT
< Content-Length: 16
<
{"Match":false}
* Connection #0 to host localhost left intact
```

### Observability
The bcrypt worker comes with a prometheus dashboard and is accessible @`localhost:3000` when using `make stack`.

<img width="1313" alt="screen shot 2018-10-30 at 8 04 45 pm" src="https://user-images.githubusercontent.com/321963/47945296-5fb78e00-ded7-11e8-9006-cd7675ef3d24.png">

This includes a number of metrics which should allow for the operation of the worker service and actionable metrics which should help identify and alert when key SLOs (avaialbility, latency) are being violated.
Ther metrics included are:
- Availability from probe results (covered later)
- Rate of HTTP Requests
- Duration of HTTP Requests
- Rate & Result (match|nomatch) of bcrypt hashing
- Latency of bcrypt hashing
- Decrypter Pool Queue Saturation
- System / Goruntime metrics

#### Launch a load test
In order to understand the performance of thew worker service a load test can be executed with:

```
$ make load-test LOAD_TEST_RATE=20
echo "POST http://localhost:8080/decrypt" | vegeta attack -body tests/fixtures/password_no_match.json -rate=20 -duration=0 | tee results.bin | vegeta report
Requests      [total, rate]            4337, 20.00
Duration      [total, attack, wait]    3m36.864626198s, 3m36.801482041s, 63.144157ms
Latencies     [mean, 50, 95, 99, max]  66.720569ms, 65.657269ms, 72.371326ms, 82.587951ms, 114.292254ms
Bytes In      [total, mean]            138320, 31.89
Bytes Out     [total, mean]            446711, 103.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:4337
Error Set:
```
(the load test was used to generate the metric screenshots above)

### Availability (Service Health)
[Probing](https://medium.com/dm03514-tech-blog/sre-availability-probing-101-using-googles-cloudprober-8c191173923c) is being used in order to determine if the service is availble an HTTP probe is being executed at a 1 minute interval:

![screen shot 2018-11-01 at 3 58 09 pm](https://user-images.githubusercontent.com/321963/47885866-88735100-de0d-11e8-9e93-1f15df135179.png)

### SLO 
- All requests complete < 100ms.
  - In order to monitor the duration each HTTP request is measured 
  - We can then alert if any are > 100ms 
  - <img width="626" alt="screen shot 2018-11-02 at 7 43 11 pm" src="https://user-images.githubusercontent.com/321963/47945341-9e4d4880-ded7-11e8-81ad-faaf9e8d24f4.png">
  
- Availabilty
   - [Cloudprober](https://cloudprober.org/) is configured to make a request every 1 minute (very much like a local pingdom/new relic synthetics)
   - The availability is calculated and displayed as a metric on the dashboard:
   - <img width="731" alt="screen shot 2018-11-02 at 7 47 19 pm" src="https://user-images.githubusercontent.com/321963/47945420-46631180-ded8-11e8-920a-fd7d27801bb8.png">
   - This should allow us to model the availability in terms of 9's and alert accordingly

### Running Tests
- The server has unit tests written in go and executable with:

```
$ make test-unit
go test github.com/dm03514/bcrypt-worker/cmd/worker github.com/dm03514/bcrypt-worker/decrypt -v
?       github.com/dm03514/bcrypt-worker/cmd/worker     [no test files]
=== RUN   TestBcrypter_IsMatch_NoMatch
--- PASS: TestBcrypter_IsMatch_NoMatch (0.07s)
=== RUN   TestBcrypter_IsMatch_True
--- PASS: TestBcrypter_IsMatch_True (0.14s)
=== RUN   TestPool_IsMatch_FalseNoMatch
--- PASS: TestPool_IsMatch_FalseNoMatch (0.08s)
PASS
ok      github.com/dm03514/bcrypt-worker/decrypt        (cached)
```

## Using the JS Client
- Start a stack following the directions above
- Execute the client through the bundled `cli.js` interface:
```
  bcrypt-worker/client/js$ nodejs --version
v10.13.0
  bcrypt-worker/client/js$ nodejs cli.js -w http://localhost:8080/decrypt -p nomatch -h $2b$10$//DXiVVE59p7G5k/4Klx/ezF7BI42QZKmoOD0NDvUuqxRE5bFFBLy
{ workerAddr: 'http://localhost:8080/decrypt',
  password: 'nomatch',
  hash: 'b0$//DXiVVE59p7G5k/4Klx/ezF7BI42QZKmoOD0NDvUuqxRE5bFFBLy' }
success:  CompareResult { match: false }
```
  
### Executing Unit Tests:
```
bcrypt-worker/client/js$ make test-unit
./node_modules/mocha/bin/mocha


  Client
    compare()
Client.compare result:  { Match: false }
CompareResult.constructor:  { Match: false }
      ✓ should resolve a compare result when successful
Client.compare result:  { INVALID_KEY: true }
CompareResult.constructor:  { INVALID_KEY: true }
      ✓ should reject an error when CompareResult parsing is invalid
      - should reject an error when an a transport error is encountered


  2 passing (10ms)
  1 pending
```

## Performance/Operation

### Hasher Pool
- Since hashing is CPU bound the worker is limited by the number of CPU's available to it.  
When the CPUs become saturated the HTTP connections will begin to queue.  The decrypter Pool queue is exposed as a prometheus metric and is available on the dashboard:
- <img width="629" alt="screen shot 2018-11-02 at 7 53 25 pm" src="https://user-images.githubusercontent.com/321963/47945542-10725d00-ded9-11e8-99a2-3ccb298b9916.png">
- The decrypter pool queue should be flatline at 0, if not that means that clients are waiting on the pool to hash their requests, we discussed a couple of strategies to deal with this:
  - Shed load whenever the pool is saturated and return 429 status code:
  ```
  select {
  case:
    p.inputChan <- work
  default:
    fmt.Println("Channel Saturated!!! Return something to alert client!")
  }
  ```
  - We could then have a retry/backoff in the client on 429.  This could allow for the majority of server latencies to remain  steady at the expense of some clients being aborted and being forced to retry.

## 0 Downtime Deploys
- Amazon Load Balancing policy could be used to accomplish this in a blue/green variant ("expand/contract")
- Deployment takes places
- New version is brought up along side the old version (both behind the LB)
- Wait until new version health checks pass
- New Version is registerd with LB for traffic
- Old version begins draining
  - No new connections to old version
  - Draining timeout (300seconds default) after which old version is killed
- New version is now accepting all traffic

## Spreading Workload
- Could be accomplished through load balancing over all the bcrypt worker instances
- Bcrypt worker instances are stateless and lend extremely well to horizontal scaling
- Additionally, CPU usage is a great signal for autoscaling of these workers
