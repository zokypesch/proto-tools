# how to run??

# please install first
proto-tools -cmd=install
# generate proto from your database
proto-tools -cmd=gen-proto-db host=localhost name=transaction user=root password=

# PR
- Schema generator
    - Read Comment
        Parsing get required hit
    - Table Comment
        Whitelist parsing
- lewati DO_NOT_REMOVE when re generate
- add options whitelist
- joins Table
- get foreign key
- enchance svc
- add whitelist
- add option db name- 

git clone git@github.com:golang/protobuf.git && (cd protobuf && git checkout v1.2.0 && go build -o $GOBIN/protoc-gen-go ./protoc-gen-go) && rm -r protobuf