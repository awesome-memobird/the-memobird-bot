package memobird

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tevino/log"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// Default API endpoints, according to http://open.memobird.cn/upload/webapi.pdf
const (
	APIPrefix        = "https://open.memobird.cn/home"
	apiFnPrintPaper  = "printpaper"
	apiFnSetUserBind = "setuserbind"
)

var fnMethods = map[string]string{
	apiFnPrintPaper:  http.MethodPost,
	apiFnSetUserBind: http.MethodPost,
}

// App is an memobird application.
type App struct {
	*AppConfig
	cli *http.Client
}

// AppConfig contains all configurations of an App.
type AppConfig struct {
	AccessKey string
	Timeout   time.Duration

	// CustomizedAPIPrefix is an optional configuration where an alternative APIPrefix should be used.
	CustomizedAPIPrefix string
}

// APIPrefix returns the APIPrefix.
func (a AppConfig) APIPrefix() string {
	if a.CustomizedAPIPrefix != "" {
		return a.CustomizedAPIPrefix
	}
	return APIPrefix
}

// NewApp creates an App with given accessKey.
func NewApp(config *AppConfig) *App {
	return &App{
		AppConfig: config,
		cli:       &http.Client{Timeout: config.Timeout},
	}
}

func (a *App) getAPIURL(fn string) string {
	return a.APIPrefix() + "/" + fn
}

var TZShanghai *time.Location

const MemobirdTimeZone = "Asia/Shanghai"

func init() {
	var err error
	TZShanghai, err = time.LoadLocation(MemobirdTimeZone)
	if err != nil {
		panic("can not load timezone")
	}
}

func (a *App) do(fn string, formMap map[string]string) (*http.Response, error) {
	method := fnMethods[fn]
	form := url.Values{}
	for k, v := range formMap {
		form.Add(k, v)
	}
	form.Add("ak", a.AccessKey)

	ts := time.Now().In(TZShanghai).Format("2006-01-02 15:04:05")
	form.Add("timestamp", ts)

	req, err := http.NewRequest(method, a.getAPIURL(fn), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	log.Debugf("%s %s %s", req.Method, req.URL.String(), form)

	return a.cli.Do(req)
}

func (a *App) doWithReply(fn string, formMap map[string]string, reply interface{}) error {
	resp, err := a.do(fn, formMap)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	log.Debugf("API response: %s", buf)
	err = json.Unmarshal(buf, reply)
	if err != nil {
		return fmt.Errorf("marshalling response JSON: %w", err)
	}
	return nil
}

type printContentReply struct {
	ReturnCode     int    `json:"showapi_res_code"` // 1: success, others: failed
	ReturnErr      string `json:"showapi_res_error"`
	PrintContentID int64  `json:"printcontentID"`
	Result         int    `json:"result"`    // 1: printed, others: not printed
	SmartGUID      string `json:"smartGuid"` // print device ID
}

// UTF8ToGBK converts UTF8 text to GBK.
func UTF8ToGBK(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewEncoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

type bindUserReply struct {
	ReturnCode int    `json:"showapi_res_code"` // 1: success, others: failed
	ReturnErr  string `json:"showapi_res_error"`
	UserID     uint64 `json:"showapi_userid"`
}

// BindDevice binds given device for further use.
func (a *App) BindDevice(deviceID string) (*BindResult, error) {
	reply := new(bindUserReply)
	if err := a.doWithReply(apiFnSetUserBind, map[string]string{
		// I don't know what this parameter is for, using user ID turns out error: showapi_res_error: "咕咕机未激活或者未绑定"
		// "useridentifying": userID,
		"memobirdID": deviceID,
	}, reply); err != nil {
		return nil, err
	}

	result := &BindResult{
		UserID: reply.UserID,
	}
	result.IsSuccess = reply.ReturnCode == 1
	if reply.ReturnErr != "" {
		result.Err = errors.New(reply.ReturnErr)
	}
	return result, nil
}

// PrintText prints txt to membird of given deviceID.
func (a *App) PrintText(txt string, deviceID string) (*PrintResult, error) {
	gbkTxt, err := UTF8ToGBK([]byte(txt))
	if err != nil {
		return nil, fmt.Errorf("encoding text: %w", err)
	}
	resp, err := a.do(apiFnPrintPaper, map[string]string{
		"printcontent": "T:" + base64.StdEncoding.EncodeToString(gbkTxt),
		"memobirdID":   deviceID,
	})
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	reply := new(printContentReply)
	buf, err := ioutil.ReadAll(resp.Body)
	log.Debugf("API response: %s", buf)
	err = json.Unmarshal(buf, reply)
	if err != nil {
		return nil, fmt.Errorf("marshalling response JSON: %w", err)
	}

	result := &PrintResult{
		IsPrinted: reply.Result == 1,
		ContentID: reply.PrintContentID,
		DeviceID:  reply.SmartGUID,
	}
	result.IsSuccess = reply.ReturnCode == 1
	if reply.ReturnErr != "" {
		result.Err = errors.New(reply.ReturnErr)
	}
	return result, nil
}
