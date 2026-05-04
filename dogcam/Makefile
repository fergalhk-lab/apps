.PHONY: proto

proto:
	cd .. && protoc \
	  --go_out=. --go_opt=module=github.com/fergalhk-lab/apps \
	  --go-grpc_out=. --go-grpc_opt=module=github.com/fergalhk-lab/apps \
	  --proto_path=dogcam/proto \
	  dogcam/proto/dogcam.proto
