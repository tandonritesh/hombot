set -x
go build -o intentexecutors/mqttexecutor/mqttexec.so -buildmode=plugin intentexecutors/mqttexecutor/mqttexec.go
sudo cp intentexecutors/mqttexecutor/mqttexec.so $GOBIN

go build -o hombotsrv/hbsrv hombotsrv/hombotSrv.go

sudo cp entities.json $GOBIN
