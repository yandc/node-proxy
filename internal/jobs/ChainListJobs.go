package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"gitlab.bixin.com/mili/node-proxy/internal/data/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

type ChainListGetNodeUrlJob struct {
	execTime string
	db       *gorm.DB
	rdb      *redis.Client
	log      *log.Helper
}

func NewChainListGetNodeUrlJob(db *gorm.DB, rdb *redis.Client, logger log.Logger) *ChainListGetNodeUrlJob {
	job := &ChainListGetNodeUrlJob{
		db:       db,
		rdb:      rdb,
		execTime: "0 0/1 * * *",
		log:      log.NewHelper(logger),
	}

	go job.Run() //部署之后立马执行一次
	return job
}

func (j *ChainListGetNodeUrlJob) Run() {

	//分布式锁
	muxKey := fmt.Sprintf("mux:%s", "ChainListGetNodeUrlJob")
	if ok, _ := j.rdb.SetNX(muxKey, true, time.Second*3).Result(); !ok {
		return
	}

	jobName := "ChainListGetNodeUrlJob"
	j.log.Infof("任务执行开始：%s", jobName)

	client := http.Client{}
	resp, err := client.Get("https://chainlist.org/zh?testnets=true")
	if err != nil {
		j.log.Errorf("任务执行失败：%s,err：%s", jobName, err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		j.log.Errorf("任务执行失败：%s,request status：%s", jobName, resp.StatusCode)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		j.log.Errorf("任务执行失败：%s,err：%s", jobName, err.Error())
		return
	}

	chainListDataStr := doc.Find("#__NEXT_DATA__").Text()
	chainListData := &ChainListData{}
	if err = json.Unmarshal([]byte(chainListDataStr), chainListData); err != nil {
		j.log.Errorf("任务执行失败：%s,err：%s", jobName, err.Error())
		return
	}

	for _, chain := range chainListData.Props.PageProps.Chains {

		blockChain := &models.BlockChain{
			Name:           chain.Name,
			Chain:          chain.Chain,
			Title:          chain.Title,
			ChainId:        chain.ChainId.String(),
			CurrencyName:   chain.NativeCurrency.Name,
			CurrencySymbol: chain.NativeCurrency.Symbol,
			Decimals:       uint8(chain.NativeCurrency.Decimals),
			ChainSlug:      chain.ChainSlug,
			IsTest:         false,
		}

		if strings.Contains(strings.ToLower(chain.Name), "test") ||
			strings.Contains(strings.ToLower(chain.Name), "devnet") ||
			strings.Contains(strings.ToLower(chain.Title), "test") ||
			strings.Contains(strings.ToLower(chain.Title), "devnet") {
			blockChain.IsTest = true
		}

		if chain.ChainSlug != "" {
			blockChain.Logo = fmt.Sprintf("https://icons.llamao.fi/icons/chains/rsz_%s.jpg", chain.ChainSlug)
		} else {
			blockChain.Logo = "https://chainlist.org/unknown-logo.png"
		}

		if chain.Explorers != nil && len(chain.Explorers) > 0 {
			blockChain.Explorer = chain.Explorers[0].Url
		}

		if err := j.db.Clauses(clause.OnConflict{DoNothing: true}).Create(blockChain).Error; err != nil {
			j.log.Errorf("任务执行异常：%s,err：%s", jobName, err.Error())
		}
		//
		var nodeUrls []*models.ChainNodeUrl
		for _, rpc := range chain.Rpc {
			nodeUrls = append(nodeUrls, &models.ChainNodeUrl{
				ChainId: chain.ChainId.String(),
				Url:     rpc.Url,
				Privacy: rpc.Tracking,
				Status:  models.ChainNodeUrlStatusUnAvailable, //默认不可用
				Source:  models.ChainNodeUrlSourcePublic,
			})
		}

		if len(nodeUrls) != 0 {
			if err := j.db.Clauses(clause.OnConflict{DoNothing: true}).Create(nodeUrls).Error; err != nil {
				j.log.Errorf("任务执行异常：%s,err：%s", jobName, err.Error())
			}
		}

	}

	j.log.Infof("任务执行完成：%s", jobName)
}

type ChainListData struct {
	Props struct {
		PageProps struct {
			Chains []struct {
				Name  string `json:"name"`
				Chain string `json:"chain"`
				Icon  string `json:"icon,omitempty"`
				Rpc   []struct {
					Url             string `json:"url"`
					Tracking        string `json:"tracking,omitempty"`
					TrackingDetails string `json:"trackingDetails,omitempty"`
					IsOpenSource    bool   `json:"isOpenSource,omitempty"`
				} `json:"rpc"`
				Features []struct {
					Name string `json:"name"`
				} `json:"features,omitempty"`
				Faucets        []string `json:"faucets"`
				NativeCurrency struct {
					Name     string `json:"name"`
					Symbol   string `json:"symbol"`
					Decimals int    `json:"decimals"`
				} `json:"nativeCurrency"`
				InfoURL   string  `json:"infoURL"`
				ShortName string  `json:"shortName"`
				ChainId   big.Int `json:"chainId"`
				NetworkId big.Int `json:"networkId"`
				Slip44    int64   `json:"slip44,omitempty"`
				Ens       struct {
					Registry string `json:"registry"`
				} `json:"ens,omitempty"`
				Explorers []struct {
					Name     string `json:"name"`
					Url      string `json:"url"`
					Standard string `json:"standard"`
					Icon     string `json:"icon,omitempty"`
				} `json:"explorers,omitempty"`
				Tvl       float64 `json:"tvl,omitempty"`
				ChainSlug string  `json:"chainSlug,omitempty"`
				Parent    struct {
					Type    string `json:"type"`
					Chain   string `json:"chain"`
					Bridges []struct {
						Url string `json:"url"`
					} `json:"bridges,omitempty"`
				} `json:"parent,omitempty"`
				Title    string   `json:"title,omitempty"`
				RedFlags []string `json:"redFlags,omitempty"`
				Status   string   `json:"status,omitempty"`
			} `json:"chains"`
		} `json:"pageProps"`
		NSSG bool `json:"__N_SSG"`
	} `json:"props"`
	Page  string `json:"page"`
	Query struct {
	} `json:"query"`
	BuildId      string        `json:"buildId"`
	IsFallback   bool          `json:"isFallback"`
	Gsp          bool          `json:"gsp"`
	ScriptLoader []interface{} `json:"scriptLoader"`
}

type NodeUrlHeightJob struct {
	execTime string
	db       *gorm.DB
	rdb      *redis.Client
	log      *log.Helper
	client   *http.Client
}

func NewNodeUrlHeightJob(db *gorm.DB, rdb *redis.Client, logger log.Logger) *NodeUrlHeightJob {
	return &NodeUrlHeightJob{
		db:       db,
		rdb:      rdb,
		execTime: "0/10 * * * *",
		log:      log.NewHelper(logger),
		client:   &http.Client{Timeout: time.Second * 5}, //超时时间5秒，超过5秒则认为节点不可用
	}
}

func (j *NodeUrlHeightJob) Run() {

	//分布式锁
	muxKey := fmt.Sprintf("mux:%s", "NodeUrlHeightJob")
	if ok, _ := j.rdb.SetNX(muxKey, true, time.Second*3).Result(); !ok {
		return
	}

	jobName := "NodeUrlHeightJob"
	j.log.Infof("任务执行开始：%s", jobName)

	var offset int
	limit := 50
	var nodeUrls []*models.ChainNodeUrl

	//分批查询节点
	for true {
		if err := j.db.Offset(offset).Limit(limit).Find(&nodeUrls).Error; err != nil {
			j.log.Warnf(err.Error())
			break
		}

		if len(nodeUrls) == 0 {
			break
		}

		offset += limit

		group := sync.WaitGroup{}
		for _, nodeUrl := range nodeUrls {
			group.Add(1)
			go j.checkAndUpdate(nodeUrl, &group)
		}
		group.Wait()
	}

	j.log.Infof("任务执行完成：%s", jobName)
}

type GetByBlockNumberResp struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  struct {
		Number string `json:"number"`
	} `json:"result"`
}

