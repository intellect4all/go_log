
compile:
	protoc -I=. --go_out=. --go_opt=paths=source_relative --proto_path=.  api/v1/*.proto