package main

import (
	"fmt"
	"os"

	"github.com/rsheasby/slog"
)

func exampleReq(mode slog.SloggerMode) (req *slog.SloggerRequest) {
	s := slog.Slogger{
		Writer: os.Stdout,
		Mode:   mode,
	}
	req = s.NewRequest(func(sr *slog.SloggerRequest) {
		sr.ClientHost = "127.0.0.1"
		sr.HttpMethod = "POST"
		sr.HttpPath = "/api/v1/testing"
		sr.HttpStatusCode = 200
		sr.ResponseSize = 35189
		sr.ExtraData["Funny_Number_lol"] = 69
		sr.ExtraData["Favorite_Colour"] = "Purple"
		sr.ExtraData["Slice"] = []int{1, 2, 3}
	})
	req.Info("info example")
	req.Warning("warning example")
	req.Error("error example")
	req.WTF("WTF example")
	return
}

func main() {
	fmt.Println("Development output:")
	exampleReq(slog.ModeDevelopment).WriteLogs()

	fmt.Println("\nProduction output:")
	exampleReq(slog.ModeProduction).WriteLogs()
}