type ChainIdResp struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  string `json:"result"`
}

func (j *NodeUrlHeightJob) checkAndUpdate(nodeUrl *models.ChainNodeUrl, group *sync.WaitGroup) {
	defer group.Done()

	//如果不是https节点，标记节点不可用
	if !strings.HasPrefix(nodeUrl.Url, "https://") {
		nodeUrl.Status = 2
		if err := j.db.Updates(nodeUrl).Error; err != nil {
			j.log.Warnf(err.Error())
		}
		return
	}

	//检查chainId
	err := checkChainId(j.client, nodeUrl)
	if err != nil {
		if err := j.db.Updates(nodeUrl).Error; err != nil {
			j.log.Warnf(err.Error())
		}
		return
	}

	//检查块高
	err = checkHeight(j.client, nodeUrl)
	if err != nil {
		if err := j.db.Updates(nodeUrl).Error; err != nil {
			j.log.Warnf(err.Error())
		}
		return
	}

	//保存或更新
	if err := j.db.Updates(nodeUrl).Error; err != nil {
		j.log.Warnf(err.Error())
	}
}

func checkChainId(client *http.Client, nodeUrl *models.ChainNodeUrl) error {

	//检查chainId
	resp, err := client.Post(nodeUrl.Url, "application/json", strings.NewReader(`{"jsonrpc": "2.0","method": "eth_chainId","params": [],"id": 1}`))

	//请求失败，标记节点不可用
	if err != nil {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
	}

	//请求失败，标记节点不可用
	if resp.StatusCode != 200 {
		nodeUrl.Status = 2
		return errors.New("check chainId error)")
	}

	respBytes, err := io.ReadAll(resp.Body)
	//读取响应数据失败，标记节点不可用
	if err != nil {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
	}

	chainIdResp := &ChainIdResp{}
	err = json.Unmarshal(respBytes, chainIdResp)
	//解析响应数据失败，标记节点不可用
	if err != nil {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
	}

	chainId, err := hexutil.DecodeBig(chainIdResp.Result)
	if err != nil {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
	}

	if chainId.String() != nodeUrl.ChainId {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("chain id not match,expect:%d,get:%d", chainId.Int64(), nodeUrl.ChainId))
	}

	return nil
}

