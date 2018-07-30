package alicloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/Dreamheart/apsarastack-mq-go-sdk/service/ons"
	"fmt"
	"github.com/hashicorp/terraform/helper/resource"
	"time"
)

func resourceAlicloudOnsTopic() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliyunOnsTopicCreate,
		Read: resourceAliyunOnsTopicRead,
		Delete: resourceAliyunOnsTopicDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"topic": &schema.Schema{
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

func resourceAliyunOnsTopicCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient).onsconn
	topic := d.Get("topic").(string)
	regionId := d.Get("regionid").(string)

	createTopicRequest := ons.CreateCreateTopicRequest(regionId, topic)
	createTopicResponse, err := client.CreateTopic(createTopicRequest)
	if err != nil {
		return fmt.Errorf("Failed to create topic with error: %s", err)
	}
	if (createTopicResponse.IsSuccess() == false || createTopicResponse.Success == false) {
		return fmt.Errorf("Failed to create topic with error: [%s]%s", createTopicResponse.Status, createTopicResponse.Message)
	}

	d.SetId(topic)

	if err := meta.(*AliyunClient).WaitForTopicReady(topic, regionId, DefaultLongTimeout); err != nil {
		return fmt.Errorf("WaitForTopic %s got error: %#v", Running, err)
	}

	return resourceAliyunOnsTopicRead(d, meta)
}

func resourceAliyunOnsTopicRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient).onsconn
	topic := d.Get("topic").(string)
	regionId := d.Get("regionid").(string)

	describeTopicRequest := ons.CreateDescribeTopicRequest(regionId, topic)
	describeTopicResponse, err := client.DescribeTopic(describeTopicRequest)

	if err != nil {
		return fmt.Errorf("Failed to describe topic with error: %s", err)
	}
	if (describeTopicResponse.IsSuccess() == false || describeTopicResponse.Success == false) {
		return fmt.Errorf("Failed to describe topic with error: [%s]%s", describeTopicResponse.Status, describeTopicResponse.Message)
	}

	d.Set("topic", describeTopicResponse.Data[0].Topic)
	d.Set("id", describeTopicResponse.Data[0].Id)
	d.Set("regionid", describeTopicResponse.Data[0].RegionId)
	d.Set("regionName", describeTopicResponse.Data[0].RegionName)
	d.Set("status", describeTopicResponse.Data[0].Status )

	return nil
}

func resourceAliyunOnsTopicDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient).onsconn
	topic := d.Get("topic").(string)
	regionId := d.Get("regionid").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		deleteTopicRequest := ons.CreateDeleteTopicRequest(regionId, topic)
		deleteTopicResponse, err := client.DeleteTopic(deleteTopicRequest)
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Delete topic timeout and got an error: %s", err))
		}
		if (deleteTopicResponse.IsSuccess() == false || deleteTopicResponse.Success == false) {
			return resource.RetryableError(fmt.Errorf("Delete topic timeout and got an error: [%s]%s", deleteTopicResponse.Status, deleteTopicResponse.Message))
		}
		return nil
	})
}