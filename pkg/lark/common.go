package lark

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/pkg/utils"
	"strconv"
	"strings"
	"time"
)

const (
	DEFAULT_MSG_LEVEL      = "FATAL"
	DEFAULT_ALARM_LEVEL    = 1
	DEFAULT_ALARM_CYCLE    = true
	DEFAULT_ALARM_INTERVAL = 3600
	LARK_MSG_KEY           = "lark:msg:%s"

	PAGE_SIZE = 200

	REDIS_NIL_KEY             = "redis: nil"
	BLOCK_HEIGHT_KEY          = "block:height:"
	BLOCK_NODE_HEIGHT_KEY     = "block:height:node:"
	BLOCK_HASH_KEY            = "block:hash:"
	USER_ADDRESS_KEY          = "usercore:"
	BLOCK_HASH_EXPIRATION_KEY = 2 * time.Hour
	TABLE_POSTFIX             = "_transaction_record"
	ADDRESS_DONE_NONCE        = "done:"    // + chainanme:address value-->nonce
	ADDRESS_PENDING_NONCE     = "pending:" // + chainname:address:nonce --->过期时间六个小时
)

const (
	STC      = "STC"
	BTC      = "BTC"
	EVM      = "EVM"
	TVM      = "TVM"
	APTOS    = "APTOS"
	SUI      = "SUI"
	SOLANA   = "SOL"
	NERVOS   = "CKB"
	CASPER   = "CSPR"
	COSMOS   = "COSMOS"
	POLKADOT = "POLKADOT"
)

const (
	WEB     = "web"
	ANDROID = "android"
	IOS     = "ios"
)

const (
	SUCCESS          = "success"
	FAIL             = "fail"
	PENDING          = "pending"          //-- 中间状态
	NO_STATUS        = "no_status"        //-- 中间状态
	DROPPED_REPLACED = "dropped_replaced" //--被丢弃或被置换 -- fail
	DROPPED          = "dropped"          //--被丢弃 -- fail
)

const (
	CANCEL      = "cancel"   //中心化操作 --- value CANCEL --success
	SPEED_UP    = "speed_up" //success
	GAS_FEE_LOW = "gasFeeLow"
	NONCE_QUEUE = "nonceQueue"
	NONCE_BREAK = "nonceBreak"
)

const (
	NATIVE                  = "native"
	TRANSFERNFT             = "transferNFT"
	APPROVENFT              = "approveNFT"
	CONTRACT                = "contract"
	EVENTLOG                = "eventLog"
	CREATEACCOUNT           = "createAccount"
	REGISTERTOKEN           = "registerToken"
	DIRECTTRANSFERNFTSWITCH = "directTransferNFTSwitch"
	OTHER                   = "other"
)

const (
	APPROVE               = "approve"
	SETAPPROVALFORALL     = "setApprovalForAll"
	TRANSFER              = "transfer"
	TRANSFERFROM          = "transferFrom"
	SAFETRANSFERFROM      = "safeTransferFrom"
	SAFEBATCHTRANSFERFROM = "safeBatchTransferFrom"
)

// 币种
const (
	CNY = "CNY" //人民币
	USD = "USD" //美元
)

const TOKEN_INFO_QUEUE_TOPIC = "token:info:queue:topic"
const TOKEN_INFO_QUEUE_PARTITION = "partition1"

// token类型
const (
	ERC20    = "ERC20"
	ERC721   = "ERC721"
	ERC1155  = "ERC1155"
	APTOSNFT = "AptosNFT"
)

var rocketMsgLevels = map[string]int{
	"DEBUG":   0,
	"NOTICE":  1,
	"WARNING": 2,
	"FATAL":   3,
}

type BroadcastRequest struct {
	SessionId string `json:"sessionId"`
	Address   string `json:"address"`
	UserName  string `json:"userName"`
	TxInput   string `json:"txInput"`
	ChainName string `json:"chainName"`
	ErrMsg    string `json:"errMsg"`
}

type BroadcastResponse struct {
	Ok      bool   `json:"ok,omitempty"`
	Message string `json:"message,omitempty"`
}

type DataDictionary struct {
	Ok                     bool     `json:"ok,omitempty"`
	ServiceTransactionType []string `json:"serviceTransactionType,omitempty"`
	ServiceStatus          []string `json:"serviceStatus,omitempty"`
}

type UserAssetUpdateRequest struct {
	ChainName string `json:"chainName"`
	Address   string `json:"address"`

	Assets []UserAsset `json:"assets"`
}

