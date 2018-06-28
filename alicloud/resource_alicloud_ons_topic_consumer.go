package alicloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"fmt"
	"github.com/Dreamheart/apsarastack-mq-go-sdk/service/ons"
	"github.com/hashicorp/terraform/helper/resource"
	"time"
)

func resourceAlicloudOnsTopicConsumer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliyunOnsTopicConsumerCreate,
		Read: resourceAliyunOnsTopicConsumerRead,
		Delete: resourceAliyunOnsTopicConsumerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"topic": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"consumerid": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"regionid": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliyunOnsTopicConsumerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient).onsconn
	topic := d.Get("topic").(string)
	consumerId := d.Get("consumerid").(string)
	regionId := d.Get("regionid").(string)

	createTopicConsumerRequest := ons.CreateCreateConsumerRequest(regionId, topic, consumerId)
	createTopicConsumerResponse, err := client.CreateConsumer(createTopicConsumerRequest)
	if err != nil {
		return fmt.Errorf("Failed to create topic_consumer with error: %s", err)
	}
	if (createTopicConsumerResponse.IsSuccess() == false || createTopicConsumerResponse.Success == false) {
		return fmt.Errorf("Failed to create topic_consumer with error: [%s]%s", createTopicConsumerResponse.Status, createTopicConsumerResponse.Message)
	}

	d.SetId(consumerId)
	return resourceAliyunOnsTopicConsumerRead(d, meta)
}

func resourceAliyunOnsTopicConsumerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient).onsconn
	topic := d.Get("topic").(string)
	consumerId := d.Get("consumerid").(string)
	regionId := d.Get("regionid").(string)

	describeTopicConsumerRequest := ons.CreateDescribeConsumerRequest(regionId, topic, consumerId)
	describeTopicConsumerResponse, err := client.DescribeConsumer(describeTopicConsumerRequest)

	if err != nil {
		return fmt.Errorf("Failed to describe topic_consumer with error: %s", err)
	}
	if (describeTopicConsumerResponse.IsSuccess() == false || describeTopicConsumerResponse.Success == false) {
		return fmt.Errorf("Failed to describe topic_consumer with error: [%s]%s", describeTopicConsumerResponse.Status, describeTopicConsumerResponse.Message)
	}

	d.Set("topic", describeTopicConsumerResponse.Data[0].Topic)
	d.Set("id", describeTopicConsumerResponse.Data[0].Id)
	d.Set("regionid", describeTopicConsumerResponse.Data[0].RegionId)
	d.Set("regionName", describeTopicConsumerResponse.Data[0].RegionName)
	d.Set("status", describeTopicConsumerResponse.Data[0].Status )

	return nil
}

func resourceAliyunOnsTopicConsumerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient).onsconn
	topic := d.Get("topic").(string)
	consumerId := d.Get("consumerid").(string)
	regionId := d.Get("regionid").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		deleteTopicConsumerRequest := ons.CreateDeleteConsumerRequest(regionId, topic, consumerId)
		deleteTopicConsumerResponse, err := client.DeleteConsumer(deleteTopicConsumerRequest)
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Delete topic_consumer timeout and got an error: %s", err))
		}
		if (deleteTopicConsumerResponse.IsSuccess() == false || deleteTopicConsumerResponse.Success == false) {
			return resource.RetryableError(fmt.Errorf("Delete topic_consumer timeout and got an error: [%s]%s", deleteTopicConsumerResponse.Status, deleteTopicConsumerResponse.Message))
		}
		return nil
	})
}