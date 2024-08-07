# myselfgo

策略一：上轨反包做空下轨反包做多策略
反包的定义描述：后一根 k 线收盘价要小于（大于）最低点（最高点）。
条件：两根 k 线至少有一根的最高点和最低点的区间在上轨或者下轨。
进场点:分批 1 反包收盘价 2 反包开盘价
止损点：两根 k 线的最高点止损
止盈点：分批挂振幅倍数或者中轨值

master 主版本，定义为稳定版本

其它分支为开发分支，测试后需合并到 master 分支

# 编译 Linux 64 位可执行程序：

# X86

GOOS=linux GOARCH=386 go build -o myselfgo

# ARM

CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o myselfgo main.go

# 编译 Windows  64 位可执行程序：

# X86

CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o myselfgo main.go

# ARM

CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -o myselfgo main.go

# 编译 MacOS 64 位可执行程序

# X86

CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o myselfgo main.go

# ARM

CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o myselfgo main.go

# 生产环境运行

ip 地址：54.199.62.212

cd /gowww

1. 方法一
   nohup ./myselfgo -c configs/cfg.production.yaml &

2. 方法二
   ./control.sh start
