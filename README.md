# grps
Golang Reverse Proxy Server

## installation

### 1. linux
```
# tar -zxf grps_rel_linux_amd64_1.0.1.0.tar.gz
# bin/grps -install
# bin/grps -start
# bin/grps -status
```

###### In order not to limit the number of connections, add the following line for service
_/etc/systemd/system/grps.service_
```
[Unit]
...

[Service]
...
LimitNOFILE=65535

[Install]
...
```

### 2. windows
1) _unzip the release file grps_rel_windows_amd64_1.0.1.0.zip_
2) _run cmd as administrator_
```
> bin\grps.exe -install
> bin\grps.exe -start
> bin\grps.exe -status
```

## web admin
[http://127.0.0.1:9618/](http://127.0.0.1:9618/)

login width following account and password
```
account: admin
password: 1
```