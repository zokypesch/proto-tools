package commands

import (
	"bytes"
	"log"
	"os/exec"
)

// Install struct for creating proto from DB
type Install struct{}

// NewInstall for new protofrom db
func NewInstall() CommandInterfacing {
	return &Install{}
}

// Execute for executing command
func (cmd *Install) Execute(args map[string]string) error {

	log.Println("installing tools")
	var res, errOut string
	// var err error
	// var res []byte
	// err = shellExecute("go get -u google.golang.org/grpc")
	// if err != nil {
	// 	return err
	// }
	res, errOut, _ = shellout("go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway")
	if len(res) > 0 || len(errOut) > 0 {
		log.Println(res, errOut)
	}
	res, errOut, _ = shellout(`go install github.com/golang/protobuf/protoc-gen-go `)
	if len(res) > 0 || len(errOut) > 0 {
		log.Println(res)
	}
	res, errOut, _ = shellout(`go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger `)
	if len(res) > 0 || len(errOut) > 0 {
		log.Println(res)
	}
	res, errOut, _ = shellout("go get github.com/zokypesch/proto-lib")
	if len(res) > 0 || len(errOut) > 0 {
		log.Println(res)
	}
	res, errOut, _ = shellout("go install github.com/zokypesch/protoc-gen-generator")
	if len(res) > 0 || len(errOut) > 0 {
		log.Println(res)
	}
	log.Println(`to install sangkuriang please run "echo 'sangkuriang() {
		protoc -I $1 $2.proto --generator_out=$3 -I=$GOPATH/src -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --go_out=plugins=grpc:$3 --grpc-gateway_out=logtostderr=true:$3
		dep ensure -v
		echo "files has been generate"
	  }' >> ~/.bashrc" or to your .zshrc`)

	log.Println("installation complete")
	return nil
}

func shellExecute(cmd string) error {
	cmdex := exec.Command(cmd)
	stdout, err := cmdex.Output()

	if err != nil {
		log.Println(err.Error())
		return err
	}

	log.Println(string(stdout))
	return nil
}

func bashExecute(cmd string, shell bool) []byte {
	if shell {
		out, err := exec.Command("bash", "-c", cmd).Output()
		if err != nil {
			panic("some error found")
		}
		return out
	} else {
		out, err := exec.Command(cmd).Output()
		if err != nil {
			panic("some error found")
		}
		return out

	}
}

const shellToUse = "bash"

func shellout(param string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(shellToUse, "-c", param)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}