func checkHeight(client *http.Client, nodeUrl *models.ChainNodeUrl) error {
	//获取节点高度
	start := time.Now().UnixMilli()
	resp, err := client.Post(nodeUrl.Url, "application/json", strings.NewReader(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["latest",false],"id":1}`))
	end := time.Now().UnixMilli()

	//请求失败，标记节点不可用
	if err != nil {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("check height error , err : %s", err.Error()))
	}

	//请求失败，标记节点不可用
	if resp.StatusCode != 200 {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("check height error"))
	}

	respBytes, err := io.ReadAll(resp.Body)
	//读取响应数据失败，标记节点不可用
	if err != nil {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("check height error , err : %s", err.Error()))
	}

	getByBlockNumberResp := &GetByBlockNumberResp{}
	err = json.Unmarshal(respBytes, getByBlockNumberResp)
	//解析响应数据失败，标记节点不可用
	if err != nil {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("check height error , err : %s", err.Error()))
	}

	height, err := hexutil.DecodeUint64(getByBlockNumberResp.Result.Number)
	//解析高度失败，标记节点不可用
	if err != nil {
		nodeUrl.Status = 2
		return errors.New(fmt.Sprintf("check height error , err : %s", err.Error()))
	}

	nodeUrl.Height = height
	nodeUrl.Status = 1
	nodeUrl.Latency = uint64(end - start)

	return nil
}
