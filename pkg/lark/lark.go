package lark

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/internal/conf"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

// Lark .
type Lark struct {
	conf *conf.Lark
	lock utils.Syncronized
}

var LarkClient *Lark

// NewLark new a lark.
func NewLark(c *conf.Lark) *Lark {
	LarkClient = &Lark{
		conf: c,
		lock: utils.NewSyncronized(c.LockNum),
	}

	return LarkClient
}

type LarkResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type Content struct {
	Tag      string `json:"tag"`
	UserID   string `json:"user_id,omitempty"`
	UserName string `json:"user_name,omitempty"`
	Text     string `json:"text,omitempty"`
}

func (lark *Lark) NotifyLark(msg string, opts ...AlarmOption) {
	lark.lock.Lock(msg)
	defer func() {
		lark.lock.Unlock(msg)
		if err := recover(); err != nil {
			log.Error("NotifyLark panic:", zap.Any("", err))
		}
	}()
	// 默认报警参数
	alarmOpts := DefaultAlarmOptions
	// 根据用户自定义信息更新报警参数
	for _, opt := range opts {
		opt.apply(&alarmOpts)
	}
	// 小于默认报警 Level，不发 lark 信息
	msgLevel, ok := rocketMsgLevels[alarmOpts.level]
	if !ok {
		msgLevel = DEFAULT_ALARM_LEVEL
	}
	if msgLevel < alarmOpts.alarmLevel {
		return
	}

	key := fmt.Sprintf(LARK_MSG_KEY, MD5(msg))
	msgTimestamp, _ := GetAlarmTimestamp(key)
	// 如果有缓存，说明之前发过相同的内容
	if int64(msgTimestamp) > 0 {
		// 如果设置 alarmCycle = False，不再重新发送
		if !alarmOpts.alarmCycle {
			return
		}
		// 如果 alarmCycle = true，并且上次发送时间距离现在小于 alarmInterval，不再重新发送
		if time.Now().Unix()-int64(msgTimestamp) < int64(alarmOpts.alarmInterval) {
			return
		}
	}

	//c := make([]Content, 0, 6)
	//c = append(c, lark.handleAtList(alarmOpts)...)
	c := lark.handleAtList(alarmOpts)
	//if lark.conf.GetLarkAtList() != "" {
	//	atList := strings.Split(lark.conf.GetLarkAtList(), ",")
	//	for _, v := range atList {
	//		c = append(c, Content{Tag: "at", UserID: v, UserName: v})
	//	}
	//} else {
	//	c = append(c, Content{Tag: "at", UserID: "all", UserName: "所有人"})
	//}
	c = append(c, Content{Tag: "text", Text: msg + "\n"})

	c = append(c, Content{Tag: "text", Text: "开始时间:\n"}, Content{Tag: "text", Text: BjNow()})
	t := time.Now().Unix()
	sign, _ := GenSign(lark.conf.LarkSecret, t)
	content := make([][]Content, 1)
	content[0] = c
	b, _ := json.Marshal(content)
	data := `{
   "msg_type": "post",
	"timestamp":"` + fmt.Sprintf("%v", t) + `",
	"sign": "` + sign + `",
   "content": {
       "post": {
           "zh_cn": {
               "title": "` + lark.conf.LarkAlarmTitle + `",
               "content":
                   ` + string(b) + `
           	}
       	}
   	}
	}`
	req, err := http.NewRequest(http.MethodPost, lark.conf.LarkHost, strings.NewReader(data))
	if err != nil {
		log.Error("lark http.NewRequest error:", zap.Error(err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	var client = http.DefaultClient
	response, err := client.Do(req)
	if err != nil {
		log.Error("lark request error: ", zap.Error(err))
		return
	}

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)
	var resp LarkResponse
	json.Unmarshal(body, &resp)
	if resp.Code > 0 {
		log.Error("send lark error:", zap.Error(errors.New(resp.Msg)))
	}
	// 更新缓存时间戳，失效时间1小时
	if response.StatusCode == 200 {
		utils.GetRedisClient().Set(key, time.Now().Unix(), 60*60*time.Second)
	}
}

func GenSign(secret string, timestamp int64) (string, error) {
	//timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}

func (lark *Lark) handleAtList(alarmOpts alarmOptions) []Content {
	c := make([]Content, 0, 6)
	channel := alarmOpts.getChannel()
	if channel != "" {
		subscriptions := lark.conf.GetLarkSubscriptions()
		if users, ok := subscriptions[channel]; ok {
			for _, user := range users.Uids {
				if v, ok := lark.conf.LarkUids[user]; ok {
					c = append(c, Content{Tag: "at", UserID: v, UserName: v})
				} else {
					log.Warn("UNKOWN LARK USER", zap.String("user", user))
				}
			}
		} else {
			log.Warn("UNKNOWN LARK CHANNEL", zap.String("channel", channel))
		}
		return c
	}

	if lark.conf.GetLarkAtList() != "" {
		atList := strings.Split(lark.conf.GetLarkAtList(), ",")
		for _, v := range atList {
			c = append(c, Content{Tag: "at", UserID: v, UserName: v})
		}
	} else {
		c = append(c, Content{Tag: "at", UserID: "all", UserName: "所有人"})
	}
	return c
}
