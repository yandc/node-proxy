package jobs

import (
	"github.com/go-kratos/kratos/v2/log"
	"gitlab.bixin.com/mili/node-proxy/pkg/token-list/tokenlist"
)

type UpdateTokenListJob struct {
	execTime string
	log      *log.Helper
}

func NewUpdateTokenListJob(logger log.Logger) *UpdateTokenListJob {
	t := &UpdateTokenListJob{
		execTime: "0 0 1 1/3 *", //从1月份开始 每3月 1号 0点执行
		log:      log.NewHelper(logger),
	}
	return t
}

func (j *UpdateTokenListJob) Run() {
	jobName := "UpdateTokenListJob"
	j.log.Infof("任务执行开始：%s", jobName)
	tokenlist.UpdateTokenListByMarket()
	j.log.Infof("任务执行完成：%s", jobName)
}
