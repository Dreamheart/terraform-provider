package alicloud

import (
	"github.com/denverdino/aliyungo/common"
	"time"
	"github.com/Dreamheart/apsarastack-mq-go-sdk/service/ons"
)

func (client *AliyunClient) WaitForTopicReady(topic string, regionid string, timeout int) error {
	if timeout <= 0 {
		timeout = DefaultTimeout
	}
	for {
		req := ons.CreateDescribeTopicRequest(regionid, topic)
		resp, err := client.onsconn.DescribeTopic(req)
		if err != nil && !NotFoundError(err) {
			return err
		}

		if len(resp.Data) == 1 && resp.Data[0].Status == 0 {
			break
		}

		if timeout <= 0 {
			return common.GetClientErrorFromString("Timeout")
		}

		timeout = timeout - DefaultIntervalMedium
		time.Sleep(DefaultIntervalMedium * time.Second)

	}
	return nil
}
