package main

import (
	"github.com/bingoohuang/go-utils"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func StartHttpSever() {
	r := mux.NewRouter()

	handleFunc(r, "/iconfont.{extension}", HandleFont, false)
	handleFunc(r, "/favicon.png", HandleFavicon, false)

	handleFunc(r, "/truncateLogFile/{loggerName}/{logMachine}", HandleTruncateLogFile, false)
	handleFunc(r, "/restartProcess/{loggerName}/{logMachine}", HandleRestartProcess, false)
	handleFunc(r, "/locateLog/{loggerName}/{logKey}/{preLines}/{lines}", HandleLocateLog, false)
	handleFunc(r, "/locateLogResult/", HandleLocateLogResult, true)
	handleFunc(r, "/tailLogFile/{loggerName}/{lines}", HandleTailLogFile, true)
	handleFunc(r, "/tailFLog/{loggerName}/{traceMobile}/{logSeq}", HandleTailFLog, true)
	handleFunc(r, "/machines", HandleMachines, false)
	handleFunc(r, "/logs", HandleLogs, false)
	handleFunc(r, "/testQywxMsg", HandQywxMsgs, false)
	handleFunc(r, "/saveConfig", HandleSaveConf, false)
	handleFunc(r, "/loadConfig", HandleLoadConf, false)
	handleFunc(r, "/exlog/{exLogId}", HandleExLog, false)
	handleFunc(r, "/machineInfo/{machineName}", HandleMachineInfo, false)

	handleFunc(r, "/", HandleHome, false)

	http.Handle(contextPath+"/", r)

	log.Println("start to listen at ", httpPort)

	RunAlterMsgSender()
	http.ListenAndServe(":"+httpPort, nil)
}

func handleFunc(r *mux.Router, path string, f func(http.ResponseWriter, *http.Request), requiredGzip bool) {
	wrap := go_utils.DumpRequest(f)
	wrap = go_utils.MustAuth(wrap, authParam)

	if requiredGzip {
		wrap = go_utils.GzipHandlerFunc(wrap)
	}

	r.HandleFunc(contextPath+path, wrap)
}

func HandQywxMsgs(w http.ResponseWriter, r *http.Request) {
	go_utils.HeadContentTypeHtml(w)
	msg := r.FormValue("msg")
	if msg == "" {
		msg = "empty message"
	}

	err := AddAlertMsg("测试消息", msg)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}
