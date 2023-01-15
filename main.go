package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/telanflow/mps"
)

// A simple mitm proxy server
func main() {
	quitSignChan := make(chan os.Signal)
	proxy := mps.NewHttpProxy()
	LoadScripts()
	mitmHandler, err := mps.NewMitmHandlerWithCertFile(proxy.Ctx, "./cert/ca.crt", "./cert/ca.key")
	if err != nil {
		CheckErr(err)
	}
	proxy.HandleConnect = mitmHandler

	proxy.UseFunc(func(req *http.Request, ctx *mps.Context) (*http.Response, error) {
		if req.Method != http.MethodConnect {
			log.Printf("[INFO] %s %s", req.Method, req.URL)
			handledReq := HandleRequest(*req)
			if handledReq.Method == "" {
				log.Println("[WARN] the request is empty")
				err := errors.New("the request is empty")
				return nil, err
			}
			resp, err := ctx.Next(&handledReq)
			CheckErr(err)
			handledResp := HandleResponse(*resp)
			return &handledResp, nil
		}
		return nil, mps.MethodNotSupportErr
	})

	// Started proxy server
	srv := http.Server{
		Addr:    "localhost:8080",
		Handler: proxy,
	}
	go func() {
		log.Printf("[INFO] Bocchi server started listen: http://%s", srv.Addr)
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			return
		}
		if err != nil {
			quitSignChan <- syscall.SIGKILL
			log.Fatalf("[FATAL] Bocchi server start fail: %v", err)
		}
	}()

	// quit signal
	signal.Notify(quitSignChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)

	<-quitSignChan
	_ = srv.Close()
	log.Fatal("[INFO] Bocchi server stop!")
}
