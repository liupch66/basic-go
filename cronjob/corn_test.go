package cronjob

import (
	"log"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
)

type myJob struct{}

func (m myJob) Run() {
	log.Println("运行了")
}

func TestCron(t *testing.T) {
	expr := cron.New(cron.WithSeconds())
	// id, err := expr.AddJob("*/1 * * * * *", myJob{})
	// id, err := expr.AddJob("0 41 * * * *", myJob{})
	id, err := expr.AddJob("@every 1s", myJob{})
	assert.NoError(t, err)
	t.Log("id: ", id)

	newId, err := expr.AddFunc("@every 3s", func() {
		t.Log("开始长任务")
		time.Sleep(12 * time.Second)
		t.Log("结束长任务")
	})
	assert.NoError(t, err)
	t.Log("new id: ", newId)
	expr.Start()
	time.Sleep(10 * time.Second)
	ctx := expr.Stop()
	t.Log("发出停止信号")
	// wait for running jobs to complete
	<-ctx.Done()
	t.Log("彻底停止")
}
