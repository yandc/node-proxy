package jobs

import (
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-redis/redis"
	"gitlab.bixin.com/mili/node-proxy/pkg/chainlist"
	"gorm.io/gorm"
	"net/http"
	"sync"
	"time"
)

type ChainListGetNodeUrlJob struct {
	execTime string
	db       *gorm.DB
	rdb      *redis.Client
	log      *log.Helper
	client   *http.Client
}

func NewChainListGetNodeUrlJob(db *gorm.DB, rdb *redis.Client, logger log.Logger) *ChainListGetNodeUrlJob {
	job := &ChainListGetNodeUrlJob{
		db:       db,
		rdb:      rdb,
		execTime: "0 0/1 * * *",
		log:      log.NewHelper(logger),
		client:   &http.Client{Timeout: time.Second * 10},
	}

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

	jobGroup := sync.WaitGroup{}
	jobGroup.Add(2)
	go func() {
		defer jobGroup.Done()
		chainlist.GetNativeCosmosChainList(j.log, j.db)
		chainlist.GetTestCosmosChainList(j.log, j.db)
	}()
	go func() {
		defer jobGroup.Done()
		chainlist.GetEVMChainList(j.log, j.db, j.client)
	}()
	jobGroup.Wait()
	j.log.Infof("任务执行完成：%s", jobName)
}

//type ChainListData struct {
//	Props struct {
//		PageProps struct {
//			Chains []struct {
//				Name  string `json:"name"`
//				Chain string `json:"chain"`
//				Icon  string `json:"icon,omitempty"`
//				Rpc   []struct {
//					Url             string `json:"url"`
//					Tracking        string `json:"tracking,omitempty"`
//					TrackingDetails string `json:"trackingDetails,omitempty"`
//					IsOpenSource    bool   `json:"isOpenSource,omitempty"`
//				} `json:"rpc"`
//				Features []struct {
//					Name string `json:"name"`
//				} `json:"features,omitempty"`
//				Faucets        []string `json:"faucets"`
//				NativeCurrency struct {
//					Name     string `json:"name"`
//					Symbol   string `json:"symbol"`
//					Decimals int    `json:"decimals"`
//				} `json:"nativeCurrency"`
//				InfoURL   string  `json:"infoURL"`
//				ShortName string  `json:"shortName"`
//				ChainId   big.Int `json:"chainId"`
//				NetworkId big.Int `json:"networkId"`
//				Slip44    int64   `json:"slip44,omitempty"`
//				Ens       struct {
//					Registry string `json:"registry"`
//				} `json:"ens,omitempty"`
//				Explorers []struct {
//					Name     string `json:"name"`
//					Url      string `json:"url"`
//					Standard string `json:"standard"`
//					Icon     string `json:"icon,omitempty"`
//				} `json:"explorers,omitempty"`
//				Tvl       float64 `json:"tvl,omitempty"`
//				ChainSlug string  `json:"chainSlug,omitempty"`
//				Parent    struct {
//					Type    string `json:"type"`
//					Chain   string `json:"chain"`
//					Bridges []struct {
//						Url string `json:"url"`
//					} `json:"bridges,omitempty"`
//				} `json:"parent,omitempty"`
//				Title    string   `json:"title,omitempty"`
//				RedFlags []string `json:"redFlags,omitempty"`
//				Status   string   `json:"status,omitempty"`
//			} `json:"chains"`
//		} `json:"pageProps"`
//		NSSG bool `json:"__N_SSG"`
//	} `json:"props"`
//	Page  string `json:"page"`
//	Query struct {
//	} `json:"query"`
//	BuildId      string        `json:"buildId"`
//	IsFallback   bool          `json:"isFallback"`
//	Gsp          bool          `json:"gsp"`
//	ScriptLoader []interface{} `json:"scriptLoader"`
//}
//
//type ChainIdResp struct {
//	Jsonrpc string `json:"jsonrpc"`
//	Id      int    `json:"id"`
//	Result  string `json:"result"`
//}
//
//func checkChainId(client *http.Client, nodeUrl *models.ChainNodeUrl, group *sync.WaitGroup) error {
//
//	defer group.Done()
//
//	//检查chainId
//	resp, err := client.Post(nodeUrl.Url, "application/json", strings.NewReader(`{"jsonrpc": "2.0","method": "eth_chainId","params": [],"id": 1}`))
//
//	//请求失败，标记节点不可用
//	if err != nil {
//		nodeUrl.Status = 2
//		return errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
//	}
//
//	//请求失败，标记节点不可用
//	if resp.StatusCode != 200 {
//		nodeUrl.Status = 2
//		return errors.New("check chainId error)")
//	}
//
//	respBytes, err := io.ReadAll(resp.Body)
//	//读取响应数据失败，标记节点不可用
//	if err != nil {
//		nodeUrl.Status = 2
//		return errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
//	}
//
//	chainIdResp := &ChainIdResp{}
//	err = json.Unmarshal(respBytes, chainIdResp)
//	//解析响应数据失败，标记节点不可用
//	if err != nil {
//		nodeUrl.Status = 2
//		return errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
//	}
//
//	chainId, err := hexutil.DecodeBig(chainIdResp.Result)
//	if err != nil {
//		nodeUrl.Status = 2
//		return errors.New(fmt.Sprintf("check chainId error , err : %s", err.Error()))
//	}
//
//	if chainId.String() != nodeUrl.ChainId {
//		nodeUrl.Status = 2
//		return errors.New(fmt.Sprintf("chain id not match,expect:%d,get:%d", chainId.Int64(), nodeUrl.ChainId))
//	}
//
//	return nil
//}
