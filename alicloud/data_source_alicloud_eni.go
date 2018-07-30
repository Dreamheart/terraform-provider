package alicloud

import (
	"fmt"
	"log"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/ecs"
)

func dataSourceAlicloudEnis() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudEnisRead,

		Schema: map[string]*schema.Schema{
			"ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
				MinItems: 1,
			},

			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"vswitch_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"primary_ip_address": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"network_interface_name": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"type": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"instance_id": {
				Type:     schema.TypeString,
				Optional: true,
			},

			// Computed values
			"network_interfaces": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},

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
				},
			},
		},
	}
}
func dataSourceAlicloudEnisRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AliyunClient).ecsconn

	args := ecs.CreateDescribeNetworkInterfacesRequest()
	args.RegionId = string(getRegion(d, meta))
	args.PageSize = requests.NewInteger(PageSizeLarge)

	if v, ok := d.GetOk("vpc_id"); ok {
		args.VpcId = Trim(v.(string))
	}

	if v, ok := d.GetOk("vswitch_id"); ok {
		args.VSwitchId = Trim(v.(string))
	}

	if v, ok := d.GetOk("primary_ip_address"); ok {
		args.PrimaryIpAddress = Trim(v.(string))
	}

	if v, ok := d.GetOk("security_group_id"); ok {
		args.SecurityGroupId = Trim(v.(string))
	}

	if v, ok := d.GetOk("network_interface_name"); ok {
		args.NetworkInterfaceName = Trim(v.(string))
	}

	if v, ok := d.GetOk("instance_id"); ok {
		args.InstanceId = Trim(v.(string))
	}

	if v, ok := d.GetOk("type"); ok {
		args.Type = Trim(v.(string))
	}

	idsList := make([]string,0)
	if v, ok := d.GetOk("ids"); ok {
		for _, vv := range(v.([]string)){
			idsList = append(idsList, vv)
		}
		args.NetworkInterfaceId = &idsList
	}

	var allEnis []ecs.NetworkInterfaceSet

	for {
		resp, err := conn.DescribeNetworkInterfaces(args)
		if err != nil {
			return err
		}

		if resp == nil || len(resp.NetworkInterfaceSets.NetworkInterfaceSet) < 1 {
			break
		}

		for _, eni := range resp.NetworkInterfaceSets.NetworkInterfaceSet {
			allEnis = append(allEnis, eni)
		}

		if len(resp.NetworkInterfaceSets.NetworkInterfaceSet) < PageSizeLarge {
			break
		}

		args.PageNumber = args.PageNumber + requests.NewInteger(1)
	}

	if len(allEnis) < 1 {
		return fmt.Errorf("Your query returned no results. Please change your search criteria and try again.")
	}

	log.Printf("[DEBUG] alicloud_eips - ENIs found: %#v", allEnis)

	return enisDecriptionAttributes(d, allEnis, meta)
}

func enisDecriptionAttributes(d *schema.ResourceData, eniSetTypes []ecs.NetworkInterfaceSet, meta interface{}) error {
	var ids []string
	var s []map[string]interface{}
	for _, eni := range eniSetTypes {

		var sg_ids  = make([]string,0)
		for _, v := range eni.SecurityGroupIds.SecurityGroupId {
			sg_ids = append(sg_ids, v)
		}

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

		mapping := map[string]interface{}{
			"id":                   	eni.NetworkInterfaceId,
			"network_interface_name":	eni.NetworkInterfaceName,
			"vpc_id":					eni.VpcId,
			"vswitch_id":				eni.VSwitchId,
			"zone_id":					eni.ZoneId,
			"security_group_ids": 		sg_ids,
			"primary_ip_address":		eni.PrivateIpAddress,
			"private_ip_set":			private_ip_sets,
			"associated_public_ip":		map[string]interface{}{
				"allocation_id":   		eni.AssociatedPublicIp.AllocationId,
				"public_ip_address":	eni.AssociatedPublicIp.PublicIpAddress,
			},
			"description":				eni.Description,
			"status":					eni.Status,
			"mac_address":				eni.MacAddress,
			"instance_id":				eni.InstanceId,
			"creation_time":        	eni.CreationTime,
		}
		log.Printf("[DEBUG] alicloud_network_interface - adding eni: %v", mapping)
		ids = append(ids, eni.NetworkInterfaceId)
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("network_interfaces", s); err != nil {
		return err
	}

	return nil
}
