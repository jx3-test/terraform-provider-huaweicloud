package ecs

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/chnsz/golangsdk"
	"github.com/chnsz/golangsdk/openstack/ecs/v1/nics"
	"github.com/chnsz/golangsdk/openstack/networking/v2/ports"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/common"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/config"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/fmtp"
	"github.com/huaweicloud/terraform-provider-huaweicloud/huaweicloud/utils/logp"
)

func ResourceComputeInterfaceAttachV2() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceComputeNicAttachCreate,
		ReadContext:   resourceComputeNicAttachRead,
		UpdateContext: resourceComputeNicAttachUpdate,
		DeleteContext: resourceComputeNicAttachDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"instance_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"network_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"network_id", "subnet_id"},
				Deprecated:   "Please use subnet_id instead.",
			},

			"subnet_id": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"network_id", "subnet_id"},
			},

			"fixed_ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},

			"security_group_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				// 查询接口没有返回， 单独的安全组是openstack的接口，不支持eps
			},

			"source_dest_check": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},

			"port_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"mac": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"attachable_quantity": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceComputeNicAttachCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)
	computeClient, err := config.ComputeV1Client(region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating compute client: %s", err)
	}

	instanceId := d.Get("instance_id").(string)

	var subnetId string
	if v, ok := d.GetOk("network_id"); ok {
		subnetId = v.(string)
	}

	if v, ok := d.GetOk("subnet_id"); ok {
		subnetId = v.(string)
	}

	var securityGroupIds []nics.IdInfo
	if v, ok := d.GetOk("security_group_id"); ok {
		securityGroupIds = append(securityGroupIds, nics.IdInfo{Id: v.(string)})
	}

	currentNics, err := nics.Get(computeClient, instanceId)
	if err != nil {
		return fmtp.DiagErrorf("Error getting the nics of ECS: %s", err)
	}

	opts := nics.CreateOps{
		Nics: []nics.NicReq{
			{
				SubnetId:       subnetId,
				IpAddress:      d.Get("fixed_ip").(string),
				SecurityGroups: securityGroupIds,
			},
		},
	}
	createRst := nics.Create(computeClient, instanceId, opts)
	if createRst.Err != nil {
		return fmtp.DiagErrorf("Error creating huaweicloud_compute_interface_attach: %s", createRst.Err)
	}

	time.Sleep(10 * time.Second)

	portID, err := GetNewPortIdOfNic(computeClient, instanceId, currentNics.InterfaceAttachments, opts.Nics[0])
	if err != nil {
		return fmtp.DiagErrorf("Error getting the nics of ECS: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:      []string{"ATTACHING"},
		Target:       []string{"ATTACHED"},
		Refresh:      nicStateChangeFunc(computeClient, instanceId, portID),
		Timeout:      d.Timeout(schema.TimeoutCreate),
		Delay:        5 * time.Second,
		PollInterval: 5 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return fmtp.DiagErrorf("Error creating ECS Nics %s: %s", instanceId, err)
	}

	// Use the instance ID and port ID as the resource ID.
	id := fmt.Sprintf("%s/%s", instanceId, portID)
	d.SetId(id)

	if sourceDestCheck := d.Get("source_dest_check").(bool); !sourceDestCheck {
		nicClient, err := config.NetworkingV2Client(region)
		if err != nil {
			return fmtp.DiagErrorf("Error creating Networking V2 client: %s", err)
		}
		if err := disableSourceDestCheck(nicClient, portID); err != nil {
			return fmtp.DiagErrorf("Error disable source dest check on port(%s) of instance(%s) failed: %s", portID,
				instanceId, err)
		}
	}

	return nil
}

func resourceComputeNicAttachRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)
	computeClient, err := config.ComputeV1Client(region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating compute V2 client: %s", err)
	}

	networkingClient, err := config.NetworkingV2Client(region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating Networking V2 client: %s", err)
	}

	instanceId, portId, err := ComputeInterfaceAttachV2ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	resp, err := nics.Get(computeClient, instanceId)
	if err != nil {
		return fmtp.DiagErrorf("Error retrieving ECS nics: %s", err)
	}

	attachment := GetNicOfEcs(resp, portId)
	if attachment == nil {
		return common.CheckDeletedDiag(d, err, "Error retrieving ECS nics")
	}

	logp.Printf("[DEBUG] Retrieved huaweicloud_compute_interface_attach %s: %#v", d.Id(), attachment)

	port, err := ports.Get(networkingClient, portId).Extract()
	if err != nil {
		return fmtp.DiagErrorf("Error retrieving port information of ECS nics%s: %s", d.Id(), err)
	}

	mErr := multierror.Append(
		d.Set("region", region),
		d.Set("instance_id", instanceId),
		d.Set("port_id", attachment.PortId),
		d.Set("mac", attachment.MacAddr),
		d.Set("fixed_ip", attachment.FixedIps[0].IPAddress),
		d.Set("network_id", attachment.NetworkId),
		d.Set("source_dest_check", len(port.AllowedAddressPairs) == 0),
		d.Set("attachable_quantity", resp.AttachableQuantity.FreeNic),
	)
	if err := mErr.ErrorOrNil(); err != nil {
		return fmtp.DiagErrorf("Error setting ECS nic fields: %s", err)
	}

	return nil
}

func resourceComputeNicAttachUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)
	nicClient, err := config.NetworkingV2Client(region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating Networking V2 client: %s", err)
	}

	portID := d.Get("port_id").(string)
	if sourceDestCheck := d.Get("source_dest_check").(bool); !sourceDestCheck {
		err = disableSourceDestCheck(nicClient, portID)
	} else {
		err = enableSourceDestCheck(nicClient, portID)
	}

	if err != nil {
		return fmtp.DiagErrorf("Error updating source_dest_check on port(%s) of instance(%s) failed: %s", portID, d.Id(), err)
	}

	return resourceComputeNicAttachRead(ctx, d, meta)
}

func resourceComputeNicAttachDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*config.Config)
	region := config.GetRegion(d)
	computeClient, err := config.ComputeV1Client(region)
	if err != nil {
		return fmtp.DiagErrorf("Error creating compute V2 client: %s", err)
	}

	instanceId, portId, err := ComputeInterfaceAttachV2ParseID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{""},
		Target:     []string{"UNFOUND"},
		Refresh:    nicStateChangeFunc(computeClient, instanceId, portId),
		Timeout:    d.Timeout(schema.TimeoutDelete),
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}

	if _, err = stateConf.WaitForStateContext(ctx); err != nil {
		return fmtp.DiagErrorf("Error detaching ECS nics %s: %s", d.Id(), err)
	}

	return nil
}

func nicStateChangeFunc(client *golangsdk.ServiceClient, instanceId, newPortId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := nics.Get(client, instanceId)
		if err != nil {
			return nil, "failed", err
		}

		for _, v := range resp.InterfaceAttachments {
			if v.PortId == newPortId {
				return resp, v.PortState, nil
			}
		}

		return resp, "UNFOUND", err
	}
}

func GetNewPortIdOfNic(client *golangsdk.ServiceClient, instanceId string, currentNics []nics.Nic, newNic nics.NicReq) (string, error) {
	newNics, err := nics.Get(client, instanceId)
	if err != nil {
		return "", fmtp.Errorf("Error getting the nics of ECS: %s", err)
	}

	curPorts := make([]string, len(currentNics))
	for i, v := range currentNics {
		curPorts[i] = v.PortId
	}

	var newPortId string
	for _, v := range newNics.InterfaceAttachments {
		if !utils.StrSliceContains(curPorts, v.PortId) &&
			v.FixedIps[0].IPAddress == newNic.IpAddress && v.FixedIps[0].SubnetID == newNic.SubnetId {
			newPortId = v.PortId
		}
	}

	return newPortId, nil
}

func GetNicOfEcs(resp *nics.NicsResp, portId string) *nics.Nic {
	var attachment *nics.Nic
	for _, v := range resp.InterfaceAttachments {
		if v.PortId == portId {
			attachment = &v
		}
	}
	return attachment
}

func ComputeInterfaceAttachV2ParseID(id string) (string, string, error) {
	idParts := strings.Split(id, "/")
	if len(idParts) < 2 {
		return "", "", fmtp.Errorf("Unable to determine huaweicloud_compute_interface_attach_v2 %s ID", id)
	}

	instanceId := idParts[0]
	attachmentId := idParts[1]

	return instanceId, attachmentId, nil
}

func disableSourceDestCheck(networkClient *golangsdk.ServiceClient, portID string) error {
	// Update the allowed-address-pairs of the port to 1.1.1.1/0
	// to disable the source/destination check
	portpairs := []ports.AddressPair{
		{
			IPAddress: "1.1.1.1/0",
		},
	}
	portUpdateOpts := ports.UpdateOpts{
		AllowedAddressPairs: &portpairs,
	}

	_, err := ports.Update(networkClient, portID, portUpdateOpts).Extract()
	return err
}

func enableSourceDestCheck(networkClient *golangsdk.ServiceClient, portID string) error {
	// cancle all allowed-address-pairs to enable the source/destination check
	portpairs := make([]ports.AddressPair, 0)
	portUpdateOpts := ports.UpdateOpts{
		AllowedAddressPairs: &portpairs,
	}

	_, err := ports.Update(networkClient, portID, portUpdateOpts).Extract()
	return err
}
