package main

import (
	"context"
	"fmt"
	"io"
	"log"

	// mavsdk "github.com/ykhedar/MAVSDK-Go/Sources/MAVSDK-Go/Generated"
	// action "github.com/ykhedar/MAVSDK-Go/action"
	telemetry "github.com/ykhedar/MAVSDK-Go/sources/telemetry"
	"google.golang.org/grpc"
)

func serverGet(cc *grpc.ClientConn) *telemetry.PositionResponse {
	newTelemetryServiceClient := telemetry.NewTelemetryServiceClient(cc)
	ctx := context.Background()
	req := &telemetry.SubscribePositionRequest{}

	gpsresponseInfo, err := newTelemetryServiceClient.SubscribePosition(ctx, req)
	if err != nil {
		fmt.Printf("Error %v", err)
	}
	message, _ := gpsresponseInfo.Recv()
	return message
}

func serverstream(cc *grpc.ClientConn) {
	for {
		fmt.Println("calling server stream")
		newTelemetryServiceClient := telemetry.NewTelemetryServiceClient(cc)
		ctx := context.Background()
		req := &telemetry.SubscribePositionRequest{}
		stream, err := newTelemetryServiceClient.SubscribePosition(ctx, req)
		if err != nil {
			fmt.Printf("%v.ListFeatures(_) = _, %v", cc, err)
		}

		message, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Printf("%v.ListFeatures(_) = _, %v", cc, err)
		}

		fmt.Printf("message %v", message)
	}

}

func clientStream(cc *grpc.ClientConn) {
	fmt.Println("calling server stream")
	newTelemetryServiceClient := telemetry.NewTelemetryServiceClient(cc)
	ctx := context.Background()
	requests := []telemetry.SubscribePositionRequest{
		telemetry.SubscribePositionRequest{},
	}
	gpsresponseInfo, err := newTelemetryServiceClient.SubscribePosition(ctx, &requests[0])
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", cc, err)
	}
	for _, r := range requests {
		if err := gpsresponseInfo.SendMsg(r); err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalf("%v.Send(%v) = %v", cc, r, err)
		}
	}
	reply := gpsresponseInfo.CloseSend()
	//  repase close send with receive CloseAndRecv()
	if err != nil {
		log.Fatalf("%v.CloseAndRecv() got error %v, want %v", cc, err, nil)
	}
	log.Printf("Route summary: %v", reply)

}

func biDirectionStream(cc *grpc.ClientConn) {
	fmt.Println("calling server stream")
	newTelemetryServiceClient := telemetry.NewTelemetryServiceClient(cc)
	ctx := context.Background()
	requests := []telemetry.SubscribePositionRequest{
		telemetry.SubscribePositionRequest{},
	}
	stream, err := newTelemetryServiceClient.SubscribePosition(ctx, &requests[0])
	if err != nil {
		log.Fatalf("%v.RecordRoute(_) = _, %v", cc, err)
	}
	waitc := make(chan struct{})
	go func() {
		for _, req := range requests {
			if err := stream.SendMsg(req); err != nil {
				log.Fatalf("Failed to send a note: %v", err)
			}
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			responseStream, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("Failed to receive a note : %v", err)
			}
			log.Printf("&v", responseStream)
		}
	}()

	<-waitc
}

func main() {
	dialoption := grpc.WithInsecure()

	serverAddr := "127.0.0.1:50051"
	cc, err := grpc.Dial(serverAddr, dialoption)
	if err != nil {
		fmt.Printf("Error while dialing %v", err)
	}
	grpc.ConnectionTimeout(5)
	// chanResp := make(chan telemetry.Position, 5)
	fmt.Printf("%v\n", serverGet(cc))

	serverstream(cc)
	// count := 0
	// for {
	// 	count++
	// 	if count == 5 {
	// 		break
	// 	}
	// 	value := <-chanResp
	// 	fmt.Printf("Response %v\n", value)
	// }
	// close(chanResp)
}
