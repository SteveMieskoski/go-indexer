package engine

import (
	"src/types"
	"testing"
)

func Test_rpcCall(t *testing.T) {
	type args struct {
		url     string
		headers map[string]string
		method  string
		params  Params
	}

	tt := args{
		url:     "http://127.0.0.1:8545",
		headers: map[string]string{},
		method:  "eth_getBlockByNumber",
		params:  Params{"latest", true},
	}
	call, err := rpcCall[types.Block](tt.url, tt.headers, tt.method, tt.params)
	if err != nil {
		println("ERROR")
		println(err)
		return
	}

	println("Result")
	println(call.String())
	//type testCase[T any] struct {
	//	name    string
	//	args    args
	//	want    *T
	//	wantErr bool
	//}
	//tests := []testCase[ types.Block ]{
	//	// TODO: Add test cases.
	//}
	//for _, tt := range tests {
	//	t.Run(tt.name, func(t *testing.T) {
	//		got, err := rpcCall(tt.args.url, tt.args.headers, tt.args.method, tt.args.params)
	//		if (err != nil) != tt.wantErr {
	//			t.Errorf("rpcCall() error = %v, wantErr %v", err, tt.wantErr)
	//			return
	//		}
	//		if !reflect.DeepEqual(got, tt.want) {
	//			t.Errorf("rpcCall() got = %v, want %v", got, tt.want)
	//		}
	//	})
	//}
}
