package main

import (
	"context"
	"fmt"
	pb "gitlab.bixin.com/mili/node-proxy/api/platform/v1"
	"google.golang.org/grpc"
)

func main(){
	conn,err := grpc.Dial("127.0.0.1:9001",grpc.WithInsecure())
	if err != nil {
		fmt.Println("error:",err)
	}
	defer conn.Close()
	p := pb.NewPlatformClient(conn)
	req := new(pb.GetBalanceRequest)
	req.Chain = "ETH"
	req.Address = "0xa06ef134313C13e03B8682B0616147607B4E375E"
	req.TokenAddress = "0xdAC17F958D2ee523a2206206994597C13D831ec7"
	req.Decimals = "6"
	resp,err := p.GetBalance(context.Background(),req)
	if err != nil {
		fmt.Println("get balacne error",err)
	}
	fmt.Println("result:",resp.Balance)
}