type UserAsset struct {
	TokenAddress string `json:"tokenAddress"`
	Balance      string `json:"balance"`
	Decimals     int    `json:"decimals"`
}

// alarmOptions 报警相关的一些可自定义参数
type alarmOptions struct {
	channel       string
	level         string
	alarmLevel    int
	alarmCycle    bool
	alarmInterval int
	chainName     string // to reflect chainType as channel
}

var DefaultAlarmOptions = alarmOptions{
	level:         DEFAULT_MSG_LEVEL,
	alarmLevel:    DEFAULT_ALARM_LEVEL,
	alarmCycle:    DEFAULT_ALARM_CYCLE,
	alarmInterval: DEFAULT_ALARM_INTERVAL,
}

func (o *alarmOptions) getChannel() string {
	if o.channel != "" {
		return o.channel
	}
	return o.channel
}

// AlarmOption 报警自定义参数接口
type AlarmOption interface {
	apply(*alarmOptions)
}

type tempFunc func(*alarmOptions)

type funcAlarmOption struct {
	f tempFunc
}

func (fdo *funcAlarmOption) apply(e *alarmOptions) {
	fdo.f(e)
}

func newfuncAlarmOption(f tempFunc) *funcAlarmOption {
	return &funcAlarmOption{f: f}
}

func WithMsgLevel(level string) AlarmOption {
	return newfuncAlarmOption(func(e *alarmOptions) {
		e.level = level
	})
}

func WithAlarmLevel(alarmLevel int) AlarmOption {
	return newfuncAlarmOption(func(e *alarmOptions) {
		e.alarmLevel = alarmLevel
	})
}

func WithAlarmCycle(alarmCycle bool) AlarmOption {
	return newfuncAlarmOption(func(e *alarmOptions) {
		e.alarmCycle = alarmCycle
	})
}

func WithAlarmInterval(alarmInterval int) AlarmOption {
	return newfuncAlarmOption(func(e *alarmOptions) {
		e.alarmInterval = alarmInterval
	})
}

//func WithAlarmAtList(uids ...string) AlarmOption {
//	return newfuncAlarmOption(func(e *alarmOptions) {
//		e.alarmAtUids = append(e.alarmAtUids, uids...)
//	})
//}

func WithAlarmChannel(channel string) AlarmOption {
	return newfuncAlarmOption(func(e *alarmOptions) {
		e.channel = channel
	})
}

func WithAlarmChainName(chainName string) AlarmOption {
	return newfuncAlarmOption(func(e *alarmOptions) {
		e.chainName = chainName
	})
}

func GetAlarmTimestamp(key string) (int64, error) {
	val, err := utils.GetRedisClient().Get(key).Result()
	if err != nil {
		return int64(0), err
	}
	timestamp, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		log.Error("GetAlarmTimestamp error: %s", err)
		return int64(0), err
	}

	return timestamp, nil
}

//func UserAddressSwitchRetryAlert(chainName, address string) (bool, string, error) {
//	enable, uid, err := UserAddressSwitch(address)
//	if err != nil {
//		// redis出错 接入lark报警
//		alarmMsg := fmt.Sprintf("请注意：%s链查询redis中用户地址失败，address:%s", chainName, address)
//		alarmOpts := WithMsgLevel("FATAL")
//		alarmOpts = WithAlarmChainName(chainName)
//		LarkClient.NotifyLark(alarmMsg, nil, nil, alarmOpts)
//	}
//	return enable, uid, err
//}

//func UserAddressSwitch(address string) (bool, string, error) {
//	if address == "" {
//		return false, "", nil
//	}
//	enable := true
//	uid, err := utils.GetRedisClient().Get(USER_ADDRESS_KEY + address).Result()
//	for i := 0; i < 3 && err != nil && err != redis.Nil; i++ {
//		time.Sleep(time.Duration(i*1) * time.Second)
//		uid, err = utils.GetRedisClient().Get(USER_ADDRESS_KEY + address).Result()
//	}
//	if !AppConfig.ScanAll {
//		if err != nil && err != redis.Nil {
//			return false, "", err
//		}
//		if uid == "" {
//			enable = false
//		}
//	}
//	return enable, uid, nil
//}

// MD5 对字符串做 MD5
func MD5(v string) string {
	d := []byte(v)
	m := md5.New()
	m.Write(d)
	return hex.EncodeToString(m.Sum(nil))
}

