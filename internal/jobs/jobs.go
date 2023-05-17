package jobs

import (
	"fmt"
	"github.com/google/wire"
	"github.com/robfig/cron/v3"
)

// ProviderSet is jobs providers.
var ProviderSet = wire.NewSet(NewChainListGetNodeUrlJob, NewNodeUrlHeightJob, NewJobManager)

func NewJobManager(job1 *ChainListGetNodeUrlJob, job2 *NodeUrlHeightJob) *cron.Cron {
	jobManager := cron.New()
	_, err := jobManager.AddJob(job1.execTime, job1)
	_, err = jobManager.AddJob(job2.execTime, job2)
	if err != nil {
		fmt.Println(err.Error())
		panic(err.Error())
	}

	return jobManager
}
