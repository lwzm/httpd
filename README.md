# HTTP servers

### Notice

```sh
go get github.com/lwzm/httpd/notice
notice
```

Test:

```sh
# shell 1
$ curl localhost:1111/zero >/dev/null 
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 1313T    0 1313T    0     0  1389M      0 --:--:--  11d 11h --:--:-- 1462M

# shell 2
$ curl -T /dev/zero localhost:1111
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100 1313T    0     0    0 1313T      0  1389M --:--:--  11d 11h --:--:-- 1447M
```

Pass url argument `broadcast` to implement pub/sub:

```sh
curl -T /dev/zero localhost:1111/zero?broadcast
```

### Agency

```sh
go get github.com/lwzm/httpd/agency
```

Apis:

* `GET /ANY`
* `POST /ANY`
* `POST /@/ID`

Headers:
* `Location`
* `Content-Type`
