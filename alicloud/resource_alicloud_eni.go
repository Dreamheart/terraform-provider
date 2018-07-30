package alicloud

import (
	"fmt"
	"time"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func resourceAliyunEni() *schema.Resource {
	return &schema.Resource{
		Create: resourceAliyunEniCreate,
		Read:   resourceAliyunEniRead,
		Update: resourceAliyunEniUpdate,
		Delete: resourceAliyunEniDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"vswitch_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"security_group_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"primary_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"network_interface_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"client_token": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"status": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"zone_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"associated_public_ip": &schema.Schema{
				Type:     schema.TypeMap,
				Computed: true,
				Elem: map[string]*schema.Schema{
					"public_ip_address": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
					"allocation_id": &schema.Schema{
						Type:     schema.TypeString,
						Computed: true,
					},
				},
			},

			"private_ip_set": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"private_ip_address": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"primary": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"associated_public_ip": &schema.Schema{
							Type:     schema.TypeMap,
							Computed: true,
							Elem: map[string]*schema.Schema{
								"public_ip_address": &schema.Schema{
									Type:     schema.TypeString,
									Computed: true,
								},
								"allocation_id": &schema.Schema{
									Type:     schema.TypeString,
									Computed: true,
								},
							},
						},
					},
				},
			},

			"private_ip_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"mac_address": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"security_group_ids": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{Type: schema.TypeString},
			},

			"instance_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"creation_time": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAliyunEniCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient)

	request := ecs.CreateCreateNetworkInterfaceRequest()
	request.RegionId = string(getRegion(d, meta))
	request.VSwitchId = d.Get("vswitch_id").(string)
	request.SecurityGroupId = d.Get("security_group_id").(string)

	if primary_ip_address, ok := d.GetOk("primary_ip_address"); ok == true && primary_ip_address.(string) != ""{
		request.PrimaryIpAddress = primary_ip_address.(string)
	}

	if network_interface_name, ok := d.GetOk("network_interface_name"); ok == true && network_interface_name.(string) != ""{
		request.NetworkInterfaceName = network_interface_name.(string)
	}

	if description, ok := d.GetOk("description"); ok == true && description.(string) != ""{
		request.Description = description.(string)
	}

	if client_token, ok := d.GetOk("client_token"); ok == true && client_token.(string) != ""{
		request.ClientToken = client_token.(string)
	}

	eni, err := client.ecsconn.CreateNetworkInterface(request)

	if err != nil {
		return fmt.Errorf("Error Creating NetworkInterface: %#v", err)
	}

	d.SetId(eni.NetworkInterfaceId)

	if err := client.WaitForNetworkInterface(d.Id(), Available, DefaultTimeoutMedium); err != nil {
		return fmt.Errorf("WaitForNetworkInterface %s got error: %#v", Available, err)
	}

	return resourceAliyunEniUpdate(d, meta)
}

func resourceAliyunEniRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient)

	eni, err := client.DescribeNetworkInterfaceById(d.Id())

	if err != nil {
		if NotFoundError(err) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error Describe ENI Attribute: %#v", err)
	}

	d.Set("status", eni.Status)
	d.Set("type", eni.Type)
	d.Set("vpc_id", eni.VpcId)
	d.Set("vswitch_id", eni.VSwitchId)
	d.Set("zone_id", eni.ZoneId)

	associated_public_ip := map[string]interface{}{
		"allocation_id":   		eni.AssociatedPublicIp.AllocationId,
		"public_ip_address":	eni.AssociatedPublicIp.PublicIpAddress,
	}

	d.Set("associated_public_ip", associated_public_ip)
	d.Set("private_ip_address", eni.PrivateIpAddress)
	d.Set("mac_address", eni.MacAddress)

	var private_ip_sets []map[string]interface{}
	for _, v := range eni.PrivateIpSets.PrivateIpSet {
		mapping := map[string]interface{}{
			"private_ip_address":   v.PrivateIpAddress,
			"primary":              v.Primary,
			"associated_public_ip":	map[string]interface{}{
				"allocation_id":   		v.AssociatedPublicIp.AllocationId,
				"public_ip_address":	v.AssociatedPublicIp.PublicIpAddress,
			},
		}
		private_ip_sets = append(private_ip_sets, mapping)
	}
	d.Set("private_ip_set", private_ip_sets)

	var sg_ids  = make([]string,0)
	for _, v := range eni.SecurityGroupIds.SecurityGroupId {
		sg_ids = append(sg_ids, v)
	}
	d.Set("security_group_ids", sg_ids)

	d.Set("network_interface_name", eni.NetworkInterfaceName)
	d.Set("description", eni.Description)
	d.Set("instance_id", eni.InstanceId)
	d.Set("creation_time", eni.CreationTime)

	return nil
}

func resourceAliyunEniUpdate(d *schema.ResourceData, meta interface{}) error {

	d.Partial(true)

	attributeUpdate := false
	args := ecs.CreateModifyNetworkInterfaceAttributeRequest()
	args.NetworkInterfaceId = d.Id()

	if d.HasChange("network_interface_name") && !d.IsNewResource() {
		d.SetPartial("network_interface_name")
		args.NetworkInterfaceName = d.Get("network_interface_name").(string)
		attributeUpdate = true
	}

	if d.HasChange("description") && !d.IsNewResource() {
		d.SetPartial("description")
		args.Description = d.Get("description").(string)
		attributeUpdate = true
	}

	if d.HasChange("security_group_id") && !d.IsNewResource() {
		d.SetPartial("security_group_id")
		args.SecurityGroupId = &[]string{ 0:d.Get("security_group_id").(string) }
		attributeUpdate = true
	}

	if attributeUpdate {
		if _, err := meta.(*AliyunClient).ecsconn.ModifyNetworkInterfaceAttribute(args); err != nil {
			return err
		}
	}

	d.Partial(false)

	return resourceAliyunEniRead(d, meta)
}

func resourceAliyunEniDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AliyunClient)

	request := ecs.CreateDeleteNetworkInterfaceRequest()

	request.NetworkInterfaceId = d.Id()

	return resource.Retry(5*time.Minute, func() *resource.RetryError {
		if _, err := client.ecsconn.DeleteNetworkInterface(request); err != nil{
			if IsExceptedError(err, DetachPrimaryEniNotAllowed) ||
				IsExceptedError(err, InvalidEniType) ||
				IsExceptedError(err, InvalidEniState) {
				return resource.RetryableError(fmt.Errorf("Delete ENI timeout and got an error:%#v.", err))
			}
			return resource.NonRetryableError(err)
		}


		eni, descErr := client.DescribeNetworkInterfaceById(d.Id())

		if descErr != nil {
			if NotFoundError(descErr) {
				return nil
			}
			return resource.NonRetryableError(descErr)
		} else if eni.NetworkInterfaceId == d.Id() {
			return resource.RetryableError(fmt.Errorf("Delete ENI timeout and it still exists."))
		}
		return nil
	})
}
