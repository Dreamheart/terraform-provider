package alicloud

import (
	"fmt"
	"strings"
	"time"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAliyunEniAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliyunEniAttachmentCreate,
		Read:   resourceAliyunEniAttachmentRead,
		Delete: resourceAliyunEniAttachmentDelete,

		Schema: map[string]*schema.Schema{
			"network_interface_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceAliyunEniAttachmentCreate(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*AliyunClient)

	args := ecs.CreateAttachNetworkInterfaceRequest()
	args.NetworkInterfaceId = Trim(d.Get("network_interface_id").(string))
	args.InstanceId = Trim(d.Get("instance_id").(string))

	if err := resource.Retry(3*time.Minute, func() *resource.RetryError {
		ar := args
		if _, err := client.ecsconn.AttachNetworkInterface(ar); err != nil {
			//if IsExceptedError(err, TaskConflict) {
			//	return resource.RetryableError(fmt.Errorf("Attach ENI got an error: %#v", err))
			//}
			return resource.NonRetryableError(fmt.Errorf("Attach ENI got an error: %#v", err))
		}
		return nil
	}); err != nil {
		return err
	}

	//if err := client.WaitForNetworkInterface(args.NetworkInterfaceId, InUse, 60); err != nil {
	//	return fmt.Errorf("Error Waitting for ENI attached: %#v", err)
	//}

	if err := client.WaitForInstanceContainNetworkInterface(args.InstanceId, args.NetworkInterfaceId, 60); err != nil{
		return fmt.Errorf("Error Waitting for ENI attached: %#v", err)
	}

	d.SetId(args.NetworkInterfaceId + ":" + args.InstanceId)

	return resourceAliyunEniAttachmentRead(d, meta)
}

func resourceAliyunEniAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient)

	networkInterfaceId, instanceId, err := getNetworkInterfaceIdAndInstanceId(d, meta)
	if err != nil {
		return err
	}

	//eni, err := client.DescribeNetworkInterfaceById(networkInterfaceId)
	inst, err := client.DescribeInstanceById(instanceId)

	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Describe Instance Attribute: %#v", err)
	}

	//if eni.InstanceId != instanceId {
	//	d.SetId("")
	//	return nil
	//}

	ready := false
	for _, eni := range inst.NetworkInterfaces.NetworkInterface{
		if eni.NetworkInterfaceId == networkInterfaceId{
			ready = true
			break
		}
	}
	if ready == false {
		d.SetId("")
		return nil
	}

	d.Set("instance_id", inst.InstanceId)
	d.Set("network_interface_id", networkInterfaceId)
	return nil
}

func resourceAliyunEniAttachmentDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*AliyunClient)

	networkInterfaceId, instanceId, err := getNetworkInterfaceIdAndInstanceId(d, meta)
	if err != nil {
		return err
	}

	request := ecs.CreateDetachNetworkInterfaceRequest()
	request.NetworkInterfaceId = networkInterfaceId
	request.InstanceId = instanceId

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := client.ecsconn.DetachNetworkInterface(request); err != nil {
			if IsExceptedError(err, InvalidEcsState)  {
				return resource.RetryableError(fmt.Errorf("Detach NetworkInterface timeout and got an error:%#v.", err))
			}
		}

		//eni, descErr := client.DescribeNetworkInterfaceById(networkInterfaceId)
		inst, descErr := client.DescribeInstanceById(instanceId)

		if descErr != nil {
			if NotFoundError(err) {
				return nil
			}
			return resource.NonRetryableError(descErr)
		}

		ready := false
		for _, eni := range inst.NetworkInterfaces.NetworkInterface{
			if eni.NetworkInterfaceId == networkInterfaceId{
				ready = true
				break
			}
		}
		if ready == true {
			return resource.RetryableError(fmt.Errorf("Detach NetworkInterface timeout and got an error:%#v.", err))
		}

		return nil
	})
}

func getNetworkInterfaceIdAndInstanceId(d *schema.ResourceData, meta interface{}) (string, string, error) {
	parts := strings.Split(d.Id(), ":")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid resource id")
	}
	return parts[0], parts[1], nil
}
