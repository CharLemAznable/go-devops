package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func StartHttpSever() {
	r := mux.NewRouter()
	r.HandleFunc(contextPath+"/log/{logger}/{timestampFrom}/{timestampTo}", FindHandleLogsBetweenTimestamps)
	r.HandleFunc(contextPath+"/machines", HandleMachines)
	r.HandleFunc(contextPath+"/logs", HandleLogs)
	r.HandleFunc(contextPath+"/favicon.ico", HandleFavicon)
	r.HandleFunc(contextPath+"/", HandleHome)

	http.Handle(contextPath+"/", r)

	fmt.Println("start to listen at ", httpPort)
	http.ListenAndServe(":"+httpPort, nil)
}
