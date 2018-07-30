package alicloud

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/Dreamheart/apsarastack-mq-go-sdk/service/ons"
)

func dataSourceAlicloudOnsRegions() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAlicloudOnsRegionsRead,

		Schema: map[string]*schema.Schema{
			"region_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			//Computed value
			"ons_regions": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"region_name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceAlicloudOnsRegionsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AliyunClient).onsconn

	req := ons.CreateDescribeRegionsRequest()
	resp, err := conn.DescribeRegions(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.Data) == 0 {
		return fmt.Errorf("no matching ons_regions found")
	}
	name, nameOk := d.GetOk("region_name")
	var filterOnsRegions []ons.Region
	for _, region := range resp.Data {
		if nameOk {
			if name.(string) == region.RegionName {
				filterOnsRegions = append(filterOnsRegions, region)
				break
			}
			continue
		}
		filterOnsRegions = append(filterOnsRegions, region)
	}
	if len(filterOnsRegions) < 1 {
		return fmt.Errorf("Your query ons_regions returned no results. Please change your search criteria and try again.")
	}

	return onsRegionsDescriptionAttributes(d, filterOnsRegions)
}

func onsRegionsDescriptionAttributes(d *schema.ResourceData, regions []ons.Region) error {
	var ids []string
	var s []map[string]interface{}
	for _, region := range regions {
		mapping := map[string]interface{}{
			"id":         region.RegionId,
			"region_id":  region.RegionId,
			"region_name": region.RegionName,
			"status": 	region.Status,
		}

		log.Printf("[DEBUG] alicloud_ons_regions - adding region mapping: %v", mapping)
		ids = append(ids, string(region.RegionId))
		s = append(s, mapping)
	}

	d.SetId(dataResourceIdHash(ids))
	if err := d.Set("ons_regions", s); err != nil {
		return err
	}

	return nil
}
