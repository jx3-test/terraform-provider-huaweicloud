package huaweicloud

import (
	"context"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/chnsz/golangsdk"
	v1groups "github.com/chnsz/golangsdk/openstack/networking/v1/security/securitygroups"
	v3groups "github.com/chnsz/golangsdk/openstack/networking/v3/security/groups"
	v3rules "github.com/chnsz/golangsdk/openstack/networking/v3/security/rules"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/logp"
)

func DataSourceNetworkingSecGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceNetworkingSecGroupRead,

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"secgroup_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enterprise_project_id": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"rules": securityGroupRuleSchema,
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func getRuleListByGroupId(client *golangsdk.ServiceClient, groupId string) ([]map[string]interface{}, error) {
	listOpts := v3rules.ListOpts{
		SecurityGroupId: groupId,
	}
	resp, err := v3rules.List(client, listOpts)
	if err != nil {
		return nil, err
	}
	return flattenSecurityGroupRulesV3(resp)
}

func dataSourceNetworkingSecGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := GetRegion(d, config)
	v3Client, err := config.NetworkingV3Client(region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud networking v3 client: %s", err)
	}

	listOpts := v3groups.ListOpts{
		ID:                  d.Get("secgroup_id").(string),
		Name:                d.Get("name").(string),
		EnterpriseProjectId: config.DataGetEnterpriseProjectID(d),
	}

	allSecGroups, err := v3groups.List(v3Client, listOpts)
	if err != nil {
		if _, ok := err.(golangsdk.ErrDefault404); ok {
			// If the v3 API does not exist or has not been published in the specified region, set again using v1 API.
			return dataSourceNetworkingSecGroupReadV1(ctx, d, meta)
		}
		return fmtp.DiagErrorf("Unable to get security groups list: %s", err)
	}

	if len(allSecGroups) < 1 {
		return fmtp.DiagErrorf("No Security Group found.")
	}

	if len(allSecGroups) > 1 {
		return fmtp.DiagErrorf("More than one Security Groups found.")
	}

	secGroup := allSecGroups[0]
	d.SetId(secGroup.ID)
	logp.Printf("[DEBUG] Retrieved Security Group (%s) by v3 client: %v", d.Id(), secGroup)

	rules, err := getRuleListByGroupId(v3Client, secGroup.ID)
	if err != nil {
		return diag.FromErr(err)
	}
	logp.Printf("[DEBUG] The rules list of security group (%s) is: %v", d.Id(), rules)

	mErr := multierror.Append(nil,
		d.Set("region", region),
		d.Set("name", secGroup.Name),
		d.Set("description", secGroup.Description),
		d.Set("rules", rules),
		d.Set("created_at", secGroup.CreatedAt),
		d.Set("updated_at", secGroup.UpdatedAt),
	)
	if mErr.ErrorOrNil() != nil {
		return diag.FromErr(mErr)
	}

	return nil
}

func dataSourceNetworkingSecGroupReadV1(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := GetRegion(d, config)
	v1Client, err := config.NetworkingV1Client(region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating HuaweiCloud networking v1 client: %s", err)
	}

	listOpts := v1groups.ListOpts{
		EnterpriseProjectId: config.DataGetEnterpriseProjectID(d),
	}

	pages, err := v1groups.List(v1Client, listOpts).AllPages()
	if err != nil {
		return diag.FromErr(err)
	}
	allGroups, err := v1groups.ExtractSecurityGroups(pages)
	if err != nil {
		return fmtp.DiagErrorf("Error retrieving security groups list: %s", err)
	}
	if len(allGroups) == 0 {
		return fmtp.DiagErrorf("No sucurity group found, please change your search criteria and try again.")
	}
	logp.Printf("[DEBUG] The retrieved group list is: %v", allGroups)

	filter := map[string]interface{}{
		"ID":   d.Get("secgroup_id"),
		"Name": d.Get("name"),
	}
	filterGroups, err := utils.FilterSliceWithField(allGroups, filter)
	if err != nil {
		return fmtp.DiagErrorf("Erroring filting security groups list: %s", err)
	}
	if len(filterGroups) < 1 {
		return fmtp.DiagErrorf("No Security Group found.")
	}
	if len(filterGroups) > 1 {
		return fmtp.DiagErrorf("More than one Security Groups found.")
	}

	resp := filterGroups[0].(v1groups.SecurityGroup)
	d.SetId(resp.ID)

	rules := flattenSecurityGroupRulesV1(&resp)
	logp.Printf("[DEBUG] The retrieved rules list is: %v", rules)

	mErr := multierror.Append(nil,
		d.Set("region", region),
		d.Set("name", resp.Name),
		d.Set("description", resp.Description),
		d.Set("enterprise_project_id", resp.EnterpriseProjectId),
		d.Set("rules", rules),
	)
	return diag.FromErr(mErr.ErrorOrNil())
}
