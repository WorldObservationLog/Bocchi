package main

import (
	"errors"
	"github.com/telanflow/mps"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// A simple mitm proxy server
func main() {
	quitSignChan := make(chan os.Signal)
	proxy := mps.NewHttpProxy()
	LoadScripts()
	mitmHandler, err := mps.NewMitmHandlerWithCertFile(proxy.Ctx, "./cert/ca.crt", "./cert/ca.key")
	if err != nil {
		log.Panic(err)
	}
	proxy.HandleConnect = mitmHandler

	proxy.UseFunc(func(req *http.Request, ctx *mps.Context) (*http.Response, error) {
		log.Printf("[INFO] middleware -- %s %s", req.Method, req.URL)
		if req.Method != http.MethodConnect {
			handledReq := HandleRequest(*req)
			resp, err := ctx.Next(&handledReq)
			CheckErr(err)
			handledResp := HandleResponse(*resp)
			return &handledResp, nil
		}
		resp, err := ctx.Next(req)
		return resp, err
	})

	// Started proxy server
	srv := http.Server{
		Addr:    "localhost:8080",
		Handler: proxy,
	}
	go func() {
		log.Printf("MitmProxy started listen: http://%s", srv.Addr)
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		if err != nil {
			quitSignChan <- syscall.SIGKILL
			log.Fatalf("MitmProxy start fail: %v", err)
		}
	}()

	// quit signal
	signal.Notify(quitSignChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)

	<-quitSignChan
	_ = srv.Close()
	log.Fatal("MitmProxy server stop!")
}
