# !!!Minimum requirement:
you have to provide .otsample.yml. You can either put under your home folder as `$HOME/.otsample.yml` or you can provide with `--config <.otsample.yml path>`

* `go run frontend.go`
* `go run backend.go`

hit through endpoint and test the app like:

* `curl -X PUT http://localhost:8080/hello/Koray -d '{ "dateOfBirth": "2000-10-01" }'`
* `curl http://localhost:8080/hello/Koray`
---

## Use Cases:
----

If you run default Jaeger at app for testing purpose, it's port is exposed on `16686`. You can reach through `http://localhost:16686/search`

### case1: happy birthday

- Req: `curl -X PUT http://localhost:8080/hello/Koray -d '{ "dateOfBirth": "2000-09-01" }'`
- Req: `curl http://localhost:8080/hello/Koray` 
- Resp: `{"message":"Hello, Koray! Happy birthday"}` 
---
### case2: 1 day left

- Req: `curl -X PUT http://localhost:8080/hello/Koray -d '{ "dateOfBirth": "2000-09-02" }'`
- Req: `curl http://localhost:8080/hello/Koray` 
- Resp: `{"message":"Hello, Koray! Your birthday is in 1 day"}` 
---
### case3: more then 1 day to birthday

- Req: `curl -X PUT http://localhost:8080/hello/Koray -d '{ "dateOfBirth": "2000-09-05" }'`
- Req: `curl http://localhost:8080/hello/Koray` 
- Resp: `{"message":"Hello, Koray! Your birthday is in 4 days"}` 

---
### case4: invalid date format

- Req: `curl -X PUT http://localhost:8080/hello/Koray -d '{ "dateOfBirth": "2000-09-5" }'`
- Resp: `{"error": "2000-09-5 date format is invalid, date must be in yyyy-mm-dd format"}` 