func BjNow() string {
	var cstSh, _ = time.LoadLocation("Asia/Shanghai")
	now := time.Now().In(cstSh).Format("2006-01-02 15:04:05")

	return now
}

func GetTableName(chainName string) string {
	return strings.ToLower(chainName) + TABLE_POSTFIX
}

//func GetDecimalsSymbol(chainName, parseData string) (int32, string, error) {
//	parseDataJson := make(map[string]interface{})
//	err := json.Unmarshal([]byte(parseData), &parseDataJson)
//	if err != nil {
//		return 0, "", err
//	}
//	tokenInfoMap := parseDataJson["token"]
//	if tokenInfoMap == nil {
//		return 0, "", errors.New("token is null")
//	}
//	tokenInfo := tokenInfoMap.(map[string]interface{})
//	var address string
//	if tokenInfo["address"] != nil {
//		address = tokenInfo["address"].(string)
//	}
//	var decimals int32
//	if tokenInfo["decimals"] != nil {
//		if decimalsInt32, ok := tokenInfo["decimals"].(int32); ok {
//			decimals = decimalsInt32
//		} else if decimalsFloat, ok := tokenInfo["decimals"].(float64); ok {
//			decimals = int32(decimalsFloat)
//		} else {
//			decimalsInt, err := strconv.Atoi(fmt.Sprintf("%v", tokenInfo["decimals"]))
//			if err != nil {
//				return 0, "", nil
//			}
//			decimals = int32(decimalsInt)
//		}
//	}
//	var symbol string
//	if tokenInfo["symbol"] != nil {
//		symbol = tokenInfo["symbol"].(string)
//	}
//	if address == "" {
//		if platInfo, ok := PlatInfoMap[chainName]; ok {
//			decimals = platInfo.Decimal
//			symbol = platInfo.NativeCurrency
//		} else {
//			return 0, "", errors.New("chain " + chainName + " is not support")
//		}
//	}
//	return decimals, symbol, nil
//}

//func ParseGetTokenInfo(chainName, parseData string) (*types.TokenInfo, error) {
//	tokenInfo, err := ParseTokenInfo(parseData)
//	if err == nil && tokenInfo.Address == "" {
//		if platInfo, ok := PlatInfoMap[chainName]; ok {
//			tokenInfo.Decimals = int64(platInfo.Decimal)
//			tokenInfo.Symbol = platInfo.NativeCurrency
//		} else {
//			return nil, errors.New("chain " + chainName + " is not support")
//		}
//	}
//	return tokenInfo, err
//}

//func ParseTokenInfo(parseData string) (*types.TokenInfo, error) {
//	parseDataJson := make(map[string]interface{})
//	err := json.Unmarshal([]byte(parseData), &parseDataJson)
//	if err != nil {
//		return nil, err
//	}
//	tokenInfoMap := parseDataJson["token"]
//	if tokenInfoMap == nil {
//		return nil, errors.New("token is null")
//	}
//	tokenInfoByte, err := json.Marshal(tokenInfoMap)
//	if err != nil {
//		return nil, err
//	}
//	tokenInfo := &types.TokenInfo{}
//	err = json.Unmarshal(tokenInfoByte, tokenInfo)
//	if err != nil {
//		return nil, err
//	}
//	return tokenInfo, nil
//}

//func NotifyBroadcastTxFailed(ctx context.Context, sessionID string, errMsg string) {
//	info, err := data.UserSendRawHistoryRepoInst.GetLatestOneBySessionId(ctx, sessionID)
//	var msg string
//	if err != nil {
//		log.Error(
//			"SEND ALARM OF BROADCATING TX FAILED",
//			zap.String("sessionId", sessionID),
//			zap.String("errMsg", errMsg),
//			zap.Error(err),
//		)
//		msg = fmt.Sprintf(
//			"交易广播失败。\nsessionId：%s\n钱包地址：%s\n用户名：%s\n错误消息：%s",
//			sessionID,
//			"Unknown",
//			"Unknown",
//			errMsg,
//		)
//	} else {
//		msg = fmt.Sprintf(
//			"%s 链交易广播失败。\nsessionId：%s\n钱包地址：%s\n用户名：%s\n错误消息：%s\ntxInput: %s",
//			info.ChainName,
//			sessionID,
//			info.Address,
//			info.UserName,
//			errMsg,
//			info.TxInput,
//		)
//	}
//	alarmOpts := WithMsgLevel("FATAL")
//	alarmOpts = WithAlarmChannel("txinput")
//	LarkClient.NotifyLark(msg, nil, nil, alarmOpts)
//}
