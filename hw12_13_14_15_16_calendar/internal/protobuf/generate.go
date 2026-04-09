//go:generate protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative --proto_path=. event.proto create_event.proto update_event.proto delete_event.proto list_events.proto calendar_service.proto

package protobuf
