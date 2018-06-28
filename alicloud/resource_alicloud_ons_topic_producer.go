package alicloud

import (
	"github.com/hashicorp/terraform/helper/schema"
	"fmt"
	"github.com/Dreamheart/apsarastack-mq-go-sdk/service/ons"
	"github.com/hashicorp/terraform/helper/resource"
	"time"
)

func resourceAlicloudOnsTopicProducer() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliyunOnsTopicProducerCreate,
		Read: resourceAliyunOnsTopicProducerRead,
		Delete: resourceAliyunOnsTopicProducerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"topic": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"producerid": &schema.Schema{
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

func resourceAliyunOnsTopicProducerCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient).onsconn
	topic := d.Get("topic").(string)
	producerId := d.Get("producerid").(string)
	regionId := d.Get("regionid").(string)

	createTopicProducerRequest := ons.CreateCreateProducerRequest(regionId, topic, producerId)
	createTopicProducerResponse, err := client.CreateProducer(createTopicProducerRequest)
	if err != nil {
		return fmt.Errorf("Failed to create topic_producer with error: %s", err)
	}
	if (createTopicProducerResponse.IsSuccess() == false || createTopicProducerResponse.Success == false) {
		return fmt.Errorf("Failed to create topic_producer with error: [%s]%s", createTopicProducerResponse.Status, createTopicProducerResponse.Message)
	}

	d.SetId(producerId)
	return resourceAliyunOnsTopicProducerRead(d, meta)
}

func resourceAliyunOnsTopicProducerRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient).onsconn
	topic := d.Get("topic").(string)
	producerId := d.Get("producerid").(string)
	regionId := d.Get("regionid").(string)

	describeTopicProducerRequest := ons.CreateDescribeProducerRequest(regionId, topic, producerId)
	describeTopicProducerResponse, err := client.DescribeProducer(describeTopicProducerRequest)

	if err != nil {
		return fmt.Errorf("Failed to describe topic_producer with error: %s", err)
	}
	if (describeTopicProducerResponse.IsSuccess() == false || describeTopicProducerResponse.Success == false) {
		return fmt.Errorf("Failed to describe topic_producer with error: [%s]%s", describeTopicProducerResponse.Status, describeTopicProducerResponse.Message)
	}

	d.Set("topic", describeTopicProducerResponse.Data[0].Topic)
	d.Set("id", describeTopicProducerResponse.Data[0].Id)
	d.Set("regionid", describeTopicProducerResponse.Data[0].RegionId)
	d.Set("regionName", describeTopicProducerResponse.Data[0].RegionName)
	d.Set("status", describeTopicProducerResponse.Data[0].Status )

	return nil
}

func resourceAliyunOnsTopicProducerDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient).onsconn
	topic := d.Get("topic").(string)
	producerId := d.Get("producerid").(string)
	regionId := d.Get("regionid").(string)

	return resource.Retry(2*time.Minute, func() *resource.RetryError {
		deleteTopicProducerRequest := ons.CreateDeleteProducerRequest(regionId, topic, producerId)
		deleteTopicProducerResponse, err := client.DeleteProducer(deleteTopicProducerRequest)
		if err != nil {
			return resource.RetryableError(fmt.Errorf("Delete topic_producer timeout and got an error: %s", err))
		}
		if (deleteTopicProducerResponse.IsSuccess() == false || deleteTopicProducerResponse.Success == false) {
			return resource.RetryableError(fmt.Errorf("Delete topic_producer timeout and got an error: [%s]%s", deleteTopicProducerResponse.Status, deleteTopicProducerResponse.Message))
		}
		return nil
	})
}