package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"net/http"
)

func StartHttpSever() {
	r := mux.NewRouter()

	r.HandleFunc(contextPath+"/font/{fontName}", HandleFont)
	r.HandleFunc(contextPath+"/favicon.ico", HandleFavicon)

	r.HandleFunc(contextPath+"/log/{logger}/{timestampFrom}/{timestampTo}", FindHandleLogsBetweenTimestamps)
	r.HandleFunc(contextPath+"/truncateLogFile/{loggerName}/{logMachine}", HandleTruncateLogFile)
	r.HandleFunc(contextPath+"/restartProcess/{loggerName}/{logMachine}", HandleRestartProcess)
	r.HandleFunc(contextPath+"/tailLogFile/{loggerName}/{lines}", gzipWrapper(HandleTailLogFile))
	r.HandleFunc(contextPath+"/tailFLog/{loggerName}/{logSeq}", gzipWrapper(HandleTailFLog))
	r.HandleFunc(contextPath+"/machines", HandleMachines)
	r.HandleFunc(contextPath+"/logs", HandleLogs)
	r.HandleFunc(contextPath+"/saveConfig", HandleSaveConf)
	r.HandleFunc(contextPath+"/loadConfig", HandleLoadConf)
	r.HandleFunc(contextPath+"/", gzipWrapper(HandleHome))

	http.Handle(contextPath+"/", r)

	fmt.Println("start to listen at ", httpPort)
	http.ListenAndServe(":"+httpPort, nil)
}
